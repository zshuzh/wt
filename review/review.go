package review

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zshuzh/wt/internal/git"
)

var renderer = lipgloss.NewRenderer(os.Stderr)

var (
	accentStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("212"))

	titleStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("255")).
			Bold(true)

	titleSelectedStyle = renderer.NewStyle().
				Foreground(lipgloss.Color("212")).
				Bold(true)

	numberStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("39"))

	numberSelectedStyle = renderer.NewStyle().
				Foreground(lipgloss.Color("212"))

	authorStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("178"))

	dimStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("243"))

	addStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("114"))

	delStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("204"))

	draftStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("243")).
			Italic(true)

	labelStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("213"))

	helpStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("241"))

	errorStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("196"))

	promptStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("212"))

	stackIconStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("212"))

	headerStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	positionStyle = renderer.NewStyle().
			Foreground(lipgloss.Color("243"))
)

// viewMode represents which screen the user is on.
type viewMode int

const (
	viewStacks viewMode = iota // top-level: list of stacks and standalone PRs
	viewDetail                 // drill-in: PRs within a single stack
	viewSetup                  // checking out a PR (spinner)
)

type prsMsg struct {
	prs []pr
}

type errMsg error

func fetchPRsCmd(repo string) tea.Cmd {
	return func() tea.Msg {
		prs, err := fetchReviewPRs(repo)
		if err != nil {
			return errMsg(err)
		}
		return prsMsg{prs: prs}
	}
}

type model struct {
	repo string
	root string

	// data
	stacks   []stack
	filtered []stack // after applying filter

	// navigation
	mode   viewMode
	cursor int

	// detail view
	activeStack *stack
	detailPRs   []pr
	detailIdx   int

	// filter
	filter textinput.Model

	// setup phase
	setting     bool
	steps       []step
	currentStep int
	done        bool
	resultPath  string

	// common
	loading  bool
	err      error
	selected bool
	spinner  spinner.Model
}

func newModel(repo, root string) model {
	ti := textinput.New()
	ti.Placeholder = "type to filter..."
	ti.Prompt = "/ "
	ti.PromptStyle = promptStyle
	ti.Width = 40
	ti.Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = accentStyle

	return model{
		repo:    repo,
		root:    root,
		loading: true,
		mode:    viewStacks,
		filter:  ti,
		spinner: s,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(fetchPRsCmd(m.repo), m.spinner.Tick, textinput.Blink)
}

// applyFilter filters stacks based on the current query.
func (m *model) applyFilter() {
	query := strings.ToLower(m.filter.Value())
	if query == "" {
		m.filtered = m.stacks
		return
	}

	m.filtered = nil
	for _, s := range m.stacks {
		// Match against stack title, author, or any PR title/number.
		searchable := strings.ToLower(s.Title)
		if len(s.PRs) > 0 {
			searchable += " " + strings.ToLower(s.PRs[0].Author.Login)
		}
		for _, p := range s.PRs {
			searchable += " " + strings.ToLower(fmt.Sprintf("#%d %s", p.Number, p.Title))
		}
		if strings.Contains(searchable, query) {
			m.filtered = append(m.filtered, s)
		}
	}
}

func (m model) startSetup(selected pr) (model, tea.Cmd) {
	path := fmt.Sprintf("%s-review-%d", m.root, selected.Number)
	branch := selected.HeadRefName

	m.selected = true
	m.mode = viewSetup
	m.setting = true
	m.resultPath = path
	m.steps = []step{
		{message: fmt.Sprintf("Fetching %s", branch), run: func() error {
			return git.FetchBranch(branch)
		}},
		{message: "Creating worktree", run: func() error {
			return git.AddWorktree(path, branch, "origin/"+branch)
		}},
	}
	m.currentStep = 0

	return m, tea.Batch(m.spinner.Tick, runStep(m.steps[0]))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Setup phase
	if m.setting {
		switch msg := msg.(type) {
		case stepDoneMsg:
			if msg.err != nil {
				m.err = msg.err
				return m, tea.Quit
			}
			m.currentStep++
			if m.currentStep >= len(m.steps) {
				m.done = true
				return m, tea.Quit
			}
			return m, runStep(m.steps[m.currentStep])

		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd

		case tea.KeyMsg:
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}
		return m, nil
	}

	// Handle shared messages
	switch msg := msg.(type) {
	case prsMsg:
		m.stacks = groupIntoStacks(msg.prs)
		m.filtered = m.stacks
		m.loading = false
		return m, nil

	case errMsg:
		m.err = msg
		m.loading = false
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	// Delegate to the active view
	switch m.mode {
	case viewStacks:
		return m.updateStacks(msg)
	case viewDetail:
		return m.updateDetail(msg)
	}

	return m, nil
}

// updateStacks handles input for the top-level stack list.
func (m model) updateStacks(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "ctrl+n":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
			return m, nil

		case "enter":
			if len(m.filtered) > 0 {
				s := m.filtered[m.cursor]
				if s.IsStack() {
					// Drill into the stack.
					m.mode = viewDetail
					m.activeStack = &s
					m.detailPRs = s.PRs
					m.detailIdx = 0
					return m, nil
				}
				// Standalone PR — check it out directly.
				return m.startSetup(s.PRs[0])
			}
			return m, nil

		case "esc":
			if m.filter.Value() != "" {
				m.filter.SetValue("")
				m.applyFilter()
				m.cursor = 0
				return m, nil
			}
			return m, tea.Quit
		}
	}

	// Pass to text input for filtering.
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	m.applyFilter()
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
	return m, cmd
}

