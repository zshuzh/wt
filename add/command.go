package add

import (
	"os"

	"github.com/charmbracelet/huh"
	"github.com/zshuzh/wt/internal/git"
)

type Options struct{}

func (o Options) Run() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	path := cwd + "-"

	var branchName string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Path").
				Description("Enter the path for the new worktree").
				Value(&path),

			huh.NewInput().
				Title("Branch name").
				Description("Name for the new branch").
				Value(&branchName),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	return git.AddWorktree(git.Worktree{
		Path:   path,
		Branch: branchName,
	})
}
