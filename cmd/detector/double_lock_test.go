package main

import (
	"os"
	"strings"
	"testing"
)

// captureStdout runs fn and returns everything it printed to stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = f
	fn()
	os.Stdout = orig
	f.Close()

	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// moby7559 must be reported as a double-lock (the loop re-enters the same
// Lock before any Unlock), and must NOT be reported as a missing-unlock.
func TestMoby7559IsDoubleLock(t *testing.T) {
	out := captureStdout(t, func() {
		if _, _, _, err := analyzeFile("../../corpus/real-world/moby7559.go"); err != nil {
			t.Errorf("analyzeFile failed: %v", err)
		}
	})

	if !strings.Contains(out, "[VIOLATION]") {
		t.Fatalf("expected a violation, got output:\n%s", out)
	}
	if !strings.Contains(out, "loop re-enters this same Lock call") {
		t.Errorf("expected a double-lock finding, got output:\n%s", out)
	}
	if strings.Contains(out, "reached WITHOUT unlock") {
		t.Errorf("expected no missing-unlock finding, got output:\n%s", out)
	}
}
