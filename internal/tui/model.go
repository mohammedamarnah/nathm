// Package tui implements the interactive branch-management view.
package tui

import (
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"

	"github.com/USER/nathm/internal/branch"
	"github.com/USER/nathm/internal/git"
)

// Model is the TUI's root.
type Model struct {
	branches []branch.Branch
	table    table.Model
	git      git.Git
	width    int
	height   int
	now      time.Time
	err      string
	selected map[string]bool
}

func NewModel(branches []branch.Branch, g git.Git) *Model {
	m := &Model{
		branches: branches,
		git:      g,
		now:      time.Now(),
		selected: make(map[string]bool),
	}
	m.rebuildTable()
	return m
}

func (m *Model) SetSize(w, h int) {
	m.width, m.height = w, h
	m.table.SetWidth(w)
	m.table.SetHeight(maxInt(h-3, 5))
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Select):
			m.toggleSelect(m.CurrentName())
			m.rebuildTable()
			return m, nil
		case key.Matches(msg, keys.SelectAll):
			m.selectAllVisible()
			m.rebuildTable()
			return m, nil
		case key.Matches(msg, keys.ClearAll):
			m.clearSelection()
			m.rebuildTable()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	header := lipgloss.NewStyle().Bold(true).Render("nathm — local branches")
	footer := dim.Render("space:select / a:all / A:clear / q:quit")
	if m.err != "" {
		footer = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(m.err)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, m.table.View(), footer)
}

func (m *Model) rebuildTable() {
	cols := []table.Column{
		{Title: "Sel", Width: 4},
		{Title: "Branch", Width: 32},
		{Title: "Status", Width: 12},
		{Title: "Age", Width: 12},
		{Title: "↑ ↓", Width: 8},
		{Title: "Last commit", Width: 40},
	}
	rows := make([]table.Row, 0, len(m.branches))
	for _, b := range branchesSorted(m.branches) {
		marker := "[ ]"
		if m.selected[b.Name] {
			marker = "[x]"
		}
		if b.IsCurrent {
			marker = " * "
		}
		rows = append(rows, table.Row{
			marker,
			b.Name,
			statusLabel(b),
			humanize.RelTime(b.LastCommitTime, m.now, "", "from now"),
			fmt.Sprintf("↑%d ↓%d", b.Ahead, b.Behind),
			truncate(b.LastCommitSubject, 40),
		})
	}
	m.table = table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
	)
	if m.width > 0 {
		m.SetSize(m.width, m.height)
	}
}

// CurrentName returns the branch name of the currently highlighted row.
func (m *Model) CurrentName() string {
	row := m.table.SelectedRow()
	if len(row) < 2 {
		return ""
	}
	return row[1]
}

// IsSelected reports whether the branch with the given name is selected.
func (m *Model) IsSelected(name string) bool {
	return m.selected[name]
}

// Selected returns the sorted list of selected branch names.
func (m *Model) Selected() []string {
	out := make([]string, 0, len(m.selected))
	for name, ok := range m.selected {
		if ok {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

func (m *Model) toggleSelect(name string) {
	if name == "" {
		return
	}
	if m.selected[name] {
		delete(m.selected, name)
	} else {
		m.selected[name] = true
	}
}

func (m *Model) selectAllVisible() {
	for _, b := range branchesSorted(m.branches) {
		if !b.Protected && !b.IsCurrent {
			m.selected[b.Name] = true
		}
	}
}

func (m *Model) clearSelection() {
	m.selected = map[string]bool{}
}

// branchesSorted: stale-first, then by last-commit age desc.
func branchesSorted(in []branch.Branch) []branch.Branch {
	out := make([]branch.Branch, len(in))
	copy(out, in)
	sort.SliceStable(out, func(i, j int) bool {
		si, sj := stalePriority(out[i]), stalePriority(out[j])
		if si != sj {
			return si > sj
		}
		return out[i].LastCommitTime.Before(out[j].LastCommitTime)
	})
	return out
}

func stalePriority(b branch.Branch) int {
	switch b.Status() {
	case branch.StatusBoth:
		return 3
	case branch.StatusGone:
		return 2
	case branch.StatusMerged:
		return 1
	default:
		return 0
	}
}

func statusLabel(b branch.Branch) string {
	if b.Protected {
		return "protected"
	}
	return b.Status().String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
