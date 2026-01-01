package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/kanban-tui/internal/config"
	"github.com/user/kanban-tui/internal/models"
	"github.com/user/kanban-tui/internal/watcher"
)

// View modes for the application.
type ViewMode int

const (
	ViewBoard ViewMode = iota
	ViewNewTicket
	ViewEditTicket
	ViewTicket // View mode (read-only)
	ViewMoveTicket
	ViewConfirmDelete
	ViewHelp
	ViewSearch
	ViewAgentFeedback // Fullscreen agent feedback view
)

// Editor modes for the ticket editor
const (
	EditorModeCreate = iota
	EditorModeEdit
	EditorModeView
)

// Messages for the Bubble Tea update loop.
type (
	tickMsg         time.Time
	fileChangeMsg   watcher.Event
	watcherErrorMsg error
	statusClearMsg  struct{}
)

// Model represents the application state.
type Model struct {
	config  *config.Config
	styles  Styles
	watcher *watcher.Watcher

	// Board state
	columns       []ColumnData
	activeColumn  int
	activeTicket  int
	width, height int

	// View state
	viewMode   ViewMode
	prevMode   ViewMode
	showDetail bool

	// Input state
	titleInput   textinput.Model
	tagsInput    textinput.Model
	contentInput textarea.Model
	searchInput  textinput.Model
	searchQuery  string
	editorFocus  int // 0 = title, 1 = tags, 2 = content
	editorMode   int // 0 = create, 1 = edit, 2 = view

	// Editing state
	editingTicket *models.Ticket // The ticket being edited (nil for create)

	// Status/feedback
	statusMessage string
	statusTimeout time.Time

	// Modal state
	confirmAction func() tea.Cmd
	moveTarget    int

	// Error state
	lastError error
}

// ColumnData holds column information and tickets.
type ColumnData struct {
	Config  config.Column
	Tickets []*models.Ticket
}

// New creates a new Model with the given configuration.
func New(cfg *config.Config) (*Model, error) {
	// Create file watcher
	w, err := watcher.New(150 * time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("creating watcher: %w", err)
	}

	// Watch all column directories
	for _, col := range cfg.Columns {
		colPath := cfg.ColumnPath(col.Dir)
		if err := w.Add(colPath); err != nil {
			return nil, fmt.Errorf("watching %s: %w", colPath, err)
		}
	}

	// Initialize text inputs
	ti := textinput.New()
	ti.Placeholder = "Enter ticket title..."
	ti.CharLimit = 100
	ti.Width = 60

	// Initialize tags input
	tg := textinput.New()
	tg.Placeholder = "Enter tags (comma-separated)..."
	tg.CharLimit = 200
	tg.Width = 60

	// Initialize textarea for content
	ta := textarea.New()
	ta.Placeholder = "Enter ticket description (markdown supported)..."
	ta.CharLimit = 0 // No limit
	ta.SetWidth(60)
	ta.SetHeight(10)
	ta.ShowLineNumbers = false

	si := textinput.New()
	si.Placeholder = "Search tickets..."
	si.CharLimit = 50
	si.Width = 30

	m := &Model{
		config:       cfg,
		styles:       DefaultStyles(),
		watcher:      w,
		columns:      make([]ColumnData, len(cfg.Columns)),
		titleInput:   ti,
		tagsInput:    tg,
		contentInput: ta,
		searchInput:  si,
		activeColumn: 0,
		activeTicket: 0,
		viewMode:     ViewBoard,
		editorFocus:  0,
		editorMode:   EditorModeCreate,
	}

	// Initialize column data
	for i, col := range cfg.Columns {
		m.columns[i] = ColumnData{
			Config:  col,
			Tickets: []*models.Ticket{},
		}
	}

	// Load initial tickets
	if err := m.loadAllTickets(); err != nil {
		return nil, fmt.Errorf("loading tickets: %w", err)
	}

	return m, nil
}

// loadAllTickets loads tickets from all columns.
func (m *Model) loadAllTickets() error {
	for i, col := range m.config.Columns {
		tickets, err := m.loadColumnTickets(col.Dir)
		if err != nil {
			return err
		}
		m.columns[i].Tickets = tickets
	}
	return nil
}

// loadColumnTickets loads tickets from a specific column.
func (m *Model) loadColumnTickets(colDir string) ([]*models.Ticket, error) {
	colPath := m.config.ColumnPath(colDir)

	entries, err := os.ReadDir(colPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*models.Ticket{}, nil
		}
		return nil, err
	}

	var tickets []*models.Ticket
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		ticketPath := filepath.Join(colPath, entry.Name())
		ticket, err := models.ParseTicket(ticketPath)
		if err != nil {
			// Skip invalid tickets but log the error
			continue
		}
		tickets = append(tickets, ticket)
	}

	// Sort by updated date (newest first)
	sort.Slice(tickets, func(i, j int) bool {
		return tickets[i].Updated.After(tickets[j].Updated)
	})

	return tickets, nil
}

