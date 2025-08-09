package cmd

import (
	"gophkeeper/internal/client/model"

	tea "github.com/charmbracelet/bubbletea"
)

type EventSaveMsg struct{}
type EventDeleteMsg struct{}

func EventSave() tea.Cmd {
	return NewCmd(EventSaveMsg{})
}

func EventDelete() tea.Cmd {
	return NewCmd(EventDeleteMsg{})
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

type ShowFormTextMsg struct {
	ID    model.ID
	NewID model.ID
	Meta  model.Meta
	Text  string
}

func ShowFormText(id, newID model.ID, meta model.Meta, text string) tea.Cmd {
	return NewCmd(ShowFormTextMsg{ID: id, NewID: newID, Meta: meta, Text: text})
}

type ShowFormLoginPassMsg struct {
	ID    model.ID
	NewID model.ID
	Meta  model.Meta
	Login string
	Pass  string
}

func ShowFormLoginPass(id, newID model.ID, meta model.Meta, login, pass string) tea.Cmd {
	return NewCmd(ShowFormLoginPassMsg{ID: id, NewID: newID, Meta: meta, Login: login, Pass: pass})
}

type ShowFormCardMsg struct {
	ID     model.ID
	NewID  model.ID
	Meta   model.Meta
	CCN    string
	Expire string
	CVV    string
}

func ShowFormCard(id, newID model.ID, meta model.Meta, ccn, expire, cvv string) tea.Cmd {
	return NewCmd(ShowFormCardMsg{ID: id, NewID: newID, Meta: meta, CCN: ccn, Expire: expire, CVV: cvv})
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
		Meta:  meta,
		Text:  text,
	})
}

type SaveLoginPassMsg struct {
	ID    model.ID
	NewID model.ID
	Meta  model.Meta
	Login string
	Pass  string
}

func SaveLoginPass(id, newID model.ID, meta model.Meta, login, pass string) tea.Cmd {
	return NewCmd(SaveLoginPassMsg{
		ID:    id,
		NewID: newID,
		Meta:  meta,
		Login: login,
		Pass:  pass,
	})
}

type SaveCardMsg struct {
	ID     model.ID
	NewID  model.ID
	Meta   model.Meta
	CCN    string
	Expire string
	CVV    string
}

func SaveCard(id, newID model.ID, meta model.Meta, ccn, expire, cvv string) tea.Cmd {
	return NewCmd(SaveCardMsg{
		ID:     id,
		NewID:  newID,
		Meta:   meta,
		CCN:    ccn,
		Expire: expire,
		CVV:    cvv,
	})
}
