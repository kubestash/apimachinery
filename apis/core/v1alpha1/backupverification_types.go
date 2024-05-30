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
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"

	"kubestash.dev/apimachinery/apis"
)

const (
	ResourceKindBackupVerification     = "BackupVerification"
	ResourceSingularBackupVerification = "backupverification"
	ResourcePluralBackupVerification   = "backupverifications"
)

// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=backupverifications,singular=backupverification,shortName=bv,categories={kubestash,appscode,all}
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// BackupVerification defines how to verify a backup of a target application.
// This is a namespaced CRD. However, you can use it from other namespaces. You can control which
// namespaces are allowed to use it using the `usagePolicy` section.
type BackupVerification struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec BackupVerificationSpec `json:"spec,omitempty"`
}

// BackupVerificationSpec specifies specification for the verification function and type.
type BackupVerificationSpec struct {
	// Function specifies the function name which will be used to verify the backup.
	// +optional
	Function string `json:"function,omitempty"`

	// UsagePolicy specifies a policy of how this BackupVerification will be used. For example,
	// you can use `allowedNamespaces` policy to restrict the usage of this BackupVerification to particular namespaces.
	// This field is optional. If you don't provide the usagePolicy, then it can be used only from the current namespace.
	// +optional
	UsagePolicy *apis.UsagePolicy `json:"usagePolicy,omitempty"`

	// Params defines a list of parameters that is used by the BackupVerification to execute its logic.
	// +optional
	Params []apis.ParameterDefinition `json:"params,omitempty"`

	// Target indicates the target application where the data will be restored for backup verification.
	// +optional
	Target *kmapi.TypedObjectReference `json:"target,omitempty"`

	// VolumeMounts specifies the mount path of the volumes specified in the VolumeTemplate section.
	// These volumes will be mounted directly on the Job created by KubeStash operator.
	// If the volume type is VolumeClaimTemplate, then KubeStash operator is responsible for creating the volume.
	// +optional
	VolumeMounts []core.VolumeMount `json:"volumeMounts,omitempty"`

	// VolumeTemplate specifies a list of volume templates that is used by the respective backup verification
	// Job to execute its logic.
	// +optional
	VolumeTemplate []VolumeTemplate `json:"volumeTemplate,omitempty"`

	// Type indicate the types of verifier that will verify the backup.
	// Valid values are:
	// - "RestoreOnly": KubeStash will create a RestoreSession with the tasks provided in BackupConfiguration's verificationStrategies section.
	// - "File": KubeStash will restore the data and then create a job to check if the files exist or not. This type is recommended for workload backup verification.
	// - "Query": KubeStash operator will restore data and then create a job to run the queries. This type is recommended for database backup verification.
	// - "Script": KubeStash operator will restore data and then create a job to run the script. This type is recommended for database backup verification.
	Type VerificationType `json:"type,omitempty"`

	// File specifies the file paths information whose existence will be checked for backup verification.
	// +optional
	File *FileVerifierSpec `json:"file,omitempty"`

	// Query specifies the queries to be run to verify backup.
	// +optional
	Query []QueryVerifierSpec `json:"query,omitempty"`

	// Script specifies the script to be run to verify backup.
	// +optional
	Script *ScriptVerifierSpec `json:"script,omitempty"`
}

// VolumeTemplate specifies the name, usage, and the source of volume that will be used by the
// backup verification job to execute its logic.
type VolumeTemplate struct {
	// Name specifies the name of the volume
	Name string `json:"name"`

	// Usage specifies the usage of the volume.
	// +optional
	Usage string `json:"usage,omitempty"`

	// Source specifies the source of this volume.
	Source *apis.VolumeSource `json:"source,omitempty"`
}

// VerificationType specifies the type of verifier that will verify the backup
// +kubebuilder:validation:Enum=RestoreOnly;File;Query;Script
type VerificationType string

const (
	RestoreOnlyVerificationType VerificationType = "RestoreOnly"
	FileVerificationType        VerificationType = "File"
	QueryVerificationType       VerificationType = "Query"
	ScriptVerificationType      VerificationType = "Script"
)

// FileVerifierSpec defines the file paths information whose existence will be checked from verifier job.
type FileVerifierSpec struct {
	// Paths specifies the list of paths whose existence will be checked.
	// These paths must be absolute paths.
	Paths []string `json:"paths,omitempty"`
}

// QueryVerifierSpec defines the queries to be run from verifier job.
type QueryVerifierSpec struct {
	// Statement specifies the query statement.
	Statement string `json:"statement,omitempty"`

	// ExpectedOutput specifies the expected output for the query.
	// If the ExpectedOutput doesn't match, the verifier job will be completed with exit code 1.
	ExpectedOutput string `json:"expectedOutput,omitempty"`
}

// ScriptVerifierSpec defines the script location in verifier job and the args to be provided with the script.
type ScriptVerifierSpec struct {
	// Location specifies the absolute path of the script file's location.
	Location string `json:"location,omitempty"`

	// Args specifies the arguments to be provided with the script.
	Args []string `json:"args,omitempty"`
}

//+kubebuilder:object:root=true

// BackupVerificationList contains a list of BackupVerification
type BackupVerificationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupVerification `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackupVerification{}, &BackupVerificationList{})
}
