// Package models defines the data structures for the kanban application.
package models

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"gopkg.in/yaml.v3"
)

// Ticket represents a kanban ticket.
type Ticket struct {
	// Metadata from frontmatter
	Title   string    `yaml:"title"`
	Tags    []string  `yaml:"tags,omitempty"`
	Created time.Time `yaml:"created"`
	Updated time.Time `yaml:"updated"`

	// Content is the markdown body (excluding frontmatter)
	Content string `yaml:"-"`

	// FilePath is the full path to the ticket file
	FilePath string `yaml:"-"`

	// Column is the directory name of the column this ticket belongs to
	Column string `yaml:"-"`
}

// NewTicket creates a new ticket with default values.
func NewTicket(title, column string) *Ticket {
	now := time.Now()
	return &Ticket{
		Title:   title,
		Tags:    []string{},
		Created: now,
		Updated: now,
		Column:  column,
	}
}

// ParseTicket reads a markdown file and parses it into a Ticket.
func ParseTicket(path string) (*Ticket, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ticket, err := ParseTicketContent(data)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	ticket.FilePath = path
	ticket.Column = filepath.Base(filepath.Dir(path))

	return ticket, nil
}

// ParseTicketContent parses ticket content from bytes.
func ParseTicketContent(data []byte) (*Ticket, error) {
	ticket := &Ticket{}

	frontmatter, content, err := splitFrontmatter(data)
	if err != nil {
		return nil, err
	}

	if len(frontmatter) > 0 {
		if err := yaml.Unmarshal(frontmatter, ticket); err != nil {
			return nil, fmt.Errorf("parsing frontmatter: %w", err)
		}
	}

	ticket.Content = strings.TrimSpace(string(content))

	// Set defaults for missing values
	if ticket.Created.IsZero() {
		ticket.Created = time.Now()
	}
	if ticket.Updated.IsZero() {
		ticket.Updated = ticket.Created
	}

	return ticket, nil
}

// splitFrontmatter separates YAML frontmatter from markdown content.
func splitFrontmatter(data []byte) (frontmatter, content []byte, err error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var inFrontmatter bool
	var fmLines, contentLines []string

	for scanner.Scan() {
		line := scanner.Text()

		if !inFrontmatter && strings.TrimSpace(line) == "---" {
			inFrontmatter = true
			continue
		}

		if inFrontmatter {
			if strings.TrimSpace(line) == "---" {
				inFrontmatter = false
				continue
			}
			fmLines = append(fmLines, line)
		} else {
			contentLines = append(contentLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	frontmatter = []byte(strings.Join(fmLines, "\n"))
	content = []byte(strings.Join(contentLines, "\n"))

	return frontmatter, content, nil
}

// ToMarkdown converts the ticket to markdown format with frontmatter.
func (t *Ticket) ToMarkdown() []byte {
	var buf bytes.Buffer

	// Write frontmatter
	buf.WriteString("---\n")

	fm := struct {
		Title   string    `yaml:"title"`
		Tags    []string  `yaml:"tags,omitempty"`
		Created time.Time `yaml:"created"`
		Updated time.Time `yaml:"updated"`
	}{
		Title:   t.Title,
		Tags:    t.Tags,
		Created: t.Created,
		Updated: t.Updated,
	}

	fmData, _ := yaml.Marshal(fm)
	buf.Write(fmData)
	buf.WriteString("---\n\n")

	// Write content
	if t.Content != "" {
		buf.WriteString(t.Content)
		buf.WriteString("\n")
	}

	return buf.Bytes()
}

// Save writes the ticket to its file path.
func (t *Ticket) Save() error {
	if t.FilePath == "" {
		return fmt.Errorf("ticket has no file path")
	}

	t.Updated = time.Now()
	data := t.ToMarkdown()

	dir := filepath.Dir(t.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(t.FilePath, data, 0644)
}

// Delete removes the ticket file.
func (t *Ticket) Delete() error {
	if t.FilePath == "" {
		return fmt.Errorf("ticket has no file path")
	}
	return os.Remove(t.FilePath)
}

// GenerateFilename creates a filename for the ticket based on date and title.
func (t *Ticket) GenerateFilename() string {
	slug := slugify(t.Title)
	date := t.Created.Format("2006-01-02")
	return fmt.Sprintf("%s-%s.md", date, slug)
}

// slugify converts a string to a URL-friendly slug.
func slugify(s string) string {
	s = strings.ToLower(s)

	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")

	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			result.WriteRune(r)
		}
	}

	s = result.String()

	// Collapse multiple hyphens
	re := regexp.MustCompile(`-+`)
	s = re.ReplaceAllString(s, "-")

	// Trim hyphens from ends
	s = strings.Trim(s, "-")

	// Limit length
	if len(s) > 50 {
		s = s[:50]
		// Don't end with a hyphen
		s = strings.TrimSuffix(s, "-")
	}

	if s == "" {
		s = "untitled"
	}

	return s
}

// Move moves the ticket to a different column.
func (t *Ticket) Move(kanbanDir, newColumn string) error {
	if t.FilePath == "" {
		return fmt.Errorf("ticket has no file path")
	}

	oldPath := t.FilePath
	newDir := filepath.Join(kanbanDir, newColumn)
	newPath := filepath.Join(newDir, filepath.Base(t.FilePath))

	// Ensure target directory exists
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return err
	}

	// Move the file
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}

	t.FilePath = newPath
	t.Column = newColumn

	return nil
}

// ShortTitle returns a truncated title for display.
func (t *Ticket) ShortTitle(maxLen int) string {
	if len(t.Title) <= maxLen {
		return t.Title
	}
	return t.Title[:maxLen-3] + "..."
}
