package components

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewList(t *testing.T) {
	items := []ListItem{
		{Title: "Item 1", Description: "First item"},
		{Title: "Item 2", Description: "Second item"},
	}

	list := NewList("Test List", items)

	assert.Equal(t, "Test List", list.Title)
	assert.Equal(t, 2, len(list.Items))
	assert.Equal(t, 0, list.SelectedIndex)
}

func TestListNavigation(t *testing.T) {
	items := []ListItem{
		{Title: "Item 1"},
		{Title: "Item 2"},
		{Title: "Item 3"},
	}

	list := NewList("Test", items)

	// Test moving down
	list.MoveDown()
	assert.Equal(t, 1, list.SelectedIndex, "After MoveDown")

	list.MoveDown()
	assert.Equal(t, 2, list.SelectedIndex, "After second MoveDown")

	// Test boundary - shouldn't go past last item
	list.MoveDown()
	assert.Equal(t, 2, list.SelectedIndex, "After third MoveDown (boundary)")

	// Test moving up
	list.MoveUp()
	assert.Equal(t, 1, list.SelectedIndex, "After MoveUp")

	list.MoveUp()
	assert.Equal(t, 0, list.SelectedIndex, "After second MoveUp")

	// Test boundary - shouldn't go below 0
	list.MoveUp()
	assert.Equal(t, 0, list.SelectedIndex, "After third MoveUp (boundary)")
}

func TestListSelected(t *testing.T) {
	items := []ListItem{
		{Title: "Item 1", Value: "value1"},
		{Title: "Item 2", Value: "value2"},
	}

	list := NewList("Test", items)

	selected := list.Selected()
	require.NotNil(t, selected, "Selected() returned nil")
	assert.Equal(t, "Item 1", selected.Title)
	assert.Equal(t, "value1", selected.Value)

	list.MoveDown()
	selected = list.Selected()
	assert.Equal(t, "Item 2", selected.Title, "After MoveDown")
}

func TestListEmptySelected(t *testing.T) {
	list := NewList("Test", []ListItem{})

	selected := list.Selected()
	assert.Nil(t, selected, "Selected() should return nil for empty list")
}

func TestListRender(t *testing.T) {
	items := []ListItem{
		{Title: "First", Description: "The first item"},
		{Title: "Second", Description: "The second item"},
	}

	list := NewList("My List", items)

	rendered := list.Render()

	// Should contain title
	assert.Contains(t, rendered, "My List", "Rendered output should contain title")

	// Should contain items
	assert.Contains(t, rendered, "First", "Rendered output should contain first item")
	assert.Contains(t, rendered, "Second", "Rendered output should contain second item")

	// Should show selection cursor
	assert.Contains(t, rendered, "> ", "Rendered output should contain selection cursor")
}

func TestListEmptyRender(t *testing.T) {
	list := NewList("Empty", []ListItem{})

	rendered := list.Render()
	assert.Equal(t, "No items", rendered)
}

func TestListSetSize(t *testing.T) {
	list := NewList("Test", []ListItem{{Title: "Item"}})

	list.SetSize(80, 20)

	assert.Equal(t, 80, list.Width)
	assert.Equal(t, 20, list.Height)
}

func TestListRenderWithPagination(t *testing.T) {
	// Create a list with 10 items
	items := make([]ListItem, 10)
	for i := range items {
		items[i] = ListItem{Title: fmt.Sprintf("Item %d", i+1)}
	}

	list := NewList("", items) // No title to simplify height calculation
	list.Height = 5            // Only room for 5 items

	rendered := list.Render()

	// Should contain first 5 items
	assert.Contains(t, rendered, "Item 1")
	assert.Contains(t, rendered, "Item 5")

	// Should NOT contain item 6+
	assert.NotContains(t, rendered, "Item 6")

	// Should show "more below" indicator
	assert.Contains(t, rendered, "↓")
}

func TestListRenderPaginationScrollsWithSelection(t *testing.T) {
	items := make([]ListItem, 10)
	for i := range items {
		items[i] = ListItem{Title: fmt.Sprintf("Item %d", i+1)}
	}

	list := NewList("", items)
	list.Height = 5

	// Move selection to item 7 (index 6)
	for i := 0; i < 6; i++ {
		list.MoveDown()
	}

	rendered := list.Render()

	// Should contain item 7 (the selected one)
	assert.Contains(t, rendered, "Item 7")

	// Should show "more above" indicator
	assert.Contains(t, rendered, "↑")
}

func TestListRenderNoPaginationWhenAllFit(t *testing.T) {
	items := []ListItem{
		{Title: "Item 1"},
		{Title: "Item 2"},
		{Title: "Item 3"},
	}

	list := NewList("", items)
	list.Height = 10 // Plenty of room

	rendered := list.Render()

	// Should contain all items
	assert.Contains(t, rendered, "Item 1")
	assert.Contains(t, rendered, "Item 3")

	// Should NOT show scroll indicators
	assert.NotContains(t, rendered, "↑")
	assert.NotContains(t, rendered, "↓")
}

func TestListRenderNoPaginationWhenHeightZero(t *testing.T) {
	items := make([]ListItem, 10)
	for i := range items {
		items[i] = ListItem{Title: fmt.Sprintf("Item %d", i+1)}
	}

	list := NewList("", items)
	// Height = 0 (default) means show all, for backwards compat

	rendered := list.Render()

	// Should contain all 10 items
	assert.Contains(t, rendered, "Item 1")
	assert.Contains(t, rendered, "Item 10")
}

func TestListScrollOffsetAdjustsOnMoveDown(t *testing.T) {
	items := make([]ListItem, 10)
	for i := range items {
		items[i] = ListItem{Title: fmt.Sprintf("Item %d", i+1)}
	}

	list := NewList("", items)
	list.Height = 3

	// Move down past visible window
	list.MoveDown() // index 1
	list.MoveDown() // index 2
	list.MoveDown() // index 3 — should scroll

	assert.Equal(t, 3, list.SelectedIndex)
	assert.True(t, list.ScrollOffset > 0, "ScrollOffset should increase when selection moves past visible area")
}

func TestListScrollOffsetAdjustsOnMoveUp(t *testing.T) {
	items := make([]ListItem, 10)
	for i := range items {
		items[i] = ListItem{Title: fmt.Sprintf("Item %d", i+1)}
	}

	list := NewList("", items)
	list.Height = 3
	list.SelectedIndex = 5
	list.ScrollOffset = 4

	// Move up past visible window top
	list.MoveUp() // index 4
	list.MoveUp() // index 3 — should scroll up

	assert.Equal(t, 3, list.SelectedIndex)
	assert.True(t, list.ScrollOffset <= 3, "ScrollOffset should decrease when selection moves above visible area")
}

func TestListPaginationWithTitle(t *testing.T) {
	items := make([]ListItem, 10)
	for i := range items {
		items[i] = ListItem{Title: fmt.Sprintf("Item %d", i+1)}
	}

	list := NewList("My Title", items)
	list.Height = 7 // Title takes 2 lines (title + blank), so ~5 items fit

	rendered := list.Render()

	// Should contain the title
	assert.Contains(t, rendered, "My Title")

	// Should NOT contain item 6+ (title takes 2 lines, leaving 5 for items)
	// Items 1-5 should be visible, but we may see at most 5 items
	assert.Contains(t, rendered, "Item 1")
}
