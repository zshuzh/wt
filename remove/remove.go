package remove

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zshuzh/wt/internal/git"
)

var renderer = lipgloss.NewRenderer(os.Stderr)

var (
	selectedStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("212"))

	normalStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("250"))

	cursorStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("212"))

	branchStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("243"))

	helpStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	errorStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("196"))
)

type worktreesMsg struct {
	worktrees       []git.Worktree
	currentWorktree git.Worktree
}
type errMsg error
type successMsg struct{}

func getWorktrees() tea.Msg {
	worktrees, err := git.ListWorktrees()
	if err != nil {
		return errMsg(err)
	}

	currentWorktree, err := git.GetCurrentWorktree()
	if err != nil {
		currentWorktree = git.Worktree{}
	}

	return worktreesMsg{
		worktrees:       worktrees,
		currentWorktree: currentWorktree,
	}
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
	worktrees       []git.Worktree
	currentWorktree git.Worktree
	cursor          int
	loading         bool
	err             error
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

		for i, wt := range msg.worktrees {
			if wt.Path == msg.currentWorktree.Path {
				m.cursor = i
				break
			}
		}

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
		return helpStyle.Render("Loading worktrees...") + "\n"
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
	}

	if len(m.worktrees) == 0 {
		return helpStyle.Render("No worktrees found.") + "\n"
	}

	homeDir, _ := os.UserHomeDir()

	var s string

	for i, wt := range m.worktrees {
		displayPath := wt.Path
		if homeDir != "" && strings.HasPrefix(displayPath, homeDir) {
			displayPath = "~" + strings.TrimPrefix(displayPath, homeDir)
		}

		if m.cursor == i {
			cursor := cursorStyle.Render(">")
			path := selectedStyle.Render(displayPath)
			branch := branchStyle.Render(wt.Branch)
			s += fmt.Sprintf("%s %s %s %s\n", cursor, path, branchStyle.Render("·"), branch)
		} else {
			path := normalStyle.Render(displayPath)
			branch := branchStyle.Render(wt.Branch)
			s += fmt.Sprintf("  %s %s %s\n", path, branchStyle.Render("·"), branch)
		}
	}

	return s
}
