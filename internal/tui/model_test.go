package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/USER/nathm/internal/branch"
	tea "github.com/charmbracelet/bubbletea"
)

// tuiFakeGit lets the model run without a real repo.
type tuiFakeGit struct {
	deleted  []string
	renamed  [][2]string
	checkout []string
}

func (f *tuiFakeGit) IsRepo() bool                                 { return true }
func (f *tuiFakeGit) ListBranches() ([]branch.Branch, error)      { return nil, nil }
func (f *tuiFakeGit) AheadBehind(string, string) (int, int, error) { return 0, 0, nil }
func (f *tuiFakeGit) MergedInto(string, string) (bool, error)     { return false, nil }
func (f *tuiFakeGit) DeleteBranch(name string, force bool) error {
	f.deleted = append(f.deleted, name)
	return nil
}
func (f *tuiFakeGit) RenameBranch(oldN, newN string) error {
	f.renamed = append(f.renamed, [2]string{oldN, newN})
	return nil
}
func (f *tuiFakeGit) Checkout(name string) error {
	f.checkout = append(f.checkout, name)
	return nil
}
func (f *tuiFakeGit) FetchPrune() error { return nil }

func TestModel_View_RendersBranches(t *testing.T) {
	bs := []branch.Branch{
		{Name: "main", IsCurrent: true, LastCommitTime: time.Now()},
		{Name: "feature/foo", LastCommitTime: time.Now()},
		{Name: "stale", UpstreamGone: true, LastCommitTime: time.Now()},
	}
	m := NewModel(bs, nil) // nil git (we don't act yet)
	m.SetSize(120, 30)
	out := m.View()
	for _, name := range []string{"main", "feature/foo", "stale"} {
		if !strings.Contains(out, name) {
			t.Errorf("view missing %q:\n%s", name, out)
		}
	}
}

func TestModel_Selection_ToggleAndClearAll(t *testing.T) {
	bs := []branch.Branch{
		{Name: "a", LastCommitTime: time.Now()},
		{Name: "b", LastCommitTime: time.Now()},
		{Name: "c", LastCommitTime: time.Now()},
	}
	m := NewModel(bs, nil)
	m.SetSize(120, 30)

	// space toggles current row
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !m.IsSelected(m.CurrentName()) {
		t.Fatal("space should select current branch")
	}
	// space again deselects
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if m.IsSelected(m.CurrentName()) {
		t.Fatal("space should deselect")
	}
	// 'a' selects all visible
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if len(m.Selected()) != 3 {
		t.Fatalf("expected 3 selected, got %d", len(m.Selected()))
	}
	// 'A' clears
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	if len(m.Selected()) != 0 {
		t.Fatalf("expected 0 selected after clear, got %d", len(m.Selected()))
	}
}

func updateModel(m *Model, msg tea.Msg) (*Model, tea.Cmd) {
	mm, cmd := m.Update(msg)
	return mm.(*Model), cmd
}

func TestModel_Filter_ByName(t *testing.T) {
	bs := []branch.Branch{
		{Name: "feature-a", LastCommitTime: time.Now()},
		{Name: "feature-b", LastCommitTime: time.Now()},
		{Name: "main", IsCurrent: true, LastCommitTime: time.Now()},
	}
	m := NewModel(bs, nil)
	m.SetSize(120, 30)
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	for _, r := range "feature" {
		m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	out := m.View()
	if !strings.Contains(out, "feature-a") || !strings.Contains(out, "feature-b") {
		t.Errorf("expected feature-* visible:\n%s", out)
	}
	if strings.Contains(out, "main") {
		t.Errorf("main should be filtered out:\n%s", out)
	}
}

func TestModel_StaleOnlyToggle(t *testing.T) {
	bs := []branch.Branch{
		{Name: "active", LastCommitTime: time.Now()},
		{Name: "gone-1", UpstreamGone: true, LastCommitTime: time.Now()},
	}
	m := NewModel(bs, nil)
	m.SetSize(120, 30)
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	out := m.View()
	if strings.Contains(out, "active") {
		t.Errorf("active should be hidden in stale-only mode:\n%s", out)
	}
	if !strings.Contains(out, "gone-1") {
		t.Errorf("gone-1 should still be shown:\n%s", out)
	}
}

func TestModel_DeleteFlow(t *testing.T) {
	bs := []branch.Branch{
		{Name: "feature", LastCommitTime: time.Now()},
		{Name: "main", IsCurrent: true, Protected: true, LastCommitTime: time.Now()},
	}
	g := &tuiFakeGit{}
	m := NewModel(bs, g)
	m.SetSize(120, 30)
	// Press d.
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if !m.Confirming() {
		t.Fatal("expected confirm modal active")
	}
	// Press y — should call DeleteBranch.
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if len(g.deleted) != 1 || g.deleted[0] != "feature" {
		t.Fatalf("expected delete of feature, got %v", g.deleted)
	}
}

func TestModel_RenameFlow(t *testing.T) {
	bs := []branch.Branch{
		{Name: "old-name", LastCommitTime: time.Now()},
	}
	g := &tuiFakeGit{}
	m := NewModel(bs, g)
	m.SetSize(120, 30)

	// press r to start rename
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if !m.Renaming() {
		t.Fatal("expected rename mode")
	}
	// Use the public helper to set the rename value (deterministic vs typing).
	m.SetRenameValue("new-name")
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})

	if len(g.renamed) != 1 || g.renamed[0] != [2]string{"old-name", "new-name"} {
		t.Fatalf("expected rename old→new, got %v", g.renamed)
	}
}

func TestModel_CheckoutFlow(t *testing.T) {
	bs := []branch.Branch{
		{Name: "main", IsCurrent: true, Protected: true, LastCommitTime: time.Now()},
		{Name: "feature", LastCommitTime: time.Now()},
	}
	g := &tuiFakeGit{}
	m := NewModel(bs, g)
	m.SetSize(120, 30)
	// Move cursor to feature.
	m.SetCursorByName("feature")
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

	if len(g.checkout) != 1 || g.checkout[0] != "feature" {
		t.Fatalf("expected checkout of feature, got %v", g.checkout)
	}
}
