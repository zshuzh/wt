package git

import (
	"os/exec"
	"strings"
)

type Worktree struct {
	Path   string
	Branch string
}

func GetWorktrees() ([]Worktree, error) {
	output, err := exec.Command("git", "worktree", "list", "--porcelain").Output()
	if err != nil {
		return nil, err
	}

	var worktrees []Worktree
	var currentPath string
	var currentBranch string

	for line := range strings.SplitSeq(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" { // blank line separates worktrees
			if currentPath != "" {
				worktrees = append(worktrees, Worktree{
					Path:   currentPath,
					Branch: currentBranch,
				})
				currentPath = ""
				currentBranch = ""
			}
			continue
		}

		if path, ok := strings.CutPrefix(line, "worktree "); ok {
			currentPath = path
		} else if branch, ok := strings.CutPrefix(line, "branch "); ok {
			currentBranch = strings.TrimPrefix(branch, "refs/heads/")
		}
	}

	return worktrees, nil
}

func AddWorktree(wt Worktree) error {
	return exec.Command("git", "worktree", "add", wt.Path, wt.Branch).Run()
}

func GetBranches() ([]string, error) {
	output, err := exec.Command("git", "branch", "--format=%(refname:short)").Output()
	if err != nil {
		return nil, err
	}

	var branches []string
	for line := range strings.SplitSeq(strings.TrimSpace(string(output)), "\n") {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "graphite-base/") {
			continue
		}

		branches = append(branches, line)
	}

	return branches, nil
}