// updateDetail handles input for the stack detail view.
func (m model) updateDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "up", "ctrl+p":
			if m.detailIdx > 0 {
				m.detailIdx--
			}
			return m, nil

		case "down", "ctrl+n":
			if m.detailIdx < len(m.detailPRs)-1 {
				m.detailIdx++
			}
			return m, nil

		case "enter":
			if len(m.detailPRs) > 0 {
				return m.startSetup(m.detailPRs[m.detailIdx])
			}
			return m, nil

		case "esc":
			// Go back to the stack list.
			m.mode = viewStacks
			m.activeStack = nil
			m.detailPRs = nil
			m.detailIdx = 0
			return m, nil
		}
	}

	return m, nil
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", h)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", days)
	}
}

func (m model) View() string {
	if m.setting {
		if m.done || m.err != nil {
			return ""
		}
		return m.spinner.View() + " " + m.steps[m.currentStep].message + "\n"
	}

	if m.loading {
		return m.spinner.View() + " Loading PRs...\n"
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
	}

	switch m.mode {
	case viewStacks:
		return m.viewStacks()
	case viewDetail:
		return m.viewDetail()
	}

	return ""
}

// viewStacks renders the top-level list of stacks and standalone PRs.
func (m model) viewStacks() string {
	if len(m.stacks) == 0 {
		return helpStyle.Render("No PRs requesting your review.") + "\n"
	}

	var s string
	s += m.filter.View() + "\n\n"

	if len(m.filtered) == 0 {
		s += helpStyle.Render("  No matching PRs.") + "\n"
		return s
	}

	for i, st := range m.filtered {
		isCurrent := m.cursor == i

		if st.IsStack() {
			s += m.renderStackRow(st, isCurrent)
		} else {
			s += m.renderPRRow(st.PRs[0], isCurrent)
		}

		if i < len(m.filtered)-1 {
			s += "\n"
		}
	}

	s += "\n" + helpStyle.Render("  ↑↓ navigate · enter select · esc quit")
	return s
}

