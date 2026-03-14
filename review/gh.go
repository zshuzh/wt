package review

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type pr struct {
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Author      author    `json:"author"`
	HeadRefName string    `json:"headRefName"`
	Additions   int       `json:"additions"`
	Deletions   int       `json:"deletions"`
	IsDraft     bool      `json:"isDraft"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Labels      []label   `json:"labels"`
}

type author struct {
	Login string `json:"login"`
}

type label struct {
	Name string `json:"name"`
}

func runGhOutput(args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
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

func fetchReviewPRs(repo string) ([]pr, error) {
	out, err := runGhOutput("pr", "list",
		"--repo", repo,
		"--search", "review-requested:@me",
		"--json", "number,title,author,headRefName,additions,deletions,isDraft,updatedAt,labels",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PRs: %w", err)
	}

	var prs []pr
	if err := json.Unmarshal([]byte(out), &prs); err != nil {
		return nil, fmt.Errorf("failed to parse PR list: %w", err)
	}

	return prs, nil
}

func fetchPRBranch(repo string, number int) (string, error) {
	out, err := runGhOutput("pr", "view",
		fmt.Sprintf("%d", number),
		"--repo", repo,
		"--json", "headRefName",
		"-q", ".headRefName",
	)
	if err != nil {
		return "", fmt.Errorf("failed to fetch PR #%d: %w", number, err)
	}

	branch := strings.TrimSpace(out)
	if branch == "" {
		return "", fmt.Errorf("PR #%d has no branch", number)
	}

	return branch, nil
}
