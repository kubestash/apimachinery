//+kubebuilder:object:generate=true
package apis

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Driver specifies the name of underlying tool that is being used to upload the backed up data.
// +kubebuilder:validation:Enum=Restic;WalG
type Driver string

const (
	DriverRestic Driver = "Restic"
	DriverWalG   Driver = "WalG"
)

// TypedObjectReference let you reference an object from different namespace
type TypedObjectReference struct {
	core.TypedLocalObjectReference `json:",inline"`
	// Namespace points to the namespace of the targeted object.
	// If you don't provide this field, the object will be looked up in the local namespace.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// VolumeSource specifies the source of volume to mount in the backup/restore executor
type VolumeSource struct {
	core.VolumeSource `json:",inline"`

	// VolumeClaimTemplate specifies a template for volume to use by the backup/restore executor
	// +optional
	VolumeClaimTemplate *core.PersistentVolumeClaimTemplate `json:"volumeClaimTemplate,omitempty"`
}

// FailurePolicy specifies what to do if a backup/restore fails
// +kubebuilder:validation:Enum=Fail;Retry
type FailurePolicy string

const (
	FailurePolicyFail  FailurePolicy = "Fail"
	FailurePolicyRetry FailurePolicy = "Retry"
)

// RetryConfig specifies the behavior of retry
type RetryConfig struct {
	// MaxRetry specifies the maximum number of times Stash should retry the backup/restore process.
	// By default, Stash will retry only 1 time.
	// +kubebuilder:validation:default=1
	MaxRetry int32 `json:"maxRetry,omitempty"`

	// Delay specifies a duration to wait until next retry.
	// By default, Stash will retry immediately.
	// +optional
	Delay string `json:"delay,omitempty"`
}

// ParameterDefinition defines the parameter names, their usage, their requirements etc.
type ParameterDefinition struct {
	// Name specifies the name of the parameter
	Name string `json:"name,omitempty"`

	// Usage specifies the usage of this parameter
	Usage string `json:"usage,omitempty"`

	// Required specify whether this parameter is required or not
	// +optional
	Required bool `json:"required,omitempty"`

	// Default specifies a default value for the parameter
	// +optional
	Default string `json:"default,omitempty"`
}

// UsagePolicy specifies a policy that restrict the usage of a resource across namespaces.
type UsagePolicy struct {
	// AllowedNamespaces specifies which namespaces are allowed to use the resource
	// +optional
	AllowedNamespaces AllowedNamespaces `json:"allowedNamespaces,omitempty"`
}

// AllowedNamespaces indicate which namespaces the resource should be selected from.
type AllowedNamespaces struct {
	// From indicates how to select the namespaces that are allowed to use this resource.
	// Possible values are:
	// * All: All namespaces can use this resource.
	// * Selector: Namespaces that matches the selector can use this resource.
	// * Same: Only current namespace can use the resource.
	//
	// +optional
	// +kubebuilder:default=Same
	From *FromNamespaces `json:"from,omitempty"`

	// Selector must be specified when From is set to "Selector". In that case,
	// only the selected namespaces are allowed to use this resource.
	// This field is ignored for other values of "From".
	//
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
}

// FromNamespaces specifies namespace from which namespaces are allowed to use the resource.
//
// +kubebuilder:validation:Enum=All;Selector;Same
type FromNamespaces string

const (
	// NamespacesFromAll specifies that all namespaces can use the resource.
	NamespacesFromAll FromNamespaces = "All"

	// NamespacesFromSelector specifies that only the namespace that matches the selector can use the resource.
	NamespacesFromSelector FromNamespaces = "Selector"

	// NamespacesFromSame specifies that only the current namespace can use the resource.
	NamespacesFromSame FromNamespaces = "Same"
)
