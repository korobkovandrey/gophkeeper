package tui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

func newCmd[T tea.Msg](v T) tea.Cmd {
	return func() tea.Msg {
		return v
	}
}

type btnsFocusMsg struct{}
type updateBtnsMsg struct{}
type selectPrivatePathMsg string
type showFilepickerMsg struct{}
type changeSecretsMsg struct{}
type tableRowsMsg struct {
	rows   []table.Row
	cursor int
}
type showTableMsg struct{}
type showFormTextMsg struct {
	text string
}

func newBtnsFocusCmd() tea.Cmd {
	return newCmd(btnsFocusMsg{})
}

func newUpdateBtnsCmd() tea.Cmd {
	return newCmd(updateBtnsMsg{})
}

func newSelectPrivatePathCmd(path string) tea.Cmd {
	return newCmd(selectPrivatePathMsg(path))
}

func newShowFilepickerCmd() tea.Cmd {
	return newCmd(showFilepickerMsg{})
}

func newShowTableCmd() tea.Cmd {
	return newCmd(showTableMsg{})
}

func newTableRowsMsg(rows []table.Row, cursor int) tea.Msg {
	return tableRowsMsg{rows: rows, cursor: cursor}
}

func NewChangeSecretsMsg() tea.Msg {
	return changeSecretsMsg{}
}
func newChangeSecretsCmd() tea.Cmd {
	return newCmd(NewChangeSecretsMsg())
}

func newShowFormTextCmd(text string) tea.Cmd {
	return newCmd(showFormTextMsg{text: text})
}
