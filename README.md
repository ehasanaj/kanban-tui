# Kanban TUI

A terminal-based kanban board that stores tickets as markdown files. Perfect for personal task management that works seamlessly with both humans and AI agents.

## Features

- **Live Reload**: Automatically updates when files change (great for AI agent collaboration)
- **Markdown Tickets**: Human-readable tickets with YAML frontmatter
- **Vim-like Navigation**: Fast keyboard-driven interface with mouse support
- **AI Agent Integration**: Copy prompts to clipboard, track agent feedback per ticket
- **Configurable Columns**: Define your own workflow stages with custom colors
- **Single Binary**: No runtime dependencies, works everywhere
- **Cross-Platform**: Linux, macOS, and Windows support
- **Gruvbox Theme**: Beautiful dark color scheme

## Installation

### From Source

```bash
git clone https://github.com/user/kanban-tui.git
cd kanban-tui
go build -o kanban ./cmd/kanban
```

```

## Quick Start

```bash
# Run with default settings (.kanban directory in current folder)
kanban

# Use a specific directory
kanban -dir ./my-project/tasks

# Use a custom config
kanban -config ./config.yaml

# Show version
kanban -version
```

## Keyboard Shortcuts

### Navigation (Board View)
| Key | Action |
|-----|--------|
| `h` / `←` | Move to left column |
| `l` / `→` | Move to right column |
| `j` / `↓` | Move to next ticket |
| `k` / `↑` | Move to previous ticket |

### Ticket Actions
| Key | Action |
|-----|--------|
| `n` | Create new ticket |
| `e` | Edit selected ticket |
| `d` | Delete ticket (with confirmation) |
| `m` | Move ticket to another column |
| `Enter` | View ticket details |

### AI Agent Integration
| Key | Action |
|-----|--------|
| `p` | Copy AI prompt for selected ticket to clipboard |
| `P` | Copy AI prompt for all todo tickets to clipboard |
| `f` | View agent feedback fullscreen (in ticket view) |

### Editor Mode (Create/Edit)
| Key | Action |
|-----|--------|
| `Tab` | Cycle focus: title → tags → content |
| `Shift+Tab` | Cycle focus backwards |
| `Ctrl+S` | Save ticket |
| `Esc` | Cancel and return to board |

### Other
| Key | Action |
|-----|--------|
| `/` | Search tickets by title |
| `r` | Refresh board |
| `?` | Toggle help |
| `q` | Quit |

## Ticket Format

Tickets are stored as markdown files with YAML frontmatter:

```markdown
---
title: "Implement user authentication"
tags: ["backend", "security"]
created: 2025-01-01T10:00:00Z
updated: 2025-01-01T10:00:00Z
agent_feedback: "Implemented JWT auth with bcrypt hashing"  # Optional: AI agent response
---

# Implementation Details

- Add JWT authentication
- Set up password hashing
- Create login endpoint
```

Filenames follow the pattern: `YYYY-MM-DD-slugified-title.md`

## Configuration

On first run, a config file is created at `.kanban/config.yaml` in the current directory. You can also specify a custom path with `-config`.

```yaml
# Root directory for kanban data (default: .kanban in current directory)
kanban_dir: .kanban

# Column definitions with optional colors
columns:
  - name: Backlog
    dir: backlog
    color: "#a78bfa"  # Optional hex color
  - name: To Do
    dir: todo
    color: "#f87171"
  - name: In Progress
    dir: doing
    color: "#fbbf24"
  - name: Review
    dir: review
    color: "#60a5fa"
  - name: Done
    dir: done
    color: "#4ade80"

# External editor (defaults to $EDITOR env variable)
editor: nvim

# AI prompt templates (Go text/template syntax)
# Available variables: .TicketPath, .DoingPath, .DonePath, .Title, .Tags, .Content
single_ticket_prompt: |
  Implement the task described in this ticket: @{{.TicketPath}}
  ...

# For batch prompts, .Tickets is available for iteration
batch_ticket_prompt: |
  Implement the following tickets in order:
  {{range .Tickets}}
  - @{{.TicketPath}}
  {{- end}}
  ...
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

### Prompt Copy Feature

Press `p` to copy an AI-ready prompt for the selected ticket to your clipboard. The prompt includes:
- The ticket file path (for AI tools like Claude Code that support `@file` references)
- Guidelines for implementation
- Workflow instructions (move to doing → implement → move to done)

Press `P` (shift) to copy a batch prompt for all tickets in the first column (typically "To Do").

### Agent Feedback

When an AI agent completes a task, it can add feedback to the ticket's `agent_feedback` field:

```yaml
agent_feedback: "Implemented feature X with Y approach. Added tests."
```

View this feedback in the ticket view, or press `f` for fullscreen mode.

### Customizing Prompts

Configure prompt templates in your config file using Go `text/template` syntax:

```yaml
single_ticket_prompt: |
  Please implement: @{{.TicketPath}}
  Move to {{.DoingPath}} when starting, {{.DonePath}} when complete.
```

### Example: Creating a ticket with an AI agent

```bash
cat > .kanban/todo/$(date +%Y-%m-%d)-new-task.md << 'EOF'
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
