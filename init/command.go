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
  if [ "$1" = "switch" ]; then
    local dir=$(command wt switch)
    if [ -n "$dir" ]; then
      cd "$dir"
    fi
  else
    command wt "$@"
  fi
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
