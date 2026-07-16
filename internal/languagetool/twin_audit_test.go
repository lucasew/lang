// Package languagetool is the root of the LT-shaped 1:1 port tree.
// Twin completeness is enforced by scripts/check_lt_test_twins.py (refactree).
package languagetool_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

// TestJavaGoTestTwins fails until every LanguageTool @Test has a hand-ported Go twin.
// Uses refactree via scripts/check_lt_test_twins.py — does not generate tests.
func TestJavaGoTestTwins(t *testing.T) {
	root := repoRoot(t)
	script := filepath.Join(root, "scripts", "check_lt_test_twins.py")
	if _, err := os.Stat(script); err != nil {
		t.Fatalf("missing auditor script: %v", err)
	}
	if _, err := exec.LookPath("rft"); err != nil {
		t.Fatalf("rft (refactree) required on PATH: %v", err)
	}
	cmd := exec.Command("python3", script)
	cmd.Dir = root
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	t.Logf("\n%s", out)
	if err != nil {
		t.Fatalf("Java↔Go test twin audit failed: %v\n(hand-port missing tests under internal/languagetool/; see scripts/check_lt_test_twins.py)", err)
	}
}
