package add

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/zshuzh/wt/internal/git"
)

type Options struct {
	Path    string `arg:"" optional:"" help:"Path for the new worktree"`
	Message string `arg:"" optional:"" help:"Commit message (creates a branch via gt c -m)"`
}

func (o Options) Run() error {
	path := o.Path
	message := o.Message

	root, err := git.GetRepoRoot()
	if err != nil {
		return err
	}

	if path == "" {
		path = root + "-"

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Path").
					Description("Enter the path for the new worktree").
					Value(&path),
				huh.NewInput().
					Title("Commit message").
					Description("Leave empty for detached HEAD").
					Value(&message),
			),
		)
		if err := form.Run(); err != nil {
			return err
		}
	} else {
		path = root + "-" + path
	}

	if message != "" {
		// Get source branch before gt c moves us off it
		source, err := git.GetCurrentBranch()
		if err != nil {
			return err
		}

		// Create stacked branch via Graphite (auto-names from message)
		if err := git.GraphiteCreate(message); err != nil {
			return err
		}

		// Get the branch name Graphite created
		branch, err := git.GetCurrentBranch()
		if err != nil {
			return err
		}

		// Go back to source so the new branch is free for the worktree
		if err := git.Checkout(source); err != nil {
			return err
		}

		// Create worktree on the Graphite-managed branch
		if err := git.AddWorktree(path, branch); err != nil {
			return err
		}
	} else {
		if err := git.AddWorktreeDetached(path); err != nil {
			return err
		}
	}

	fmt.Println(path)
	return nil
}
