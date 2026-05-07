package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRename_LocalBranch(t *testing.T) {
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "nathm")
	if out, err := exec.Command("go", "build", "-o", bin, "github.com/USER/nathm").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	repo := t.TempDir()
	for _, c := range [][]string{
		{"git", "init", "-q", "-b", "main"},
		{"git", "config", "user.email", "x@x"},
		{"git", "config", "user.name", "x"},
		{"git", "config", "commit.gpgsign", "false"},
		{"git", "commit", "--allow-empty", "-q", "-m", "init"},
		{"git", "branch", "old"},
	} {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", c, err, out)
		}
	}
	cfgDir := filepath.Join(tmp, "cfg")
	_ = os.MkdirAll(cfgDir, 0o755)

	cmd := exec.Command(bin, "rename", "old", "new")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+cfgDir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("rename: %v\n%s", err, stderr.String())
	}

	out, _ := exec.Command("git", "-C", repo, "branch", "--list").Output()
	if !strings.Contains(string(out), "new") || strings.Contains(string(out), "old") {
		t.Fatalf("expected old→new, got %s", out)
	}
}
