package main

// Can I add a border to the TUI????

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedListStyle = lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("63"))

	unfocusedListStyle = lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.NormalBorder())

	cmdPalletFocusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	cmdPalletUnFocusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)
