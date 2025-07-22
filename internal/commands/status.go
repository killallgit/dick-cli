package commands

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/killallgit/dick/internal/cleanup"
	"github.com/killallgit/dick/internal/common"
	"github.com/killallgit/dick/internal/config"
	"github.com/killallgit/dick/internal/tui"
)

// StatusOptions holds configuration for the status command
type StatusOptions struct {
	Config *config.Config
	Watch  bool
}

// RunStatus executes the status command with the given options
func RunStatus(opts StatusOptions) error {
	cfg := opts.Config

	// Check for expired clusters before showing status
	// This replaces the global pre-run hook
	if err := checkExpiration(cfg); err != nil {
		// Don't fail on expiration check errors, just warn
		fmt.Printf("Warning: expiration check failed: %v\n", err)
	}

	if opts.Watch {
		return runWatch(cfg)
	}
	
	return runSimple(cfg)
}

// checkExpiration handles expired cluster checking for status command
func checkExpiration(cfg *config.Config) error {
	return cleanup.CheckExpirationForCommand(cfg)
}

// runSimple shows the basic status output
func runSimple(cfg *config.Config) error {
	if common.ShouldShowTaskOutput() {
		// Verbose mode - show detailed information
		return runVerboseStatus(cfg)
	}
	
	// Minimal mode - single line format
	switch cfg.Status {
	case "active":
		remaining := cfg.TimeRemaining()
		if remaining > 0 {
			totalDuration, err := cfg.ParseTTL()
			if err != nil {
				// Fallback to simple format if TTL parsing fails
				fmt.Printf("%s %s %s\n", cfg.Name, tui.FormatStatus(cfg.Status), remaining.String())
				return nil
			}
			
			// Show progress bar format: name [progress] time-remaining
			fmt.Println(tui.RenderProgressBar(cfg.Name, remaining, totalDuration, 20))
		} else {
			// Expired cluster
			fmt.Printf("%s %s %s\n", cfg.Name, 
				tui.StatusExpiredStyle.Render("EXPIRED"),
				tui.ProgressTextStyle.Render("overdue"))
		}
		
	case "destroyed":
		fmt.Printf("%s %s\n", cfg.Name, tui.FormatStatus(cfg.Status))
		
	default:
		// No cluster or unknown status
		fmt.Printf("No active cluster (run 'dick new' to create)\n")
	}
	
	return nil
}

// runVerboseStatus shows detailed status information
func runVerboseStatus(cfg *config.Config) error {
	// Header
	fmt.Print(tui.HeaderStyle.Render(fmt.Sprintf("%s Dick Cluster Status", tui.Icon("cluster"))))
	fmt.Println()
	fmt.Print(tui.Divider(50))
	fmt.Println()
	
	// Project information
	projectPath := cfg.ProjectPath
	if projectPath == "" {
		projectPath = "Current directory"
	}
	
	fmt.Printf("%s: %s\n", tui.Icon("project"), projectPath)
	fmt.Printf("%s: %s\n", tui.Icon("name"), cfg.Name)
	fmt.Printf("%s: %s\n", tui.Icon("cluster"), cfg.Provider)
	fmt.Printf("%s: %s\n", tui.Icon("ttl"), cfg.TTL)
	
	// Status information
	fmt.Printf("\n%s: %s\n", tui.Icon("active"), tui.FormatStatus(cfg.Status))

	switch cfg.Status {
	case "active":
		remaining := cfg.TimeRemaining()
		if remaining > 0 {
			fmt.Printf("%s: %s\n", tui.Icon("created"), cfg.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("%s: %s\n", tui.Icon("expires"), cfg.ExpiresAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("%s: %s\n", tui.Icon("remaining"), remaining.String())
		} else {
			fmt.Printf("%s: Should have been destroyed at %s\n", 
				tui.Icon("warning"),
				cfg.ExpiresAt.Format("2006-01-02 15:04:05"))
		}

	case "destroyed":
		if !cfg.CreatedAt.IsZero() {
			fmt.Printf("%s: %s\n", tui.Icon("created"), cfg.CreatedAt.Format("2006-01-02 15:04:05"))
		}

	default:
		fmt.Printf("%s: Run 'dick new' to create a cluster\n", tui.Icon("info"))
	}

	// Configuration options
	fmt.Println()
	showConfigOptions(cfg)
	
	return nil
}

// runWatch shows the continuous monitoring TUI
func runWatch(cfg *config.Config) error {
	model := tui.NewWatchModel(cfg)
	p := tea.NewProgram(model, tea.WithAltScreen())
	
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run watch TUI: %w", err)
	}

	return nil
}

// showConfigOptions displays configuration options
func showConfigOptions(cfg *config.Config) {
	fmt.Print(tui.TitleStyle.Render(" Configuration Options "))
	fmt.Println()

	// Force mode
	forceStatus := "Disabled"
	if cfg.Force {
		forceStatus = "Enabled (auto-destroy expired clusters)"
	}
	fmt.Printf("  • %s %s\n", 
		tui.InfoLabelStyle.Render("Force mode:"), 
		tui.InfoValueStyle.Render(forceStatus))
	
	// Cleanup attempts
	if cfg.CleanupAttempts > 0 {
		fmt.Printf("  • %s %s\n", 
			tui.InfoLabelStyle.Render("Cleanup status:"), 
			tui.InfoValueStyle.Render(cfg.GetCleanupStatus()))
	}
}