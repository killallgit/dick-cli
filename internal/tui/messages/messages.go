package messages

import (
	"time"
)

// ViewType represents the type of view
type ViewType int

const (
	StatusView ViewType = iota
	MonitorView
	ConfirmView
	HelpView
	SettingsView
)

// String returns the string representation of the view type
func (vt ViewType) String() string {
	switch vt {
	case StatusView:
		return "Status"
	case MonitorView:
		return "Monitor"
	case ConfirmView:
		return "Confirm"
	case HelpView:
		return "Help"
	case SettingsView:
		return "Settings"
	default:
		return "Unknown"
	}
}

// TickMsg is sent periodically for updates
type TickMsg struct {
	Time time.Time
}

// NavigateMsg is sent when the user wants to switch views
type NavigateMsg struct {
	To ViewType
}

// ErrorMsg is sent when an error occurs
type ErrorMsg struct {
	Err error
}

// ConfigReloadMsg is sent when config should be reloaded
type ConfigReloadMsg struct{}

// ConfirmResultMsg is sent when confirmation dialog completes
type ConfirmResultMsg struct {
	Confirmed bool
}

// EventMsg is sent when a new event occurs for monitoring
type EventMsg struct {
	Message string
	Time    time.Time
}