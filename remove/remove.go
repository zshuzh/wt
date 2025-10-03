package remove

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zshuzh/wt/internal/git"
)

type worktreesMsg []git.Worktree
type errMsg error
type successMsg struct{}

func getWorktrees() tea.Msg {
	worktrees, err := git.ListWorktrees()
	if err != nil {
		return errMsg(err)
	}
	return worktreesMsg(worktrees)
}

func removeWorktree(path string) tea.Cmd {
	return func() tea.Msg {
		err := git.RemoveWorktree(path)
		if err != nil {
			return errMsg(err)
		}

		return successMsg{}
	}
}

type model struct {
	worktrees []git.Worktree
	cursor    int
	loading   bool
	err       error
}

func (m model) Init() tea.Cmd {
	return getWorktrees
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case worktreesMsg:
		m.worktrees = msg
		m.loading = false
		return m, nil

	case errMsg:
		m.err = msg
		m.loading = false
		return m, nil

	case successMsg:
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "ctrl+n":
			if m.cursor < len(m.worktrees)-1 {
				m.cursor++
			}

		case "enter":
			return m, removeWorktree(m.worktrees[m.cursor].Path)
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.loading {
		return "Loading worktrees...\n"
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.\n", m.err)
	}

	if len(m.worktrees) == 0 {
		return "No worktrees found.\n\nPress q to quit.\n"
	}

	s := "Select a worktree to remove:\n\n"

	for i, wt := range m.worktrees {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %s - %s\n", cursor, wt.Path, wt.Branch)
	}

	s += "\nPress enter to remove, q to quit.\n"

	return s
}
