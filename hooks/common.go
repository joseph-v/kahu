/*
Copyright 2022 The SODA Authors.
Copyright 2020 the Velero contributors.

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

// Package hooks implement hook execution in backup and restore scenario
package hooks

import (
	"encoding/json"

	"github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"

	kahuapi "github.com/soda-cdm/kahu/apis/kahu/v1"
	"github.com/soda-cdm/kahu/utils"
)

type commonHookSpec struct {
	Name              string
	IncludeNamespaces []string
	ExcludeNamespaces []string
	IncludeResources  []kahuapi.ResourceSpec
	ExcludeResources  []kahuapi.ResourceSpec
	LabelSelector     *metav1.LabelSelector
}

func filterHookNamespaces(allNamespaces sets.String, IncludeNamespaces, ExcludeNamespaces []string) sets.String {
	// Filter namespaces for hook
	hooksNsIncludes := sets.NewString()
	hooksNsExcludes := sets.NewString()
	hooksNsIncludes.Insert(IncludeNamespaces...)
	hooksNsExcludes.Insert(ExcludeNamespaces...)
	filteredHookNs := filterIncludesExcludes(allNamespaces,
		hooksNsIncludes.UnsortedList(),
		hooksNsExcludes.UnsortedList())
	return filteredHookNs
}

func filterIncludesExcludes(rawItems sets.String, includes, excludes []string) sets.String {
	// perform include/exclude item on cache
	excludeItems := sets.NewString()

	// process include item
	includeItems := sets.NewString(includes...)
	if len(includeItems) != 0 {
		for _, item := range rawItems.UnsortedList() {
			// if available item are not included exclude them
			if !includeItems.Has(item) {
				excludeItems.Insert(item)
			}
		}
		for _, item := range includeItems.UnsortedList() {
			if !rawItems.Has(item) {
				excludeItems.Insert(item)
			}
		}
	} else {
		// if not specified include all namespaces
		includeItems = rawItems
	}
	excludeItems.Insert(excludes...)

	for _, excludeItem := range excludeItems.UnsortedList() {
		includeItems.Delete(excludeItem)
	}
	return includeItems
}

func checkInclude(in []string, ex []string, key string) bool {
	setIn := sets.NewString().Insert(in...)
	setEx := sets.NewString().Insert(ex...)
	if setEx.Has(key) {
		return false
	}
	return len(in) == 0 || setIn.Has(key)
}

func validateHook(log logrus.FieldLogger,
	hookSpec commonHookSpec,
	name, namespace string,
	slabels labels.Set) bool {
	// Check namespace
	namespacesIn := hookSpec.IncludeNamespaces
	namespacesEx := hookSpec.ExcludeNamespaces

	if !checkInclude(namespacesIn, namespacesEx, namespace) {
		log.Infof("invalid namespace (%s), skipping  hook execution", namespace)
		return false
	}
	// Check resource
	resourcesIn := hookSpec.IncludeResources
	resourcesEx := hookSpec.ExcludeResources

	var resourceNames []string
	resourceNames = append(resourceNames, name)
	names := utils.FindMatchedStrings(PodResource, resourceNames, resourcesIn, resourcesEx)
	setNames := sets.NewString().Insert(names...)
	if !setNames.Has(name) {
		log.Infof("invalid resource (%s), skipping hook execution", name)
		return false
	}
	// Check label
	var selector labels.Selector
	if hookSpec.LabelSelector != nil {
		labelSelector, err := metav1.LabelSelectorAsSelector(hookSpec.LabelSelector)
		if err != nil {
			log.Infof("error getting label selector(%s), skipping hook execution", err.Error())
			return false
		}
		selector = labelSelector
	}
	if selector != nil && !selector.Matches(slabels) {
		log.Infof("invalid label for hook execution, skipping hook execution")
		return false
	}
	return true
}

func parseStringToCommand(commandValue string) []string {
	var command []string
	// check for json array
	if commandValue[0] == '[' {
		if err := json.Unmarshal([]byte(commandValue), &command); err != nil {
			command = []string{commandValue}
		}
	} else {
		command = append(command, commandValue)
	}
	return command
}
