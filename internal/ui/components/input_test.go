package components

import (
	"strings"
	"testing"
)

func TestTextInputCursor(t *testing.T) {
	input := NewTextInput("Name", "Enter name")

	// Unfocused - no cursor
	input.SetFocused(false)
	input.SetValue("John")
	rendered := input.Render()

	if strings.Contains(rendered, "▌") {
		t.Error("Unfocused input should not show cursor")
	}

	// Focused - cursor should appear
	input.SetFocused(true)
	rendered = input.Render()

	if !strings.Contains(rendered, "▌") {
		t.Error("Focused input should show cursor")
	}

	// Verify cursor appears after the value
	if !strings.Contains(rendered, "John") {
		t.Error("Input should contain the value 'John'")
	}
}

func TestTextInputCursorWithEmptyValue(t *testing.T) {
	input := NewTextInput("Name", "Enter name")
	input.SetFocused(true)
	input.SetValue("")

	rendered := input.Render()

	// Should NOT show cursor when empty (only placeholder)
	if strings.Contains(rendered, "▌") {
		t.Error("Focused empty input should not show cursor, only placeholder")
	}

	// Should show placeholder
	if !strings.Contains(rendered, "Enter name") {
		t.Error("Empty input should show placeholder")
	}
}

func TestTextInputCursorAfterTyping(t *testing.T) {
	input := NewTextInput("Name", "Enter name")
	input.SetFocused(true)

	// Type some text
	input.SetValue("G")
	rendered := input.Render()
	if !strings.Contains(rendered, "▌") {
		t.Error("Cursor should appear after typing")
	}

	input.SetValue("Ga")
	rendered = input.Render()
	if !strings.Contains(rendered, "▌") {
		t.Error("Cursor should appear after typing more")
	}

	input.SetValue("Gandalf")
	rendered = input.Render()
	if !strings.Contains(rendered, "Gandalf") {
		t.Error("Should show full typed text")
	}
	if !strings.Contains(rendered, "▌") {
		t.Error("Cursor should appear at the end")
	}
}
