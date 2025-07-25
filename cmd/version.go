/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is set by build process
	Version = "dev"
	// Commit is set by build process
	Commit = "unknown"
	// Date is set by build process  
	Date = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Display version information for dick CLI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("dick version %s\n", Version)
		fmt.Printf("commit: %s\n", Commit)
		fmt.Printf("built: %s\n", Date)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	
	// Set the version on the root command for --version flag support
	rootCmd.Version = Version
}