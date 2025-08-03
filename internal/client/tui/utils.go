package tui

import "gophkeeper/internal/client/tui/form"

func setMsg(m *Model, msg string) {
	//m.focused = false
	m.msg = msg
}

func updateBtns(m *Model) {
	var currentBtn btn
	if len(m.btns) > m.cursor {
		currentBtn = m.btns[m.cursor]
	}
	m.cursor = 0
	m.btns = make([]btn, 0)

	var isScreenTable, isScreenFilepicker, isForm bool
	switch m.screen.(type) {
	case tableModel:
		isScreenTable = true
	case filepickerModel:
		isScreenFilepicker = true
	case form.Model:
		isForm = true
	}

	isWorkAvailable := m.app.GetPrivateKeyPath() != ""

	if isWorkAvailable {
		if isForm {
			if m.formIsValid {
				m.btns = append(m.btns, btnSave)
			}
			if !m.formIsNew {
				m.btns = append(m.btns, btnDelete)
			}
		}
		if isScreenTable {
			m.btns = append(m.btns, btnFormAddText)
		} else {
			m.btns = append(m.btns, btnTable)
		}
	}
	if !m.app.IsLogged() {
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
