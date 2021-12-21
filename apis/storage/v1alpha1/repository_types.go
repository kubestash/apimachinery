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
	"stash.appscode.dev/kubestash/apis"
)

// +kubebuilder:object:root=true
// +kubebuilder:object:generate=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=repositories,singular=repository,shortName=repo,categories={kubestash,appscode,all}
// +kubebuilder:printcolumn:name="App",type="string",JSONPath=".spec.appRef.Kind/.spec.appRef.name"
// +kubebuilder:printcolumn:name="BackupStorage",type="string",JSONPath=".spec.storageRef.namespace/.spec.storageRef.name"
// +kubebuilder:printcolumn:name="Integrity",type="boolean",JSONPath=".status.integrity"
// +kubebuilder:printcolumn:name="Snapshot-Count",type="integer",JSONPath=".status.snapshotCount"
// +kubebuilder:printcolumn:name="Size",type="string",JSONPath=".status.size"
// +kubebuilder:printcolumn:name="Last-Successful-Backup",type="date",format="date-time",JSONPath=".status.lastBackupTime"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Repository specifies the information about the targeted application that has been backed up
// and the BackupStorage where the backed up data is being stored. It also holds a list of recent
// Snapshots that have been taken in this Repository.
// Repository is a namespaced object. It must be in the same namespace as the targeted application.
type Repository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepositorySpec   `json:"spec,omitempty"`
	Status RepositoryStatus `json:"status,omitempty"`
}

// RepositorySpec specifies the application reference and the BackupStorage reference.It also specifies
// what should be the behavior when a Repository CR is deleted from the cluster.
type RepositorySpec struct {
	// AppRef refers to the application that is being backed up in this Repository.
	AppRef core.TypedLocalObjectReference `json:"appRef,omitempty"`

	// StorageRef refers to the BackupStorage CR which contain the backend information where the backed
	// up data will be stored. The BackupStorage could be in a different namespace. However, the Repository
	// namespace must be allowed to use the BackupStorage.
	StorageRef apis.TypedObjectReference `json:"storageRef,omitempty"`

	// Path represents the directory inside the BackupStorage where this Repository is storing its data
	// This path is relative to the path of BackupStorage.
	Path string `json:"path,omitempty"`

	// DeletionPolicy specifies what to do when you delete a Repository CR.
	// The valid values are:
	// "Delete": This will delete the respective Snapshot CRs from the cluster but keep the backed up data in the remote backend. This is the default behavior.
	// "WipeOut": This will delete the respective Snapshot CRs as well as the backed up data from the backend.
	// +kubebuilder:validation:default=Delete
	// +optional
	DeletionPolicy DeletionPolicy `json:"deletionPolicy"`

	// Paused specifies whether the Repository is paused or not. If the Repository is paused,
	// Stash will not process any further event for the Repository.
	// +optional
	Paused bool `json:"paused,omitempty"`
}

// RepositoryStatus defines the observed state of Repository
type RepositoryStatus struct {
	// LastBackupTime specifies the timestamp when the last successful backup has been taken
	// +optional
	LastBackupTime string `json:"lastBackupTime,omitempty"`

	// Integrity specifies whether the backed up data of this Repository has been corrupted or not
	// +optional
	Integrity *bool `json:"integrity,omitempty"`

	// SnapshotCount specifies the number of current Snapshots stored in this Repository
	// +optional
	SnapshotCount *int32 `json:"snapshotCount,omitempty"`

	// Size specifies the amount of backed up data stored in the Repository
	// +optional
	Size string `json:"size,omitempty"`

	// RecentSnapshots holds a list of recent Snapshot information that has been taken in this Repository
	// +optional
	RecentSnapshots []SnapshotInfo `json:"recentSnapshots,omitempty"`
}

// SnapshotInfo specifies some basic information about the Snapshots stored in this Repository
type SnapshotInfo struct {
	// Name represents the name of the Snapshot
	Name string `json:"name,omitempty"`

	// Phase represents the phase of the Snapshot
	// +optional
	Phase string `json:"phase,omitempty"`

	// Session represents the name of the session that is responsible for this Snapshot
	Session string `json:"session,omitempty"`

	// Size represents the size of the Snapshot
	// +optional
	Size string `json:"size,omitempty"`

	// SnapshotTime represents the time when this Snapshot was taken
	// +optional
	SnapshotTime string `json:"snapshotTime,omitempty"`
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
