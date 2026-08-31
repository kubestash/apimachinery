package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

// A BackupSession created by hand rarely ends in -<digits>, and every caller
// indexed the regex match without checking it, so the operator panicked while
// removing the finalizer and the session could never be deleted.
func TestSplitBackupSessionNameWithoutNumericSuffix(t *testing.T) {
	for _, tc := range []struct{ in, base, suffix string }{
		{"vmguest-backup-daily-1787641252", "vmguest-backup-daily", "1787641252"},
		{"guest-r2b", "guest-r2b", ""},
		{"guest-r1", "guest-r1", ""},
		{"plain", "plain", ""},
		{"", "", ""},
	} {
		base, suffix := SplitBackupSessionName(tc.in)
		if base != tc.base || suffix != tc.suffix {
			t.Errorf("SplitBackupSessionName(%q) = (%q,%q), want (%q,%q)", tc.in, base, suffix, tc.base, tc.suffix)
		}
	}
}

// A resident archiver's snapshot never completes, so a transient metadata
// upload failure must not present a healthy archiver as broken.
func TestRunningIncrementalIgnoresATransientMetadataFailure(t *testing.T) {
	s := &Snapshot{}
	s.Spec.Type = BackupTypeIncremental
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
