package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTextInputCursor(t *testing.T) {
	input := NewTextInput("Name", "Enter name")

	// Unfocused - no cursor
	input.SetFocused(false)
	input.SetValue("John")
	rendered := input.Render()

	assert.NotContains(t, rendered, "▌", "Unfocused input should not show cursor")

	// Focused - cursor should appear
	input.SetFocused(true)
	rendered = input.Render()

	assert.Contains(t, rendered, "▌", "Focused input should show cursor")

	// Verify cursor appears after the value
	assert.Contains(t, rendered, "John", "Input should contain the value 'John'")
}

func TestTextInputCursorWithEmptyValue(t *testing.T) {
	input := NewTextInput("Name", "Enter name")
	input.SetFocused(true)
	input.SetValue("")

	rendered := input.Render()

	// Should NOT show cursor when empty (only placeholder)
	assert.NotContains(t, rendered, "▌", "Focused empty input should not show cursor, only placeholder")

	// Should show placeholder
	assert.Contains(t, rendered, "Enter name", "Empty input should show placeholder")
}

func TestTextInputCursorAfterTyping(t *testing.T) {
	input := NewTextInput("Name", "Enter name")
	input.SetFocused(true)

	// Type some text
	input.SetValue("G")
	rendered := input.Render()
	assert.Contains(t, rendered, "▌", "Cursor should appear after typing")

	input.SetValue("Ga")
	rendered = input.Render()
	assert.Contains(t, rendered, "▌", "Cursor should appear after typing more")

	input.SetValue("Gandalf")
	rendered = input.Render()
	assert.Contains(t, rendered, "Gandalf", "Should show full typed text")
	assert.Contains(t, rendered, "▌", "Cursor should appear at the end")
}
