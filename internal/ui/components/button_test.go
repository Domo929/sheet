package components

import (
	"strings"
	"testing"
)

func TestNewButton(t *testing.T) {
	btn := NewButton("Click Me")

	if btn.Label != "Click Me" {
		t.Errorf("Label = %s, want Click Me", btn.Label)
	}
	if btn.Focused {
		t.Error("New button should not be focused by default")
	}
	if btn.Disabled {
		t.Error("New button should not be disabled by default")
	}
}

func TestButtonRender(t *testing.T) {
	btn := NewButton("Test")

	rendered := btn.Render()
	if !strings.Contains(rendered, "Test") {
		t.Error("Rendered button should contain label")
	}
}

func TestNewButtonGroup(t *testing.T) {
	bg := NewButtonGroup("OK", "Cancel", "Help")

	if len(bg.Buttons) != 3 {
		t.Errorf("len(Buttons) = %d, want 3", len(bg.Buttons))
	}
	if bg.SelectedIndex != 0 {
		t.Errorf("SelectedIndex = %d, want 0", bg.SelectedIndex)
	}
	if !bg.Buttons[0].Focused {
		t.Error("First button should be focused")
	}
}

func TestButtonGroupNavigation(t *testing.T) {
	bg := NewButtonGroup("One", "Two", "Three")

	// Test moving right
	bg.MoveRight()
	if bg.SelectedIndex != 1 {
		t.Errorf("After MoveRight: SelectedIndex = %d, want 1", bg.SelectedIndex)
	}
	if bg.Buttons[0].Focused {
		t.Error("First button should not be focused after moving right")
	}
	if !bg.Buttons[1].Focused {
		t.Error("Second button should be focused after moving right")
	}

	bg.MoveRight()
	if bg.SelectedIndex != 2 {
		t.Errorf("After second MoveRight: SelectedIndex = %d, want 2", bg.SelectedIndex)
	}

	// Test boundary - shouldn't go past last button
	bg.MoveRight()
	if bg.SelectedIndex != 2 {
		t.Errorf("After third MoveRight: SelectedIndex = %d, want 2 (boundary)", bg.SelectedIndex)
	}

	// Test moving left
	bg.MoveLeft()
	if bg.SelectedIndex != 1 {
		t.Errorf("After MoveLeft: SelectedIndex = %d, want 1", bg.SelectedIndex)
	}

	bg.MoveLeft()
	if bg.SelectedIndex != 0 {
		t.Errorf("After second MoveLeft: SelectedIndex = %d, want 0", bg.SelectedIndex)
	}

	// Test boundary - shouldn't go below 0
	bg.MoveLeft()
	if bg.SelectedIndex != 0 {
		t.Errorf("After third MoveLeft: SelectedIndex = %d, want 0 (boundary)", bg.SelectedIndex)
	}
}

func TestButtonGroupSelected(t *testing.T) {
	bg := NewButtonGroup("First", "Second")

	selected := bg.Selected()
	if selected == nil {
		t.Fatal("Selected() returned nil")
	}
	if selected.Label != "First" {
		t.Errorf("Selected().Label = %s, want First", selected.Label)
	}

	bg.MoveRight()
	selected = bg.Selected()
	if selected.Label != "Second" {
		t.Errorf("After MoveRight: Selected().Label = %s, want Second", selected.Label)
	}
}

func TestButtonGroupRender(t *testing.T) {
	bg := NewButtonGroup("OK", "Cancel")

	rendered := bg.Render()

	// Should contain both button labels
	if !strings.Contains(rendered, "OK") {
		t.Error("Rendered output should contain OK button")
	}
	if !strings.Contains(rendered, "Cancel") {
		t.Error("Rendered output should contain Cancel button")
	}
}

func TestButtonGroupEmpty(t *testing.T) {
	bg := ButtonGroup{Buttons: []Button{}}

	selected := bg.Selected()
	if selected != nil {
		t.Error("Selected() should return nil for empty button group")
	}

	rendered := bg.Render()
	if rendered != "" {
		t.Error("Rendered output should be empty for empty button group")
	}
}
