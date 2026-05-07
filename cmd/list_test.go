package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestList_Smoke(t *testing.T) {
	// Build the binary into the tempdir.
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "nathm")
	build := exec.Command("go", "build", "-o", bin, "github.com/mohammedamarnah/nathm")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}

	// Tempdir repo with two branches.
	repo := t.TempDir()
	for _, c := range [][]string{
		{"git", "init", "-q", "-b", "main"},
		{"git", "config", "user.email", "x@x"},
		{"git", "config", "user.name", "x"},
		{"git", "config", "commit.gpgsign", "false"},
		{"git", "commit", "--allow-empty", "-q", "-m", "init"},
		{"git", "branch", "feature"},
	} {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", c, err, out)
		}
	}

	// Isolate config to avoid touching the real user's config.
	cfgDir := filepath.Join(tmp, "cfg")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(bin, "list")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+cfgDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("nathm list: %v\nstderr: %s", err, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "main\t") || !strings.Contains(out, "feature\t") {
		t.Fatalf("missing branches in output: %q", out)
	}
}
