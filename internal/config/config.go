package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var (
	// GlobalViper is the global viper instance used throughout the app
	GlobalViper *viper.Viper
)

// Initialize sets up the global Viper configuration
func Initialize(cfgFile string) error {
	if GlobalViper == nil {
		GlobalViper = viper.New()
	}

	// Set up environment variable support
	GlobalViper.SetEnvPrefix("DICK")
	GlobalViper.AutomaticEnv()

	// Set up config file
	if cfgFile != "" {
		// Use config file from the flag
		GlobalViper.SetConfigFile(cfgFile)
	} else {
		// Search for config in current directory and home directory
		GlobalViper.AddConfigPath(".")
		GlobalViper.AddConfigPath("$HOME")
		GlobalViper.SetConfigType("yaml")
		GlobalViper.SetConfigName(".dick")
	}

	// Set defaults
	SetDefaults(GlobalViper)

	// Read config file if it exists
	if err := GlobalViper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; create with defaults if we're in a project directory
			return createDefaultConfigFile()
		}
		return fmt.Errorf("error reading config file: %w", err)
	}

	return nil
}

// LoadConfig loads and returns the current configuration
func LoadConfig() (*Config, error) {
	if GlobalViper == nil {
		if err := Initialize(""); err != nil {
			return nil, err
		}
	}

	var config Config
	if err := GlobalViper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

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

// ApplyFlagOverrides applies command line flag values to config
func ApplyFlagOverrides(config *Config, provider, ttl, name *string, force *bool) {
	// provider parameter is now optional (can be nil)
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