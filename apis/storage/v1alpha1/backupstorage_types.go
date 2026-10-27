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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

type BackupStoragePhase string

const (
	ResourceKindBackupStorage     = "BackupStorage"
	ResourceSingularBackupStorage = "backupstorage"
	ResourcePluralBackupStorage   = "backupstorages"

	BackupStorageReady    BackupStoragePhase = "Ready"
	BackupStorageNotReady BackupStoragePhase = "NotReady"
)

// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=backupstorages,singular=backupstorage,categories={kubestash,appscode,all}
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".spec.storage.provider"
// +kubebuilder:printcolumn:name="Default",type="boolean",JSONPath=".spec.default"
// +kubebuilder:printcolumn:name="Deletion-Policy",type="string",JSONPath=".spec.deletionPolicy"
// +kubebuilder:printcolumn:name="Total-Size",type="string",JSONPath=".status.totalSize"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// BackupStorage specifies the backend information where the backed up data of different applications will be stored.
// You can consider BackupStorage as a representation of a bucket in Kubernetes native way.
// This is a namespaced object. However, you can use the BackupStorage from any namespace
// as long as it is permitted by the `.spec.usagePolicy` field.
type BackupStorage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupStorageSpec   `json:"spec,omitempty"`
	Status BackupStorageStatus `json:"status,omitempty"`
}

// BackupStorageSpec defines information regarding remote backend, its access credentials, usage policy etc.
type BackupStorageSpec struct {
	// Storage specifies the remote storage information
	Storage Backend `json:"storage,omitempty"`
	// UsagePolicy specifies a policy of how this BackupStorage will be used. For example, you can use `allowedNamespaces`
	// policy to restrict the usage of this BackupStorage to particular namespaces.
	// This field is optional. If you don't provide the usagePolicy, then it can be used only from the current namespace.
	// +optional
	UsagePolicy *apis.UsagePolicy `json:"usagePolicy,omitempty"`

	// Default specifies whether to use this BackupStorage as default storage for the current namespace
	// as well as the allowed namespaces. One namespace can have at most one default BackupStorage configured.
	// +optional
	Default bool `json:"default,omitempty"`

	// DeletionPolicy specifies what to do when you delete a BackupStorage CR.
	// The valid values are:
	// "Delete": This will delete the respective Repository and Snapshot CRs from the cluster but keep the backed up data in the remote backend. This is the default behavior.
	// "WipeOut": This will delete the respective Repository and Snapshot CRs as well as the backed up data from the backend.
	// +kubebuilder:default=Delete
	// +optional
	DeletionPolicy DeletionPolicy `json:"deletionPolicy,omitempty"`

	// RuntimeSettings allow to specify Resources, NodeSelector, Affinity, Toleration, ReadinessProbe etc.
	// for the storage initializer/cleaner job.
	// +optional
	RuntimeSettings ofst.RuntimeSettings `json:"runtimeSettings,omitempty"`

	// ObjectLock specifies how KubeStash should use the Object Lock (WORM) feature of the bucket.
	// KubeStash always detects whether the bucket has Object Lock enabled and, if it has, keeps the
	// objects that are still used by the backups locked. This field lets you make Object Lock mandatory
	// and tune the retention period. Currently, only the S3 provider supports Object Lock.
	// +optional
	ObjectLock *ObjectLockSpec `json:"objectLock,omitempty"`
}

// ObjectLockMode specifies the Object Lock retention mode of a bucket.
// +kubebuilder:validation:Enum=Governance;Compliance
type ObjectLockMode string

const (
	// ObjectLockModeGovernance allows a user with special permission to shorten or remove the retention of an object.
	ObjectLockModeGovernance ObjectLockMode = "Governance"
	// ObjectLockModeCompliance does not allow anyone, including the root user, to shorten or remove the retention of an object.
	ObjectLockModeCompliance ObjectLockMode = "Compliance"
)

