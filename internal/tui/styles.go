package tui

import "github.com/charmbracelet/lipgloss"

var (
	statusActive    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	statusGone      = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	statusMerged    = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	statusBoth      = lipgloss.NewStyle().Foreground(lipgloss.Color("213"))
	statusProtected = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	currentMarker   = lipgloss.NewStyle().Bold(true)
	dim             = lipgloss.NewStyle().Faint(true)
)
