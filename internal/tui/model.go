// Package tui implements the interactive branch-management view.
package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"

	"github.com/USER/nathm/internal/branch"
	"github.com/USER/nathm/internal/git"
)

type sortMode int

const (
	sortStaleFirst sortMode = iota
	sortName
	sortAge
)

type confirmKind int

const (
	confirmNone confirmKind = iota
	confirmDelete
	confirmForceDelete
)

// Model is the TUI's root.
type Model struct {
	branches      []branch.Branch
	table         table.Model
	git           git.Git
	width, height int
	now           time.Time
	err           string
	selected      map[string]bool

	filter     textinput.Model
	filterOn   bool
	filterText string
	sortMode   sortMode
	staleOnly  bool

	// confirm modal state
	confirmKind    confirmKind
	confirmTargets []string
}

func NewModel(branches []branch.Branch, g git.Git) *Model {
	ti := textinput.New()
	ti.Placeholder = "filter..."
	ti.CharLimit = 80
	ti.Width = 40
	m := &Model{
		branches: branches,
		git:      g,
		now:      time.Now(),
		selected: make(map[string]bool),
		filter:   ti,
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

// Confirming reports whether the confirm modal is active.
func (m *Model) Confirming() bool { return m.confirmKind != confirmNone }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		// Confirm modal.
		if m.Confirming() {
			switch msg.String() {
			case "y", "Y":
				m.runConfirmedAction()
				return m, nil
			default:
				m.confirmKind = confirmNone
				m.confirmTargets = nil
				return m, nil
			}
		}

		// Filter input mode.
		if m.filterOn {
			switch msg.Type {
			case tea.KeyEnter, tea.KeyEsc:
				m.filterOn = false
				m.filter.Blur()
				if msg.Type == tea.KeyEsc {
					m.filterText = ""
					m.filter.SetValue("")
				} else {
					m.filterText = m.filter.Value()
				}
				m.rebuildTable()
				return m, nil
			}
			var cmd tea.Cmd
			m.filter, cmd = m.filter.Update(msg)
			return m, cmd
		}

		// Main keymap.
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
		case key.Matches(msg, keys.Filter):
			m.filterOn = true
			m.filter.SetValue(m.filterText)
			m.filter.Focus()
			return m, nil
		case key.Matches(msg, keys.Sort):
			m.sortMode = (m.sortMode + 1) % 3
			m.rebuildTable()
			return m, nil
		case key.Matches(msg, keys.StaleOnly):
			m.staleOnly = !m.staleOnly
			m.rebuildTable()
			return m, nil
		case key.Matches(msg, keys.Delete):
			m.beginDelete(false)
			return m, nil
		case key.Matches(msg, keys.ForceDelete):
			m.beginDelete(true)
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	header := lipgloss.NewStyle().Bold(true).Render("nathm — local branches")
	mid := m.table.View()
	footer := dim.Render("space:select / a:all / A:clear / /:filter / s:sort / p:stale-only / d:del / D:force / q:quit")
	if m.filterOn {
		footer = m.filter.View()
	}
	if m.err != "" {
		footer = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(m.err)
	}
	if m.Confirming() {
		mid = m.renderConfirm()
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, mid, footer)
}

func (m *Model) renderConfirm() string {
	verb := "Delete"
	if m.confirmKind == confirmForceDelete {
		verb = "FORCE DELETE"
	}
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("203")).
		Render(fmt.Sprintf("%s %d branch(es)?", verb, len(m.confirmTargets)))
	body := strings.Join(m.confirmTargets, "\n  ")
	help := dim.Render("y to confirm · any other key to cancel")
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
	return box.Render(lipgloss.JoinVertical(lipgloss.Left, title, "  "+body, "", help))
}

// targetsForAction returns the selection if non-empty, otherwise the cursor row.
func (m *Model) targetsForAction() []string {
	sel := m.Selected()
	if len(sel) > 0 {
		return sel
	}
	if name := m.CurrentName(); name != "" {
		return []string{name}
	}
	return nil
}

func (m *Model) beginDelete(force bool) {
	targets := m.targetsForAction()
	final := make([]string, 0, len(targets))
	for _, name := range targets {
		if b, ok := m.byName(name); ok && !b.Protected {
			final = append(final, name)
		}
	}
	if len(final) == 0 {
		m.err = "nothing to delete (all targets protected)"
		return
	}
	m.confirmTargets = final
	if force {
		m.confirmKind = confirmForceDelete
	} else {
		m.confirmKind = confirmDelete
	}
}

func (m *Model) byName(name string) (branch.Branch, bool) {
	for _, b := range m.branches {
		if b.Name == name {
			return b, true
		}
	}
	return branch.Branch{}, false
}

func (m *Model) runConfirmedAction() {
	force := m.confirmKind == confirmForceDelete
	var failed []string
	for _, name := range m.confirmTargets {
		if err := m.git.DeleteBranch(name, force); err != nil {
			failed = append(failed, name+": "+err.Error())
		}
	}
	// Drop deleted from the in-memory list so the UI updates.
	deleted := map[string]bool{}
	for _, n := range m.confirmTargets {
		deleted[n] = true
	}
	keep := make([]branch.Branch, 0, len(m.branches))
	for _, b := range m.branches {
		if !deleted[b.Name] {
			keep = append(keep, b)
		}
	}
	m.branches = keep
	m.clearSelection()
	if len(failed) > 0 {
		m.err = "errors: " + strings.Join(failed, "; ")
	} else {
		m.err = ""
	}
	m.confirmKind = confirmNone
	m.confirmTargets = nil
	m.rebuildTable()
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
	for _, b := range m.visibleBranches() {
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
	for _, b := range m.visibleBranches() {
		if !b.Protected && !b.IsCurrent {
			m.selected[b.Name] = true
		}
	}
}

func (m *Model) clearSelection() {
	m.selected = map[string]bool{}
}

// visibleBranches returns the branches that pass the current filter and stale-only toggle,
// sorted according to the current sort mode.
func (m *Model) visibleBranches() []branch.Branch {
	out := make([]branch.Branch, 0, len(m.branches))
	for _, b := range m.branches {
		if m.staleOnly && !b.IsStale() {
			continue
		}
		if m.filterText != "" && !strings.Contains(b.Name, m.filterText) {
			continue
		}
		out = append(out, b)
	}
	return m.applySort(out)
}

func (m *Model) applySort(in []branch.Branch) []branch.Branch {
	out := make([]branch.Branch, len(in))
	copy(out, in)
	switch m.sortMode {
	case sortName:
		sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	case sortAge:
		sort.SliceStable(out, func(i, j int) bool {
			return out[i].LastCommitTime.Before(out[j].LastCommitTime)
		})
	default: // sortStaleFirst
		sort.SliceStable(out, func(i, j int) bool {
			si, sj := stalePriority(out[i]), stalePriority(out[j])
			if si != sj {
				return si > sj
			}
			return out[i].LastCommitTime.Before(out[j].LastCommitTime)
		})
	}
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
