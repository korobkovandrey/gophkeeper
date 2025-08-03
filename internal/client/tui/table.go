package tui

import (
	"gophkeeper/internal/client/model2"
	"gophkeeper/internal/client/tui/cmd"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tableModel struct {
	table.Model
	secrets  []*model2.Secret
	selectID string
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
		m.Model.SetHeight(tMsg.Height - 2)
	case cmd.ScreenFocusMsg:
		m.Model.Focus()
	case cmd.ScreenBlurMsg:
		m.Model.Blur()
	case tea.KeyMsg:
		if tMsg.Type == tea.KeyTab {
			return m, cmd.ScreenBlur()
		}
	case cmd.SelectIDMsg:
		m.selectID = string(tMsg)
		return m, nil
	case cmd.TableRowsMsg:
		m.Model.SetRows(tMsg.Rows)
		m.Model.SetCursor(tMsg.Cursor)
		return m, m.checkSelect()
	}
	m.Model, _ = m.Model.Update(msg)
	return m, m.checkSelect()
}

func (m tableModel) View() string {
	return tableBaseStyle.Render(m.Model.View())
}

func (m tableModel) checkSelect() tea.Cmd {
	selectID := ""
	if selected := m.Model.SelectedRow(); selected != nil {
		selectID = selected[1]
	}
	if selectID == m.selectID {
		return nil
	}
	return cmd.SelectID(selectID)
}
