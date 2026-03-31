package ui

import (
	"encoding/json"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xyustc/cc-switch/config"
)

type formField int

const (
	fieldName formField = iota
	fieldDesc
	fieldSettings
)

type formModel struct {
	profiles  *config.Profiles
	editName  string // empty = new profile
	field     formField
	name      textinput.Model
	desc      textinput.Model
	settings  textarea.Model
	errMsg    string
	width     int
	height    int
}

func newForm(profiles *config.Profiles, editName string, width, height int) formModel {
	name := textinput.New()
	name.Placeholder = "profile 名称 (a-z, 0-9, _, -)"
	name.Focus()

	desc := textinput.New()
	desc.Placeholder = "描述（可选）"

	ta := textarea.New()
	ta.Placeholder = `{"env": {"ANTHROPIC_AUTH_TOKEN": "sk-..."}}`
	ta.SetHeight(8)
	if width > 20 {
		ta.SetWidth(width - 4)
	} else {
		ta.SetWidth(60)
	}

	m := formModel{
		profiles: profiles,
		editName: editName,
		field:    fieldName,
		name:     name,
		desc:     desc,
		settings: ta,
		width:    width,
		height:   height,
	}

	// Pre-fill when editing
	if editName != "" {
		for _, p := range profiles.Profiles {
			if p.Name == editName {
				m.name.SetValue(p.Name)
				m.desc.SetValue(p.Description)
				data, _ := json.MarshalIndent(p.Settings, "", "  ")
				m.settings.SetValue(string(data))
				break
			}
		}
	}
	return m
}

func (m *formModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	if width > 20 {
		m.settings.SetWidth(width - 4)
	}
}

func (m formModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			// Check which field was clicked based on Y position
			// Approximate Y positions (calculated from View)
			nameY := 2          // After title
			descY := nameY + 3  // After name field
			settingsY := descY + 3 // After desc field

			clickY := msg.Y
			switch {
			case clickY >= nameY && clickY < nameY+2:
				m.setFocus(fieldName)
			case clickY >= descY && clickY < descY+2:
				m.setFocus(fieldDesc)
			case clickY >= settingsY:
				m.setFocus(fieldSettings)
			}
		}
		return m.updateInputs(msg)

	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil
	}

	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m.updateInputs(msg)
	}

	switch key.String() {
	case "esc":
		return m, func() tea.Msg { return switchToListMsg{} }

	case "tab", "shift+tab":
		if key.String() == "tab" {
			m.field = (m.field + 1) % 3
		} else {
			m.field = (m.field + 2) % 3
		}
		m.updateFocus()
		return m, nil

	case "enter":
		if m.field != fieldSettings {
			// Tab to next field on Enter (except in textarea)
			m.field = (m.field + 1) % 3
			m.updateFocus()
			return m, nil
		}

	case "ctrl+s":
		return m.save()
	}

	return m.updateInputs(msg)
}

func (m *formModel) setFocus(field formField) {
	m.field = field
	m.updateFocus()
}

func (m *formModel) updateFocus() {
	m.name.Blur()
	m.desc.Blur()
	m.settings.Blur()
	switch m.field {
	case fieldName:
		m.name.Focus()
	case fieldDesc:
		m.desc.Focus()
	case fieldSettings:
		m.settings.Focus()
	}
}

func (m formModel) save() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.name.Value())
	desc := strings.TrimSpace(m.desc.Value())
	settingsRaw := strings.TrimSpace(m.settings.Value())

	// Validate settings JSON
	var settingsMap map[string]interface{}
	if settingsRaw == "" {
		settingsMap = map[string]interface{}{}
	} else if err := json.Unmarshal([]byte(settingsRaw), &settingsMap); err != nil {
		m.errMsg = "settings JSON 格式错误: " + err.Error()
		return m, nil
	}

	p := config.Profile{
		Name:        name,
		Description: desc,
		Settings:    settingsMap,
	}

	var err error
	if m.editName == "" {
		err = config.AddProfile(m.profiles, p)
	} else {
		err = config.UpdateProfile(m.profiles, m.editName, p)
	}
	if err != nil {
		m.errMsg = err.Error()
		return m, nil
	}
	return m, func() tea.Msg { return switchToListMsg{} }
}

func (m formModel) updateInputs(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.name, cmd = m.name.Update(msg)
	cmds = append(cmds, cmd)
	m.desc, cmd = m.desc.Update(msg)
	cmds = append(cmds, cmd)
	m.settings, cmd = m.settings.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

var (
	labelStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	activeField = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	formHelp    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func (m formModel) View() string {
	title := "新增 Profile"
	if m.editName != "" {
		title = "编辑 Profile: " + m.editName
	}

	var b strings.Builder
	b.WriteString(labelStyle.Render(title) + "\n\n")

	b.WriteString(fieldLabel("名称", m.field == fieldName) + "\n")
	b.WriteString(m.name.View() + "\n\n")

	b.WriteString(fieldLabel("描述", m.field == fieldDesc) + "\n")
	b.WriteString(m.desc.View() + "\n\n")

	b.WriteString(fieldLabel("Settings JSON", m.field == fieldSettings) + "\n")
	b.WriteString(m.settings.View() + "\n\n")

	if m.errMsg != "" {
		b.WriteString(errStyle.Render("✗ "+m.errMsg) + "\n\n")
	}

	b.WriteString(formHelp.Render("[Tab] 切换字段  [Ctrl+S] 保存  [Esc] 取消"))
	return b.String()
}

func fieldLabel(label string, active bool) string {
	if active {
		return activeField.Render("> " + label)
	}
	return "  " + label
}
