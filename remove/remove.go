package remove

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zshuzh/wt/internal/git"
	"github.com/zshuzh/wt/internal/tui"
)

type successMsg struct {
	path   string
	branch string
}

func removeWorktree(path, branch string, force bool) tea.Cmd {
	return func() tea.Msg {
		err := git.RemoveWorktree(path, force)
		if err != nil {
			return tui.ErrMsg(err)
		}

		return successMsg{path: path, branch: branch}
	}
}

type model struct {
	tui.Model
	force    bool
	removing bool
	spinner  spinner.Model
	result   *successMsg
}

func (m model) Init() tea.Cmd {
	return m.Model.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case successMsg:
		m.result = &msg
		m.removing = false
		return m, tea.Quit

	case spinner.TickMsg:
		if m.removing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		if msg.String() == "enter" && len(m.Worktrees) > 0 && !m.removing {
			wt := m.Worktrees[m.Cursor]
			m.removing = true
			return m, tea.Batch(
				m.spinner.Tick,
				removeWorktree(wt.Path, wt.Branch, m.force),
			)
		}
	}

	updated, cmd := m.Model.Update(msg)
	m.Model = updated
	return m, cmd
}

func (m model) View() string {
	if m.result != nil {
		name := filepath.Base(m.result.path)
		msg := fmt.Sprintf("✓ Removed %s", name)
		if m.result.branch != "" {
			msg += fmt.Sprintf(" (branch: %s)", m.result.branch)
		}
		return lipgloss.NewRenderer(os.Stderr).
			NewStyle().
			Foreground(lipgloss.Color("green")).
			Render(msg) + "\n"
	}

	if m.removing {
		name := filepath.Base(m.Worktrees[m.Cursor].Path)
		return m.spinner.View() + fmt.Sprintf(" Removing %s...\n", name)
	}

	return m.Model.View()
}
