/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
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
	Use:       "status [noun]",
	Short:     "Show environment status and remaining TTL",
	ValidArgs: []string{"k8s", "kubernetes"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return nil
		}
		if len(args) > 1 {
			return fmt.Errorf("too many arguments: %d", len(args))
		}
		validArgs := []string{"k8s", "kubernetes"}
		for _, validArg := range validArgs {
			if args[0] == validArg {
				return nil
			}
		}
		return fmt.Errorf("invalid argument %q for %q", args[0], cmd.CommandPath())
	},
	Long: `Display the current status of your ephemeral environment including
remaining time before automatic TTL cleanup.`,
	Example: `  dick status
  dick status k8s --watch
  dick status kubernetes --watch`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load config using Viper
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// Handle noun parameter (optional) - ValidArgs handles validation
		// In the future, this could filter status by specific resource types

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