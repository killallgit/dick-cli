/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// k8sCmd represents the k8s command
var k8sCmd = &cobra.Command{
	Use:   "k8s",
	Short: "K8s namespace",
	Long: `Used to target the noun: k8s`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("k8s called")
	},
}

func init() {
	// TODO: Remove this subcommand - we now use noun parameters instead
	// newCmd.AddCommand(k8sCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// k8sCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// k8sCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
