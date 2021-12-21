/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
