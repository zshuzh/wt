package list

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/zshuzh/wt/internal/git"
)

type Options struct{}

func (o Options) Run() error {
	worktrees, err := git.ListWorktrees()
	if err != nil {
		return err
	}

	current, err := git.GetCurrentWorktree()
	if err != nil {
		current = git.Worktree{}
	}

	homeDir, _ := os.UserHomeDir()

	currentPathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("212"))

	currentBranchStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))

	dotStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))

	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250"))

	branchStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))

	// Find max path length for alignment
	maxPathLen := 0
	for _, wt := range worktrees {
		displayPath := wt.Path
		if homeDir != "" && strings.HasPrefix(wt.Path, homeDir) {
			displayPath = "~" + strings.TrimPrefix(wt.Path, homeDir)
		}
		if len(displayPath) > maxPathLen {
			maxPathLen = len(displayPath)
		}
	}

	for _, wt := range worktrees {
		// Replace home directory with ~
		displayPath := wt.Path
		if homeDir != "" && strings.HasPrefix(wt.Path, homeDir) {
			displayPath = "~" + strings.TrimPrefix(wt.Path, homeDir)
		}

		// Pad the display path manually before styling to preserve alignment
		paddedPath := fmt.Sprintf("%-*s", maxPathLen, displayPath)

		dot := dotStyle.Render("·")
		if wt.Path == current.Path {
			fmt.Printf("  %s %s %s\n",
				currentPathStyle.Render(paddedPath),
				dot,
				currentBranchStyle.Render(wt.Branch))
		} else {
			fmt.Printf("  %s %s %s\n",
				pathStyle.Render(paddedPath),
				dot,
				branchStyle.Render(wt.Branch))
		}
	}

	return nil
}
