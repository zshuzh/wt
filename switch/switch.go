package switchcmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zshuzh/wt/internal/tui"
)

type model struct {
	tui.Model
}

func (m model) Init() tea.Cmd {
	return m.Model.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := m.Model.Update(msg)
	m.Model = updated
	return m, cmd
}

func (m model) View() string {
	return m.Model.View()
}
