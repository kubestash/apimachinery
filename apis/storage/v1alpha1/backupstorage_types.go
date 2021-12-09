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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackupStorageSpec defines the desired state of BackupStorage
type BackupStorageSpec struct {
	Storage        Backend        `json:"storage"`
	UsagePolicy    UsagePolicy    `json:"usagePolicy,omitempty"`
	Default        bool           `json:"default,omitempty"`
	DeletionPolicy DeletionPolicy `json:"deletionPolicy,omitempty"`
}

// BackupStorageStatus defines the observed state of BackupStorage
type BackupStorageStatus struct {
	Ready        bool               `json:"ready,omitempty"`
	TotalSize    string             `json:"totalSize,omitempty"`
	Repositories []RepositoryInfo   `json:"repositories,omitempty"`
	Conditions   []metav1.Condition `json:"conditions,omitempty"`
}

type RepositoryInfo struct {
	Name      string  `json:"name,omitempty"`
	Namespace string  `json:"namespace,omitempty"`
	Path      string  `json:"path,omitempty"`
	Size      string  `json:"size,omitempty"`
	Synced    bool    `json:"synced,omitempty"`
	Error     *string `json:"error,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// BackupStorage is the Schema for the backupstorages API
type BackupStorage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupStorageSpec   `json:"spec,omitempty"`
	Status BackupStorageStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BackupStorageList contains a list of BackupStorage
type BackupStorageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupStorage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackupStorage{}, &BackupStorageList{})
}
