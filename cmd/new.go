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
	providerFlag    string
	ttlFlag         string
	nameFlag        string
	waitFlag        bool
	newForceFlag    bool
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new [noun]",
	Short: "Create a new ephemeral environment",
	Long: `Create a new ephemeral cluster with TTL cleanup.

The cluster will be automatically destroyed after the specified TTL expires.
By default, the command exits immediately after creating the cluster and 
scheduling the cleanup. Use --wait to keep the process running until cleanup.

Use --force to completely overwrite the existing .dick.yaml configuration
with defaults and any arguments provided via command line.
	
Examples:
  dick new k8s --ttl 5m --provider kind
  dick new k8s --ttl 10m --name my-cluster
  dick new k8s --ttl 1h --wait  # Keep running until cleanup
  dick new k8s --force --ttl 30m  # Force new config with 30m TTL
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load config using Viper
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// Handle noun parameter (optional)
		var nounProvider string
		if len(args) > 0 {
			noun := args[0]
			switch noun {
			case "k8s", "kubernetes":
				nounProvider = "kind"
			default:
				log.Fatalf("Unknown noun: %s. Supported nouns: k8s", noun)
			}
		}

		// Apply flag overrides
		config.ApplyFlagOverrides(cfg, &providerFlag, &ttlFlag, &nameFlag, &newForceFlag)
		
		// Override provider from noun if provided
		if nounProvider != "" {
			cfg.Provider = nounProvider
		}

		opts := commands.NewOptions{
			Provider: cfg.Provider,
			TTL:      cfg.TTL,
			Name:     cfg.Name,
			Wait:     waitFlag,
			Force:    cfg.Force,
		}
		
		if err := commands.RunNew(opts); err != nil {
			log.Fatalf("Error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	// Add flags for CLI argument overrides
	newCmd.Flags().StringVar(&providerFlag, "provider", "", "Infrastructure provider (kind, etc.)")
	newCmd.Flags().StringVar(&ttlFlag, "ttl", "", "Time to live (e.g., 5m, 10m, 1h)")
	newCmd.Flags().StringVar(&nameFlag, "name", "", "Cluster name")
	newCmd.Flags().BoolVar(&waitFlag, "wait", false, "Keep process running until cleanup completes")
	newCmd.Flags().BoolVar(&newForceFlag, "force", false, "Force overwrite existing config with defaults and provided args")

	// Bind flags to Viper
	if config.GlobalViper != nil {
		viper.BindPFlag("provider", newCmd.Flags().Lookup("provider"))
		viper.BindPFlag("ttl", newCmd.Flags().Lookup("ttl"))
		viper.BindPFlag("name", newCmd.Flags().Lookup("name"))
		viper.BindPFlag("force", newCmd.Flags().Lookup("force"))
	}
}