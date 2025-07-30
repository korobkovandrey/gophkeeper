package tui

func setMsg(m *Model, msg string) {
	//m.btnsFocus = false
	m.msg = msg
}

func updateBtns(m *Model) {
	var currentBtn btn
	if len(m.btns) > m.btnsCursor {
		currentBtn = m.btns[m.btnsCursor]
	}
	m.btnsCursor = 0
	m.btns = make([]btn, 0)

	var isScreenTable, isScreenFilepicker bool
	switch m.screen.(type) {
	case tableModel:
		isScreenTable = true
	case filepickerModel:
		isScreenFilepicker = true
	}
	if !isScreenTable && m.app.GetPrivateKeyPath() != "" {
		m.btns = append(m.btns, btnTable)
	}
	if !m.app.IsLogged() {
		if m.app.GetPrivateKeyPath() != "" {
			m.btns = append(m.btns, btnLogin, btnRegister)
		}
		if !isScreenFilepicker {
			m.btns = append(m.btns, btnFile)
		}
	}
	if m.btnsFocus {
		for i := range m.btns {
			if currentBtn == m.btns[i] {
				m.btnsCursor = i
				break
			}
		}
	}
}
