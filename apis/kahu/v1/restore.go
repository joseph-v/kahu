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

type ResourceSpec struct {
	// +required
	// Name of the resource
	// The name can have empty, * in regular expression
	// or valid resource name
	Name string `json:"name"`

	// +required
	// Kind of the resource
	Kind string `json:"kind"`

	// +optional
	// IsRegex indicates if Name is regular expression
	IsRegex bool `json:"isRegex,omitempty"`
}

// RestoreSpec defines the desired state of Restore
type RestoreSpec struct {
	// BackupName is backup CR name specified during backup
	// +required
	BackupName string `json:"backupName"`

	// IncludeNamespaces are set of namespaces name considered for restore
	// +optional
	IncludeNamespaces []string `json:"includeNamespaces,omitempty"`

	// ExcludeNamespaces are set of namespace name should not get considered for restore
	// +optional
	ExcludeNamespaces []string `json:"excludeNamespaces,omitempty"`

	// IncludeResources are set of kubernetes resource name considered for restore
	// +optional
	IncludeResources []ResourceSpec `json:"includeResources,omitempty"`

	// ExcludeResources are set of kubernetes resource name should not get considered for restore
	// +optional
	ExcludeResources []ResourceSpec `json:"excludeResources,omitempty"`

	// LabelSelector are label get evaluated against resource selection
	// +optional
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`

	// NamespaceMapping is mapping between backed up namespace name against restore namespace name
	// +optional
	NamespaceMapping map[string]string `json:"namespaceMapping,omitempty"`

	// IncludeClusterResource is a flag for considering cluster wider resource during restore
	// +optional
	IncludeClusterResource bool `json:"includeClusterResource,omitempty"`

	// ResourcePrefix gets prepended in each restored resource name
	// +optional
	ResourcePrefix string `json:"resourcePrefix,omitempty"`

	// Hooks represent custom behaviors that should be executed during or post restore.
	// +optional
	Hooks RestoreHookSpec `json:"hook,omitempty"`
}

// HookSpec is hook which should be executed
// at different phase of backup
type RestoreHookSpec struct {
	// +optional
	Resources []RestoreResourceHookSpec `json:"resources,omitempty"`
}

// ResourceHookSpec is hook which should be executed
// at different phase of backup
type RestoreResourceHookSpec struct {
	// +optional
	Name string `json:"name"`

	// IncludeNamespaces is a list of all namespaces included for hook. If empty, all namespaces
	// are included
	// +optional
	IncludeNamespaces []string `json:"includeNamespaces,omitempty"`

	// ExcludeNamespaces is a list of all namespaces excluded for hook
	// +optional
	ExcludeNamespaces []string `json:"excludeNamespaces,omitempty"`

	// IncludeResources is a list of all resources included for hook. If empty, all resources
	// are included
	// +optional
	IncludeResources []ResourceSpec `json:"includeResources,omitempty"`

	// ExcludeResources is a list of all resources excluded for backup
	// +optional
	ExcludeResources []ResourceSpec `json:"excludeResources,omitempty"`

	// Label is used to filter the resources
	// +optional
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`

	// PostHooks is a list of ResourceHooks to execute after storing the item in the backup.
	// These are executed after all "additional items" from item actions are processed.
	// +optional
	PostHooks []RestoreResourceHook `json:"post,omitempty"`
}

// ResourceHook defines a hook for a resource.
type RestoreResourceHook struct {
	// Exec defines an exec hook.
	// +optional
	Exec *RestoreExecHook `json:"exec"`

	// Init defines an init restore hook.
	// +optional
	Init *InitRestoreHook `json:"init,omitempty"`
}

// ExecHook is a hook that uses the pod exec API to execute a command in a container in a pod.
type RestoreExecHook struct {
	// Container is the container in the pod where the command should be executed. If not specified,
	// the pod's first container is used.
	// +optional
	Container string `json:"container,omitempty"`

	// Command is the command and arguments to execute.
	// +kubebuilder:validation:MinItems=1
	Command []string `json:"command"`

	// OnError specifies how to behave if it encounters an error executing this hook.
	// +optional
	OnError HookErrorMode `json:"onError,omitempty"`

	// Timeout defines the maximum amount of time service should wait for the hook to complete before
	// considering the execution a failure.
	// +optional
	Timeout metav1.Duration `json:"timeout,omitempty"`

	// WaitTimeout defines the maximum amount of time Velero should wait for the container to be Ready
	// before attempting to run the command.
	// +optional
	WaitTimeout metav1.Duration `json:"waitTimeout,omitempty"`
}

