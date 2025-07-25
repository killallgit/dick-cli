package styles

import (
	"fmt"
	"strings"
	"time"
	
	"github.com/charmbracelet/lipgloss"
)

var (
	// Base16 color palette for better terminal compatibility
	ColorPrimary   = lipgloss.Color("1")   // Red
	ColorSecondary = lipgloss.Color("6")   // Cyan  
	ColorSuccess   = lipgloss.Color("2")   // Green
	ColorWarning   = lipgloss.Color("3")   // Yellow
	ColorDanger    = lipgloss.Color("1")   // Red
	ColorMuted     = lipgloss.Color("8")   // Bright Black/Gray
	ColorDark      = lipgloss.Color("0")   // Black
	ColorLight     = lipgloss.Color("7")   // White
)

var (
	// Base styles - simplified
	BaseStyle = lipgloss.NewStyle()

	// Header styles - minimal
	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	// Title styles - minimal
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)

	// Status styles
	StatusActiveStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true)

	StatusDestroyedStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Bold(true)

	StatusExpiredStyle = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true)

	StatusUnknownStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Bold(true)

	// Info styles
	InfoLabelStyle = lipgloss.NewStyle().
			Foreground(ColorDark).
			Bold(true)

	InfoValueStyle = lipgloss.NewStyle().
			Foreground(ColorDark)

	// Progress styles
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess)

	ProgressTextStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)

	// Border styles - simplified
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder())

	// Button styles - minimal
	ButtonStyle = lipgloss.NewStyle().
			Foreground(ColorLight).
			Background(ColorPrimary).
			Bold(true)

	ButtonSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorLight).
				Background(ColorSecondary).
				Bold(true)

	// Error styles
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true)

	// Success styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)
	
	// Warning styles
	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)
)

// Divider creates a styled separator line
func Divider(width int) string {
	divider := ""
	for i := 0; i < width; i++ {
		divider += "━"
	}
	return lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render(divider)
}

// Icon returns simple text indicators for different states
func Icon(iconType string) string {
	icons := map[string]string{
		"cluster":   "CLUSTER",
		"time":      "TIME",
		"timer":     "TIMER",
		"success":   "OK",
		"error":     "ERROR",
		"warning":   "WARN",
		"destroyed": "DESTROYED",
		"active":    "OK",
		"expired":   "EXPIRED",
		"unknown":   "UNKNOWN",
		"project":   "PROJECT",
		"name":      "NAME",
		"ttl":       "TTL",
		"created":   "CREATED",
		"expires":   "EXPIRES",
		"remaining": "REMAINING",
		"destroy":   "DESTROY",
		"info":      "INFO",
	}
	
	if icon, exists := icons[iconType]; exists {
		return icon
	}
	return ""
}

// FormatDuration creates a human-readable duration string with color
func FormatDuration(duration string) string {
	return InfoValueStyle.Render(duration)
}

// FormatStatus returns a styled status string
func FormatStatus(status string) string {
	switch status {
	case "active":
		return StatusActiveStyle.Render("ACTIVE")
	case "destroyed":
		return StatusDestroyedStyle.Render("DESTROYED")
	case "expired":
		return StatusExpiredStyle.Render("EXPIRED")
	default:
		return StatusUnknownStyle.Render("UNKNOWN")
	}
}

// RenderProgressBar creates a progress bar for TTL remaining time
func RenderProgressBar(name string, remaining, total time.Duration, width int) string {
	if width <= 0 {
		width = 20
	}
	
	elapsed := total - remaining
	progress := float64(elapsed) / float64(total)
	
	// Ensure progress is between 0 and 1
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	
	filled := int(progress * float64(width))
	if filled > width {
		filled = width
	}
	
	bar := "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
	
	return fmt.Sprintf("%s %s %s", 
		name,
		ProgressBarStyle.Render(bar),
		ProgressTextStyle.Render(remaining.String()))
}