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

// BackupBatch is the Schema for the backupbatches API
type BackupBatch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupBatchSpec   `json:"spec,omitempty"`
	Status BackupBatchStatus `json:"status,omitempty"`
}

// BackupBatchSpec defines the desired state of BackupBatch
type BackupBatchSpec struct {
	Backends []BackendReference `json:"backends,omitempty"`
	Targets  []TargetReference  `json:"targets,omitempty"`
	Sessions []BatchSession     `json:"sessions,omitempty"`
}

type TargetReference struct {
	Name   string                          `json:"name,omitempty"`
	AppRef *core.TypedLocalObjectReference `json:"appRef,omitempty"`
}

type BatchSession struct {
	*SessionConfig
	Targets []TargetBackupSpec `json:"targets,omitempty"`
}

type TargetBackupSpec struct {
	Name         string           `json:"name,omitempty"`
	Addon        *AddonInfo       `json:"addon,omitempty"`
	Repositories []RepositoryInfo `json:"repositories,omitempty"`
}

// BackupBatchStatus defines the observed state of BackupBatch
type BackupBatchStatus struct {
	*OffshootStatus
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
