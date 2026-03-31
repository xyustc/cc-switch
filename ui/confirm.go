package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type confirmModel struct {
	message  string
	onYes    tea.Cmd
	width    int
	height   int
}

type confirmResultMsg struct {
	confirmed bool
}

func newConfirm(message string, onYes tea.Cmd, width, height int) confirmModel {
	return confirmModel{message: message, onYes: onYes, width: width, height: height}
}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			return m, m.onYes
		case "n", "N", "esc":
			return m, func() tea.Msg { return confirmResultMsg{confirmed: false} }
		}

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			// Check if click is on [y] or [n] area
			// The confirm dialog is centered, approximate positions
			content := m.View()
			lines := strings.Split(content, "\n")
			for i, line := range lines {
				if strings.Contains(line, "[y]") && msg.Y == i {
					// Click on the button line
					linePos := strings.Index(line, "[y]")
					if msg.X >= linePos && msg.X < linePos+3 {
						return m, m.onYes
					}
					linePosN := strings.Index(line, "[n]")
					if msg.X >= linePosN && msg.X < linePosN+3 {
						return m, func() tea.Msg { return confirmResultMsg{confirmed: false} }
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m confirmModel) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		BorderForeground(lipgloss.Color("9"))

	buttonStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	buttons := buttonStyle.Render("[y]") + " 确认  " + buttonStyle.Render("[n]") + " 取消"
	return style.Render(m.message + "\n\n" + buttons)
}
