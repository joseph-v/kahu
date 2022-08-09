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

package restore

import (
	"context"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"

	kahuapi "github.com/soda-cdm/kahu/apis/kahu/v1"
	"github.com/soda-cdm/kahu/hooks"
	"github.com/soda-cdm/kahu/utils"
)

var (
	excludeResources = sets.NewString(
		"Node",
		"Namespace",
		"Event",
	)
)

type backupInfo struct {
	backup         *kahuapi.Backup
	backupLocation *kahuapi.BackupLocation
	backupProvider *kahuapi.Provider
}

func (ctx *restoreContext) syncMetadataRestore(restore *kahuapi.Restore) error {
	// metadata restore should be last step for restore
	ctx.logger.Infof("Restore in %s phase", kahuapi.RestoreStageResources)

	// process CRD resource first
	err := ctx.applyCRD(restore)
	if err != nil {
		restore.Status.State = kahuapi.RestoreStateFailed
		restore.Status.FailureReason = fmt.Sprintf("Failed to apply CRD resources. %s", err)
		restore, err = ctx.updateRestoreStatus(restore)
		return err
	}

	// process resources
	err = ctx.applyIndexedResource(restore)
	if err != nil {
		restore.Status.State = kahuapi.RestoreStateFailed
		restore.Status.FailureReason = fmt.Sprintf("Failed to apply resources. %s", err)
		restore, err = ctx.updateRestoreStatus(restore)
		return err
	}

	// Wait for all post-restore exec hooks with same logic as restic wait above.
	go func() {
		ctx.logger.Info("Waiting for all post-restore-exec hooks to complete")

		ctx.hooksWaitGroup.Wait()
		close(ctx.hooksErrs)
	}()
	var errs []string
	for err := range ctx.hooksErrs {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		ctx.logger.Errorf("Done waiting for all post-restore exec hooks to complete %+v", errs)
	}
	ctx.logger.Info("Done waiting for all post-restore exec hooks to complete")

	restore.Status.Stage = kahuapi.RestoreStageFinished
	restore.Status.State = kahuapi.RestoreStateCompleted
	time := metav1.Now()
	restore.Status.CompletionTimestamp = &time
	restore, err = ctx.updateRestoreStatus(restore)
	return err
}

