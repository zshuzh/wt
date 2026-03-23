package review

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
)

type pr struct {
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Author      author    `json:"author"`
	HeadRefName string    `json:"headRefName"`
	BaseRefName string    `json:"baseRefName"`
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

// stack represents a group of related PRs that form a Graphite-style stack,
// or a single standalone PR.
type stack struct {
	Title string // derived from the root PR's title
	PRs   []pr   // ordered from bottom (root) to top (tip)
}

// TotalAdditions returns the sum of additions across all PRs in the stack.
func (s stack) TotalAdditions() int {
	n := 0
	for _, p := range s.PRs {
		n += p.Additions
	}
	return n
}

// TotalDeletions returns the sum of deletions across all PRs in the stack.
func (s stack) TotalDeletions() int {
	n := 0
	for _, p := range s.PRs {
		n += p.Deletions
	}
	return n
}

// LatestUpdate returns the most recent UpdatedAt across all PRs.
func (s stack) LatestUpdate() time.Time {
	var latest time.Time
	for _, p := range s.PRs {
		if p.UpdatedAt.After(latest) {
			latest = p.UpdatedAt
		}
	}
	return latest
}

// IsStack returns true if this contains more than one PR.
func (s stack) IsStack() bool {
	return len(s.PRs) > 1
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
		"--json", "number,title,author,headRefName,baseRefName,additions,deletions,isDraft,updatedAt,labels",
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

// groupIntoStacks takes a flat list of PRs and groups them into stacks.
//
// A stack is detected by examining baseRefName/headRefName relationships:
// if PR B's baseRefName equals PR A's headRefName, they are in the same stack.
// PRs whose baseRefName is a trunk branch (e.g. main, master) and whose
// headRefName is not referenced by any other PR's baseRefName are standalone.
func groupIntoStacks(prs []pr) []stack {
	if len(prs) == 0 {
		return nil
	}

	// Index PRs by their head branch name for lookup.
	byHead := make(map[string]*pr, len(prs))
	for i := range prs {
		byHead[prs[i].HeadRefName] = &prs[i]
	}

	// For each PR, find the root of its chain (the PR whose base is not
	// another PR's head within our set). Group by root.
	rootOf := make(map[int]int, len(prs))    // pr number -> root pr number
	visited := make(map[int]bool, len(prs))

	var findRoot func(p *pr) int
	findRoot = func(p *pr) int {
		if r, ok := rootOf[p.Number]; ok {
			return r
		}
		if visited[p.Number] {
			// Cycle guard — treat as its own root.
			rootOf[p.Number] = p.Number
			return p.Number
		}
		visited[p.Number] = true

		parent, exists := byHead[p.BaseRefName]
		if !exists {
			// Base branch is not another PR in our set → this is the root.
			rootOf[p.Number] = p.Number
			return p.Number
		}
		root := findRoot(parent)
		rootOf[p.Number] = root
		return root
	}

	for i := range prs {
		findRoot(&prs[i])
	}

	// Collect PRs by root.
	groups := make(map[int][]pr)
	for i := range prs {
		root := rootOf[prs[i].Number]
		groups[root] = append(groups[root], prs[i])
	}

	// Order PRs within each group by their dependency chain (tip first).
	var stacks []stack
	for _, group := range groups {
		ordered := orderChain(group, byHead)
		// Reverse so the tip (newest) is first and the root is last.
		for i, j := 0, len(ordered)-1; i < j; i, j = i+1, j-1 {
			ordered[i], ordered[j] = ordered[j], ordered[i]
		}
		s := stack{
			Title: ordered[0].Title,
			PRs:   ordered,
		}
		stacks = append(stacks, s)
	}

	// Sort stacks by most recently updated first.
	sort.Slice(stacks, func(i, j int) bool {
		return stacks[i].LatestUpdate().After(stacks[j].LatestUpdate())
	})

	return stacks
}

// orderChain sorts a group of PRs so that the root (whose base is not in the
// group) comes first, followed by each subsequent child.
func orderChain(group []pr, byHead map[string]*pr) []pr {
	if len(group) <= 1 {
		return group
	}

	// Build a map from base branch → PR for this group.
	headSet := make(map[string]bool, len(group))
	for _, p := range group {
		headSet[p.HeadRefName] = true
	}

	// Find the root: the PR whose base is not another PR's head in this group.
	var root *pr
	childOf := make(map[string]pr) // parent head -> child PR
	for _, p := range group {
		if !headSet[p.BaseRefName] {
			root = &p
		} else {
			childOf[p.BaseRefName] = p
		}
	}

	if root == nil {
		// Fallback: just return sorted by number.
		sort.Slice(group, func(i, j int) bool {
			return group[i].Number < group[j].Number
		})
		return group
	}

	ordered := []pr{*root}
	current := root.HeadRefName
	for {
		next, ok := childOf[current]
		if !ok {
			break
		}
		ordered = append(ordered, next)
		current = next.HeadRefName
	}

	// If we missed any (e.g. branching stacks), append them.
	seen := make(map[int]bool, len(ordered))
	for _, p := range ordered {
		seen[p.Number] = true
	}
	for _, p := range group {
		if !seen[p.Number] {
			ordered = append(ordered, p)
		}
	}

	return ordered
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
