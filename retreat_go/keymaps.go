package main

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Quit   key.Binding
	Select key.Binding
	Back   key.Binding
}

var DefaultKeyMap = KeyMap{

	Quit: key.NewBinding(
		key.WithKeys("q"),
	),

	Up: key.NewBinding(
		key.WithKeys("k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
	),

	// Need to support chording - I want my gd T_T
	Select: key.NewBinding(
		key.WithKeys("enter"),
	),

	Back: key.NewBinding(
		key.WithKeys("b"),
	),
}
