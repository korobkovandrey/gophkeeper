package form

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type loginPassModel struct {
	login textinput.Model
	pass  textinput.Model
}

func newLoginPassModel(login, pass string) loginPassModel {
	m := loginPassModel{
		login: textinput.New(),
		pass:  textinput.New(),
	}
	m.login.Placeholder = "Login"
	m.login.Prompt = prompt
	m.login.Width = 20
	m.login.SetValue(login)
	m.login.Validate = requiredValidator
	m.pass.Placeholder = "Password"
	m.pass.Prompt = prompt
	m.pass.Width = 20
	m.pass.SetValue(pass)
	m.pass.Validate = requiredValidator
	return m
}

func (m loginPassModel) update(msg tea.Msg) (fieldsModel, tea.Cmd) {
	var c [2]tea.Cmd
	m.login, c[0] = m.login.Update(msg)
	m.pass, c[1] = m.pass.Update(msg)
	return m, tea.Batch(c[0], c[1])
}

func (m loginPassModel) view() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		renderInput("Login", m.login),
		renderInput("Password", m.pass),
	)
}

func (m loginPassModel) lenInputs() int {
	return 2
}

func (m loginPassModel) focus(f int) (fieldsModel, tea.Cmd) {
	var c tea.Cmd
	switch f {
	case 0:
		c = m.login.Focus()
	case 1:
		c = m.pass.Focus()
	}
	return m, c
}

func (m loginPassModel) blur(f int) fieldsModel {
	switch f {
	case 0:
		m.login.Blur()
	case 1:
		m.pass.Blur()
	}
	return m
}

func (m loginPassModel) isValid() bool {
	return requiredValidator(m.login.Value()) == nil && requiredValidator(m.pass.Value()) == nil
}
