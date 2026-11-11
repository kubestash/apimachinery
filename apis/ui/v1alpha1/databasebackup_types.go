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
	ResourceKindDatabaseBackup     = "DatabaseBackup"
	ResourceSingularDatabaseBackup = "databasebackup"
	ResourcePluralDatabaseBackup   = "databasebackups"
)

// DatabaseBackup is the backup state of one KubeDB database, whether or not any BackupConfiguration targets it.
// Its name is "{.metadata.name}~Kind.Group", e.g. "pg~Postgres.kubedb.com", and its namespace is the database's.

// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
type DatabaseBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DatabaseBackupSpec `json:"spec,omitempty"`
}

type DatabaseBackupSpec struct {
	Database kmapi.TypedObjectReference `json:"database"`
	// Version is the database's spec.version.
	// +optional
	Version string `json:"version,omitempty"`
	// Health is the worst health across BackupConfigurations, or Unprotected when there are none.
	Health BackupHealth `json:"health"`
	// +optional
	BackupConfigurations []kmapi.ObjectReference `json:"backupConfigurations,omitempty"`
	// Schedule is the most frequent cron schedule across the sessions of all BackupConfigurations.
	// +optional
	Schedule string `json:"schedule,omitempty"`
	PITR     bool   `json:"pitr"`
	// +optional
	LastSuccessfulBackupTime *metav1.Time `json:"lastSuccessfulBackupTime,omitempty"`
	// Retention is the longest MaxRetentionPeriod across the backends of all BackupConfigurations.
	// +optional
	Retention storageapi.RetentionPeriod `json:"retention,omitempty"`
	// RPO is taken from the BackupConfiguration with the most recent recoverable point.
	// +optional
	RPO *RPOStats `json:"rpo,omitempty"`
}

// +kubebuilder:object:root=true
type DatabaseBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseBackup{}, &DatabaseBackupList{})
}
