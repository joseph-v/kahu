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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1beta1 "github.com/soda-cdm/kahu/apis/kahu/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeVolumeBackupContents implements VolumeBackupContentInterface
type FakeVolumeBackupContents struct {
	Fake *FakeKahuV1beta1
}

var volumebackupcontentsResource = schema.GroupVersionResource{Group: "kahu.io", Version: "v1beta1", Resource: "volumebackupcontents"}

var volumebackupcontentsKind = schema.GroupVersionKind{Group: "kahu.io", Version: "v1beta1", Kind: "VolumeBackupContent"}

// Get takes name of the volumeBackupContent, and returns the corresponding volumeBackupContent object, and an error if there is any.
func (c *FakeVolumeBackupContents) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.VolumeBackupContent, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(volumebackupcontentsResource, name), &v1beta1.VolumeBackupContent{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.VolumeBackupContent), err
}

// List takes label and field selectors, and returns the list of VolumeBackupContents that match those selectors.
func (c *FakeVolumeBackupContents) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.VolumeBackupContentList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(volumebackupcontentsResource, volumebackupcontentsKind, opts), &v1beta1.VolumeBackupContentList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.VolumeBackupContentList{ListMeta: obj.(*v1beta1.VolumeBackupContentList).ListMeta}
	for _, item := range obj.(*v1beta1.VolumeBackupContentList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested volumeBackupContents.
func (c *FakeVolumeBackupContents) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(volumebackupcontentsResource, opts))
}

// Create takes the representation of a volumeBackupContent and creates it.  Returns the server's representation of the volumeBackupContent, and an error, if there is any.
func (c *FakeVolumeBackupContents) Create(ctx context.Context, volumeBackupContent *v1beta1.VolumeBackupContent, opts v1.CreateOptions) (result *v1beta1.VolumeBackupContent, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(volumebackupcontentsResource, volumeBackupContent), &v1beta1.VolumeBackupContent{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.VolumeBackupContent), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeVolumeBackupContents) UpdateStatus(ctx context.Context, volumeBackupContent *v1beta1.VolumeBackupContent, opts v1.UpdateOptions) (*v1beta1.VolumeBackupContent, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(volumebackupcontentsResource, "status", volumeBackupContent), &v1beta1.VolumeBackupContent{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.VolumeBackupContent), err
}

// Delete takes name of the volumeBackupContent and deletes it. Returns an error if one occurs.
func (c *FakeVolumeBackupContents) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(volumebackupcontentsResource, name), &v1beta1.VolumeBackupContent{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVolumeBackupContents) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(volumebackupcontentsResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1beta1.VolumeBackupContentList{})
	return err
}

// Patch applies the patch and returns the patched volumeBackupContent.
func (c *FakeVolumeBackupContents) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.VolumeBackupContent, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(volumebackupcontentsResource, name, pt, data, subresources...), &v1beta1.VolumeBackupContent{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.VolumeBackupContent), err
}