package config

import "github.com/spf13/viper"

// SetDefaults establishes default configuration values using modern namespaced approach
func SetDefaults(v *viper.Viper) {
	// Global defaults
	v.SetDefault("global.verbose", false)
	v.SetDefault("global.silent", false)
	
	// New command defaults  
	v.SetDefault("new.provider", "kind")
	v.SetDefault("new.ttl", "5m")
	v.SetDefault("new.name", "dev-cluster")
	v.SetDefault("new.force", false)
	
	// Status command defaults
	v.SetDefault("status_cmd.watch", false)
	
	// Destroy command defaults
	v.SetDefault("destroy.force", false)
	
	// Legacy defaults for backward compatibility
	v.SetDefault("provider", "kind")
	v.SetDefault("ttl", "5m")
	v.SetDefault("name", "dev-cluster")
	v.SetDefault("force", false)
	
	// State defaults (these are usually set at runtime)
	v.SetDefault("status", "")
	v.SetDefault("cleanup_attempted", false)
	v.SetDefault("scheduled_job_id", "")
}