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
)

func TestSnapshotComponentsPhase(t *testing.T) {
	tests := []struct {
		name          string
		snapshot      *Snapshot
		expectedPhase SnapshotPhase
	}{
		{
			name:          "Snapshot should be Pending if no component is initialized",
			snapshot:      sampleSnapshot(2, map[string]Component{}),
			expectedPhase: SnapshotPending,
		},
		{
			name: "Snapshot should be Running if any component is in progress",
			snapshot: sampleSnapshot(2, map[string]Component{
				"dump": {Phase: ComponentPhaseRunning},
			}),
			expectedPhase: SnapshotRunning,
		},
		{
			name: "Snapshot should be Running if any component Failed",
			snapshot: sampleSnapshot(2, map[string]Component{
				"manifest": {Phase: ComponentPhaseFailed},
				"dump":     {Phase: ComponentPhaseRunning},
			}),
			expectedPhase: SnapshotFailed,
		},
		{
			name: "Snapshot should be Failed if any component Failed and all components reported",
			snapshot: sampleSnapshot(2, map[string]Component{
				"manifest": {Phase: ComponentPhaseSucceeded},
				"dump":     {Phase: ComponentPhaseFailed},
			}),
			expectedPhase: SnapshotFailed,
		},
		{
			name: "Snapshot should be Failed if a component Failed and the rest were never reported",
			snapshot: sampleSnapshot(2, map[string]Component{
				"dump": {Phase: ComponentPhaseFailed},
			}),
			expectedPhase: SnapshotFailed,
		},
		{
			name: "Snapshot should be Succeeded if all components Succeeded",
			snapshot: sampleSnapshot(2, map[string]Component{
				"manifest": {Phase: ComponentPhaseSucceeded},
				"dump":     {Phase: ComponentPhaseSucceeded},
			}),
			expectedPhase: SnapshotSucceeded,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedPhase, test.snapshot.GetComponentsPhase())
		})
	}
}

// nolint:unparam
func sampleSnapshot(totalComponents int32, components map[string]Component) *Snapshot {
	return &Snapshot{
		Status: SnapshotStatus{
			TotalComponents: totalComponents,
			Components:      components,
		},
	}
}
