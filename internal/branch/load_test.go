package branch

import (
	"errors"
	"testing"
	"time"
)

// fakeGit is a minimal Git stand-in for testing the loader.
type fakeGit struct {
	branches []Branch
	merged   map[string]bool
	ab       map[string][2]int
}

func (f *fakeGit) IsRepo() bool { return true }
func (f *fakeGit) ListBranches() ([]Branch, error) {
	cp := make([]Branch, len(f.branches))
	copy(cp, f.branches)
	return cp, nil
}
func (f *fakeGit) AheadBehind(b, base string) (int, int, error) {
	v, ok := f.ab[b]
	if !ok {
		return 0, 0, errors.New("no ab for " + b)
	}
	return v[0], v[1], nil
}
func (f *fakeGit) MergedInto(b, base string) (bool, error) {
	return f.merged[b], nil
}

func TestLoad_BasicFlow(t *testing.T) {
	t0 := time.Unix(1700000000, 0)
	f := &fakeGit{
		branches: []Branch{
			{Name: "main", IsCurrent: true, LastCommitTime: t0},
			{Name: "feature", Upstream: "origin/feature", Ahead: 1, LastCommitTime: t0},
			{Name: "stale", UpstreamGone: true, LastCommitTime: t0},
			{Name: "no-upstream", LastCommitTime: t0},
		},
		merged: map[string]bool{
			"feature":     false,
			"stale":       true,
			"no-upstream": false,
		},
		ab: map[string][2]int{
			"no-upstream": {2, 0}, // backfill candidate
		},
	}
	cfg := LoadConfig{
		BaseBranches:      []string{"main", "master"},
		ProtectedPatterns: []string{"release/*"},
	}
	bs, err := Load(f, cfg)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]Branch{}
	for _, b := range bs {
		byName[b.Name] = b
	}
	if !byName["main"].Protected {
		t.Error("main (base+current) should be protected")
	}
	if byName["stale"].Status() != StatusBoth {
		t.Errorf("stale should be StatusBoth, got %v", byName["stale"].Status())
	}
	if byName["no-upstream"].Ahead != 2 {
		t.Errorf("no-upstream Ahead should be backfilled to 2, got %d", byName["no-upstream"].Ahead)
	}
	if byName["feature"].MergedIntoBase {
		t.Error("feature should not be merged")
	}
}
