package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
)

type ShowFormTextMsg struct {
	ID       string
	Text     string
	MetaKeys []string
	MetaVals []string
}

func NewShowFormTextCmd(id, text string, metaKeys, metaVals []string) tea.Cmd {
	return NewCmd(ShowFormTextMsg{ID: id, Text: text, MetaKeys: metaKeys, MetaVals: metaVals})
}

type FormValidMsg struct {
	IsValid bool
}

func NewFormValidCmd(valid bool) tea.Cmd {
	return NewCmd(FormValidMsg{IsValid: valid})
}