// Init initializes the model.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.watcherCmd(),
		textinput.Blink,
	)
}

// watcherCmd listens for file system events.
func (m *Model) watcherCmd() tea.Cmd {
	return func() tea.Msg {
		select {
		case event := <-m.watcher.Events:
			return fileChangeMsg(event)
		case err := <-m.watcher.Errors:
			return watcherErrorMsg(err)
		}
	}
}

// Update handles messages and updates the model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd := m.handleKeyPress(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case fileChangeMsg:
		// Reload tickets when files change
		m.loadAllTickets()
		cmds = append(cmds, m.watcherCmd())

	case watcherErrorMsg:
		m.lastError = msg
		cmds = append(cmds, m.watcherCmd())

	case statusClearMsg:
		m.statusMessage = ""
	}

	// Update text inputs if in input mode (create/edit modes only, not view)
	if m.viewMode == ViewNewTicket || m.viewMode == ViewEditTicket {
		var cmd tea.Cmd
		switch m.editorFocus {
		case 0:
			m.titleInput, cmd = m.titleInput.Update(msg)
		case 1:
			m.tagsInput, cmd = m.tagsInput.Update(msg)
		case 2:
			m.contentInput, cmd = m.contentInput.Update(msg)
		}
		cmds = append(cmds, cmd)
	}

	if m.viewMode == ViewSearch {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleKeyPress processes keyboard input.
func (m *Model) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	// Global keys
	switch msg.String() {
	case "ctrl+c":
		m.watcher.Close()
		return tea.Quit
	}

	// Mode-specific handling
	switch m.viewMode {
	case ViewBoard:
		return m.handleBoardKeys(msg)
	case ViewNewTicket, ViewEditTicket, ViewTicket:
		return m.handleTicketEditorKeys(msg)
	case ViewMoveTicket:
		return m.handleMoveTicketKeys(msg)
	case ViewConfirmDelete:
		return m.handleConfirmDeleteKeys(msg)
	case ViewHelp:
		return m.handleHelpKeys(msg)
	case ViewSearch:
		return m.handleSearchKeys(msg)
	case ViewAgentFeedback:
		return m.handleAgentFeedbackKeys(msg)
	}

	return nil
}

// handleBoardKeys handles keys in board view.
func (m *Model) handleBoardKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "q":
		m.watcher.Close()
		return tea.Quit

	case "h", "left":
		if m.activeColumn > 0 {
			m.activeColumn--
			m.activeTicket = 0
		}

	case "l", "right":
		if m.activeColumn < len(m.columns)-1 {
			m.activeColumn++
			m.activeTicket = 0
		}

	case "j", "down":
		tickets := m.getFilteredTickets(m.activeColumn)
		if m.activeTicket < len(tickets)-1 {
			m.activeTicket++
		}

	case "k", "up":
		if m.activeTicket > 0 {
			m.activeTicket--
		}

	case "n":
		m.viewMode = ViewNewTicket
		m.editorMode = EditorModeCreate
		m.editingTicket = nil
		m.titleInput.SetValue("")
		m.tagsInput.SetValue("")
		m.contentInput.SetValue("")
		m.editorFocus = 0
		m.titleInput.Focus()
		m.tagsInput.Blur()
		m.contentInput.Blur()
		return textinput.Blink

	case "enter":
		if m.hasSelectedTicket() {
			return m.openTicketEditor(EditorModeView)
		}

	case "d":
		if m.hasSelectedTicket() {
			m.viewMode = ViewConfirmDelete
		}

	case "m":
		if m.hasSelectedTicket() {
			m.viewMode = ViewMoveTicket
			m.moveTarget = m.activeColumn
		}

	case "e":
		if m.hasSelectedTicket() {
			return m.openTicketEditor(EditorModeEdit)
		}

	case "/":
		m.viewMode = ViewSearch
		m.searchInput.SetValue("")
		m.searchInput.Focus()
		return textinput.Blink

	case "?":
		m.viewMode = ViewHelp

	case "r":
		m.loadAllTickets()
		m.setStatus("Refreshed")

	case "p":
		return m.copySelectedTicketPrompt()

	case "P":
		return m.copyTodoTicketsPrompt()
	}

	return nil
}

