package ui

import (
	"fmt"
	"time"

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
	list        list.Model
	profiles    *config.Profiles
	err         string
	width       int
	height      int
	lastClick   time.Time
	lastClickY  int
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

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			Underline(true)

	// List item dimensions (must match DefaultDelegate settings)
	itemHeight  = 3 // delegate height (2) + spacing (1)
	titleHeight = 1
	// Double-click detection threshold
	doubleClickThreshold = 300 * time.Millisecond
)

func newList(profiles *config.Profiles, width, height int) listModel {
	items := make([]list.Item, len(profiles.Profiles))
	for i, p := range profiles.Profiles {
		items[i] = profileItem{
			profile:  p,
			isActive: p.Name == profiles.Active,
		}
	}
	l := list.New(items, list.NewDefaultDelegate(), width, height)
	l.Title = "cc-switch"
	l.Styles.Title = titleStyle
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	return listModel{list: l, profiles: profiles, width: width, height: height}
}

func (m *listModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
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

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			// Help area is at the bottom (after list content)
			// List height is m.height, help line is at Y = m.height
			if msg.Y >= m.height {
				// Click in help area - check which button
				btnX := 0
				buttons := []struct {
					label string
					key   string
				}{
					{"[n] 新增", "n"},
					{"[e] 编辑", "e"},
					{"[d] 删除", "d"},
					{"[q] 退出", "q"},
				}
				for _, btn := range buttons {
					btnX += len(btn.label) + 2
					if msg.X < btnX {
						switch btn.key {
						case "n":
							return m, func() tea.Msg { return switchToFormMsg{} }
						case "e":
							if item, ok := m.list.SelectedItem().(profileItem); ok {
								return m, func() tea.Msg { return switchToFormMsg{editName: item.profile.Name} }
							}
						case "d":
							if item, ok := m.list.SelectedItem().(profileItem); ok {
								return m, func() tea.Msg { return switchToConfirmMsg{profileName: item.profile.Name} }
							}
						case "q":
							return m, tea.Quit
						}
						break
					}
				}
				return m, nil
			}

			// Click in list area - calculate which item was clicked
			// List view structure: title area + items

			// Calculate item index from click position
			clickY := msg.Y
			if clickY < titleHeight {
				// Clicked on title, ignore
				return m, nil
			}

			itemIndex := (clickY - titleHeight) / itemHeight
			if itemIndex >= 0 && itemIndex < len(m.list.Items()) {
				// Double-click detection (within threshold on same item)
				now := time.Now()
				if now.Sub(m.lastClick) < doubleClickThreshold && itemIndex == m.lastClickY {
					// Double click - switch profile
					m.list.Select(itemIndex)
					return m, m.switchSelected()
				}
				m.lastClick = now
				m.lastClickY = itemIndex

				// Single click - select item
				m.list.Select(itemIndex)
			}
			return m, nil
		}

	case deleteConfirmedMsg:
		if err := config.DeleteProfile(m.profiles, msg.name); err != nil {
			m.err = err.Error()
			return m, nil
		}
		return m, func() tea.Msg { return switchToListMsg{} }

	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height-4)
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
	help := helpStyle.Render("[↑↓] 导航  [Enter] 切换  ") +
		buttonStyle.Render("[n] 新增") + "  " +
		buttonStyle.Render("[e] 编辑") + "  " +
		buttonStyle.Render("[d] 删除") + "  " +
		buttonStyle.Render("[q] 退出")
	errLine := ""
	if m.err != "" {
		errLine = "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("错误: "+m.err)
	}
	return fmt.Sprintf("%s\n%s%s", m.list.View(), help, errLine)
}

func loadProfilesOrExit() (*config.Profiles, error) {
	return config.LoadProfiles()
}
