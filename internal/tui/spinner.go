package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbletea"
)

// Simple spinner characters
var spinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// SpinnerModel represents a loading spinner
type SpinnerModel struct {
	frame   int
	message string
	done    bool
	err     error
}

// NewSpinner creates a new spinner
func NewSpinner(message string) *SpinnerModel {
	return &SpinnerModel{
		message: message,
	}
}

// Init initializes the spinner
func (m *SpinnerModel) Init() tea.Cmd {
	return m.tick()
}

// Update handles spinner updates
func (m *SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case spinnerTickMsg:
		if !m.done {
			m.frame = (m.frame + 1) % len(spinnerChars)
			return m, m.tick()
		}

	case spinnerDoneMsg:
		m.done = true
		return m, tea.Quit

	case spinnerErrorMsg:
		m.err = msg.err
		m.done = true
		return m, tea.Quit
	}

	return m, nil
}

// View renders the spinner
func (m *SpinnerModel) View() string {
	if m.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("%s Error: %v", Icon("error"), m.err))
	}

	if m.done {
		return SuccessStyle.Render(fmt.Sprintf("%s %s", Icon("success"), m.message))
	}

	spinner := ProgressBarStyle.Render(spinnerChars[m.frame])
	return fmt.Sprintf("%s %s", spinner, InfoLabelStyle.Render(m.message))
}

// tick returns a command that triggers spinner animation
func (m *SpinnerModel) tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

// Complete marks the spinner as done
func (m *SpinnerModel) Complete() tea.Cmd {
	return func() tea.Msg {
		return spinnerDoneMsg{}
	}
}

// Error marks the spinner as failed
func (m *SpinnerModel) Error(err error) tea.Cmd {
	return func() tea.Msg {
		return spinnerErrorMsg{err: err}
	}
}

// Message types for spinner
type spinnerTickMsg struct{}

type spinnerDoneMsg struct{}

type spinnerErrorMsg struct {
	err error
}

// ShowSpinner displays a simple inline spinner for operations
func ShowSpinner(message string) string {
	return fmt.Sprintf("%s %s", 
		ProgressBarStyle.Render("⠋"), 
		InfoLabelStyle.Render(message))
}