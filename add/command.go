package add

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/zshuzh/wt/internal/config"
	"github.com/zshuzh/wt/internal/git"
)

type Options struct{}

func filterBranches(query string, branches []string) []huh.Option[string] {
	var matches []huh.Option[string]

	if query == "" {
		// Show all branches when no filter is applied
		for _, branch := range branches {
			matches = append(matches, huh.NewOption(branch, branch))
		}
		return matches
	}

	queryLower := strings.ToLower(query)

	for _, branch := range branches {
		if strings.Contains(strings.ToLower(branch), queryLower) {
			matches = append(matches, huh.NewOption(branch, branch))
		}
	}

	return matches
}

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

	var branchName string
	var selectedBranch string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Path").
				Description("Enter the path for the new worktree").
				Value(&path),

			huh.NewInput().
				Title("Branch name").
				Description("Type to filter existing branches or create new").
				Value(&branchName),

			huh.NewSelect[string]().
				TitleFunc(func() string {
					matches := filterBranches(branchName, branches)
					if len(matches) == 0 && branchName != "" {
						return "Creating new branch: " + branchName
					}
					return "Or select existing branch:"
				}, &branchName).
				OptionsFunc(func() []huh.Option[string] {
					matches := filterBranches(branchName, branches)
					if len(matches) == 0 {
						// Return a dummy option that won't be selected
						return []huh.Option[string]{huh.NewOption("(press enter to create new branch)", "")}
					}
					return matches
				}, &branchName).
				Value(&selectedBranch),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	// Use selected branch if available, otherwise use typed branch name
	branch := selectedBranch
	if branch == "" {
		branch = branchName
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
