package cmd

import (
	"gophkeeper/internal/client/model"

	tea "github.com/charmbracelet/bubbletea"
)

func NewCmd[T tea.Msg](v T) tea.Cmd {
	return func() tea.Msg {
		return v
	}
}

type MsgMsg string

func Msg(msg string) tea.Cmd {
	return NewCmd(MsgMsg(msg))
}

type UpdateBtnsMsg struct{}
type SelectPrivatePathMsg string
type ShowFilepickerMsg struct{}
type ChangeSecretsMsg struct {
	ID string
}
type UpdateTableMsg struct {
	ID      model.ID
	Secrets []*model.Secret
}
type ShowTableMsg struct{}
type ScreenFocusMsg struct{}
type ScreenBlurMsg struct{}

func NewUpdateTableMsg(id model.ID, secrets []*model.Secret) tea.Msg {
	return UpdateTableMsg{ID: id, Secrets: secrets}
}

func NewChangeSecretsMsg(id string) tea.Msg {
	return ChangeSecretsMsg{ID: id}
}

func ScreenFocus() tea.Cmd {
	return NewCmd(ScreenFocusMsg{})
}

func ScreenBlur() tea.Cmd {
	return NewCmd(ScreenBlurMsg{})
}

func UpdateBtns() tea.Cmd {
	return NewCmd(UpdateBtnsMsg{})
}

func SelectPrivatePath(path string) tea.Cmd {
	return NewCmd(SelectPrivatePathMsg(path))
}

func ShowFilepicker() tea.Cmd {
	return NewCmd(ShowFilepickerMsg{})
}

func ShowTable() tea.Cmd {
	return NewCmd(ShowTableMsg{})
}

func UpdateTable(id model.ID, secrets []*model.Secret) tea.Cmd {
	return NewCmd(NewUpdateTableMsg(id, secrets))
}

func ChangeSecrets(id string) tea.Cmd {
	return NewCmd(NewChangeSecretsMsg(id))
}
