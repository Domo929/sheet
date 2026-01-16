package components

import (
	"github.com/charmbracelet/lipgloss"
)

// Button represents a clickable button.
type Button struct {
	Label    string
	Selected bool // This button's option is currently chosen
	Focused  bool // This button is the active one in navigation
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
		// Focused button (has input focus) - bright purple border
		style = style.
			Foreground(lipgloss.Color("170")).
			BorderForeground(lipgloss.Color("170")).
			Bold(true)
	} else if b.Selected {
		// Selected button (this option is chosen) - green indicator
		style = style.
			Foreground(lipgloss.Color("35")).
			BorderForeground(lipgloss.Color("35")).
			Bold(true)
	} else {
		// Default unselected button
		style = style.
			Foreground(lipgloss.Color("252")).
			BorderForeground(lipgloss.Color("240"))
	}

	return style.Render(b.Label)
}

// ButtonGroup represents a group of buttons.
type ButtonGroup struct {
	Buttons       []Button
	SelectedIndex int  // Which button is currently selected/chosen
	FocusedIndex  int  // Which button has navigation focus (when group is focused)
	GroupFocused  bool // Whether this button group has input focus
}

// NewButtonGroup creates a new button group.
func NewButtonGroup(labels ...string) ButtonGroup {
	buttons := make([]Button, len(labels))
	for i, label := range labels {
		buttons[i] = NewButton(label)
	}
	// First button starts as selected
	if len(buttons) > 0 {
		buttons[0].Selected = true
	}
	return ButtonGroup{
		Buttons:       buttons,
		SelectedIndex: 0,
		FocusedIndex:  0,
		GroupFocused:  false,
	}
}

// SetFocused sets whether this button group has input focus.
func (bg *ButtonGroup) SetFocused(focused bool) {
	bg.GroupFocused = focused
	// Update button states
	for i := range bg.Buttons {
		bg.Buttons[i].Focused = focused && (i == bg.FocusedIndex)
		bg.Buttons[i].Selected = (i == bg.SelectedIndex)
	}
}

// MoveLeft moves navigation focus to the left.
func (bg *ButtonGroup) MoveLeft() {
	if bg.FocusedIndex > 0 {
		bg.FocusedIndex--
		bg.SelectedIndex = bg.FocusedIndex
		bg.SetFocused(bg.GroupFocused) // Refresh button states
	}
}

// MoveRight moves navigation focus to the right.
func (bg *ButtonGroup) MoveRight() {
	if bg.FocusedIndex < len(bg.Buttons)-1 {
		bg.FocusedIndex++
		bg.SelectedIndex = bg.FocusedIndex
		bg.SetFocused(bg.GroupFocused) // Refresh button states
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
