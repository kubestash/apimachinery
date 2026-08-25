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
	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

const (
	ResourceKindKubeVirtArchiver     = "KubeVirtArchiver"
	ResourceSingularKubeVirtArchiver = "kubevirtarchiver"
	ResourcePluralKubeVirtArchiver   = "kubevirtarchivers"
)

// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=kubevirtarchivers,singular=kubevirtarchiver,shortName=kvarchiver,categories={archiver,kubestash,appscode,all}
// +kubebuilder:printcolumn:name="Paused",type="boolean",JSONPath=".spec.pause"
// +kubebuilder:printcolumn:name="VMs",type="integer",JSONPath=".status.virtualMachinesMatched"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// KubeVirtArchiver ties together the two halves of incremental VirtualMachine backup:
// a scheduled full that re-anchors the chain once a day, and a resident Job that
// fills every gap between with changed-block checkpoints.
type KubeVirtArchiver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeVirtArchiverSpec   `json:"spec,omitempty"`
	Status KubeVirtArchiverStatus `json:"status,omitempty"`
}

// KubeVirtArchiverSpec defines the desired state of KubeVirtArchiver
type KubeVirtArchiverSpec struct {
	// VirtualMachines selects the VirtualMachines this archiver backs up.
	// +optional
	VirtualMachines *AllowedVirtualMachines `json:"virtualMachines,omitempty"`

	// BackupStorage is the backend the archiver's Repositories are created in.
	// +optional
	BackupStorage *BackupStorage `json:"backupStorage,omitempty"`

	// EncryptionSecret refers to the Secret holding the key the backed up data is
	// encrypted with.
	// +optional
	EncryptionSecret *kmapi.ObjectReference `json:"encryptionSecret,omitempty"`

	// RetentionPolicy governs how long the full and manifest Snapshots are kept.
	// Checkpoints retire separately, under CBTBackup.RetentionPeriod.
	// +optional
	RetentionPolicy *kmapi.ObjectReference `json:"retentionPolicy,omitempty"`

	// FullBackup configures the scheduled full backup session.
	// +optional
	FullBackup *FullBackupOptions `json:"fullBackup,omitempty"`

	// ManifestBackup configures the scheduled backup of the VM's object graph.
	// +optional
	ManifestBackup *ManifestBackupOptions `json:"manifestBackup,omitempty"`

	// CBTBackup configures the resident checkpoint loop. It carries no scheduler.
	// +optional
	CBTBackup *CBTBackupOptions `json:"cbtBackup,omitempty"`

	// Pause stops the resident loop and suspends the scheduled sessions. The
	// incremental Snapshot is finalised to Succeeded, which is the only terminal
	// transition it has.
	// +optional
	Pause bool `json:"pause,omitempty"`

	// DeletionPolicy specifies what happens to the created Repositories when this
	// archiver is deleted.
	// +optional
	DeletionPolicy *storageapi.BackupConfigDeletionPolicy `json:"deletionPolicy,omitempty"`
}

// AllowedVirtualMachines selects VirtualMachines by namespace and by label.
//
// Opt-in is selector-only. KubeStash does not own the VirtualMachine CRD, so there
// is no field on the VM to set — a VM joins an archiver by carrying the label.
type AllowedVirtualMachines struct {
	// Namespaces indicates which namespaces the VirtualMachines are selected from.
	// +optional
	Namespaces *apis.AllowedNamespaces `json:"namespaces,omitempty"`

	// Selector matches the VirtualMachines within those namespaces. An empty
	// selector matches every VirtualMachine in them.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
}

