package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Quit      key.Binding
	Select    key.Binding
	SelectAll key.Binding
	ClearAll  key.Binding
}

var keys = keyMap{
	Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c", "esc"), key.WithHelp("q", "quit")),
	Select:    key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "select")),
	SelectAll: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "select all")),
	ClearAll:  key.NewBinding(key.WithKeys("A"), key.WithHelp("A", "clear selection")),
}
