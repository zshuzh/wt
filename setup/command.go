package setup

import (
	"github.com/zshuzh/wt/internal/git"
	"github.com/zshuzh/wt/internal/scripts"
)

// Options represents the `wt setup` command.
type Options struct {
	Scripts []string `arg:"" optional:"" help:"Names of scripts to run (without .sh). Omit to get an interactive picker."`
}

func (o Options) Run() error {
	root, err := git.GetRepoRoot()
	if err != nil {
		return err
	}

	wt, err := git.GetCurrentWorktree()
	if err != nil {
		return err
	}

	return scripts.SelectAndRun(root, wt.Path, wt.Branch, o.Scripts)
}
