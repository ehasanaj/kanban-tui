package ui

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/atotto/clipboard"
	"github.com/user/kanban-tui/internal/models"
)

// TicketPromptData holds data for single ticket template rendering.
type TicketPromptData struct {
	Title      string
	Tags       string
	Content    string
	TicketPath string
	DonePath   string
	DoingPath  string
}

// BatchPromptData holds data for batch ticket template rendering.
type BatchPromptData struct {
	Tickets []TicketPromptData
}

// buildTicketPromptData creates template data from a ticket.
func (m *Model) buildTicketPromptData(ticket *models.Ticket) TicketPromptData {
	// Project root is parent of .kanban directory
	projectRoot := filepath.Dir(m.config.KanbanDir)
	relativePath, err := filepath.Rel(projectRoot, ticket.FilePath)
	if err != nil {
		relativePath = ticket.FilePath
	}

	// Build paths relative to project root
	filename := filepath.Base(ticket.FilePath)
	donePath := filepath.Join(".kanban", "done", filename)
	doingPath := filepath.Join(".kanban", "doing", filename)

	return TicketPromptData{
		Title:      ticket.Title,
		Tags:       strings.Join(ticket.Tags, ", "),
		Content:    ticket.Content,
		TicketPath: relativePath,
		DonePath:   donePath,
		DoingPath:  doingPath,
	}
}

// renderSingleTicketPrompt renders the single ticket template.
func (m *Model) renderSingleTicketPrompt(ticket *models.Ticket) (string, error) {
	tmpl, err := template.New("single").Parse(m.config.SingleTicketPrompt)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	data := m.buildTicketPromptData(ticket)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}

// renderBatchTicketPrompt renders the batch ticket template.
func (m *Model) renderBatchTicketPrompt(tickets []*models.Ticket) (string, error) {
	tmpl, err := template.New("batch").Parse(m.config.BatchTicketPrompt)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var ticketData []TicketPromptData
	for _, t := range tickets {
		ticketData = append(ticketData, m.buildTicketPromptData(t))
	}

	data := BatchPromptData{Tickets: ticketData}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}

// copyToClipboard copies text to the system clipboard.
func copyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}
