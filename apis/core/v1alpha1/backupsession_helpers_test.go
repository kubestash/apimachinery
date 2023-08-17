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

	"kubestash.dev/apimachinery/apis/storage/v1alpha1"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	cutil "kmodules.xyz/client-go/conditions"
)

func TestBackupSessionPhaseBasedOnSnapshotPhase(t *testing.T) {
	cond := kmapi.Condition{
		Type:   TypeMetricsPushed,
		Status: metav1.ConditionTrue,
		Reason: ReasonSuccessfullyPushedMetrics,
	}

	tests := []struct {
		name          string
		backupSession *BackupSession
		expectedPhase BackupSessionPhase
	}{
		{
			name: "BackupSession should be Pending if all Snapshots are Pending",
			backupSession: getSampleBackupSession(func(b *BackupSession) {
				b.Status.Snapshots = []SnapshotStatus{
					{
						Name:       "manifest",
						Phase:      v1alpha1.SnapshotPending,
						Repository: "gcs-repo",
					},
					{
						Name:       "shard-0",
						Phase:      v1alpha1.SnapshotPending,
						Repository: "s3-repo",
					},
					{
						Name:       "shard-1",
						Phase:      v1alpha1.SnapshotPending,
						Repository: "s3-repo",
					},
				}
			}),
			expectedPhase: BackupSessionPending,
		},
		{
			name: "BackupSession should be Running if any Snapshot is Running",
			backupSession: getSampleBackupSession(func(b *BackupSession) {
				b.Status.Snapshots = []SnapshotStatus{
					{
						Name:       "manifest",
						Phase:      v1alpha1.SnapshotPending,
						Repository: "gcs-repo",
					},
					{
						Name:       "shard-0",
						Phase:      v1alpha1.SnapshotRunning,
						Repository: "s3-repo",
					},
					{
						Name:       "shard-1",
						Phase:      v1alpha1.SnapshotFailed,
						Repository: "s3-repo",
					},
					{
						Name:       "shard-2",
						Phase:      v1alpha1.SnapshotSucceeded,
						Repository: "s3-repo",
					},
				}
			}),
			expectedPhase: BackupSessionRunning,
		},
		{
			name: "BackupSession should be Running if any Snapshot is not completed",
			backupSession: getSampleBackupSession(func(b *BackupSession) {
				b.Status.Snapshots = []SnapshotStatus{
					{
						Name:       "manifest",
						Phase:      v1alpha1.SnapshotPending,
						Repository: "gcs-repo",
					},
					{
						Name:       "shard-0",
						Phase:      v1alpha1.SnapshotSucceeded,
						Repository: "s3-repo",
					},
					{
						Name:       "shard-1",
						Phase:      v1alpha1.SnapshotFailed,
						Repository: "s3-repo",
					},
					{
						Name:       "shard-2",
						Phase:      v1alpha1.SnapshotSucceeded,
						Repository: "s3-repo",
					},
				}
			}),
			expectedPhase: BackupSessionRunning,
		},
		{
			name: "BackupSession should be Running if all Snapshot completed but final step not executed",
			backupSession: getSampleBackupSession(func(b *BackupSession) {
				b.Status.Snapshots = []SnapshotStatus{
					{
						Name:       "manifest",
						Phase:      v1alpha1.SnapshotSucceeded,
						Repository: "gcs-repo",
					},
					{
						Name:       "shard-0",
						Phase:      v1alpha1.SnapshotSucceeded,
						Repository: "s3-repo",
					},
					{
						Name:       "shard-1",
						Phase:      v1alpha1.SnapshotFailed,
						Repository: "s3-repo",
					},
					{
						Name:       "shard-2",
						Phase:      v1alpha1.SnapshotSucceeded,
						Repository: "s3-repo",
					},
				}
			}),
			expectedPhase: BackupSessionRunning,
		},
		{
			name: "BackupSession should be Failed if any Snapshot Failed and all are completed",
			backupSession: getSampleBackupSession(func(b *BackupSession) {
				b.Status.Snapshots = []SnapshotStatus{
					{
						Name:       "manifest",
						Phase:      v1alpha1.SnapshotSucceeded,
						Repository: "gcs-repo",
					},
					{
						Name:       "shard-0",
						Phase:      v1alpha1.SnapshotSucceeded,
						Repository: "s3-repo",
					},
					{
						Name:       "shard-1",
						Phase:      v1alpha1.SnapshotFailed,
						Repository: "s3-repo",
					},
				}

				b.Status.Conditions = cutil.SetCondition(b.Status.Conditions, cond)
			}),
			expectedPhase: BackupSessionFailed,
		},
		{
			name: "BackupSession should be Failed if all Snapshots Failed",
			backupSession: getSampleBackupSession(func(b *BackupSession) {
				b.Status.Snapshots = []SnapshotStatus{
					{
						Name:       "manifest",
						Phase:      v1alpha1.SnapshotFailed,
						Repository: "gcs-repo",
					},
					{
						Name:       "shard-0",
						Phase:      v1alpha1.SnapshotFailed,
						Repository: "s3-repo",
					},
					{
						Name:       "shard-1",
						Phase:      v1alpha1.SnapshotFailed,
						Repository: "s3-repo",
					},
				}

				b.Status.Conditions = cutil.SetCondition(b.Status.Conditions, cond)
			}),
			expectedPhase: BackupSessionFailed,
		},
		{
			name: "BackupSession should be Successful if all Snapshots Succeeded",
			backupSession: getSampleBackupSession(func(b *BackupSession) {
				b.Status.Snapshots = []SnapshotStatus{
					{
						Name:       "manifest",
						Phase:      v1alpha1.SnapshotSucceeded,
						Repository: "gcs-repo",
					},
					{
						Name:       "shard-0",
						Phase:      v1alpha1.SnapshotSucceeded,
						Repository: "s3-repo",
					},
					{
						Name:       "shard-1",
						Phase:      v1alpha1.SnapshotSucceeded,
						Repository: "s3-repo",
					},
				}

				b.Status.Conditions = cutil.SetCondition(b.Status.Conditions, cond)
			}),

			expectedPhase: BackupSessionSucceeded,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedPhase, test.backupSession.CalculatePhase())
		})
	}
}

