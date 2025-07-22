package commands

import (
	"fmt"

	"github.com/killallgit/dick/internal/cleanup"
	"github.com/killallgit/dick/internal/config"
	"github.com/killallgit/dick/internal/tui"
)

// DestroyOptions holds configuration for the destroy command
type DestroyOptions struct {
	Config *config.Config
	Force  bool
}

// RunDestroy executes the destroy command with the given options
func RunDestroy(opts DestroyOptions) error {
	cfg := opts.Config

	// Check for expired clusters before manual destroy
	// This handles automatic cleanup of expired clusters
	if err := cleanup.CheckExpirationForCommand(cfg); err != nil {
		// Don't fail on expiration check errors, just warn
		fmt.Printf("Warning: expiration check failed: %v\n", err)
	}

	// Check if cluster is active (might have been cleaned up by expiration check)
	if cfg.Status != "active" {
		fmt.Printf("%s Cluster %s is not active (status: %s)\n", 
			tui.Icon("warning"),
			tui.InfoValueStyle.Render(cfg.Name), 
			tui.FormatStatus(cfg.Status))
		return nil
	}

	// Show confirmation unless --force is used
	if !opts.Force {
		confirmed, err := tui.RunConfirmation(
			"Destroy Cluster",
			fmt.Sprintf("Are you sure you want to destroy cluster '%s'?\n\nThis action cannot be undone.", cfg.Name),
		)
		if err != nil {
			return fmt.Errorf("failed to show confirmation: %w", err)
		}
		
		if !confirmed {
			fmt.Printf("%s Destroy cancelled\n", tui.Icon("warning"))
			return nil
		}
	}

	fmt.Printf("%s Destroying cluster %s...\n", 
		tui.Icon("destroy"), 
		tui.InfoValueStyle.Render(cfg.Name))

	// Cancel any scheduled cleanup first
	if err := cleanup.CancelScheduledCleanup(cfg); err != nil {
		fmt.Printf("%s Warning: failed to cancel scheduled cleanup: %v\n", 
			tui.Icon("warning"), err)
	}

	// Force cleanup immediately
	if err := cleanup.ForceCleanup(cfg); err != nil {
		return fmt.Errorf("failed to destroy cluster: %w", err)
	}

	fmt.Printf("%s %s\n", 
		tui.Icon("success"), 
		tui.SuccessStyle.Render(fmt.Sprintf("Cluster '%s' destroyed successfully!", cfg.Name)))
	return nil
}

