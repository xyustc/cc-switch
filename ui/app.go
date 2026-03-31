package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type view int

const (
	viewList view = iota
	viewForm
	viewConfirm
)

// switchToListMsg returns to the list view.
type switchToListMsg struct{}

// switchToFormMsg opens the form for add/edit.
type switchToFormMsg struct {
	editName string // empty = new profile
}

// switchToConfirmMsg opens the delete confirmation.
type switchToConfirmMsg struct {
	profileName string
}

// App is the top-level bubbletea model.
type App struct {
	current view
	width   int
	height  int
	list    listModel
	form    formModel
	confirm confirmModel
}

func NewApp() (*App, error) {
	profiles, err := loadProfilesOrExit()
	if err != nil {
		return nil, err
	}
	return &App{
		current: viewList,
		list:    newList(profiles, 60, 20),
	}, nil
}

func (a *App) Init() tea.Cmd {
	return a.list.Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.list.SetSize(msg.Width, msg.Height-4)
		if a.current == viewForm {
			a.form.SetSize(msg.Width, msg.Height)
		}
		return a, nil

	case switchToFormMsg:
		a.form = newForm(a.list.profiles, msg.editName, a.width, a.height)
		a.current = viewForm
		return a, a.form.Init()

	case switchToConfirmMsg:
		a.confirm = newConfirm(
			"删除 profile \""+msg.profileName+"\"？",
			func() tea.Msg { return deleteConfirmedMsg{name: msg.profileName} },
			a.width, a.height,
		)
		a.current = viewConfirm
		return a, nil

	case switchToListMsg:
		profiles, err := loadProfilesOrExit()
		if err != nil {
			return a, tea.Quit
		}
		a.list = newList(profiles, a.width, a.height-4)
		a.current = viewList
		return a, a.list.Init()

	case confirmResultMsg:
		if !msg.confirmed {
			a.current = viewList
		}
		return a, nil

	case deleteConfirmedMsg:
		a.current = viewList
		return a, nil
	}

	switch a.current {
	case viewList:
		m, cmd := a.list.Update(msg)
		a.list = m.(listModel)
		return a, cmd
	case viewForm:
		m, cmd := a.form.Update(msg)
		a.form = m.(formModel)
		return a, cmd
	case viewConfirm:
		m, cmd := a.confirm.Update(msg)
		a.confirm = m.(confirmModel)
		return a, cmd
	}
	return a, nil
}

func (a *App) View() string {
	switch a.current {
	case viewForm:
		return a.form.View()
	case viewConfirm:
		return a.confirm.View()
	default:
		return a.list.View()
	}
}
