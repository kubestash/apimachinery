package apis

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Driver string

const (
	DriverRestic Driver = "Restic"
	DriverWalG   Driver = "WalG"
)

type TypedObjectReference struct {
	core.TypedLocalObjectReference
	Namespace string `json:"namespace,omitempty"`
}

type VolumeSource struct {
	core.VolumeSource
	VolumeClaimTemplate *core.PersistentVolumeClaimTemplate `json:"volumeClaimTemplate,omitempty"`
}

type FailurePolicy string

const (
	FailurePolicyFail  FailurePolicy = "Fail"
	FailurePolicyRetry FailurePolicy = "Retry"
)

type RetryConfig struct {
	MaxRetry int32  `json:"maxRetry,omitempty"`
	Delay    string `json:"delay,omitempty"`
}

type ParameterDefinition struct {
	Name     string `json:"name,omitempty"`
	Usage    string `json:"usage,omitempty"`
	Required bool   `json:"required,omitempty"`
	Default  string `json:"default,omitempty"`
}

type UsagePolicy struct {
	AllowedNamespaces AllowedNamespaces `json:"allowedNamespaces,omitempty"`
}

// FromNamespaces specifies namespace from which Secret Engines may be attached to a
// VaultServer.
//
// +kubebuilder:validation:Enum=All;Selector;Same
type FromNamespaces string

const (
	// Secret Engines in all namespaces may be attached to this VaultServer.
	NamespacesFromAll FromNamespaces = "All"
	// Only Secret Engines in namespaces selected by the selector may be attached to
	// this VaultServer.
	NamespacesFromSelector FromNamespaces = "Selector"
	// Only Secret Engines in the same namespace as the VaultServer may be attached to this
	// VaultServer.
	NamespacesFromSame FromNamespaces = "Same"
)

// SecretEngineNamespaces indicate which namespaces Secret Engines should be selected from.
type AllowedNamespaces struct {
	// From indicates where Secret Engines will be selected for this VaultServer. Possible
	// values are:
	// * All: Secret Engines in all namespaces may be used by this VaultServer.
	// * Selector: Secret Engines in namespaces selected by the selector may be used by
	//   this VaultServer.
	// * Same: Only Secret Engines in the same namespace may be used by this VaultServer.
	//
	// +optional
	// +kubebuilder:default=Same
	From *FromNamespaces `json:"from,omitempty" protobuf:"bytes,1,opt,name=from,casttype=FromNamespaces"`

	// Selector must be specified when From is set to "Selector". In that case,
	// only Secret Engines in Namespaces matching this Selector will be selected by this
	// VaultServer. This field is ignored for other values of "From".
	//
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,2,opt,name=selector"`
}
