/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"kubestash.dev/apimachinery/apis"
	coreapi "kubestash.dev/apimachinery/apis/core/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

// BackupStorage points at the BackupStorage the archiver's Repositories are created in.
type BackupStorage struct {
	// Ref refers to the BackupStorage CR. It may live in another namespace as long
	// as that namespace is allowed by the BackupStorage's usage policy.
	// +optional
	Ref *kmapi.ObjectReference `json:"ref,omitempty"`

	// SubDir is the directory inside the BackupStorage the archiver writes under.
	// +optional
	SubDir string `json:"subDir,omitempty"`
}

// FullBackupOptions configures the scheduled half that produces full backups.
// A full re-reads every allocated extent, so it depends on no earlier data and
// becomes the point the checkpoint chain can be pruned back to.
type FullBackupOptions struct {
	// Driver specifies the tool that moves the data. The default is BlockCAS,
	// which is also what the resident half writes with — sharing one block store
	// is what makes a daily full a hash comparison instead of a whole-disk upload.
	// +kubebuilder:default:=BlockCAS
	// +optional
	Driver apis.Driver `json:"driver,omitempty"`

	// Scheduler specifies the configuration for the backup triggering CronJob.
	// +optional
	Scheduler *coreapi.SchedulerSpec `json:"scheduler,omitempty"`

	// JobTemplate specifies the pod template for the backup Job.
	// +optional
	JobTemplate *ofst.PodTemplateSpec `json:"jobTemplate,omitempty"`

	// RetryConfig specifies the behavior of retry in case of a backup failure.
	// +optional
	RetryConfig *coreapi.RetryConfig `json:"retryConfig,omitempty"`

	// Timeout specifies the maximum duration of the backup.
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// SessionHistoryLimit specifies how many backup Jobs and associated resources
	// KubeStash should keep for debugging purpose.
	// +optional
	SessionHistoryLimit int32 `json:"sessionHistoryLimit,omitempty"`
}

// ManifestBackupOptions configures the scheduled backup of the VirtualMachine's
// object graph — the VM, its VMI and its DataVolumes — which Restic stores in a
// repository of its own, separate from the disk blocks.
type ManifestBackupOptions struct {
	// Scheduler specifies the configuration for the backup triggering CronJob.
	// +optional
	Scheduler *coreapi.SchedulerSpec `json:"scheduler,omitempty"`

	// JobTemplate specifies the pod template for the backup Job.
	// +optional
	JobTemplate *ofst.PodTemplateSpec `json:"jobTemplate,omitempty"`

	// RetryConfig specifies the behavior of retry in case of a backup failure.
	// +optional
	RetryConfig *coreapi.RetryConfig `json:"retryConfig,omitempty"`

	// Timeout specifies the maximum duration of the backup.
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// SessionHistoryLimit specifies how many backup Jobs and associated resources
	// KubeStash should keep for debugging purpose.
	// +optional
	SessionHistoryLimit int32 `json:"sessionHistoryLimit,omitempty"`
}

// CBTBackupOptions configures the resident half: a Job that loops on Interval,
// reading the blocks KubeVirt reports dirty since the previous checkpoint.
//
// There is deliberately no Scheduler field here. Its absence is what marks this
// half resident — the same convention KubeDB uses for LogBackupOptions. Adding a
// scheduler would turn the loop back into a CronJob, whose ~70-90s of Job and PVC
// churn per run is more than the interval it is meant to fill.
type CBTBackupOptions struct {
	// Interval is how long the loop sleeps between checkpoints. The floor is set
	// by the KubeVirt export handshake, which costs ~15-20s per increment.
	// +kubebuilder:default="10m"
	// +optional
	Interval *metav1.Duration `json:"interval,omitempty"`

	// ScratchStorageClass is the StorageClass for the scratch PVC. KubeVirt's
	// webhook requires one even in Pull mode, where nothing reads it back.
	// +optional
	ScratchStorageClass string `json:"scratchStorageClass,omitempty"`

	// TTLDuration bounds how long an export outlives a crashed archiver. It is a
	// safety net, not the completion mechanism: deleting the VirtualMachineBackup
	// is what completes a run and advances the checkpoint.
	// +optional
	TTLDuration *metav1.Duration `json:"ttlDuration,omitempty"`

	// RetentionPeriod bounds the restorable window, expressed as `XXu` where `XX`
	// is a positive integer and `u` is one of `dwmy` — days, weeks, months, years.
	// Checkpoints retire on this clock rather than under the RetentionPolicy,
	// because they are not Snapshots — they are entries inside one.
	// +kubebuilder:validation:Pattern=^[1-9][0-9]*[dwmy]$
	// +kubebuilder:default="7d"
	// +optional
	RetentionPeriod string `json:"retentionPeriod,omitempty"`

	// RetentionSchedule is the cron expression for the checkpoint pruning task.
	// +kubebuilder:default="0 0 * * *"
	// +optional
	RetentionSchedule string `json:"retentionSchedule,omitempty"`

	// SuccessfulHistoryLimit is how many succeeded checkpoints the incremental
	// Snapshot keeps in its bounded status window.
	// +kubebuilder:default=5
	// +optional
	SuccessfulHistoryLimit int32 `json:"successfulHistoryLimit,omitempty"`

	// FailedHistoryLimit is how many failed checkpoints the incremental Snapshot
	// keeps. A failed iteration never fails the Snapshot, so these counters are
	// the only signal that the archiver is sick.
	// +kubebuilder:default=5
	// +optional
	FailedHistoryLimit int32 `json:"failedHistoryLimit,omitempty"`

	// RuntimeSettings allows tuning the resident Job's pod and container.
	// +optional
	RuntimeSettings *ofst.RuntimeSettings `json:"runtimeSettings,omitempty"`
}
