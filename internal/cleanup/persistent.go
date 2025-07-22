package cleanup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/killallgit/dick/internal/config"
)

// ScheduleCleanup schedules OS-level cleanup at expiration time
func ScheduleCleanup(cfg *config.Config) error {
	if cfg.Status != "active" {
		return fmt.Errorf("cluster is not active, cannot schedule cleanup")
	}

	// Validate expiration time is in the future
	if cfg.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("expiration time is in the past: %s", cfg.ExpiresAt.Format(time.RFC3339))
	}

	// Get the absolute path to the dick executable
	dickPath, err := getDickExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get dick executable path: %w", err)
	}

	// Verify the executable exists and is executable
	if stat, err := os.Stat(dickPath); err != nil {
		return fmt.Errorf("dick executable not found at %s: %w", dickPath, err)
	} else if stat.Mode()&0111 == 0 {
		return fmt.Errorf("dick executable is not executable: %s", dickPath)
	}

	// Get project directory for context
	projectDir := cfg.ProjectPath
	if projectDir == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		projectDir = pwd
	}

	// Verify project directory exists
	if _, err := os.Stat(projectDir); err != nil {
		return fmt.Errorf("project directory does not exist: %s", projectDir)
	}

	// Cancel any existing scheduled job first
	if err := CancelScheduledCleanup(cfg); err != nil {
		fmt.Printf("Warning: failed to cancel existing scheduled job: %v\n", err)
	}

	// Schedule the cleanup based on OS
	jobID, err := scheduleOSCleanup(dickPath, projectDir, cfg.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to schedule OS cleanup: %w", err)
	}

	// Store the job ID in config
	cfg.SetScheduledJobID(jobID)
	
	// Verify the job was actually scheduled
	exists, err := CheckScheduledCleanup(cfg)
	if err != nil {
		fmt.Printf("Warning: failed to verify scheduled job: %v\n", err)
	} else if !exists {
		return fmt.Errorf("scheduled job was not created successfully")
	}
	
	return nil
}

// CancelScheduledCleanup removes the scheduled cleanup job
func CancelScheduledCleanup(cfg *config.Config) error {
	if cfg.ScheduledJobID == "" {
		return nil // No job to cancel
	}

	err := cancelOSCleanup(cfg.ScheduledJobID)
	if err != nil {
		// Don't fail hard on cleanup cancellation errors
		// Job might have already run or been removed
		fmt.Printf("Warning: failed to cancel scheduled cleanup: %v\n", err)
	}

	cfg.ClearScheduledJob()
	return nil
}

// CheckScheduledCleanup verifies if cleanup job still exists
func CheckScheduledCleanup(cfg *config.Config) (bool, error) {
	if cfg.ScheduledJobID == "" {
		return false, nil
	}

	exists, err := checkOSCleanupExists(cfg.ScheduledJobID)
	if err != nil {
		return false, fmt.Errorf("failed to check scheduled cleanup: %w", err)
	}

	return exists, nil
}

