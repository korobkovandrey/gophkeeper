package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

func (m Model) View() string {
	if !m.initialized {
		return ""
	}
	var btns []string
	for i := 0; i < len(m.btns); i++ {
		name, ok := m.btnNames[m.btns[i]]
		if !ok {
			name = strconv.Itoa(int(m.btns[i]))
		}
		if m.focused && m.cursor == i {
			btns = append(btns, btnFocusedStyle.Render(name))
		} else {
			btns = append(btns, btnStyle.Render(name))
		}
	}
	btnsStr := lipgloss.JoinHorizontal(lipgloss.Center, btns...)

	var statusText string
	if m.app.IsOnline() {
		statusText = onlineStyle.Render("ONLINE")
	} else {
		statusText = offlineStyle.Render("OFFLINE")
	}

	header := lipgloss.JoinHorizontal(lipgloss.Center,
		headerStyle.Render("GophKeeper"), " ",
		statusText, " ",
		selectedStyle.Render(m.key.GetPrivateKeyPath()), " ",
		m.key.Fingerprint(),
	)

	var body, help string
	if _, ok := m.screen.(filepickerModel); ok {
		help = "esc: parent directory"
	} else {
		help = "esc: change focus"
	}
	help += " • tab: focus next • enter: select • ctrl+c, f10: exit"
	if m.debug != nil {
		help += " | debug: " + fmt.Sprint(m.debug)
	}

	if m.msg != "" {
		body = lipgloss.JoinVertical(lipgloss.Center,
			header,
			msgStyle.Render(wordwrap.String(strings.Replace(m.msg, ": ", ":\n", 1), max(100, lipgloss.Width(header)))),
			selectedStyle.Render("press any key to return"))
		return lipgloss.JoinVertical(lipgloss.Left, body, btnsStr, helpStyle.Render(help))
	} else if m.screen != nil {
		body = m.screen.View()
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, body, btnsStr, helpStyle.Render(help))
}
