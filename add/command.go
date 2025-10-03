package add

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/zshuzh/wt/internal/config"
	"github.com/zshuzh/wt/internal/git"
)

type Options struct{}

func (o Options) Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	root, err := git.GetRepoRoot()
	if err != nil {
		return err
	}

	var path string
	var branch string

	if cfg.PathDefault != "" {
		path = filepath.Join(root, cfg.PathDefault)
	} else {
		path = root
	}

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

	err = git.AddWorktree(git.Worktree{
		Path:   path,
		Branch: branch,
	})
	if err != nil {
		return err
	}

	for _, hook := range cfg.Hooks {
		if hook.Root != root {
			continue
		}

		workDir := filepath.Join(path, hook.Subdir)

		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}

		// The -l flag sources login files (e.g. ~/.zshrc) to set up a full environment. This also
		// allows wt to set env variables that hooks can use (e.g. $GIT_REPO_ROOT).
		cmd := exec.Command(shell, "-l", "-c", hook.Command)
		cmd.Dir = workDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			fmt.Printf("Hook failed: %v\n", err)
		}
	}

	return nil
}
