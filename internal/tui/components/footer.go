package components

import (
	"fmt"
	"strings"
	"time"
	
	"github.com/killallgit/dick/internal/styles"
	"github.com/killallgit/dick/internal/tui/messages"
)

// Footer represents a reusable footer component
type Footer struct {
	ActiveView messages.ViewType
	LastUpdate time.Time
	ShowNav    bool
	CustomHint string
}

// NewFooter creates a new footer
func NewFooter() *Footer {
	return &Footer{
		ShowNav:    true,
		LastUpdate: time.Now(),
	}
}

// SetActiveView updates the active view
func (f *Footer) SetActiveView(view messages.ViewType) *Footer {
	f.ActiveView = view
	return f
}

// SetCustomHint sets a custom hint message
func (f *Footer) SetCustomHint(hint string) *Footer {
	f.CustomHint = hint
	return f
}

// UpdateTime updates the last update time
func (f *Footer) UpdateTime(t time.Time) {
	f.LastUpdate = t
}

// Render returns the rendered footer
func (f *Footer) Render() string {
	lines := []string{}
	
	// Navigation hints
	if f.ShowNav {
		navHints := f.getNavigationHints()
		lines = append(lines, styles.InfoLabelStyle.Render("Navigation:"))
		lines = append(lines, navHints)
	}
	
	// Custom hint
	if f.CustomHint != "" {
		lines = append(lines, styles.InfoValueStyle.Render(f.CustomHint))
	}
	
	// Last update time
	lastUpdate := fmt.Sprintf("Last updated: %s", f.LastUpdate.Format("15:04:05"))
	lines = append(lines, "", styles.ProgressTextStyle.Render(lastUpdate))
	
	return strings.Join(lines, "\n")
}

func (f *Footer) getNavigationHints() string {
	hints := []string{}
	
	// View-specific hints
	switch f.ActiveView {
	case messages.StatusView:
		hints = append(hints, "2:monitor", "?:help", "r:refresh", "q:quit")
	case messages.MonitorView:
		hints = append(hints, "1:status", "c:clear", "r:refresh", "q:quit")
	case messages.ConfirmView:
		hints = append(hints, "y/n:choose", "←→:navigate", "enter:select", "esc:cancel")
	case messages.HelpView:
		hints = append(hints, "esc:back", "q:quit")
	default:
		hints = append(hints, "tab:next", "?:help", "q:quit")
	}
	
	return strings.Join(hints, " • ")
}