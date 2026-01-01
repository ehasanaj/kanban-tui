// Package config handles application configuration loading and management.
package config

// DefaultSingleTicketPrompt is the default template for copying a single ticket prompt.
const DefaultSingleTicketPrompt = `Implement the task described in this ticket: @{{.TicketPath}}

## Guidelines
- First, read and understand the ticket requirements thoroughly
- Plan your approach before writing code
- Follow the existing coding style, patterns, and conventions of this project
- Keep changes focused - only modify what's necessary for the task
- Ensure existing functionality is not broken
- Test your changes if the project has tests

## Workflow
1. Move the ticket to doing: mv "{{.TicketPath}}" "{{.DoingPath}}"
2. Implement the task as described in the ticket
3. When complete, move the ticket to done: mv "{{.DoingPath}}" "{{.DonePath}}"
4. Update the agent_feedback field in the ticket's YAML frontmatter with a brief summary of the changes made
`

// DefaultBatchTicketPrompt is the default template for copying all todo tickets prompt.
const DefaultBatchTicketPrompt = `Implement the tasks described in the following tickets, in order:
{{range .Tickets}}
- @{{.TicketPath}}
{{- end}}

## Guidelines
- Read and understand each ticket's requirements before starting
- Plan your approach for each task before writing code
- Follow the existing coding style, patterns, and conventions of this project
- Keep changes focused - only modify what's necessary for each task
- Ensure existing functionality is not broken
- Test your changes if the project has tests
- Complete each ticket fully before moving to the next

## Workflow (for each ticket)
1. Move the ticket to doing: mv "<ticket_path>" ".kanban/doing/<filename>"
2. Implement the task as described in the ticket
3. When complete, move the ticket to done: mv ".kanban/doing/<filename>" ".kanban/done/<filename>"
4. Update the agent_feedback field in the ticket's YAML frontmatter with a brief summary of the changes made

Process tickets in the order listed above.
`
