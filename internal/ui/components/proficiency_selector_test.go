package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestNewProficiencySelector(t *testing.T) {
	options := []string{"Option 1", "Option 2", "Option 3"}
	ps := NewProficiencySelector("Test Selector", options, 2)

	assert.Equal(t, "Test Selector", ps.title)
	assert.Equal(t, 3, len(ps.options))
	assert.Equal(t, 2, ps.maxSelect)
	assert.Equal(t, 0, ps.SelectedCount(), "expected 0 selections initially")
	assert.True(t, ps.focused, "selector should be focused by default")
}

func TestProficiencySelectorNavigation(t *testing.T) {
	options := []string{"Option 1", "Option 2", "Option 3"}
	ps := NewProficiencySelector("Test", options, 2)

	// Initial cursor position
	assert.Equal(t, 0, ps.cursor, "initial cursor position")

	// Move down
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, ps.cursor, "cursor after down")

	// Move down again
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 2, ps.cursor, "cursor after second down")

	// Try to move beyond end
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 2, ps.cursor, "cursor should stay at 2 at bottom")

	// Move up
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 1, ps.cursor, "cursor after up")

	// Move to top
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 0, ps.cursor, "cursor after second up")

	// Try to move beyond start
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 0, ps.cursor, "cursor should stay at 0 at top")
}

func TestProficiencySelectorSelection(t *testing.T) {
	options := []string{"Option 1", "Option 2", "Option 3", "Option 4"}
	ps := NewProficiencySelector("Test", options, 2)

	// Select first option
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeySpace})
	assert.Equal(t, 1, ps.SelectedCount())
	assert.True(t, ps.selected[0], "expected option 0 to be selected")

	// Move to second and select
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyDown})
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeySpace})
	assert.Equal(t, 2, ps.SelectedCount())

	// Try to select a third (should not work, maxSelect is 2)
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyDown})
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeySpace})
	assert.Equal(t, 2, ps.SelectedCount(), "expected 2 selections (maxed out)")

	// Deselect first option
	ps.cursor = 0
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeySpace})
	assert.Equal(t, 1, ps.SelectedCount(), "expected 1 selection after deselect")
	assert.False(t, ps.selected[0], "expected option 0 to be deselected")
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
	assert.Equal(t, 2, len(indices))

	// Test GetSelectedOptions
	selected := ps.GetSelectedOptions()
	assert.Equal(t, 2, len(selected))

	// Should contain "Skill A" and "Skill C"
	assert.Contains(t, selected, "Skill A")
	assert.Contains(t, selected, "Skill C")
}

func TestProficiencySelectorSetSelected(t *testing.T) {
	options := []string{"Option 1", "Option 2", "Option 3"}
	ps := NewProficiencySelector("Test", options, 2)

	// Set indices 0 and 2 as selected
	ps.SetSelected([]int{0, 2})

	assert.Equal(t, 2, ps.SelectedCount(), "expected 2 selections after SetSelected")
	assert.True(t, ps.selected[0], "expected index 0 to be selected")
	assert.True(t, ps.selected[2], "expected index 2 to be selected")
	assert.False(t, ps.selected[1], "expected index 1 to not be selected")
}

func TestProficiencySelectorIsComplete(t *testing.T) {
	options := []string{"Option 1", "Option 2", "Option 3"}
	ps := NewProficiencySelector("Test", options, 2)

	assert.False(t, ps.IsComplete(), "selector should not be complete initially")

	// Select one
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeySpace})
	assert.False(t, ps.IsComplete(), "selector should not be complete with 1/2 selected")

	// Select second
	ps.cursor = 1
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeySpace})
	assert.True(t, ps.IsComplete(), "selector should be complete with 2/2 selected")
}

func TestProficiencySelectorFocus(t *testing.T) {
	options := []string{"Option 1", "Option 2"}
	ps := NewProficiencySelector("Test", options, 1)

	assert.True(t, ps.IsFocused(), "selector should be focused by default")

	ps.SetFocused(false)
	assert.False(t, ps.IsFocused(), "selector should not be focused after SetFocused(false)")

	// When unfocused, keys should not work
	initialCursor := ps.cursor
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, initialCursor, ps.cursor, "cursor should not move when unfocused")

	// Re-focus
	ps.SetFocused(true)
	assert.True(t, ps.IsFocused(), "selector should be focused after SetFocused(true)")

	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.NotEqual(t, initialCursor, ps.cursor, "cursor should move when focused")
}

func TestProficiencySelectorViKeys(t *testing.T) {
	options := []string{"Option 1", "Option 2", "Option 3"}
	ps := NewProficiencySelector("Test", options, 2)

	// Test 'j' for down
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	assert.Equal(t, 1, ps.cursor, "cursor with 'j'")

	// Test 'k' for up
	ps, _ = ps.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	assert.Equal(t, 0, ps.cursor, "cursor with 'k'")
}

func TestProficiencySelectorDimensions(t *testing.T) {
	options := []string{"Option 1", "Option 2"}
	ps := NewProficiencySelector("Test", options, 1)

	ps.SetWidth(100)
	assert.Equal(t, 100, ps.width)

	ps.SetHeight(50)
	assert.Equal(t, 50, ps.height)
}
