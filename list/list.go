package list

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zshuzh/wt/internal/git"
)

type worktreesMsg struct {
	worktrees        []git.Worktree
	currentWorktree  git.Worktree
}

type errMsg error

func getWorktrees() tea.Msg {
	worktrees, err := git.ListWorktrees()
	if err != nil {
		return errMsg(err)
	}

	currentWorktree, err := git.GetCurrentWorktree()
	if err != nil {
		currentWorktree = git.Worktree{} // Not in a worktree or error
	}

	return worktreesMsg{
		worktrees:        worktrees,
		currentWorktree:  currentWorktree,
	}
}

type model struct {
	worktrees        []git.Worktree
	currentWorktree  git.Worktree
	loading          bool
	err              error
}

func (m model) Init() tea.Cmd {
	return getWorktrees
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case worktreesMsg:
		m.worktrees = msg.worktrees
		m.currentWorktree = msg.currentWorktree
		m.loading = false
		return m, tea.Quit

	case errMsg:
		m.err = msg
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
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
		return "No worktrees found.\n"
	}

	s := "Git Worktrees:\n\n"

	for _, wt := range m.worktrees {
		cursor := " "
		if wt.Path == m.currentWorktree.Path {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s - %s\n", cursor, wt.Path, wt.Branch)
	}

	return s
}
