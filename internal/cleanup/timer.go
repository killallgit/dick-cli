package cleanup

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/killallgit/dick/internal/common"
	"github.com/killallgit/dick/internal/config"
)

// StartTTLTimer starts a background timer that will cleanup the cluster when TTL expires
func StartTTLTimer(cfg *config.Config) error {
	if cfg.Status != "active" {
		return fmt.Errorf("cluster is not active, cannot start TTL timer")
	}

	duration := cfg.TimeRemaining()
	if duration <= 0 {
		return fmt.Errorf("cluster has already expired")
	}

	// Start the timer in a goroutine
	go func() {
		log.Printf("TTL timer started: cluster will be destroyed in %v", duration)
		
		// Wait for the TTL to expire
		time.Sleep(duration)
		
		// Perform cleanup
		if err := performCleanup(cfg); err != nil {
			log.Printf("Failed to cleanup cluster: %v", err)
		} else {
			log.Printf("Cluster '%s' destroyed after TTL expiration", cfg.Name)
		}
	}()

	return nil
}

// performCleanup executes the cleanup task and updates state
func performCleanup(cfg *config.Config) error {
	// Get the project directory where .dick.yaml is located
	projectDir := cfg.ProjectPath
	if projectDir == "" {
		// Fallback to current directory
		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		projectDir = pwd
	}

	// Execute the destroy task using provider-specific taskfile
	var taskFile string
	switch cfg.Provider {
	case "kind":
		taskFile = filepath.Join(projectDir, "tasks", "Taskfile.k8s.yaml")
	default:
		// Fallback to legacy taskfile
		taskFile = filepath.Join(projectDir, "tasks", "Taskfile.new.yaml")
	}
	
	// Check if new taskfile exists, fallback to legacy if needed
	if _, err := os.Stat(taskFile); err != nil {
		// Fallback to legacy taskfile for backward compatibility
		taskFile = filepath.Join(projectDir, "tasks", "Taskfile.new.yaml")
	}
	
	if err := executeDestroyTask(taskFile, cfg.Name, cfg.Provider); err != nil {
		return fmt.Errorf("failed to execute destroy task: %w", err)
	}

	// Update the config to mark as destroyed
	cfg.SetDestroyed()
	
	// Change to project directory to save config
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	
	if err := os.Chdir(projectDir); err != nil {
		return fmt.Errorf("failed to change to project directory: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	return nil
}

// executeDestroyTask runs the kind:destroy task using the task command
func executeDestroyTask(taskFile, clusterName, provider string) error {
	// Check if task command is available
	if _, err := exec.LookPath("task"); err != nil {
		return fmt.Errorf("task command not found: %w", err)
	}

	// Execute: task -t <taskfile> hook:teardown CLUSTER_NAME=<name>
	taskArgs := []string{"-t", taskFile}
	
	// Add --silent flag by default unless verbose mode is enabled
	if !common.ShouldShowTaskOutput() {
		taskArgs = append(taskArgs, "--silent")
	}
	
	// Use standardized hook:teardown task, fallback to provider-specific task
	var taskName string
	if strings.Contains(taskFile, "Taskfile.k8s.yaml") {
		taskName = "hook:teardown"
	} else {
		// Legacy taskfile - use provider-specific task name
		switch provider {
		case "kind":
			taskName = "kind:destroy"
		default:
			taskName = "hook:teardown" // Try standardized hook as fallback
		}
	}
	
	taskArgs = append(taskArgs, taskName, fmt.Sprintf("CLUSTER_NAME=%s", clusterName))
	cmd := exec.Command("task", taskArgs...)
	
	// Set the working directory to the project directory
	cmd.Dir = filepath.Dir(filepath.Dir(taskFile))
	
	// Always capture output for background cleanup operations
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("task execution failed: %w, output: %s", err, string(output))
	}

	// Log output only in verbose mode for background operations
	if common.ShouldShowTaskOutput() {
		log.Printf("Task output: %s", string(output))
	}
	return nil
}

// ForceCleanup immediately performs cleanup without waiting for TTL
func ForceCleanup(cfg *config.Config) error {
	if cfg.Status != "active" {
		return fmt.Errorf("cluster is not active")
	}

	return performCleanup(cfg)
}