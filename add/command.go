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
	root, err := git.GetRepoRoot()
	if err != nil {
		return err
	}

	path := o.Path
	if path == "" {
		path = root + "-"
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Path").
					Description("Enter the path for the new worktree").
					Value(&path),
			),
		)
		if err := form.Run(); err != nil {
			return err
		}
	}

	if err := git.AddWorktreeDetached(path); err != nil {
		return err
	}

	fmt.Println(path)
	if o.Message != "" {
		fmt.Println(o.Message)
	}
	return nil
}
