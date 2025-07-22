package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmModel represents a confirmation dialog
type ConfirmModel struct {
	title       string
	message     string
	confirmText string
	cancelText  string
	confirmed   bool
	cancelled   bool
	focused     int // 0 = confirm, 1 = cancel
	width       int
	height      int
}

// NewConfirmModel creates a new confirmation dialog
func NewConfirmModel(title, message string) *ConfirmModel {
	return &ConfirmModel{
		title:       title,
		message:     message,
		confirmText: "Yes",
		cancelText:  "No",
		focused:     1, // Default to "No" for safety
	}
}

// SetButtons customizes the button text
func (m *ConfirmModel) SetButtons(confirm, cancel string) *ConfirmModel {
	m.confirmText = confirm
	m.cancelText = cancel
	return m
}

// Init initializes the confirmation model
func (m *ConfirmModel) Init() tea.Cmd {
	return nil
}

// Update handles input for the confirmation dialog
func (m *ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit

		case "tab", "left", "right":
			m.focused = (m.focused + 1) % 2

		case "enter", " ":
			if m.focused == 0 {
				m.confirmed = true
			} else {
				m.cancelled = true
			}
			return m, tea.Quit

		case "y", "Y":
			m.confirmed = true
			return m, tea.Quit

		case "n", "N":
			m.cancelled = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the confirmation dialog
func (m *ConfirmModel) View() string {
	if m.width == 0 {
		m.width = 60 // Default width
	}

	var sections []string

	// Title
	title := HeaderStyle.Render(fmt.Sprintf("%s %s", Icon("warning"), m.title))
	sections = append(sections, title)

	// Message
	messageStyle := lipgloss.NewStyle().
		Foreground(ColorDark).
		Padding(1, 0).
		Width(m.width - 4)
	
	message := messageStyle.Render(m.message)
	sections = append(sections, message)

	// Buttons
	buttons := m.renderButtons()
	sections = append(sections, buttons)

	// Instructions
	instructions := m.renderInstructions()
	sections = append(sections, instructions)

	content := strings.Join(sections, "\n")

	// Add border
	return BorderStyle.
		Width(m.width - 4).
		Render(content)
}

func (m *ConfirmModel) renderButtons() string {
	var confirmStyle, cancelStyle lipgloss.Style

	if m.focused == 0 {
		confirmStyle = ButtonSelectedStyle
		cancelStyle = ButtonStyle
	} else {
		confirmStyle = ButtonStyle
		cancelStyle = ButtonSelectedStyle
	}

	confirm := confirmStyle.Render(m.confirmText)
	cancel := cancelStyle.Render(m.cancelText)

	buttonRow := lipgloss.JoinHorizontal(lipgloss.Left, confirm, cancel)
	
	return lipgloss.NewStyle().
		Padding(1, 0).
		Render(buttonRow)
}

func (m *ConfirmModel) renderInstructions() string {
	instructions := []string{
		"Navigate: ← → Tab",
		"Select: Enter/Space",
		"Quick: Y/N",
		"Cancel: Esc",
	}

	return ProgressTextStyle.Render(strings.Join(instructions, " • "))
}

// IsConfirmed returns true if user confirmed
func (m *ConfirmModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled returns true if user cancelled
func (m *ConfirmModel) IsCancelled() bool {
	return m.cancelled
}

// RunConfirmation runs a confirmation dialog and returns the result
func RunConfirmation(title, message string) (bool, error) {
	model := NewConfirmModel(title, message)
	p := tea.NewProgram(model)
	
	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}

	confirmModel, ok := finalModel.(*ConfirmModel)
	if !ok {
		return false, fmt.Errorf("unexpected model type")
	}

	return confirmModel.IsConfirmed(), nil
}