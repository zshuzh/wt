package checkout

import (
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/zshuzh/wt/internal/git"
)

type Options struct{}

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

func (o Options) Run() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	path := cwd + "-"

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

	return git.AddWorktree(git.Worktree{
		Path:   path,
		Branch: selectedBranch,
	})
}
