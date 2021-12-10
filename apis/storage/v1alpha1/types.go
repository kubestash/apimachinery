package v1alpha1

import (
	core "k8s.io/api/core/v1"
)

type DeletionPolicy string

const (
	DeletionPolicyDelete  DeletionPolicy = "Delete"
	DeletionPolicyWipeOut DeletionPolicy = "WipeOut"
)

type Backend struct {
	Local *LocalSpec      `json:"local,omitempty" protobuf:"bytes,2,opt,name=local"`
	S3    *S3Spec         `json:"s3,omitempty" protobuf:"bytes,3,opt,name=s3"`
	GCS   *GCSSpec        `json:"gcs,omitempty" protobuf:"bytes,4,opt,name=gcs"`
	Azure *AzureSpec      `json:"azure,omitempty" protobuf:"bytes,5,opt,name=azure"`
	Swift *SwiftSpec      `json:"swift,omitempty" protobuf:"bytes,6,opt,name=swift"`
	B2    *B2Spec         `json:"b2,omitempty" protobuf:"bytes,7,opt,name=b2"`
	Rest  *RestServerSpec `json:"rest,omitempty" protobuf:"bytes,8,opt,name=rest"`
}

type LocalSpec struct {
	core.VolumeSource `json:",inline" protobuf:"bytes,1,opt,name=volumeSource"`
	MountPath         string `json:"mountPath,omitempty" protobuf:"bytes,2,opt,name=mountPath"`
	SubPath           string `json:"subPath,omitempty" protobuf:"bytes,3,opt,name=subPath"`
}

type S3Spec struct {
	Endpoint string `json:"endpoint,omitempty" protobuf:"bytes,1,opt,name=endpoint"`
	Bucket   string `json:"bucket,omitempty" protobuf:"bytes,2,opt,name=bucket"`
	Prefix   string `json:"prefix,omitempty" protobuf:"bytes,3,opt,name=prefix"`
	Region   string `json:"region,omitempty" protobuf:"bytes,4,opt,name=region"`
	Secret   string `json:"secret,omitempty"`
}

type GCSSpec struct {
	Bucket         string `json:"bucket,omitempty" protobuf:"bytes,1,opt,name=bucket"`
	Prefix         string `json:"prefix,omitempty" protobuf:"bytes,2,opt,name=prefix"`
	MaxConnections int64  `json:"maxConnections,omitempty" protobuf:"varint,3,opt,name=maxConnections"`
	Secret         string `json:"secret,omitempty"`
}

type AzureSpec struct {
	Container      string `json:"container,omitempty" protobuf:"bytes,1,opt,name=container"`
	Prefix         string `json:"prefix,omitempty" protobuf:"bytes,2,opt,name=prefix"`
	MaxConnections int64  `json:"maxConnections,omitempty" protobuf:"varint,3,opt,name=maxConnections"`
	Secret         string `json:"secret,omitempty"`
}

type SwiftSpec struct {
	Container string `json:"container,omitempty" protobuf:"bytes,1,opt,name=container"`
	Prefix    string `json:"prefix,omitempty" protobuf:"bytes,2,opt,name=prefix"`
	Secret    string `json:"secret,omitempty"`
}

type B2Spec struct {
	Bucket         string `json:"bucket,omitempty" protobuf:"bytes,1,opt,name=bucket"`
	Prefix         string `json:"prefix,omitempty" protobuf:"bytes,2,opt,name=prefix"`
	MaxConnections int64  `json:"maxConnections,omitempty" protobuf:"varint,3,opt,name=maxConnections"`
	Secret         string `json:"secret,omitempty"`
}

type RestServerSpec struct {
	URL    string `json:"url,omitempty" protobuf:"bytes,1,opt,name=url"`
	Secret string `json:"secret,omitempty"`
}
