package tui

import (
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type filepickerModel struct {
	filepicker.Model
}

func newFilepickerModel() filepickerModel {
	fp := filepicker.New()
	fp.CurrentDirectory, _ = os.UserHomeDir()
	fp.ShowHidden = true
	return filepickerModel{
		Model: fp,
	}
}

func (m filepickerModel) Init() tea.Cmd {
	return m.Model.Init()
}

func (m filepickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch tMsg := msg.(type) {
	case tea.WindowSizeMsg:
		tMsg.Height -= 1
		msg = tMsg
	case tea.KeyMsg:
		if tMsg.Type == tea.KeyEsc || tMsg.Type == tea.KeyTab {
			return m, newBtnsFocusCmd()
		}
	}
	var c tea.Cmd
	m.Model, c = m.Model.Update(msg)
	if didSelect, path := m.Model.DidSelectFile(msg); didSelect {
		return m, tea.Sequence(c, newSelectPrivatePathCmd(path))
	}
	return m, c
}

func (m filepickerModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Top,
		helpStyle.Render("Выберите приватный ключ:"),
		m.Model.View(),
	)
}
