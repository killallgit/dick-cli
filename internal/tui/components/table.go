package components

import (
	"fmt"
	"strings"
	
	"github.com/killallgit/dick/internal/styles"
)

// TableRow represents a row in the table
type TableRow struct {
	Label string
	Value string
	Icon  string
}

// Table represents a reusable table component
type Table struct {
	Title string
	Rows  []TableRow
	Width int
}

// NewTable creates a new table
func NewTable(title string) *Table {
	return &Table{
		Title: title,
		Rows:  []TableRow{},
		Width: 0,
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(label, value, icon string) *Table {
	t.Rows = append(t.Rows, TableRow{
		Label: label,
		Value: value,
		Icon:  icon,
	})
	return t
}

// Clear clears all rows
func (t *Table) Clear() {
	t.Rows = []TableRow{}
}

// Render returns the rendered table
func (t *Table) Render() string {
	lines := []string{}
	
	if t.Title != "" {
		lines = append(lines, styles.TitleStyle.Render(t.Title))
	}
	
	// Find the longest label for alignment
	maxLabelLen := 0
	for _, row := range t.Rows {
		labelLen := len(row.Label)
		if row.Icon != "" {
			labelLen += len(row.Icon) + 1
		}
		if labelLen > maxLabelLen {
			maxLabelLen = labelLen
		}
	}
	
	// Render rows
	for _, row := range t.Rows {
		var line string
		
		if row.Icon != "" {
			label := fmt.Sprintf("%s %s", row.Icon, row.Label)
			line = fmt.Sprintf("%-*s %s", 
				maxLabelLen, 
				styles.InfoLabelStyle.Render(label+":"),
				styles.InfoValueStyle.Render(row.Value))
		} else {
			line = fmt.Sprintf("%-*s %s",
				maxLabelLen,
				styles.InfoLabelStyle.Render(row.Label+":"),
				styles.InfoValueStyle.Render(row.Value))
		}
		
		lines = append(lines, line)
	}
	
	return strings.Join(lines, "\n")
}

// SetWidth sets the table width
func (t *Table) SetWidth(width int) {
	t.Width = width
}