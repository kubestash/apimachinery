package v1alpha1

import (
	core "k8s.io/api/core/v1"
)

// DeletionPolicy specifies what to do if a resource is deleted
// +kubebuilder:validation:Enum=Delete;WipeOUt
type DeletionPolicy string

const (
	DeletionPolicyDelete  DeletionPolicy = "Delete"
	DeletionPolicyWipeOut DeletionPolicy = "WipeOut"
)

type Backend struct {
	Local *LocalSpec      `json:"local,omitempty"`
	S3    *S3Spec         `json:"s3,omitempty"`
	GCS   *GCSSpec        `json:"gcs,omitempty"`
	Azure *AzureSpec      `json:"azure,omitempty"`
	Swift *SwiftSpec      `json:"swift,omitempty"`
	B2    *B2Spec         `json:"b2,omitempty"`
	Rest  *RestServerSpec `json:"rest,omitempty"`
}

type LocalSpec struct {
	core.VolumeSource `json:",inline"`
	MountPath         string `json:"mountPath,omitempty"`
	SubPath           string `json:"subPath,omitempty"`
}

type S3Spec struct {
	Endpoint string `json:"endpoint,omitempty"`
	Bucket   string `json:"bucket,omitempty"`
	Prefix   string `json:"prefix,omitempty"`
	Region   string `json:"region,omitempty"`
	Secret   string `json:"secret,omitempty"`
}

type GCSSpec struct {
	Bucket         string `json:"bucket,omitempty"`
	Prefix         string `json:"prefix,omitempty"`
	MaxConnections int64  `json:"maxConnections,omitempty"`
	Secret         string `json:"secret,omitempty"`
}

type AzureSpec struct {
	Container      string `json:"container,omitempty"`
	Prefix         string `json:"prefix,omitempty"`
	MaxConnections int64  `json:"maxConnections,omitempty"`
	Secret         string `json:"secret,omitempty"`
}

type SwiftSpec struct {
	Container string `json:"container,omitempty"`
	Prefix    string `json:"prefix,omitempty"`
	Secret    string `json:"secret,omitempty"`
}

type B2Spec struct {
	Bucket         string `json:"bucket,omitempty"`
	Prefix         string `json:"prefix,omitempty"`
	MaxConnections int64  `json:"maxConnections,omitempty"`
	Secret         string `json:"secret,omitempty"`
}

type RestServerSpec struct {
	URL    string `json:"url,omitempty"`
	Secret string `json:"secret,omitempty"`
}
