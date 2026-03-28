package remove

import (
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zshuzh/wt/internal/tui"
)

type Options struct {
	Force bool `short:"f" help:"Force removal even if worktree is dirty"`
}

func (o Options) Run() error {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	m := model{
		Model:   tui.Model{Loading: true},
		force:   o.Force,
		spinner: s,
	}

	_, err := tea.NewProgram(m, tea.WithOutput(os.Stderr)).Run()
	return err
}
