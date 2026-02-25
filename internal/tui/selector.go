package tui

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
	SelectedStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("212"))

	NormalStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("250"))

	CursorStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("212"))

	BranchStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("243"))

	HelpStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	ErrorStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("196"))
)

type WorktreesMsg struct {
	Worktrees       []git.Worktree
	CurrentWorktree git.Worktree
}

type ErrMsg error

func FetchWorktrees() tea.Msg {
	worktrees, err := git.ListWorktrees()
	if err != nil {
		return ErrMsg(err)
	}

	currentWorktree, err := git.GetCurrentWorktree()
	if err != nil {
		currentWorktree = git.Worktree{}
	}

	return WorktreesMsg{
		Worktrees:       worktrees,
		CurrentWorktree: currentWorktree,
	}
}

type Model struct {
	Worktrees       []git.Worktree
	CurrentWorktree git.Worktree
	Cursor          int
	Loading         bool
	Err             error
	Selected        bool
}

func (m Model) Init() tea.Cmd {
	return FetchWorktrees
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case WorktreesMsg:
		m.Worktrees = msg.Worktrees
		m.CurrentWorktree = msg.CurrentWorktree
		m.Loading = false

		for i, wt := range msg.Worktrees {
			if wt.Path == msg.CurrentWorktree.Path {
				m.Cursor = i
				break
			}
		}

		return m, nil

	case ErrMsg:
		m.Err = msg
		m.Loading = false
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "ctrl+p":
			if m.Cursor > 0 {
				m.Cursor--
			}

		case "down", "ctrl+n":
			if m.Cursor < len(m.Worktrees)-1 {
				m.Cursor++
			}

		case "enter":
			m.Selected = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.Loading {
		return HelpStyle.Render("Loading worktrees...") + "\n"
	}

	if m.Err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v", m.Err)) + "\n"
	}

	if len(m.Worktrees) == 0 {
		return HelpStyle.Render("No worktrees found.") + "\n"
	}

	homeDir, _ := os.UserHomeDir()

	var s string

	for i, wt := range m.Worktrees {
		displayPath := wt.Path
		if homeDir != "" && strings.HasPrefix(displayPath, homeDir) {
			displayPath = "~" + strings.TrimPrefix(displayPath, homeDir)
		}

		if m.Cursor == i {
			cursor := CursorStyle.Render(">")
			path := SelectedStyle.Render(displayPath)
			branch := BranchStyle.Render(wt.Branch)
			s += fmt.Sprintf("%s %s %s %s\n", cursor, path, BranchStyle.Render("·"), branch)
		} else {
			path := NormalStyle.Render(displayPath)
			branch := BranchStyle.Render(wt.Branch)
			s += fmt.Sprintf("  %s %s %s\n", path, BranchStyle.Render("·"), branch)
		}
	}

	return s
}