// handleTicketEditorKeys handles keys in ticket editor (create/edit/view modes).
func (m *Model) handleTicketEditorKeys(msg tea.KeyMsg) tea.Cmd {
	// View mode specific handling
	if m.editorMode == EditorModeView {
		switch msg.String() {
		case "esc", "q":
			m.viewMode = ViewBoard
			m.resetEditorInputs()
			return nil
		case "e":
			// Switch to edit mode
			m.editorMode = EditorModeEdit
			m.viewMode = ViewEditTicket
			m.editorFocus = 0
			m.titleInput.Focus()
			return textinput.Blink
		case "f":
			// Open fullscreen agent feedback view
			if m.editingTicket != nil && m.editingTicket.AgentFeedback != "" {
				m.viewMode = ViewAgentFeedback
			}
			return nil
		}
		return nil
	}

	// Create and Edit mode handling
	switch msg.String() {
	case "esc":
		m.viewMode = ViewBoard
		m.resetEditorInputs()
		return nil

	case "tab":
		// Cycle focus: title → tags → content → title
		m.editorFocus = (m.editorFocus + 1) % 3
		m.updateEditorFocus()
		return nil

	case "shift+tab":
		// Cycle focus backwards
		m.editorFocus = (m.editorFocus + 2) % 3
		m.updateEditorFocus()
		return nil

	case "ctrl+s":
		// Save the ticket
		if m.editorMode == EditorModeEdit {
			return m.saveTicket()
		}
		return m.createTicket()
	}

	return nil
}

// updateEditorFocus updates which input field is focused.
func (m *Model) updateEditorFocus() {
	m.titleInput.Blur()
	m.tagsInput.Blur()
	m.contentInput.Blur()

	switch m.editorFocus {
	case 0:
		m.titleInput.Focus()
	case 1:
		m.tagsInput.Focus()
	case 2:
		m.contentInput.Focus()
	}
}

// handleMoveTicketKeys handles keys in move ticket view.
func (m *Model) handleMoveTicketKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.viewMode = ViewBoard

	case "h", "left":
		if m.moveTarget > 0 {
			m.moveTarget--
		}

	case "l", "right":
		if m.moveTarget < len(m.columns)-1 {
			m.moveTarget++
		}

	case "enter":
		return m.moveSelectedTicket()
	}

	return nil
}

// handleConfirmDeleteKeys handles keys in delete confirmation view.
func (m *Model) handleConfirmDeleteKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "n":
		m.viewMode = ViewBoard

	case "y", "enter":
		return m.deleteSelectedTicket()
	}

	return nil
}

// handleHelpKeys handles keys in help view.
func (m *Model) handleHelpKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "?", "q":
		m.viewMode = ViewBoard
	}

	return nil
}

// handleSearchKeys handles keys in search view.
func (m *Model) handleSearchKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.viewMode = ViewBoard
		m.searchQuery = ""
		m.activeTicket = 0 // Reset selection when clearing search
		m.searchInput.Blur()

	case "enter":
		m.searchQuery = m.searchInput.Value()
		m.activeTicket = 0 // Reset selection for filtered results
		m.viewMode = ViewBoard
		m.searchInput.Blur()
	}

	return nil
}

// handleAgentFeedbackKeys handles keys in agent feedback fullscreen view.
func (m *Model) handleAgentFeedbackKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q", "f":
		m.viewMode = ViewTicket
	}
	return nil
}

// getFilteredTickets returns tickets for a column, filtered by search query if active.
func (m *Model) getFilteredTickets(colIndex int) []*models.Ticket {
	if colIndex >= len(m.columns) {
		return nil
	}
	tickets := m.columns[colIndex].Tickets
	if m.searchQuery != "" {
		tickets = m.filterTickets(tickets)
	}
	return tickets
}

// hasSelectedTicket returns true if there's a valid ticket selected.
func (m *Model) hasSelectedTicket() bool {
	if m.activeColumn >= len(m.columns) {
		return false
	}
	tickets := m.getFilteredTickets(m.activeColumn)
	return m.activeTicket < len(tickets)
}

// getSelectedTicket returns the currently selected ticket.
func (m *Model) getSelectedTicket() *models.Ticket {
	tickets := m.getFilteredTickets(m.activeColumn)
	if m.activeTicket >= len(tickets) {
		return nil
	}
	return tickets[m.activeTicket]
}

