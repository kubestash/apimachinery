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
)

// TimeRange is a closed interval of time.
type TimeRange struct {
	Start metav1.Time `json:"start"`
	End   metav1.Time `json:"end"`
}

// BackupSessionRef summarizes a BackupSession.
type BackupSessionRef struct {
	Name    string `json:"name"`
	Session string `json:"session"`
	// Phase mirrors the BackupSession status phase.
	Phase string `json:"phase,omitempty"`
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// CompletionTime is StartTime plus the reported session duration.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
	// +optional
	Duration *metav1.Duration `json:"duration,omitempty"`
	// +optional
	Error string `json:"error,omitempty"`
}

// VerificationSessionRef summarizes a BackupVerificationSession.
type VerificationSessionRef struct {
	Name     string `json:"name"`
	Snapshot string `json:"snapshot"`
	// Phase mirrors the BackupVerificationSession status phase.
	Phase string `json:"phase,omitempty"`
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// +optional
	Duration *metav1.Duration `json:"duration,omitempty"`
}

// +kubebuilder:validation:Enum=ContinuousLog;Snapshot
type RPOBasis string

const (
	RPOBasisContinuousLog RPOBasis = "ContinuousLog"
	RPOBasisSnapshot      RPOBasis = "Snapshot"
)

// RPOStats reports the recovery point objective actually achieved.
type RPOStats struct {
	Basis RPOBasis `json:"basis,omitempty"`
	// Current is the age of the newest recoverable point.
	// +optional
	Current *metav1.Duration `json:"current,omitempty"`
	// WorstCase is the longest span without a recoverable point inside Window, including Current.
	// +optional
	WorstCase *metav1.Duration `json:"worstCase,omitempty"`
	Window    metav1.Duration  `json:"window"`
	// Target comes from the ui.kubestash.com/rpo-target annotation of the BackupConfiguration.
	// +optional
	Target *metav1.Duration `json:"target,omitempty"`
	// Breached is true when Current exceeds Target, or when there is a Target but no recoverable point.
	Breached bool `json:"breached"`
}

// +kubebuilder:validation:Enum=Restore;Verification
type RTOSource string

const (
	RTOSourceRestore      RTOSource = "Restore"
	RTOSourceVerification RTOSource = "Verification"
)

// RTOStats reports the recovery time measured from completed restores and verifications.
type RTOStats struct {
	// +optional
	LastMeasured *metav1.Duration `json:"lastMeasured,omitempty"`
	// +optional
	LastMeasuredSource RTOSource `json:"lastMeasuredSource,omitempty"`
	// +optional
	LastMeasuredAt *metav1.Time `json:"lastMeasuredAt,omitempty"`
	// Estimated is the mean of the most recent successful samples.
	// +optional
	Estimated *metav1.Duration `json:"estimated,omitempty"`
	Samples   int32            `json:"samples"`
}
