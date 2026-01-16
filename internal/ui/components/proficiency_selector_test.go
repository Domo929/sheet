package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewProficiencySelector(t *testing.T) {
	options := []string{"Option 1", "Option 2", "Option 3"}
	ps := NewProficiencySelector("Test Selector", options, 2)

	if ps.title != "Test Selector" {
		t.Errorf("expected title 'Test Selector', got '%s'", ps.title)
	}

	if len(ps.options) != 3 {
		t.Errorf("expected 3 options, got %d", len(ps.options))
	}

	if ps.maxSelect != 2 {
		t.Errorf("expected maxSelect 2, got %d", ps.maxSelect)
	}

	if ps.SelectedCount() != 0 {
		t.Errorf("expected 0 selections initially, got %d", ps.SelectedCount())
	}

	if !ps.focused {
		t.Error("selector should be focused by default")
	}
}

func TestProficiencySelectorNavigation(t *testing.T) {
	options := []string{"Option 1", "Option 2", "Option 3"}
	ps := NewProficiencySelector("Test", options, 2)

	// Initial cursor position
	if ps.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", ps.cursor)
	}

	// Move down
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyDown})
	if ps.cursor != 1 {
		t.Errorf("expected cursor at 1 after down, got %d", ps.cursor)
	}

	// Move down again
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyDown})
	if ps.cursor != 2 {
		t.Errorf("expected cursor at 2 after second down, got %d", ps.cursor)
	}

	// Try to move beyond end
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyDown})
	if ps.cursor != 2 {
		t.Errorf("cursor should stay at 2 at bottom, got %d", ps.cursor)
	}

	// Move up
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyUp})
	if ps.cursor != 1 {
		t.Errorf("expected cursor at 1 after up, got %d", ps.cursor)
	}

	// Move to top
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyUp})
	if ps.cursor != 0 {
		t.Errorf("expected cursor at 0 after second up, got %d", ps.cursor)
	}

	// Try to move beyond start
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyUp})
	if ps.cursor != 0 {
		t.Errorf("cursor should stay at 0 at top, got %d", ps.cursor)
	}
}

func TestProficiencySelectorSelection(t *testing.T) {
	options := []string{"Option 1", "Option 2", "Option 3", "Option 4"}
	ps := NewProficiencySelector("Test", options, 2)

	// Select first option
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeySpace})
	if ps.SelectedCount() != 1 {
		t.Errorf("expected 1 selection, got %d", ps.SelectedCount())
	}

	if !ps.selected[0] {
		t.Error("expected option 0 to be selected")
	}

	// Move to second and select
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyDown})
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeySpace})
	if ps.SelectedCount() != 2 {
		t.Errorf("expected 2 selections, got %d", ps.SelectedCount())
	}

	// Try to select a third (should not work, maxSelect is 2)
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyDown})
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeySpace})
	if ps.SelectedCount() != 2 {
		t.Errorf("expected 2 selections (maxed out), got %d", ps.SelectedCount())
	}

	// Deselect first option
	ps.cursor = 0
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeySpace})
	if ps.SelectedCount() != 1 {
		t.Errorf("expected 1 selection after deselect, got %d", ps.SelectedCount())
	}

	if ps.selected[0] {
		t.Error("expected option 0 to be deselected")
	}
}

func TestProficiencySelectorEnterKey(t *testing.T) {
	options := []string{"Option 1", "Option 2"}
	ps := NewProficiencySelector("Test", options, 1)

	// Test enter key also works for selection
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ps.SelectedCount() != 1 {
		t.Errorf("expected 1 selection with enter key, got %d", ps.SelectedCount())
	}
}

func TestProficiencySelectorGetters(t *testing.T) {
	options := []string{"Skill A", "Skill B", "Skill C"}
	ps := NewProficiencySelector("Test", options, 2)

	// Select options 0 and 2
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeySpace})
	ps.cursor = 2
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeySpace})

	// Test GetSelectedIndices
	indices := ps.GetSelectedIndices()
	if len(indices) != 2 {
		t.Errorf("expected 2 indices, got %d", len(indices))
	}

	// Test GetSelectedOptions
	selected := ps.GetSelectedOptions()
	if len(selected) != 2 {
		t.Errorf("expected 2 options, got %d", len(selected))
	}

	// Should contain "Skill A" and "Skill C"
	found := make(map[string]bool)
	for _, s := range selected {
		found[s] = true
	}

	if !found["Skill A"] {
		t.Error("expected 'Skill A' in selected options")
	}

	if !found["Skill C"] {
		t.Error("expected 'Skill C' in selected options")
	}
}

func TestProficiencySelectorSetSelected(t *testing.T) {
	options := []string{"Option 1", "Option 2", "Option 3"}
	ps := NewProficiencySelector("Test", options, 2)

	// Set indices 0 and 2 as selected
	ps.SetSelected([]int{0, 2})

	if ps.SelectedCount() != 2 {
		t.Errorf("expected 2 selections after SetSelected, got %d", ps.SelectedCount())
	}

	if !ps.selected[0] || !ps.selected[2] {
		t.Error("expected indices 0 and 2 to be selected")
	}

	if ps.selected[1] {
		t.Error("expected index 1 to not be selected")
	}
}

func TestProficiencySelectorIsComplete(t *testing.T) {
	options := []string{"Option 1", "Option 2", "Option 3"}
	ps := NewProficiencySelector("Test", options, 2)

	if ps.IsComplete() {
		t.Error("selector should not be complete initially")
	}

	// Select one
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeySpace})
	if ps.IsComplete() {
		t.Error("selector should not be complete with 1/2 selected")
	}

	// Select second
	ps.cursor = 1
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !ps.IsComplete() {
		t.Error("selector should be complete with 2/2 selected")
	}
}

func TestProficiencySelectorFocus(t *testing.T) {
	options := []string{"Option 1", "Option 2"}
	ps := NewProficiencySelector("Test", options, 1)

	if !ps.IsFocused() {
		t.Error("selector should be focused by default")
	}

	ps.SetFocused(false)
	if ps.IsFocused() {
		t.Error("selector should not be focused after SetFocused(false)")
	}

	// When unfocused, keys should not work
	initialCursor := ps.cursor
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyDown})
	if ps.cursor != initialCursor {
		t.Error("cursor should not move when unfocused")
	}

	// Re-focus
	ps.SetFocused(true)
	if !ps.IsFocused() {
		t.Error("selector should be focused after SetFocused(true)")
	}

	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyDown})
	if ps.cursor == initialCursor {
		t.Error("cursor should move when focused")
	}
}

func TestProficiencySelectorViKeys(t *testing.T) {
	options := []string{"Option 1", "Option 2", "Option 3"}
	ps := NewProficiencySelector("Test", options, 2)

	// Test 'j' for down
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if ps.cursor != 1 {
		t.Errorf("expected cursor at 1 with 'j', got %d", ps.cursor)
	}

	// Test 'k' for up
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if ps.cursor != 0 {
		t.Errorf("expected cursor at 0 with 'k', got %d", ps.cursor)
	}
}

func TestProficiencySelectorDimensions(t *testing.T) {
	options := []string{"Option 1", "Option 2"}
	ps := NewProficiencySelector("Test", options, 1)

	ps.SetWidth(100)
	if ps.width != 100 {
		t.Errorf("expected width 100, got %d", ps.width)
	}

	ps.SetHeight(50)
	if ps.height != 50 {
		t.Errorf("expected height 50, got %d", ps.height)
	}
}
