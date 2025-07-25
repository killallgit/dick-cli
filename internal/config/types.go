package config

import (
	"fmt"
	"time"
)

// GlobalConfig represents the global CLI configuration flags
type GlobalConfig struct {
	Verbose bool `mapstructure:"verbose" yaml:"verbose,omitempty"`
	Silent  bool `mapstructure:"silent" yaml:"silent,omitempty"`
}

// NewConfig represents configuration for the 'new' command
type NewConfig struct {
	TTL      string `mapstructure:"ttl" yaml:"ttl,omitempty"`
	Name     string `mapstructure:"name" yaml:"name,omitempty"`
	Provider string `mapstructure:"provider" yaml:"provider,omitempty"`
	Force    bool   `mapstructure:"force" yaml:"force,omitempty"`
}

// StatusConfig represents configuration for the 'status' command  
type StatusConfig struct {
	Watch bool `mapstructure:"watch" yaml:"watch,omitempty"`
}

// DestroyConfig represents configuration for the 'destroy' command
type DestroyConfig struct {
	Force bool `mapstructure:"force" yaml:"force,omitempty"`
}

// Config represents the complete application configuration with proper namespacing
type Config struct {
	// Command-specific configurations with proper namespacing
	Global     GlobalConfig  `mapstructure:"global" yaml:"global,omitempty"`
	New        NewConfig     `mapstructure:"new" yaml:"new,omitempty"`
	StatusCmd  StatusConfig  `mapstructure:"status_cmd" yaml:"status_cmd,omitempty"`
	Destroy    DestroyConfig `mapstructure:"destroy" yaml:"destroy,omitempty"`

	// Legacy fields for backward compatibility and state tracking
	// These will be populated from new.* fields when needed
	Provider string `mapstructure:"provider" yaml:"provider"`
	TTL      string `mapstructure:"ttl" yaml:"ttl"`
	Name     string `mapstructure:"name" yaml:"name"`
	Force    bool   `mapstructure:"force" yaml:"force,omitempty"`

	// State fields (unchanged)
	Status           string    `mapstructure:"status" yaml:"status,omitempty"`
	CreatedAt        time.Time `mapstructure:"created_at" yaml:"created_at,omitempty"`
	ExpiresAt        time.Time `mapstructure:"expires_at" yaml:"expires_at,omitempty"`
	ProjectPath      string    `mapstructure:"project_path" yaml:"project_path,omitempty"`
	CleanupAttempted bool      `mapstructure:"cleanup_attempted" yaml:"cleanup_attempted,omitempty"`
	ScheduledJobID   string    `mapstructure:"scheduled_job_id" yaml:"scheduled_job_id,omitempty"`
	
	// Enhanced cleanup tracking
	LastCleanupAttempt time.Time `mapstructure:"last_cleanup_attempt" yaml:"last_cleanup_attempt,omitempty"`
	CleanupAttempts    int       `mapstructure:"cleanup_attempts" yaml:"cleanup_attempts,omitempty"`
	LastCleanupError   string    `mapstructure:"last_cleanup_error" yaml:"last_cleanup_error,omitempty"`
}

// ParseTTL converts TTL string to duration
func (c *Config) ParseTTL() (time.Duration, error) {
	duration, err := time.ParseDuration(c.TTL)
	if err != nil {
		return 0, err
	}
	return duration, nil
}

// SetActive marks the cluster as active with proper timestamps
func (c *Config) SetActive() error {
	duration, err := c.ParseTTL()
	if err != nil {
		return err
	}

	c.Status = "active"
	c.CreatedAt = time.Now()
	c.ExpiresAt = c.CreatedAt.Add(duration)
	c.CleanupAttempted = false
	
	return nil
}

// SetDestroyed marks the cluster as destroyed
func (c *Config) SetDestroyed() {
	c.Status = "destroyed"
	c.CleanupAttempted = true
	c.ScheduledJobID = ""
}

// IsExpired returns true if the cluster has expired
func (c *Config) IsExpired() bool {
	return c.Status == "active" && time.Now().After(c.ExpiresAt)
}

// TimeRemaining returns the time until expiration
func (c *Config) TimeRemaining() time.Duration {
	if c.Status != "active" {
		return 0
	}
	remaining := c.ExpiresAt.Sub(time.Now())
	if remaining < 0 {
		return 0
	}
	return remaining
}

// CheckExpiration returns if cluster is expired and how long it's been expired
func (c *Config) CheckExpiration() (expired bool, expiredSince time.Duration) {
	if c.Status != "active" {
		return false, 0
	}
	
	now := time.Now()
	if now.After(c.ExpiresAt) {
		return true, now.Sub(c.ExpiresAt)
	}
	
	return false, 0
}

// ShouldPromptDestroy determines if user should be prompted for cleanup
func (c *Config) ShouldPromptDestroy() bool {
	expired, _ := c.CheckExpiration()
	
	// Don't prompt if not expired, already destroyed, or cleanup attempted
	if !expired || c.Status != "active" || c.CleanupAttempted {
		return false
	}
	
	// Don't prompt if force mode is enabled
	if c.Force {
		return false
	}
	
	// Always prompt for manual cleanup when not in force mode
	return true
}

