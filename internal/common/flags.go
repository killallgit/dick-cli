package common

// Global flag values set by the root command
var (
	VerboseFlag bool
	SilentFlag  bool
)

// IsVerbose returns true if verbose output is enabled
func IsVerbose() bool {
	return VerboseFlag
}

// IsSilent returns true if silent output is explicitly requested
func IsSilent() bool {
	return SilentFlag
}

// ShouldShowTaskOutput determines if task command output should be shown
// Default behavior: silent unless --verbose is passed
// Explicit --silent overrides --verbose
func ShouldShowTaskOutput() bool {
	if SilentFlag {
		return false
	}
	return VerboseFlag
}