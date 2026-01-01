# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run Commands

```bash
# Build the binary
go build -o kanban ./cmd/kanban

# Run the application
./kanban

# Run with specific directory
./kanban -dir ./tasks

# Run with custom config
./kanban -config ./config.yaml
```

## Architecture

This is a Go TUI application using the Bubbletea framework. Tickets are markdown files with YAML frontmatter stored in filesystem directories (one directory per kanban column).

### Package Structure

- **cmd/kanban/** - Entry point, CLI flag parsing, program initialization
- **internal/config/** - YAML config loading, defaults, directory creation
- **internal/models/** - Ticket struct, markdown/YAML parsing, file operations (Save, Move, Delete)
- **internal/ui/** - Bubbletea Model with view modes, keyboard handlers, and renderers
- **internal/watcher/** - fsnotify-based file watcher with debouncing for live reload

### UI Model Pattern

The UI uses Bubbletea's Elm-inspired architecture:
- `Model` struct holds all application state including view mode, active column/ticket indices, and input fields
- `Update()` routes messages to mode-specific handlers (`handleBoardKeys`, etc.)
- `View()` routes to mode-specific renderers based on `ViewMode` enum
- Eight view modes: Board, NewTicket, EditTicket, ViewTicket, MoveTicket, ConfirmDelete, Help, Search

### Data Flow

1. Config loaded from `~/.config/kanban-tui/config.yaml` (or defaults)
2. Column directories created via `EnsureDirectories()`
3. Tickets loaded by reading `.md` files from each column directory
4. File watcher monitors directories; changes trigger `loadAllTickets()` reload
5. CRUD operations write/rename/delete files, watcher detects and reloads

### Ticket File Format

```markdown
---
title: "Task title"
tags: ["tag1", "tag2"]
created: 2025-01-01T10:00:00Z
updated: 2025-01-01T10:00:00Z
---

Markdown content here
```

Filenames: `{YYYY-MM-DD}-{slugified-title}.md`
