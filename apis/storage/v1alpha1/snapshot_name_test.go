package v1alpha1

import "testing"

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