// parseTagsInput parses the comma-separated tags input into a slice.
func (m *Model) parseTagsInput() []string {
	input := strings.TrimSpace(m.tagsInput.Value())
	if input == "" {
		return []string{}
	}
	parts := strings.Split(input, ",")
	var tags []string
	for _, p := range parts {
		tag := strings.TrimSpace(p)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

// resetEditorInputs clears all editor input fields.
func (m *Model) resetEditorInputs() {
	m.titleInput.SetValue("")
	m.tagsInput.SetValue("")
	m.contentInput.SetValue("")
	m.titleInput.Blur()
	m.tagsInput.Blur()
	m.contentInput.Blur()
	m.editorFocus = 0
	m.editingTicket = nil
}

// openTicketEditor opens a ticket in the editor with the specified mode.
func (m *Model) openTicketEditor(mode int) tea.Cmd {
	ticket := m.getSelectedTicket()
	if ticket == nil {
		return nil
	}

	m.editorMode = mode
	m.editingTicket = ticket

	// Populate fields from ticket
	m.titleInput.SetValue(ticket.Title)
	m.tagsInput.SetValue(strings.Join(ticket.Tags, ", "))
	m.contentInput.SetValue(ticket.Content)

	if mode == EditorModeView {
		m.viewMode = ViewTicket
		// Blur all inputs in view mode
		m.titleInput.Blur()
		m.tagsInput.Blur()
		m.contentInput.Blur()
	} else {
		m.viewMode = ViewEditTicket
		m.editorFocus = 0
		m.titleInput.Focus()
	}

	return textinput.Blink
}

// createTicket creates a new ticket with title, tags, and content.
func (m *Model) createTicket() tea.Cmd {
	title := strings.TrimSpace(m.titleInput.Value())
	if title == "" {
		m.setStatus("Error: Title cannot be empty")
		return nil
	}

	col := m.columns[m.activeColumn]
	ticket := models.NewTicket(title, col.Config.Dir)
	ticket.Tags = m.parseTagsInput()
	ticket.Content = strings.TrimSpace(m.contentInput.Value())
	ticket.FilePath = filepath.Join(
		m.config.ColumnPath(col.Config.Dir),
		ticket.GenerateFilename(),
	)

	if err := ticket.Save(); err != nil {
		m.setStatus(fmt.Sprintf("Error: %v", err))
	} else {
		m.setStatus(fmt.Sprintf("Created: %s", title))
	}

	m.viewMode = ViewBoard
	m.resetEditorInputs()
	m.loadAllTickets()

	return nil
}

// saveTicket saves changes to an existing ticket.
func (m *Model) saveTicket() tea.Cmd {
	if m.editingTicket == nil {
		return nil
	}

	title := strings.TrimSpace(m.titleInput.Value())
	if title == "" {
		m.setStatus("Error: Title cannot be empty")
		return nil
	}

	m.editingTicket.Title = title
	m.editingTicket.Tags = m.parseTagsInput()
	m.editingTicket.Content = strings.TrimSpace(m.contentInput.Value())

	if err := m.editingTicket.Save(); err != nil {
		m.setStatus(fmt.Sprintf("Error: %v", err))
	} else {
		m.setStatus(fmt.Sprintf("Updated: %s", title))
	}

	m.viewMode = ViewBoard
	m.resetEditorInputs()
	m.loadAllTickets()

	return nil
}

// deleteSelectedTicket deletes the selected ticket.
func (m *Model) deleteSelectedTicket() tea.Cmd {
	ticket := m.getSelectedTicket()
	if ticket == nil {
		return nil
	}

	if err := ticket.Delete(); err != nil {
		m.setStatus(fmt.Sprintf("Error: %v", err))
	} else {
		m.setStatus(fmt.Sprintf("Deleted: %s", ticket.Title))
	}

	m.viewMode = ViewBoard
	m.loadAllTickets()

	// Adjust selection if needed
	col := m.columns[m.activeColumn]
	if m.activeTicket >= len(col.Tickets) && m.activeTicket > 0 {
		m.activeTicket--
	}

	return nil
}

// moveSelectedTicket moves the selected ticket to a new column.
func (m *Model) moveSelectedTicket() tea.Cmd {
	ticket := m.getSelectedTicket()
	if ticket == nil {
		return nil
	}

	if m.moveTarget == m.activeColumn {
		m.viewMode = ViewBoard
		return nil
	}

	targetCol := m.columns[m.moveTarget].Config.Dir

	if err := ticket.Move(m.config.KanbanDir, targetCol); err != nil {
		m.setStatus(fmt.Sprintf("Error: %v", err))
	} else {
		m.setStatus(fmt.Sprintf("Moved to %s", m.columns[m.moveTarget].Config.Name))
	}

	m.viewMode = ViewBoard
	m.loadAllTickets()

	// Adjust selection if needed
	col := m.columns[m.activeColumn]
	if m.activeTicket >= len(col.Tickets) && m.activeTicket > 0 {
		m.activeTicket--
	}

	return nil
}

// setStatus sets a temporary status message.
func (m *Model) setStatus(msg string) {
	m.statusMessage = msg
	m.statusTimeout = time.Now().Add(3 * time.Second)
}

// copySelectedTicketPrompt copies the prompt for the selected ticket to clipboard.
func (m *Model) copySelectedTicketPrompt() tea.Cmd {
	ticket := m.getSelectedTicket()
	if ticket == nil {
		m.setStatus("No ticket selected")
		return nil
	}

	prompt, err := m.renderSingleTicketPrompt(ticket)
	if err != nil {
		m.setStatus(fmt.Sprintf("Error: %v", err))
		return nil
	}

	if err := copyToClipboard(prompt); err != nil {
		m.setStatus(fmt.Sprintf("Clipboard error: %v", err))
		return nil
	}

	m.setStatus(fmt.Sprintf("Copied prompt for: %s", ticket.ShortTitle(30)))
	return nil
}

// copyTodoTicketsPrompt copies prompts for all tickets in the first column.
func (m *Model) copyTodoTicketsPrompt() tea.Cmd {
	if len(m.columns) == 0 {
		m.setStatus("No columns configured")
		return nil
	}

	todoColumn := m.columns[0]
	if len(todoColumn.Tickets) == 0 {
		m.setStatus("No tickets in todo column")
		return nil
	}

	prompt, err := m.renderBatchTicketPrompt(todoColumn.Tickets)
	if err != nil {
		m.setStatus(fmt.Sprintf("Error: %v", err))
		return nil
	}

	if err := copyToClipboard(prompt); err != nil {
		m.setStatus(fmt.Sprintf("Clipboard error: %v", err))
		return nil
	}

	m.setStatus(fmt.Sprintf("Copied %d todo ticket(s) to clipboard", len(todoColumn.Tickets)))
	return nil
}

// View renders the UI.
func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	switch m.viewMode {
	case ViewHelp:
		return m.renderHelp()
	case ViewNewTicket, ViewEditTicket, ViewTicket:
		return m.renderTicketEditor()
	case ViewConfirmDelete:
		return m.renderDeleteConfirmScreen()
	case ViewMoveTicket:
		return m.renderMoveScreen()
	case ViewSearch:
		return m.renderSearchScreen()
	case ViewAgentFeedback:
		return m.renderAgentFeedbackScreen()
	default:
		return m.renderBoard()
	}
}

// renderBoard renders the main board view.
func (m *Model) renderBoard() string {
	var b strings.Builder

	// Header
	header := m.styles.Header.Width(m.width - 4).Render("  Kanban Board")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Calculate column width
	colWidth := (m.width - 4 - len(m.columns)*2) / len(m.columns)
	if colWidth < 20 {
		colWidth = 20
	}

	// Render columns
	var columnViews []string
	for i, col := range m.columns {
		isActive := i == m.activeColumn
		columnViews = append(columnViews, m.renderColumn(col, i, colWidth, isActive))
	}

	// Join columns horizontally
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, columnViews...))
	b.WriteString("\n")

	// Status message
	if m.statusMessage != "" && time.Now().Before(m.statusTimeout) {
		b.WriteString("\n")
		b.WriteString(m.styles.StatusMessage.Render(m.statusMessage))
	}

	// Help bar at bottom
	b.WriteString("\n\n")
	b.WriteString(m.renderHelpBar())

	return m.styles.App.Render(b.String())
}