func TestBackupSessionPhaseFailedIfFailedToEnsureSnapshots(t *testing.T) {
	cond := kmapi.Condition{
		Type:   TypeSnapshotsEnsured,
		Status: metav1.ConditionFalse,
		Reason: ReasonFailedToEnsureSnapshots,
	}

	finalStep := kmapi.Condition{
		Type:   TypeMetricsPushed,
		Status: metav1.ConditionTrue,
		Reason: ReasonSuccessfullyPushedMetrics,
	}

	bs := getSampleBackupSession(func(b *BackupSession) {
		b.Status.Conditions = cutil.SetCondition(b.Status.Conditions, cond)
		b.Status.Conditions = cutil.SetCondition(b.Status.Conditions, finalStep)
	})

	assert.Equal(t, BackupSessionFailed, bs.CalculatePhase())
}

func TestBackupSessionPhaseSkippedIfSkippedConditionTrue(t *testing.T) {
	cond := kmapi.Condition{
		Type:   TypeBackupSkipped,
		Status: metav1.ConditionTrue,
		Reason: ReasonSkippedTakingNewBackup,
	}

	bs := getSampleBackupSession(func(b *BackupSession) {
		b.Status.Conditions = cutil.SetCondition(b.Status.Conditions, cond)
	})

	assert.Equal(t, BackupSessionSkipped, bs.CalculatePhase())
}

