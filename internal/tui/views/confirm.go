package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/killallgit/dick/internal/styles"
	"github.com/killallgit/dick/internal/tui/components"
	"github.com/killallgit/dick/internal/tui/messages"
)

// Confirm represents the confirmation dialog view
type Confirm struct {
	title       string
	message     string
	confirmText string
	cancelText  string
	confirmed   bool
	cancelled   bool
	focused     int // 0 = confirm, 1 = cancel
	width       int
	height      int
	
	// Components
	footer *components.Footer
}

// ConfirmOptions holds options for creating a confirm view
type ConfirmOptions struct {
	Title       string
	Message     string
	ConfirmText string
	CancelText  string
}

// NewConfirmView creates a new confirmation view
func NewConfirmView(opts ConfirmOptions) View {
	confirmText := opts.ConfirmText
	if confirmText == "" {
		confirmText = "Yes"
	}
	
	cancelText := opts.CancelText
	if cancelText == "" {
		cancelText = "No"
	}
	
	return &Confirm{
		title:       opts.Title,
		message:     opts.Message,
		confirmText: confirmText,
		cancelText:  cancelText,
		focused:     1, // Default to "No" for safety
		footer:      components.NewFooter().SetActiveView(messages.ConfirmView),
	}
}

// Init initializes the confirm view
func (c *Confirm) Init() tea.Cmd {
	return nil
}

// Update handles messages for the confirm view
func (c *Confirm) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
		return c, nil
		
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			c.cancelled = true
			return c, func() tea.Msg {
				return messages.ConfirmResultMsg{Confirmed: false}
			}
			
		case "tab", "left", "right":
			c.focused = (c.focused + 1) % 2
			
		case "enter", " ":
			if c.focused == 0 {
				c.confirmed = true
				return c, func() tea.Msg {
					return messages.ConfirmResultMsg{Confirmed: true}
				}
			} else {
				c.cancelled = true
				return c, func() tea.Msg {
					return messages.ConfirmResultMsg{Confirmed: false}
				}
			}
			
		case "y", "Y":
			c.confirmed = true
			return c, func() tea.Msg {
				return messages.ConfirmResultMsg{Confirmed: true}
			}
			
		case "n", "N":
			c.cancelled = true
			return c, func() tea.Msg {
				return messages.ConfirmResultMsg{Confirmed: false}
			}
		}
	}
	
	return c, nil
}

// View renders the confirmation dialog
func (c *Confirm) View() string {
	if c.width == 0 {
		c.width = 60 // Default width
	}
	
	var sections []string
	
	// Title
	title := styles.HeaderStyle.Render(fmt.Sprintf("%s %s", styles.Icon("warning"), c.title))
	sections = append(sections, title)
	
	// Message
	messageStyle := lipgloss.NewStyle().
		Foreground(styles.ColorDark).
		Padding(1, 0).
		Width(c.width - 4)
	
	message := messageStyle.Render(c.message)
	sections = append(sections, message)
	
	// Buttons
	buttons := c.renderButtons()
	sections = append(sections, buttons)
	
	// Instructions
	instructions := c.renderInstructions()
	sections = append(sections, instructions)
	
	content := strings.Join(sections, "\n")
	
	// Add border
	return styles.BorderStyle.
		Width(c.width - 4).
		Render(content)
}

func (c *Confirm) renderButtons() string {
	var confirmStyle, cancelStyle lipgloss.Style
	
	if c.focused == 0 {
		confirmStyle = styles.ButtonSelectedStyle
		cancelStyle = styles.ButtonStyle
	} else {
		confirmStyle = styles.ButtonStyle
		cancelStyle = styles.ButtonSelectedStyle
	}
	
	confirm := confirmStyle.Render(c.confirmText)
	cancel := cancelStyle.Render(c.cancelText)
	
	buttonRow := lipgloss.JoinHorizontal(lipgloss.Left, confirm, "  ", cancel)
	
	return lipgloss.NewStyle().
		Padding(1, 0).
		Render(buttonRow)
}

func (c *Confirm) renderInstructions() string {
	instructions := []string{
		"Navigate: ← → Tab",
		"Select: Enter/Space",
		"Quick: Y/N",
		"Cancel: Esc",
	}
	
	return styles.ProgressTextStyle.Render(strings.Join(instructions, " • "))
}

// IsConfirmed returns true if user confirmed
func (c *Confirm) IsConfirmed() bool {
	return c.confirmed
}

// IsCancelled returns true if user cancelled
func (c *Confirm) IsCancelled() bool {
	return c.cancelled
}