// renderColumn renders a single column.
func (m *Model) renderColumn(col ColumnData, colIndex, width int, isActive bool) string {
	var b strings.Builder

	// Filter tickets if searching
	tickets := col.Tickets
	if m.searchQuery != "" {
		tickets = m.filterTickets(tickets)
	}

	// Column header with color (show filtered count when searching)
	headerColor := GetColumnColor(col.Config.Dir)
	headerStyle := m.styles.ColumnHeader.Copy().Background(headerColor)
	count := m.styles.ColumnCount.Render(fmt.Sprintf("(%d)", len(tickets)))
	header := headerStyle.Render(col.Config.Name) + count
	b.WriteString(header)
	b.WriteString("\n")

	// Render tickets
	maxTickets := (m.height - 12) / 4
	if maxTickets < 3 {
		maxTickets = 3
	}

	for i, ticket := range tickets {
		if i >= maxTickets {
			remaining := len(tickets) - maxTickets
			b.WriteString(m.styles.TicketDate.Render(fmt.Sprintf("  +%d more...", remaining)))
			break
		}

		isSelected := isActive && i == m.activeTicket
		b.WriteString(m.renderTicket(ticket, width-4, isSelected))
	}

	if len(tickets) == 0 {
		empty := m.styles.TicketDate.Render("  No tickets")
		b.WriteString(empty)
	}

	// Apply column style
	style := m.styles.Column
	if isActive {
		style = m.styles.ColumnActive
	}

	return style.Width(width).Height(m.height - 10).Render(b.String())
}

