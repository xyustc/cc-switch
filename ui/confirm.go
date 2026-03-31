package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type confirmModel struct {
	message  string
	onYes    tea.Cmd
	onCancel tea.Cmd
}

type confirmResultMsg struct {
	confirmed bool
}

func newConfirm(message string, onYes tea.Cmd) confirmModel {
	return confirmModel{message: message, onYes: onYes}
}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "y", "Y":
			return m, m.onYes
		case "n", "N", "esc":
			return m, func() tea.Msg { return confirmResultMsg{confirmed: false} }
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		BorderForeground(lipgloss.Color("9"))
	return style.Render(m.message + "\n\n[y] 确认  [n/Esc] 取消")
}
