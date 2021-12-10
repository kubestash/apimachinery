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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SnapshotSpec defines the desired state of Snapshot
type SnapshotSpec struct {
	ULID           string                         `json:"ulid,omitempty"`
	Repository     string                         `json:"repository,omitempty"`
	Session        string                         `json:"session,omitempty"`
	Version        string                         `json:"version,omitempty"`
	AppReference   core.TypedLocalObjectReference `json:"appRef,omitempty"`
	DeletionPolicy DeletionPolicy                 `json:"deletionPolicy,omitempty"`
	Paused         bool                           `json:"paused,omitempty"`
}

// SnapshotStatus defines the observed state of Snapshot
type SnapshotStatus struct {
	Phase              SnapshotPhase      `json:"phase,omitempty"`
	VerificationStatus VerificationStatus `json:"verificationStatus,omitempty"`
	SnapshotTime       string             `json:"snapshotTime,omitempty"`
	LastUpdateTime     string             `json:"lastUpdateTime,omitempty"`
	Size               string             `json:"size,omitempty"`
	Integrity          bool               `json:"integrity,omitempty"`
	Components         []ComponentStatus  `json:"components,omitempty"`
	BackupSession      string             `json:"backupSession,omitempty"`
}

type SnapshotPhase string

const (
	SnapshotPending   SnapshotPhase = "Pending"
	SnapshotRunning   SnapshotPhase = "Running"
	SnapshotSucceeded SnapshotPhase = "Succeeded"
	SnapshotFailed    SnapshotPhase = "Failed"
)

type VerificationStatus string

const (
	SnapshotVerified           VerificationStatus = "Verified"
	SnapshotNotVerified        VerificationStatus = "NotVerified"
	SnapshotVerificationFailed VerificationStatus = "VerificationFailed"
)

type ComponentStatus struct {
	Name        string         `json:"name,omitempty"`
	Path        string         `json:"path,omitempty"`
	Phase       ComponentPhase `json:"phase,omitempty"`
	Driver      apis.Driver    `json:"driver,omitempty"`
	ResticStats ResticStats    `json:"resticStats,omitempty"`
}

type ComponentPhase string

const (
	ComponentPhasePending   ComponentPhase = "Pending"
	ComponentPhaseRunning   ComponentPhase = "Running"
	ComponentPhaseSucceeded ComponentPhase = "Succeeded"
	ComponentPhaseFailed    ComponentPhase = "Failed"
)

type ResticStats struct {
	Id        string `json:"id,omitempty"`
	Uploaded  string `json:"uploaded,omitempty"`
	Size      string `json:"size,omitempty"`
	Integrity bool   `json:"integrity,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Snapshot is the Schema for the snapshots API
type Snapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnapshotSpec   `json:"spec,omitempty"`
	Status SnapshotStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SnapshotList contains a list of Snapshot
type SnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Snapshot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Snapshot{}, &SnapshotList{})
}
