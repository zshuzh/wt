package add

import (
	"fmt"

	"github.com/zshuzh/wt/internal/git"
)

type Options struct {
	Path    string `arg:"" required:"" help:"Path for the new worktree"`
	Message string `arg:"" optional:"" help:"Commit message (creates a branch via gt c -m)"`
}

func (o Options) Run() error {
	if err := git.AddWorktreeDetached(o.Path); err != nil {
		return err
	}

	fmt.Println(o.Path)
	if o.Message != "" {
		fmt.Println(o.Message)
	}
	return nil
}
