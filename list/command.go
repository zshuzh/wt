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

	currentDotStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EE6FF8"))

	currentPathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EE6FF8")).
		Bold(true)

	currentBranchStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EE6FF8"))

	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D7D7D"))

	branchStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D7D7D"))

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

		prefix := "  "
		if wt.Path == current.Path {
			prefix = currentDotStyle.Render("● ")
			fmt.Printf("%s%s  %s\n",
				prefix,
				currentPathStyle.Render(paddedPath),
				currentBranchStyle.Render(wt.Branch))
		} else {
			fmt.Printf("%s%s  %s\n",
				prefix,
				pathStyle.Render(paddedPath),
				branchStyle.Render(wt.Branch))
		}
	}

	return nil
}
