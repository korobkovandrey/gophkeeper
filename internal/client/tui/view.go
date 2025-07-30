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
		if m.btnsFocus && m.btnsCursor == i {
			btns = append(btns, btnFocusedStyle.Render(name))
		} else {
			btns = append(btns, btnStyle.Render(name))
		}
	}
	btnsStr := lipgloss.JoinHorizontal(lipgloss.Left, btns...)

	header := headerStyle.Render(lipgloss.JoinVertical(lipgloss.Center, "GophKeeper", selectedStyle.Render(m.app.GetPrivateKeyPath())))

	var body, help string
	if m.msg != "" {
		body = lipgloss.JoinVertical(lipgloss.Center,
			msgStyle.Render(wordwrap.String(strings.Replace(m.msg, ": ", ":\n", 1), 100)),
			selectedStyle.Render("press any key to return"))
	} else if m.screen != nil {
		body = m.screen.View()
	}
	help += "tab: focus next • enter: select • ctrl+c, f10: exit"
	if m.app.IsOnline() {
		help += " • ONLINE"
	} else {
		help += " • OFFLINE"
	}
	help += m.debug
	return lipgloss.JoinVertical(lipgloss.Top, header, lipgloss.JoinVertical(lipgloss.Top, body, btnsStr, helpStyle.Render(help)))
}
