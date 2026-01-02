// Kanban TUI - A terminal-based kanban board with markdown tickets.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/kanban-tui/internal/config"
	"github.com/user/kanban-tui/internal/ui"
)

var (
	version = "0.1.0"
)

func main() {
	// Command line flags
	configPath := flag.String("config", "", "Path to config file")
	kanbanDir := flag.String("dir", "", "Kanban directory (overrides config)")
	showVersion := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("kanban-tui v%s\n", version)
		os.Exit(0)
	}

	// Determine config path
	cfgPath := *configPath
	if cfgPath == "" {
		cfgPath = ".kanban/config.yaml"
	}

	// Load configuration
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Override kanban directory if specified
	if *kanbanDir != "" {
		absDir, err := filepath.Abs(*kanbanDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving directory: %v\n", err)
			os.Exit(1)
		}
		cfg.KanbanDir = absDir
	}

	// Ensure directories exist
	if err := cfg.EnsureDirectories(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directories: %v\n", err)
		os.Exit(1)
	}

	// Create the UI model
	model, err := ui.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing UI: %v\n", err)
		os.Exit(1)
	}

	// Run the program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
