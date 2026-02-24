package components

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Panel creates a bordered panel with a title.
type Panel struct {
	Title   string
	Content string
	Width   int
	Height  int
	Style   lipgloss.Style
}

// DefaultPanelStyle returns the default style for panels.
func DefaultPanelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)
}

// NewPanel creates a new panel with default styling.
func NewPanel(title, content string, width, height int) Panel {
	return Panel{
		Title:   title,
		Content: content,
		Width:   width,
		Height:  height,
		Style:   DefaultPanelStyle(),
	}
}

// Render renders the panel as a string.
func (p Panel) Render() string {
	// Set dimensions if specified
	style := p.Style
	if p.Width > 0 {
		style = style.Width(p.Width)
	}
	if p.Height > 0 {
		style = style.Height(p.Height)
	}

	// Add title if present
	content := p.Content
	if p.Title != "" {
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			UnderlineStyle(lipgloss.UnderlineCurly).
			UnderlineColor(lipgloss.Color("99"))
		title := titleStyle.Render(p.Title)
		content = title + "\n\n" + content
	}

	return style.Render(content)
}

// Box creates a simple bordered box around content.
func Box(content string, width int) string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	if width > 0 {
		style = style.Width(width)
	}

	return style.Render(content)
}

// JoinHorizontal joins multiple strings horizontally with spacing.
func JoinHorizontal(gap int, strs ...string) string {
	spacer := strings.Repeat(" ", gap)
	return lipgloss.JoinHorizontal(lipgloss.Top, strs[0], spacer, strings.Join(strs[1:], spacer))
}

// JoinVertical joins multiple strings vertically.
func JoinVertical(strs ...string) string {
	return lipgloss.JoinVertical(lipgloss.Left, strs...)
}
