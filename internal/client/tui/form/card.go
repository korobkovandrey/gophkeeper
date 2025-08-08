package form

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type cardModel struct {
	ccn    textinput.Model
	expire textinput.Model
	cvv    textinput.Model
}

func newCardModel(ccn, expire, cvv string) cardModel {
	m := cardModel{
		ccn:    textinput.New(),
		expire: textinput.New(),
		cvv:    textinput.New(),
	}
	m.ccn.Placeholder = "4505 **** **** 1234"
	m.ccn.Prompt = prompt
	m.ccn.Width = 30
	m.ccn.CharLimit = 20
	m.ccn.SetValue(ccn)
	m.ccn.Validate = ccnValidator
	m.expire.Placeholder = "MM/YY "
	m.expire.Prompt = prompt
	m.expire.CharLimit = 5
	m.expire.Width = 5
	m.expire.SetValue(expire)
	m.expire.Validate = expValidator
	m.cvv.Placeholder = "XXX"
	m.cvv.Prompt = prompt
	m.cvv.CharLimit = 3
	m.cvv.Width = 5
	m.cvv.SetValue(cvv)
	m.cvv.Validate = cvvValidator
	return m
}

func (m cardModel) update(msg tea.Msg) (fieldsModel, tea.Cmd) {
	var c [3]tea.Cmd
	m.ccn, c[0] = m.ccn.Update(msg)
	m.expire, c[1] = m.expire.Update(msg)
	m.cvv, c[2] = m.cvv.Update(msg)
	return m, tea.Batch(c[0], c[1], c[2])
}

func (m cardModel) view() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		renderInput("CCN", m.ccn),
		renderInput("EXP", m.expire),
		renderInput("CVV", m.cvv),
	)
}

func (m cardModel) lenInputs() int {
	return 3
}

func (m cardModel) focus(f int) (fieldsModel, tea.Cmd) {
	var c tea.Cmd
	switch f {
	case 0:
		c = m.ccn.Focus()
	case 1:
		c = m.expire.Focus()
	case 2:
		c = m.cvv.Focus()
	}
	return m, c
}

func (m cardModel) blur(f int) fieldsModel {
	switch f {
	case 0:
		m.ccn.Blur()
	case 1:
		m.expire.Blur()
	case 2:
		m.cvv.Blur()
	}
	return m
}

func (m cardModel) isValid() bool {
	return ccnValidator(m.ccn.Value()) == nil &&
		expValidator(m.expire.Value()) == nil &&
		cvvValidator(m.cvv.Value()) == nil
}
