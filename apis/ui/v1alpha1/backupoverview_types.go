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
	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

const (
	ResourceKindBackupOverview     = "BackupOverview"
	ResourceSingularBackupOverview = "backupoverview"
	ResourcePluralBackupOverview   = "backupoverviews"

	// AnnotationRPOTarget on a BackupConfiguration sets the RPO it must meet, as a Go duration ("1m", "2h").
	AnnotationRPOTarget = "ui.kubestash.com/rpo-target"
)

// +kubebuilder:validation:Enum=Logical;Manifest;Physical;VolumeSnapshot;ContinuousLog
type BackupMethod string

const (
	BackupMethodLogical        BackupMethod = "Logical"
	BackupMethodManifest       BackupMethod = "Manifest"
	BackupMethodPhysical       BackupMethod = "Physical"
	BackupMethodVolumeSnapshot BackupMethod = "VolumeSnapshot"
	BackupMethodContinuousLog  BackupMethod = "ContinuousLog"
)

// +kubebuilder:validation:Enum=Healthy;Paused;NotReady;StorageUnreachable;LastBackupFailed;NoBackupYet;Unprotected
type BackupHealth string

const (
	BackupHealthHealthy            BackupHealth = "Healthy"
	BackupHealthPaused             BackupHealth = "Paused"
	BackupHealthNotReady           BackupHealth = "NotReady"
	BackupHealthStorageUnreachable BackupHealth = "StorageUnreachable"
	BackupHealthLastBackupFailed   BackupHealth = "LastBackupFailed"
	BackupHealthNoBackupYet        BackupHealth = "NoBackupYet"
	// BackupHealthUnprotected is reported only for databases that no BackupConfiguration targets.
	BackupHealthUnprotected BackupHealth = "Unprotected"
)

// BackupOverview is the computed state of a BackupConfiguration: schedules, last runs,
// repositories, backends, continuous-log (PITR) window, recovery timeline, and achieved RPO/RTO.
// Its name and namespace are those of the BackupConfiguration.

// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
type BackupOverview struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec BackupOverviewSpec `json:"spec,omitempty"`
}

type BackupOverviewSpec struct {
	Invoker kmapi.TypedObjectReference `json:"invoker"`
	// Target is nil for backups that are not bound to a single application, e.g. cluster manifests.
	// +optional
	Target *kmapi.TypedObjectReference `json:"target,omitempty"`
	// Phase mirrors the BackupConfiguration status phase.
	Phase    string            `json:"phase,omitempty"`
	Paused   bool              `json:"paused"`
	Methods  []BackupMethod    `json:"methods,omitempty"`
	Sessions []SessionOverview `json:"sessions,omitempty"`
	Backends []BackendOverview `json:"backends,omitempty"`
	// +optional
	PITR *PITROverview `json:"pitr,omitempty"`
	RPO  RPOStats      `json:"rpo"`
	RTO  RTOStats      `json:"rto"`
	// +optional
	TotalSize      string `json:"totalSize,omitempty"`
	TotalSizeBytes int64  `json:"totalSizeBytes"`
	// +optional
	LastSuccessfulBackupTime *metav1.Time `json:"lastSuccessfulBackupTime,omitempty"`
	// FirstRecoveryPoint is the earliest time this backup can restore to: the start of the
	// continuous-log window or the oldest succeeded snapshot, whichever is earlier.
	// +optional
	FirstRecoveryPoint *metav1.Time `json:"firstRecoveryPoint,omitempty"`
	Health             BackupHealth `json:"health"`
	// Timeline covers RPO.Window up to now.
	Timeline Timeline `json:"timeline"`
}

// Timeline lists every point a BackupConfiguration can be recovered to inside Range.
type Timeline struct {
	Range TimeRange `json:"range"`
	// ContinuousWindows is the continuous-log window clipped to Range, with failed log pushes cut out.
	// +optional
	ContinuousWindows []TimeRange     `json:"continuousWindows,omitempty"`
	Points            []RecoveryPoint `json:"points"`
}

type RecoveryPoint struct {
	Snapshot   string `json:"snapshot"`
	Repository string `json:"repository"`
	Session    string `json:"session"`
	// Time is the snapshot time, or the creation time for snapshots that never completed.
	Time  metav1.Time              `json:"time"`
	Phase storageapi.SnapshotPhase `json:"phase,omitempty"`
	Type  storageapi.BackupType    `json:"type,omitempty"`
	// +optional
	VerificationStatus storageapi.VerificationStatus `json:"verificationStatus,omitempty"`
	// +optional
	Size string `json:"size,omitempty"`
}

type SessionOverview struct {
	Name string `json:"name"`
	// +optional
	Addon string `json:"addon,omitempty"`
	// +optional
	Tasks []string `json:"tasks,omitempty"`
	// +optional
	Schedule  string `json:"schedule,omitempty"`
	Suspended bool   `json:"suspended"`
	// +optional
	NextSchedule *metav1.Time `json:"nextSchedule,omitempty"`
	// +optional
	LastBackup *BackupSessionRef `json:"lastBackup,omitempty"`
	// +optional
	LastSuccessfulBackup *BackupSessionRef    `json:"lastSuccessfulBackup,omitempty"`
	Repositories         []RepositoryOverview `json:"repositories,omitempty"`
}

