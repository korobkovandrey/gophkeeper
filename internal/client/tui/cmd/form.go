package cmd

import (
	"gophkeeper/internal/client/model"

	tea "github.com/charmbracelet/bubbletea"
)

type ShowFormTextMsg struct {
	ID   model.ID
	Meta model.Meta
	Text string
}

func ShowFormText(id model.ID, meta model.Meta, text string) tea.Cmd {
	return NewCmd(ShowFormTextMsg{ID: id, Meta: meta, Text: text})
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
	ID    model.ID
	NewID model.ID
	Meta  model.Meta
	Text  string
}

func SaveText(id, newID model.ID, meta model.Meta, text string) tea.Cmd {
	return NewCmd(SaveTextMsg{
		ID:    id,
		NewID: newID,
		Text:  text,
		Meta:  meta,
	})
}

type DeleteIDMsg struct {
	ID model.ID
}

func DeleteID(id model.ID) tea.Cmd {
	return NewCmd(DeleteIDMsg{ID: id})
}

type ShowFormIDMsg struct {
	ID model.ID
}

func ShowFormID(id model.ID) tea.Cmd {
	return NewCmd(ShowFormIDMsg{ID: id})
}
