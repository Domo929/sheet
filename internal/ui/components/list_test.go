package components

import (
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