// getDickExecutablePath returns the absolute path to the dick executable
func getDickExecutablePath() (string, error) {
	// First try to get the current executable path
	exePath, err := os.Executable()
	if err != nil {
		// Fallback: try to find dick in PATH
		exePath, err = exec.LookPath("dick")
		if err != nil {
			return "", fmt.Errorf("could not find dick executable: %w", err)
		}
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(exePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return absPath, nil
}

// scheduleOSCleanup schedules cleanup using OS-specific methods
func scheduleOSCleanup(dickPath, projectDir string, expireTime time.Time) (string, error) {
	switch runtime.GOOS {
	case "darwin", "linux":
		return scheduleWithAt(dickPath, projectDir, expireTime)
	case "windows":
		return scheduleWithSchtasks(dickPath, projectDir, expireTime)
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// scheduleWithAt schedules cleanup using the 'at' command (macOS/Linux)
func scheduleWithAt(dickPath, projectDir string, expireTime time.Time) (string, error) {
	// Check if 'at' command is available
	if _, err := exec.LookPath("at"); err != nil {
		return "", fmt.Errorf("'at' command not found (required for background cleanup): %w", err)
	}

	// Validate scheduling time is in the future (with at least 1 minute buffer)
	now := time.Now()
	if expireTime.Before(now.Add(time.Minute)) {
		return "", fmt.Errorf("expiration time must be at least 1 minute in the future: %s", expireTime.Format(time.RFC3339))
	}

	// Format time for 'at' command: "HH:MM MM/DD/YY"
	atTime := expireTime.Format("15:04 01/02/06")
	
	// Create the command to run: cd to project dir and run destroy with force flag
	// Use absolute paths and add error handling
	command := fmt.Sprintf("cd %s && %s destroy --force 2>&1 | logger -t dick-cleanup", 
		shellEscape(projectDir), 
		shellEscape(dickPath))

	// Execute: echo "command" | at time
	cmd := exec.Command("at", atTime)
	cmd.Stdin = strings.NewReader(command)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to schedule with 'at' command at time %s: %w, output: %s", atTime, err, string(output))
	}

	// Extract job ID from output (format: "job 123 at ...")
	jobID := extractAtJobID(string(output))
	if jobID == "" {
		return "", fmt.Errorf("failed to extract job ID from 'at' output: %s", string(output))
	}

	// Log successful scheduling
	fmt.Printf("Successfully scheduled cleanup job %s at %s\n", jobID, expireTime.Format("2006-01-02 15:04:05"))

	return jobID, nil
}

// scheduleWithSchtasks schedules cleanup using 'schtasks' (Windows)
func scheduleWithSchtasks(dickPath, projectDir string, expireTime time.Time) (string, error) {
	// Check if 'schtasks' command is available
	if _, err := exec.LookPath("schtasks"); err != nil {
		return "", fmt.Errorf("'schtasks' command not found: %w", err)
	}

	// Generate unique task name
	taskName := fmt.Sprintf("dick-cleanup-%d", time.Now().Unix())
	
	// Format time and date for schtasks
	scheduleTime := expireTime.Format("15:04")
	scheduleDate := expireTime.Format("01/02/2006")
	
	// Create command
	command := fmt.Sprintf(`cmd /c "cd /d %s && %s destroy --force"`, 
		projectDir, dickPath)

	// Create scheduled task
	cmd := exec.Command("schtasks", "/create", "/tn", taskName, "/tr", command,
		"/sc", "once", "/st", scheduleTime, "/sd", scheduleDate)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to schedule with 'schtasks': %w, output: %s", err, string(output))
	}

	return taskName, nil
}

// cancelOSCleanup cancels scheduled cleanup using OS-specific methods
func cancelOSCleanup(jobID string) error {
	switch runtime.GOOS {
	case "darwin", "linux":
		return cancelAtJob(jobID)
	case "windows":
		return cancelSchtask(jobID)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// cancelAtJob cancels an 'at' job
func cancelAtJob(jobID string) error {
	cmd := exec.Command("atrm", jobID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to cancel 'at' job %s: %w, output: %s", jobID, err, string(output))
	}
	return nil
}

// cancelSchtask cancels a scheduled task
func cancelSchtask(taskName string) error {
	cmd := exec.Command("schtasks", "/delete", "/tn", taskName, "/f")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to cancel scheduled task %s: %w, output: %s", taskName, err, string(output))
	}
	return nil
}

// checkOSCleanupExists checks if scheduled cleanup still exists
func checkOSCleanupExists(jobID string) (bool, error) {
	switch runtime.GOOS {
	case "darwin", "linux":
		return checkAtJobExists(jobID)
	case "windows":
		return checkSchtaskExists(jobID)
	default:
		return false, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// checkAtJobExists checks if an 'at' job exists
func checkAtJobExists(jobID string) (bool, error) {
	cmd := exec.Command("atq")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to list 'at' jobs: %w", err)
	}

	// Check if job ID appears in the output
	return strings.Contains(string(output), jobID), nil
}

// checkSchtaskExists checks if a scheduled task exists
func checkSchtaskExists(taskName string) (bool, error) {
	cmd := exec.Command("schtasks", "/query", "/tn", taskName)
	err := cmd.Run()
	
	// If command succeeds, task exists
	return err == nil, nil
}

// extractAtJobID extracts job ID from 'at' command output
func extractAtJobID(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "job ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1] // Return the job number
			}
		}
	}
	return ""
}

// shellEscape escapes a string for safe shell execution
func shellEscape(s string) string {
	// Simple shell escaping - wrap in single quotes and escape any single quotes
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}