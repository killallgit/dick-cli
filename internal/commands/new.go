package commands

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/killallgit/dick/internal/cleanup"
	"github.com/killallgit/dick/internal/common"
	"github.com/killallgit/dick/internal/config"
	"github.com/killallgit/dick/internal/tui"
)

// NewOptions holds configuration for the new command
type NewOptions struct {
	Provider string
	TTL      string
	Name     string
	Wait     bool
	Force    bool
}

// RunNew executes the new command with the given options
func RunNew(opts NewOptions) error {
	var cfg *config.Config
	var err error

	if opts.Force {
		// Create new config with defaults when force flag is used
		cfg = &config.Config{
			Provider: "kind",
			TTL:      "5m",
			Name:     "dev-cluster",
		}
		fmt.Printf("%s Force flag enabled - creating new configuration\n", tui.Icon("warning"))
	} else {
		// Load existing config
		cfg, err = config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	// Check for expired clusters before creating new ones
	// This ensures any existing expired clusters are cleaned up first
	if err := cleanup.CheckExpirationForCommand(cfg); err != nil {
		// Don't fail on expiration check errors, just warn
		fmt.Printf("%s Warning: expiration check failed: %v\n", tui.Icon("warning"), err)
	} else if cfg.Status == "destroyed" {
		// Expired cluster was found and cleaned up
		fmt.Printf("%s Cleaned up expired cluster before creating new one\n", tui.Icon("success"))
	}

	// Apply CLI flag overrides
	applyFlags(cfg, opts)

	// Validate TTL format
	duration, err := cfg.ParseTTL()
	if err != nil {
		return fmt.Errorf("invalid TTL: %w", err)
	}

	// Display fancy header with colored row
	fmt.Printf("\n%s\n", tui.Divider(60))
	fmt.Printf("%s %s %s %s %s %s %s\n",
		tui.HeaderStyle.Render("CREATING"),
		tui.InfoLabelStyle.Render("Provider:"),
		tui.SuccessStyle.Render(cfg.Provider),
		tui.InfoLabelStyle.Render("Name:"),
		tui.InfoValueStyle.Render(cfg.Name),
		tui.InfoLabelStyle.Render("TTL:"),
		tui.WarningStyle.Render(duration.String()))
	fmt.Printf("%s\n\n", tui.Divider(60))

	// Execute the create task
	if err := executeCreateTask(cfg); err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}

	// Mark cluster as active and save state
	if err := cfg.SetActive(); err != nil {
		return fmt.Errorf("failed to set cluster active: %w", err)
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Save config again after scheduling to persist job ID
	defer func() {
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("%s Warning: failed to save config: %v\n", tui.Icon("warning"), err)
		}
	}()

	// Start Go-based TTL timer (no system scheduling)
	if err := cleanup.StartTTLTimer(cfg); err != nil {
		return fmt.Errorf("failed to start TTL timer: %w", err)
	}

	// Display success with fancy formatting
	fmt.Printf("\n%s\n", tui.Divider(60))
	fmt.Printf("%s %s\n", 
		tui.SuccessStyle.Render("✓"),
		tui.SuccessStyle.Render(fmt.Sprintf("Cluster '%s' created successfully!", cfg.Name)))
	fmt.Printf("%s\n", tui.Divider(60))
	
	// Display cluster details in a formatted table
	fmt.Printf("\n%-15s %s\n", 
		tui.InfoLabelStyle.Render("STATUS:"), 
		tui.StatusActiveStyle.Render("ACTIVE"))
	fmt.Printf("%-15s %s\n", 
		tui.InfoLabelStyle.Render("PROVIDER:"), 
		tui.InfoValueStyle.Render(cfg.Provider))
	fmt.Printf("%-15s %s\n", 
		tui.InfoLabelStyle.Render("NAME:"), 
		tui.InfoValueStyle.Render(cfg.Name))
	fmt.Printf("%-15s %s\n", 
		tui.InfoLabelStyle.Render("TTL:"), 
		tui.WarningStyle.Render(duration.String()))
	fmt.Printf("%-15s %s\n", 
		tui.InfoLabelStyle.Render("EXPIRES AT:"), 
		tui.WarningStyle.Render(cfg.ExpiresAt.Format("15:04:05")))
	fmt.Printf("%-15s %s\n", 
		tui.InfoLabelStyle.Render("PROJECT PATH:"), 
		tui.InfoValueStyle.Render(cfg.ProjectPath))

	// Always run in watch mode (no conditional check)
	fmt.Printf("\n%s Starting real-time dashboard - process will remain active until cleanup\n", 
		tui.Icon("timer"))
	fmt.Printf("%s Press Ctrl+C to exit early (TTL timer will continue in background)\n", 
		tui.Icon("info"))
	
	// Start watch mode using the existing TUI
	if err := startWatchMode(cfg); err != nil {
		return fmt.Errorf("error in watch mode: %w", err)
	}

	return nil
}

// applyFlags applies command line flag overrides to the config
func applyFlags(cfg *config.Config, opts NewOptions) {
	if opts.Provider != "" {
		cfg.Provider = opts.Provider
	}
	if opts.TTL != "" {
		cfg.TTL = opts.TTL
	}
	if opts.Name != "" {
		cfg.Name = opts.Name
	}
}

