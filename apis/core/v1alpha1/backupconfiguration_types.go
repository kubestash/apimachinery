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
	batchv1 "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kmapi "kmodules.xyz/client-go/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	"stash.appscode.dev/kubestash/apis"
)

// BackupConfigurationSpec defines the desired state of BackupConfiguration
type BackupConfigurationSpec struct {
	Target   *core.TypedLocalObjectReference `json:"target,omitempty"`
	Backends []BackendReference              `json:"backends,omitempty"`
	Sessions []Session                       `json:"sessions,omitempty"`
}

type BackendReference struct {
	Name          string                    `json:"name,omitempty"`
	BackupStorage apis.TypedObjectReference `json:"backupStorage,omitempty"`
}

type Session struct {
	Name                   string                 `json:"name,omitempty"`
	Scheduler              SchedulerSpec          `json:"scheduler,omitempty"`
	Repositories           []RepositoryInfo       `json:"repositories,omitempty"`
	Addon                  AddonInfo              `json:"addon,omitempty"`
	RetentionPolicy        kmapi.ObjectReference  `json:"retentionPolicy,omitempty"`
	VerificationStrategies []VerificationStrategy `json:"verificationStrategies,omitempty"`
	Hooks                  BackupHooks            `json:"hooks,omitempty"`
	FailurePolicy          apis.FailurePolicy     `json:"failurePolicy,omitempty"`
	RetryConfig            *apis.RetryConfig      `json:"retryConfig,omitempty"`
	SessionHistoryLimit    *int32                 `json:"sessionHistoryLimit,omitempty"`
}

type SchedulerSpec struct {
	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule" protobuf:"bytes,1,opt,name=schedule"`

	// Optional deadline in seconds for starting the job if it misses scheduled
	// time for any reason.  Missed jobs executions will be counted as failed ones.
	// +optional
	StartingDeadlineSeconds *int64 `json:"startingDeadlineSeconds,omitempty" protobuf:"varint,2,opt,name=startingDeadlineSeconds"`

	// Specifies how to treat concurrent executions of a Job.
	// Valid values are:
	// - "Allow" (default): allows CronJobs to run concurrently;
	// - "Forbid": forbids concurrent runs, skipping next run if previous run hasn't finished yet;
	// - "Replace": cancels currently running job and replaces it with a new one
	// +optional
	ConcurrencyPolicy batchv1.ConcurrencyPolicy `json:"concurrencyPolicy,omitempty" protobuf:"bytes,3,opt,name=concurrencyPolicy,casttype=ConcurrencyPolicy"`

	// This flag tells the controller to suspend subsequent executions, it does
	// not apply to already started executions.  Defaults to false.
	// +optional
	Suspend *bool `json:"suspend,omitempty" protobuf:"varint,4,opt,name=suspend"`

	// Specifies the job that will be created when executing a CronJob.
	JobTemplate batchv1.JobTemplateSpec `json:"jobTemplate" protobuf:"bytes,5,opt,name=jobTemplate"`

	// The number of successful finished jobs to retain. Value must be non-negative integer.
	// Defaults to 3.
	// +optional
	SuccessfulJobsHistoryLimit *int32 `json:"successfulJobsHistoryLimit,omitempty" protobuf:"varint,6,opt,name=successfulJobsHistoryLimit"`

	// The number of failed finished jobs to retain. Value must be non-negative integer.
	// Defaults to 1.
	// +optional
	FailedJobsHistoryLimit *int32 `json:"failedJobsHistoryLimit,omitempty" protobuf:"varint,7,opt,name=failedJobsHistoryLimit"`
}

