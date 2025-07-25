/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"fmt"
	"os"
	
	"github.com/killallgit/dick/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
