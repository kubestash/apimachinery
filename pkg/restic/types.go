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

package restic

import (
	kmapi "kmodules.xyz/client-go/api/v1"
)

type RestoreMemberStatus struct {
	// Ref is the reference to the respective target whose status is shown here.
	Ref kmapi.TypedObjectReference `json:"ref"`
	// Conditions shows the condition of different steps to restore this member.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty"`
	// TotalHosts specifies total number of hosts that will be restored for this member.
	// +optional
	TotalHosts *int32 `json:"totalHosts,omitempty"`
	// Phase indicates restore phase of this member
	// +optional
	Phase RestoreTargetPhase `json:"phase,omitempty"`
	// Stats shows restore statistics of individual hosts for this member
	// +optional
	Stats []HostRestoreStats `json:"stats,omitempty"`
}

// +kubebuilder:validation:Enum=Pending;Succeeded;Running;Failed
type RestoreTargetPhase string

const (
	TargetRestorePending      RestoreTargetPhase = "Pending"
	TargetRestoreRunning      RestoreTargetPhase = "Running"
	TargetRestoreSucceeded    RestoreTargetPhase = "Succeeded"
	TargetRestoreFailed       RestoreTargetPhase = "Failed"
	TargetRestorePhaseUnknown RestoreTargetPhase = "Unknown"
)

type HostRestoreStats struct {
	// Hostname indicate name of the host that has been restored
	// +optional
	Hostname string `json:"hostname,omitempty"`
	// Phase indicates restore phase of this host
	// +optional
	Phase HostRestorePhase `json:"phase,omitempty"`
	// Duration indicates total time taken to complete restore for this hosts
	// +optional
	Duration string `json:"duration,omitempty"`
	// Error indicates string value of error in case of restore failure
	// +optional
	Error string `json:"error,omitempty"`
}

// +kubebuilder:validation:Enum=Succeeded;Failed;Running;Unknown
type HostRestorePhase string

const (
	HostRestoreSucceeded HostRestorePhase = "Succeeded"
	HostRestoreFailed    HostRestorePhase = "Failed"
	HostRestoreRunning   HostRestorePhase = "Running"
	HostRestoreUnknown   HostRestorePhase = "Unknown"
)

type HostBackupStats struct {
	// Hostname indicate name of the host that has been backed up
	// +optional
	Hostname string `json:"hostname,omitempty"`
	// Phase indicates backup phase of this host
	// +optional
	Phase HostBackupPhase `json:"phase,omitempty"`
	// Snapshots specifies the stats of individual snapshots that has been taken for this host in current backup session
	// +optional
	Snapshots []SnapshotStats `json:"snapshots,omitempty"`
	// Duration indicates total time taken to complete backup for this hosts
	// +optional
	Duration string `json:"duration,omitempty"`
	// Error indicates string value of error in case of backup failure
	// +optional
	Error string `json:"error,omitempty"`
}

type SnapshotStats struct {
	// Name indicates the name of the backup snapshot created for this host
	Name string `json:"name,omitempty"`
	// Path indicates the directory that has been backed up in this snapshot
	Path string `json:"path,omitempty"`
	// TotalSize indicates the size of data to backup in target directory
	TotalSize string `json:"totalSize,omitempty"`
	// Uploaded indicates size of data uploaded to backend for this snapshot
	Uploaded string `json:"uploaded,omitempty"`
	// ProcessingTime indicates time taken to process the target data
	ProcessingTime string `json:"processingTime,omitempty"`
	// FileStats shows statistics of files of this snapshot
	FileStats FileStats `json:"fileStats,omitempty"`
}

type FileStats struct {
	// TotalFiles shows total number of files that has been backed up
	TotalFiles *int64 `json:"totalFiles,omitempty"`
	// NewFiles shows total number of new files that has been created since last backup
	NewFiles *int64 `json:"newFiles,omitempty"`
	// ModifiedFiles shows total number of files that has been modified since last backup
	ModifiedFiles *int64 `json:"modifiedFiles,omitempty"`
	// UnmodifiedFiles shows total number of files that has not been changed since last backup
	UnmodifiedFiles *int64 `json:"unmodifiedFiles,omitempty"`
}

// +kubebuilder:validation:Enum=Pending;Succeeded;Running;Failed
type TargetPhase string

const (
	TargetBackupPending   TargetPhase = "Pending"
	TargetBackupSucceeded TargetPhase = "Succeeded"
	TargetBackupRunning   TargetPhase = "Running"
	TargetBackupFailed    TargetPhase = "Failed"
)

// +kubebuilder:validation:Enum=Succeeded;Failed
type HostBackupPhase string

const (
	HostBackupSucceeded HostBackupPhase = "Succeeded"
	HostBackupFailed    HostBackupPhase = "Failed"
)