// InitRestoreHook is a hook that adds an init container to a PodSpec to run commands before the
// workload pod is able to start.
type InitRestoreHook struct {
	// InitContainers is list of init containers to be added to a pod during its restore.
	// +optional
	InitContainers []v1.Container `json:"initContainers"`

	// Timeout defines the maximum amount of time Velero should wait for the initContainers to complete.
	// +optional
	Timeout metav1.Duration `json:"timeout,omitempty"`
}

// // HookErrorMode defines how service should treat an error from a hook.
// // +kubebuilder:validation:Enum=Continue;Fail
// type HookErrorMode string

// const (
// 	// HookErrorModeContinue means that an error from a hook is acceptable, and the backup can
// 	// proceed.
// 	HookErrorModeContinue HookErrorMode = "Continue"

// 	// HookErrorModeFail means that an error from a hook is problematic, and the backup should be in
// 	// error.
// 	HookErrorModeFail HookErrorMode = "Fail"
// )

// +kubebuilder:validation:Enum=Initial;PreHook;Resources;Volumes;PostHook;Finished

type RestoreStage string

// +kubebuilder:validation:Enum=New;Validating;Failed;Processing;Completed;Deleting

type RestoreState string

const (
	RestoreStageInitial   RestoreStage = "Initial"
	RestoreStagePreHook   RestoreStage = "PreHook"
	RestoreStageResources RestoreStage = "Resources"
	RestoreStageVolumes   RestoreStage = "Volumes"
	RestoreStagePostHook  RestoreStage = "PostHook"
	RestoreStageFinished  RestoreStage = "Finished"

	RestoreStateNew        RestoreState = "New"
	RestoreStateValidating RestoreState = "Validating"
	RestoreStateFailed     RestoreState = "Failed"
	RestoreStateProcessing RestoreState = "Processing"
	RestoreStateCompleted  RestoreState = "Completed"
	RestoreStateDeleting   RestoreState = "Deleting"
)

// RestoreProgress expresses overall progress of restore
type RestoreProgress struct {
	// TotalItems is count of resource to be process
	// +optional
	TotalItems int `json:"totalItems,omitempty"`

	// ItemsRestored is count of resource got processed
	// +optional
	ItemsRestored int `json:"itemsRestored,omitempty"`
}

// RestoreStatus defines the observed state of Restore
type RestoreStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	Stage RestoreStage `json:"stage,omitempty"`

	// +optional
	// +kubebuilder:default=New
	State RestoreState `json:"state,omitempty"`

	// +optional
	// +nullable
	StartTimestamp *metav1.Time `json:"startTimestamp,omitempty"`

	// +optional
	// +nullable
	CompletionTimestamp *metav1.Time `json:"completionTimestamp,omitempty"`

	// +optional
	Progress RestoreProgress `json:"progress,omitempty"`

	// +optional
	FailureReason string `json:"failureReason,omitempty"`

	// ValidationErrors is a slice of validation errors during restore
	// +optional
	// +nullable
	ValidationErrors []string `json:"validationErrors,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +genclient:skipVerbs=update,patch
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="BackupName",type=string,JSONPath=`.spec.backupName`
// +kubebuilder:printcolumn:name="Stage",type=string,JSONPath=`.status.stage`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="StartTimestamp",type=string,JSONPath=`.status.startTimestamp`
// +kubebuilder:printcolumn:name="CompletionTimestamp",type=string,JSONPath=`.status.completionTimestamp`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Restore is the Schema for the restores API
type Restore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec RestoreSpec `json:"spec,omitempty"`
	// +optional
	Status RestoreStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RestoreList contains a list of Restore
type RestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Restore `json:"items"`
}
