package views

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/killallgit/dick/internal/config"
	"github.com/killallgit/dick/internal/styles"
	"github.com/killallgit/dick/internal/tui/components"
	"github.com/killallgit/dick/internal/tui/messages"
)

// Monitor represents the monitoring view (watch mode)
type Monitor struct {
	config     *config.Config
	width      int
	height     int
	lastUpdate time.Time
	err        error
	
	// File monitoring
	configPath string
	configStat os.FileInfo
	
	// Components
	header    *components.Header
	footer    *components.Footer
	progress  *components.ProgressBar
	eventLog  *components.EventLog
}

// NewMonitorView creates a new monitor view
func NewMonitorView(cfg *config.Config) View {
	configStat, _ := os.Stat(".dick.yaml")
	
	eventLog := components.NewEventLog(20, 5)
	eventLog.Add(fmt.Sprintf("Started monitoring at %s", time.Now().Format("15:04:05")))
	
	return &Monitor{
		config:     cfg,
		lastUpdate: time.Now(),
		configPath: ".dick.yaml",
		configStat: configStat,
		header:     components.NewHeader("Dick Cluster Monitor", "cluster"),
		footer:     components.NewFooter().SetActiveView(messages.MonitorView),
		eventLog:   eventLog,
	}
}

// Init initializes the monitor view
func (m *Monitor) Init() tea.Cmd {
	return nil
}

// Update handles messages for the monitor view
func (m *Monitor) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.header.SetWidth(msg.Width)
		return m, nil
		
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			// Refresh config and add event
			if cfg, err := config.LoadConfig(); err == nil {
				m.config = cfg
				m.eventLog.Add("Config manually refreshed")
			}
			return m, nil
		case "c":
			// Clear events
			m.eventLog.Clear()
			m.eventLog.Add("Events cleared")
			return m, nil
		}
		
	case messages.TickMsg:
		m.lastUpdate = msg.Time
		m.footer.UpdateTime(msg.Time)
		
		// Check for config file changes
		if stat, err := os.Stat(m.configPath); err == nil {
			if m.configStat != nil && stat.ModTime().After(m.configStat.ModTime()) {
				m.eventLog.Add("Config file changed, reloading...")
				if cfg, err := config.LoadConfig(); err == nil {
					m.config = cfg
				}
			}
			m.configStat = stat
		}
		
		// Update progress bar if active
		if m.config.Status == "active" {
			m.updateProgressBar()
		}
		
		return m, nil
		
	case messages.ConfigReloadMsg:
		if cfg, err := config.LoadConfig(); err == nil {
			m.config = cfg
			m.eventLog.Add("Config reloaded from external source")
		}
		return m, nil
		
	case messages.EventMsg:
		m.eventLog.Add(msg.Message)
		return m, nil
		
	case messages.ErrorMsg:
		m.err = msg.Err
		return m, nil
	}
	
	return m, nil
}

// View renders the monitor view
func (m *Monitor) View() string {
	if m.err != nil {
		return styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}
	
	var sections []string
	
	// Header
	sections = append(sections, m.header.Render())
	
	// Status section
	statusSection := m.renderStatusSection()
	sections = append(sections, statusSection)
	
	// Debug section
	debugSection := m.renderDebugSection()
	sections = append(sections, debugSection)
	
	// Events section
	sections = append(sections, m.eventLog.Render())
	
	// Footer with custom hint
	m.footer.SetCustomHint("'c' to clear events")
	sections = append(sections, m.footer.Render())
	
	content := strings.Join(sections, "\n")
	
	// Add border if we have enough space
	if m.width > 70 {
		return styles.BorderStyle.Width(m.width - 4).Render(content)
	}
	
	return content
}

func (m *Monitor) renderStatusSection() string {
	table := components.NewTable("Live Status")
	table.AddRow("Name", m.config.Name, "")
	table.AddRow("Provider", m.config.Provider, "")
	table.AddRow("Status", styles.FormatStatus(m.config.Status), "")
	
	statusLines := []string{table.Render()}
	
	// Status-specific info
	if m.config.Status == "active" {
		remaining := m.config.TimeRemaining()
		if remaining > 0 {
			table2 := components.NewTable("")
			table2.AddRow("Remaining", 
				styles.ProgressTextStyle.Render(remaining.Round(time.Second).String()), "")
			statusLines = append(statusLines, table2.Render())
			
			// Progress bar
			if m.progress != nil {
				statusLines = append(statusLines, "  "+m.progress.Render())
			}
		} else {
			statusLines = append(statusLines,
				styles.WarningStyle.Render(fmt.Sprintf("%s EXPIRED %s ago!", 
					styles.Icon("warning"), 
					(-remaining).Round(time.Second).String())),
			)
		}
	}
	
	return strings.Join(statusLines, "\n") + "\n"
}

func (m *Monitor) renderDebugSection() string {
	table := components.NewTable("Debug Information")
	
	// Config file info
	if m.configStat != nil {
		table.AddRow("Config file", m.configPath, "")
		table.AddRow("Last modified", 
			m.configStat.ModTime().Format("15:04:05"), "")
	}
	
	// Scheduled job info
	if m.config.ScheduledJobID != "" {
		table.AddRow("Job ID", m.config.ScheduledJobID, "")
	}
	
	// Timestamps
	if !m.config.CreatedAt.IsZero() {
		table.AddRow("Created", 
			m.config.CreatedAt.Format("15:04:05"), "")
	}
	if !m.config.ExpiresAt.IsZero() {
		table.AddRow("Expires", 
			m.config.ExpiresAt.Format("15:04:05"), "")
	}
	
	return table.Render() + "\n"
}

func (m *Monitor) updateProgressBar() {
	if m.config.Status != "active" {
		m.progress = nil
		return
	}
	
	totalDuration, err := m.config.ParseTTL()
	if err != nil {
		m.progress = nil
		return
	}
	
	remaining := m.config.TimeRemaining()
	elapsed := totalDuration - remaining
	
	if m.progress == nil {
		barWidth := 40
		if m.width > 0 && m.width < 80 {
			barWidth = m.width - 30
			if barWidth < 10 {
				barWidth = 10
			}
		}
		m.progress = components.NewProgressBar(elapsed, totalDuration, barWidth)
	} else {
		m.progress.Update(elapsed)
		if m.width > 0 && m.width < 80 {
			barWidth := m.width - 30
			if barWidth < 10 {
				barWidth = 10
			}
			m.progress.SetWidth(barWidth)
		}
	}
}