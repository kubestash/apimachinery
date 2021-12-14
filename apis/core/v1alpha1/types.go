package v1alpha1

import (
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kmapi "kmodules.xyz/client-go/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	"stash.appscode.dev/kubestash/apis"
)

type AddonInfo struct {
	Name                     string                         `json:"name,omitempty"`
	Tasks                    []TaskReference                `json:"tasks,omitempty"`
	ContainerRuntimeSettings *ofst.ContainerRuntimeSettings `json:"containerRuntimeSettings,omitempty"`
	JobTemplate              *ofst.PodTemplateSpec          `json:"jobTemplate,omitempty"`
}

type TaskReference struct {
	Name          string                `json:"name,omitempty"`
	Variables     []core.EnvVar         `json:"variables,omitempty"`
	Params        *runtime.RawExtension `json:"params,omitempty"`
	TargetVolumes *TargetVolumeInfo     `json:"targetVolumes,omitempty"`
	AddonVolumes  []apis.VolumeSource   `json:"addonVolumes,omitempty"`
}

type TargetVolumeInfo struct {
	Volumes      []core.Volume      `json:"volumes,omitempty"`
	VolumeMounts []core.VolumeMount `json:"volumeMounts,omitempty"`
}

type HookInfo struct {
	Name            string                 `json:"name,omitempty"`
	HookTemplate    *kmapi.ObjectReference `json:"hookTemplate,omitempty"`
	Params          *runtime.RawExtension  `json:"params,omitempty"`
	MaxRetry        int32                  `json:"maxRetry,omitempty"`
	ExecutionPolicy HookExecutionPolicy    `json:"executionPolicy,omitempty"`
	Variables       []core.EnvVar          `json:"variables,omitempty"`
	VolumeMounts    []core.VolumeMount     `json:"volumeMounts,omitempty"`
	Volumes         []core.Volume          `json:"volumes,omitempty"`
	RuntimeSettings *ofst.RuntimeSettings  `json:"runtimeSettings,omitempty"`
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
