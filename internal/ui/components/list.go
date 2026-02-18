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
	}
}

// MoveDown moves the selection down.
func (l *List) MoveDown() {
	if l.SelectedIndex < len(l.Items)-1 {
		l.SelectedIndex++
	}
}

// Selected returns the currently selected item.
func (l *List) Selected() *ListItem {
	if l.SelectedIndex < 0 || l.SelectedIndex >= len(l.Items) {
		return nil
	}
	return &l.Items[l.SelectedIndex]
}

// Render renders the list as a string.
func (l List) Render() string {
	if len(l.Items) == 0 {
		return "No items"
	}

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

	for i, item := range l.Items {
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

	return b.String()
}

// SetSize sets the dimensions of the list.
func (l *List) SetSize(width, height int) {
	l.Width = width
	l.Height = height
}
