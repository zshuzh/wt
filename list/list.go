package list

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zshuzh/wt/internal/git"
)

var (
	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250"))

	branchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)

type worktreesMsg struct {
	worktrees       []git.Worktree
	currentWorktree git.Worktree
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
		return helpStyle.Render("Loading worktrees...") + "\n"
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
	}

	if len(m.worktrees) == 0 {
		return helpStyle.Render("No worktrees found.") + "\n"
	}

	var s string

	for _, wt := range m.worktrees {
		sep := separatorStyle.Render("·")
		if wt.Path == m.currentWorktree.Path {
			cursor := cursorStyle.Render(">")
			path := selectedStyle.Render(wt.Path)
			branch := branchStyle.Render(wt.Branch)
			s += fmt.Sprintf("%s %s %s %s\n", cursor, path, sep, branch)
		} else {
			path := pathStyle.Render(wt.Path)
			branch := branchStyle.Render(wt.Branch)
			s += fmt.Sprintf("  %s %s %s\n", path, sep, branch)
		}
	}

	return s
}
