package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/killallgit/dick/internal/config"
)

// StatusModel represents the status TUI model
type StatusModel struct {
	config      *config.Config
	width       int
	height      int
	lastUpdate  time.Time
	err         error
	refreshRate time.Duration
}

// NewStatusModel creates a new status TUI model
func NewStatusModel(cfg *config.Config) *StatusModel {
	return &StatusModel{
		config:      cfg,
		refreshRate: time.Second,
		lastUpdate:  time.Now(),
	}
}

// Init initializes the model
func (m *StatusModel) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.tick(),
	)
}

// Update handles messages
func (m *StatusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "r":
			// Refresh config
			if cfg, err := config.LoadConfig(); err == nil {
				m.config = cfg
				m.lastUpdate = time.Now()
			}
			return m, nil
		}

	case tickMsg:
		m.lastUpdate = time.Now()
		return m, m.tick()

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

// View renders the TUI
func (m *StatusModel) View() string {
	if m.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	var sections []string

	// Header
	header := HeaderStyle.Render(fmt.Sprintf("%s Dick Cluster Status", Icon("cluster")))
	sections = append(sections, header)
	
	// Divider
	dividerWidth := 50
	if m.width > 0 && m.width < dividerWidth {
		dividerWidth = m.width - 4
	}
	sections = append(sections, Divider(dividerWidth))

	// Project info section
	projectSection := m.renderProjectInfo()
	sections = append(sections, projectSection)

	// Status section
	statusSection := m.renderStatusInfo()
	sections = append(sections, statusSection)

	// Controls
	controlsSection := m.renderControls()
	sections = append(sections, controlsSection)

	content := strings.Join(sections, "\n")
	
	// Add border if we have enough space
	if m.width > 60 {
		return BorderStyle.Width(m.width - 4).Render(content)
	}
	
	return content
}

func (m *StatusModel) renderProjectInfo() string {
	lines := []string{
		TitleStyle.Render("Project Information"),
	}

	projectPath := m.config.ProjectPath
	if projectPath == "" {
		projectPath = "Current directory"
	}

	lines = append(lines,
		fmt.Sprintf("%s %s %s", 
			Icon("project"), 
			InfoLabelStyle.Render("Project:"), 
			InfoValueStyle.Render(projectPath)),
		fmt.Sprintf("%s %s %s", 
			Icon("name"), 
			InfoLabelStyle.Render("Name:"), 
			InfoValueStyle.Render(m.config.Name)),
		fmt.Sprintf("%s %s %s", 
			Icon("cluster"), 
			InfoLabelStyle.Render("Provider:"), 
			InfoValueStyle.Render(m.config.Provider)),
		fmt.Sprintf("%s %s %s", 
			Icon("ttl"), 
			InfoLabelStyle.Render("TTL:"), 
			InfoValueStyle.Render(m.config.TTL)),
	)

	return strings.Join(lines, "\n") + "\n"
}

func (m *StatusModel) renderStatusInfo() string {
	lines := []string{
		TitleStyle.Render("Status Information"),
	}

	// Status with icon
	statusIcon := Icon(m.config.Status)
	if m.config.Status == "" {
		statusIcon = Icon("unknown")
	}
	
	lines = append(lines,
		fmt.Sprintf("%s %s %s", 
			statusIcon,
			InfoLabelStyle.Render("Status:"), 
			FormatStatus(m.config.Status)),
	)

	// Status-specific information
	switch m.config.Status {
	case "active":
		remaining := m.config.TimeRemaining()
		if remaining > 0 {
			lines = append(lines,
				fmt.Sprintf("%s %s %s", 
					Icon("created"), 
					InfoLabelStyle.Render("Created:"), 
					InfoValueStyle.Render(m.config.CreatedAt.Format("2006-01-02 15:04:05"))),
				fmt.Sprintf("%s %s %s", 
					Icon("expires"), 
					InfoLabelStyle.Render("Expires:"), 
					InfoValueStyle.Render(m.config.ExpiresAt.Format("2006-01-02 15:04:05"))),
				fmt.Sprintf("%s %s %s", 
					Icon("remaining"), 
					InfoLabelStyle.Render("Remaining:"), 
					ProgressTextStyle.Render(remaining.String())),
			)
			
			// Add progress bar
			progressBar := m.renderProgressBar(remaining)
			lines = append(lines, progressBar)
		} else {
			lines = append(lines,
				WarningStyle.Render(fmt.Sprintf("%s Should have been destroyed at: %s", 
					Icon("warning"),
					m.config.ExpiresAt.Format("2006-01-02 15:04:05"))),
			)
		}

	case "destroyed":
		if !m.config.CreatedAt.IsZero() {
			lines = append(lines,
				fmt.Sprintf("%s %s %s", 
					Icon("created"), 
					InfoLabelStyle.Render("Was created:"), 
					InfoValueStyle.Render(m.config.CreatedAt.Format("2006-01-02 15:04:05"))),
			)
		}

	default:
		lines = append(lines,
			InfoLabelStyle.Render("ℹ️  Run 'dick new' to create a cluster"),
		)
	}

	return strings.Join(lines, "\n") + "\n"
}

func (m *StatusModel) renderProgressBar(remaining time.Duration) string {
	if m.config.Status != "active" {
		return ""
	}

	totalDuration, err := m.config.ParseTTL()
	if err != nil {
		return ""
	}

	elapsed := totalDuration - remaining
	progress := float64(elapsed) / float64(totalDuration)
	
	barWidth := 30
	if m.width > 0 && m.width < 60 {
		barWidth = m.width - 20
	}
	
	filled := int(progress * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	percentage := int(progress * 100)
	
	return fmt.Sprintf("    %s %s %d%%", 
		ProgressBarStyle.Render(bar),
		ProgressTextStyle.Render("TTL Progress:"),
		percentage)
}

func (m *StatusModel) renderControls() string {
	controls := []string{
		"Controls:",
		"• Press 'r' to refresh",
		"• Press 'q' or Ctrl+C to quit",
	}
	
	lastUpdate := fmt.Sprintf("Last updated: %s", 
		m.lastUpdate.Format("15:04:05"))
	
	controls = append(controls, "", 
		ProgressTextStyle.Render(lastUpdate))
	
	return InfoLabelStyle.Render(strings.Join(controls, "\n"))
}

// tick returns a command that triggers a refresh
func (m *StatusModel) tick() tea.Cmd {
	return tea.Tick(m.refreshRate, func(t time.Time) tea.Msg {
		return tickMsg{t}
	})
}

// Message types
type tickMsg struct {
	time time.Time
}

type errMsg struct {
	err error
}

// WarningStyle for warnings
var WarningStyle = lipgloss.NewStyle().
	Foreground(ColorWarning).
	Bold(true)