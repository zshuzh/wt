package remove

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zshuzh/wt/internal/git"
	"github.com/zshuzh/wt/internal/tui"
)

type successMsg struct{}

func removeWorktree(path string) tea.Cmd {
	return func() tea.Msg {
		err := git.RemoveWorktree(path)
		if err != nil {
			return tui.ErrMsg(err)
		}

		return successMsg{}
	}
}

type model struct {
	tui.Model
}

func (m model) Init() tea.Cmd {
	return m.Model.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case successMsg:
		return m, tea.Quit

	case tea.KeyMsg:
		if msg.String() == "enter" && len(m.Worktrees) > 0 {
			return m, removeWorktree(m.Worktrees[m.Cursor].Path)
		}
	}

	updated, cmd := m.Model.Update(msg)
	m.Model = updated
	return m, cmd
}

func (m model) View() string {
	return m.Model.View()
}
