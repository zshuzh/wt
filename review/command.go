package review

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zshuzh/wt/internal/git"
)

type Options struct {
	Number int `arg:"" optional:"" help:"PR number to review"`
}

func (o Options) Run() error {
	repo, err := git.GetRepoSlug()
	if err != nil {
		return err
	}

	root, err := git.GetRepoRoot()
	if err != nil {
		return err
	}

	if o.Number > 0 {
		branch, err := fetchPRBranch(repo, o.Number)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("%s-review-%d", root, o.Number)

		steps := []step{
			{message: fmt.Sprintf("Fetching %s", branch), run: func() error {
				return git.FetchBranch(branch)
			}},
			{message: "Creating worktree", run: func() error {
				return git.AddWorktree(path, branch, "origin/"+branch)
			}},
		}

		s := spinner.New()
		s.Spinner = spinner.Dot
		s.Style = accentStyle
		sm := setupModel{steps: steps, spinner: s}
		p := tea.NewProgram(sm, tea.WithOutput(os.Stderr))
		finalModel, err := p.Run()
		if err != nil {
			return err
		}

		final := finalModel.(setupModel)
		if final.err != nil {
			return final.err
		}

		fmt.Println(path)
		return nil
	}

	m := newModel(repo, root)
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	final := finalModel.(model)
	if !final.selected {
		return nil
	}
	if final.err != nil {
		return final.err
	}

	fmt.Println(final.resultPath)
	return nil
}

type step struct {
	message string
	run     func() error
}

type stepDoneMsg struct{ err error }

func runStep(s step) tea.Cmd {
	return func() tea.Msg {
		return stepDoneMsg{err: s.run()}
	}
}

type setupModel struct {
	steps   []step
	current int
	spinner spinner.Model
	err     error
	done    bool
}

func (m setupModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, runStep(m.steps[0]))
}

func (m setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case stepDoneMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.current++
		if m.current >= len(m.steps) {
			m.done = true
			return m, tea.Quit
		}
		return m, runStep(m.steps[m.current])

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

func (m setupModel) View() string {
	if m.done || m.err != nil {
		return ""
	}

	return m.spinner.View() + " " + m.steps[m.current].message + "\n"
}
