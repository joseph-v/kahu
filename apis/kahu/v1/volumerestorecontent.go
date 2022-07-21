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

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VolumeRestoreContentSpec defines the desired state of VolumeRestoreContent
type VolumeRestoreContentSpec struct {
	// BackupName is backup CR name specified during backup
	// +required
	BackupName string `json:"backupName"`

	// Volume represents kubernetes volume to be restored
	// +optional
	// +nullable
	Volumes []RestoreVolumeSpec `json:"volumes"`

	// Volume provider for set of volumes
	VolumeProvider *string `json:"volumeProvider,omitempty"`

	// Supported volume backup provider information
	// +required
	BackupProviderLocation string `json:"backupProviderLocation"`
}

type RestoreVolumeSpec struct {
	// +required
	BackupHandle string `json:"backupHandle"`

	// +required
	Claim v1.PersistentVolumeClaimSpec `json:"claim"`
}

// +kubebuilder:validation:Enum=New;InProgress;Completed;Failed;Deleting

type VolumeRestoreContentPhase string

const (
	VolumeRestoreContentPhaseInit       VolumeRestoreContentPhase = "New"
	VolumeRestoreContentPhaseInProgress VolumeRestoreContentPhase = "InProgress"
	VolumeRestoreContentPhaseCompleted  VolumeRestoreContentPhase = "Completed"
	VolumeRestoreContentPhaseFailed     VolumeRestoreContentPhase = "Failed"
	VolumeRestoreContentPhaseDeleting   VolumeRestoreContentPhase = "Deleting"
)

type VolumeRestoreState struct {
	VolumeRef *v1.ObjectReference `json:"volumeRef,omitempty"`

	VolumeHandle string `json:"backupHandle,omitempty"`
}

// VolumeRestoreContentStatus defines the observed state of VolumeRestoreContent
type VolumeRestoreContentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	Phase VolumeRestoreContentPhase `json:"phase,omitempty"`

	// +optional
	// +nullable
	StartTimestamp *metav1.Time `json:"startTimestamp,omitempty"`

	// +optional
	// +nullable
	CompletionTimestamp *metav1.Time `json:"completionTimestamp,omitempty"`

	// +optional
	FailureReason string `json:"failureReason,omitempty"`

	// +optional
	RestoreState []VolumeRestoreState `json:"backupState,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +genclient:skipVerbs=update
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName="vrc"

// VolumeRestoreContent is the Schema for the VolumeRestoreContents API
type VolumeRestoreContent struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec VolumeRestoreContentSpec `json:"spec,omitempty"`
	// +optional
	Status VolumeRestoreContentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeRestoreContentList contains a list of VolumeRestoreContent
type VolumeRestoreContentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumeRestoreContent `json:"items"`
}
