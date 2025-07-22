package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/killallgit/dick/internal/config"
)

// WatchModel represents the watch monitoring TUI model
type WatchModel struct {
	config       *config.Config
	width        int
	height       int
	lastUpdate   time.Time
	err          error
	refreshRate  time.Duration
	events       []string
	configPath   string
	configStat   os.FileInfo
}

// NewWatchModel creates a new watch TUI model
func NewWatchModel(cfg *config.Config) *WatchModel {
	configStat, _ := os.Stat(".dick.yaml")
	
	return &WatchModel{
		config:      cfg,
		refreshRate: time.Second,
		lastUpdate:  time.Now(),
		events:      []string{fmt.Sprintf("Started monitoring at %s", time.Now().Format("15:04:05"))},
		configPath:  ".dick.yaml",
		configStat:  configStat,
	}
}

// Init initializes the model
func (m *WatchModel) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.tick(),
	)
}

// Update handles messages
func (m *WatchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			// Refresh config and add event
			if cfg, err := config.LoadConfig(); err == nil {
				m.config = cfg
				m.addEvent("Config manually refreshed")
			}
			return m, nil
		case "c":
			// Clear events
			m.events = []string{"Events cleared"}
			return m, nil
		}

	case tickMsg:
		m.lastUpdate = time.Now()
		
		// Check for config file changes
		if stat, err := os.Stat(m.configPath); err == nil {
			if m.configStat != nil && stat.ModTime().After(m.configStat.ModTime()) {
				m.addEvent("Config file changed, reloading...")
				if cfg, err := config.LoadConfig(); err == nil {
					m.config = cfg
				}
			}
			m.configStat = stat
		}
		
		return m, m.tick()

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

// View renders the watch TUI
func (m *WatchModel) View() string {
	if m.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	var sections []string

	// Header
	header := HeaderStyle.Render(fmt.Sprintf("%s Dick Cluster Monitor", Icon("cluster")))
	sections = append(sections, header)
	
	// Status section
	statusSection := m.renderStatusSection()
	sections = append(sections, statusSection)

	// Debug section  
	debugSection := m.renderDebugSection()
	sections = append(sections, debugSection)

	// Events section
	eventsSection := m.renderEventsSection()
	sections = append(sections, eventsSection)

	// Controls
	controlsSection := m.renderControlsSection()
	sections = append(sections, controlsSection)

	content := strings.Join(sections, "\n")
	
	// Add border if we have enough space
	if m.width > 70 {
		return BorderStyle.Width(m.width - 4).Render(content)
	}
	
	return content
}

func (m *WatchModel) renderStatusSection() string {
	lines := []string{
		TitleStyle.Render("📊 Live Status"),
	}

	// Basic info
	lines = append(lines,
		fmt.Sprintf("%s %s", InfoLabelStyle.Render("Name:"), InfoValueStyle.Render(m.config.Name)),
		fmt.Sprintf("%s %s", InfoLabelStyle.Render("Provider:"), InfoValueStyle.Render(m.config.Provider)),
		fmt.Sprintf("%s %s", InfoLabelStyle.Render("Status:"), FormatStatus(m.config.Status)),
	)

	// Status-specific info
	if m.config.Status == "active" {
		remaining := m.config.TimeRemaining()
		if remaining > 0 {
			lines = append(lines,
				fmt.Sprintf("%s %s", InfoLabelStyle.Render("Remaining:"), 
					ProgressTextStyle.Render(remaining.Round(time.Second).String())),
			)
			
			// Progress bar
			progressBar := m.renderProgressBar(remaining)
			if progressBar != "" {
				lines = append(lines, progressBar)
			}
		} else {
			lines = append(lines,
				WarningStyle.Render(fmt.Sprintf("%s EXPIRED %s ago!", 
					Icon("warning"), 
					(-remaining).Round(time.Second).String())),
			)
		}
	}

	return strings.Join(lines, "\n") + "\n"
}

func (m *WatchModel) renderDebugSection() string {
	lines := []string{
		TitleStyle.Render("🔧 Debug Information"),
	}

	// Config file info
	if m.configStat != nil {
		lines = append(lines,
			fmt.Sprintf("%s %s", InfoLabelStyle.Render("Config file:"), InfoValueStyle.Render(m.configPath)),
			fmt.Sprintf("%s %s", InfoLabelStyle.Render("Last modified:"), 
				InfoValueStyle.Render(m.configStat.ModTime().Format("15:04:05"))),
		)
	}

	// Scheduled job info
	if m.config.ScheduledJobID != "" {
		lines = append(lines,
			fmt.Sprintf("%s %s", InfoLabelStyle.Render("Job ID:"), InfoValueStyle.Render(m.config.ScheduledJobID)),
		)
	}

	// Timestamps
	if !m.config.CreatedAt.IsZero() {
		lines = append(lines,
			fmt.Sprintf("%s %s", InfoLabelStyle.Render("Created:"), 
				InfoValueStyle.Render(m.config.CreatedAt.Format("15:04:05"))),
		)
	}
	if !m.config.ExpiresAt.IsZero() {
		lines = append(lines,
			fmt.Sprintf("%s %s", InfoLabelStyle.Render("Expires:"), 
				InfoValueStyle.Render(m.config.ExpiresAt.Format("15:04:05"))),
		)
	}

	return strings.Join(lines, "\n") + "\n"
}

func (m *WatchModel) renderEventsSection() string {
	lines := []string{
		TitleStyle.Render("📋 Events"),
	}

	// Show last 5 events
	eventCount := len(m.events)
	startIdx := 0
	if eventCount > 5 {
		startIdx = eventCount - 5
	}

	for i := startIdx; i < eventCount; i++ {
		lines = append(lines, fmt.Sprintf("  • %s", m.events[i]))
	}

	if len(lines) == 1 {
		lines = append(lines, "  No events yet")
	}

	return strings.Join(lines, "\n") + "\n"
}

func (m *WatchModel) renderControlsSection() string {
	controls := []string{
		InfoLabelStyle.Render("Controls:"),
		"• 'r' refresh • 'c' clear events • 'q' quit",
		"",
		ProgressTextStyle.Render(fmt.Sprintf("Last update: %s", m.lastUpdate.Format("15:04:05"))),
	}
	
	return strings.Join(controls, "\n")
}

func (m *WatchModel) renderProgressBar(remaining time.Duration) string {
	if m.config.Status != "active" {
		return ""
	}

	totalDuration, err := m.config.ParseTTL()
	if err != nil {
		return ""
	}

	elapsed := totalDuration - remaining
	progress := float64(elapsed) / float64(totalDuration)
	
	barWidth := 40
	if m.width > 0 && m.width < 80 {
		barWidth = m.width - 30
		if barWidth < 10 {
			barWidth = 10
		}
	}
	
	filled := int(progress * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	percentage := int(progress * 100)
	if percentage > 100 {
		percentage = 100
	}
	
	return fmt.Sprintf("  %s %d%%", 
		ProgressBarStyle.Render(bar),
		percentage)
}

func (m *WatchModel) addEvent(message string) {
	timestamp := time.Now().Format("15:04:05")
	event := fmt.Sprintf("[%s] %s", timestamp, message)
	m.events = append(m.events, event)
	
	// Keep only last 20 events
	if len(m.events) > 20 {
		m.events = m.events[1:]
	}
}

// tick returns a command that triggers a refresh
func (m *WatchModel) tick() tea.Cmd {
	return tea.Tick(m.refreshRate, func(t time.Time) tea.Msg {
		return tickMsg{t}
	})
}