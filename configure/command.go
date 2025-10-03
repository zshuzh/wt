package configure

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/zshuzh/wt/internal/config"
	"github.com/zshuzh/wt/internal/git"
)

type Options struct{}

// extractFileFromCopyCommand extracts file path from cp command
func extractFileFromCopyCommand(command string) string {
	re := regexp.MustCompile(`cp \$WT_REPO_ROOT/(.+?) \$WT_WORKTREE_PATH/.+`)
	matches := re.FindStringSubmatch(command)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (o Options) Run() error {
	// Load existing config
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Get repo root
	root, err := git.GetRepoRoot()
	if err != nil {
		return err
	}

	// Separate hooks into copy and command hooks
	var copyHooks []config.Hook
	var commandHooks []config.Hook

	for _, hook := range cfg.Hooks {
		if hook.Root == root {
			if strings.Contains(hook.Command, "cp") {
				copyHooks = append(copyHooks, hook)
			} else {
				commandHooks = append(commandHooks, hook)
			}
		}
	}

	// Question 1: Path Default
	pathDefault := cfg.PathDefault
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Where do you want to create your worktrees?").
				Description("Relative path from repo root (e.g., '../worktrees-')").
				Value(&pathDefault),
		),
	).Run()
	if err != nil {
		return err
	}

	// Question 2: Copy Hooks
	for {
		var selected string
		options := []huh.Option[string]{
			huh.NewOption("✨ Create new copy hook", "__new__"),
		}

		for i, hook := range copyHooks {
			filePath := extractFileFromCopyCommand(hook.Command)
			if filePath == "" {
				filePath = hook.Command
			}
			options = append(options, huh.NewOption(filePath, fmt.Sprintf("__edit_%d__", i)))
		}

		options = append(options, huh.NewOption("✓ Done with copy hooks", "__done__"))

		err = huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Copy Hooks").
					Description("Select a hook to edit or create a new one").
					Options(options...).
					Value(&selected),
			),
		).Run()
		if err != nil {
			return err
		}

		if selected == "__done__" {
			break
		}

		var filePath string
		var editIndex int = -1

		if selected == "__new__" {
			// New hook
			filePath = ""
		} else {
			// Edit existing
			fmt.Sscanf(selected, "__edit_%d__", &editIndex)
			filePath = extractFileFromCopyCommand(copyHooks[editIndex].Command)
		}

		// Show file path input
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("File path").
					Description("Path relative to repo root (e.g., 'python/services/backend/.env')").
					Suggestions([]string{".env", ".env.local", ".env.development", ".env.production", "config/secrets.yml"}).
					Value(&filePath),
			),
		).Run()
		if err != nil {
			return err
		}

		if filePath == "" {
			continue
		}

		// Ask Edit or Delete
		var shouldDelete bool
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Delete this hook?").
					Description("No = Save, Yes = Delete").
					Affirmative("Delete").
					Negative("Save").
					Value(&shouldDelete),
			),
		).Run()
		if err != nil {
			return err
		}

		if shouldDelete {
			// Delete
			if editIndex >= 0 {
				copyHooks = append(copyHooks[:editIndex], copyHooks[editIndex+1:]...)
			}
		} else {
			// Save/Update
			newHook := config.Hook{
				Root:    root,
				Command: fmt.Sprintf("cp $WT_REPO_ROOT/%s $WT_WORKTREE_PATH/%s", filePath, filePath),
				Subdir:  "",
			}

			if editIndex >= 0 {
				copyHooks[editIndex] = newHook
			} else {
				copyHooks = append(copyHooks, newHook)
			}
		}
	}

	// Question 3: Command Hooks
	for {
		var selected string
		options := []huh.Option[string]{
			huh.NewOption("✨ Create new command hook", "__new__"),
		}

		for i, hook := range commandHooks {
			label := hook.Command
			if hook.Subdir != "" {
				label += fmt.Sprintf(" (in %s)", hook.Subdir)
			}
			options = append(options, huh.NewOption(label, fmt.Sprintf("__edit_%d__", i)))
		}

		options = append(options, huh.NewOption("✓ Done with command hooks", "__done__"))

		err = huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Command Hooks").
					Description("Select a hook to edit or create a new one").
					Options(options...).
					Value(&selected),
			),
		).Run()
		if err != nil {
			return err
		}

		if selected == "__done__" {
			break
		}

		var cmd string
		var subdir string
		var editIndex int = -1

		if selected == "__new__" {
			// New hook
			cmd = ""
			subdir = ""
		} else {
			// Edit existing
			fmt.Sscanf(selected, "__edit_%d__", &editIndex)
			cmd = commandHooks[editIndex].Command
			subdir = commandHooks[editIndex].Subdir
		}

		// Show command and subdir inputs
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Command").
					Description("Command to run (e.g., 'pnpm install')").
					Value(&cmd),

				huh.NewInput().
					Title("Subdirectory").
					Description("Subdirectory relative to worktree root (leave empty for root)").
					Value(&subdir),
			),
		).Run()
		if err != nil {
			return err
		}

		if cmd == "" {
			continue
		}

		// Ask Edit or Delete
		var shouldDelete bool
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Delete this hook?").
					Description("No = Save, Yes = Delete").
					Affirmative("Delete").
					Negative("Save").
					Value(&shouldDelete),
			),
		).Run()
		if err != nil {
			return err
		}

		if shouldDelete {
			// Delete
			if editIndex >= 0 {
				commandHooks = append(commandHooks[:editIndex], commandHooks[editIndex+1:]...)
			}
		} else {
			// Save/Update
			newHook := config.Hook{
				Root:    root,
				Command: cmd,
				Subdir:  strings.TrimSpace(subdir),
			}

			if editIndex >= 0 {
				commandHooks[editIndex] = newHook
			} else {
				commandHooks = append(commandHooks, newHook)
			}
		}
	}

	// Combine all hooks (preserve hooks from other repos)
	var allHooks []config.Hook
	for _, hook := range cfg.Hooks {
		if hook.Root != root {
			allHooks = append(allHooks, hook)
		}
	}
	allHooks = append(allHooks, copyHooks...)
	allHooks = append(allHooks, commandHooks...)

	// Save config
	newCfg := &config.Config{
		PathDefault: pathDefault,
		Hooks:       allHooks,
	}

	if err := config.Save(newCfg); err != nil {
		return err
	}

	fmt.Println("\n✓ Configuration saved successfully!")
	return nil
}