// KubeVirtArchiverStatus defines the observed state of KubeVirtArchiver
type KubeVirtArchiverStatus struct {
	// ObservedGeneration is the most recent generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions represents the current state of the archiver as a whole.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty"`

	// VirtualMachinesMatched is how many VirtualMachines the selector matched.
	// +optional
	VirtualMachinesMatched int32 `json:"virtualMachinesMatched,omitempty"`

	// VirtualMachines reports, per matched VirtualMachine, whether it is actually
	// backupable. Enabling changed-block tracking requires restarting the VM, so
	// the controller validates and reports rather than mutating — silently
	// restarting a production VM because a label changed would be hostile. This
	// list is where the user reads what to do about it.
	// +optional
	// +listType=map
	// +listMapKey=namespace
	// +listMapKey=name
	VirtualMachines []VirtualMachineArchiverStatus `json:"virtualMachines,omitempty"`
}

// VirtualMachineArchiverStatus is the archiver's view of one matched VirtualMachine.
type VirtualMachineArchiverStatus struct {
	// Name of the VirtualMachine.
	Name string `json:"name"`

	// Namespace of the VirtualMachine.
	Namespace string `json:"namespace"`

	// ChangedBlockTrackingReady is true only when vmi.status.changedBlockTracking
	// reports Enabled. Until then no backup for this VM can run at all: KubeVirt
	// refuses the VirtualMachineBackup outright.
	// +optional
	ChangedBlockTrackingReady bool `json:"changedBlockTrackingReady,omitempty"`

	// ChangedBlockTrackingState mirrors vmi.status.changedBlockTracking.state
	// verbatim. Each value maps to a different user action, so the raw value is
	// more useful than a boolean alone: Enabled, PendingRestart, Initializing,
	// Disabled, IncrementalBackupFeatureGateDisabled, or empty when the VM does
	// not match the cluster-level CBT label selector.
	// +optional
	ChangedBlockTrackingState string `json:"changedBlockTrackingState,omitempty"`

	// Reason is a short, machine-readable cause when ChangedBlockTrackingReady is
	// false.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message tells the user what to do about Reason — restart the VM, label it,
	// or enable the IncrementalBackup feature gate.
	// +optional
	Message string `json:"message,omitempty"`

	// LastObservedTime is when the controller last looked at this VM.
	// +optional
	LastObservedTime *metav1.Time `json:"lastObservedTime,omitempty"`
}

const (
	// TypeVirtualMachinesReady is True only when every matched VirtualMachine has
	// changed-block tracking Enabled.
	TypeVirtualMachinesReady           = "VirtualMachinesReady"
	ReasonAllVirtualMachinesReady      = "AllVirtualMachinesReady"
	ReasonChangedBlockTrackingNotReady = "ChangedBlockTrackingNotReady"

	// TypeBackupConfigurationsEnsured covers the scheduled half.
	TypeBackupConfigurationsEnsured          = "BackupConfigurationsEnsured"
	ReasonBackupConfigurationEnsureSucceeded = "BackupConfigurationEnsureSucceeded"
	ReasonBackupConfigurationEnsureFailed    = "BackupConfigurationEnsureFailed"

	// TypeArchiverJobsEnsured covers the resident half.
	TypeArchiverJobsEnsured          = "ArchiverJobsEnsured"
	ReasonArchiverJobEnsureSucceeded = "ArchiverJobEnsureSucceeded"
	ReasonArchiverJobEnsureFailed    = "ArchiverJobEnsureFailed"
)

// Reasons a VirtualMachine is not backupable, mapped from the CBT readiness state.
const (
	ReasonVMNotSelectedForCBT      = "NotSelectedForChangedBlockTracking"
	ReasonVMCBTPendingRestart      = "ChangedBlockTrackingPendingRestart"
	ReasonVMCBTInitializing        = "ChangedBlockTrackingInitializing"
	ReasonVMCBTFeatureGateDisabled = "IncrementalBackupFeatureGateDisabled"
	ReasonVMNotRunning             = "VirtualMachineNotRunning"
)

//+kubebuilder:object:root=true

// KubeVirtArchiverList contains a list of KubeVirtArchiver
type KubeVirtArchiverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubeVirtArchiver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubeVirtArchiver{}, &KubeVirtArchiverList{})
}
