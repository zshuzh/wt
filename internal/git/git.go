package git

import (
	"bytes"
	"fmt"
	"os"
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

func runGraphite(args ...string) error {
	cmd := exec.Command("gt", args...)
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

func AddWorktree(path, branch string, ref ...string) error {
	err := runGit("rev-parse", "--verify", branch)
	branchExists := err == nil

	if branchExists {
		return runGit("worktree", "add", path, branch)
	}
	startPoint := "HEAD"
	if len(ref) > 0 && ref[0] != "" {
		startPoint = ref[0]
	}
	return runGit("worktree", "add", "-b", branch, path, startPoint)
}

func RemoveWorktree(path string, force bool) error {
	if force {
		return runGit("worktree", "remove", "--force", path)
	}
	return runGit("worktree", "remove", path)
}

func DeleteBranch(branch string) error {
	if IsGraphiteRepo() {
		if err := runGraphite("delete", branch, "--force"); err == nil {
			return nil
		}
	}
	return runGit("branch", "-D", branch)
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

func FetchBranch(branch string) error {
	return runGit("fetch", "origin", branch)
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

func GetRepoSlug() (string, error) {
	output, err := runGitOutput("remote", "get-url", "origin")
	if err != nil {
		return "", err
	}

	url := strings.TrimSpace(output)

	// Handle SSH: git@github.com:owner/repo.git
	if strings.HasPrefix(url, "git@") {
		url = strings.TrimPrefix(url, "git@github.com:")
		url = strings.TrimSuffix(url, ".git")
		return url, nil
	}

	// Handle HTTPS: https://github.com/owner/repo.git
	url = strings.TrimPrefix(url, "https://github.com/")
	url = strings.TrimPrefix(url, "http://github.com/")
	url = strings.TrimSuffix(url, ".git")
	return url, nil
}

func IsGraphiteRepo() bool {
	root, err := GetRepoRoot()
	if err != nil {
		return false
	}
	_, err = os.Stat(root + "/.git/.graphite_repo_config")
	return err == nil
}

func RunGraphiteInteractive(args ...string) error {
	cmd := exec.Command("gt", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stderr // TUI output goes to stderr so stdout stays clean for path capture
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func TrackWithGraphite(cwd, parent string) error {
	return runGraphite("--cwd", cwd, "track", "--parent", parent)
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
