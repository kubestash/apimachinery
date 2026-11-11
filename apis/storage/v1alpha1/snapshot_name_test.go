package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

// A KubeVirtArchiver's resident snapshot never completes, so a transient
// metadata upload failure must not present a healthy archiver as broken.
func TestRunningKubeVirtIncrementalIgnoresATransientMetadataFailure(t *testing.T) {
	s := &Snapshot{}
	s.Spec.Type = BackupTypeIncremental
	s.Spec.AppRef = kmapi.TypedObjectReference{APIGroup: "kubevirt.io", Kind: "VirtualMachine", Name: "vm"}
	s.Status.TotalComponents = 1
	s.Status.Components = map[string]Component{"volume-data": {Phase: ComponentPhaseRunning}}
	s.Status.Conditions = []kmapi.Condition{{
		Type: TypeSnapshotMetadataUploaded, Status: metav1.ConditionFalse, Reason: "transient",
	}}
	if got := s.CalculatePhase(); got != SnapshotRunning {
		t.Fatalf("a running incremental must stay Running, got %q", got)
	}

	// Once archiving stops and the component is terminal, the ordinary rules
	// apply again and the failed upload does mean Failed.
	s.Status.Components["volume-data"] = Component{Phase: ComponentPhaseSucceeded}
	if got := s.CalculatePhase(); got != SnapshotFailed {
		t.Fatalf("a finished snapshot with a failed upload must be Failed, got %q", got)
	}
}

// Other archivers' incremental snapshots (e.g. KubeDB WAL/binlog archiving)
// keep the generic rules: a failed metadata upload means Failed even while
// the component is still Running.
func TestRunningNonKubeVirtIncrementalKeepsGenericPhaseRules(t *testing.T) {
	s := &Snapshot{}
	s.Spec.Type = BackupTypeIncremental
	s.Spec.AppRef = kmapi.TypedObjectReference{APIGroup: "kubedb.com", Kind: "Postgres", Name: "db"}
	s.Status.TotalComponents = 1
	s.Status.Components = map[string]Component{"wal": {Phase: ComponentPhaseRunning}}
	s.Status.Conditions = []kmapi.Condition{{
		Type: TypeSnapshotMetadataUploaded, Status: metav1.ConditionFalse, Reason: "transient",
	}}
	if got := s.CalculatePhase(); got != SnapshotFailed {
		t.Fatalf("a non-KubeVirt incremental must keep the generic rules, got %q", got)
	}
}
