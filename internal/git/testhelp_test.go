package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// newTestRepo creates a fresh git repo in a tempdir with one commit on `main`.
// Caller gets the directory path. Cleanup is automatic via t.TempDir().
func newTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runIn(t, dir, "git", "init", "-q", "-b", "main")
	runIn(t, dir, "git", "config", "user.email", "nathm-test@example.com")
	runIn(t, dir, "git", "config", "user.name", "nathm test")
	runIn(t, dir, "git", "config", "commit.gpgsign", "false")
	runIn(t, dir, "git", "commit", "--allow-empty", "-q", "-m", "init")
	return dir
}

func runIn(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v in %s: %v\n%s", name, args, dir, err, out)
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