func TestBackupSessionPhaseFailedIfSessionHistoryCleanupFailed(t *testing.T) {
	cond := kmapi.Condition{
		Type:   TypeSessionHistoryCleaned,
		Status: metav1.ConditionFalse,
		Reason: ReasonFailedToCleanSessionHistory,
	}

	finalStep := kmapi.Condition{
		Type:   TypeMetricsPushed,
		Status: metav1.ConditionTrue,
		Reason: ReasonSuccessfullyPushedMetrics,
	}

	bs := getSampleBackupSession(func(b *BackupSession) {
		b.Status.Conditions = cutil.SetCondition(b.Status.Conditions, cond)
		b.Status.Conditions = cutil.SetCondition(b.Status.Conditions, finalStep)
	})

	assert.Equal(t, BackupSessionFailed, bs.CalculatePhase())
}

func TestBackupSessionPhaseFailedIfRetentionPolicyFailedToApply(t *testing.T) {
	finalStep := kmapi.Condition{
		Type:   TypeMetricsPushed,
		Status: metav1.ConditionTrue,
		Reason: ReasonSuccessfullyPushedMetrics,
	}

	bs := getSampleBackupSession(func(b *BackupSession) {
		b.Status.Conditions = cutil.SetCondition(b.Status.Conditions, finalStep)
		b.Status.RetentionPolicies = append(b.Status.RetentionPolicies, RetentionPolicyApplyStatus{
			Phase: RetentionPolicyFailedToApply,
		})
	})

	assert.Equal(t, BackupSessionFailed, bs.CalculatePhase())
}

func TestBackupSessionPhaseFailedIfBackupExecutorFailedToEnsure(t *testing.T) {
	cond := kmapi.Condition{
		Type:   TypeBackupExecutorEnsured,
		Status: metav1.ConditionFalse,
		Reason: ReasonFailedToEnsureBackupExecutor,
	}

	finalStep := kmapi.Condition{
		Type:   TypeMetricsPushed,
		Status: metav1.ConditionTrue,
		Reason: ReasonSuccessfullyPushedMetrics,
	}

	bs := getSampleBackupSession(func(b *BackupSession) {
		b.Status.Conditions = cutil.SetCondition(b.Status.Conditions, cond)
		b.Status.Conditions = cutil.SetCondition(b.Status.Conditions, finalStep)
	})

	assert.Equal(t, BackupSessionFailed, bs.CalculatePhase())
}

func TestBackupSessionPhaseSucceededIfAllCriteriaSatisfied(t *testing.T) {
	cond := kmapi.Condition{
		Type:   TypeSessionHistoryCleaned,
		Status: metav1.ConditionTrue,
		Reason: ReasonSuccessfullyCleanedSessionHistory,
	}

	finalStep := kmapi.Condition{
		Type:   TypeMetricsPushed,
		Status: metav1.ConditionTrue,
		Reason: ReasonSuccessfullyPushedMetrics,
	}

	bs := getSampleBackupSession(func(b *BackupSession) {
		b.Status.Snapshots = []SnapshotStatus{
			{
				Name:  "demo-snapshot",
				Phase: v1alpha1.SnapshotSucceeded,
			},
		}

		b.Status.Conditions = cutil.SetCondition(b.Status.Conditions, cond)
		b.Status.Conditions = cutil.SetCondition(b.Status.Conditions, finalStep)
	})

	assert.Equal(t, BackupSessionSucceeded, bs.CalculatePhase())
}

func getSampleBackupSession(transformFuncs ...func(configuration *BackupSession)) *BackupSession {
	bs := &BackupSession{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-backupsession",
			Namespace: "default",
			CreationTimestamp: metav1.Time{
				Time: time.Now(),
			},
		},
	}

	executorCond := kmapi.Condition{
		Type:   TypeBackupExecutorEnsured,
		Status: metav1.ConditionTrue,
		Reason: ReasonSuccessfullyEnsuredBackupExecutor,
	}

	snapCond := kmapi.Condition{
		Type:   TypeSnapshotsEnsured,
		Status: metav1.ConditionTrue,
		Reason: ReasonSuccessfullyEnsuredSnapshots,
	}

	bs.Status.Conditions = cutil.SetCondition(bs.Status.Conditions, executorCond)
	bs.Status.Conditions = cutil.SetCondition(bs.Status.Conditions, snapCond)

	for _, fn := range transformFuncs {
		fn(bs)
	}

	return bs
}
