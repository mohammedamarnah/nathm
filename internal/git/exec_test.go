package git

import (
	"sort"
	"strings"
	"testing"
)

func TestExec_IsRepo_TrueInsideRepo(t *testing.T) {
	dir := newTestRepo(t)
	g := NewExec(dir)
	if !g.IsRepo() {
		t.Fatalf("expected IsRepo() = true inside %s", dir)
	}
}

func TestExec_IsRepo_FalseOutsideRepo(t *testing.T) {
	dir := t.TempDir() // not initialized
	g := NewExec(dir)
	if g.IsRepo() {
		t.Fatalf("expected IsRepo() = false in non-repo dir")
	}
}

func TestExec_ListBranches_OneBranch(t *testing.T) {
	dir := newTestRepo(t)
	g := NewExec(dir)
	bs, err := g.ListBranches()
	if err != nil {
		t.Fatal(err)
	}
	if len(bs) != 1 || bs[0].Name != "main" || !bs[0].IsCurrent {
		t.Fatalf("got %+v", bs)
	}
}

func TestExec_ListBranches_MultipleBranches(t *testing.T) {
	dir := newTestRepo(t)
	runIn(t, dir, "git", "branch", "feature-a")
	runIn(t, dir, "git", "branch", "feature-b")
	g := NewExec(dir)
	bs, err := g.ListBranches()
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, len(bs))
	for i, b := range bs {
		names[i] = b.Name
	}
	sort.Strings(names)
	if got, want := strings.Join(names, ","), "feature-a,feature-b,main"; got != want {
		t.Fatalf("names = %q, want %q", got, want)
	}
}

func TestExec_AheadBehind(t *testing.T) {
	dir := newTestRepo(t)
	// add 2 commits on main
	writeFile(t, dir, "a.txt", "a")
	runIn(t, dir, "git", "add", "a.txt")
	runIn(t, dir, "git", "commit", "-q", "-m", "a")
	// branch off and add 1 commit
	runIn(t, dir, "git", "checkout", "-q", "-b", "feature")
	writeFile(t, dir, "b.txt", "b")
	runIn(t, dir, "git", "add", "b.txt")
	runIn(t, dir, "git", "commit", "-q", "-m", "b")
	// add another commit on main
	runIn(t, dir, "git", "checkout", "-q", "main")
	writeFile(t, dir, "c.txt", "c")
	runIn(t, dir, "git", "add", "c.txt")
	runIn(t, dir, "git", "commit", "-q", "-m", "c")

	g := NewExec(dir)
	ahead, behind, err := g.AheadBehind("feature", "main")
	if err != nil {
		t.Fatal(err)
	}
	if ahead != 1 || behind != 1 {
		t.Fatalf("ahead/behind = %d/%d, want 1/1", ahead, behind)
	}
}
