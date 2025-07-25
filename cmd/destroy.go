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

var destroyCmd = &cobra.Command{
	Use:       "destroy [k8s|kubernetes]",
	Short:     "Immediately destroy the environment bypassing TTL",
	ValidArgs: []string{"k8s", "kubernetes"},
	Args:      cobra.MaximumNArgs(1), // Use built-in validator instead of custom function
	Long: `Immediately destroy the ephemeral environment without waiting for TTL
expiration. This stops any running TTL timer and cleans up all resources.

The environment type is optional - defaults to destroying all environments.`,
	Example: `  dick destroy                # Destroy all environments
  dick destroy k8s --force    # Force destroy k8s environment
  dick destroy kubernetes     # Destroy kubernetes environment`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Bind destroy command flags with proper namespacing
		return config.BindDestroyFlags(config.GlobalViper, cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config using Viper
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get the effective configuration for the destroy command
		destroyConfig := cfg.GetEffectiveDestroyConfig()

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

		opts := commands.DestroyOptions{
			Config: cfg,
			Force:  destroyConfig.Force,
		}
		
		return commands.RunDestroy(opts)
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
	
	// Define flags with modern patterns - no package variables needed
	destroyCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	
	// Add completion for environment types (ValidArgs provides this automatically)
}