package scripts

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
)

// Dir is the folder inside a repo root where init scripts live.
const Dir = ".wt/setup"

// Find discovers all .sh script names (without the .sh extension) under
// <repoRoot>/.wt/. Returns nil (no error) if the directory does not exist.
func Find(repoRoot string) ([]string, error) {
	dir := filepath.Join(repoRoot, Dir)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", dir, err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".sh") {
			names = append(names, strings.TrimSuffix(e.Name(), ".sh"))
		}
	}
	sort.Strings(names)
	return names, nil
}

// Run executes the named scripts (without the .sh suffix) from
// <repoRoot>/.wt/ inside wtPath. Scripts are run sequentially; execution
// stops at the first failure.
//
// Each script receives the following environment variables:
//
//	WT_PATH      – absolute path to the worktree being initialised
//	WT_BRANCH    – branch checked out in the worktree (empty for detached HEAD)
//	WT_MAIN_REPO – absolute path to the main (bare) repo root
func Run(repoRoot, wtPath, branch string, names []string) error {
	for _, name := range names {
		scriptPath := filepath.Join(repoRoot, Dir, name+".sh")
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			return fmt.Errorf("script not found: %s.sh", name)
		}

		fmt.Fprintf(os.Stderr, "▶ running %s.sh\n", name)

		cmd := exec.Command("sh", scriptPath)
		cmd.Dir = wtPath
		// Always write to stderr so stdout stays clean for path output used
		// by the shell wrapper (e.g. dir=$(command wt add)).
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		cmd.Env = append(os.Environ(),
			"WT_PATH="+wtPath,
			"WT_BRANCH="+branch,
			"WT_MAIN_REPO="+repoRoot,
		)

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("script %s.sh failed: %w", name, err)
		}
	}
	return nil
}

// SelectAndRun prompts the user to choose which scripts to run when names is
// empty; otherwise it runs the named scripts directly. It is a no-op (with a
// notice) when no scripts exist in .wt/.
func SelectAndRun(repoRoot, wtPath, branch string, names []string) error {
	available, err := Find(repoRoot)
	if err != nil {
		return err
	}

	if len(available) == 0 {
		fmt.Fprintln(os.Stderr, "no init scripts found in .wt/ — skipping")
		return nil
	}

	toRun := names

	if len(toRun) == 0 {
		opts := make([]huh.Option[string], len(available))
		for i, s := range available {
			opts[i] = huh.NewOption(s, s)
		}

		var selected []string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Select init scripts to run").
					Description("Space to toggle · enter to confirm").
					Options(opts...).
					Value(&selected),
			),
		)
		if err := form.Run(); err != nil {
			return err
		}
		toRun = selected
	}

	if len(toRun) == 0 {
		return nil
	}

	// Validate all names up-front before running anything.
	scriptSet := make(map[string]bool, len(available))
	for _, s := range available {
		scriptSet[s] = true
	}
	for _, name := range toRun {
		if !scriptSet[name] {
			return fmt.Errorf("script not found: %s.sh", name)
		}
	}

	return Run(repoRoot, wtPath, branch, toRun)
}
