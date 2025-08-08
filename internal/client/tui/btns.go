package tui

import "gophkeeper/internal/client/tui/form"

type btn int

const (
	btnTable btn = iota
	btnLogin
	btnRegister
	btnFile
	btnFormAddText
	btnFormAddLoginPass
	btnFormAddCard
	btnSave
	btnDelete
)

//nolint:gocyclo // ignore
func updateBtns(m *Model) {
	var currentBtn btn
	if len(m.btns) > m.cursor {
		currentBtn = m.btns[m.cursor]
	}
	m.cursor = 0
	m.btns = make([]btn, 0)

	var isScreenTable, isScreenFilepicker, isForm, formIsNew, formIsValid bool
	switch model := m.screen.(type) {
	case tableModel:
		isScreenTable = true
	case filepickerModel:
		isScreenFilepicker = true
	case form.Model:
		isForm = true
		formIsNew = model.IsNew()
		formIsValid = model.IsValid
	}

	isWorkAvailable := m.key.GetPrivateKeyPath() != ""

	if isWorkAvailable {
		if isForm {
			if formIsValid {
				m.btns = append(m.btns, btnSave)
			}
			if !formIsNew {
				m.btns = append(m.btns, btnDelete)
			}
		}
		if isScreenTable {
			m.btns = append(m.btns, btnFormAddText)
		} else {
			m.btns = append(m.btns, btnTable)
		}
	}
	if !m.key.IsLogged() {
		if isWorkAvailable {
			m.btns = append(m.btns, btnLogin, btnRegister)
		}
		if !isScreenFilepicker {
			m.btns = append(m.btns, btnFile)
		}
	}
	if m.focused {
		for i := range m.btns {
			if currentBtn == m.btns[i] {
				m.cursor = i
				break
			}
		}
	}
}
