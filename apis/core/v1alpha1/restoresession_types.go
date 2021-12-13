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

// RestoreSessionSpec defines the desired state of RestoreSession
type RestoreSessionSpec struct {
	Target        *core.TypedLocalObjectReference `json:"target,omitempty"`
	DataSource    RestoreDataSource               `json:"dataSource,omitempty"`
	Addon         AddonInfo                       `json:"addon,omitempty"`
	Hooks         RestoreHooks                    `json:"hooks,omitempty"`
	FailurePolicy apis.FailurePolicy              `json:"failurePolicy,omitempty"`
	RetryConfig   *apis.RetryConfig               `json:"retryConfig,omitempty"`
}

type RestoreDataSource struct {
	Repository string   `json:"repository,omitempty"`
	Snapshot   string   `json:"snapshot,omitempty"`
	PITR       PITR     `json:"pitr,omitempty"`
	Components []string `json:"components,omitempty"`
}

type PITR struct {
	TargetTime string `json:"targetTime,omitempty"`
	Exclusive  bool   `json:"exclusive,omitempty"`
}

type RestoreHooks struct {
	PreRestore  []HookInfo `json:"preRestore,omitempty"`
	PostRestore []HookInfo `json:"postRestore,omitempty"`
}

// RestoreSessionStatus defines the observed state of RestoreSession
type RestoreSessionStatus struct {
	Phase      RestorePhase             `json:"phase,omitempty"`
	Components []ComponentRestoreStatus `json:"components,omitempty"`
	Hooks      []HookExecutionStatus    `json:"hooks,omitempty"`
	Backup     *BackupPausedStatus      `json:"backup,omitempty"`
	Conditions []kmapi.Condition        `json:"conditions,omitempty"`
}

type RestorePhase string

const (
	RestorePending RestorePhase = "Pending"
	RestoreRunning RestorePhase = "Running"
	RestoreFailed  RestorePhase = "Failed"
	RestoreSkipped RestorePhase = "Skipped"
)

type ComponentRestoreStatus struct {
	Name  string       `json:"name,omitempty"`
	Phase RestorePhase `json:"phase,omitempty"`
}

type BackupPausedStatus struct {
	Paused  bool                           `json:"paused,omitempty"`
	Invoker core.TypedLocalObjectReference `json:"invoker,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RestoreSession is the Schema for the restoresessions API
type RestoreSession struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RestoreSessionSpec   `json:"spec,omitempty"`
	Status RestoreSessionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RestoreSessionList contains a list of RestoreSession
type RestoreSessionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RestoreSession `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RestoreSession{}, &RestoreSessionList{})
}
