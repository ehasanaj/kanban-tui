# Kanban TUI

A terminal-based kanban board that stores tickets as markdown files. Perfect for personal task management that works seamlessly with both humans and AI agents.

## Features

- **Live Reload**: Automatically updates when files change (great for AI agent collaboration)
- **Markdown Tickets**: Human-readable tickets with YAML frontmatter
- **Vim-like Navigation**: Fast keyboard-driven interface
- **Configurable Columns**: Define your own workflow stages
- **Single Binary**: No runtime dependencies, works everywhere
- **Cross-Platform**: Linux, macOS, and Windows support

## Installation

### From Source

```bash
git clone https://github.com/user/kanban-tui.git
cd kanban-tui
go build -o kanban ./cmd/kanban
```

### Go Install

```bash
go install github.com/user/kanban-tui/cmd/kanban@latest
```

## Quick Start

```bash
# Run with default settings (~/.kanban directory)
kanban

# Use a specific directory
kanban -dir ./my-project/tasks

# Use a custom config
kanban -config ./config.yaml
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `h` / `l` | Move between columns |
| `j` / `k` | Move between tickets |
| `n` | Create new ticket |
| `e` | Edit ticket in $EDITOR |
| `d` | Delete ticket |
| `m` | Move ticket to another column |
| `Enter` | View ticket details |
| `/` | Search tickets |
| `r` | Refresh board |
| `?` | Show help |
| `q` | Quit |

## Ticket Format

Tickets are stored as markdown files with YAML frontmatter:

```markdown
---
title: "Implement user authentication"
tags: ["backend", "security"]
created: 2025-01-01T10:00:00Z
updated: 2025-01-01T10:00:00Z
---

# Implementation Details

- Add JWT authentication
- Set up password hashing
- Create login endpoint
```

## Configuration

Create `~/.config/kanban-tui/config.yaml`:

```yaml
# Root directory for kanban data
kanban_dir: ~/.kanban

# Column definitions
columns:
  - name: Backlog
    dir: backlog
  - name: To Do
    dir: todo
  - name: In Progress
    dir: doing
  - name: Review
    dir: review
  - name: Done
    dir: done

# External editor (defaults to $EDITOR)
editor: nvim
```

## Directory Structure

```
~/.kanban/
├── todo/
│   ├── 2025-01-01-implement-auth.md
│   └── 2025-01-02-add-logging.md
├── doing/
│   └── 2025-01-01-fix-bug.md
└── done/
    └── 2024-12-30-setup-project.md
```

## AI Agent Integration

This kanban board is designed to work with AI agents. Agents can:

1. **Create tickets**: Write markdown files to column directories
2. **Update tickets**: Modify existing markdown files
3. **Move tickets**: Move files between column directories
4. **Read tickets**: Parse markdown files to understand tasks

The TUI automatically detects file changes and updates in real-time.

### Example: Creating a ticket with an AI agent

```bash
cat > ~/.kanban/todo/$(date +%Y-%m-%d)-new-task.md << 'EOF'
---
title: "Task created by AI"
tags: ["ai-generated"]
created: 2025-01-01T10:00:00Z
updated: 2025-01-01T10:00:00Z
---

This task was created by an AI agent.
EOF
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Disclaimer

This project was primarily vibe-coded with AI assistance.

## License

MIT License - see [LICENSE](LICENSE) for details.
