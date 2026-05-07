package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Quit        key.Binding
	Select      key.Binding
	SelectAll   key.Binding
	ClearAll    key.Binding
	Filter      key.Binding
	Sort        key.Binding
	StaleOnly   key.Binding
	Delete      key.Binding
	ForceDelete key.Binding
	Rename      key.Binding
	Checkout    key.Binding
}

var keys = keyMap{
	Quit:        key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Select:      key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "select")),
	SelectAll:   key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "select all")),
	ClearAll:    key.NewBinding(key.WithKeys("A"), key.WithHelp("A", "clear selection")),
	Filter:      key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
	Sort:        key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "cycle sort")),
	StaleOnly:   key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "stale only")),
	Delete:      key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	ForceDelete: key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "force delete")),
	Rename:      key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rename")),
	Checkout:    key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "checkout")),
}
