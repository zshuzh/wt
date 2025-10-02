package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
)

// ============================================
// KONG - CLI argument parsing
// ============================================
var CLI struct {
	Version kong.VersionFlag `short:"v" help:"Print version information"`
	Init    struct{}         `cmd:"" help:"Start the wt setup wizard"`
	List    struct{}         `cmd:"" help:"List current worktrees"`
}

func main() {
	ctx := kong.Parse(&CLI,
		kong.Name("wt"),
		kong.Description("Git worktree manager optimized for AI workflows"),
		kong.UsageOnError(),
		kong.Vars{"version": "0.1.0"},
	)

	switch ctx.Command() {
	case "init":
		handleInit()
	case "list":
		handleList()
	default:
		// This shouldn't happen as kong handles unknown commands
		panic(ctx.Command())
	}
}

func handleInit() {
	fmt.Println("Initializing wt...")
	// Add your init logic here
}

func handleList() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

// ============================================
// BUBBLETEA - TUI
// ============================================
type model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

// pre event-loop initialization
func initialModel() model {
	return model{
		choices:  []string{"carrots", "celery", "kohlrabi"},
		selected: make(map[int]struct{}),
	}
}

// post event-loop initialization
func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}

	}

	return m, nil
}

func (m model) View() string {
	s := "What should we buy at the market?\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	s += "\nPress q to quit.\n"

	return s
}