// renderTicket renders a single ticket card.
func (m *Model) renderTicket(ticket *models.Ticket, width int, isSelected bool) string {
	var b strings.Builder

	title := m.styles.TicketTitle.Render(ticket.ShortTitle(width - 4))
	b.WriteString(title)
	b.WriteString("\n")

	if len(ticket.Tags) > 0 {
		tags := m.styles.TicketTags.Render(strings.Join(ticket.Tags, ", "))
		b.WriteString(tags)
		b.WriteString("\n")
	}

	date := m.styles.TicketDate.Render(ticket.Updated.Format("Jan 02"))
	b.WriteString(date)

	style := m.styles.Ticket
	if isSelected {
		style = m.styles.TicketSelected
	}

	return style.Width(width).Render(b.String())
}

// filterTickets filters tickets by search query.
func (m *Model) filterTickets(tickets []*models.Ticket) []*models.Ticket {
	if m.searchQuery == "" {
		return tickets
	}

	query := strings.ToLower(m.searchQuery)
	var filtered []*models.Ticket

	for _, t := range tickets {
		if strings.Contains(strings.ToLower(t.Title), query) {
			filtered = append(filtered, t)
		}
	}

	return filtered
}

// renderTicketEditor renders the unified ticket editor (create/edit/view modes).
func (m *Model) renderTicketEditor() string {
	var b strings.Builder

	isViewMode := m.editorMode == EditorModeView

	// Calculate content width (leave margins)
	contentWidth := m.width - 8
	if contentWidth > 80 {
		contentWidth = 80
	}
	if contentWidth < 40 {
		contentWidth = 40
	}

	// Update input widths to match
	m.titleInput.Width = contentWidth - 4
	m.tagsInput.Width = contentWidth - 4
	m.contentInput.SetWidth(contentWidth - 4)

	// Calculate content height for textarea
	taHeight := m.height - 22 // Account for tags field
	if taHeight < 5 {
		taHeight = 5
	}
	if taHeight > 15 {
		taHeight = 15
	}
	m.contentInput.SetHeight(taHeight)

	// Header based on mode
	var headerText string
	var columnText string
	switch m.editorMode {
	case EditorModeCreate:
		headerText = "  New Ticket"
		columnText = "Creating in: "
	case EditorModeEdit:
		headerText = "  Edit Ticket"
		columnText = "Editing in: "
	case EditorModeView:
		headerText = "  View Ticket"
		columnText = "Column: "
	}

	// Get column info
	var colName string
	var colDir string
	if m.editingTicket != nil {
		colDir = m.editingTicket.Column
		// Find column name
		for _, c := range m.columns {
			if c.Config.Dir == colDir {
				colName = c.Config.Name
				break
			}
		}
	} else {
		col := m.columns[m.activeColumn]
		colDir = col.Config.Dir
		colName = col.Config.Name
	}

	headerColor := GetColumnColor(colDir)
	columnBadge := lipgloss.NewStyle().
		Background(headerColor).
		Foreground(GruvboxBg0).
		Padding(0, 1).
		Bold(true).
		Render(colName)

	header := m.styles.Header.Width(contentWidth).Render(headerText)
	b.WriteString(header)
	b.WriteString("\n\n")

	// Column indicator
	b.WriteString(m.styles.HelpDesc.Render(columnText))
	b.WriteString(columnBadge)
	b.WriteString("\n\n")

	// Title field
	titleLabel := m.styles.ModalTitle.Render("Title")
	if !isViewMode && m.editorFocus == 0 {
		titleLabel = m.styles.ModalTitle.Copy().Foreground(GruvboxYellow).Render("▶ Title")
	}
	b.WriteString(titleLabel)
	b.WriteString("\n")

	if isViewMode {
		// View mode: show styled text
		titleContent := m.titleInput.Value()
		if titleContent == "" {
			titleContent = "(no title)"
		}
		b.WriteString(m.styles.Input.Width(contentWidth).Render(
			m.styles.TicketTitle.Render(titleContent)))
	} else {
		// Edit mode: show input
		titleStyle := m.styles.Input
		if m.editorFocus == 0 {
			titleStyle = m.styles.InputFocused
		}
		b.WriteString(titleStyle.Width(contentWidth).Render(m.titleInput.View()))
	}
	b.WriteString("\n\n")

	// Tags field
	tagsLabel := m.styles.ModalTitle.Render("Tags")
	if !isViewMode && m.editorFocus == 1 {
		tagsLabel = m.styles.ModalTitle.Copy().Foreground(GruvboxYellow).Render("▶ Tags")
	}
	b.WriteString(tagsLabel)
	b.WriteString("\n")

	if isViewMode {
		// View mode: show styled text
		tagsContent := m.tagsInput.Value()
		if tagsContent == "" {
			tagsContent = "(no tags)"
		}
		b.WriteString(m.styles.Input.Width(contentWidth).Render(
			m.styles.TicketTags.Render(tagsContent)))
	} else {
		// Edit mode: show input
		tagsStyle := m.styles.Input
		if m.editorFocus == 1 {
			tagsStyle = m.styles.InputFocused
		}
		b.WriteString(tagsStyle.Width(contentWidth).Render(m.tagsInput.View()))
	}
	b.WriteString("\n\n")

	// Content field
	contentLabel := m.styles.ModalTitle.Render("Content")
	if !isViewMode && m.editorFocus == 2 {
		contentLabel = m.styles.ModalTitle.Copy().Foreground(GruvboxYellow).Render("▶ Content")
	}
	b.WriteString(contentLabel)
	b.WriteString("\n")

	if isViewMode {
		// View mode: show styled text
		contentText := m.contentInput.Value()
		if contentText == "" {
			contentText = "(no content)"
		}
		b.WriteString(m.styles.Input.Width(contentWidth).Height(taHeight + 2).Render(contentText))
	} else {
		// Edit mode: show textarea
		contentStyle := m.styles.Input
		if m.editorFocus == 2 {
			contentStyle = m.styles.InputFocused
		}
		b.WriteString(contentStyle.Width(contentWidth).Height(taHeight + 2).Render(m.contentInput.View()))
	}
	b.WriteString("\n\n")

	// Agent feedback preview (view mode only, when feedback exists)
	if isViewMode && m.editingTicket != nil && m.editingTicket.AgentFeedback != "" {
		feedbackLabel := m.styles.ModalTitle.Copy().Foreground(GruvboxBlue).Render("Agent Feedback")
		b.WriteString(feedbackLabel)
		b.WriteString("\n")

		// Show truncated preview (first 100 chars or 2 lines)
		feedback := m.editingTicket.AgentFeedback
		previewLines := strings.SplitN(feedback, "\n", 3)
		preview := strings.Join(previewLines[:min(len(previewLines), 2)], "\n")
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		} else if len(previewLines) > 2 {
			preview += "..."
		}

		feedbackStyle := m.styles.Input.Width(contentWidth).Foreground(GruvboxBlue)
		b.WriteString(feedbackStyle.Render(preview))
		b.WriteString("\n")
		b.WriteString(m.styles.HelpDesc.Render("Press 'f' to view full feedback"))
		b.WriteString("\n\n")
	}

	// Status message if any
	if m.statusMessage != "" && time.Now().Before(m.statusTimeout) {
		b.WriteString(m.styles.StatusMessage.Render(m.statusMessage))
		b.WriteString("\n\n")
	}

	// Help bar based on mode
	var helpKeys []struct{ key, desc string }
	if isViewMode {
		helpKeys = []struct{ key, desc string }{
			{"e", "edit"},
			{"Esc", "back"},
		}
		// Show feedback shortcut only if agent feedback exists
		if m.editingTicket != nil && m.editingTicket.AgentFeedback != "" {
			helpKeys = []struct{ key, desc string }{
				{"e", "edit"},
				{"f", "feedback"},
				{"Esc", "back"},
			}
		}
	} else {
		helpKeys = []struct{ key, desc string }{
			{"Tab", "next field"},
			{"Ctrl+S", "save"},
			{"Esc", "cancel"},
		}
	}

	var parts []string
	for _, k := range helpKeys {
		key := m.styles.HelpKey.Render(k.key)
		desc := m.styles.HelpDesc.Render(k.desc)
		parts = append(parts, fmt.Sprintf("%s %s", key, desc))
	}

	helpText := strings.Join(parts, "    ")
	b.WriteString(m.styles.HelpBar.Width(contentWidth).Render(helpText))

	return m.styles.App.Render(b.String())
}

