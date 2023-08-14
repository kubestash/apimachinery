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

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	cutil "kmodules.xyz/client-go/conditions"
)

func TestRestoreSessionPhaseBasedOnComponentsPhase(t *testing.T) {
	tests := []struct {
		name           string
		restoreSession *RestoreSession
		expectedPhase  RestorePhase
	}{
		{
			name: "RestoreSession should be Pending if no component is initialized",
			restoreSession: sampleRestoreSession(func(r *RestoreSession) {
				r.Status.TotalComponents = 4
			}),

			expectedPhase: RestorePending,
		},
		{
			name: "RestoreSession should be Running if any component is Running",
			restoreSession: sampleRestoreSession(func(r *RestoreSession) {
				r.Status.TotalComponents = 4
				r.Status.Components = map[string]ComponentRestoreStatus{
					"manifest": {
						Phase: RestoreRunning,
					},
					"configserver": {
						Phase: RestorePending,
					},
					"shard-0": {
						Phase: RestorePending,
					},
					"shard-1": {
						Phase: RestorePending,
					},
				}
			}),

			expectedPhase: RestoreRunning,
		},
		{
			name: "RestoreSession should be Running if any component is not completed",
			restoreSession: sampleRestoreSession(func(r *RestoreSession) {
				r.Status.TotalComponents = 4
				r.Status.Components = map[string]ComponentRestoreStatus{
					"manifest": {
						Phase: RestoreSucceeded,
					},
					"configserver": {
						Phase: RestoreFailed,
					},
					"shard-0": {
						Phase: RestoreFailed,
					},
					"shard-1": {
						Phase: RestorePending,
					},
				}
			}),

			expectedPhase: RestoreRunning,
		},
		{
			name: "RestoreSession should be Failed if any component Failed",
			restoreSession: sampleRestoreSession(func(r *RestoreSession) {
				setPostRestoreHooksExecutionSucceededConditionToTrue(r)
				r.Status.TotalComponents = 4
				r.Status.Components = map[string]ComponentRestoreStatus{
					"manifest": {
						Phase: RestoreFailed,
					},
					"configserver": {
						Phase: RestoreFailed,
					},
					"shard-0": {
						Phase: RestoreSucceeded,
					},
					"shard-1": {
						Phase: RestoreSucceeded,
					},
				}
			}),

			expectedPhase: RestoreFailed,
		},
		{
			name: "RestoreSession should be Failed if all components Failed",
			restoreSession: sampleRestoreSession(func(r *RestoreSession) {
				setPostRestoreHooksExecutionSucceededConditionToTrue(r)
				r.Status.TotalComponents = 4
				r.Status.Components = map[string]ComponentRestoreStatus{
					"manifest": {
						Phase: RestoreFailed,
					},
					"configserver": {
						Phase: RestoreFailed,
					},
					"shard-0": {
						Phase: RestoreFailed,
					},
					"shard-1": {
						Phase: RestoreFailed,
					},
				}
			}),

			expectedPhase: RestoreFailed,
		},
		{
			name: "RestoreSession should be Succeeded if all components Succeeded",
			restoreSession: sampleRestoreSession(func(r *RestoreSession) {
				setPostRestoreHooksExecutionSucceededConditionToTrue(r)
				r.Status.TotalComponents = 4
				r.Status.Components = map[string]ComponentRestoreStatus{
					"manifest": {
						Phase: RestoreSucceeded,
					},
					"configserver": {
						Phase: RestoreSucceeded,
					},
					"shard-0": {
						Phase: RestoreSucceeded,
					},
					"shard-1": {
						Phase: RestoreSucceeded,
					},
				}
			}),

			expectedPhase: RestoreSucceeded,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedPhase, test.restoreSession.CalculatePhase())
		})
	}
}

func TestRestoreSessionPhaseIsFailedIfPreRestoreHooksExecutionSucceededConditionIsFalse(t *testing.T) {
	rs := sampleRestoreSession(func(r *RestoreSession) {
		r.Status.Conditions = append(r.Status.Conditions,
			kmapi.Condition{
				Type:   TypePreRestoreHooksExecutionSucceeded,
				Status: metav1.ConditionFalse,
				Reason: ReasonFailedToExecutePreRestoreHooks,
			},
			kmapi.Condition{
				Type:   TypeMetricsPushed,
				Status: metav1.ConditionTrue,
				Reason: ReasonSuccessfullyPushedMetrics,
			},
		)
	})

	assert.Equal(t, RestoreFailed, rs.CalculatePhase())
}

