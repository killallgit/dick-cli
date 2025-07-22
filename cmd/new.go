/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"log"

	"github.com/killallgit/dick/internal/commands"
	"github.com/killallgit/dick/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ttlFlag         string
	nameFlag        string
	newForceFlag    bool
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new Kubernetes environment and watch in real-time",
	Long: `Create a new ephemeral Kubernetes cluster that automatically destroys itself
after the specified TTL expires. The command stays active showing a real-time 
dashboard with TTL countdown until the cluster is destroyed or you exit.`,
	Example: `  dick new                    # Create k8s cluster with 5m TTL, then watch
  dick new --ttl 10m          # Create with 10 minute TTL, then watch
  dick new --name my-cluster  # Create with custom name, then watch
  dick new --force --ttl 30m  # Force new config and watch`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load config using Viper
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// Default to Kubernetes provider (no noun parameter needed)
		cfg.Provider = "kind"

		// Apply flag overrides (removed providerFlag)
		config.ApplyFlagOverrides(cfg, nil, &ttlFlag, &nameFlag, &newForceFlag)

		opts := commands.NewOptions{
			Provider: cfg.Provider,
			TTL:      cfg.TTL,
			Name:     cfg.Name,
			Wait:     true, // Always watch by default
			Force:    cfg.Force,
		}
		
		if err := commands.RunNew(opts); err != nil {
			log.Fatalf("Error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	// Add flags for CLI argument overrides (removed --provider and --wait)
	newCmd.Flags().StringVar(&ttlFlag, "ttl", "", "Time to live (e.g., 5m, 10m, 1h)")
	newCmd.Flags().StringVar(&nameFlag, "name", "", "Cluster name")
	newCmd.Flags().BoolVar(&newForceFlag, "force", false, "Force overwrite existing config with defaults and provided args")

	// Bind flags to Viper
	if config.GlobalViper != nil {
		viper.BindPFlag("ttl", newCmd.Flags().Lookup("ttl"))
		viper.BindPFlag("name", newCmd.Flags().Lookup("name"))
		viper.BindPFlag("force", newCmd.Flags().Lookup("force"))
	}
}