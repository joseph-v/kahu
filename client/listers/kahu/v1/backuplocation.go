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

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/soda-cdm/kahu/apis/kahu/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// BackupLocationLister helps list BackupLocations.
// All objects returned here must be treated as read-only.
type BackupLocationLister interface {
	// List lists all BackupLocations in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.BackupLocation, err error)
	// Get retrieves the BackupLocation from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.BackupLocation, error)
	BackupLocationListerExpansion
}

// backupLocationLister implements the BackupLocationLister interface.
type backupLocationLister struct {
	indexer cache.Indexer
}

// NewBackupLocationLister returns a new BackupLocationLister.
func NewBackupLocationLister(indexer cache.Indexer) BackupLocationLister {
	return &backupLocationLister{indexer: indexer}
}

// List lists all BackupLocations in the indexer.
func (s *backupLocationLister) List(selector labels.Selector) (ret []*v1.BackupLocation, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.BackupLocation))
	})
	return ret, err
}

// Get retrieves the BackupLocation from the index for a given name.
func (s *backupLocationLister) Get(name string) (*v1.BackupLocation, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("backuplocation"), name)
	}
	return obj.(*v1.BackupLocation), nil
}