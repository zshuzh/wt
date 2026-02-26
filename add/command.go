package add

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/zshuzh/wt/internal/git"
)

type Options struct{}

func (o Options) Run() error {
	root, err := git.GetRepoRoot()
	if err != nil {
		return err
	}
	path := root + "-"

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

	currentWorktree, err := git.GetCurrentWorktree()
	if err != nil {
		return err
	}

	if err := git.AddWorktree(git.Worktree{
		Path:   path,
		Branch: branchName,
	}); err != nil {
		return err
	}

	if git.IsGraphiteRepo() {
		if err := git.TrackWithGraphite(path, currentWorktree.Branch); err != nil {
			return err
		}
	}

	fmt.Println(path)
	return nil
}
