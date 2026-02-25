package switchcmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type Options struct{}

func (o Options) Run() error {
	m := model{
		loading: true,
	}

	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	finalState := finalModel.(model)

	// print worktree path which is then captured by shell function
	if finalState.selected {
		fmt.Println(finalState.worktrees[finalState.cursor].Path)
	}

	return nil
}
