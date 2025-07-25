package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// GlobalViper is the global viper instance used throughout the app
	GlobalViper *viper.Viper
)

// Initialize sets up the global Viper configuration with modern best practices
func Initialize(cfgFile string) error {
	if GlobalViper == nil {
		GlobalViper = viper.New()
	}

	// Set up environment variable support with better patterns
	GlobalViper.SetEnvPrefix("DICK")
	GlobalViper.AutomaticEnv()
	// Replace dots with underscores in env var names (modern practice)
	GlobalViper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Set up config file discovery with proper search paths
	if cfgFile != "" {
		// Use config file from the flag
		GlobalViper.SetConfigFile(cfgFile)
	} else {
		// Search for config in multiple standard locations
		GlobalViper.SetConfigName(".dick")
		GlobalViper.SetConfigType("yaml")
		
		// Add config search paths in order of precedence
		GlobalViper.AddConfigPath(".")                    // Current directory
		if home, err := os.UserHomeDir(); err == nil {
			GlobalViper.AddConfigPath(home)               // Home directory
			GlobalViper.AddConfigPath(filepath.Join(home, ".config", "dick")) // XDG config
		}
		GlobalViper.AddConfigPath("/etc/dick")           // System-wide config
	}

	// Set defaults first
	SetDefaults(GlobalViper)

	// Read config file with improved error handling
	if err := GlobalViper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; create with defaults if we're in a project directory
			return createDefaultConfigFile()
		}
		return fmt.Errorf("error reading config file: %w", err)
	}

	return nil
}

// LoadConfig loads and returns the current configuration with improved error handling
func LoadConfig() (*Config, error) {
	if GlobalViper == nil {
		if err := Initialize(""); err != nil {
			return nil, fmt.Errorf("failed to initialize config: %w", err)
		}
	}

	var config Config
	if err := GlobalViper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Sync between namespaced and legacy fields for compatibility
	config.SyncLegacyFields()

	// Set project path if not already set
	if config.ProjectPath == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
		config.ProjectPath = pwd
	}

	return &config, nil
}

// NewIsolatedViper creates a new isolated Viper instance for testing or specific use cases
// Following modern Viper best practices for multiple instances
func NewIsolatedViper() *viper.Viper {
	v := viper.New()
	v.SetEnvPrefix("DICK")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	SetDefaults(v)
	return v
}

// SaveConfig writes the configuration back to file
func SaveConfig(config *Config) error {
	if GlobalViper == nil {
		return fmt.Errorf("viper not initialized")
	}

	// Set project path if not already set
	if config.ProjectPath == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		config.ProjectPath = pwd
	}

	// Update viper with current config values
	GlobalViper.Set("provider", config.Provider)
	GlobalViper.Set("ttl", config.TTL)
	GlobalViper.Set("name", config.Name)
	GlobalViper.Set("force", config.Force)
	GlobalViper.Set("status", config.Status)
	GlobalViper.Set("created_at", config.CreatedAt)
	GlobalViper.Set("expires_at", config.ExpiresAt)
	GlobalViper.Set("project_path", config.ProjectPath)
	GlobalViper.Set("cleanup_attempted", config.CleanupAttempted)
	GlobalViper.Set("scheduled_job_id", config.ScheduledJobID)
	
	// Enhanced cleanup tracking fields
	GlobalViper.Set("last_cleanup_attempt", config.LastCleanupAttempt)
	GlobalViper.Set("cleanup_attempts", config.CleanupAttempts)
	GlobalViper.Set("last_cleanup_error", config.LastCleanupError)

	// Write to config file
	configFile := GlobalViper.ConfigFileUsed()
	if configFile == "" {
		// Default to .dick.yaml in current directory
		configFile = ".dick.yaml"
	}

	return GlobalViper.WriteConfigAs(configFile)
}

// createDefaultConfigFile creates a default .dick.yaml file
func createDefaultConfigFile() error {
	configPath := ".dick.yaml"
	
	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // File already exists
	}

	// Create default config using viper defaults
	if err := GlobalViper.SafeWriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to create default config file: %w", err)
	}

	return nil
}

// GetConfigFilePath returns the path to the active config file
func GetConfigFilePath() string {
	if GlobalViper == nil {
		return ""
	}
	
	configFile := GlobalViper.ConfigFileUsed()
	if configFile == "" {
		// Default to .dick.yaml in current directory
		configFile = ".dick.yaml"
	}
	
	abs, err := filepath.Abs(configFile)
	if err != nil {
		return configFile
	}
	
	return abs
}

// ApplyFlagOverrides applies command line flag values to config with improved patterns
func ApplyFlagOverrides(config *Config, provider, ttl, name *string, force *bool) {
	// Apply non-empty string values using modern pattern
	if provider != nil && *provider != "" {
		config.Provider = *provider
	}
	if ttl != nil && *ttl != "" {
		config.TTL = *ttl
	}
	if name != nil && *name != "" {
		config.Name = *name
	}
	if force != nil {
		config.Force = *force
	}
}

