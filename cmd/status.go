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

// statusCmd represents the status command
var (
	statusWatchFlag bool
)

var statusCmd = &cobra.Command{
	Use:   "status [noun]",
	Short: "Show provider status and TTL information",
	Long: `Display the current status of the provider in this project,
including remaining time before TTL cleanup.

Use --watch flag for continuous monitoring with live updates.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load config using Viper
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// Handle noun parameter (optional) - currently just for validation
		// In the future, this could filter status by specific resource types
		if len(args) > 0 {
			noun := args[0]
			switch noun {
			case "k8s", "kubernetes":
				// Valid noun - could filter for k8s-specific status in future
			default:
				log.Fatalf("Unknown noun: %s. Supported nouns: k8s", noun)
			}
		}

		opts := commands.StatusOptions{
			Config: cfg,
			Watch:  statusWatchFlag,
		}
		
		if err := commands.RunStatus(opts); err != nil {
			log.Fatalf("Error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolVar(&statusWatchFlag, "watch", false, "Watch cluster status with live updates")
}