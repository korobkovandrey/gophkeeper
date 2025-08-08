package tui

import (
	"gophkeeper/internal/client/tui/cmd"
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type filepickerModel struct {
	filepicker.Model
	focused bool
}

func newFilepickerModel() filepickerModel {
	fp := filepicker.New()
	fp.CurrentDirectory, _ = os.UserHomeDir()
	fp.ShowHidden = true
	fp.AutoHeight = false
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
		m.Model.SetHeight(tMsg.Height - 2)
	case cmd.ScreenFocusMsg:
		m.focused = true
	case cmd.ScreenBlurMsg:
		m.focused = false
	}
	if !m.focused {
		return m, nil
	}
	if tMsg, ok := msg.(tea.KeyMsg); ok && tMsg.Type == tea.KeyTab {
		return m, cmd.ScreenBlur()
	}
	var c tea.Cmd
	m.Model, c = m.Model.Update(msg)
	if didSelect, path := m.Model.DidSelectFile(msg); didSelect {
		return m, tea.Sequence(c, cmd.SelectPrivatePath(path))
	}
	return m, c
}

func (m filepickerModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Top,
		helpStyle.Render("Выберите приватный ключ:"),
		m.Model.View(),
	)
}
