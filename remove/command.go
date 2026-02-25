package remove

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zshuzh/wt/internal/tui"
)

type Options struct{}

func (o Options) Run() error {
	m := model{
		Model: tui.Model{Loading: true},
	}

	_, err := tea.NewProgram(m, tea.WithOutput(os.Stderr)).Run()
	return err
}
