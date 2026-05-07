package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/USER/nathm/internal/branch"
	tea "github.com/charmbracelet/bubbletea"
)

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