type RepositoryOverview struct {
	Name    string                     `json:"name"`
	Backend string                     `json:"backend"`
	Phase   storageapi.RepositoryPhase `json:"phase,omitempty"`
	// +optional
	Integrity     *bool `json:"integrity,omitempty"`
	SnapshotCount int32 `json:"snapshotCount"`
	// +optional
	Size      string `json:"size,omitempty"`
	SizeBytes int64  `json:"sizeBytes"`
	// +optional
	LastBackupTime *metav1.Time `json:"lastBackupTime,omitempty"`
	Encrypted      bool         `json:"encrypted"`
	// +optional
	Verification *VerificationOverview `json:"verification,omitempty"`
}

type VerificationOverview struct {
	Verifier kmapi.ObjectReference `json:"verifier"`
	// Type mirrors the BackupVerifier type: RestoreOnly, Query or Script.
	Type string `json:"type,omitempty"`
	// +optional
	Schedule string `json:"schedule,omitempty"`
	// +optional
	NextSchedule *metav1.Time `json:"nextSchedule,omitempty"`
	// +optional
	LastSession *VerificationSessionRef `json:"lastSession,omitempty"`
	// Verified, VerificationFailed and NotVerified count this repository's Snapshots inside the RPO window.
	Verified           int32 `json:"verified"`
	VerificationFailed int32 `json:"verificationFailed"`
	NotVerified        int32 `json:"notVerified"`
}

type BackendOverview struct {
	Name    string                `json:"name"`
	Storage kmapi.ObjectReference `json:"storage"`
	// +optional
	Provider storageapi.StorageProvider `json:"provider,omitempty"`
	Ready    bool                       `json:"ready"`
	// +optional
	StoragePhase storageapi.BackupStoragePhase `json:"storagePhase,omitempty"`
	// Reachable means the BackupStorage backend is initialized and its credentials were found.
	Reachable bool `json:"reachable"`
	// Bucket is the S3/GCS bucket or Azure container.
	// +optional
	Bucket string `json:"bucket,omitempty"`
	// Endpoint is the S3 endpoint or Azure storage account.
	// +optional
	Endpoint string `json:"endpoint,omitempty"`
	// +optional
	Region string `json:"region,omitempty"`
	// +optional
	Prefix string `json:"prefix,omitempty"`
	// UsedSize is the size of everything on the BackupStorage, not only this BackupConfiguration's repositories.
	// +optional
	UsedSize      string `json:"usedSize,omitempty"`
	UsedSizeBytes int64  `json:"usedSizeBytes"`
	// ClientSideEncrypted is true when every repository of this BackupConfiguration on this backend is encrypted.
	ClientSideEncrypted bool `json:"clientSideEncrypted"`
	// +optional
	RetentionPolicy *RetentionSummary `json:"retentionPolicy,omitempty"`
	// +optional
	LastBackupTime *metav1.Time `json:"lastBackupTime,omitempty"`
	// LagBehindNewest is how far this backend's newest backup trails the newest backup across all backends.
	// It is nil for the newest backend.
	// +optional
	LagBehindNewest *metav1.Duration `json:"lagBehindNewest,omitempty"`
}

type RetentionSummary struct {
	Ref kmapi.ObjectReference `json:"ref"`
	// +optional
	MaxRetentionPeriod storageapi.RetentionPeriod `json:"maxRetentionPeriod,omitempty"`
	// +optional
	Successful *storageapi.SuccessfulSnapshotsKeepPolicy `json:"successful,omitempty"`
	// +optional
	FailedLast *int32 `json:"failedLast,omitempty"`
}

// PITROverview describes the continuous-log archive of a KubeDB archiver backup.
type PITROverview struct {
	Repository string `json:"repository"`
	// Snapshot is the IncrementalBackup Snapshot whose log component carries the log statistics.
	Snapshot  string `json:"snapshot"`
	Component string `json:"component"`
	// +optional
	Window *TimeRange `json:"window,omitempty"`
	// +optional
	LastPushTime *metav1.Time `json:"lastPushTime,omitempty"`
	// +optional
	LSN             string `json:"lsn,omitempty"`
	SucceededPushes int64  `json:"succeededPushes"`
	FailedPushes    int64  `json:"failedPushes"`
	// +optional
	LastFailure *LogFailure `json:"lastFailure,omitempty"`
	// +optional
	LastRetention *storageapi.LogRetentionStatus `json:"lastRetention,omitempty"`
}

type LogFailure struct {
	Range TimeRange `json:"range"`
	// +optional
	Error string `json:"error,omitempty"`
}

// +kubebuilder:object:root=true
type BackupOverviewList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupOverview `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackupOverview{}, &BackupOverviewList{})
}
