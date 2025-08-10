package tui

import (
	"gophkeeper/internal/client/model"
	"gophkeeper/internal/client/tui/cmd"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tableModel struct {
	table.Model
	ids []model.ID
}

func newTableModel() tableModel {
	columns := []table.Column{
		{Title: "Статус", Width: 10},
		{Title: "Тип", Width: 15},
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
		return m, nil
	case cmd.ScreenFocusMsg:
		m.Model.Focus()
		return m, nil
	case cmd.ScreenBlurMsg:
		m.Model.Blur()
		return m, nil
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
	case cmd.UpdateTableMsg:
		selectID := tMsg.ID
		cursor := m.Model.Cursor()
		var rows []table.Row
		var newCursor int
		rows, m.ids, newCursor = secretsToTableRows(selectID, tMsg.Secrets)
		if selectID == "" {
			if cursor >= 0 {
				newCursor = cursor
				if newCursor >= len(m.ids) {
					newCursor = len(m.ids) - 1
				}
			}
		}
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

func secretsToTableRows(id model.ID, secrets []*model.Secret) ([]table.Row, []model.ID, int) {
	r := make([]table.Row, len(secrets))
	ids := make([]model.ID, len(secrets))
	cursor := 0
	for i := range secrets {
		r[i] = table.Row{
			"",
			"",
			string(secrets[i].NewID),
			secrets[i].DecryptMeta.ToString(),
			secrets[i].CreatedAt.Format(time.DateTime),
			secrets[i].UpdatedAt.Format(time.DateTime),
		}
		switch secrets[i].Status {
		case model.StatusSynced:
			r[i][0] = "Сохранен"
		case model.StatusDeleting:
			r[i][0] = "Удаление"
		case model.StatusNew:
			r[i][0] = "Новый"
		case model.StatusUpdating:
			r[i][0] = "Изменен"
		default:
			r[i][0] = "Неизвестно"
		}
		switch secrets[i].Type {
		case model.TypeText:
			r[i][1] = "Текст"
		case model.TypeLoginPass:
			r[i][1] = "Логин/Пароль"
		case model.TypeCard:
			r[i][1] = "Карта"
		default:
			r[i][1] = "Неизвестно"
		}
		if secrets[i].ID == id {
			cursor = i
		}
		ids[i] = secrets[i].ID
	}
	return r, ids, cursor
}
