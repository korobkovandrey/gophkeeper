package tui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tableModel struct {
	table.Model
	debug string
}

func newTableModel() tableModel {
	columns := []table.Column{
		{Title: "Статус", Width: 10},
		{Title: "ID", Width: 20},
		{Title: "Desc", Width: 20},
		{Title: "Created", Width: 10},
		{Title: "Updated", Width: 10},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	return tableModel{Model: t}
}

func (m tableModel) Init() tea.Cmd {
	return nil
}

func (m tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch tMsg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Model.SetHeight(tMsg.Height - 6)
	case tea.KeyMsg:
		if tMsg.Type == tea.KeyEsc || tMsg.Type == tea.KeyTab {
			return m, newBtnsFocusCmd()
		}
	case tableRowsMsg:
		m.Model.SetRows(tMsg.rows)
		m.Model.SetCursor(tMsg.cursor)
		return m, nil
	}
	var c tea.Cmd
	m.Model, c = m.Model.Update(msg)
	return m, c
}

func (m tableModel) View() string {
	return tableBaseStyle.Render(m.Model.View()) + "\n" + m.debug
}