// ShouldAutoDestroy determines if automatic cleanup should happen
func (c *Config) ShouldAutoDestroy() bool {
	expired, _ := c.CheckExpiration()
	
	// Only auto-destroy if expired, active, and force mode enabled
	return expired && c.Status == "active" && c.Force && !c.CleanupAttempted
}

// ShouldAttemptCleanup determines if we should attempt any form of cleanup
func (c *Config) ShouldAttemptCleanup() bool {
	expired, _ := c.CheckExpiration()
	
	// Attempt cleanup if expired, active, and not already attempted
	return expired && c.Status == "active" && !c.CleanupAttempted
}

// MarkCleanupAttempted marks that cleanup has been attempted
func (c *Config) MarkCleanupAttempted() {
	c.CleanupAttempted = true
	c.LastCleanupAttempt = time.Now()
	c.CleanupAttempts++
}

// MarkCleanupFailed records a failed cleanup attempt with error details
func (c *Config) MarkCleanupFailed(err error) {
	c.MarkCleanupAttempted()
	if err != nil {
		c.LastCleanupError = err.Error()
	}
}

// MarkCleanupSuccessful records a successful cleanup attempt
func (c *Config) MarkCleanupSuccessful() {
	c.MarkCleanupAttempted()
	c.LastCleanupError = ""
}

// ShouldRetryCleanup determines if cleanup should be retried based on retry policy
func (c *Config) ShouldRetryCleanup() bool {
	expired, _ := c.CheckExpiration()
	
	// Don't retry if not expired or not active
	if !expired || c.Status != "active" {
		return false
	}
	
	// Don't retry if we've exceeded max attempts
	const maxCleanupAttempts = 3
	if c.CleanupAttempts >= maxCleanupAttempts {
		return false
	}
	
	// Don't retry if last attempt was too recent (exponential backoff)
	if !c.LastCleanupAttempt.IsZero() {
		timeSinceLastAttempt := time.Since(c.LastCleanupAttempt)
		minRetryInterval := time.Duration(c.CleanupAttempts+1) * 5 * time.Minute
		
		if timeSinceLastAttempt < minRetryInterval {
			return false
		}
	}
	
	return true
}

// GetCleanupStatus returns a human-readable cleanup status
func (c *Config) GetCleanupStatus() string {
	if c.CleanupAttempts == 0 {
		return "No cleanup attempts"
	}
	
	status := fmt.Sprintf("%d attempt(s)", c.CleanupAttempts)
	if !c.LastCleanupAttempt.IsZero() {
		status += fmt.Sprintf(", last: %s", c.LastCleanupAttempt.Format("2006-01-02 15:04:05"))
	}
	if c.LastCleanupError != "" {
		status += fmt.Sprintf(", error: %s", c.LastCleanupError)
	}
	
	return status
}

// SetScheduledJobID stores the OS scheduler job ID
func (c *Config) SetScheduledJobID(jobID string) {
	c.ScheduledJobID = jobID
}

// ClearScheduledJob removes the scheduled job ID
func (c *Config) ClearScheduledJob() {
	c.ScheduledJobID = ""
}

// SyncLegacyFields synchronizes between namespaced and legacy fields
// This ensures backward compatibility while transitioning to new structure
func (c *Config) SyncLegacyFields() {
	// Sync from legacy to namespaced (for loading old configs)
	if c.Provider != "" && c.New.Provider == "" {
		c.New.Provider = c.Provider
	}
	if c.TTL != "" && c.New.TTL == "" {
		c.New.TTL = c.TTL
	}
	if c.Name != "" && c.New.Name == "" {
		c.New.Name = c.Name
	}
	if c.Force && !c.New.Force {
		c.New.Force = c.Force
	}

	// Sync from namespaced to legacy (for current operations)
	if c.New.Provider != "" {
		c.Provider = c.New.Provider
	}
	if c.New.TTL != "" {
		c.TTL = c.New.TTL
	}
	if c.New.Name != "" {
		c.Name = c.New.Name
	}
	if c.New.Force {
		c.Force = c.New.Force
	}
}

// GetEffectiveNewConfig returns the effective new command configuration
// merging defaults, config file, and any overrides
func (c *Config) GetEffectiveNewConfig() NewConfig {
	c.SyncLegacyFields()
	return c.New
}

// GetEffectiveStatusConfig returns the effective status command configuration
func (c *Config) GetEffectiveStatusConfig() StatusConfig {
	return c.StatusCmd
}

// GetEffectiveDestroyConfig returns the effective destroy command configuration
func (c *Config) GetEffectiveDestroyConfig() DestroyConfig {
	return c.Destroy
}

// GetEffectiveGlobalConfig returns the effective global configuration
func (c *Config) GetEffectiveGlobalConfig() GlobalConfig {
	return c.Global
}