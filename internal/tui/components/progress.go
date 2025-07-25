package components

import (
	"fmt"
	"strings"
	"time"
	
	"github.com/killallgit/dick/internal/styles"
)

// ProgressBar represents a reusable progress bar component
type ProgressBar struct {
	Width    int
	Elapsed  time.Duration
	Total    time.Duration
	ShowTime bool
}

// NewProgressBar creates a new progress bar
func NewProgressBar(elapsed, total time.Duration, width int) *ProgressBar {
	return &ProgressBar{
		Width:    width,
		Elapsed:  elapsed,
		Total:    total,
		ShowTime: true,
	}
}

// Render returns the rendered progress bar
func (p *ProgressBar) Render() string {
	if p.Total <= 0 {
		return ""
	}
	
	progress := float64(p.Elapsed) / float64(p.Total)
	
	// Ensure progress is between 0 and 1
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	
	// Calculate bar width
	barWidth := p.Width
	if barWidth <= 0 {
		barWidth = 30
	}
	
	filled := int(progress * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	percentage := int(progress * 100)
	
	result := styles.ProgressBarStyle.Render(bar)
	
	if p.ShowTime {
		remaining := p.Total - p.Elapsed
		result = fmt.Sprintf("%s %s %d%%", 
			result,
			styles.ProgressTextStyle.Render(remaining.Round(time.Second).String()),
			percentage)
	} else {
		result = fmt.Sprintf("%s %d%%", result, percentage)
	}
	
	return result
}

// SetWidth updates the progress bar width
func (p *ProgressBar) SetWidth(width int) {
	p.Width = width
}

// Update updates the elapsed time
func (p *ProgressBar) Update(elapsed time.Duration) {
	p.Elapsed = elapsed
}