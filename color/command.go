package color

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/zshuzh/wt/internal/git"
)

type Options struct{}

type colorChoice struct {
	Name       string
	Background string
	Foreground string
}

var colors = []colorChoice{
	{Name: "Isometric Pink", Background: "#f9dfe2", Foreground: "#1a1a1a"},
	{Name: "Reforestation Green", Background: "#3a9e52", Foreground: "#ffffff"},
	{Name: "Soil Brown", Background: "#8c6a2e", Foreground: "#ffffff"},
	{Name: "Climate Disaster Red", Background: "#ee1111", Foreground: "#ffffff"},
	{Name: "None", Background: "", Foreground: ""},
}

func (o Options) Run() error {
	wt, err := git.GetCurrentWorktree()
	if err != nil {
		return err
	}

	opts := make([]huh.Option[int], len(colors))
	for i, c := range colors {
		label := c.Name
		if c.Background != "" {
			label = lipgloss.NewStyle().Foreground(lipgloss.Color(c.Background)).Render(c.Name)
		}
		opts[i] = huh.NewOption(label, i)
	}

	var selected int
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Select a color for this worktree").
				Options(opts...).
				Value(&selected),
		),
	).WithOutput(os.Stderr)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}

	choice := colors[selected]
	return applyColor(wt.Path, choice)
}

func applyColor(wtPath string, choice colorChoice) error {
	settingsDir := filepath.Join(wtPath, ".vscode")
	settingsPath := filepath.Join(settingsDir, "settings.json")

	// Read existing settings or start fresh
	settings := make(map[string]any)
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("parsing %s: %w", settingsPath, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("reading %s: %w", settingsPath, err)
	}

	if choice.Background == "" {
		// Clear colors
		delete(settings, "workbench.colorCustomizations")
	} else {
		customizations, ok := settings["workbench.colorCustomizations"].(map[string]any)
		if !ok {
			customizations = make(map[string]any)
		}

		customizations["titleBar.activeBackground"] = choice.Background
		customizations["titleBar.activeForeground"] = choice.Foreground
		customizations["statusBar.background"] = choice.Background
		customizations["statusBar.foreground"] = choice.Foreground

		settings["workbench.colorCustomizations"] = customizations
	}

	// If clearing left us with an empty map, remove the file
	if len(settings) == 0 {
		if err := os.Remove(settingsPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		fmt.Fprintln(os.Stderr, "Cleared worktree color.")
		return nil
	}

	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", settingsDir, err)
	}

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')

	if err := os.WriteFile(settingsPath, out, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", settingsPath, err)
	}

	if choice.Background == "" {
		fmt.Fprintln(os.Stderr, "Cleared worktree color.")
	} else {
		coloredName := lipgloss.NewStyle().Foreground(lipgloss.Color(choice.Background)).Render(choice.Name)
		fmt.Fprintf(os.Stderr, "Set worktree color to %s.\n", coloredName)
	}

	return nil
}
