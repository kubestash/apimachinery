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
	kmapi "kmodules.xyz/client-go/api/v1"
	storage "stash.appscode.dev/kubestash/apis/storage/v1alpha1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// BackupSession represent one backup run for the target(s) pointed by the
// respective BackupConfiguration or BackupBatch
type BackupSession struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupSessionSpec   `json:"spec,omitempty"`
	Status BackupSessionStatus `json:"status,omitempty"`
}

// BackupSessionSpec specifies the information related to the respective backup invoker and session.
type BackupSessionSpec struct {
	// Invoker points to the respective BackupConfiguration or BackupBatch
	// which is responsible for triggering this backup.
	Invoker *core.TypedLocalObjectReference `json:"invoker,omitempty"`

	// Session specifies the name of the session that triggered this backup
	Session string `json:"session,omitempty"`
}

// BackupSessionStatus defines the observed state of BackupSession
type BackupSessionStatus struct {
	// Phase represents the current state of the backup process.
	// +optional
	Phase BackupSessionPhase `json:"phase,omitempty"`

	// Duration specifies the time required to complete the backup process
	// +optional
	Duration string `json:"duration,omitempty"`

	// Snapshots specifies the Snapshots status
	// +optional
	Snapshots []SnapshotStatus `json:"snapshots,omitempty"`

	// Hooks specifies the hook execution status
	// +optional
	Hooks []HookExecutionStatus `json:"hooks,omitempty"`

	// Verifications specifies the backup verification status
	// +optional
	Verifications []VerificationStatus `json:"verifications,omitempty"`

	// RetentionPolicy specifies whether the retention policy was properly applied or not
	// +optional
	RetentionPolicy *RetentionPolicyApplyStatus `json:"retentionPolicy,omitempty"`
}

// BackupSessionPhase specifies the current state of the backup process
// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed;Skipped
type BackupSessionPhase string

const (
	BackupSessionPending   BackupSessionPhase = "Pending"
	BackupSessionRunning   BackupSessionPhase = "Running"
	BackupSessionSucceeded BackupSessionPhase = "Succeeded"
	BackupSessionFailed    BackupSessionPhase = "Failed"
	BackupSessionSkipped   BackupSessionPhase = "Skipped"
)

// SnapshotStatus represents the current state of respective the Snapshot
type SnapshotStatus struct {
	// Name indicate to the name of the Snapshot
	Name string `json:"name,omitempty"`

	// Phase indicate the phase of the Snapshot
	// +optional
	Phase storage.SnapshotPhase `json:"phase,omitempty"`

	// AppRef indicate to the application that is being backed up in this Snapshot
	AppRef *core.LocalObjectReference `json:"appRef,omitempty"`

	// Repository indicates the name of the Repository where the Snapshot is being stored.
	Repository string `json:"repository,omitempty"`
}

// VerificationStatus specifies the status of a backup verification
type VerificationStatus struct {
	// Name indicates the name of the respective verification strategy
	Name string `json:"name,omitempty"`

	// Phase represents the state of the verification process
	// +optional
	Phase BackupVerificationPhase `json:"phase,omitempty"`
}

// BackupVerificationPhase represents the state of the backup verification process
// +kubebuilder:validation:Enum=Verified;NotVerified;VerificationFailed
type BackupVerificationPhase string

const (
	Verified           BackupVerificationPhase = "Verified"
	NotVerified        BackupVerificationPhase = "NotVerified"
	VerificationFailed BackupVerificationPhase = "VerificationFailed"
)

// RetentionPolicyApplyStatus represents the state of the applying retention policy
type RetentionPolicyApplyStatus struct {
	// Ref points to the RetentionPolicy CR that is being used to cleanup the old Snapshots for this session.
	Ref kmapi.ObjectReference `json:"ref,omitempty"`

	// Phase specifies the state of retention policy apply process
	// +optional
	Phase RetentionPolicyApplyPhase `json:"phase,omitempty"`

	// Error represents the reason if the retention policy applying fail
	// +optional
	Error string `json:"error,omitempty"`
}

// RetentionPolicyApplyPhase represents the state of the retention policy apply process
// +kubebuilder:validation:Enum=Pending;Applied;FailedToApply
type RetentionPolicyApplyPhase string

const (
	RetentionPolicyPending       RetentionPolicyApplyPhase = "Pending"
	RetentionPolicyApplied       RetentionPolicyApplyPhase = "Applied"
	RetentionPolicyFailedToApply RetentionPolicyApplyPhase = "FailedToApply"
)

//+kubebuilder:object:root=true

// BackupSessionList contains a list of BackupSession
type BackupSessionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupSession `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackupSession{}, &BackupSessionList{})
}
