/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"log"

	"github.com/killallgit/dick/internal/commands"
	"github.com/killallgit/dick/internal/config"
	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var (
	forceFlag bool
)

var destroyCmd = &cobra.Command{
	Use:   "destroy [noun]",
	Short: "Manually destroy the cluster",
	Long: `Immediately destroy the cluster without waiting for TTL expiration.
This will stop any running TTL timer and clean up all resources.

Use --force to skip confirmation prompt.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load config using Viper
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// Handle noun parameter (optional) - currently just for validation
		if len(args) > 0 {
			noun := args[0]
			switch noun {
			case "k8s", "kubernetes":
				// Valid noun - matches current provider
			default:
				log.Fatalf("Unknown noun: %s. Supported nouns: k8s", noun)
			}
		}

		opts := commands.DestroyOptions{
			Config: cfg,
			Force:  forceFlag,
		}
		
		if err := commands.RunDestroy(opts); err != nil {
			log.Fatalf("Error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
	destroyCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Skip confirmation prompt")
}