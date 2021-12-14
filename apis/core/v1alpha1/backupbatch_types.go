/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// BackupBatch specifies the configurations for taking backup of multiple co-related applications.
type BackupBatch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupBatchSpec   `json:"spec,omitempty"`
	Status BackupBatchStatus `json:"status,omitempty"`
}

// BackupBatchSpec defines the targets of backup, the backend where the backed up data will be stored,
// and the session configuration which specifies the when and how to take the backup.
type BackupBatchSpec struct {
	// Backends specifies a list of storage references where the backed up data will be stored.
	// The respective BackupStorages can be in a different namespace than the BackupBatch.
	// However, it must be allowed by the `usagePolicy` of the BackupStorage to refer from this namespace.
	//
	// This field is optional, if you don't provide any backend here, Stash will use the default BackupStorage for the namespace.
	// If a default BackupStorage does not exist in the same namespace, then Stash will look for a default BackupStorage
	// in other namespaces that allows using it from the BackupBatch namespace.
	// +optional
	Backends []BackendReference `json:"backends,omitempty"`

	// Targets specifies a list of targets that are subject to backup.
	Targets []TargetReference `json:"targets,omitempty"`

	// Session defines a list of session configuration that specifies when and how to take backup.
	Sessions []BatchSession `json:"sessions,omitempty"`
}

// TargetReference specifies a reference to the target that is subject to backup
type TargetReference struct {
	// Name specifies an identifier for this target. This name will be used in the session to refer this target.
	Name string `json:"name,omitempty"`

	// AppRef points to the target that is subject to backup. The target should be in same namespace as the BackupBatch.
	AppRef *core.TypedLocalObjectReference `json:"appRef,omitempty"`
}

// BatchSession specifies the session configuration for the targets.
type BatchSession struct {
	*SessionConfig

	// Targets specifies a list of target backup specification.
	Targets []TargetBackupSpec `json:"targets,omitempty"`
}

// TargetBackupSpec specifies the information needed to backup a target.
type TargetBackupSpec struct {
	// Name point to the identifier of the target that is being backed up.
	// It should match the name used as the identifier of a target in the `spec.targets` section.
	Name string `json:"name,omitempty"`

	// Addon specifies addon configuration that will be used to backup this target.
	Addon *AddonInfo `json:"addon,omitempty"`

	// Repositories specifies a list of repository information where the backed up data will be stored.
	// Stash will create the respective Repository CRs using this information.
	Repositories []RepositoryInfo `json:"repositories,omitempty"`
}

// BackupBatchStatus defines the observed state of BackupBatch
type BackupBatchStatus struct {
	*OffshootStatus

	// Targets specifies whether the targets of backup do exist or not
	Targets []ResourceFoundStatus `json:"targets,omitempty"`
}

//+kubebuilder:object:root=true

// BackupBatchList contains a list of BackupBatch
type BackupBatchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupBatch `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackupBatch{}, &BackupBatchList{})
}
