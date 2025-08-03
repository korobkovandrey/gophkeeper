package form

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type textModel struct {
	text textarea.Model
}

func newTextModel(text string) textModel {
	m := textModel{
		text: textarea.New(),
	}
	m.text.Placeholder = "Text"
	m.text.SetWidth(60)
	m.text.ShowLineNumbers = true
	m.text.Blur()
	m.text.SetValue(text)
	return m
}

func (m textModel) update(msg tea.Msg) (fieldsModel, tea.Cmd) {
	var c tea.Cmd
	m.text, c = m.text.Update(msg)
	return m, c
}

func (m textModel) view() string {
	textInput := m.text.View()
	if m.text.Focused() {
		if m.text.Value() == "" {
			textInput = focusedInvalidStyle.Render(textInput)
		} else {
			textInput = focusedStyle.Render(textInput)
		}
	} else {
		if m.text.Value() == "" {
			textInput = unfocusedInvalidStyle.Render(textInput)
		} else {
			textInput = unfocusedStyle.Render(textInput)
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render("Text:"), textInput,
	)
}

func (m textModel) lenInputs() int {
	return 1
}

func (m textModel) focus(f int) (fieldsModel, tea.Cmd) {
	var c tea.Cmd
	if f == 0 {
		c = m.text.Focus()
	}
	return m, c
}

func (m textModel) blur(f int) fieldsModel {
	if f == 0 {
		m.text.Blur()
	}
	return m
}

func (m textModel) isValid() bool {
	return m.text.Value() != ""
}
