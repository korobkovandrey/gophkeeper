package tui

import (
	"context"
	"gophkeeper/internal/client/tui/cmd"
	"gophkeeper/internal/client/tui/form"
	"gophkeeper/internal/client/tuiadapter"

	tea "github.com/charmbracelet/bubbletea"
)

type keyService interface {
	GetPrivateKeyPath() string
	Fingerprint() string
	SetPrivateKeyPath(privateKeyPath string) error
	IsLogged() bool
}

type authClient interface {
	Login(ctx context.Context) error
	Register(ctx context.Context) error
	//IsOnline() bool
}

type storageService interface {
	List() []tuiadapter.Row
	SaveText(id, newID, text string) (string, error)
	Delete(id string) error
}

type Model struct {
	ctx     context.Context
	key     keyService
	auth    authClient
	storage *tuiadapter.Storage

	msg    string
	table  tea.Model
	screen tea.Model

	focused     bool
	cursor      int
	btns        []btn
	btnNames    map[btn]string
	debug       string
	initialized bool
}

func NewModel(ctx context.Context, key keyService, auth authClient, storage *tuiadapter.Storage) Model {
	m := Model{
		ctx:     ctx,
		key:     key,
		auth:    auth,
		storage: storage,
		btnNames: map[btn]string{
			btnTable:            "Таблица",
			btnLogin:            "Логин",
			btnRegister:         "Регистрация",
			btnFile:             "Ключ...",
			btnSave:             "Сохранить",
			btnDelete:           "Удалить",
			btnFormAddText:      "Добавить Текст",
			btnFormAddLoginPass: "Добавить Логин/Пароль",
			btnFormAddCard:      "Добавить Карту",
		},
		table: newTableModel(),
	}
	updateBtns(&m)
	return m
}

func (m Model) Init() tea.Cmd {
	var c tea.Cmd
	if m.key.GetPrivateKeyPath() == "" {
		c = cmd.ShowFilepicker()
	} else {
		c = cmd.ShowTable()
	}
	return tea.Batch(tea.SetWindowTitle("Gophkeeper"), c)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch tMsg := msg.(type) {
	case tea.WindowSizeMsg:
		tMsg.Height -= 7
		msg = tMsg
	case cmd.MsgMsg:
		m.msg = string(tMsg)
		return m, nil
	case tea.KeyMsg:
		if tMsg.Type == tea.KeyF10 || tMsg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		if m.msg != "" {
			m.msg = ""
			return m, nil
		}
		if tMsg.Type == tea.KeyEsc {
			if m.focused {
				return m, cmd.ScreenFocus()
			}
			return m, cmd.ScreenBlur()
		}
	case cmd.ScreenFocusMsg:
		m.focused = false
	case cmd.ScreenBlurMsg:
		m.cursor = 0
		m.focused = true
	case cmd.UpdateBtnsMsg:
		updateBtns(&m)
		m.initialized = true
		return m, nil
	case cmd.SelectPrivatePathMsg:
		err := m.key.SetPrivateKeyPath(string(tMsg))
		if err != nil {
			return m, cmd.Msg(err.Error())
		}
		return m, nil
	case cmd.ShowFilepickerMsg:
		m.screen = newFilepickerModel()
		return m, tea.Sequence(cmd.ScreenFocus(), m.screen.Init(), tea.WindowSize(), cmd.UpdateBtns())
	case cmd.ShowTableMsg:
		m.screen = m.table
		return m, tea.Sequence(cmd.ScreenFocus(), tea.WindowSize(), cmd.UpdateBtns())
	case cmd.ShowFormIDMsg:
		t, err := m.storage.Type(string(tMsg))
		if err != nil {
			return m, cmd.Msg(err.Error())
		}
		switch t {
		case tuiadapter.DataTypeText:
			data, err := m.storage.DataText(string(tMsg))
			if err != nil {
				return m, cmd.Msg(err.Error())
			}
			return m, cmd.ShowFormText(data.ID, data.Text, data.Meta)
		}
		return m, nil
	case cmd.ShowFormTextMsg:
		m.screen = form.MakeFormTextModel(tMsg.ID, tMsg.Text, tMsg.Meta)
		return m, tea.Sequence(cmd.ScreenFocus(), m.screen.Init(), tea.WindowSize(), cmd.UpdateBtns())
	case tuiadapter.ChangeMsg:
		switch m.screen.(type) {
		case tableModel:
			return m, cmd.TableRows(tMsg.ID, m.storage.List())
		}
		var c tea.Cmd
		m.table, c = m.table.Update(cmd.NewRowsMsg(tMsg.ID, m.storage.List()))
		return m, c
	case cmd.SaveTextMsg:
		if err := m.storage.SaveText(tMsg.ID, tMsg.NewID, tMsg.Text, tMsg.Meta); err != nil {
			return m, cmd.Msg(err.Error())
		}
		return m, cmd.ShowTable()
	case cmd.DeleteIDMsg:
		if err := m.storage.Delete(string(tMsg)); err != nil {
			return m, cmd.Msg(err.Error())
		}
		return m, cmd.ShowTable()
	}

	if m.focused {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			var c tea.Cmd
			switch msg.Type {
			case tea.KeyCtrlD:
				break
			case tea.KeyTab:
				if m.cursor < len(m.btns)-1 {
					m.cursor++
					return m, nil
				} else {
					m.cursor = 0
					return m, cmd.ScreenFocus()
				}
			case tea.KeyUp:
				if m.cursor > 0 {
					m.cursor--
				} else {
					m.cursor = len(m.btns) - 1
				}
				return m, nil
			case tea.KeyDown:
				if m.cursor < len(m.btns)-1 {
					m.cursor++
				} else {
					m.cursor = 0
				}
				return m, c
			case tea.KeyEnter:
				if m.cursor < 0 || m.cursor >= len(m.btns) {
					break
				}
				switch m.btns[m.cursor] {
				case btnTable:
					c = cmd.ShowTable()
				case btnFile:
					c = cmd.ShowFilepicker()
				case btnLogin:
					if err := m.auth.Login(m.ctx); err != nil {
						c = cmd.Msg(err.Error())
					} else {
						c = cmd.ShowTable()
					}
				case btnRegister:
					if err := m.auth.Register(m.ctx); err != nil {
						c = cmd.Msg(err.Error())
					} else {
						c = cmd.ShowTable()
					}
				case btnFormAddText:
					return m, cmd.ShowFormText("", "", tuiadapter.NewMeta(""))
				case btnSave:
					return m, cmd.EventSave()
				case btnDelete:
					return m, cmd.EventDelete()
				default:
				}
				return m, c
			default:
				return m, nil
			}
		case tea.MouseMsg:
			return m, nil
		}
	}

	if m.screen == nil {
		return m, nil
	}

	var c tea.Cmd
	m.screen, c = m.screen.Update(msg)
	switch screenModel := m.screen.(type) {
	case tableModel:
		m.table = screenModel
	}
	return m, c
}
