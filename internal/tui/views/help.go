package views

import (
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/killallgit/dick/internal/styles"
	"github.com/killallgit/dick/internal/tui/components"
	"github.com/killallgit/dick/internal/tui/messages"
)

// Help represents the help view
type Help struct {
	width  int
	height int
	
	// Components
	header *components.Header
	footer *components.Footer
}

// NewHelpView creates a new help view
func NewHelpView() View {
	return &Help{
		header: components.NewHeader("Dick Help", "info"),
		footer: components.NewFooter().SetActiveView(messages.HelpView),
	}
}

// Init initializes the help view
func (h *Help) Init() tea.Cmd {
	return nil
}

// Update handles messages for the help view
func (h *Help) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height
		h.header.SetWidth(msg.Width)
		return h, nil
	}
	
	return h, nil
}

// View renders the help view
func (h *Help) View() string {
	var sections []string
	
	// Header
	sections = append(sections, h.header.Render())
	
	// Navigation help
	sections = append(sections, h.renderNavigationHelp())
	
	// View-specific help
	sections = append(sections, h.renderViewHelp())
	
	// Command help
	sections = append(sections, h.renderCommandHelp())
	
	// Footer
	sections = append(sections, h.footer.Render())
	
	content := strings.Join(sections, "\n")
	
	// Add border if we have enough space
	if h.width > 60 {
		return styles.BorderStyle.Width(h.width - 4).Render(content)
	}
	
	return content
}

func (h *Help) renderNavigationHelp() string {
	lines := []string{
		styles.TitleStyle.Render("Navigation"),
		"",
		"  1        - Switch to Status view",
		"  2        - Switch to Monitor view", 
		"  3        - Switch to Settings view (future)",
		"  Tab      - Cycle through views forward",
		"  Shift+Tab - Cycle through views backward",
		"  ?        - Toggle this help",
		"  ESC      - Go back to previous view",
		"  q        - Quit Dick CLI",
		"  Ctrl+C   - Force quit",
	}
	
	return strings.Join(lines, "\n") + "\n"
}

func (h *Help) renderViewHelp() string {
	lines := []string{
		styles.TitleStyle.Render("View-Specific Controls"),
		"",
		styles.InfoLabelStyle.Render("Status View:"),
		"  r        - Refresh cluster status",
		"",
		styles.InfoLabelStyle.Render("Monitor View:"),
		"  r        - Refresh cluster status",
		"  c        - Clear event log",
		"",
		styles.InfoLabelStyle.Render("Confirmation Dialogs:"),
		"  y/n      - Quick yes/no selection",
		"  Enter    - Confirm selection",
		"  ←/→      - Navigate between options",
	}
	
	return strings.Join(lines, "\n") + "\n"
}

func (h *Help) renderCommandHelp() string {
	lines := []string{
		styles.TitleStyle.Render("Command Line Usage"),
		"",
		"  dick status         - Show cluster status",
		"  dick status --watch - Monitor cluster in real-time",
		"  dick new k8s        - Create new Kubernetes cluster",
		"  dick destroy        - Destroy active cluster",
		"",
		styles.InfoLabelStyle.Render("Common Flags:"),
		"  --verbose, -v       - Show detailed output",
		"  --help, -h          - Show command help",
	}
	
	return strings.Join(lines, "\n") + "\n"
}