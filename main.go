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
	"github.com/zshuzh/wt/remove"
	switchcmd "github.com/zshuzh/wt/switch"
)

type Wt struct {
	Version  kong.VersionFlag   `short:"v" help:"Print version information"`
	Init     initcmd.Options   `cmd:"" help:"Print shell integration code"`
	List     list.Options      `cmd:"" help:"List current worktrees"`
	Switch   switchcmd.Options `cmd:"" help:"Switch to a worktree"`
	Add      add.Options       `cmd:"" help:"Add a new worktree with a new branch"`
	Checkout checkout.Options  `cmd:"" help:"Add a new worktree for an existing branch"`
	Remove   remove.Options    `cmd:"" help:"Remove an existing worktree"`
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
