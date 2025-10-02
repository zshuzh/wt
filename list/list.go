package list

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Worktree struct {
	Path   string
	Branch string
}

type worktreesMsg []Worktree
type errMsg error

func getWorktrees() tea.Msg {
	output, err := exec.Command("git", "worktree", "list", "--porcelain").Output()
	if err != nil {
		return errMsg(err)
	}

	var worktrees []Worktree
	var currentPath string
	var currentBranch string

	for line := range strings.SplitSeq(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" { // blank line separates worktrees
			if currentPath != "" {
				worktrees = append(worktrees, Worktree{
					Path:   currentPath,
					Branch: currentBranch,
				})
				currentPath = ""
				currentBranch = ""
			}
			continue
		}

		if path, ok := strings.CutPrefix(line, "worktree "); ok {
			currentPath = path
		} else if branch, ok := strings.CutPrefix(line, "branch "); ok {
			currentBranch = strings.TrimPrefix(branch, "refs/heads/")
		}
	}

	return worktreesMsg(worktrees)
}

type model struct {
	worktrees []Worktree
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
		return "No worktrees found.\n\nPress q to quit.\n"
	}

	s := "Git Worktrees:\n\n"

	for _, wt := range m.worktrees {
		s += fmt.Sprintf("%s - %s\n", wt.Path, wt.Branch)
	}

	s += "\nPress q to quit.\n"

	return s
}
