package main

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up   key.Binding
	Down key.Binding
	Quit key.Binding
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
}
