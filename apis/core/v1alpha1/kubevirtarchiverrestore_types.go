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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "kmodules.xyz/client-go/api/v1"
)

const (
	ResourceKindKubeVirtArchiverRestore     = "KubeVirtArchiverRestore"
	ResourceSingularKubeVirtArchiverRestore = "kubevirtarchiverrestore"
	ResourcePluralKubeVirtArchiverRestore   = "kubevirtarchiverrestores"
)

// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=kubevirtarchiverrestores,singular=kubevirtarchiverrestore,shortName=kvarestore,categories={archiver,kubestash,appscode,all}
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Checkpoint",type="string",JSONPath=".status.checkpoint"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// KubeVirtArchiverRestore restores one VirtualMachine from the complete CBT
// recovery points maintained by a KubeVirtArchiver. It is intentionally a
// KubeVirt-specific entrypoint: generic RestoreSession continues to restore an
// explicitly supplied source without understanding live checkpoint chains.
type KubeVirtArchiverRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeVirtArchiverRestoreSpec   `json:"spec,omitempty"`
	Status KubeVirtArchiverRestoreStatus `json:"status,omitempty"`
}

type KubeVirtArchiverRestoreSpec struct {
	// ArchiverRef identifies the same-namespace archiver that owns the source VM.
	ArchiverRef v1.ObjectReference `json:"archiverRef"`

	// Source identifies a VirtualMachine selected by ArchiverRef. Namespace is
	// omitted because v1 restores are deliberately same-namespace only.
	Source v1.ObjectReference `json:"source"`

	// Target is the name of a new VirtualMachine. The controller fails rather
	// than overwriting an existing target.
	Target v1.ObjectReference `json:"target"`

	// RecoveryPoint selects the newest complete checkpoint at or before the
	// requested time. An omitted targetTime selects the newest complete point.
	// +optional
	RecoveryPoint *KubeVirtRecoveryPoint `json:"recoveryPoint,omitempty"`
}

type KubeVirtRecoveryPoint struct {
	// TargetTime is an RFC3339 timestamp. The selection is inclusive.
	// +optional
	TargetTime *metav1.Time `json:"targetTime,omitempty"`
}

type KubeVirtArchiverRestoreStatus struct {
	// ObservedGeneration is the most recent generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Phase is Pending, Running, Succeeded, or Failed.
	// +optional
	Phase RestorePhase `json:"phase,omitempty"`

	// Conditions reports validation, checkpoint selection, and both restore stages.
	// +optional
	Conditions []v1.Condition `json:"conditions,omitempty"`

	// Snapshot is the resident Snapshot frozen for this restore.
	// +optional
	Snapshot *v1.ObjectReference `json:"snapshot,omitempty"`

	// Checkpoint and RecoveryTime identify the immutable recovery point selected.
	// +optional
	Checkpoint string `json:"checkpoint,omitempty"`
	// +optional
	RecoveryTime *metav1.Time `json:"recoveryTime,omitempty"`

	// VolumeRestoreSession and ManifestRestoreSession are the owned standard
	// RestoreSessions used for the two stages.
	// +optional
	VolumeRestoreSession *v1.ObjectReference `json:"volumeRestoreSession,omitempty"`
	// +optional
	ManifestRestoreSession *v1.ObjectReference `json:"manifestRestoreSession,omitempty"`
}

const (
	TypeKubeVirtArchiverRestoreValidated         = "Validated"
	TypeKubeVirtArchiverRecoveryPointSelected    = "RecoveryPointSelected"
	TypeKubeVirtArchiverVolumeRestoreSucceeded   = "VolumeRestoreSucceeded"
	TypeKubeVirtArchiverManifestRestoreSucceeded = "ManifestRestoreSucceeded"
)

// +kubebuilder:object:root=true

type KubeVirtArchiverRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubeVirtArchiverRestore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubeVirtArchiverRestore{}, &KubeVirtArchiverRestoreList{})
}