// executeCreateTask runs the appropriate create task based on provider
func executeCreateTask(cfg *config.Config) error {
	// Check if task command is available
	if _, err := exec.LookPath("task"); err != nil {
		return fmt.Errorf("task command not found. Please install Task: https://taskfile.dev/installation/")
	}

	// Get current working directory
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Construct task file path based on provider (noun-based)
	var taskFile string
	switch cfg.Provider {
	case "kind":
		taskFile = filepath.Join(pwd, "tasks", "Taskfile.k8s.yaml")
	default:
		return fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
	
	// Check if new taskfile exists, fallback to legacy if needed
	if _, err := os.Stat(taskFile); err != nil {
		// Fallback to legacy taskfile for backward compatibility
		taskFile = filepath.Join(pwd, "tasks", "Taskfile.new.yaml")
		if _, err := os.Stat(taskFile); err != nil {
			return fmt.Errorf("no taskfile found for provider %s: tried %s and %s", 
				cfg.Provider, taskFile, filepath.Join(pwd, "tasks", "Taskfile.k8s.yaml"))
		}
	}

	// Use standardized hook:setup task, fallback to provider-specific task
	taskName := "hook:setup"

	// Execute: task -t tasks/Taskfile.new.yaml kind:create CLUSTER_NAME=<name>
	taskArgs := []string{"-t", taskFile}
	
	// Add --silent flag by default unless verbose mode is enabled
	if !common.ShouldShowTaskOutput() {
		taskArgs = append(taskArgs, "--silent")
	}
	
	taskArgs = append(taskArgs, taskName, fmt.Sprintf("CLUSTER_NAME=%s", cfg.Name))
	command := exec.Command("task", taskArgs...)
	command.Dir = pwd
	
	// Show spinner while creating cluster
	spinnerDone := make(chan bool)
	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-spinnerDone:
				fmt.Printf("\r%s\r", strings.Repeat(" ", 50)) // Clear the line
				return
			default:
				fmt.Printf("\r%s %s %s",
					tui.ProgressBarStyle.Render(frames[i]),
					tui.InfoLabelStyle.Render("Creating cluster"),
					tui.InfoValueStyle.Render(cfg.Name))
				i = (i + 1) % len(frames)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
	
	// Handle output based on verbose/silent flags
	var cmdErr error
	if common.ShouldShowTaskOutput() {
		// Stop spinner for verbose mode
		spinnerDone <- true
		fmt.Printf("\n")
		
		// Verbose mode: show all output
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		
		cmdErr = command.Run()
	} else {
		// Silent mode: capture output and only show on error
		output, err := command.CombinedOutput()
		cmdErr = err
		
		// Stop spinner
		spinnerDone <- true
		
		if cmdErr != nil {
			return fmt.Errorf("task execution failed: %w, output: %s", cmdErr, string(output))
		}
		
		// Show success message with checkmark
		fmt.Printf("\r%s %s %s\n", 
			tui.SuccessStyle.Render("✓"),
			tui.InfoLabelStyle.Render("Created cluster"),
			tui.InfoValueStyle.Render(cfg.Name))
	}

	if cmdErr != nil {
		return fmt.Errorf("task execution failed: %w", cmdErr)
	}

	return nil
}

// startWatchMode starts the TUI watch interface
func startWatchMode(cfg *config.Config) error {
	// Use the existing status command watch mode
	statusOpts := StatusOptions{
		Config: cfg,
		Watch:  true,
	}
	return RunStatus(statusOpts)
}

// waitForCleanup blocks until the cluster is cleaned up or the process is interrupted
func waitForCleanup(cfg *config.Config) error {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	// Create a channel to signal when cleanup is complete
	cleanupDone := make(chan bool, 1)
	
	// Start a goroutine to monitor for cleanup completion
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				// Reload config to check current status
				currentConfig, err := config.LoadConfig()
				if err != nil {
					continue
				}
				
				// Check if cluster has been destroyed
				if currentConfig.Status == "destroyed" {
					cleanupDone <- true
					return
				}
				
				// Check if we've passed the expiration time
				if time.Now().After(currentConfig.ExpiresAt) {
					// Give a bit more time for cleanup to complete
					time.Sleep(10 * time.Second)
					cleanupDone <- true
					return
				}
			}
		}
	}()
	
	// Display countdown timer
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				remaining := cfg.TimeRemaining()
				if remaining > 0 {
					fmt.Printf("%s Time remaining: %s\n", 
						tui.Icon("timer"), 
						tui.InfoValueStyle.Render(remaining.Round(time.Second).String()))
				}
			case <-cleanupDone:
				return
			}
		}
	}()
	
	// Wait for either cleanup completion or interrupt signal
	select {
	case <-cleanupDone:
		fmt.Printf("\n%s Cleanup completed successfully!\n", tui.Icon("success"))
		return nil
		
	case sig := <-sigChan:
		fmt.Printf("\n%s Received signal %v, exiting...\n", tui.Icon("warning"), sig)
		fmt.Printf("%s Scheduled cleanup will still occur at %s\n", 
			tui.Icon("info"), 
			tui.InfoValueStyle.Render(cfg.ExpiresAt.Format("15:04:05")))
		return nil
	}
}

