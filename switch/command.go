package switchcmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zshuzh/wt/internal/tui"
)

type Options struct{}

func (o Options) Run() error {
	m := model{
		Model: tui.Model{Loading: true},
	}

	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	finalState := finalModel.(model)

	// print worktree path which is then captured by shell function
	if finalState.Selected {
		fmt.Println(finalState.Worktrees[finalState.Cursor].Path)
	}

	return nil
}
