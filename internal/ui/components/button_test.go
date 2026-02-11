package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewButton(t *testing.T) {
	btn := NewButton("Click Me")

	assert.Equal(t, "Click Me", btn.Label)
	assert.False(t, btn.Focused, "New button should not be focused by default")
	assert.False(t, btn.Disabled, "New button should not be disabled by default")
}

func TestButtonRender(t *testing.T) {
	btn := NewButton("Test")

	rendered := btn.Render()
	assert.Contains(t, rendered, "Test", "Rendered button should contain label")
}

func TestNewButtonGroup(t *testing.T) {
	bg := NewButtonGroup("OK", "Cancel", "Help")

	assert.Equal(t, 3, len(bg.Buttons))
	assert.Equal(t, 0, bg.SelectedIndex)
	assert.True(t, bg.Buttons[0].Selected, "First button should be selected")
}

func TestButtonGroupNavigation(t *testing.T) {
	bg := NewButtonGroup("One", "Two", "Three")
	bg.SetFocused(true) // Set group as focused for navigation

	// Test moving right
	bg.MoveRight()
	assert.Equal(t, 1, bg.SelectedIndex, "After MoveRight")
	assert.False(t, bg.Buttons[0].Focused, "First button should not be focused after moving right")
	assert.True(t, bg.Buttons[1].Focused, "Second button should be focused after moving right")
	assert.True(t, bg.Buttons[1].Selected, "Second button should be selected after moving right")

	bg.MoveRight()
	assert.Equal(t, 2, bg.SelectedIndex, "After second MoveRight")

	// Test boundary - shouldn't go past last button
	bg.MoveRight()
	assert.Equal(t, 2, bg.SelectedIndex, "After third MoveRight (boundary)")

	// Test moving left
	bg.MoveLeft()
	assert.Equal(t, 1, bg.SelectedIndex, "After MoveLeft")

	bg.MoveLeft()
	assert.Equal(t, 0, bg.SelectedIndex, "After second MoveLeft")

	// Test boundary - shouldn't go below 0
	bg.MoveLeft()
	assert.Equal(t, 0, bg.SelectedIndex, "After third MoveLeft (boundary)")
}

func TestButtonGroupSelected(t *testing.T) {
	bg := NewButtonGroup("First", "Second")

	selected := bg.Selected()
	require.NotNil(t, selected, "Selected() returned nil")
	assert.Equal(t, "First", selected.Label)

	bg.MoveRight()
	selected = bg.Selected()
	assert.Equal(t, "Second", selected.Label, "After MoveRight")
}

func TestButtonGroupRender(t *testing.T) {
	bg := NewButtonGroup("OK", "Cancel")

	rendered := bg.Render()

	// Should contain both button labels
	assert.Contains(t, rendered, "OK", "Rendered output should contain OK button")
	assert.Contains(t, rendered, "Cancel", "Rendered output should contain Cancel button")
}

func TestButtonGroupEmpty(t *testing.T) {
	bg := ButtonGroup{Buttons: []Button{}}

	selected := bg.Selected()
	assert.Nil(t, selected, "Selected() should return nil for empty button group")

	rendered := bg.Render()
	assert.Empty(t, rendered, "Rendered output should be empty for empty button group")
}

func TestButtonGroupFocusedVsSelected(t *testing.T) {
	bg := NewButtonGroup("XP Tracking", "Milestone")

	// Initially: first button selected, group not focused
	assert.True(t, bg.Buttons[0].Selected, "First button should be selected initially")
	assert.False(t, bg.Buttons[0].Focused, "Buttons should not be focused when group is not focused")

	// Set group as focused (like when user tabs to it)
	bg.SetFocused(true)
	assert.True(t, bg.Buttons[0].Focused, "First button should be focused when group is focused")
	assert.True(t, bg.Buttons[0].Selected, "First button should still be selected")

	// Set group as not focused (like when user tabs away)
	bg.SetFocused(false)
	assert.False(t, bg.Buttons[0].Focused, "Button should not be focused when group loses focus")
	assert.True(t, bg.Buttons[0].Selected, "Button should remain selected even when group loses focus")

	// Move to second button and unfocus
	bg.SetFocused(true)
	bg.MoveRight()
	bg.SetFocused(false)

	// First button should no longer be selected
	assert.False(t, bg.Buttons[0].Selected, "First button should not be selected after moving right")
	// Second button should be selected but not focused
	assert.True(t, bg.Buttons[1].Selected, "Second button should be selected")
	assert.False(t, bg.Buttons[1].Focused, "Second button should not be focused when group is not focused")
}
