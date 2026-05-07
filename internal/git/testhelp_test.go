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

// newTestRepoWithRemote creates a test repo with a real (local-path) remote
// and a single tracking branch "feature" whose upstream is later deleted on
// the remote — useful for testing "gone" detection.
func newTestRepoWithRemote(t *testing.T) (repo, remote string) {
	t.Helper()
	remote = t.TempDir()
	runIn(t, remote, "git", "init", "-q", "--bare", "-b", "main")
	repo = newTestRepo(t)
	runIn(t, repo, "git", "remote", "add", "origin", remote)
	runIn(t, repo, "git", "push", "-q", "-u", "origin", "main")
	runIn(t, repo, "git", "checkout", "-q", "-b", "feature")
	writeFile(t, repo, "f.txt", "f")
	runIn(t, repo, "git", "add", "f.txt")
	runIn(t, repo, "git", "commit", "-q", "-m", "f")
	runIn(t, repo, "git", "push", "-q", "-u", "origin", "feature")
	runIn(t, repo, "git", "checkout", "-q", "main")
	// Delete the remote branch so the next fetch --prune marks local "feature" as gone.
	runIn(t, remote, "git", "branch", "-D", "feature")
	return repo, remote
}
