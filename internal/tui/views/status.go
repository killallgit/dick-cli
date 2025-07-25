package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/killallgit/dick/internal/config"
	"github.com/killallgit/dick/internal/styles"
	"github.com/killallgit/dick/internal/tui/components"
	"github.com/killallgit/dick/internal/tui/messages"
)

// Status represents the status view
type Status struct {
	config     *config.Config
	width      int
	height     int
	lastUpdate time.Time
	err        error
	
	// Components
	header   *components.Header
	footer   *components.Footer
	progress *components.ProgressBar
}

// NewStatusView creates a new status view
func NewStatusView(cfg *config.Config) View {
	return &Status{
		config:     cfg,
		lastUpdate: time.Now(),
		header:     components.NewHeader("Dick Cluster Status", "cluster"),
		footer:     components.NewFooter().SetActiveView(messages.StatusView),
		progress:   nil,
	}
}

// Init initializes the status view
func (s *Status) Init() tea.Cmd {
	return nil
}

// Update handles messages for the status view
func (s *Status) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.header.SetWidth(msg.Width)
		return s, nil
		
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			// Refresh config
			if cfg, err := config.LoadConfig(); err == nil {
				s.config = cfg
				s.lastUpdate = time.Now()
			}
			return s, nil
		}
		
	case messages.TickMsg:
		s.lastUpdate = msg.Time
		s.footer.UpdateTime(msg.Time)
		
		// Update progress bar if active
		if s.config.Status == "active" {
			s.updateProgressBar()
		}
		return s, nil
		
	case messages.ConfigReloadMsg:
		if cfg, err := config.LoadConfig(); err == nil {
			s.config = cfg
		}
		return s, nil
		
	case messages.ErrorMsg:
		s.err = msg.Err
		return s, nil
	}
	
	return s, nil
}

// View renders the status view
func (s *Status) View() string {
	if s.err != nil {
		return styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", s.err))
	}
	
	var sections []string
	
	// Header
	sections = append(sections, s.header.Render())
	
	// Project info section
	projectSection := s.renderProjectInfo()
	sections = append(sections, projectSection)
	
	// Status section
	statusSection := s.renderStatusInfo()
	sections = append(sections, statusSection)
	
	// Footer
	sections = append(sections, s.footer.Render())
	
	content := strings.Join(sections, "\n")
	
	// Add border if we have enough space
	if s.width > 60 {
		return styles.BorderStyle.Width(s.width - 4).Render(content)
	}
	
	return content
}

func (s *Status) renderProjectInfo() string {
	projectPath := s.config.ProjectPath
	if projectPath == "" {
		projectPath = "Current directory"
	}
	
	table := components.NewTable("Project Information")
	table.AddRow("Project", projectPath, styles.Icon("project"))
	table.AddRow("Name", s.config.Name, styles.Icon("name"))
	table.AddRow("Provider", s.config.Provider, styles.Icon("cluster"))
	table.AddRow("TTL", s.config.TTL, styles.Icon("ttl"))
	
	return table.Render() + "\n"
}

func (s *Status) renderStatusInfo() string {
	lines := []string{
		styles.TitleStyle.Render("Status Information"),
	}
	
	// Status with icon
	statusIcon := styles.Icon(s.config.Status)
	if s.config.Status == "" {
		statusIcon = styles.Icon("unknown")
	}
	
	lines = append(lines,
		fmt.Sprintf("%s %s %s", 
			statusIcon,
			styles.InfoLabelStyle.Render("Status:"), 
			styles.FormatStatus(s.config.Status)),
	)
	
	// Status-specific information
	switch s.config.Status {
	case "active":
		remaining := s.config.TimeRemaining()
		if remaining > 0 {
			lines = append(lines,
				fmt.Sprintf("%s %s %s", 
					styles.Icon("created"), 
					styles.InfoLabelStyle.Render("Created:"), 
					styles.InfoValueStyle.Render(s.config.CreatedAt.Format("2006-01-02 15:04:05"))),
				fmt.Sprintf("%s %s %s", 
					styles.Icon("expires"), 
					styles.InfoLabelStyle.Render("Expires:"), 
					styles.InfoValueStyle.Render(s.config.ExpiresAt.Format("2006-01-02 15:04:05"))),
				fmt.Sprintf("%s %s %s", 
					styles.Icon("remaining"), 
					styles.InfoLabelStyle.Render("Remaining:"), 
					styles.ProgressTextStyle.Render(remaining.String())),
			)
			
			// Add progress bar
			if s.progress != nil {
				lines = append(lines, "    "+s.progress.Render())
			}
		} else {
			lines = append(lines,
				styles.WarningStyle.Render(fmt.Sprintf("%s Should have been destroyed at: %s", 
					styles.Icon("warning"),
					s.config.ExpiresAt.Format("2006-01-02 15:04:05"))),
			)
		}
		
	case "destroyed":
		if !s.config.CreatedAt.IsZero() {
			lines = append(lines,
				fmt.Sprintf("%s %s %s", 
					styles.Icon("created"), 
					styles.InfoLabelStyle.Render("Was created:"), 
					styles.InfoValueStyle.Render(s.config.CreatedAt.Format("2006-01-02 15:04:05"))),
			)
		}
		
	default:
		lines = append(lines,
			styles.InfoLabelStyle.Render("Run 'dick new' to create a cluster"),
		)
	}
	
	return strings.Join(lines, "\n") + "\n"
}

func (s *Status) updateProgressBar() {
	if s.config.Status != "active" {
		s.progress = nil
		return
	}
	
	totalDuration, err := s.config.ParseTTL()
	if err != nil {
		s.progress = nil
		return
	}
	
	remaining := s.config.TimeRemaining()
	elapsed := totalDuration - remaining
	
	if s.progress == nil {
		barWidth := 30
		if s.width > 0 && s.width < 60 {
			barWidth = s.width - 20
		}
		s.progress = components.NewProgressBar(elapsed, totalDuration, barWidth)
	} else {
		s.progress.Update(elapsed)
		if s.width > 0 && s.width < 60 {
			s.progress.SetWidth(s.width - 20)
		}
	}
}