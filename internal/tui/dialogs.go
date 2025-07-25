package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
	"github.com/killallgit/dick/internal/tui/messages"
	"github.com/killallgit/dick/internal/tui/views"
)

// ConfirmDialog is a specialized model for standalone confirmation dialogs
type ConfirmDialog struct {
	view      views.View
	confirmed bool
	done      bool
}

// NewConfirmDialog creates a new confirmation dialog model
func NewConfirmDialog(title, message string) *ConfirmDialog {
	opts := views.ConfirmOptions{
		Title:       title,
		Message:     message,
		ConfirmText: "Yes",
		CancelText:  "No",
	}
	
	return &ConfirmDialog{
		view: views.NewConfirmView(opts),
	}
}

// Init initializes the dialog
func (d *ConfirmDialog) Init() tea.Cmd {
	return d.view.Init()
}

// Update handles messages
func (d *ConfirmDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return d, tea.Quit
		}
		
	case messages.ConfirmResultMsg:
		d.confirmed = msg.Confirmed
		d.done = true
		return d, tea.Quit
	}
	
	updatedView, cmd := d.view.Update(msg)
	d.view = updatedView
	return d, cmd
}

// View renders the dialog
func (d *ConfirmDialog) View() string {
	return d.view.View()
}

// IsConfirmed returns whether the user confirmed
func (d *ConfirmDialog) IsConfirmed() bool {
	return d.confirmed
}

// RunConfirmDialog runs a standalone confirmation dialog
func RunConfirmDialog(title, message string) (bool, error) {
	model := NewConfirmDialog(title, message)
	p := tea.NewProgram(model)
	
	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}
	
	confirmModel, ok := finalModel.(*ConfirmDialog)
	if !ok {
		return false, fmt.Errorf("unexpected model type")
	}
	
	return confirmModel.IsConfirmed(), nil
}