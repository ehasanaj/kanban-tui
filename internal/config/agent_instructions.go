// Package config handles application configuration loading and management.
package config

// DefaultAgentInstructions is the content for the AGENT.md file created in the kanban directory.
const DefaultAgentInstructions = `# Kanban Agent Instructions

This directory contains a kanban board stored as markdown files. Each ticket is a markdown file with YAML frontmatter, organized into column directories.

## Directory Structure

` + "```" + `
.kanban/
├── AGENT.md        # This file
├── config.yaml     # Configuration (optional)
├── todo/           # Tasks to be done
├── doing/          # Tasks in progress
└── done/           # Completed tasks
` + "```" + `

## Ticket Format

Tickets are markdown files with YAML frontmatter:

` + "```" + `markdown
---
title: "Task title"
tags: ["tag1", "tag2"]
created: 2025-01-01T10:00:00Z
updated: 2025-01-01T10:00:00Z
agent_feedback: "Summary of work done"
---

Task description and details in markdown format.
` + "```" + `

### Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| title | Yes | Short task title |
| tags | No | Array of tags for categorization |
| created | Yes | ISO 8601 timestamp when ticket was created |
| updated | Yes | ISO 8601 timestamp when ticket was last modified |
| agent_feedback | No | Brief summary of changes made (add when completing) |

### Filename Convention

Filenames follow the pattern: ` + "`YYYY-MM-DD-slugified-title.md`" + `

Example: ` + "`2025-01-15-implement-user-auth.md`" + `

## Creating a Ticket

1. Create a new markdown file in the appropriate column directory
2. Add YAML frontmatter with required fields
3. Add task description in markdown below the frontmatter

Example:
` + "```" + `bash
cat > .kanban/todo/$(date +%Y-%m-%d)-my-new-task.md << 'EOF'
---
title: "My new task"
tags: ["feature"]
created: $(date -u +%Y-%m-%dT%H:%M:%SZ)
updated: $(date -u +%Y-%m-%dT%H:%M:%SZ)
---

Description of what needs to be done.
EOF
` + "```" + `

## Moving Tickets

Move tickets between columns by moving the file:

` + "```" + `bash
# Start working on a ticket
mv .kanban/todo/2025-01-15-my-task.md .kanban/doing/

# Complete a ticket
mv .kanban/doing/2025-01-15-my-task.md .kanban/done/
` + "```" + `

## Updating Tickets

When modifying a ticket:
1. Update the ` + "`updated`" + ` timestamp to the current time
2. Modify the content as needed

When completing a ticket:
1. Add the ` + "`agent_feedback`" + ` field with a brief summary of changes made
2. Move the ticket to the ` + "`done/`" + ` directory

## Workflow

1. **Start**: Move ticket from ` + "`todo/`" + ` to ` + "`doing/`" + `
2. **Work**: Implement the task as described
3. **Complete**: Add ` + "`agent_feedback`" + `, move to ` + "`done/`" + `

## Configuration

The ` + "`config.yaml`" + ` file (if present) can customize:
- ` + "`kanban_dir`" + `: Root directory for kanban data
- ` + "`columns`" + `: Column names, directories, and colors
- ` + "`editor`" + `: External editor command
- ` + "`single_ticket_prompt`" + `: Template for single ticket AI prompts
- ` + "`batch_ticket_prompt`" + `: Template for batch AI prompts

Prompt templates use Go text/template syntax with variables like ` + "`{{.TicketPath}}`" + `, ` + "`{{.DoingPath}}`" + `, ` + "`{{.DonePath}}`" + `.
`
