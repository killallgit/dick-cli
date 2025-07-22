/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

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
	Short: "Create and manage ephemeral environments",
	Long:  `Run help for more information.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set global flag values
		common.VerboseFlag = verbose
		common.SilentFlag = silent
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .dick.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output for task commands")
	rootCmd.PersistentFlags().BoolVar(&silent, "silent", false, "silent output for task commands")
}

func initConfig() {
	// Initialize Viper configuration
	if err := config.Initialize(cfgFile); err != nil {
		// Don't exit on config errors, let commands handle it
		return
	}
}


