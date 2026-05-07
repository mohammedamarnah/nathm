package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Quit key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(key.WithKeys("q", "ctrl+c", "esc"), key.WithHelp("q", "quit")),
}
