# wt

An interactive worktree manager. Adds a nice UI and several conveniences on top of the native `git worktree` commands.

## Installation

Note that no matter how you install you'll have to add the following line to your `.zshrc`:

```sh
eval "$(wt init zsh)"
```

This enables auto-`cd` when you switch or create worktrees, and unlocks the `wt ai` shortcut.

### MacOS

```sh
brew install zshuzh/tap/wt
```

### Source

```sh
go install github.com/zshuzh/wt@latest
```

## Commands

| Command       | Shorthand | What it does                             |
| ------------- | --------- | ---------------------------------------- |
| `wt list`     | `l`       | List all worktrees                       |
| `wt switch`   | `s`       | Interactive picker to switch worktrees   |
| `wt add`      | `a`       | Create a new worktree with a new branch  |
| `wt checkout` | `co`      | Create a worktree for an existing branch |
| `wt remove`   | `rm`      | Remove a worktree                        |
| `wt nuke`     | `n`       | Remove a worktree and delete its branch  |
| `wt ai`       |           | Create a worktree and start an AI agent  |
