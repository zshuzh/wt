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
    switch|add|checkout)
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