// renderMoveModal renders the move ticket modal.
func (m *Model) renderMoveModal() string {
	var b strings.Builder

	b.WriteString(m.styles.ModalTitle.Render("Move Ticket"))
	b.WriteString("\n\n")

	for i, col := range m.columns {
		style := m.styles.Button
		if i == m.moveTarget {
			style = m.styles.ButtonActive
		}
		b.WriteString(style.Render(col.Config.Name))
	}

	b.WriteString("\n\n")
	b.WriteString(m.styles.HelpDesc.Render("h/l to select, Enter to confirm, Esc to cancel"))

	return m.styles.Modal.Width(60).Render(b.String())
}

// renderDeleteConfirm renders the delete confirmation modal.
func (m *Model) renderDeleteConfirm() string {
	ticket := m.getSelectedTicket()
	title := ""
	if ticket != nil {
		title = ticket.Title
	}

	var b strings.Builder
	b.WriteString(m.styles.ModalTitle.Render("Delete Ticket?"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Are you sure you want to delete:\n%s", title))
	b.WriteString("\n\n")
	b.WriteString(m.styles.HelpDesc.Render("y to confirm, n/Esc to cancel"))

	return m.styles.Modal.Width(50).Render(b.String())
}

// renderSearchModal renders the search modal.
func (m *Model) renderSearchModal() string {
	var b strings.Builder

	b.WriteString(m.styles.ModalTitle.Render("Search Tickets"))
	b.WriteString("\n\n")
	b.WriteString(m.searchInput.View())
	b.WriteString("\n\n")
	b.WriteString(m.styles.HelpDesc.Render("Enter to search, Esc to cancel"))

	return m.styles.Modal.Width(50).Render(b.String())
}

// renderDeleteConfirmScreen renders the delete confirmation as a centered full-screen view.
func (m *Model) renderDeleteConfirmScreen() string {
	modal := m.renderDeleteConfirm()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}

// renderMoveScreen renders the move ticket modal as a centered full-screen view.
func (m *Model) renderMoveScreen() string {
	modal := m.renderMoveModal()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}

// renderSearchScreen renders the search modal as a centered full-screen view.
func (m *Model) renderSearchScreen() string {
	modal := m.renderSearchModal()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}

// renderAgentFeedbackScreen renders the agent feedback in fullscreen.
func (m *Model) renderAgentFeedbackScreen() string {
	var b strings.Builder

	// Calculate content width
	contentWidth := max(min(m.width-8, 100), 40)

	// Header
	header := m.styles.Header.Width(contentWidth).Render("  Agent Feedback")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Ticket title for context
	if m.editingTicket != nil {
		titleLabel := m.styles.HelpDesc.Render("Ticket: ")
		titleText := m.styles.TicketTitle.Render(m.editingTicket.Title)
		b.WriteString(titleLabel)
		b.WriteString(titleText)
		b.WriteString("\n\n")
	}

	// Feedback content
	feedbackLabel := m.styles.ModalTitle.Render("Feedback from AI Agent")
	b.WriteString(feedbackLabel)
	b.WriteString("\n\n")

	feedback := ""
	if m.editingTicket != nil {
		feedback = m.editingTicket.AgentFeedback
	}
	if feedback == "" {
		feedback = "(no agent feedback available)"
	}

	// Calculate available height for feedback content
	feedbackHeight := max(m.height-14, 5)

	feedbackStyle := m.styles.Input.Width(contentWidth).Height(feedbackHeight)
	b.WriteString(feedbackStyle.Render(feedback))
	b.WriteString("\n\n")

	// Help bar
	helpKeys := []struct{ key, desc string }{
		{"Esc/f", "back"},
	}

	var parts []string
	for _, k := range helpKeys {
		key := m.styles.HelpKey.Render(k.key)
		desc := m.styles.HelpDesc.Render(k.desc)
		parts = append(parts, fmt.Sprintf("%s %s", key, desc))
	}

	helpText := strings.Join(parts, "    ")
	b.WriteString(m.styles.HelpBar.Width(contentWidth).Render(helpText))

	return m.styles.App.Render(b.String())
}

// renderHelpBar renders the always-visible help bar.
func (m *Model) renderHelpBar() string {
	keys := []struct{ key, desc string }{
		{"h/l", "columns"},
		{"j/k", "tickets"},
		{"n", "new"},
		{"e", "edit"},
		{"d", "delete"},
		{"m", "move"},
		{"p", "copy ticket prompt"},
		{"P", "copy all todo prompts"},
		{"Enter", "view"},
		{"/", "search"},
		{"?", "help"},
		{"q", "quit"},
	}

	var parts []string
	for _, k := range keys {
		key := m.styles.HelpKey.Render(k.key)
		desc := m.styles.HelpDesc.Render(k.desc)
		parts = append(parts, fmt.Sprintf("%s %s", key, desc))
	}

	helpText := strings.Join(parts, "  ")
	return m.styles.HelpBar.Width(m.width - 4).Render(helpText)
}

// renderHelp renders the detailed help view.
func (m *Model) renderHelp() string {
	help := `
KANBAN TUI - Keyboard Shortcuts

Navigation
  h / ←      Move to left column
  l / →      Move to right column
  j / ↓      Move to next ticket
  k / ↑      Move to previous ticket

Actions
  n          Create new ticket
  e          Edit selected ticket (opens $EDITOR)
  d          Delete selected ticket
  m          Move ticket to another column
  Enter      View ticket details

Agent Integration
  p          Copy AI agent prompt for selected ticket to clipboard
  P          Copy AI agent prompt for all todo tickets to clipboard

Other
  /          Search tickets
  r          Refresh board
  ?          Toggle this help
  q          Quit

Press Esc or ? to close this help
`
	return m.styles.App.Render(help)
}
