package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xyustc/cc-switch/config"
)

// profileItem implements list.Item.
type profileItem struct {
	profile  config.Profile
	isActive bool
}

func (i profileItem) Title() string {
	if i.isActive {
		return "● " + i.profile.Name
	}
	return "  " + i.profile.Name
}
func (i profileItem) Description() string { return i.profile.Description }
func (i profileItem) FilterValue() string { return i.profile.Name }

type deleteConfirmedMsg struct{ name string }

type listModel struct {
	list     list.Model
	profiles *config.Profiles
	err      string
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Padding(0, 1)

	activeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

func newList(profiles *config.Profiles) listModel {
	items := make([]list.Item, len(profiles.Profiles))
	for i, p := range profiles.Profiles {
		items[i] = profileItem{
			profile:  p,
			isActive: p.Name == profiles.Active,
		}
	}
	l := list.New(items, list.NewDefaultDelegate(), 60, 20)
	l.Title = "cc-switch"
	l.Styles.Title = titleStyle
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	// custom help shown in View() instead
	return listModel{list: l, profiles: profiles}
}

func (m listModel) Init() tea.Cmd { return nil }

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m, m.switchSelected()
		case "n":
			return m, func() tea.Msg { return switchToFormMsg{} }
		case "e":
			if item, ok := m.list.SelectedItem().(profileItem); ok {
				return m, func() tea.Msg { return switchToFormMsg{editName: item.profile.Name} }
			}
		case "d":
			if item, ok := m.list.SelectedItem().(profileItem); ok {
				name := item.profile.Name
				return m, func() tea.Msg { return switchToConfirmMsg{profileName: name} }
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case deleteConfirmedMsg:
		if err := config.DeleteProfile(m.profiles, msg.name); err != nil {
			m.err = err.Error()
			return m, nil
		}
		return m, func() tea.Msg { return switchToListMsg{} }

	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-4)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listModel) switchSelected() tea.Cmd {
	item, ok := m.list.SelectedItem().(profileItem)
	if !ok {
		return nil
	}
	name := item.profile.Name
	return func() tea.Msg {
		// Find the profile
		var target config.Profile
		for _, p := range m.profiles.Profiles {
			if p.Name == name {
				target = p
				break
			}
		}
		if err := config.ApplyProfile(target); err != nil {
			return errMsg{err}
		}
		m.profiles.Active = name
		if err := config.SaveProfiles(m.profiles); err != nil {
			return errMsg{err}
		}
		return switchToListMsg{}
	}
}

type errMsg struct{ err error }

func (m listModel) View() string {
	help := helpStyle.Render("[↑↓] 导航  [Enter] 切换  [n] 新增  [e] 编辑  [d] 删除  [q] 退出")
	errLine := ""
	if m.err != "" {
		errLine = "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("错误: "+m.err)
	}
	return fmt.Sprintf("%s\n%s%s", m.list.View(), help, errLine)
}

func loadProfilesOrExit() (*config.Profiles, error) {
	return config.LoadProfiles()
}
