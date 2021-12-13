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

//+kubebuilder:object:root=true

// Addon is the Schema for the addons API
type Addon struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AddonSpec `json:"spec,omitempty"`
}

// AddonSpec defines the desired state of Addon
type AddonSpec struct {
	BackupTasks  []Task `json:"backupTasks,omitempty"`
	RestoreTasks []Task `json:"restoreTasks,omitempty"`
}

type Task struct {
	Name              string                     `json:"name,omitempty"`
	Function          string                     `json:"function,omitempty"`
	Driver            apis.Driver                `json:"driver,omitempty"`
	Executor          TaskExecutor               `json:"executor,omitempty"`
	Parameters        []apis.ParameterDefinition `json:"parameters,omitempty"`
	VolumeTemplate    []VolumeTemplate           `json:"volumeTemplate,omitempty"`
	VolumeMounts      []core.VolumeMount         `json:"volumeMounts,omitempty"`
	PassThroughMounts []core.VolumeMount         `json:"passThroughMounts,omitempty"`
}

type TaskExecutor string

const (
	ExecutorJob                TaskExecutor = "Job"
	ExecutorSidecar            TaskExecutor = "Sidecar"
	ExecutorEphemeralContainer TaskExecutor = "EphemeralContainer"
	ExecutorMultiLevelJob      TaskExecutor = "MultiLevelJob"
)

type VolumeTemplate struct {
	Name   string             `json:"name,omitempty"`
	Usage  string             `json:"usage,omitempty"`
	Source *apis.VolumeSource `json:"source,omitempty"`
}

//+kubebuilder:object:root=true

// AddonList contains a list of Addon
type AddonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Addon `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Addon{}, &AddonList{})
}