func TestRestoreSessionPhaseIsFailedIfPostRestoreHooksExecutionSucceededConditionIsFalse(t *testing.T) {
	rs := sampleRestoreSession(func(r *RestoreSession) {
		r.Status.Conditions = append(r.Status.Conditions,
			kmapi.Condition{
				Type:   TypePostRestoreHooksExecutionSucceeded,
				Status: metav1.ConditionFalse,
				Reason: ReasonFailedToExecutePostRestoreHooks,
			},
			kmapi.Condition{
				Type:   TypeMetricsPushed,
				Status: metav1.ConditionTrue,
				Reason: ReasonSuccessfullyPushedMetrics,
			},
		)
	})

	assert.Equal(t, RestoreFailed, rs.CalculatePhase())
}

func TestRestoreSessionPhaseIsFailedIfRestoreExecutorEnsuredConditionIsFalse(t *testing.T) {
	rs := sampleRestoreSession(func(r *RestoreSession) {
		r.Status.Conditions = append(r.Status.Conditions,
			kmapi.Condition{
				Type:   TypeRestoreExecutorEnsured,
				Status: metav1.ConditionFalse,
				Reason: ReasonFailedToEnsureRestoreExecutor,
			},
			kmapi.Condition{
				Type:   TypeMetricsPushed,
				Status: metav1.ConditionTrue,
				Reason: ReasonSuccessfullyPushedMetrics,
			},
		)
	})

	assert.Equal(t, RestoreFailed, rs.CalculatePhase())
}

func TestRestoreSessionPhaseIsFailedIfDeadlineExceededConditionIsTrue(t *testing.T) {
	rs := sampleRestoreSession(func(r *RestoreSession) {
		r.Status.Conditions = append(r.Status.Conditions,
			kmapi.Condition{
				Type:   TypeDeadlineExceeded,
				Status: metav1.ConditionTrue,
				Reason: ReasonFailedToCompleteWithinDeadline,
			},
			kmapi.Condition{
				Type:   TypeMetricsPushed,
				Status: metav1.ConditionTrue,
				Reason: ReasonSuccessfullyPushedMetrics,
			},
		)
	})

	assert.Equal(t, RestoreFailed, rs.CalculatePhase())
}

func TestRestoreSessionPhaseIsRunningIfPostRestoreHooksNotExecuted(test *testing.T) {
	rs := sampleRestoreSession(func(r *RestoreSession) {
		r.Status.Components = map[string]ComponentRestoreStatus{
			"manifest": {
				Phase: RestoreSucceeded,
			},
			"configserver": {
				Phase: RestoreSucceeded,
			},
			"shard-0": {
				Phase: RestoreSucceeded,
			},
			"shard-1": {
				Phase: RestoreSucceeded,
			},
		}
	})
	assert.Equal(test, RestoreRunning, rs.CalculatePhase())
}

func sampleRestoreSession(transformFuncs ...func(*RestoreSession)) *RestoreSession {
	rs := &RestoreSession{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-mysql-restore",
			Namespace: "demo",
		},
		Spec: RestoreSessionSpec{
			Target: &kmapi.TypedObjectReference{
				APIGroup: "appcatalog.appscode.com",
				Kind:     "AppBinding",
				Name:     "sample-mysql",
			},
			DataSource: &RestoreDataSource{
				Snapshot: "sample-mysql-backup-1561974001",
			},
			Addon: &AddonInfo{
				Name: "stash-mysql-90.31",
				Tasks: []TaskReference{
					{
						Name: "ManifestRestore",
					},
					{
						Name: "LogicalBackupRestore",
					},
				},
			},
			Hooks: &RestoreHooks{
				PreRestore: []HookInfo{
					{
						Name: "cleanup-old-databases",
						HookTemplate: &kmapi.ObjectReference{
							Name:      "mysql-query-executor",
							Namespace: "demo",
						},
					},
				},
				PostRestore: []HookInfo{
					{
						Name: "run-migration",
						HookTemplate: &kmapi.ObjectReference{
							Name:      "mysql-query-executor",
							Namespace: "demo",
						},
					},
				},
			},
		},
	}

	for _, fn := range transformFuncs {
		fn(rs)
	}

	return rs
}

func setPostRestoreHooksExecutionSucceededConditionToTrue(rs *RestoreSession) {
	newCond := kmapi.Condition{
		Type:    TypePostRestoreHooksExecutionSucceeded,
		Status:  metav1.ConditionTrue,
		Reason:  ReasonSuccessfullyExecutedPostRestoreHooks,
		Message: "Post-Restore Hooks have been executed successfully.",
	}
	rs.Status.Conditions = cutil.SetCondition(rs.Status.Conditions, newCond)
}
