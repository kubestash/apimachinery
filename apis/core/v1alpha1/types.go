package v1alpha1

import (
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kmapi "kmodules.xyz/client-go/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	"stash.appscode.dev/kubestash/apis"
)

// AddonInfo specifies addon configuration that will be used to backup/restore the respective target.
type AddonInfo struct {
	// Name specifies the name of the addon that will be used for the backup/restore purpose
	Name string `json:"name,omitempty"`

	// Tasks specifies a list of backup/restore tasks and their configuration parameters
	Tasks []TaskReference `json:"tasks,omitempty"`

	// ContainerRuntimeSettings specifies runtime settings for the backup/restore executor container
	// +optional
	ContainerRuntimeSettings *ofst.ContainerRuntimeSettings `json:"containerRuntimeSettings,omitempty"`

	// JobTemplate specifies runtime configurations for the bakcup/restore Job
	// +optional
	JobTemplate *ofst.PodTemplateSpec `json:"jobTemplate,omitempty"`
}

// TaskReference specifies a task and its configuration parameters
type TaskReference struct {
	// Name indicates to the name of the task
	Name string `json:"name,omitempty"`

	// Variables specifies a list of variables and their sources that will be used to resolve the task.
	// +optional
	Variables []core.EnvVar `json:"variables,omitempty"`

	// Params specifies parameters for the task. You must provide the parameter in the Addon desired structure.
	// +optional
	Params *runtime.RawExtension `json:"params,omitempty"`

	// TargetVolumes specifies which volumes from the target should be mounted in the backup/restore job/container.
	// +optional
	TargetVolumes *TargetVolumeInfo `json:"targetVolumes,omitempty"`

	// AddonVolumes let you overwrite the volume sources used in the VolumeTemplate section of Addon.
	// Make sure that name of your volume matches with the name of the volume you want to overwrite.
	// +optional
	AddonVolumes []apis.VolumeSource `json:"addonVolumes,omitempty"`
}

// TargetVolumeInfo specifies the volumes and their mounts of the targeted application that should
// be mounted in backup/restore Job/container.
type TargetVolumeInfo struct {
	// Volumes indicates the list of volumes of targeted application that should be mounted on the backup/restore job.
	Volumes []core.Volume `json:"volumes,omitempty"`

	// VolumeMounts specifies the mount for the volumes specified in `Volumes` section
	VolumeMounts []core.VolumeMount `json:"volumeMounts,omitempty"`
}

// HookInfo specifies the information about the backup/restore hooks
type HookInfo struct {
	// Name specifies a name for the hook
	Name string `json:"name,omitempty"`

	// HookTemplate points to a HookTemplate CR that will be used to execute the hook.
	// You can refer to a HookTemplate from other namespaces as long as your current
	// namespace is allowed by the `usagePolicy` in the respective HookTemplate.
	HookTemplate *kmapi.ObjectReference `json:"hookTemplate,omitempty"`

	// Params specifies parameters for the hook. You must provide the parameter in the HookTemplates desired structure.
	// +optional
	Params *runtime.RawExtension `json:"params,omitempty"`

	// MaxRetry specified how many times Stash should retry the hook execution in case of failure.
	// The default value of this field is 0 which means no retry.
	// +optional
	MaxRetry int32 `json:"maxRetry,omitempty"`

	// ExecutionPolicy specifies when to execute the hook.
	// Valid values are:
	// - "Always": Stash will execute this hook no matter the backup/restore failed. This is the default execution policy.
	// - "OnSuccess": Stash will execute this hook only if the backup/restore has succeeded.
	// - "OnFailure": Stash will execute this hook only if the backup/restore has failed.
	// +kubebuilder:validation:default=Always
	// +optional
	ExecutionPolicy HookExecutionPolicy `json:"executionPolicy,omitempty"`

	// Variables specifies a list of variables and their sources that will be used to resolve the HookTemplate.
	// +optional
	Variables []core.EnvVar `json:"variables,omitempty"`

	// Volumes indicates the list of volumes of targeted application that should be mounted on the hook executor.
	// Use this field only for `Function` type hook executor.
	// +optional
	VolumeMounts []core.VolumeMount `json:"volumeMounts,omitempty"`

	// VolumeMounts specifies the mount for the volumes specified in `Volumes` section
	// Use this field only for `Function` type hook executor.
	// +optional
	Volumes []core.Volume `json:"volumes,omitempty"`

	// RuntimeSettings specifies runtime configurations for the hook executor Job.
	// Use this field only for `Function` type hook executor.
	// +optional
	RuntimeSettings *ofst.RuntimeSettings `json:"runtimeSettings,omitempty"`
}

// HookExecutionPolicy specifies when to execute the hook.
// +kubebuilder:validation:Enum=Always;OnSuccess;OnFailure
type HookExecutionPolicy string

const (
	ExecuteAlways    HookExecutionPolicy = "Always"
	ExecuteOnSuccess HookExecutionPolicy = "OnSuccess"
	ExecuteOnFailure HookExecutionPolicy = "OnFailure"
)

// HookExecutionStatus represents the state of the hook execution
type HookExecutionStatus struct {
	// Name indicate the name of the hook whose status is being shown here.
	Name string `json:"name,omitempty"`

	// Phase represents the hook execution phase
	// +optional
	Phase HookExecutionPhase `json:"phase,omitempty"`
}

// HookExecutionPhase specifies the state of the hook execution
// +kubebuilder:validation:Enum=Succeeded;Failed
type HookExecutionPhase string

const (
	HookExecutionSucceeded HookExecutionPhase = "Succeeded"
	HookExecutionFailed    HookExecutionPhase = "Failed"
)

// ResourceFoundStatus specifies whether a resource was found or not
type ResourceFoundStatus struct {
	// Name represents the name of the resource
	Name string `json:"name,omitempty"`

	// Found indicates whether the resource was found or not
	Found bool `json:"found,omitempty"`
}