// BindGlobalFlags binds global persistent flags to Viper with proper namespacing
func BindGlobalFlags(v *viper.Viper, cmd interface{}) error {
	// Type assertion to get cobra.Command
	cobraCmd, ok := cmd.(*cobra.Command)
	if !ok {
		return fmt.Errorf("invalid command type, expected *cobra.Command")
	}

	// Bind global flags with namespace
	if flag := cobraCmd.PersistentFlags().Lookup("verbose"); flag != nil {
		if err := v.BindPFlag("global.verbose", flag); err != nil {
			return fmt.Errorf("failed to bind verbose flag: %w", err)
		}
	}
	
	if flag := cobraCmd.PersistentFlags().Lookup("silent"); flag != nil {
		if err := v.BindPFlag("global.silent", flag); err != nil {
			return fmt.Errorf("failed to bind silent flag: %w", err)
		}
	}

	return nil
}

// BindNewFlags binds 'new' command flags to Viper with proper namespacing
func BindNewFlags(v *viper.Viper, cmd interface{}) error {
	cobraCmd, ok := cmd.(*cobra.Command)
	if !ok {
		return fmt.Errorf("invalid command type, expected *cobra.Command")
	}

	// Bind new command flags with namespace
	if flag := cobraCmd.Flags().Lookup("ttl"); flag != nil {
		if err := v.BindPFlag("new.ttl", flag); err != nil {
			return fmt.Errorf("failed to bind ttl flag: %w", err)
		}
	}
	
	if flag := cobraCmd.Flags().Lookup("name"); flag != nil {
		if err := v.BindPFlag("new.name", flag); err != nil {
			return fmt.Errorf("failed to bind name flag: %w", err)
		}
	}
	
	if flag := cobraCmd.Flags().Lookup("provider"); flag != nil {
		if err := v.BindPFlag("new.provider", flag); err != nil {
			return fmt.Errorf("failed to bind provider flag: %w", err)
		}
	}
	
	if flag := cobraCmd.Flags().Lookup("force"); flag != nil {
		if err := v.BindPFlag("new.force", flag); err != nil {
			return fmt.Errorf("failed to bind force flag: %w", err)
		}
	}

	return nil
}

// BindStatusFlags binds 'status' command flags to Viper with proper namespacing
func BindStatusFlags(v *viper.Viper, cmd interface{}) error {
	cobraCmd, ok := cmd.(*cobra.Command)
	if !ok {
		return fmt.Errorf("invalid command type, expected *cobra.Command")
	}

	// Bind status command flags with namespace
	if flag := cobraCmd.Flags().Lookup("watch"); flag != nil {
		if err := v.BindPFlag("status_cmd.watch", flag); err != nil {
			return fmt.Errorf("failed to bind watch flag: %w", err)
		}
	}

	return nil
}

// BindDestroyFlags binds 'destroy' command flags to Viper with proper namespacing
func BindDestroyFlags(v *viper.Viper, cmd interface{}) error {
	cobraCmd, ok := cmd.(*cobra.Command)
	if !ok {
		return fmt.Errorf("invalid command type, expected *cobra.Command")
	}

	// Bind destroy command flags with namespace
	if flag := cobraCmd.Flags().Lookup("force"); flag != nil {
		if err := v.BindPFlag("destroy.force", flag); err != nil {
			return fmt.Errorf("failed to bind force flag: %w", err)
		}
	}

	return nil
}

// ValidateTTL validates a TTL string format
func ValidateTTL(ttl string) error {
	if ttl == "" {
		return nil // Empty TTL is allowed (will use default)
	}
	
	_, err := time.ParseDuration(ttl)
	if err != nil {
		return fmt.Errorf("invalid TTL format '%s': %w (examples: 5m, 1h, 30s)", ttl, err)
	}
	
	return nil
}

// ValidateProvider validates a provider name
func ValidateProvider(provider string) error {
	if provider == "" {
		return nil // Empty provider is allowed (will use default)
	}
	
	validProviders := []string{"kind", "tofu"}
	for _, valid := range validProviders {
		if provider == valid {
			return nil
		}
	}
	
	return fmt.Errorf("unsupported provider '%s' (supported: %s)", provider, strings.Join(validProviders, ", "))
}

// ValidateName validates an environment name
func ValidateName(name string) error {
	if name == "" {
		return nil // Empty name is allowed (will use default)
	}
	
	// Basic validation - should be DNS compatible for Kubernetes
	if len(name) > 63 {
		return fmt.Errorf("name too long (max 63 characters): %s", name)
	}
	
	// Should start and end with alphanumeric
	matched, err := regexp.MatchString(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`, name)
	if err != nil {
		return fmt.Errorf("error validating name pattern: %w", err)
	}
	if !matched {
		return fmt.Errorf("invalid name '%s': must start and end with alphanumeric, contain only lowercase letters, numbers, and hyphens", name)
	}
	
	return nil
}

// ValidateConfig validates the current configuration
func ValidateConfig(config *Config) error {
	// Validate TTL
	if err := ValidateTTL(config.TTL); err != nil {
		return err
	}
	
	// Validate provider
	if err := ValidateProvider(config.Provider); err != nil {
		return err
	}
	
	// Validate name
	if err := ValidateName(config.Name); err != nil {
		return err
	}
	
	return nil
}