package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidbudnick/redis-tui/internal/types"
)

func (m Model) handleMasterPasswordScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.Loading = false
		m.ConnectionError = ""
		m.Screen = types.ScreenConnections

		m.StatusMsg = "Warning: Continuing without Secret Store. Secrets will not be saved, or retreived."
		return m, m.Cmds.LoadConnections()
	case "enter":
		pwd := m.MasterPasswordInput.Value()

		if pwd != "" {
			m.Loading = true
			m.Screen = types.ScreenConnections
			return m, m.Cmds.InitializeFileStore(m.ConfigDir, pwd)
		}
	default:
		var inputCmd tea.Cmd
		m.MasterPasswordInput, inputCmd = m.MasterPasswordInput.Update(msg)
		return m, inputCmd
	}
	return m, nil
}
