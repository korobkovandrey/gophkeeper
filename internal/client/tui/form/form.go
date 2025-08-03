package form

import (
	"fmt"
	"gophkeeper/internal/client/tui/cmd"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	labelStyle            = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	inputStyle            = lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.RoundedBorder())
	focusedStyle          = inputStyle.BorderForeground(lipgloss.Color("#04B575"))
	unfocusedStyle        = inputStyle.BorderForeground(lipgloss.Color("#333"))
	focusedInvalidStyle   = inputStyle.BorderForeground(lipgloss.Color("#b50404"))
	unfocusedInvalidStyle = inputStyle.BorderForeground(lipgloss.Color("#450202"))
)

type MetaData struct {
	Key string
	Val string
}

type fieldsModel interface {
	update(tea.Msg) (fieldsModel, tea.Cmd)
	view() string
	lenInputs() int
	focus(int) (fieldsModel, tea.Cmd)
	blur(int) fieldsModel
	isValid() bool
}

type Model struct {
	id       string
	newID    textinput.Model
	metaKeys []textinput.Model
	metaVals []textinput.Model
	viewport viewport.Model
	model    fieldsModel
	cursor   int
	focused  bool
	isValid  bool
}

func NewModel(id string, fields fieldsModel, metaKeys, metaVals []string) Model {
	vp := viewport.New(0, 0)
	vp.MouseWheelEnabled = true
	m := Model{
		id:       id,
		newID:    textinput.New(),
		model:    fields,
		viewport: vp,
		cursor:   -1,
	}
	m.newID.Placeholder = "ID"
	m.newID.Prompt = "> "
	m.newID.Width = 20
	m.newID.SetValue(id)
	addMetas(&m, metaKeys, metaVals)
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func addMetas(m *Model, metaKeys, metaVals []string) {
	lenVals := len(metaVals)
	for i := range metaKeys {
		key := textinput.New()
		key.Width = 10
		key.Placeholder = "Key"
		key.Prompt = "> "
		key.SetValue(metaKeys[i])
		m.metaKeys = append(m.metaKeys, key)
		val := textinput.New()
		val.Width = 10
		val.Placeholder = "Value"
		val.Prompt = "> "
		if i < lenVals {
			val.SetValue(metaVals[i])
		}
		m.metaVals = append(m.metaVals, val)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var c tea.Cmd
	m.viewport, c = m.viewport.Update(msg)
	if c != nil {
		return m, c
	}
	newCursor := m.cursor
	lenModelInputs := m.model.lenInputs()
	lenMetaInputs := len(m.metaKeys) * 2
	lenInputs := 1 + lenModelInputs + lenMetaInputs
	focused := m.focused
	isValid := m.isValid
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Height = msg.Height
		return m, nil
	case cmd.ScreenFocusMsg:
		m.focused = true
	case cmd.ScreenBlurMsg:
		newCursor = -1
		m.focused = false
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			if m.cursor < lenInputs-1 {
				newCursor = m.cursor + 1
			} else {
				return m, cmd.ScreenBlur()
			}
		case tea.KeyCtrlD:
			addMetas(&m, []string{""}, []string{""})
			lenMetaInputs += 2
			lenInputs += 2
			if m.focused {
				newCursor = lenInputs - 2
			}
		}
	}
	if newCursor < 0 && m.focused {
		newCursor = 0
	}
	if newCursor == m.cursor && focused == m.focused {
		if m.cursor == 0 {
			m.newID, c = m.newID.Update(msg)
		} else if m.cursor > lenModelInputs {
			f := m.cursor - lenModelInputs - 1
			metaIndex := f / 2
			if f%2 == 0 {
				m.metaKeys[metaIndex], c = m.metaKeys[metaIndex].Update(msg)
			} else {
				m.metaVals[metaIndex], c = m.metaVals[metaIndex].Update(msg)
			}
		} else {
			m.model, c = m.model.update(msg)
		}
		m.isValid = m.validate()
		if isValid != m.validate() {
			c = tea.Sequence(c, cmd.NewFormValidCmd(m.isValid))
		}
	} else {
		if newCursor == 0 {
			c = m.newID.Focus()
		} else if newCursor > lenModelInputs {
			f := newCursor - lenModelInputs - 1
			metaIndex := f / 2
			if f%2 == 0 {
				c = m.metaKeys[metaIndex].Focus()
			} else {
				c = m.metaVals[metaIndex].Focus()
			}
		} else if newCursor > 0 {
			m.model, c = m.model.focus(newCursor - 1)
		}

		if m.cursor == 0 {
			m.newID.Blur()
		} else if m.cursor > lenModelInputs {
			f := m.cursor - lenModelInputs - 1
			metaIndex := f / 2
			if f%2 == 0 {
				m.metaKeys[metaIndex].Blur()
			} else {
				m.metaVals[metaIndex].Blur()
			}
		} else {
			m.model = m.model.blur(m.cursor - 1)
		}
		m.cursor = newCursor
	}
	m.viewport.SetContent(m.view())
	return m, c
}

func (m Model) View() string {
	return m.viewport.View()
}

func (m Model) view() string {
	var rows []string
	if m.id == "" {
		rows = append(rows, labelStyle.Render("Новый текст"))
	} else {
		rows = append(rows, labelStyle.Render("Обновление: ")+m.id)
	}
	newIDInput := m.newID.View()
	if m.newID.Focused() {
		if m.newID.Value() == "" {
			newIDInput = focusedInvalidStyle.Render(newIDInput)
		} else {
			newIDInput = focusedStyle.Render(newIDInput)
		}
	} else {
		if m.newID.Value() == "" {
			newIDInput = unfocusedInvalidStyle.Render(newIDInput)
		} else {
			newIDInput = unfocusedStyle.Render(newIDInput)
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

func (m Model) validate() bool {
	return m.newID.Value() != "" && m.model.isValid()
}
