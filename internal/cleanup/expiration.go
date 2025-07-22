package cleanup

import (
	"fmt"
	"log"
	"time"

	"github.com/killallgit/dick/internal/config"
	"github.com/killallgit/dick/internal/tui"
)

// CheckAndHandleExpiration checks for expired clusters and handles them appropriately
func CheckAndHandleExpiration(cfg *config.Config) (bool, error) {
	// Check if we should attempt any cleanup (including retries)
	if !cfg.ShouldAttemptCleanup() && !cfg.ShouldRetryCleanup() {
		return false, nil // Not expired, not active, or no retry needed
	}

	_, expiredSince := cfg.CheckExpiration()
	
	// Handle expired cluster based on configuration
	if cfg.ShouldAutoDestroy() || (cfg.ShouldRetryCleanup() && cfg.Force) {
		// Force mode: automatic cleanup without prompt
		return handleAutoDestroy(cfg, expiredSince)
	} else if cfg.ShouldPromptDestroy() || cfg.ShouldRetryCleanup() {
		// Interactive mode: prompt user for cleanup (or retry)
		return handlePromptDestroy(cfg, expiredSince)
	} else {
		// Expired cluster with cleanup attempts exhausted - show detailed status
		fmt.Printf("%s Cluster '%s' expired %s ago - cleanup attempts exhausted\n",
			tui.Icon("warning"),
			tui.InfoValueStyle.Render(cfg.Name),
			tui.InfoValueStyle.Render(expiredSince.String()))
		fmt.Printf("%s Cleanup status: %s\n",
			tui.Icon("info"),
			tui.InfoValueStyle.Render(cfg.GetCleanupStatus()))
		fmt.Printf("%s Run 'dick destroy --force' to manually cleanup\n",
			tui.Icon("info"))
		return false, nil
	}
}

// handleAutoDestroy performs automatic cleanup in force mode
func handleAutoDestroy(cfg *config.Config, expiredSince time.Duration) (bool, error) {
	retryText := ""
	if cfg.CleanupAttempts > 0 {
		retryText = fmt.Sprintf(" (retry %d)", cfg.CleanupAttempts+1)
	}
	
	fmt.Printf("%s Cluster expired %s ago. Auto-destroying (force mode)%s...\n",
		tui.Icon("warning"),
		tui.InfoValueStyle.Render(expiredSince.String()),
		retryText)

	// Perform cleanup
	err := ForceCleanup(cfg)
	
	// Update cleanup state based on result
	if err != nil {
		cfg.MarkCleanupFailed(err)
		if saveErr := config.SaveConfig(cfg); saveErr != nil {
			log.Printf("Warning: failed to save config after cleanup failure: %v", saveErr)
		}
		return false, fmt.Errorf("automatic cleanup failed: %w", err)
	}

	cfg.MarkCleanupSuccessful()
	if saveErr := config.SaveConfig(cfg); saveErr != nil {
		log.Printf("Warning: failed to save config after successful cleanup: %v", saveErr)
	}

	fmt.Printf("%s %s\n",
		tui.Icon("success"),
		tui.SuccessStyle.Render("Cluster automatically destroyed"))

	return true, nil
}

// CheckExpirationForCommand handles expired cluster checking for any command
func CheckExpirationForCommand(cfg *config.Config) error {
	// Only check active clusters
	if cfg.Status != "active" {
		return nil
	}

	// Check and handle expiration
	_, err := CheckAndHandleExpiration(cfg)
	return err
}

// handlePromptDestroy prompts user for cleanup confirmation  
func handlePromptDestroy(cfg *config.Config, expiredSince time.Duration) (bool, error) {
	title := "Cluster Expired"
	retryText := ""
	if cfg.CleanupAttempts > 0 {
		retryText = fmt.Sprintf("\n\nPrevious cleanup attempts: %s", cfg.GetCleanupStatus())
	}
	
	message := fmt.Sprintf(
		"Cluster '%s' expired %s ago.\n\n"+
			"Would you like to destroy it now?\n\n"+
			"• Yes: Destroy the cluster immediately\n"+
			"• No: Keep the cluster (you can destroy it manually later)\n\n"+
			"Note: You can set 'force: true' in .dick.yaml to auto-destroy expired clusters.%s",
		cfg.Name,
		expiredSince.String(),
		retryText)

	confirmed, err := tui.RunConfirmation(title, message)
	if err != nil {
		return false, fmt.Errorf("failed to show confirmation: %w", err)
	}

	if confirmed {
		// User confirmed cleanup
		fmt.Printf("%s Destroying expired cluster...\n",
			tui.Icon("destroy"))

		err := ForceCleanup(cfg)
		if err != nil {
			cfg.MarkCleanupFailed(err)
			if saveErr := config.SaveConfig(cfg); saveErr != nil {
				log.Printf("Warning: failed to save config after cleanup failure: %v", saveErr)
			}
			return false, fmt.Errorf("cleanup failed: %w", err)
		}

		cfg.MarkCleanupSuccessful()
		if saveErr := config.SaveConfig(cfg); saveErr != nil {
			log.Printf("Warning: failed to save config after successful cleanup: %v", saveErr)
		}

		fmt.Printf("%s %s\n",
			tui.Icon("success"),
			tui.SuccessStyle.Render("Expired cluster destroyed"))
		
		return true, nil
	} else {
		// User declined cleanup - record attempt
		cfg.MarkCleanupAttempted()
		if err := config.SaveConfig(cfg); err != nil {
			log.Printf("Warning: failed to save config: %v", err)
		}
		
		fmt.Printf("%s Cleanup cancelled. Cluster remains active.\n",
			tui.Icon("warning"))
		
		return false, nil
	}
}

