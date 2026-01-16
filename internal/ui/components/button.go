package components

import (
	"github.com/charmbracelet/lipgloss"
)

// Button represents a clickable button.
type Button struct {
	Label    string
	Focused  bool
	Disabled bool
	Action   interface{} // The message to send when clicked
}

// NewButton creates a new button.
func NewButton(label string) Button {
	return Button{
		Label: label,
	}
}

// Render renders the button.
func (b Button) Render() string {
	style := lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.RoundedBorder())

	if b.Disabled {
		style = style.
			Foreground(lipgloss.Color("240")).
			BorderForeground(lipgloss.Color("240"))
	} else if b.Focused {
		style = style.
			Foreground(lipgloss.Color("170")).
			BorderForeground(lipgloss.Color("170")).
			Bold(true)
	} else {
		style = style.
			Foreground(lipgloss.Color("252")).
			BorderForeground(lipgloss.Color("240"))
	}

	return style.Render(b.Label)
}

// ButtonGroup represents a group of buttons.
type ButtonGroup struct {
	Buttons       []Button
	SelectedIndex int
}

// NewButtonGroup creates a new button group.
func NewButtonGroup(labels ...string) ButtonGroup {
	buttons := make([]Button, len(labels))
	for i, label := range labels {
		buttons[i] = NewButton(label)
	}
	if len(buttons) > 0 {
		buttons[0].Focused = true
	}
	return ButtonGroup{
		Buttons:       buttons,
		SelectedIndex: 0,
	}
}

// MoveLeft moves selection to the left.
func (bg *ButtonGroup) MoveLeft() {
	if bg.SelectedIndex > 0 {
		bg.Buttons[bg.SelectedIndex].Focused = false
		bg.SelectedIndex--
		bg.Buttons[bg.SelectedIndex].Focused = true
	}
}

// MoveRight moves selection to the right.
func (bg *ButtonGroup) MoveRight() {
	if bg.SelectedIndex < len(bg.Buttons)-1 {
		bg.Buttons[bg.SelectedIndex].Focused = false
		bg.SelectedIndex++
		bg.Buttons[bg.SelectedIndex].Focused = true
	}
}

// Selected returns the currently selected button.
func (bg *ButtonGroup) Selected() *Button {
	if bg.SelectedIndex >= 0 && bg.SelectedIndex < len(bg.Buttons) {
		return &bg.Buttons[bg.SelectedIndex]
	}
	return nil
}

// Render renders the button group.
func (bg ButtonGroup) Render() string {
	if len(bg.Buttons) == 0 {
		return ""
	}

	rendered := make([]string, len(bg.Buttons))
	for i, button := range bg.Buttons {
		rendered[i] = button.Render()
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, rendered...)
}
