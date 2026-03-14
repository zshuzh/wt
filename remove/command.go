package remove

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zshuzh/wt/internal/tui"
)

type Options struct {
	Force bool `short:"f" help:"Force removal even if worktree is dirty"`
}

func (o Options) Run() error {
	m := model{
		Model: tui.Model{Loading: true},
		force: o.Force,
	}

	_, err := tea.NewProgram(m, tea.WithOutput(os.Stderr)).Run()
	return err
}
