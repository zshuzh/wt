package checkout

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/zshuzh/wt/internal/git"
)

type Options struct{}

func (o Options) Run() error {
	if git.IsGraphiteRepo() {
		return o.runGraphite()
	}
	return o.runDefault()
}

func (o Options) runGraphite() error {
	root, err := git.GetRepoRoot()
	if err != nil {
		return err
	}
	path := root + "-"

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

	// Create a temporary branch name based on the worktree directory name
	dirName := filepath.Base(path)
	tempBranch := fmt.Sprintf("wt-%s-%d", dirName, time.Now().Unix())

	if err := git.AddWorktree(path, tempBranch); err != nil {
		return err
	}

	// Run gt co interactively inside the new worktree
	if err := git.RunGraphiteInteractive("--cwd", path, "co"); err != nil {
		return err
	}

	fmt.Println(path)
	return nil
}

func filterBranches(query string, branches []string) []huh.Option[string] {
	var matches []huh.Option[string]

	if query == "" {
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

func (o Options) runDefault() error {
	root, err := git.GetRepoRoot()
	if err != nil {
		return err
	}
	path := root + "-"

	branches, err := git.GetBranches()
	if err != nil {
		return err
	}

	var filter string
	var selectedBranch string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Path").
				Description("Enter the path for the new worktree").
				Value(&path),

			huh.NewInput().
				Title("Filter").
				Description("Type to filter branches").
				Value(&filter),

			huh.NewSelect[string]().
				Title("Select branch").
				OptionsFunc(func() []huh.Option[string] {
					return filterBranches(filter, branches)
				}, &filter).
				Value(&selectedBranch),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if err := git.AddWorktree(path, selectedBranch); err != nil {
		return err
	}

	fmt.Println(path)
	return nil
}
