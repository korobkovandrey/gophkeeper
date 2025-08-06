package cmd

import (
	"gophkeeper/internal/client/tuiadapter"

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
type RowsMsg struct {
	ID   string
	Rows []tuiadapter.Row
}
type ShowTableMsg struct{}
type ScreenFocusMsg struct{}
type ScreenBlurMsg struct{}
type SelectIDMsg string

func NewRowsMsg(id string, rows []tuiadapter.Row) tea.Msg {
	return RowsMsg{ID: id, Rows: rows}
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

func TableRows(id string, rows []tuiadapter.Row) tea.Cmd {
	return NewCmd(NewRowsMsg(id, rows))
}

func ChangeSecrets(id string) tea.Cmd {
	return NewCmd(NewChangeSecretsMsg(id))
}

func SelectID(id string) tea.Cmd {
	return NewCmd(SelectIDMsg(id))
}
