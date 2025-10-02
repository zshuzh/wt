package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/zshuzh/wt/configure"
	"github.com/zshuzh/wt/list"
)

type Wt struct {
	Version kong.VersionFlag  `short:"v" help:"Print version information"`
	Init    configure.Options `cmd:"" help:"Start the wt setup wizard"`
	List    list.Options      `cmd:"" help:"List current worktrees"`
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
