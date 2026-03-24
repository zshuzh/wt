package checkout

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/zshuzh/wt/internal/git"
	"github.com/zshuzh/wt/internal/scripts"
)

type Options struct {
	Setup bool `help:"Run init scripts from .wt/ after creating the worktree"`
}

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

	if err := git.AddWorktreeDetached(path); err != nil {
		return err
	}

	fmt.Println(path)
	fmt.Println("graphite")

	if o.Setup {
		if err := scripts.SelectAndRun(root, path, "", nil); err != nil {
			return err
		}
	}

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

	if o.Setup {
		if err := scripts.SelectAndRun(root, path, selectedBranch, nil); err != nil {
			return err
		}
	}

	return nil
}
