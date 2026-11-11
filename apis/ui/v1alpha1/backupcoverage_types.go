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
	ResourceKindBackupCoverage     = "BackupCoverage"
	ResourceSingularBackupCoverage = "backupcoverage"
	ResourcePluralBackupCoverage   = "backupcoverages"

	// BackupCoverageName is the name of the only BackupCoverage object.
	BackupCoverageName = "cluster"
)

// BackupCoverage summarizes backup coverage of the whole cluster: database counts and cluster manifest backups.
// It is cluster scoped and has a single object named "cluster".
// Counts include only databases and BackupConfigurations the requester may get.

// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
type BackupCoverage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec BackupCoverageSpec `json:"spec,omitempty"`
}

type BackupCoverageSpec struct {
	Cluster   ClusterIdentity `json:"cluster"`
	Databases int32           `json:"databases"`
	// Protected counts databases targeted by at least one BackupConfiguration.
	Protected int32 `json:"protected"`
	// Failing counts protected databases whose health is LastBackupFailed, StorageUnreachable or NotReady.
	Failing     int32 `json:"failing"`
	PITREnabled int32 `json:"pitrEnabled"`
	RPOBreached int32 `json:"rpoBreached"`
	// ManifestBackups are the BackupConfigurations that back up cluster manifests rather than one application.
	// +optional
	ManifestBackups []ManifestBackup `json:"manifestBackups,omitempty"`
}

// ClusterIdentity is the subset of kmodules ClusterMetadata a hub needs to group results.
type ClusterIdentity struct {
	UID string `json:"uid"`
	// +optional
	Name string `json:"name,omitempty"`
	// +optional
	DisplayName string `json:"displayName,omitempty"`
	// +optional
	OwnerID string `json:"ownerID,omitempty"`
	// +optional
	OwnerType string `json:"ownerType,omitempty"`
}

type ManifestBackup struct {
	BackupConfiguration kmapi.ObjectReference `json:"backupConfiguration"`
	Health              BackupHealth          `json:"health"`
	// Schedule is the most frequent cron schedule across the BackupConfiguration's sessions.
	// +optional
	Schedule string `json:"schedule,omitempty"`
	// +optional
	LastSuccessfulBackupTime *metav1.Time `json:"lastSuccessfulBackupTime,omitempty"`
	// +optional
	Size      string `json:"size,omitempty"`
	SizeBytes int64  `json:"sizeBytes"`
	// +optional
	Retention storageapi.RetentionPeriod `json:"retention,omitempty"`
	RPO       RPOStats                   `json:"rpo"`
}

func init() {
	SchemeBuilder.Register(&BackupCoverage{})
}
