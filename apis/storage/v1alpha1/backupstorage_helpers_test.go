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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

func TestBackupStorageCalculatePhaseWithObjectLock(t *testing.T) {
	initialized := kmapi.Condition{Type: TypeBackendInitialized, Status: metav1.ConditionTrue}
	lockVerified := kmapi.Condition{Type: TypeObjectLockVerified, Status: metav1.ConditionTrue}
	lockNotVerified := kmapi.Condition{Type: TypeObjectLockVerified, Status: metav1.ConditionFalse}

	tests := []struct {
		name       string
		objectLock *ObjectLockSpec
		conditions []kmapi.Condition
		expected   BackupStoragePhase
	}{
		{
			name:       "Object Lock not required",
			conditions: []kmapi.Condition{initialized},
			expected:   BackupStorageReady,
		},
		{
			name:       "Object Lock not required and not verified",
			objectLock: &ObjectLockSpec{RetentionPeriod: &metav1.Duration{Duration: time.Hour}},
			conditions: []kmapi.Condition{initialized, lockNotVerified},
			expected:   BackupStorageReady,
		},
		{
			name:       "Object Lock required and verified",
			objectLock: &ObjectLockSpec{Required: true},
			conditions: []kmapi.Condition{initialized, lockVerified},
			expected:   BackupStorageReady,
		},
		{
			name:       "Object Lock required but not verified",
			objectLock: &ObjectLockSpec{Required: true},
			conditions: []kmapi.Condition{initialized, lockNotVerified},
			expected:   BackupStorageNotReady,
		},
		{
			name:       "Object Lock mode required but not checked yet",
			objectLock: &ObjectLockSpec{Mode: ObjectLockModeCompliance},
			conditions: []kmapi.Condition{initialized},
			expected:   BackupStorageNotReady,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BackupStorage{
				Spec:   BackupStorageSpec{ObjectLock: tt.objectLock},
				Status: BackupStorageStatus{Conditions: tt.conditions},
			}
			assert.Equal(t, tt.expected, bs.CalculatePhase())
		})
	}
}

func TestBackupStorageObjectLockRetentionPeriod(t *testing.T) {
	day := 24 * time.Hour
	enabled := &ObjectLockStatus{Enabled: true, Mode: ObjectLockModeCompliance, DefaultRetention: &metav1.Duration{Duration: day}}

	tests := []struct {
		name     string
		spec     *ObjectLockSpec
		status   *ObjectLockStatus
		expected time.Duration
	}{
		{
			name:     "Object Lock not detected",
			expected: 0,
		},
		{
			name:     "Object Lock not enabled on the bucket",
			spec:     &ObjectLockSpec{RetentionPeriod: &metav1.Duration{Duration: 7 * day}},
			status:   &ObjectLockStatus{Enabled: false},
			expected: 0,
		},
		{
			name:     "Bucket default retention",
			status:   enabled,
			expected: day,
		},
		{
			name:     "Retention period from spec",
			spec:     &ObjectLockSpec{RetentionPeriod: &metav1.Duration{Duration: 7 * day}},
			status:   enabled,
			expected: 7 * day,
		},
		{
			name:     "Retention extension disabled",
			spec:     &ObjectLockSpec{DisableRetentionExtension: true},
			status:   enabled,
			expected: 0,
		},
		{
			name:     "No default retention and no retention period",
			status:   &ObjectLockStatus{Enabled: true},
			expected: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BackupStorage{
				Spec:   BackupStorageSpec{ObjectLock: tt.spec},
				Status: BackupStorageStatus{ObjectLock: tt.status},
			}
			assert.Equal(t, tt.expected, bs.ObjectLockRetentionPeriod())
		})
	}
}

func TestBackupStorageObjectLockMode(t *testing.T) {
	bs := &BackupStorage{Status: BackupStorageStatus{ObjectLock: &ObjectLockStatus{Enabled: true, Mode: ObjectLockModeGovernance}}}
	assert.Equal(t, ObjectLockModeGovernance, bs.ObjectLockMode())

	bs.Spec.ObjectLock = &ObjectLockSpec{Mode: ObjectLockModeCompliance}
	assert.Equal(t, ObjectLockModeCompliance, bs.ObjectLockMode())
}
