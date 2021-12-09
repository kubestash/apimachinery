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

// RepositorySpec defines the desired state of Repository
type RepositorySpec struct {
	AppRef           core.TypedLocalObjectReference `json:"appRef,omitempty"`
	BackupStorageRef TypedObjectReference           `json:"backupStorageRef,omitempty"`
	Path             string                         `json:"path,omitempty"`
	DeletionPolicy   DeletionPolicy                 `json:"deletionPolicy"`
	Paused           bool                           `json:"paused,omitempty"`
}

// RepositoryStatus defines the observed state of Repository
type RepositoryStatus struct {
	LastBackupTime  string         `json:"lastBackupTime,omitempty"`
	Integrity       bool           `json:"integrity,omitempty"`
	SnapshotCount   int32          `json:"snapshotCount,omitempty"`
	Size            string         `json:"size,omitempty"`
	RecentSnapshots []SnapshotInfo `json:"recentSnapshots,omitempty"`
}

type SnapshotInfo struct {
	Name         string  `json:"name,omitempty"`
	Phase        string  `json:"phase,omitempty"`
	Session      string  `json:"session,omitempty"`
	Size         string  `json:"size,omitempty"`
	SnapshotTime string  `json:"snapshotTime,omitempty"`
	Error        *string `json:"error,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Repository is the Schema for the repositories API
type Repository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepositorySpec   `json:"spec,omitempty"`
	Status RepositoryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RepositoryList contains a list of Repository
type RepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Repository `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Repository{}, &RepositoryList{})
}
