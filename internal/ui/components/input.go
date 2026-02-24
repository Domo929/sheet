package components

import (
	"charm.land/lipgloss/v2"
)

// TextInput represents a simple text input field.
type TextInput struct {
	Label       string
	Placeholder string
	Value       string
	Width       int
	Focused     bool
}

// NewTextInput creates a new text input.
func NewTextInput(label, placeholder string) TextInput {
	return TextInput{
		Label:       label,
		Placeholder: placeholder,
		Width:       40,
	}
}

// SetValue sets the input value.
func (t *TextInput) SetValue(value string) {
	t.Value = value
}

// SetFocused sets the focus state.
func (t *TextInput) SetFocused(focused bool) {
	t.Focused = focused
}

// Render renders the text input.
func (t TextInput) Render() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99")).
		Bold(true)

	label := labelStyle.Render(t.Label + ":")

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	if t.Focused {
		valueStyle = valueStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("170")).
			Padding(0, 1)
	} else {
		valueStyle = valueStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)
	}

	if t.Width > 0 {
		valueStyle = valueStyle.Width(t.Width)
	}

	displayValue := t.Value

	// Add cursor when focused and there's actual text (not just placeholder)
	if t.Focused && t.Value != "" {
		cursorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)
		displayValue += cursorStyle.Render("▌")
	}

	// Show placeholder if empty
	if t.Value == "" && t.Placeholder != "" {
		placeholderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		displayValue = placeholderStyle.Render(t.Placeholder)
	}

	value := valueStyle.Render(displayValue)

	return label + "\n" + value
}
