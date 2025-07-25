/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/killallgit/dick/internal/commands"
	"github.com/killallgit/dick/internal/config"
	"github.com/spf13/cobra"
)

// Remove package-level flag variables - use Viper directly

var statusCmd = &cobra.Command{
	Use:       "status [k8s|kubernetes]",
	Short:     "Show environment status and remaining TTL",
	ValidArgs: []string{"k8s", "kubernetes"},
	Args:      cobra.MaximumNArgs(1), // Use built-in validator instead of custom function
	Long: `Display the current status of your ephemeral environment including
remaining time before automatic TTL cleanup.

The environment type is optional - defaults to showing status for all environments.`,
	Example: `  dick status                 # Show status for all environments
  dick status k8s --watch     # Watch k8s environment status
  dick status kubernetes      # Show kubernetes environment status`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Bind status command flags with proper namespacing
		return config.BindStatusFlags(config.GlobalViper, cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config using Viper
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get the effective configuration for the status command
		statusConfig := cfg.GetEffectiveStatusConfig()

		// Handle noun parameter (optional)
		if len(args) == 1 {
			noun := args[0]
			if noun == "kubernetes" {
				noun = "k8s" // normalize
			}
			// For now, we only support k8s, so just validate it
			if noun != "k8s" {
				return fmt.Errorf("unsupported environment type: %s (only k8s is currently supported)", noun)
			}
		}

		opts := commands.StatusOptions{
			Config: cfg,
			Watch:  statusConfig.Watch,
		}
		
		return commands.RunStatus(opts)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	
	// Define flags with modern patterns - no package variables needed
	statusCmd.Flags().BoolP("watch", "w", false, "Watch environment status with live updates")
	
	// Add completion for environment types (ValidArgs provides this automatically)
}