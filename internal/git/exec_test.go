package git

import (
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
