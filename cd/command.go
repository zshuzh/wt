package cd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/zshuzh/wt/internal/git"
)

const aliasesFile = ".wt/aliases.json"

// Options represents the `wt cd` command.
type Options struct {
	Alias string `arg:"" optional:"" help:"Alias to navigate to. Omit for an interactive picker."`
}

type alias struct {
	Alias string `json:"alias"`
	Label string `json:"label"`
	Path  string `json:"path"`
}

func loadAliases(repoRoot string) ([]alias, error) {
	p := filepath.Join(repoRoot, aliasesFile)
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", aliasesFile, err)
	}

	var aliases []alias
	if err := json.Unmarshal(data, &aliases); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", aliasesFile, err)
	}
	return aliases, nil
}

func (o Options) Run() error {
	repoRoot, err := git.GetRepoRoot()
	if err != nil {
		return err
	}

	aliases, err := loadAliases(repoRoot)
	if err != nil {
		return err
	}

	if len(aliases) == 0 {
		fmt.Fprintf(os.Stderr, "no aliases defined — create %s with content like:\n  [{\"alias\": \"be\", \"label\": \"Backend\", \"path\": \"python/services/backend\"}]\n", aliasesFile)
		return nil
	}

	wt, err := git.GetCurrentWorktree()
	if err != nil {
		return err
	}

	// Direct lookup
	if o.Alias != "" {
		for _, a := range aliases {
			if a.Alias == o.Alias {
				fmt.Println(filepath.Join(wt.Path, a.Path))
				return nil
			}
		}
		return fmt.Errorf("unknown alias %q", o.Alias)
	}

	// Interactive picker — sort alphabetically by label (falling back to alias)
	sorted := make([]alias, len(aliases))
	copy(sorted, aliases)
	sort.Slice(sorted, func(i, j int) bool {
		li := sorted[i].Label
		if li == "" {
			li = sorted[i].Alias
		}
		lj := sorted[j].Label
		if lj == "" {
			lj = sorted[j].Alias
		}
		return li < lj
	})

	opts := make([]huh.Option[string], len(sorted))
	for i, a := range sorted {
		label := a.Label
		if label == "" {
			label = a.Alias
		}
		opts[i] = huh.NewOption(fmt.Sprintf("%-16s %s", label, a.Path), filepath.Join(wt.Path, a.Path))
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Navigate to").
				Options(opts...).
				Value(&selected),
		),
	).WithOutput(os.Stderr)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}

	if selected != "" {
		fmt.Println(selected)
	}
	return nil
}
