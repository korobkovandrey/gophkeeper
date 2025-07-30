package tui

import "github.com/charmbracelet/lipgloss"

var (
	headerStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			Margin(0, 1).Padding(1, 5).
			Foreground(lipgloss.Color("#04326f")).Bold(true)
	headerDescStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	btnStyle        = lipgloss.NewStyle().
			Padding(0, 1).
		//Width(15).
		Height(1).
		Align(lipgloss.Center, lipgloss.Center).
		BorderStyle(lipgloss.HiddenBorder())
	btnFocusedStyle = lipgloss.NewStyle().
			Padding(0, 1).
		//Width(15).
		Height(1).
		Align(lipgloss.Center, lipgloss.Center).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("69"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	msgStyle      = lipgloss.NewStyle().Margin(0, 0, 1, 2).Bold(true).
			Blink(true)
	tableBaseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))
)
