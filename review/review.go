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
	repo     string
	root     string
	prs      []pr
	filtered []pr
	cursor   int
	loading  bool
	err      error
	selected bool
	filter   textinput.Model
	spinner  spinner.Model

	// setup phase
	setting    bool
	steps      []step
	currentStep int
	done       bool
	resultPath string
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
		filter:  ti,
		spinner: s,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(fetchPRsCmd(m.repo), m.spinner.Tick, textinput.Blink)
}

func (m *model) applyFilter() {
	query := strings.ToLower(m.filter.Value())
	if query == "" {
		m.filtered = m.prs
		return
	}

	m.filtered = nil
	for _, p := range m.prs {
		searchable := strings.ToLower(fmt.Sprintf("#%d %s %s", p.Number, p.Title, p.Author.Login))
		if strings.Contains(searchable, query) {
			m.filtered = append(m.filtered, p)
		}
	}
}

func (m model) startSetup(selected pr) (model, tea.Cmd) {
	path := fmt.Sprintf("%s-review-%d", m.root, selected.Number)
	branch := selected.HeadRefName

	m.selected = true
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

	// Selection phase
	switch msg := msg.(type) {
	case prsMsg:
		m.prs = msg.prs
		m.filtered = msg.prs
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
				return m.startSetup(m.filtered[m.cursor])
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

	// Pass to text input for filtering
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	m.applyFilter()
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
	return m, cmd
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

	if len(m.prs) == 0 {
		return helpStyle.Render("No PRs requesting your review.") + "\n"
	}

	var s string

	s += m.filter.View() + "\n\n"

	if len(m.filtered) == 0 {
		s += helpStyle.Render("  No matching PRs.") + "\n"
		return s
	}

	for i, p := range m.filtered {
		isCurrent := m.cursor == i
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

		// Line 2: author, diff stats, time, labels
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

		s += line1 + "\n" + line2 + "\n"
		if i < len(m.filtered)-1 {
			s += "\n"
		}
	}

	return s
}
