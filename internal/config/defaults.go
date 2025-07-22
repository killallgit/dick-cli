package config

import "github.com/spf13/viper"

// SetDefaults establishes default configuration values
func SetDefaults(v *viper.Viper) {
	// Core configuration defaults
	v.SetDefault("provider", "kind")
	v.SetDefault("ttl", "5m")
	v.SetDefault("name", "dev-cluster")
	v.SetDefault("force", false)
	
	// State defaults (these are usually set at runtime)
	v.SetDefault("status", "")
	v.SetDefault("cleanup_attempted", false)
	v.SetDefault("scheduled_job_id", "")
}