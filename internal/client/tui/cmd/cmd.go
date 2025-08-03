package cmd

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

func NewCmd[T tea.Msg](v T) tea.Cmd {
	return func() tea.Msg {
		return v
	}
}

type UpdateBtnsMsg struct{}
type SelectPrivatePathMsg string
type ShowFilepickerMsg struct{}
type ChangeSecretsMsg struct{}
type TableRowsMsg struct {
	Rows   []table.Row
	Cursor int
}
type ShowTableMsg struct{}
type ScreenFocusMsg struct{}
type ScreenBlurMsg struct{}
type SelectIDMsg string

func NewTableRowsMsg(rows []table.Row, cursor int) tea.Msg {
	return TableRowsMsg{Rows: rows, Cursor: cursor}
}

func NewChangeSecretsMsg() tea.Msg {
	return ChangeSecretsMsg{}
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

func TableRows(rows []table.Row, cursor int) tea.Cmd {
	return NewCmd(NewTableRowsMsg(rows, cursor))
}

func ChangeSecrets() tea.Cmd {
	return NewCmd(NewChangeSecretsMsg())
}

func SelectID(id string) tea.Cmd {
	return NewCmd(SelectIDMsg(id))
}
