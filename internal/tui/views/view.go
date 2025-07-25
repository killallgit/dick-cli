package views

import (
	"github.com/charmbracelet/bubbletea"
)

// View is the interface that all views must implement
type View interface {
	// Update handles messages and returns the updated view and any commands
	Update(msg tea.Msg) (View, tea.Cmd)
	
	// View renders the view to a string
	View() string
	
	// Init initializes the view and returns any initial commands
	Init() tea.Cmd
}

// ViewContext provides shared context to all views
type ViewContext struct {
	Width  int
	Height int
}