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
	"stash.appscode.dev/kubestash/apis"
)

// BackupBatchSpec defines the desired state of BackupBatch
type BackupBatchSpec struct {
	Backends []BackendReference `json:"backends,omitempty"`
	Targets  []TargetReference  `json:"targets,omitempty"`
	Sessions []BatchSession     `json:"sessions,omitempty"`
}

type TargetReference struct {
	Name string                          `json:"name,omitempty"`
	Ref  *core.TypedLocalObjectReference `json:"ref,omitempty"`
}

type BatchSession struct {
	Name                   string                 `json:"name,omitempty"`
	Scheduler              SchedulerSpec          `json:"scheduler,omitempty"`
	Targets                []TargetBackupSpec     `json:"targets,omitempty"`
	RetentionPolicy        kmapi.ObjectReference  `json:"retentionPolicy,omitempty"`
	VerificationStrategies []VerificationStrategy `json:"verificationStrategies,omitempty"`
	FailurePolicy          apis.FailurePolicy     `json:"failurePolicy,omitempty"`
	RetryConfig            *apis.RetryConfig      `json:"retryConfig,omitempty"`
	SessionHistoryLimit    *int32                 `json:"sessionHistoryLimit,omitempty"`
}

type TargetBackupSpec struct {
	Name         string           `json:"name,omitempty"`
	Addon        AddonInfo        `json:"addon,omitempty"`
	Repositories []RepositoryInfo `json:"repositories,omitempty"`
}

// BackupBatchStatus defines the observed state of BackupBatch
type BackupBatchStatus struct {
	Ready         bool                  `json:"ready,omitempty"`
	Backends      []BackendStatus       `json:"backends,omitempty"`
	Targets       []ResourceFoundStatus `json:"targets,omitempty"`
	Addons        []ResourceFoundStatus `json:"addons,omitempty"`
	Repositories  []RepoStatus          `json:"repositories,omitempty"`
	Verifiers     []ResourceFoundStatus `json:"verifiers,omitempty"`
	HookTemplates []ResourceFoundStatus `json:"hookTemplates,omitempty"`
	Sessions      []SessionStatus       `json:"sessions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// BackupBatch is the Schema for the backupbatches API
type BackupBatch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupBatchSpec   `json:"spec,omitempty"`
	Status BackupBatchStatus `json:"status,omitempty"`
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
