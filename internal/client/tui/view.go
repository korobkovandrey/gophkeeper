package tui

import (
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
		selectedStyle.Render(m.app.GetPrivateKeyPath()), " ",
		m.app.Fingerprint(),
	)

	var body, help string
	if m.msg != "" {
		body = lipgloss.JoinVertical(lipgloss.Center,
			msgStyle.Render(wordwrap.String(strings.Replace(m.msg, ": ", ":\n", 1), 100)),
			selectedStyle.Render("press any key to return"))
	} else if m.screen != nil {
		body = m.screen.View()
	}
	help += "esc: change focus • tab: focus next • enter: select • ctrl+c, f10: exit"
	m.debug = m.selectID
	if m.debug != "" {
		help += " debug: " + m.debug
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, body, btnsStr, helpStyle.Render(help))
}
