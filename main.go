package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/lipgloss"
	"github.com/zshuzh/wt/add"
	"github.com/zshuzh/wt/checkout"
	initcmd "github.com/zshuzh/wt/init"
	"github.com/zshuzh/wt/list"
	"github.com/zshuzh/wt/nuke"
	"github.com/zshuzh/wt/remove"
	switchcmd "github.com/zshuzh/wt/switch"
)

type Wt struct {
	Version  kong.VersionFlag   `short:"v" help:"Print version information"`
	Init     initcmd.Options   `cmd:"" help:"Print shell integration code"`
	List     list.Options      `cmd:"" aliases:"l" help:"List current worktrees"`
	Switch   switchcmd.Options `cmd:"" aliases:"s" help:"Switch to a worktree"`
	Add      add.Options       `cmd:"" aliases:"a" help:"Add a new worktree with a new branch"`
	Checkout checkout.Options  `cmd:"" aliases:"co" help:"Add a new worktree for an existing branch"`
	Remove   remove.Options    `cmd:"" aliases:"rm" help:"Remove an existing worktree"`
	Nuke     nuke.Options      `cmd:"" aliases:"n" help:"Remove a worktree and delete its branch"`
}

func init() {
	// Use stderr for color detection so that colors work when stdout is
	// captured by the shell wrapper (e.g. dir=$(command wt add)).
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(os.Stderr))
}

func main() {
	cli := &Wt{}
	ctx := kong.Parse(cli,
		kong.Name("wt"),
		kong.Description("Git worktree manager optimized for AI workflows"),
		kong.UsageOnError(),
		kong.Vars{"version": "0.1.0"},
	)

	err := ctx.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
