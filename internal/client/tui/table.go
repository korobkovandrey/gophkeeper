package tui

import (
	"gophkeeper/internal/client/tui/cmd"
	"gophkeeper/internal/client/tuiadapter"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tableModel struct {
	table.Model
	ids []string
}

func newTableModel() tableModel {
	columns := []table.Column{
		{Title: "Статус", Width: 10},
		{Title: "Тип", Width: 10},
		{Title: "ID", Width: 30},
		{Title: "Meta", Width: 40},
		{Title: "Created", Width: 20},
		{Title: "Updated", Width: 20},
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
		m.Model.SetHeight(tMsg.Height - 2)
	case cmd.ScreenFocusMsg:
		m.Model.Focus()
	case cmd.ScreenBlurMsg:
		m.Model.Blur()
	case tea.KeyMsg:
		if tMsg.Type == tea.KeyTab {
			return m, cmd.ScreenBlur()
		}
		if tMsg.Type == tea.KeyEnter {
			cursor := m.Model.Cursor()
			if cursor >= 0 && cursor < len(m.ids) {
				return m, cmd.ShowFormID(m.ids[cursor])
			}
		}
	case cmd.RowsMsg:
		selectID := tMsg.ID
		cursor := m.Model.Cursor()
		if selectID == "" {
			if cursor >= 0 && cursor < len(m.ids) {
				selectID = m.ids[cursor]
			}
		}
		var rows []table.Row
		var newCursor int
		rows, m.ids, newCursor = rowsToTableRows(selectID, tMsg.Rows)
		m.Model.SetRows(rows)
		if cursor != newCursor {
			m.Model.SetCursor(newCursor)
		}
		return m, nil
	}
	m.Model, _ = m.Model.Update(msg)
	return m, nil
}

func (m tableModel) View() string {
	return tableBaseStyle.Render(m.Model.View())
}

func rowsToTableRows(id string, rows []tuiadapter.Row) ([]table.Row, []string, int) {
	r := make([]table.Row, len(rows))
	ids := make([]string, len(rows))
	cursor := 0
	for i := range rows {
		r[i] = table.Row{
			rows[i].Status,
			rows[i].Type,
			rows[i].ID,
			rows[i].Meta,
			rows[i].Created,
			rows[i].Updated,
		}
		if rows[i].ID == id {
			cursor = i
		}
		ids[i] = rows[i].ID
	}
	return r, ids, cursor
}
