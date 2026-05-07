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

	"github.com/mohammedamarnah/nathm/internal/branch"
	"github.com/mohammedamarnah/nathm/internal/git"
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

	// rename input state
	renameOn     bool
	renameSource string
	renameInput  textinput.Model

	showHelp bool
}

func NewModel(branches []branch.Branch, g git.Git) *Model {
	ti := textinput.New()
	ti.Placeholder = "filter..."
	ti.CharLimit = 80
	ti.Width = 40

	ri := textinput.New()
	ri.CharLimit = 200
	ri.Width = 40

	m := &Model{
		branches:    branches,
		git:         g,
		now:         time.Now(),
		selected:    make(map[string]bool),
		filter:      ti,
		renameInput: ri,
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

// Renaming reports whether the rename input is active.
func (m *Model) Renaming() bool { return m.renameOn }

// SetRenameValue sets the rename input value (used by tests for deterministic input).
func (m *Model) SetRenameValue(s string) {
	m.renameInput.SetValue(s)
}

// SetCursorByName moves the table cursor to the row for the named branch.
func (m *Model) SetCursorByName(name string) {
	visible := m.visibleBranches()
	for i, b := range visible {
		if b.Name == name {
			m.table.SetCursor(i)
			return
		}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		// Rename input mode has highest priority.
		if m.renameOn {
			switch msg.Type {
			case tea.KeyEsc:
				m.cancelRename()
				return m, nil
			case tea.KeyEnter:
				m.commitRename()
				return m, nil
			}
			var cmd tea.Cmd
			m.renameInput, cmd = m.renameInput.Update(msg)
			return m, cmd
		}

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
		case key.Matches(msg, keys.Help):
			m.showHelp = !m.showHelp
			return m, nil
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
		case key.Matches(msg, keys.Rename):
			m.beginRename()
			return m, nil
		case key.Matches(msg, keys.Checkout):
			m.doCheckout()
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
	footer := dim.Render("space:select / a:all / A:clear / /:filter / s:sort / p:stale-only / d:del / D:force / r:rename / c:checkout / ?:help / q:quit")
	if m.filterOn {
		footer = m.filter.View()
	}
	if m.renameOn {
		footer = "rename: " + m.renameInput.View() + "  (enter:save · esc:cancel)"
	}
	if m.err != "" {
		footer = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(m.err)
	}
	if m.showHelp {
		mid = m.renderHelp()
	}
	if m.Confirming() {
		mid = m.renderConfirm()
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, mid, footer)
}

func (m *Model) renderHelp() string {
	lines := []string{
		"nathm — keybindings",
		"",
		"  ↑/↓ or j/k    navigate",
		"  space         toggle selection",
		"  a             select all visible",
		"  A             clear selection",
		"  enter / d     delete (cursor or selected)",
		"  D             force delete",
		"  r             rename (cursor only)",
		"  c             checkout",
		"  /             filter by name",
		"  s             cycle sort",
		"  p             toggle stale-only",
		"  ?             toggle this help",
		"  q             quit",
	}
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
	return box.Render(strings.Join(lines, "\n"))
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
	succeeded := map[string]bool{}
	for _, name := range m.confirmTargets {
		if err := m.git.DeleteBranch(name, force); err != nil {
			failed = append(failed, name+": "+err.Error())
			continue
		}
		succeeded[name] = true
	}
	// Drop only the actually-deleted branches; failed ones stay in the list
	// so the user can retry (e.g. with D for force).
	keep := make([]branch.Branch, 0, len(m.branches))
	for _, b := range m.branches {
		if !succeeded[b.Name] {
			keep = append(keep, b)
		}
	}
	m.branches = keep
	for name := range succeeded {
		delete(m.selected, name)
	}
	if len(failed) > 0 {
		m.err = "errors: " + strings.Join(failed, "; ")
	} else {
		m.err = ""
	}
	m.confirmKind = confirmNone
	m.confirmTargets = nil
	m.rebuildTable()
}

func (m *Model) beginRename() {
	name := m.CurrentName()
	if name == "" {
		return
	}
	if b, ok := m.byName(name); ok && b.Protected {
		m.err = "cannot rename protected branch"
		return
	}
	m.renameOn = true
	m.renameSource = name
	m.renameInput.SetValue(name)
	m.renameInput.Focus()
}

func (m *Model) commitRename() {
	newName := strings.TrimSpace(m.renameInput.Value())
	if newName == "" || newName == m.renameSource {
		m.cancelRename()
		return
	}
	if err := m.git.RenameBranch(m.renameSource, newName); err != nil {
		m.err = "rename failed: " + err.Error()
		m.cancelRename()
		return
	}
	for i := range m.branches {
		if m.branches[i].Name == m.renameSource {
			m.branches[i].Name = newName
		}
	}
	m.cancelRename()
	m.rebuildTable()
}

func (m *Model) cancelRename() {
	m.renameOn = false
	m.renameSource = ""
	m.renameInput.SetValue("")
	m.renameInput.Blur()
}

func (m *Model) doCheckout() {
	name := m.CurrentName()
	if name == "" {
		return
	}
	if err := m.git.Checkout(name); err != nil {
		m.err = "checkout failed: " + err.Error()
		return
	}
	for i := range m.branches {
		m.branches[i].IsCurrent = (m.branches[i].Name == name)
	}
	m.err = ""
	m.rebuildTable()
}

func (m *Model) rebuildTable() {
	// Preserve cursor across rebuilds by remembering the current row's branch name.
	var prevName string
	if oldRows := m.table.Rows(); len(oldRows) > 0 {
		if cur := m.table.Cursor(); cur >= 0 && cur < len(oldRows) && len(oldRows[cur]) >= 2 {
			prevName = oldRows[cur][1]
		}
	}

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

	if prevName != "" {
		for i, r := range rows {
			if len(r) >= 2 && r[1] == prevName {
				m.table.SetCursor(i)
				break
			}
		}
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
