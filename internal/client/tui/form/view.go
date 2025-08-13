package form

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	return m.viewport.View()
}

func (m Model) view() string {
	var rows []string
	if m.id == "" {
		rows = append(rows, labelStyle.Render("Новый текст"))
	} else {
		rows = append(rows, labelStyle.Render("Обновление: ")+string(m.id))
	}
	newIDInput := m.newID.View()
	if m.newID.Focused() {
		if m.newID.Err == nil {
			newIDInput = focusedStyle.Render(newIDInput)
		} else {
			newIDInput = focusedInvalidStyle.Render(newIDInput)
		}
	} else {
		if m.newID.Err == nil {
			newIDInput = unfocusedStyle.Render(newIDInput)
		} else {
			newIDInput = unfocusedInvalidStyle.Render(newIDInput)
		}
	}
	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Center,
		labelStyle.Render("ID:"), newIDInput,
	), m.model.view())

	metaRows := make([]string, len(m.metaKeys)+1)
	for i := range m.metaKeys {
		k := m.metaKeys[i].View()
		v := m.metaVals[i].View()

		if m.metaKeys[i].Focused() {
			k = focusedStyle.Render(k)
		} else {
			k = unfocusedStyle.Render(k)
		}
		if m.metaVals[i].Focused() {
			v = focusedStyle.Render(v)
		} else {
			v = unfocusedStyle.Render(v)
		}
		metaRows[i] = lipgloss.JoinHorizontal(lipgloss.Center,
			labelStyle.Render(fmt.Sprintf("Meta [%d]:", i+1)),
			k,
			v,
		)
	}
	metaRows[len(metaRows)-1] = "• ctrl+d: add meta"
	return lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Left, rows...), " ",
		lipgloss.JoinVertical(lipgloss.Left, metaRows...))
}
