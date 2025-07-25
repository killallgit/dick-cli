package components

import (
	"fmt"
	"strings"
	"time"
	
	"github.com/killallgit/dick/internal/styles"
)

// Event represents a single event in the event log
type Event struct {
	Message   string
	Timestamp time.Time
}

// EventLog represents a reusable event log component
type EventLog struct {
	events     []Event
	maxEvents  int
	maxDisplay int
}

// NewEventLog creates a new event log
func NewEventLog(maxEvents, maxDisplay int) *EventLog {
	if maxEvents <= 0 {
		maxEvents = 20
	}
	if maxDisplay <= 0 {
		maxDisplay = 5
	}
	
	return &EventLog{
		events:     []Event{},
		maxEvents:  maxEvents,
		maxDisplay: maxDisplay,
	}
}

// Add adds a new event to the log
func (e *EventLog) Add(message string) {
	event := Event{
		Message:   message,
		Timestamp: time.Now(),
	}
	
	e.events = append(e.events, event)
	
	// Keep only last maxEvents
	if len(e.events) > e.maxEvents {
		e.events = e.events[len(e.events)-e.maxEvents:]
	}
}

// Clear clears all events
func (e *EventLog) Clear() {
	e.events = []Event{}
}

// GetEvents returns the most recent events for display
func (e *EventLog) GetEvents() []Event {
	if len(e.events) <= e.maxDisplay {
		return e.events
	}
	
	return e.events[len(e.events)-e.maxDisplay:]
}

// Render returns the rendered event log
func (e *EventLog) Render() string {
	lines := []string{
		styles.TitleStyle.Render("Events"),
	}
	
	displayEvents := e.GetEvents()
	
	if len(displayEvents) == 0 {
		lines = append(lines, "  No events yet")
	} else {
		for _, event := range displayEvents {
			eventLine := fmt.Sprintf("  [%s] %s", 
				event.Timestamp.Format("15:04:05"),
				event.Message)
			lines = append(lines, eventLine)
		}
	}
	
	return strings.Join(lines, "\n")
}

// Count returns the total number of events
func (e *EventLog) Count() int {
	return len(e.events)
}