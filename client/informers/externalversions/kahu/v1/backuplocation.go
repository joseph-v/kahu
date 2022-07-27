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

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	"context"
	time "time"

	kahuv1 "github.com/soda-cdm/kahu/apis/kahu/v1"
	versioned "github.com/soda-cdm/kahu/client/clientset/versioned"
	internalinterfaces "github.com/soda-cdm/kahu/client/informers/externalversions/internalinterfaces"
	v1 "github.com/soda-cdm/kahu/client/listers/kahu/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// BackupLocationInformer provides access to a shared informer and lister for
// BackupLocations.
type BackupLocationInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.BackupLocationLister
}

type backupLocationInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewBackupLocationInformer constructs a new informer for BackupLocation type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewBackupLocationInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredBackupLocationInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredBackupLocationInformer constructs a new informer for BackupLocation type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredBackupLocationInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KahuV1().BackupLocations().List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KahuV1().BackupLocations().Watch(context.TODO(), options)
			},
		},
		&kahuv1.BackupLocation{},
		resyncPeriod,
		indexers,
	)
}

func (f *backupLocationInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredBackupLocationInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *backupLocationInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&kahuv1.BackupLocation{}, f.defaultInformer)
}

func (f *backupLocationInformer) Lister() v1.BackupLocationLister {
	return v1.NewBackupLocationLister(f.Informer().GetIndexer())
}