type JobTemplate struct {
	// Specifies the maximum desired number of pods the job should
	// run at any given time. The actual number of pods running in steady state will
	// be less than this number when ((.spec.completions - .status.successful) < .spec.parallelism),
	// i.e. when the work left to do is less than max parallelism.
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/
	// +optional
	Parallelism *int32 `json:"parallelism,omitempty" protobuf:"varint,1,opt,name=parallelism"`

	// Specifies the desired number of successfully finished pods the
	// job should be run with.  Setting to nil means that the success of any
	// pod signals the success of all pods, and allows parallelism to have any positive
	// value.  Setting to 1 means that parallelism is limited to 1 and the success of that
	// pod signals the success of the job.
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/
	// +optional
	Completions *int32 `json:"completions,omitempty" protobuf:"varint,2,opt,name=completions"`

	// Specifies the duration in seconds relative to the startTime that the job
	// may be continuously active before the system tries to terminate it; value
	// must be positive integer. If a Job is suspended (at creation or through an
	// update), this timer will effectively be stopped and reset when the Job is
	// resumed again.
	// +optional
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty" protobuf:"varint,3,opt,name=activeDeadlineSeconds"`

	// Specifies the number of retries before marking this job failed.
	// Defaults to 6
	// +optional
	BackoffLimit *int32 `json:"backoffLimit,omitempty" protobuf:"varint,7,opt,name=backoffLimit"`

	// Describes the pod that will be created when executing a job.
	Template ofst.PodTemplateSpec `json:"template" protobuf:"bytes,6,opt,name=template"`

	// ttlSecondsAfterFinished limits the lifetime of a Job that has finished
	// execution (either Complete or Failed). If this field is set,
	// ttlSecondsAfterFinished after the Job finishes, it is eligible to be
	// automatically deleted. When the Job is being deleted, its lifecycle
	// guarantees (e.g. finalizers) will be honored. If this field is unset,
	// the Job won't be automatically deleted. If this field is set to zero,
	// the Job becomes eligible to be deleted immediately after it finishes.
	// This field is alpha-level and is only honored by servers that enable the
	// TTLAfterFinished feature.
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty" protobuf:"varint,8,opt,name=ttlSecondsAfterFinished"`

	// CompletionMode specifies how Pod completions are tracked. It can be
	// `NonIndexed` (default) or `Indexed`.
	//
	// `NonIndexed` means that the Job is considered complete when there have
	// been .spec.completions successfully completed Pods. Each Pod completion is
	// homologous to each other.
	//
	// `Indexed` means that the Pods of a
	// Job get an associated completion index from 0 to (.spec.completions - 1),
	// available in the annotation batch.kubernetes.io/job-completion-index.
	// The Job is considered complete when there is one successfully completed Pod
	// for each index.
	// When value is `Indexed`, .spec.completions must be specified and
	// `.spec.parallelism` must be less than or equal to 10^5.
	//
	// This field is alpha-level and is only honored by servers that enable the
	// IndexedJob feature gate. More completion modes can be added in the future.
	// If the Job controller observes a mode that it doesn't recognize, the
	// controller skips updates for the Job.
	// +optional
	CompletionMode *batchv1.CompletionMode `json:"completionMode,omitempty" protobuf:"bytes,9,opt,name=completionMode,casttype=CompletionMode"`

	// Suspend specifies whether the Job controller should create Pods or not. If
	// a Job is created with suspend set to true, no Pods are created by the Job
	// controller. If a Job is suspended after creation (i.e. the flag goes from
	// false to true), the Job controller will delete all active Pods associated
	// with this Job. Users must design their workload to gracefully handle this.
	// Suspending a Job will reset the StartTime field of the Job, effectively
	// resetting the ActiveDeadlineSeconds timer too. This is an alpha field and
	// requires the SuspendJob feature gate to be enabled; otherwise this field
	// may not be set to true. Defaults to false.
	// +optional
	Suspend *bool `json:"suspend,omitempty" protobuf:"varint,10,opt,name=suspend"`
}

type RepositoryInfo struct {
	Name             string `json:"name,omitempty"`
	Backend          string `json:"backend,omitempty"`
	Directory        string `json:"directory,omitempty"`
	EncryptionSecret string `json:"encryptionSecret,omitempty"`
}

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
	TargetVolumes TargetVolumeInfo      `json:"targetVolumes,omitempty"`
	AddonVolumes  []apis.VolumeSource   `json:"addonVolumes,omitempty"`
}

type TargetVolumeInfo struct {
	Volumes      []core.Volume      `json:"volumes,omitempty"`
	VolumeMounts []core.VolumeMount `json:"volumeMounts,omitempty"`
}

type VerificationStrategy struct {
	Name        string                    `json:"name,omitempty"`
	Repository  string                    `json:"repository,omitempty"`
	Verifier    apis.TypedObjectReference `json:"verifier,omitempty"`
	Params      *runtime.RawExtension     `json:"params,omitempty"`
	VerifyEvery int32                     `json:"verifyEvery,omitempty"`
	OnFailure   apis.FailurePolicy        `json:"onFailure,omitempty"`
}

type BackupHooks struct {
	PreBackup  []HookInfo `json:"preBackup,omitempty"`
	PostBackup []HookInfo `json:"postBackup,omitempty"`
}

// BackupConfigurationStatus defines the observed state of BackupConfiguration
type BackupConfigurationStatus struct {
	Ready         bool                  `json:"ready,omitempty"`
	Target        ResourceFoundStatus   `json:"target,omitempty"`
	Backends      []BackendStatus       `json:"backends,omitempty"`
	Addons        []ResourceFoundStatus `json:"addons,omitempty"`
	Repositories  []RepoStatus          `json:"repositories,omitempty"`
	Verifiers     []ResourceFoundStatus `json:"verifiers,omitempty"`
	HookTemplates []ResourceFoundStatus `json:"hookTemplates,omitempty"`
	Sessions      []SessionStatus       `json:"sessions,omitempty"`
}

type BackendStatus struct {
	Name  string `json:"name,omitempty"`
	Found bool   `json:"found,omitempty"`
	Ready bool   `json:"ready,omitempty"`
}

type RepoStatus struct {
	Name        string `json:"name,omitempty"`
	Initialized bool   `json:"initialized,omitempty"`
	Error       string `json:"error,omitempty"`
}

type SessionStatus struct {
	Name         string `json:"name,omitempty"`
	NextSchedule string `json:"nextSchedule,omitempty"`
	Conditions   []kmapi.Condition
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// BackupConfiguration is the Schema for the backupconfigurations API
type BackupConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupConfigurationSpec   `json:"spec,omitempty"`
	Status BackupConfigurationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BackupConfigurationList contains a list of BackupConfiguration
type BackupConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupConfiguration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackupConfiguration{}, &BackupConfigurationList{})
}
