package remove

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type Options struct{}

func (o Options) Run() error {
	m := model{
		loading: true,
	}

	_, err := tea.NewProgram(m, tea.WithOutput(os.Stderr)).Run()
	return err
}