type ObjectLockSpec struct {
	// Required specifies whether the bucket must have Object Lock enabled with a default retention.
	// If it is true and the bucket does not have Object Lock enabled, the BackupStorage will not be Ready.
	// +optional
	Required bool `json:"required,omitempty"`

	// Mode specifies the Object Lock mode the bucket must use. If it is set and the bucket uses a
	// different mode, the BackupStorage will not be Ready. Use "Compliance" for ransomware protection,
	// as the retention of an object in "Governance" mode can be removed by a user with bypass permission.
	// +optional
	Mode ObjectLockMode `json:"mode,omitempty"`

	// RetentionPeriod specifies how long the objects that are still used by the backups stay locked.
	// After each backup, KubeStash extends the retain-until date of these objects to now + RetentionPeriod.
	// The objects removed by the retention policy are not extended, so they become deletable once
	// their retention expires. If it is not set, the default retention period of the bucket is used.
	// +optional
	RetentionPeriod *metav1.Duration `json:"retentionPeriod,omitempty"`

	// DisableRetentionExtension disables extending the retention of the objects that are still used by the backups.
	// Note that, if it is disabled, an old object that is still used by a new backup can be deleted once its retention expires.
	// +optional
	DisableRetentionExtension bool `json:"disableRetentionExtension,omitempty"`
}

// BackupStorageStatus defines the observed state of BackupStorage
type BackupStorageStatus struct {
	// Phase indicates the overall phase of the backup BackupStorage. Phase will be "Ready" only
	// if the Backend is initialized and Repositories are synced.
	// +optional
	Phase BackupStoragePhase `json:"phase,omitempty"`

	// TotalSize represents the total backed up data size in this storage.
	// This is simply the summation of sizes of all Repositories using this BackupStorage.
	// +optional
	TotalSize string `json:"totalSize,omitempty"`

	// Repositories holds the information of all Repositories using this BackupStorage
	// +optional
	Repositories []RepositoryInfo `json:"repositories,omitempty"`

	// Conditions represents list of conditions regarding this BackupStorage
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty"`

	// ObjectLock shows the Object Lock (WORM) configuration found on the bucket.
	// +optional
	ObjectLock *ObjectLockStatus `json:"objectLock,omitempty"`
}

type ObjectLockStatus struct {
	// Enabled specifies whether Object Lock is enabled on the bucket.
	Enabled bool `json:"enabled"`

	// Mode specifies the default retention mode of the bucket.
	// +optional
	Mode ObjectLockMode `json:"mode,omitempty"`

	// DefaultRetention specifies the default retention period of the bucket.
	// +optional
	DefaultRetention *metav1.Duration `json:"defaultRetention,omitempty"`
}

// RepositoryInfo specifies information regarding a Repository using the BackupStorage
type RepositoryInfo struct {
	// Name represents the name of the respective Repository CR
	Name string `json:"name,omitempty"`

	// Namespace represent the namespace where the Repository CR has been created
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Path represents the directory inside the BackupStorage where this Repository is storing its data
	// This path is relative to the path of BackupStorage.
	Path string `json:"path,omitempty"`

	// Size represents the size of the backed up data in this Repository
	// +optional
	Size string `json:"size,omitempty"`

	// Synced specifies whether this Repository state has been synced with the cloud state or not
	// +optional
	Synced *bool `json:"synced,omitempty"`

	// Error specifies the reason in case of Repository sync failure.
	// +optional
	Error *string `json:"error,omitempty"`
}

const (
	TypeBackendInitialized               = "BackendInitialized"
	ReasonBackendInitializationSucceeded = "BackendInitializationSucceeded"
	ReasonBackendInitializationFailed    = "BackendInitializationFailed"

	TypeBackendSecretFound          = "BackendSecretFound"
	ReasonBackendSecretNotAvailable = "BackendSecretNotAvailable"
	ReasonBackendSecretAvailable    = "BackendSecretAvailable"

	TypeObjectLockVerified         = "ObjectLockVerified"
	ReasonObjectLockEnabled        = "ObjectLockEnabled"
	ReasonObjectLockNotEnabled     = "ObjectLockNotEnabled"
	ReasonObjectLockModeMismatched = "ObjectLockModeMismatched"
	ReasonObjectLockCheckFailed    = "ObjectLockCheckFailed"
	ReasonObjectLockNotSupported   = "ObjectLockNotSupported"
)

//+kubebuilder:object:root=true

// BackupStorageList contains a list of BackupStorage
type BackupStorageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupStorage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackupStorage{}, &BackupStorageList{})
}