// viewDetail renders the PRs within the active stack.
func (m model) viewDetail() string {
	var s string

	s += "  " + headerStyle.Render("◀ "+m.activeStack.Title) +
		dimStyle.Render(fmt.Sprintf("  (%d PRs)", len(m.detailPRs))) + "\n\n"

	for i, p := range m.detailPRs {
		isCurrent := m.detailIdx == i
		cursor := "  "
		if isCurrent {
			cursor = accentStyle.Render("> ")
		}

		// Line 1: number + title (+ draft badge)
		var num, title string
		if isCurrent {
			num = numberSelectedStyle.Render(fmt.Sprintf("#%d", p.Number))
			title = titleSelectedStyle.Render(p.Title)
		} else {
			num = numberStyle.Render(fmt.Sprintf("#%d", p.Number))
			title = titleStyle.Render(p.Title)
		}
		line1 := fmt.Sprintf("%s%s %s", cursor, num, title)
		if p.IsDraft {
			line1 += " " + draftStyle.Render("[draft]")
		}

		// Line 2: author, diff stats, time, position
		var parts []string
		parts = append(parts, authorStyle.Render("@"+p.Author.Login))
		parts = append(parts, addStyle.Render(fmt.Sprintf("+%d", p.Additions))+delStyle.Render(fmt.Sprintf(" -%d", p.Deletions)))
		parts = append(parts, dimStyle.Render(timeAgo(p.UpdatedAt)))
		parts = append(parts, positionStyle.Render(fmt.Sprintf("%d of %d", i+1, len(m.detailPRs))))
		if len(p.Labels) > 0 {
			var names []string
			for _, l := range p.Labels {
				names = append(names, l.Name)
			}
			parts = append(parts, labelStyle.Render(strings.Join(names, ", ")))
		}

		line2 := "     " + strings.Join(parts, dimStyle.Render(" · "))

		s += line1 + "\n" + line2 + "\n"
		if i < len(m.detailPRs)-1 {
			s += "\n"
		}
	}

	s += "\n" + helpStyle.Render("  ↑↓ navigate · enter checkout · esc back")
	return s
}

// renderStackRow renders a collapsed stack entry in the top-level view.
func (m model) renderStackRow(s stack, isCurrent bool) string {
	cursor := "  "
	if isCurrent {
		cursor = accentStyle.Render("> ")
	}

	icon := stackIconStyle.Render("▶")
	var title string
	if isCurrent {
		title = titleSelectedStyle.Render(s.Title)
	} else {
		title = titleStyle.Render(s.Title)
	}
	count := dimStyle.Render(fmt.Sprintf("(%d PRs)", len(s.PRs)))
	line1 := fmt.Sprintf("%s%s %s %s", cursor, icon, title, count)

	// Line 2: author, aggregate stats, time
	var parts []string
	if len(s.PRs) > 0 {
		parts = append(parts, authorStyle.Render("@"+s.PRs[0].Author.Login))
	}
	parts = append(parts, addStyle.Render(fmt.Sprintf("+%d", s.TotalAdditions()))+delStyle.Render(fmt.Sprintf(" -%d", s.TotalDeletions())))
	parts = append(parts, dimStyle.Render(timeAgo(s.LatestUpdate())))

	line2 := "     " + strings.Join(parts, dimStyle.Render(" · "))

	return line1 + "\n" + line2 + "\n"
}

// renderPRRow renders a standalone PR entry in the top-level view.
func (m model) renderPRRow(p pr, isCurrent bool) string {
	cursor := "  "
	if isCurrent {
		cursor = accentStyle.Render("> ")
	}

	var num, title string
	if isCurrent {
		num = numberSelectedStyle.Render(fmt.Sprintf("#%d", p.Number))
		title = titleSelectedStyle.Render(p.Title)
	} else {
		num = numberStyle.Render(fmt.Sprintf("#%d", p.Number))
		title = titleStyle.Render(p.Title)
	}
	line1 := fmt.Sprintf("%s%s %s", cursor, num, title)
	if p.IsDraft {
		line1 += " " + draftStyle.Render("[draft]")
	}

	var parts []string
	parts = append(parts, authorStyle.Render("@"+p.Author.Login))
	parts = append(parts, addStyle.Render(fmt.Sprintf("+%d", p.Additions))+delStyle.Render(fmt.Sprintf(" -%d", p.Deletions)))
	parts = append(parts, dimStyle.Render(timeAgo(p.UpdatedAt)))
	if len(p.Labels) > 0 {
		var names []string
		for _, l := range p.Labels {
			names = append(names, l.Name)
		}
		parts = append(parts, labelStyle.Render(strings.Join(names, ", ")))
	}

	line2 := "     " + strings.Join(parts, dimStyle.Render(" · "))

	return line1 + "\n" + line2 + "\n"
}
