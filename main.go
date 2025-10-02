package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/zshuzh/wt/add"
	"github.com/zshuzh/wt/configure"
	"github.com/zshuzh/wt/list"
	"github.com/zshuzh/wt/remove"
	switchcmd "github.com/zshuzh/wt/switch"
)

type Wt struct {
	Version   kong.VersionFlag  `short:"v" help:"Print version information"`
	Configure configure.Options `cmd:"" help:"Start the wt setup wizard"`
	List      list.Options      `cmd:"" help:"List current worktrees"`
	Switch    switchcmd.Options `cmd:"" help:"Switch to a worktree"`
	Add       add.Options       `cmd:"" help:"Add a new worktree"`
	Remove    remove.Options    `cmd:"" help:"Remove an existing worktree"`
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
		fmt.Printf("Error: %v\n", err)
	}
}
