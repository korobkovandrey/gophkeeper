package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type appService interface {
	GetPrivateKeyPath() string
	SetPrivateKeyPath(privateKeyPath string) error
	Login(ctx context.Context) error
	Register(ctx context.Context) error
	IsLogged() bool
	IsOnline() bool
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
	btnFormAddFile
)

type Model struct {
	ctx context.Context
	app appService

	msg    string
	table  tea.Model
	screen tea.Model

	btnsFocus   bool
	btns        []btn
	btnsCursor  int
	btnNames    map[btn]string
	debug       string
	initialized bool
}

func NewModel(ctx context.Context, app appService) Model {
	m := Model{
		ctx: ctx,
		app: app,
		btnNames: map[btn]string{
			btnTable:            "Таблица",
			btnLogin:            "Логин",
			btnRegister:         "Регистрация",
			btnFile:             "Файл...",
			btnFormAddText:      "Добавить Текст",
			btnFormAddLoginPass: "Добавить Логин/Пароль",
			btnFormAddCard:      "Добавить Карту",
			btnFormAddFile:      "Добавить файл",
		},
		table: newTableModel(),
	}
	updateBtns(&m)
	return m
}

func (m Model) Init() tea.Cmd {
	if m.app.GetPrivateKeyPath() != "" {
		return newShowFilepickerCmd()
	}
	return newShowTableCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch tMsg := msg.(type) {
	case tea.WindowSizeMsg:
		tMsg.Height -= 4
		msg = tMsg
	case tea.KeyMsg:
		if tMsg.Type == tea.KeyF10 || tMsg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		if m.msg != "" {
			m.msg = ""
			return m, nil
		}
	case btnsFocusMsg:
		m.btnsFocus = true
		return m, nil
	case updateBtnsMsg:
		updateBtns(&m)
		m.initialized = true
		return m, nil
	case selectPrivatePathMsg:
		err := m.app.SetPrivateKeyPath(string(tMsg))
		if err != nil {
			setMsg(&m, err.Error())
		}
		return m, nil
	case showFilepickerMsg:
		m.btnsCursor = 0
		m.btnsFocus = false
		m.screen = newFilepickerModel()
		return m, tea.Sequence(m.screen.Init(), tea.WindowSize(), newUpdateBtnsCmd())
	case showTableMsg:
		m.btnsCursor = 0
		m.btnsFocus = false
		m.screen = m.table
		return m, tea.Sequence(tea.WindowSize(), newUpdateBtnsCmd())
	case changeSecretsMsg:
		//secrets := m.app.Secrets()
		rows := []table.Row{
			{"статус1", "ид1", "деск1", "created_at1", "updated_at1"},
			{"статус2", "ид2", "деск2", "created_at2", "updated_at2"},
			{"статус3", "ид2", "деск2", "created_at2", "updated_at2"},
			{"статус4", "ид2", "деск2", "created_at2", "updated_at2"},
			{"статус5", "ид2", "деск2", "created_at2", "updated_at2"},
			{"статус6", "ид2", "деск2", "created_at2", "updated_at2"},
			{"статус7", "ид2", "деск2", "created_at2", "updated_at2"},
			{"статус8", "ид2", "деск2", "created_at2", "updated_at2"},
			{"статус9", "ид2", "деск2", "created_at2", "updated_at2"},
			{"статус10", "ид2", "деск2", "created_at2", "updated_at2"},
			{"статус11", "ид2", "деск2", "created_at2", "updated_at2"},
			{"статус12", "ид2", "деск2", "created_at2", "updated_at2"},
			{"статус13", "ид2", "деск2", "created_at2", "updated_at2"},
			{"статус14", "ид2", "деск2", "created_at2", "updated_at2"},
		}
		switch m.screen.(type) {
		case tableModel:
			return m, newCmd(newTableRowsMsg(rows, 0))
		}
		var c tea.Cmd
		m.table, c = m.table.Update(newTableRowsMsg(rows, 0))
		return m, c
	case showFormTextMsg:

	}

	if m.screen == nil {
		return m, nil
	}

	if m.btnsFocus {
		switch tMsg := msg.(type) {
		case tea.KeyMsg:
			var cmd tea.Cmd
			switch tMsg.Type {
			case tea.KeyEsc:
				m.btnsFocus = false
			case tea.KeyTab:
				if m.btnsCursor < len(m.btns)-1 {
					m.btnsCursor++
				} else {
					m.btnsCursor = 0
					m.btnsFocus = false
				}
			case tea.KeyEnter:
				if m.btnsCursor < 0 || m.btnsCursor >= len(m.btns) {
					break
				}
				switch m.btns[m.btnsCursor] {
				case btnTable:
					cmd = newShowTableCmd()
				case btnFile:
					cmd = newShowFilepickerCmd()
				case btnLogin:
					if err := m.app.Login(m.ctx); err != nil {
						setMsg(&m, err.Error())
					} else {
						cmd = newShowTableCmd()
					}
				case btnRegister:
					if err := m.app.Register(m.ctx); err != nil {
						setMsg(&m, err.Error())
					} else {
						cmd = newShowTableCmd()
					}
				default:
				}
			}
			return m, cmd
		case tea.MouseMsg:
			return m, nil
		}
	}

	var c tea.Cmd
	m.screen, c = m.screen.Update(msg)
	switch screenModel := m.screen.(type) {
	case tableModel:
		m.table = screenModel
	}
	return m, c
}
