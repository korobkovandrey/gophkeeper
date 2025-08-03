package tui

import "github.com/charmbracelet/lipgloss"

var (
	headerStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).Padding(1).
			Foreground(lipgloss.Color("#04326f")).Bold(true)
	btnStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Height(1).
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.HiddenBorder())
	btnFocusedStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Height(1).
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("69"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	onlineStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#09732d")).Bold(true)
	offlineStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#751e21")).Bold(true)
	msgStyle      = lipgloss.NewStyle().Margin(0, 0, 1, 2).Bold(true).
			Blink(true)
	tableBaseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))
)
