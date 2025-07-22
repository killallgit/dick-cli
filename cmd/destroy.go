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

// destroyCmd represents the destroy command
var (
	forceFlag bool
)

var destroyCmd = &cobra.Command{
	Use:       "destroy [noun]",
	Short:     "Immediately destroy the environment bypassing TTL",
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
	Long: `Immediately destroy the ephemeral environment without waiting for TTL
expiration. This stops any running TTL timer and cleans up all resources.`,
	Example: `  dick destroy
  dick destroy k8s --force
  dick destroy kubernetes`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load config using Viper
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// Handle noun parameter (optional) - ValidArgs handles validation

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