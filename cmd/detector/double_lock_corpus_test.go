package main

import "testing"

// A second Lock on the same receiver while the first is held must be
// flagged even when the first Lock has a defer Unlock.
func TestDeferCoveredDoubleLockIsCaught(t *testing.T) {
	for _, f := range globCorpus(t, "double_lock") {
		if v := violationsIn(t, f); v == 0 {
			t.Errorf("%s: expected a double-lock violation, got 0", f)
		}
	}
}
