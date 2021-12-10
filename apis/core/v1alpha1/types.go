package v1alpha1

import (
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

type HookInfo struct {
	Name            string                `json:"name,omitempty"`
	HookTemplate    HookReference         `json:"hookTemplate,omitempty"`
	Params          *runtime.RawExtension `json:"params,omitempty"`
	MaxRetry        int32                 `json:"maxRetry,omitempty"`
	ExecutionPolicy HookExecutionPolicy   `json:"executionPolicy,omitempty"`
	Variables       []core.EnvVar         `json:"variables,omitempty"`
	VolumeMounts    []core.VolumeMount    `json:"volumeMounts,omitempty"`
	Volumes         []core.Volume         `json:"volumes,omitempty"`
	RuntimeSettings ofst.RuntimeSettings  `json:"runtimeSettings,omitempty"`
}

type HookReference struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type HookExecutionPolicy string

const (
	ExecuteAlways    HookExecutionPolicy = "Always"
	ExecuteOnSuccess HookExecutionPolicy = "OnSuccess"
	ExecuteOnFailure HookExecutionPolicy = "OnFailure"
)

type HookExecutionStatus struct {
	Name  string             `json:"name,omitempty"`
	Phase HookExecutionPhase `json:"phase,omitempty"`
}

type HookExecutionPhase string

const (
	HookExecutionSucceeded HookExecutionPhase = "Succeeded"
	HookExecutionFailed    HookExecutionPhase = "Failed"
)

type ResourceFoundStatus struct {
	Name  string `json:"name,omitempty"`
	Found bool   `json:"found,omitempty"`
}
