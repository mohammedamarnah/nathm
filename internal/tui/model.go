// Package tui implements the interactive branch-management view.
package tui

import (
	"fmt"
	"sort"
	"time"

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
}

func NewModel(branches []branch.Branch, g git.Git) *Model {
	m := &Model{
		branches: branches,
		git:      g,
		now:      time.Now(),
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
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	header := lipgloss.NewStyle().Bold(true).Render("nathm — local branches")
	footer := dim.Render("q quit · (more keys coming)")
	if m.err != "" {
		footer = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(m.err)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, m.table.View(), footer)
}

func (m *Model) rebuildTable() {
	cols := []table.Column{
		{Title: "", Width: 2},   // current marker
		{Title: "Branch", Width: 32},
		{Title: "Status", Width: 12},
		{Title: "Age", Width: 12},
		{Title: "↑ ↓", Width: 8},
		{Title: "Last commit", Width: 40},
	}
	rows := make([]table.Row, 0, len(m.branches))
	for _, b := range branchesSorted(m.branches) {
		marker := " "
		if b.IsCurrent {
			marker = "*"
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

// branchesSorted: stale-first, then by last-commit age desc.
func branchesSorted(in []branch.Branch) []branch.Branch {
	out := make([]branch.Branch, len(in))
	copy(out, in)
	sort.SliceStable(out, func(i, j int) bool {
		si, sj := stalePriority(out[i]), stalePriority(out[j])
		if si != sj {
			return si > sj
		}
		// older = older time = smaller — but we want oldest first, so:
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
