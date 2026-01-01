// Package config handles application configuration loading and management.
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Column represents a kanban column configuration.
type Column struct {
	Name  string `yaml:"name"`
	Dir   string `yaml:"dir"`
	Color string `yaml:"color,omitempty"`
}

// Config holds the application configuration.
type Config struct {
	// KanbanDir is the root directory for kanban data
	KanbanDir string `yaml:"kanban_dir"`
	// Columns defines the kanban columns
	Columns []Column `yaml:"columns"`
	// Editor is the external editor command (defaults to $EDITOR)
	Editor string `yaml:"editor,omitempty"`
	// SingleTicketPrompt is the template for copying a single ticket's agent prompt
	SingleTicketPrompt string `yaml:"single_ticket_prompt,omitempty"`
	// BatchTicketPrompt is the template for copying all todo tickets' agent prompt
	BatchTicketPrompt string `yaml:"batch_ticket_prompt,omitempty"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	// Use current working directory by default
	cwd, _ := os.Getwd()
	kanbanDir := filepath.Join(cwd, ".kanban")

	return &Config{
		KanbanDir: kanbanDir,
		Columns: []Column{
			{Name: "To Do", Dir: "todo", Color: "#f87171"},
			{Name: "Doing", Dir: "doing", Color: "#fbbf24"},
			{Name: "Done", Dir: "done", Color: "#4ade80"},
		},
		Editor:             os.Getenv("EDITOR"),
		SingleTicketPrompt: DefaultSingleTicketPrompt,
		BatchTicketPrompt:  DefaultBatchTicketPrompt,
	}
}

// Load reads configuration from a YAML file.
// If the file doesn't exist, it creates a default configuration file.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create config file with defaults on first run
			if saveErr := cfg.Save(path); saveErr != nil {
				// Non-fatal: continue with defaults if we can't save
				return cfg, nil
			}
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Apply defaults for missing values
	if cfg.KanbanDir == "" {
		cfg.KanbanDir = DefaultConfig().KanbanDir
	}
	if len(cfg.Columns) == 0 {
		cfg.Columns = DefaultConfig().Columns
	}
	if cfg.Editor == "" {
		cfg.Editor = os.Getenv("EDITOR")
	}
	if cfg.SingleTicketPrompt == "" {
		cfg.SingleTicketPrompt = DefaultSingleTicketPrompt
	}
	if cfg.BatchTicketPrompt == "" {
		cfg.BatchTicketPrompt = DefaultBatchTicketPrompt
	}

	return cfg, nil
}

// Save writes the configuration to a YAML file.
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// EnsureDirectories creates the kanban directory structure.
func (c *Config) EnsureDirectories() error {
	if err := os.MkdirAll(c.KanbanDir, 0755); err != nil {
		return err
	}

	for _, col := range c.Columns {
		colPath := filepath.Join(c.KanbanDir, col.Dir)
		if err := os.MkdirAll(colPath, 0755); err != nil {
			return err
		}
	}

	return nil
}

// ColumnPath returns the full path for a column directory.
func (c *Config) ColumnPath(colDir string) string {
	return filepath.Join(c.KanbanDir, colDir)
}
