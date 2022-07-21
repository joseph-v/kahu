/*
Copyright 2022 The SODA Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package backup

import (
	"context"
	"encoding/json"
	kahuapi "github.com/soda-cdm/kahu/apis/kahu/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	kahuclient "github.com/soda-cdm/kahu/client/clientset/versioned/typed/kahu/v1"
	kahulister "github.com/soda-cdm/kahu/client/listers/kahu/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// Reconciler runs a periodic loop to reconcile the current state of volume back and update
// backup annotation
type Reconciler interface {
	Run(stopCh <-chan struct{})
}

// newReconciler returns a new instance of Reconciler that waits loopPeriod
// between successive executions.
func newReconciler(
	loopPeriod time.Duration,
	logger log.FieldLogger,
	volumeBackupLister kahulister.VolumeBackupContentLister,
	backupClient kahuclient.BackupInterface,
	backupLister kahulister.BackupLister) Reconciler {
	return &reconciler{
		loopPeriod:         loopPeriod,
		logger:             logger,
		volumeBackupLister: volumeBackupLister,
		backupLister:       backupLister,
		backupClient:       backupClient,
	}
}

type reconciler struct {
	loopPeriod         time.Duration
	logger             log.FieldLogger
	volumeBackupLister kahulister.VolumeBackupContentLister
	backupClient       kahuclient.BackupInterface
	backupLister       kahulister.BackupLister
}

func (rc *reconciler) Run(stopCh <-chan struct{}) {
	wait.Until(rc.reconciliationLoopFunc(), rc.loopPeriod, stopCh)
}

func (rc *reconciler) reconciliationLoopFunc() func() {
	return func() {
		rc.reconcile()
	}
}

// reconcile update backup with volume backup updates
func (rc *reconciler) reconcile() {
	backups, err := rc.backupLister.List(labels.Everything())
	if err != nil {
		rc.logger.Errorf("Unable to list backup list to reconcile")
	}

	for _, backup := range backups {
		backupName := backup.Name
		if backup.DeletionTimestamp != nil {
			deleted, err := rc.isVolumeBackupContentDeleted(backup)
			if err != nil {
				rc.logger.Errorf("Failed to check deleted volume backup status. %s", err)
				continue
			}

			if deleted {
				err := rc.annotateBackup(annVolumeBackupDeleteCompleted, backupName)
				if err != nil {
					rc.logger.Errorf("Unable to add annotation(%s) for backup(%s)",
						annVolumeBackupDeleteCompleted,
						backupName)
				}
				continue
			}
		}

		if !rc.isReconcileRequired(backup) {
			rc.logger.Debugf("Skipping reconcile for backup %s", backupName)
			continue
		}

		// annotate with volume completeness if no volume for backup
		vbContents, err := rc.volumeBackupLister.List(labels.Set{
			volumeContentBackupLabel: backupName,
		}.AsSelector())
		if err != nil {
			rc.logger.Errorf("Unable to list volume backup content for backup(%s). %s",
				backupName, err)
			continue
		}

		// may be lister not populated with volume backup contents
		if len(vbContents) == 0 {
			continue
		}

		volumesBackupDone := true
		volumeBackupFailed := false
		for _, vbc := range vbContents {
			if vbc.Status.Phase == kahuapi.VolumeBackupContentPhaseFailed {
				volumeBackupFailed = true
				break
			}
			if vbc.Status.Phase != kahuapi.VolumeBackupContentPhaseCompleted {
				volumesBackupDone = false
				break
			}
		}

		if volumeBackupFailed {
			// update backup status failure
			_, err = rc.updateBackupFailure(backup)
			if err == nil {
				rc.logger.Errorf("Unable to update backup(%s) failure", backup.Name)
			}
		}

		if !volumesBackupDone {
			continue
		}

		err = rc.annotateBackup(annVolumeBackupCompleted, backupName)
		if err != nil {
			rc.logger.Errorf("Unable to annotate backup(%s) with volume backup completeness", backupName)
		}
	}
}

func (rc *reconciler) isVolumeBackupContentDeleted(backup *kahuapi.Backup) (bool, error) {
	backupName := backup.Name
	// annotate with volume backup deletion if no volume for backup
	vbContents, err := rc.volumeBackupLister.List(labels.Set{
		volumeContentBackupLabel: backupName,
	}.AsSelector())
	if err == nil && len(vbContents) > 0 {
		rc.logger.Debug("VolumeBackupContents still available in backup %s", backupName)
		return false, nil
	}

	if apierrors.IsNotFound(err) ||
		len(vbContents) == 0 {
		return true, nil
	}

	return false, err
}

func (rc *reconciler) isReconcileRequired(backup *kahuapi.Backup) bool {
	// skip reconcile if backup.Status.Phase is not volume backup
	// skip reconcile if volume backup completed
	if backup.Status.Stage != kahuapi.BackupStageVolumes ||
		metav1.HasAnnotation(backup.ObjectMeta, annVolumeBackupCompleted) {
		return false
	}
	return true
}

func (rc *reconciler) annotateBackup(
	annotation string,
	backupName string) error {
	rc.logger.Infof("Annotating backup(%s) with %s", backupName, annotation)

	backup, err := rc.backupLister.Get(backupName)
	if err != nil {
		rc.logger.Errorf("Unable to get backup(%s) for %s. %s",
			backupName, annotation, err)
		return errors.Wrap(err, "Unable to update backup")
	}

	_, ok := backup.Annotations[annotation]
	if ok {
		rc.logger.Infof("Backup(%s) all-ready annotated with %s", backupName, annotation)
		return nil
	}

	backupClone := backup.DeepCopy()
	metav1.SetMetaDataAnnotation(&backupClone.ObjectMeta, annotation, "true")

	origBytes, err := json.Marshal(backup)
	if err != nil {
		return errors.Wrap(err, "error marshalling backup")
	}

	updatedBytes, err := json.Marshal(backupClone)
	if err != nil {
		return errors.Wrap(err, "error marshalling updated backup")
	}

	patchBytes, err := jsonpatch.CreateMergePatch(origBytes, updatedBytes)
	if err != nil {
		return errors.Wrap(err, "error creating json merge patch for backup")
	}

	_, err = rc.backupClient.Patch(context.TODO(), backupName,
		types.MergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		rc.logger.Error("Unable to update backup(%s) for volume completeness. %s",
			backupName, err)
		errors.Wrap(err, "error annotating volume backup completeness")
	}

	return nil
}

func (rc *reconciler) updateBackupFailure(
	backup *kahuapi.Backup) (*kahuapi.Backup, error) {
	var err error

	backupClone := backup.DeepCopy()
	backupClone.Status.State = kahuapi.BackupStateFailed
	backupClone, err = rc.backupClient.UpdateStatus(context.TODO(), backupClone, metav1.UpdateOptions{})
	if err != nil {
		rc.logger.Errorf("updating backup(%s) status: update status failed %s", backup.Name, err)
	}

	return backupClone, err
}
