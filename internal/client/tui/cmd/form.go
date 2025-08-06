package cmd

import (
	"gophkeeper/internal/client/tuiadapter"

	tea "github.com/charmbracelet/bubbletea"
)

type ShowFormTextMsg struct {
	ID   string
	Text string
	Meta tuiadapter.Meta
}

func ShowFormText(id, text string, meta tuiadapter.Meta) tea.Cmd {
	return NewCmd(ShowFormTextMsg{ID: id, Text: text, Meta: meta})
}

type EventSaveMsg struct{}
type EventDeleteMsg struct{}

func EventSave() tea.Cmd {
	return NewCmd(EventSaveMsg{})
}

func EventDelete() tea.Cmd {
	return NewCmd(EventDeleteMsg{})
}

type SaveTextMsg struct {
	ID    string
	NewID string
	Text  string
	Meta  tuiadapter.Meta
}

func SaveText(id, newID, text string, meta tuiadapter.Meta) tea.Cmd {
	return NewCmd(SaveTextMsg{
		ID:    id,
		NewID: newID,
		Text:  text,
		Meta:  meta,
	})
}

type DeleteIDMsg string

func DeleteID(id string) tea.Cmd {
	return NewCmd(DeleteIDMsg(id))
}

type ShowFormIDMsg string

func ShowFormID(id string) tea.Cmd {
	return NewCmd(ShowFormIDMsg(id))
}
