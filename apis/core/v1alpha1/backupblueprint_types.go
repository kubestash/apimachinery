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

//+kubebuilder:object:root=true

// BackupBlueprint is the Schema for the backupblueprints API
type BackupBlueprint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec BackupBlueprintSpec `json:"spec,omitempty"`
}

// BackupBlueprintSpec defines the desired state of BackupBlueprint
type BackupBlueprintSpec struct {
	Backends []BackendReference `json:"backends,omitempty"`
	Sessions []Session          `json:"sessions,omitempty"`
}

//+kubebuilder:object:root=true

// BackupBlueprintList contains a list of BackupBlueprint
type BackupBlueprintList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupBlueprint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackupBlueprint{}, &BackupBlueprintList{})
}
