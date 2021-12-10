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
	"stash.appscode.dev/kubestash/apis"
)

// RetentionPolicySpec defines the desired state of RetentionPolicy
type RetentionPolicySpec struct {
	Default             bool                          `json:"default,omitempty"`
	MaxRetentionPeriod  string                        `json:"maxRetentionPeriod,omitempty"`
	UsagePolicy         apis.UsagePolicy              `json:"usagePolicy,omitempty"`
	SuccessfulSnapshots SuccessfulSnapshotsKeepPolicy `json:"successfulSnapshots,omitempty"`
	FailedSnapshots     FailedSnapshotsKeepPolicy     `json:"failedSnapshots,omitempty"`
}

type SuccessfulSnapshotsKeepPolicy struct {
	Last    int32 `json:"last,omitempty"`
	Hourly  int32 `json:"hourly,omitempty"`
	Daily   int32 `json:"daily,omitempty"`
	Weekly  int32 `json:"weekly,omitempty"`
	Monthly int32 `json:"monthly,omitempty"`
	Yearly  int32 `json:"yearly,omitempty"`
}

type FailedSnapshotsKeepPolicy struct {
	Last int32 `json:"last,omitempty"`
}

//+kubebuilder:object:root=true

// RetentionPolicy is the Schema for the retentionpolicies API
type RetentionPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RetentionPolicySpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// RetentionPolicyList contains a list of RetentionPolicy
type RetentionPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RetentionPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RetentionPolicy{}, &RetentionPolicyList{})
}
