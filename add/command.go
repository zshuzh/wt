package add

import (
	"github.com/charmbracelet/huh"
	"github.com/zshuzh/wt/internal/git"
)

type Options struct{}

func (o Options) Run() error {
	root, err := git.GetRepoRoot()
	if err != nil {
		return err
	}

	path := root
	var branch string

	branches, err := git.GetBranches()
	if err != nil {
		return err
	}

	options := make([]huh.Option[string], len(branches))
	for i, b := range branches {
		options[i] = huh.NewOption(b, b)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Path").
				Description("Enter the path for the new worktree").
				Value(&path),

			huh.NewSelect[string]().
				Title("Branch").
				Description("Select or search for a branch (press / to filter)").
				Options(options...).
				Value(&branch),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	return git.AddWorktree(git.Worktree{
		Path:   path,
		Branch: branch,
	})
}
