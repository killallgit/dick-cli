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

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new [k8s|kubernetes]",
	Short: "Create a new ephemeral environment and watch in real-time",
	Long: `Create a new ephemeral infrastructure environment that automatically destroys itself
after the specified TTL expires. The command stays active showing a real-time 
dashboard with TTL countdown until the environment is destroyed or you exit.

The environment type is optional - defaults to k8s (Kubernetes via Kind).`,
	ValidArgs: []string{"k8s", "kubernetes"},
	Example: `  dick new                    # Create k8s cluster with 5m TTL, then watch
  dick new k8s --ttl 10m      # Create with 10 minute TTL, then watch
  dick new --name my-cluster  # Create with custom name, then watch
  dick new --force --ttl 30m  # Force new config and watch`,
	Args: cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return config.BindNewFlags(config.GlobalViper, cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		newConfig := cfg.GetEffectiveNewConfig()

		if err := config.ValidateConfig(cfg); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}

		if len(args) == 1 {
			noun := args[0]
			if noun == "kubernetes" {
				noun = "k8s" // normalize
			}
			if noun != "k8s" {
				return fmt.Errorf("unsupported environment type: %s (only k8s is currently supported)", noun)
			}
		}

		opts := commands.NewOptions{
			Provider: newConfig.Provider,
			TTL:      newConfig.TTL,
			Name:     newConfig.Name,
			Wait:     true, // Always watch by default
			Force:    newConfig.Force,
		}
		
		return commands.RunNew(opts)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	newCmd.Flags().StringP("ttl", "t", "", "Time to live (e.g., 5m, 10m, 1h)")
	newCmd.Flags().StringP("name", "n", "", "Environment name")  
	newCmd.Flags().StringP("provider", "p", "", "Infrastructure provider (kind, tofu)")
	newCmd.Flags().BoolP("force", "f", false, "Force overwrite existing config with defaults and provided args")

	newCmd.RegisterFlagCompletionFunc("ttl", cobra.FixedCompletions([]string{"5m", "10m", "30m", "1h", "2h"}, cobra.ShellCompDirectiveDefault))
	newCmd.RegisterFlagCompletionFunc("provider", cobra.FixedCompletions([]string{"kind", "tofu"}, cobra.ShellCompDirectiveDefault))

}