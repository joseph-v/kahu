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
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"

	kahuapi "github.com/soda-cdm/kahu/apis/kahu/v1beta1"
	"github.com/soda-cdm/kahu/utils"
)

type mutationHandler interface {
	handle(restore *kahuapi.Restore) error
}

func constructMutationHandler(cache cache.Indexer, logger log.FieldLogger) mutationHandler {
	return newNamespaceMutator(cache,
		newServiceMutation(cache, logger))
}

type namespaceMutation struct {
	cache cache.Indexer
	next  mutationHandler
}

func newNamespaceMutator(cache cache.Indexer, next mutationHandler) mutationHandler {
	return &namespaceMutation{
		cache: cache,
		next:  next,
	}
}

func (handler *namespaceMutation) handle(restore *kahuapi.Restore) error {
	// perform namespace mutation
	for oldNamespace, newNamespace := range restore.Spec.NamespaceMapping {
		resourceList, err := handler.cache.ByIndex(backupObjectNamespaceIndex, oldNamespace)
		if err != nil {
			return fmt.Errorf("failed to retrieve resources from cache for "+
				"namespace mutation. %s", err)
		}

		for _, resource := range resourceList {
			object, ok := resource.(*unstructured.Unstructured)
			if !ok {
				return fmt.Errorf("restore index cache with invalid object type %v",
					reflect.TypeOf(resource))
			}

			newObject := object.DeepCopy()
			newObject.SetNamespace(newNamespace)

			// delete old cached object
			err := handler.cache.Delete(resource)
			if err != nil {
				return fmt.Errorf("failed to delete resource from cache for "+
					"namespace resource mutation. %s", err)
			}

			// add new object in cache
			err = handler.cache.Add(newObject)
			if err != nil {
				return fmt.Errorf("failed to add resource in cache for "+
					"namespace resource mutation. %s", err)
			}
		}
	}

	return handler.next.handle(restore)
}

type serviceMutation struct {
	cache  cache.Indexer
	logger log.FieldLogger
}

func newServiceMutation(cache cache.Indexer, logger log.FieldLogger) mutationHandler {
	return &serviceMutation{
		cache:  cache,
		logger: logger,
	}
}

func (handler *serviceMutation) handle(_ *kahuapi.Restore) error {
	// perform service IP
	resourceList, err := handler.cache.ByIndex(backupObjectResourceIndex, utils.Service)
	if err != nil {
		return fmt.Errorf("failed to retrieve resources from cache for "+
			"namespace mutation. %s", err)
	}

	for _, resource := range resourceList {
		object, ok := resource.(*unstructured.Unstructured)
		if !ok {
			return fmt.Errorf("restore index cache with invalid object type %v",
				reflect.TypeOf(resource))
		}

		newObject := object.DeepCopy()
		err := unstructured.SetNestedField(newObject.Object, "", "spec", "clusterIP")
		if err != nil {
			return fmt.Errorf("failed to unset cluster IP in service mutation. %s", err)
		}

		// delete old cached object
		err = handler.cache.Delete(resource)
		if err != nil {
			return fmt.Errorf("failed to delete resource from cache for "+
				"service mutation. %s", err)
		}

		// add new object in cache
		err = handler.cache.Add(newObject)
		if err != nil {
			return fmt.Errorf("failed to add resource in cache for "+
				"service mutation. %s", err)
		}
	}

	return nil
}