func (ctx *restoreContext) applyCRD(restore *kahuapi.Restore) error {
	crds, err := ctx.backupObjectIndexer.ByIndex(backupObjectResourceIndex, crdName)
	if err != nil {
		ctx.logger.Errorf("error fetching CRDs from indexer %s", err)
		return err
	}

	unstructuredCRDs := make([]*unstructured.Unstructured, 0)
	for _, crd := range crds {
		ubstructure, ok := crd.(*unstructured.Unstructured)
		if !ok {
			ctx.logger.Warningf("Restore index cache has invalid object type. %v",
				reflect.TypeOf(crd))
			continue
		}
		unstructuredCRDs = append(unstructuredCRDs, ubstructure)
	}

	for _, unstructuredCRD := range unstructuredCRDs {
		err := ctx.applyResource(unstructuredCRD, restore)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ctx *restoreContext) applyIndexedResource(restore *kahuapi.Restore) error {
	indexedResources := ctx.backupObjectIndexer.List()

	unstructuredResources := make([]*unstructured.Unstructured, 0)
	for _, indexedResource := range indexedResources {
		unstructuredResource, ok := indexedResource.(*unstructured.Unstructured)
		if !ok {
			ctx.logger.Warningf("Restore index cache has invalid object type. %v",
				reflect.TypeOf(unstructuredResource))
			continue
		}

		// ignore CRDs
		if unstructuredResource.GetObjectKind().GroupVersionKind().Kind == crdName {
			continue
		}
		unstructuredResources = append(unstructuredResources, unstructuredResource)
	}

	for _, unstructuredResource := range unstructuredResources {
		ctx.logger.Infof("Processing %s/%s for restore",
			unstructuredResource.GroupVersionKind(),
			unstructuredResource.GetName())
		err := ctx.applyResource(unstructuredResource, restore)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ctx *restoreContext) applyResource(resource *unstructured.Unstructured, restore *kahuapi.Restore) error {
	gvk := resource.GroupVersionKind()
	gvr, _, err := ctx.discoveryHelper.ByGroupVersionKind(gvk)
	if err != nil {
		ctx.logger.Errorf("unable to fetch GroupVersionResource for %s : %x", gvk.String(), []byte(gvk.String()))
		return err
	}

	resourceClient := ctx.dynamicClient.Resource(gvr).Namespace(resource.GetNamespace())

	err = ctx.preProcessResource(resource)
	if err != nil {
		ctx.logger.Errorf("unable to preprocess resource %s. %s", resource.GetName(), err)
		return err
	}

	existingResource, err := resourceClient.Get(context.TODO(), resource.GetName(), metav1.GetOptions{})
	if err == nil && existingResource != nil {
		// resource already exist. ignore creating
		ctx.logger.Warningf("ignoring %s.%s/%s restore. Resource already exist",
			existingResource.GetKind(),
			existingResource.GetAPIVersion(),
			existingResource.GetName())
		return nil
	}

	// init container hook
	if resource.GetKind() == utils.Pod {
		ctx.logger.Infof("Start processing init hook for Pod resource (%s)", resource.GetName())
		hookHandler := hooks.InitHooksHandler{}
		var initRes *unstructured.Unstructured
		// if hookHandler.IsHooksSpecified(restore.Spec.Hooks.Resources, hooks.PreHookPhase) {
		initRes, err = hookHandler.HandleInitHook(ctx.logger, hooks.Pods, resource, &restore.Spec.Hooks)
		if err != nil {
			ctx.logger.Errorf("Failed to process init hook for Pod resource (%s)", resource.GetName())
		}
		if initRes != nil {
			resource = initRes
		}
		// }
		// ctx.logger.Infof("-------Container after init: %+v", initRes)
		_, err = resourceClient.Create(context.TODO(), resource, metav1.CreateOptions{})
		if err != nil && apierrors.IsAlreadyExists(err) {
			// ignore if already exist
			return nil
		}
		// ctx.logger.Infof("-------Container after create: %+v", resource)
		// post hook
		// if hookHandler.IsHooksSpecified(restore.Spec.Hooks.Resources, hooks.PostHookPhase) {
		ctx.waitExec(restore.Spec.Hooks.Resources, resource)
		// }

	} else {
		_, err = resourceClient.Create(context.TODO(), resource, metav1.CreateOptions{})
		if err != nil && apierrors.IsAlreadyExists(err) {
			// ignore if already exist
			return nil
		}
		// post hook
		return err

	}
	return nil
}

func (ctx *restoreContext) preProcessResource(resource *unstructured.Unstructured) error {
	// ensure namespace existence
	if err := ctx.ensureNamespace(resource); err != nil {
		return err
	}
	// remove resource version
	resource.SetResourceVersion("")

	// TODO(Amit Roushan): Add resource specific handling
	return nil
}

func (ctx *restoreContext) ensureNamespace(resource *unstructured.Unstructured) error {
	// check if namespace exist
	namespace := resource.GetNamespace()
	if namespace == "" {
		// ignore if namespace is empty
		// possibly cluster scope resource
		return nil
	}

	// check if namespace exist
	n, err := ctx.kubeClient.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err == nil && n != nil {
		// namespace already exist
		return nil
	}

	// create namespace
	_, err = ctx.kubeClient.CoreV1().
		Namespaces().
		Create(context.TODO(), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}, metav1.CreateOptions{})
	if err != nil {
		ctx.logger.Errorf("Unable to ensure namespace. %s", err)
		return err
	}

	return nil
}

// waitExec executes hooks in a restored pod's containers when they become ready.
func (ctx *restoreContext) waitExec(resourceHooks []kahuapi.RestoreResourceHookSpec, createdObj *unstructured.Unstructured) {
	ctx.hooksWaitGroup.Add(1)
	go func() {
		// Done() will only be called after all errors have been successfully sent
		// on the ctx.resticErrs channel.
		defer ctx.hooksWaitGroup.Done()

		pod := new(corev1.Pod)
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(createdObj.UnstructuredContent(), &pod); err != nil {
			ctx.logger.WithError(err).Error("error converting unstructured pod")
			ctx.hooksErrs <- err
			return
		}
		execHooksByContainer, err := hooks.GroupRestoreExecHooks(
			resourceHooks,
			pod,
			ctx.logger,
		)
		if err != nil {
			ctx.logger.WithError(err).Errorf("error getting exec hooks for pod %s/%s", pod.Namespace, pod.Name)
			ctx.hooksErrs <- err
			return
		}

		if errs := ctx.waitExecHookHandler.HandleHooks(ctx.hooksContext, ctx.logger, pod, execHooksByContainer); len(errs) > 0 {
			// ctx.logger.WithError(kubeerrs.NewAggregate(errs)).Error("unable to successfully execute post-restore hooks")
			ctx.hooksCancelFunc()

			for _, err := range errs {
				// Errors are already logged in the HandleHooks method.
				ctx.hooksErrs <- err
			}
		}
	}()
}
