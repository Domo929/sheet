package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ListItem represents an item in a list.
type ListItem struct {
	Title       string
	Description string
	Value       any
}

// List is a selectable list component.
type List struct {
	Items         []ListItem
	SelectedIndex int
	ScrollOffset  int
	Width         int
	Height        int
	Title         string
}

// NewList creates a new list component.
func NewList(title string, items []ListItem) List {
	return List{
		Title:         title,
		Items:         items,
		SelectedIndex: 0,
	}
}

// MoveUp moves the selection up.
func (l *List) MoveUp() {
	if l.SelectedIndex > 0 {
		l.SelectedIndex--
		l.ensureVisible()
	}
}

// MoveDown moves the selection down.
func (l *List) MoveDown() {
	if l.SelectedIndex < len(l.Items)-1 {
		l.SelectedIndex++
		l.ensureVisible()
	}
}

// Selected returns the currently selected item.
func (l *List) Selected() *ListItem {
	if l.SelectedIndex < 0 || l.SelectedIndex >= len(l.Items) {
		return nil
	}
	return &l.Items[l.SelectedIndex]
}

// visibleItemCount returns how many items can be displayed given the current Height.
// Returns len(Items) if Height is 0 (no pagination).
func (l *List) visibleItemCount() int {
	if l.Height <= 0 {
		return len(l.Items)
	}

	available := l.Height

	// Title takes 2 lines (title text + blank line)
	if l.Title != "" {
		available -= 2
	}

	if available < 1 {
		available = 1
	}

	if available > len(l.Items) {
		available = len(l.Items)
	}

	return available
}

// ensureVisible adjusts ScrollOffset so that SelectedIndex is within the visible window.
func (l *List) ensureVisible() {
	if l.Height <= 0 {
		return
	}

	visible := l.visibleItemCount()

	// If selected is above the visible window, scroll up
	if l.SelectedIndex < l.ScrollOffset {
		l.ScrollOffset = l.SelectedIndex
	}

	// If selected is below the visible window, scroll down
	if l.SelectedIndex >= l.ScrollOffset+visible {
		l.ScrollOffset = l.SelectedIndex - visible + 1
	}

	// Clamp scroll offset
	maxOffset := len(l.Items) - visible
	if maxOffset < 0 {
		maxOffset = 0
	}
	if l.ScrollOffset > maxOffset {
		l.ScrollOffset = maxOffset
	}
	if l.ScrollOffset < 0 {
		l.ScrollOffset = 0
	}
}

// Render renders the list as a string.
func (l List) Render() string {
	if len(l.Items) == 0 {
		return "No items"
	}

	// ensureVisible on a copy since Render has a value receiver
	l.ensureVisible()

	var b strings.Builder

	// Add title if present
	if l.Title != "" {
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99"))
		b.WriteString(titleStyle.Render(l.Title))
		b.WriteString("\n\n")
	}

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	indicatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))

	// Determine visible range
	visibleCount := l.visibleItemCount()
	startIdx := l.ScrollOffset
	endIdx := startIdx + visibleCount
	if endIdx > len(l.Items) {
		endIdx = len(l.Items)
	}

	// "More above" indicator
	if startIdx > 0 {
		b.WriteString(indicatorStyle.Render(fmt.Sprintf("  ↑ %d more", startIdx)))
		b.WriteString("\n")
	}

	for i := startIdx; i < endIdx; i++ {
		item := l.Items[i]
		cursor := "  "
		style := normalStyle

		if i == l.SelectedIndex {
			cursor = "> "
			style = selectedStyle
		}

		line := fmt.Sprintf("%s%s", cursor, item.Title)
		if item.Description != "" {
			line = fmt.Sprintf("%s - %s", line, item.Description)
		}

		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	// "More below" indicator
	if endIdx < len(l.Items) {
		b.WriteString(indicatorStyle.Render(fmt.Sprintf("  ↓ %d more", len(l.Items)-endIdx)))
		b.WriteString("\n")
	}

	return b.String()
}

// SetSize sets the dimensions of the list.
func (l *List) SetSize(width, height int) {
	l.Width = width
	l.Height = height
}
