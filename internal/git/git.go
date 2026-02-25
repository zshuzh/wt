package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func runGit(args ...string) error {
	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return fmt.Errorf("%s", msg)
		}
		return err
	}
	return nil
}

func runGitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return "", fmt.Errorf("%s", msg)
		}
		return "", err
	}
	return string(out), nil
}

type Worktree struct {
	Path   string
	Branch string
}

func ListWorktrees() ([]Worktree, error) {
	output, err := runGitOutput("worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	var worktrees []Worktree
	var currentPath string
	var currentBranch string

	for line := range strings.SplitSeq(output, "\n") {
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

	if currentPath != "" {
		worktrees = append(worktrees, Worktree{Path: currentPath, Branch: currentBranch})
	}

	return worktrees, nil
}

func AddWorktree(wt Worktree) error {
	err := runGit("rev-parse", "--verify", wt.Branch)
	branchExists := err == nil

	if branchExists {
		return runGit("worktree", "add", wt.Path, wt.Branch)
	} else {
		return runGit("worktree", "add", "-b", wt.Branch, wt.Path)
	}
}

func RemoveWorktree(path string) error {
	return runGit("worktree", "remove", path)
}

func GetBranches() ([]string, error) {
	output, err := runGitOutput("branch", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}

	var branches []string
	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "graphite-base/") {
			continue
		}

		branches = append(branches, line)
	}

	return branches, nil
}

func GetRepoRoot() (string, error) {
	output, err := runGitOutput("worktree", "list", "--porcelain")
	if err != nil {
		return "", err
	}

	// first worktree in the list is the main worktree
	for line := range strings.SplitSeq(output, "\n") {
		if path, ok := strings.CutPrefix(line, "worktree "); ok {
			return path, nil
		}
	}

	return "", nil
}

func GetCurrentWorktree() (Worktree, error) {
	path, err := runGitOutput("rev-parse", "--show-toplevel")
	if err != nil {
		return Worktree{}, err
	}

	branch, err := runGitOutput("branch", "--show-current")
	if err != nil {
		return Worktree{}, err
	}

	return Worktree{
		Path:   strings.TrimSpace(path),
		Branch: strings.TrimSpace(branch),
	}, nil
}
