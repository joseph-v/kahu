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

	"github.com/soda-cdm/kahu/utils"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	metaservice "github.com/soda-cdm/kahu/providerframework/metaservice/lib/go"
)

func (ctrl *controller) GetDeploymentAndBackup(name, namespace string,
	backupClient metaservice.MetaService_BackupClient) error {
	deployment, err := ctrl.kubeClient.AppsV1().
		Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	err = ctrl.backupSend(deployment, deployment.Name, backupClient)
	if err != nil {
		return err
	}
	return nil

}

func (ctrl *controller) deploymentBackup(namespace string,
	backup *PrepareBackup, backupClient metaservice.MetaService_BackupClient) error {
	ctrl.logger.Infoln("Starting collecting deployments")
	var labelSelectors map[string]string
	if backup.Spec.Label != nil {
		labelSelectors = backup.Spec.Label.MatchLabels
	}

	selectors := labels.Set(labelSelectors).String()
	dList, err := ctrl.kubeClient.AppsV1().Deployments(namespace).
		List(context.TODO(), metav1.ListOptions{
			LabelSelector: selectors,
		})
	if err != nil {
		return err
	}
	var deploymentAllList []string
	for _, deployment := range dList.Items {
		deploymentAllList = append(deploymentAllList, deployment.Name)
	}

	deploymentAllList = utils.FindMatchedStrings(utils.Deployment, deploymentAllList, backup.Spec.IncludeResources,
		backup.Spec.ExcludeResources)

	for _, deployment := range dList.Items {
		if utils.Contains(deploymentAllList, deployment.Name) {
			// backup the deployment yaml
			err = ctrl.GetDeploymentAndBackup(deployment.Name, deployment.Namespace, backupClient)
			if err != nil {
				return err
			}

			// backup the volumespec releted object like, configmaps, secret, pvc and sc
			err = ctrl.GetVolumesSpec(deployment.Spec.Template.Spec, deployment.Namespace, backupClient)
			if err != nil {
				return err
			}

			// get service account relared objects
			err = ctrl.GetServiceAccountSpec(deployment.Spec.Template.Spec, deployment.Namespace, backupClient)
			if err != nil {
				return err
			}

			// get services based on selectors
			var selectorList []map[string]string
			selectorList = append(selectorList, deployment.Spec.Selector.MatchLabels)
			err = ctrl.GetServiceForPod(deployment.Namespace, selectorList, backupClient, ctrl.kubeClient)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ctrl *controller) GetDaemonSetAndBackup(name, namespace string,
	backupClient metaservice.MetaService_BackupClient) error {
	daemonset, err := ctrl.kubeClient.AppsV1().
		DaemonSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	err = ctrl.backupSend(daemonset, daemonset.Name, backupClient)
	if err != nil {
		return err
	}
	return nil

}

func (ctrl *controller) daemonSetBackup(namespace string,
	backup *PrepareBackup, backupClient metaservice.MetaService_BackupClient) error {
	ctrl.logger.Infoln("Starting collecting daemonset")
	var labelSelectors map[string]string
	if backup.Spec.Label != nil {
		labelSelectors = backup.Spec.Label.MatchLabels
	}
	selectors := labels.Set(labelSelectors).String()

	dList, err := ctrl.kubeClient.AppsV1().DaemonSets(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selectors,
	})
	if err != nil {
		return err
	}
	var daemonsetAllList []string
	for _, daemonset := range dList.Items {
		daemonsetAllList = append(daemonsetAllList, daemonset.Name)
	}

	daemonsetAllList = utils.FindMatchedStrings(utils.Daemonset, daemonsetAllList, backup.Spec.IncludeResources,
		backup.Spec.ExcludeResources)

	for _, daemonset := range dList.Items {
		if utils.Contains(daemonsetAllList, daemonset.Name) {
			// backup the daemonset yaml
			err = ctrl.GetDaemonSetAndBackup(daemonset.Name, daemonset.Namespace, backupClient)
			if err != nil {
				return err
			}

			// backup the volumespec releted object like, configmaps, secret, pvc and sc
			err = ctrl.GetVolumesSpec(daemonset.Spec.Template.Spec, daemonset.Namespace, backupClient)
			if err != nil {
				return err
			}

			// get service account relared objects
			err = ctrl.GetServiceAccountSpec(daemonset.Spec.Template.Spec, daemonset.Namespace, backupClient)
			if err != nil {
				return err
			}

			// get services based on selectors
			var selectorList []map[string]string
			selectorList = append(selectorList, daemonset.Spec.Selector.MatchLabels)
			err = ctrl.GetServiceForPod(daemonset.Namespace, selectorList, backupClient, ctrl.kubeClient)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ctrl *controller) GetServiceAccount(namespace, name string) (*corev1.ServiceAccount, error) {
	sa, err := ctrl.kubeClient.CoreV1().
		ServiceAccounts(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return sa, err
}

func (ctrl *controller) GetConfigMap(namespace, name string) (*corev1.ConfigMap, error) {
	configmap, err := ctrl.kubeClient.CoreV1().
		ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return configmap, err
}

func (ctrl *controller) ListNamespaces(backup *PrepareBackup) ([]string, error) {
	var namespaceList []string
	namespaces, err := ctrl.kubeClient.CoreV1().
		Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return namespaceList, err
	}
	for _, ns := range namespaces.Items {
		namespaceList = append(namespaceList, ns.Name)

	}
	return namespaceList, nil
}

func (ctrl *controller) getPersistentVolumeClaims(namespace string, backup *PrepareBackup,
	backupClient metaservice.MetaService_BackupClient) error {
	ctrl.logger.Infoln("Starting collecting persistentvolumeclaims")
	var labelSelectors map[string]string
	if backup.Spec.Label != nil {
		labelSelectors = backup.Spec.Label.MatchLabels
	}

	selectors := labels.Set(labelSelectors).String()
	allPVC, err := ctrl.kubeClient.CoreV1().
		PersistentVolumeClaims(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selectors,
	})
	if err != nil {
		return err
	}

	var allPVCList []string
	for _, pvc := range allPVC.Items {
		allPVCList = append(allPVCList, pvc.Name)
	}

	allPVCList = utils.FindMatchedStrings(utils.PVC, allPVCList, backup.Spec.IncludeResources,
		backup.Spec.ExcludeResources)

	for _, item := range allPVC.Items {
		if utils.Contains(allPVCList, item.Name) {
			pvcData, err := ctrl.GetPVC(namespace, item.Name)
			if err != nil {
				return err
			}
			err = ctrl.backupSend(pvcData, pvcData.Name, backupClient)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func (ctrl *controller) GetPVC(namespace, name string) (*corev1.PersistentVolumeClaim, error) {
	pvc, err := ctrl.kubeClient.CoreV1().
		PersistentVolumeClaims(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pvc, err
}

func (ctrl *controller) getStorageClass(backup *PrepareBackup,
	backupClient metaservice.MetaService_BackupClient) error {
	ctrl.logger.Infoln("Starting collecting storageclass")

	var labelSelectors map[string]string
	if backup.Spec.Label != nil {
		labelSelectors = backup.Spec.Label.MatchLabels
	}

	selectors := labels.Set(labelSelectors).String()
	allSC, err := ctrl.kubeClient.StorageV1().StorageClasses().List(context.TODO(), metav1.ListOptions{
		LabelSelector: selectors,
	})
	if err != nil {
		return err
	}

	var allSCList []string
	for _, sc := range allSC.Items {
		allSCList = append(allSCList, sc.Name)
	}

	allSCList = utils.FindMatchedStrings(utils.Sc, allSCList, backup.Spec.IncludeResources,
		backup.Spec.ExcludeResources)

	for _, item := range allSC.Items {
		if utils.Contains(allSCList, item.Name) {
			scData, err := ctrl.GetSC(item.Name)
			err = ctrl.backupSend(scData, scData.Name, backupClient)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ctrl *controller) GetSC(name string) (*storagev1.StorageClass, error) {
	sc, err := ctrl.kubeClient.StorageV1().StorageClasses().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return sc, err
}
