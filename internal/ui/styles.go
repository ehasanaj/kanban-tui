// Package ui provides the terminal user interface components.
package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Gruvbox color palette (dark mode)
var (
	// Background tones
	GruvboxBg0Hard = lipgloss.Color("#1d2021")
	GruvboxBg0     = lipgloss.Color("#282828")
	GruvboxBg1     = lipgloss.Color("#3c3836")
	GruvboxBg2     = lipgloss.Color("#504945")
	GruvboxBg3     = lipgloss.Color("#665c54")
	GruvboxBg4     = lipgloss.Color("#7c6f64")

	// Foreground tones
	GruvboxFg0 = lipgloss.Color("#fbf1c7")
	GruvboxFg1 = lipgloss.Color("#ebdbb2")
	GruvboxFg2 = lipgloss.Color("#d5c4a1")
	GruvboxFg3 = lipgloss.Color("#bdae93")
	GruvboxFg4 = lipgloss.Color("#a89984")

	// Accent colors
	GruvboxRed    = lipgloss.Color("#fb4934")
	GruvboxGreen  = lipgloss.Color("#b8bb26")
	GruvboxYellow = lipgloss.Color("#fabd2f")
	GruvboxBlue   = lipgloss.Color("#83a598")
	GruvboxPurple = lipgloss.Color("#d3869b")
	GruvboxAqua   = lipgloss.Color("#8ec07c")
	GruvboxOrange = lipgloss.Color("#fe8019")
	GruvboxGray   = lipgloss.Color("#928374")
)

// Semantic color aliases
var (
	ColorPrimary   = GruvboxOrange
	ColorSecondary = GruvboxBlue
	ColorSuccess   = GruvboxGreen
	ColorWarning   = GruvboxYellow
	ColorDanger    = GruvboxRed
	ColorMuted     = GruvboxGray
	ColorBorder    = GruvboxBg3
	ColorBg        = GruvboxBg0
	ColorBgLight   = GruvboxBg1
	ColorFg        = GruvboxFg1
	ColorFgDim     = GruvboxFg4
)

// Styles contains all the lipgloss styles for the application.
type Styles struct {
	App            lipgloss.Style
	Header         lipgloss.Style
	HeaderTitle    lipgloss.Style
	Column         lipgloss.Style
	ColumnActive   lipgloss.Style
	ColumnHeader   lipgloss.Style
	ColumnCount    lipgloss.Style
	Ticket         lipgloss.Style
	TicketSelected lipgloss.Style
	TicketTitle    lipgloss.Style
	TicketTags     lipgloss.Style
	TicketDate     lipgloss.Style
	HelpBar        lipgloss.Style
	HelpKey        lipgloss.Style
	HelpDesc       lipgloss.Style
	StatusBar      lipgloss.Style
	StatusMessage  lipgloss.Style
	Modal          lipgloss.Style
	ModalTitle     lipgloss.Style
	Input          lipgloss.Style
	InputFocused   lipgloss.Style
	Button         lipgloss.Style
	ButtonActive   lipgloss.Style
}

// DefaultStyles creates the default style set.
func DefaultStyles() Styles {
	return Styles{
		App: lipgloss.NewStyle().
			Padding(1, 2),

		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(GruvboxBg0).
			Background(GruvboxOrange).
			Padding(0, 2).
			MarginBottom(1),

		HeaderTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(GruvboxFg0),

		Column: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(GruvboxBg3).
			Padding(0, 1).
			MarginRight(1),

		ColumnActive: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(GruvboxOrange).
			Padding(0, 1).
			MarginRight(1),

		ColumnHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(GruvboxBg0).
			MarginBottom(1).
			Padding(0, 1),

		ColumnCount: lipgloss.NewStyle().
			Foreground(GruvboxFg4).
			MarginLeft(1),

		Ticket: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(GruvboxBg3).
			Padding(0, 1).
			MarginBottom(1),

		TicketSelected: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(GruvboxYellow).
			Background(GruvboxBg1).
			Padding(0, 1).
			MarginBottom(1),

		TicketTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(GruvboxFg1),

		TicketTags: lipgloss.NewStyle().
			Foreground(GruvboxPurple).
			Italic(true),

		TicketDate: lipgloss.NewStyle().
			Foreground(GruvboxGray),

		HelpBar: lipgloss.NewStyle().
			Foreground(GruvboxFg3).
			Background(GruvboxBg1).
			Padding(0, 1),

		HelpKey: lipgloss.NewStyle().
			Foreground(GruvboxYellow).
			Bold(true),

		HelpDesc: lipgloss.NewStyle().
			Foreground(GruvboxFg4),

		StatusBar: lipgloss.NewStyle().
			Foreground(GruvboxBg0).
			Background(GruvboxOrange).
			Padding(0, 1),

		StatusMessage: lipgloss.NewStyle().
			Foreground(GruvboxGreen),

		Modal: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(GruvboxOrange).
			Padding(1, 2).
			Background(GruvboxBg0),

		ModalTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(GruvboxOrange).
			MarginBottom(1),

		Input: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(GruvboxBg3).
			Padding(0, 1),

		InputFocused: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(GruvboxYellow).
			Padding(0, 1),

		Button: lipgloss.NewStyle().
			Foreground(GruvboxFg1).
			Background(GruvboxBg2).
			Padding(0, 2).
			MarginRight(1),

		ButtonActive: lipgloss.NewStyle().
			Foreground(GruvboxBg0).
			Background(GruvboxOrange).
			Padding(0, 2).
			MarginRight(1),
	}
}

// ColumnColors returns a map of column colors by column directory name.
func ColumnColors() map[string]lipgloss.Color {
	return map[string]lipgloss.Color{
		"todo":    GruvboxRed,
		"doing":   GruvboxYellow,
		"done":    GruvboxGreen,
		"backlog": GruvboxGray,
		"review":  GruvboxBlue,
	}
}

// GetColumnColor returns the color for a column, with a default fallback.
func GetColumnColor(colDir string) lipgloss.Color {
	colors := ColumnColors()
	if color, ok := colors[colDir]; ok {
		return color
	}
	return GruvboxAqua
}
