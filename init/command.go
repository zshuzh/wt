package init

import (
	"fmt"
	"strings"
)

type Options struct {
	Shell string `arg:"" help:"Shell type (zsh, bash)"`
}

func (o Options) Run() error {
	shellFunction := `
wt() {
  case "$1" in
    ai)
      local dir=$(command wt add)
      if [ -n "$dir" ]; then
        cd "$dir" && claude
      fi
      ;;
    checkout|co)
      local output=$(command wt "$@")
      local dir=$(echo "$output" | head -n1)
      local mode=$(echo "$output" | sed -n '2p')
      if [ -n "$dir" ]; then
        cd "$dir"
        if [ "$mode" = "graphite" ]; then
          gt co
        fi
      fi
      ;;
    add|a)
      local dir=$(command wt "$@")
      if [ -n "$dir" ]; then
        cd "$dir"
      fi
      ;;
    switch|s|review|r) # keep in sync with aliases in main.go
      local dir=$(command wt "$@")
      if [ -n "$dir" ]; then
        cd "$dir"
      fi
      ;;
    *)
      command wt "$@"
      ;;
  esac
}
`

	switch o.Shell {
	case "zsh", "bash":
		fmt.Println(strings.TrimSpace(shellFunction))
	default:
		return fmt.Errorf("unsupported shell: %s (supported: zsh, bash)", o.Shell)
	}
	return nil
}
