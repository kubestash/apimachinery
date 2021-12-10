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
	prober "kmodules.xyz/prober/api/v1"
	"stash.appscode.dev/kubestash/apis"
)

// HookTemplateSpec defines the desired state of HookTemplate
type HookTemplateSpec struct {
	UsagePolicy apis.UsagePolicy
	Params      apis.ParameterDefinition `json:"params,omitempty"`
	Action      *prober.Handler          `json:"action,omitempty"`
	TimeOut     string                   `json:"timeOut,omitempty"`
	Executor    HookExecutor             `json:"executor,omitempty"`
}

type HookExecutor struct {
	Type     HookExecutorType         `json:"type,omitempty"`
	Function FunctionHookExecutorSpec `json:"function,omitempty"`
	Pod      PodHookExecutorSpec      `json:"pod,omitempty"`
}

type HookExecutorType string

const (
	HookExecutorFunction HookExecutorType = "Function"
	HookExecutorPod      HookExecutorType = "Pod"
	HookExecutorOperator HookExecutorType = "Operator"
)

type FunctionHookExecutorSpec struct {
	Name         string             `json:"name,omitempty"`
	Variables    []core.EnvVar      `json:"variables,omitempty"`
	VolumeMounts []core.VolumeMount `json:"volumeMounts,omitempty"`
	Volumes      []core.Volume      `json:"volumes,omitempty"`
}

type PodHookExecutorSpec struct {
	Selector string                   `json:"selector,omitempty"`
	Owner    *metav1.OwnerReference   `json:"owner,omitempty"`
	Strategy PodHookExecutionStrategy `json:"strategy,omitempty"`
}

type PodHookExecutionStrategy string

const (
	ExecuteOnOnce PodHookExecutionStrategy = "ExecuteOnOnce"
	ExecuteOnAll  PodHookExecutionStrategy = "ExecuteOnAll"
)

//+kubebuilder:object:root=true

// HookTemplate is the Schema for the hooktemplates API
type HookTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HookTemplateSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// HookTemplateList contains a list of HookTemplate
type HookTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HookTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HookTemplate{}, &HookTemplateList{})
}
