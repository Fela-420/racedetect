package main

import (
	"path/filepath"
	"testing"
)

// violationsIn runs the detector on one file and returns the violation count.
func violationsIn(t *testing.T, path string) int {
	t.Helper()
	v, _, _, err := analyzeFile(path)
	if err != nil {
		t.Fatalf("could not parse %s: %v", path, err)
	}
	return v
}

// globCorpus returns all .go files in corpus/<dir>.
func globCorpus(t *testing.T, dir string) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join("..", "..", "corpus", dir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatalf("no .go files found in corpus/%s", dir)
	}
	return files
}

// Files that should be clean must report zero violations.
func TestCleanCorpusHasNoViolations(t *testing.T) {
	for _, dir := range []string{"clean", "defer_ok"} {
		for _, f := range globCorpus(t, dir) {
			if v := violationsIn(t, f); v != 0 {
				t.Errorf("%s: expected 0 violations, got %d", f, v)
			}
		}
	}
}

// Files that contain a missing unlock must report at least one violation.
func TestMissingUnlockCorpusIsCaught(t *testing.T) {
	for _, f := range globCorpus(t, "missing_unlock") {
		if v := violationsIn(t, f); v == 0 {
			t.Errorf("%s: expected at least 1 violation, got 0", f)
		}
	}
}

// The real-world bugs from CockroachDB, gRPC, and Moby must all be caught.
func TestRealWorldBugsAreCaught(t *testing.T) {
	for _, f := range globCorpus(t, "real-world") {
		if v := violationsIn(t, f); v == 0 {
			t.Errorf("%s: expected at least 1 violation, got 0", f)
		}
	}
}
