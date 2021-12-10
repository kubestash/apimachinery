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

// BackupSessionSpec defines the desired state of BackupSession
type BackupSessionSpec struct {
	Invoker core.TypedLocalObjectReference `json:"invoker,omitempty"`
	Session string                         `json:"session,omitempty"`
}

// BackupSessionStatus defines the observed state of BackupSession
type BackupSessionStatus struct {
	Phase           BackupSessionPhase           `json:"phase,omitempty"`
	Duration        string                       `json:"duration,omitempty"`
	Snapshots       []SnapshotStatus             `json:"snapshots,omitempty"`
	Hooks           []HookExecutionStatus        `json:"hooks,omitempty"`
	Verifications   []VerificationStatus         `json:"verifications,omitempty"`
	RetentionPolicy []RetentionPolicyApplyStatus `json:"retentionPolicy,omitempty"`
}

type BackupSessionPhase string

const (
	BackupSessionPending BackupSessionPhase = "Pending"
	BackupSessionRunning BackupSessionPhase = "Running"
	BackupSessionFailed  BackupSessionPhase = "Failed"
	BackupSessionSkipped BackupSessionPhase = "Skipped"
)

type SnapshotStatus struct {
	Name       string                     `json:"name,omitempty"`
	Phase      storage.SnapshotPhase      `json:"phase,omitempty"`
	AppRef     *core.LocalObjectReference `json:"appRef,omitempty"`
	Repository string                     `json:"repository,omitempty"`
}

type VerificationStatus struct {
	Name  string                  `json:"name,omitempty"`
	Phase BackupVerificationPhase `json:"phase,omitempty"`
}

type BackupVerificationPhase string

const (
	VerificationSucceeded BackupVerificationPhase = "Succeeded"
	VerificationFailed    BackupVerificationPhase = "Failed"
)

type RetentionPolicyApplyStatus struct {
	Ref   kmapi.ObjectReference     `json:"ref,omitempty"`
	Phase RetentionPolicyApplyPhase `json:"phase,omitempty"`
	Error string                    `json:"error,omitempty"`
}

type RetentionPolicyApplyPhase string

const (
	RetentionPolicyPending       RetentionPolicyApplyPhase = "Pending"
	RetentionPolicyApplied       RetentionPolicyApplyPhase = "Applied"
	RetentionPolicyFailedToApply RetentionPolicyApplyPhase = "FailedToApply"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// BackupSession is the Schema for the backupsessions API
type BackupSession struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupSessionSpec   `json:"spec,omitempty"`
	Status BackupSessionStatus `json:"status,omitempty"`
}

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
