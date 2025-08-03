package tui

import (
	"context"
	"gophkeeper/internal/client/tui/cmd"
	"gophkeeper/internal/client/tui/form"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type appService interface {
	GetPrivateKeyPath() string
	Fingerprint() string
	SetPrivateKeyPath(privateKeyPath string) error
	Login(ctx context.Context) error
	Register(ctx context.Context) error
	IsLogged() bool
	IsOnline() bool
	//List() [][]string
	//Secrets() []model.Secret
}

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

type Model struct {
	ctx context.Context
	app appService

	msg    string
	table  tea.Model
	screen tea.Model

	focused     bool
	cursor      int
	btns        []btn
	btnNames    map[btn]string
	debug       string
	initialized bool
	formIsValid bool
	formIsNew   bool
	selectID    string
}

func NewModel(ctx context.Context, app appService) Model {
	m := Model{
		ctx: ctx,
		app: app,
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
	if m.app.GetPrivateKeyPath() == "" {
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
	case cmd.FormValidMsg:
		m.formIsValid = tMsg.IsValid
		updateBtns(&m)
		return m, nil
	case cmd.SelectPrivatePathMsg:
		err := m.app.SetPrivateKeyPath(string(tMsg))
		if err != nil {
			setMsg(&m, err.Error())
		}
		return m, nil
	case cmd.ShowFilepickerMsg:
		m.screen = newFilepickerModel()
		return m, tea.Sequence(cmd.ScreenFocus(), m.screen.Init(), tea.WindowSize(), cmd.UpdateBtns())
	case cmd.ShowTableMsg:
		m.screen = m.table
		return m, tea.Sequence(cmd.ScreenFocus(), tea.WindowSize(), cmd.UpdateBtns())
	case cmd.ShowFormTextMsg:
		m.formIsValid = false
		m.formIsNew = true
		m.screen = form.MakeFormTextModel(tMsg.ID, tMsg.Text, tMsg.MetaKeys, tMsg.MetaVals)
		return m, tea.Sequence(cmd.ScreenFocus(), m.screen.Init(), tea.WindowSize(), cmd.UpdateBtns())
	case cmd.SelectIDMsg:
		m.selectID = string(tMsg)
	case cmd.ChangeSecretsMsg:
		//secrets := m.app.Secrets()
		rows := []table.Row{
			{"статус1", "ид1", "деск1", "created_at1", "updated_at1"},
			{"статус2", "ид2", "деск2", "created_at2", "updated_at2"},
			{"статус3", "ид3", "деск2", "created_at2", "updated_at2"},
			{"статус4", "ид4", "деск2", "created_at2", "updated_at2"},
			{"статус5", "ид5", "деск2", "created_at2", "updated_at2"},
			{"статус6", "ид6", "деск2", "created_at2", "updated_at2"},
			{"статус7", "ид7", "деск2", "created_at2", "updated_at2"},
			{"статус8", "ид8", "деск2", "created_at2", "updated_at2"},
			{"статус9", "ид9", "деск2", "created_at2", "updated_at2"},
			{"статус10", "ид10", "деск2", "created_at2", "updated_at2"},
			{"статус11", "ид2", "деск2", "created_at2", "updated_at2"},
			{"статус12", "ид2", "деск2", "created_at2", "updated_at2"},
			{"статус13", "ид2", "деск2", "created_at2", "updated_at2"},
			{"статус14", "ид2", "деск2", "created_at2", "updated_at2"},
		}
		switch m.screen.(type) {
		case tableModel:
			return m, cmd.TableRows(rows, 0)
		}
		var c tea.Cmd
		m.table, c = m.table.Update(cmd.NewTableRowsMsg(rows, 0))
		return m, c
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
					if err := m.app.Login(m.ctx); err != nil {
						setMsg(&m, err.Error())
					} else {
						c = cmd.ShowTable()
					}
				case btnRegister:
					if err := m.app.Register(m.ctx); err != nil {
						setMsg(&m, err.Error())
					} else {
						c = cmd.ShowTable()
					}
				case btnFormAddText:
					return m, cmd.NewShowFormTextCmd("", "", []string{""}, []string{""})
				case btnSave:
				case btnDelete:

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
