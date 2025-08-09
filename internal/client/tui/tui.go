package tui

import (
	"context"
	"gophkeeper/internal/client/app"
	"gophkeeper/internal/client/model"
	"gophkeeper/internal/client/service"
	"gophkeeper/internal/client/tui/cmd"
	"gophkeeper/internal/client/tui/form"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	ctx     context.Context
	key     *app.KeyManager
	app     *app.App
	storage *service.Storage

	msg    string
	table  tea.Model
	screen tea.Model

	focused     bool
	cursor      int
	btns        []btn
	btnNames    map[btn]string
	debug       any
	initialized bool
	online      bool
}

func NewModel(ctx context.Context, key *app.KeyManager, app *app.App, storage *service.Storage) Model {
	m := Model{
		ctx:     ctx,
		key:     key,
		app:     app,
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
	return tea.Batch(tea.SetWindowTitle("Gophkeeper"), c, m.listenOnline(), m.listenUpdate())
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
			if _, ok := m.screen.(filepickerModel); !ok {
				return m, cmd.ScreenBlur()
			}
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
		m.storage.Clear()
		return m, cmd.UpdateTable("", m.storage.List())
	case cmd.ShowFilepickerMsg:
		m.screen = newFilepickerModel()
		return m, tea.Sequence(cmd.ScreenFocus(), m.screen.Init(), tea.WindowSize(), cmd.UpdateBtns())
	case cmd.ShowTableMsg:
		m.screen = m.table
		return m, tea.Sequence(cmd.ScreenFocus(), tea.WindowSize(), cmd.UpdateBtns())
	case cmd.ShowFormIDMsg:
		secret, err := m.storage.Get(tMsg.ID)
		if err != nil {
			return m, cmd.Msg(err.Error())
		}
		switch secret.Type {
		case model.TypeText:
			data, meta, err := m.storage.DataText(*secret)
			if err != nil {
				return m, cmd.Msg(err.Error())
			}
			return m, cmd.ShowFormText(secret.ID, secret.NewID, meta, data.Text)
		case model.TypeLoginPass:
			data, meta, err := m.storage.DataLoginPass(*secret)
			if err != nil {
				return m, cmd.Msg(err.Error())
			}
			return m, cmd.ShowFormLoginPass(secret.ID, secret.NewID, meta, data.Login, data.Pass)
		case model.TypeCard:
			data, meta, err := m.storage.DataCard(*secret)
			if err != nil {
				return m, cmd.Msg(err.Error())
			}
			return m, cmd.ShowFormCard(secret.ID, secret.NewID, meta, data.CCN, data.Expire, data.CVV)
		default:
		}
		return m, nil
	case cmd.ShowFormTextMsg:
		m.screen = form.MakeFormTextModel(tMsg.ID, tMsg.NewID, tMsg.Meta, tMsg.Text)
		return m, tea.Sequence(cmd.ScreenFocus(), m.screen.Init(), tea.WindowSize(), cmd.UpdateBtns())
	case cmd.ShowFormLoginPassMsg:
		m.screen = form.MakeFormLoginPassModel(tMsg.ID, tMsg.NewID, tMsg.Meta, tMsg.Login, tMsg.Pass)
		return m, tea.Sequence(cmd.ScreenFocus(), m.screen.Init(), tea.WindowSize(), cmd.UpdateBtns())
	case cmd.ShowFormCardMsg:
		m.screen = form.MakeFormCardModel(tMsg.ID, tMsg.NewID, tMsg.Meta, tMsg.CCN, tMsg.Expire, tMsg.CVV)
		return m, tea.Sequence(cmd.ScreenFocus(), m.screen.Init(), tea.WindowSize(), cmd.UpdateBtns())
	case cmd.UpdateTableMsg:
		if _, ok := m.screen.(tableModel); !ok {
			m.table, _ = m.table.Update(tMsg)
			return m, nil
		}
	case cmd.SaveTextMsg:
		if err := m.storage.SaveText(tMsg.ID, tMsg.NewID, tMsg.Meta, tMsg.Text); err != nil {
			return m, cmd.Msg(err.Error())
		}
		return m, tea.Sequence(cmd.ShowTable(), cmd.UpdateTable(tMsg.ID, m.storage.List()))
	case cmd.SaveLoginPassMsg:
		if err := m.storage.SaveLoginPass(tMsg.ID, tMsg.NewID, tMsg.Meta, tMsg.Login, tMsg.Pass); err != nil {
			return m, cmd.Msg(err.Error())
		}
		return m, tea.Sequence(cmd.ShowTable(), cmd.UpdateTable(tMsg.ID, m.storage.List()))
	case cmd.SaveCardMsg:
		if err := m.storage.SaveCard(tMsg.ID, tMsg.NewID, tMsg.Meta, tMsg.CCN, tMsg.Expire, tMsg.CVV); err != nil {
			return m, cmd.Msg(err.Error())
		}
		return m, tea.Sequence(cmd.ShowTable(), cmd.UpdateTable(tMsg.ID, m.storage.List()))
	case cmd.DeleteIDMsg:
		m.storage.Delete(tMsg.ID)
		return m, tea.Sequence(cmd.ShowTable(), cmd.UpdateTable("", m.storage.List()))
	case OnlineMsg:
		m.online = m.app.IsOnline()
		return m, m.listenOnline()
	case UpdateMsg:
		return m, tea.Batch(cmd.UpdateTable("", m.storage.List()), m.listenUpdate())
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
			case tea.KeyUp, tea.KeyLeft:
				if m.cursor > 0 {
					m.cursor--
				} else {
					m.cursor = len(m.btns) - 1
				}
				return m, nil
			case tea.KeyDown, tea.KeyRight:
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
						c = cmd.Msg(err.Error())
					} else {
						c = cmd.ShowTable()
					}
				case btnRegister:
					if err := m.app.Register(m.ctx); err != nil {
						c = cmd.Msg(err.Error())
					} else {
						c = cmd.ShowTable()
					}
				case btnFormAddText:
					return m, cmd.ShowFormText("", "", model.NewMeta(""), "")
				case btnFormAddLoginPass:
					return m, cmd.ShowFormLoginPass("", "", model.NewMeta(""), "", "")
				case btnFormAddCard:
					return m, cmd.ShowFormCard("", "", model.NewMeta(""), "", "", "")
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
	if t, ok := m.screen.(tableModel); ok {
		m.table = t
	}
	return m, c
}

type OnlineMsg struct{}
type UpdateMsg struct{}

func listenEvent[T any](ch <-chan struct{}) tea.Cmd {
	return func() tea.Msg {
		if _, ok := <-ch; ok {
			var t T
			return t
		}
		return nil
	}
}

func (m Model) listenOnline() tea.Cmd {
	return listenEvent[OnlineMsg](m.app.OnlineCh)
}

func (m Model) listenUpdate() tea.Cmd {
	return listenEvent[UpdateMsg](m.app.UpdateCh)
}
