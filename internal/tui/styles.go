package tui

import "github.com/charmbracelet/lipgloss"

// Soft peach/coral accent — bright but desaturated, used for the outer frame
// and the title.
var frameAccent = lipgloss.Color("#A0D8B3")

var (
	statusActive    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	statusGone      = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	statusMerged    = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	statusBoth      = lipgloss.NewStyle().Foreground(lipgloss.Color("213"))
	statusProtected = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	currentMarker   = lipgloss.NewStyle().Bold(true)
	dim             = lipgloss.NewStyle().Faint(true)

	frameStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(frameAccent).
			Padding(1, 2)

	headerStyle = lipgloss.NewStyle().Foreground(frameAccent).Bold(true)
)

// Frame overhead in characters: 1 border + 2 padding on each side horizontally,
// 1 border + 1 padding on each side vertically. Used for sizing.
const (
	frameHPad = 6
	frameVPad = 4
)
