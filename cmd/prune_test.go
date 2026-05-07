package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Runs `nathm prune --yes` against a tempdir repo with a known-stale
// (merged) branch, and verifies the branch was deleted.
func TestPrune_Yes_DeletesMergedBranch(t *testing.T) {
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "nathm")
	if out, err := exec.Command("go", "build", "-o", bin, "github.com/mohammedamarnah/nathm").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}

	repo := t.TempDir()
	for _, c := range [][]string{
		{"git", "init", "-q", "-b", "main"},
		{"git", "config", "user.email", "x@x"},
		{"git", "config", "user.name", "x"},
		{"git", "config", "commit.gpgsign", "false"},
		{"git", "commit", "--allow-empty", "-q", "-m", "init"},
		{"git", "checkout", "-q", "-b", "merged-feature"},
		{"git", "commit", "--allow-empty", "-q", "-m", "f"},
		{"git", "checkout", "-q", "main"},
		{"git", "merge", "--no-ff", "-q", "-m", "merge", "merged-feature"},
	} {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", c, err, out)
		}
	}

	cfgDir := filepath.Join(tmp, "cfg")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(bin, "prune", "--yes")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+cfgDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("prune: %v\nstderr: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "merged-feature") {
		t.Fatalf("expected merged-feature in output, got %q", stdout.String())
	}

	// Verify branch is gone.
	listCmd := exec.Command("git", "branch", "--list")
	listCmd.Dir = repo
	out, err := listCmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "merged-feature") {
		t.Fatalf("merged-feature still present:\n%s", out)
	}
}
