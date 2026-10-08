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
	"k8s.io/apimachinery/pkg/runtime"
	kmapi "kmodules.xyz/client-go/api/v1"
)

const (
	ResourceKindRestoreRequest     = "RestoreRequest"
	ResourceSingularRestoreRequest = "restorerequest"
	ResourcePluralRestoreRequest   = "restorerequests"

	// AnnotationDeleteAfter marks a clone for garbage collection once the RFC3339 time has passed.
	AnnotationDeleteAfter = "ui.kubestash.com/delete-after"
)

// +kubebuilder:validation:Enum=Clone;Manifest
type RestoreRequestType string

const (
	// RestoreRequestTypeClone restores an application into a new object in a separate namespace,
	// leaving the source untouched.
	RestoreRequestTypeClone RestoreRequestType = "Clone"
	// RestoreRequestTypeManifest re-applies backed up Kubernetes manifests.
	RestoreRequestTypeManifest RestoreRequestType = "Manifest"
)

// RestoreRequest renders and creates the objects needed to restore from a repository.
// It is create-only and never persisted. Its metadata.labels are copied onto every created object.
// With dryRun=All nothing is created and the rendered objects are returned in status.rendered.

// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
type RestoreRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RestoreRequestSpec `json:"spec"`
	// +optional
	Status RestoreRequestStatus `json:"status,omitempty"`
}

type RestoreRequestSpec struct {
	Type   RestoreRequestType `json:"type"`
	Source RestoreSource      `json:"source"`
	// +optional
	Clone *CloneOptions `json:"clone,omitempty"`
	// +optional
	Manifest *ManifestOptions `json:"manifest,omitempty"`
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`
}

type RestoreSource struct {
	// Repository is in the namespace of the RestoreRequest.
	Repository string `json:"repository"`
	// Snapshot is a Snapshot name or "latest". It is mutually exclusive with PITR.
	// +optional
	Snapshot string `json:"snapshot,omitempty"`
	// PITR is only valid for repositories holding a continuous-log archive.
	// +optional
	PITR *PITR `json:"pitr,omitempty"`
}

// PITR selects a point in time inside a continuous-log window.
type PITR struct {
	TargetTime metav1.Time `json:"targetTime"`
}

type CloneOptions struct {
	// Namespace receives the clone. It is created when missing.
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	// DeleteAfter garbage-collects the clone once it has existed this long.
	// +optional
	DeleteAfter *metav1.Duration `json:"deleteAfter,omitempty"`
}

type ManifestOptions struct {
	IncludeClusterResources bool `json:"includeClusterResources"`
	// +optional
	IncludeNamespaces []string `json:"includeNamespaces,omitempty"`
	// +optional
	ExcludeNamespaces []string `json:"excludeNamespaces,omitempty"`
	// +optional
	IncludeResources []string `json:"includeResources,omitempty"`
	// +optional
	ExcludeResources []string `json:"excludeResources,omitempty"`
	// +optional
	ANDedLabelSelectors []string `json:"andedLabelSelectors,omitempty"`
	OverrideResources   bool     `json:"overrideResources"`
	RestorePVs          bool     `json:"restorePVs"`
	// ServiceAccountName runs the restore job. It needs write access to every restored resource,
	// so the read-only account used for backup does not suffice.
	ServiceAccountName string `json:"serviceAccountName"`
}

type RestoreRequestStatus struct {
	// +optional
	Objects []kmapi.TypedObjectReference `json:"objects,omitempty"`
	// Rendered holds the objects that would be created, filled only for dry runs.
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Rendered []runtime.RawExtension `json:"rendered,omitempty"`
	// +optional
	Warnings []string `json:"warnings,omitempty"`
}

func init() {
	SchemeBuilder.Register(&RestoreRequest{})
}
