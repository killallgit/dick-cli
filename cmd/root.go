/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	
	"github.com/killallgit/dick/internal/common"
	"github.com/killallgit/dick/internal/config"
)

var (
	cfgFile string
	verbose bool
	silent  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dick",
	Short: "Ephemeral infrastructure manager with automatic TTL cleanup",
	Long: `Dick creates temporary cloud environments that automatically self-destruct
after a specified time period, preventing resource waste and cost overruns.

Currently supports Kubernetes clusters via Kind with automatic cleanup scheduling.

Environment variables (with command namespacing):
  Global flags:
    DICK_GLOBAL_VERBOSE=true         - Enable verbose output
    DICK_GLOBAL_SILENT=true          - Enable silent output

  New command flags:
    DICK_NEW_PROVIDER=kind           - Default provider (kind, tofu)
    DICK_NEW_TTL=5m                  - Default TTL for new environments
    DICK_NEW_NAME=dev-cluster        - Default cluster/environment name
    DICK_NEW_FORCE=true              - Force overwrite config

  Status command flags:
    DICK_STATUS_WATCH=true           - Watch status by default

  Destroy command flags:
    DICK_DESTROY_FORCE=true          - Skip confirmation prompts

  Legacy environment variables (deprecated but supported):
    DICK_PROVIDER, DICK_TTL, DICK_NAME, DICK_FORCE`,
	
	// Use PersistentPreRunE for proper error handling
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Set global flag values
		common.VerboseFlag = verbose
		common.SilentFlag = silent
		
		// Initialize configuration if needed
		if err := initConfig(cmd); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}
		
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Remove cobra.OnInitialize since we handle config in PersistentPreRunE
	
	// Define persistent flags with modern patterns
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default searches for .dick.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output for task commands")
	rootCmd.PersistentFlags().BoolVar(&silent, "silent", false, "silent output for task commands")
	
	// Mark flags as mutually exclusive
	rootCmd.MarkFlagsMutuallyExclusive("verbose", "silent")
	
	// Add completion for config flag to suggest YAML files
	rootCmd.RegisterFlagCompletionFunc("config", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveFilterFileExt
	})
}

// initConfig initializes the configuration with better error handling
func initConfig(cmd *cobra.Command) error {
	// Initialize Viper configuration with the config file path
	if err := config.Initialize(cfgFile); err != nil {
		return err
	}
	
	// Bind global flags to Viper with proper namespacing
	// Use the root command from the parameter to avoid initialization cycle
	if err := config.BindGlobalFlags(config.GlobalViper, cmd.Root()); err != nil {
		return fmt.Errorf("failed to bind global flags: %w", err)
	}
	
	return nil
}


