package components

import (
	"fmt"
	
	"github.com/killallgit/dick/internal/styles"
)

// Header represents a reusable header component
type Header struct {
	Title    string
	Icon     string
	Subtitle string
	Width    int
}

// NewHeader creates a new header
func NewHeader(title, icon string) *Header {
	return &Header{
		Title: title,
		Icon:  icon,
	}
}

// SetSubtitle sets the subtitle
func (h *Header) SetSubtitle(subtitle string) *Header {
	h.Subtitle = subtitle
	return h
}

// Render returns the rendered header
func (h *Header) Render() string {
	title := h.Title
	if h.Icon != "" {
		iconText := styles.Icon(h.Icon)
		if iconText != "" {
			title = fmt.Sprintf("%s %s", iconText, title)
		}
	}
	
	header := styles.HeaderStyle.Render(title)
	
	if h.Subtitle != "" {
		header += "\n" + styles.InfoLabelStyle.Render(h.Subtitle)
	}
	
	// Add divider
	dividerWidth := 50
	if h.Width > 0 && h.Width < dividerWidth {
		dividerWidth = h.Width - 4
	}
	header += "\n" + styles.Divider(dividerWidth)
	
	return header
}

// SetWidth sets the header width
func (h *Header) SetWidth(width int) {
	h.Width = width
}