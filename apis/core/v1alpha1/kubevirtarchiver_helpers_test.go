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
	"reflect"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// The absence of a scheduler is the only thing that marks CBTBackup as the resident
// half. Adding one would silently turn the loop back into a CronJob.
func TestCBTBackupOptionsCarriesNoScheduler(t *testing.T) {
	tp := reflect.TypeOf(CBTBackupOptions{})
	for i := 0; i < tp.NumField(); i++ {
		if strings.Contains(strings.ToLower(tp.Field(i).Name), "schedul") &&
			tp.Field(i).Name != "RetentionSchedule" {
			t.Fatalf("CBTBackupOptions must not carry a scheduler, found field %q", tp.Field(i).Name)
		}
	}
}

func TestGetCBTIntervalFallsBackToTenMinutes(t *testing.T) {
	cases := map[string]struct {
		archiver KubeVirtArchiver
		want     time.Duration
	}{
		"no cbtBackup at all": {
			archiver: KubeVirtArchiver{},
			want:     10 * time.Minute,
		},
		"cbtBackup without interval": {
			archiver: KubeVirtArchiver{Spec: KubeVirtArchiverSpec{CBTBackup: &CBTBackupOptions{}}},
			want:     10 * time.Minute,
		},
		"explicit interval wins": {
			archiver: KubeVirtArchiver{Spec: KubeVirtArchiverSpec{CBTBackup: &CBTBackupOptions{
				Interval: &metav1.Duration{Duration: 90 * time.Second},
			}}},
			want: 90 * time.Second,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := tc.archiver.GetCBTInterval(); got != tc.want {
				t.Errorf("GetCBTInterval() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestUpsertVirtualMachineStatusReplacesInPlaceAndKeepsOrder(t *testing.T) {
	s := KubeVirtArchiverStatus{}
	s.UpsertVirtualMachineStatus(VirtualMachineArchiverStatus{Namespace: "demo", Name: "a"})
	s.UpsertVirtualMachineStatus(VirtualMachineArchiverStatus{Namespace: "demo", Name: "b"})
	s.UpsertVirtualMachineStatus(VirtualMachineArchiverStatus{
		Namespace:                 "demo",
		Name:                      "a",
		ChangedBlockTrackingReady: true,
	})

	if len(s.VirtualMachines) != 2 {
		t.Fatalf("got %d entries, want 2", len(s.VirtualMachines))
	}
	if s.VirtualMachines[0].Name != "a" || !s.VirtualMachines[0].ChangedBlockTrackingReady {
		t.Errorf("entry for a was not updated in place: %+v", s.VirtualMachines[0])
	}
	if s.VirtualMachines[1].Name != "b" {
		t.Errorf("order changed: %+v", s.VirtualMachines)
	}
}

// Same name in a different namespace is a different VirtualMachine.
func TestUpsertVirtualMachineStatusKeysOnNamespaceAndName(t *testing.T) {
	s := KubeVirtArchiverStatus{}
	s.UpsertVirtualMachineStatus(VirtualMachineArchiverStatus{Namespace: "demo", Name: "vm"})
	s.UpsertVirtualMachineStatus(VirtualMachineArchiverStatus{Namespace: "other", Name: "vm"})

	if len(s.VirtualMachines) != 2 {
		t.Fatalf("got %d entries, want 2", len(s.VirtualMachines))
	}
}

func TestRemoveVirtualMachineStatusDropsOnlyTheNamedVM(t *testing.T) {
	s := KubeVirtArchiverStatus{}
	s.UpsertVirtualMachineStatus(VirtualMachineArchiverStatus{Namespace: "demo", Name: "a"})
	s.UpsertVirtualMachineStatus(VirtualMachineArchiverStatus{Namespace: "demo", Name: "b"})

	s.RemoveVirtualMachineStatus("demo", "a")

	if len(s.VirtualMachines) != 1 || s.VirtualMachines[0].Name != "b" {
		t.Fatalf("got %+v, want only b", s.VirtualMachines)
	}
}

// A VM that needs a restart must not read as ready; the reason is what tells the
// user to restart it, since the controller deliberately will not.
func TestAllVirtualMachinesReadyIsFalseWhenAnyVMNeedsARestart(t *testing.T) {
	s := KubeVirtArchiverStatus{}
	s.UpsertVirtualMachineStatus(VirtualMachineArchiverStatus{
		Namespace: "demo", Name: "a", ChangedBlockTrackingReady: true,
		ChangedBlockTrackingState: "Enabled",
	})
	s.UpsertVirtualMachineStatus(VirtualMachineArchiverStatus{
		Namespace: "demo", Name: "b",
		ChangedBlockTrackingState: "PendingRestart",
		Reason:                    ReasonVMCBTPendingRestart,
	})

	if s.AllVirtualMachinesReady() {
		t.Error("AllVirtualMachinesReady() = true while a VM is PendingRestart")
	}
}

// Matching no VirtualMachine is a selector typo, not a healthy archiver.
func TestAllVirtualMachinesReadyIsFalseWhenNothingMatched(t *testing.T) {
	s := KubeVirtArchiverStatus{}
	if s.AllVirtualMachinesReady() {
		t.Error("AllVirtualMachinesReady() = true with no matched VirtualMachines")
	}
}

// The generated CRD file name is derived from the plural; a mismatch only shows up
// as a panic at runtime.
func TestCustomResourceDefinitionResolvesTheEmbeddedCRD(t *testing.T) {
	if crd := (KubeVirtArchiver{}).CustomResourceDefinition(); crd.V1 == nil {
		t.Fatal("no v1 CustomResourceDefinition embedded for kubevirtarchivers")
	}
}
