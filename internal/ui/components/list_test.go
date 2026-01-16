package components

import (
	"strings"
	"testing"
)

func TestNewList(t *testing.T) {
	items := []ListItem{
		{Title: "Item 1", Description: "First item"},
		{Title: "Item 2", Description: "Second item"},
	}

	list := NewList("Test List", items)

	if list.Title != "Test List" {
		t.Errorf("Title = %s, want Test List", list.Title)
	}
	if len(list.Items) != 2 {
		t.Errorf("len(Items) = %d, want 2", len(list.Items))
	}
	if list.SelectedIndex != 0 {
		t.Errorf("SelectedIndex = %d, want 0", list.SelectedIndex)
	}
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
	if list.SelectedIndex != 1 {
		t.Errorf("After MoveDown: SelectedIndex = %d, want 1", list.SelectedIndex)
	}

	list.MoveDown()
	if list.SelectedIndex != 2 {
		t.Errorf("After second MoveDown: SelectedIndex = %d, want 2", list.SelectedIndex)
	}

	// Test boundary - shouldn't go past last item
	list.MoveDown()
	if list.SelectedIndex != 2 {
		t.Errorf("After third MoveDown: SelectedIndex = %d, want 2 (boundary)", list.SelectedIndex)
	}

	// Test moving up
	list.MoveUp()
	if list.SelectedIndex != 1 {
		t.Errorf("After MoveUp: SelectedIndex = %d, want 1", list.SelectedIndex)
	}

	list.MoveUp()
	if list.SelectedIndex != 0 {
		t.Errorf("After second MoveUp: SelectedIndex = %d, want 0", list.SelectedIndex)
	}

	// Test boundary - shouldn't go below 0
	list.MoveUp()
	if list.SelectedIndex != 0 {
		t.Errorf("After third MoveUp: SelectedIndex = %d, want 0 (boundary)", list.SelectedIndex)
	}
}

func TestListSelected(t *testing.T) {
	items := []ListItem{
		{Title: "Item 1", Value: "value1"},
		{Title: "Item 2", Value: "value2"},
	}

	list := NewList("Test", items)

	selected := list.Selected()
	if selected == nil {
		t.Fatal("Selected() returned nil")
	}
	if selected.Title != "Item 1" {
		t.Errorf("Selected().Title = %s, want Item 1", selected.Title)
	}
	if selected.Value != "value1" {
		t.Errorf("Selected().Value = %v, want value1", selected.Value)
	}

	list.MoveDown()
	selected = list.Selected()
	if selected.Title != "Item 2" {
		t.Errorf("After MoveDown: Selected().Title = %s, want Item 2", selected.Title)
	}
}

func TestListEmptySelected(t *testing.T) {
	list := NewList("Test", []ListItem{})

	selected := list.Selected()
	if selected != nil {
		t.Error("Selected() should return nil for empty list")
	}
}

func TestListRender(t *testing.T) {
	items := []ListItem{
		{Title: "First", Description: "The first item"},
		{Title: "Second", Description: "The second item"},
	}

	list := NewList("My List", items)

	rendered := list.Render()

	// Should contain title
	if !strings.Contains(rendered, "My List") {
		t.Error("Rendered output should contain title")
	}

	// Should contain items
	if !strings.Contains(rendered, "First") {
		t.Error("Rendered output should contain first item")
	}
	if !strings.Contains(rendered, "Second") {
		t.Error("Rendered output should contain second item")
	}

	// Should show selection cursor
	if !strings.Contains(rendered, "> ") {
		t.Error("Rendered output should contain selection cursor")
	}
}

func TestListEmptyRender(t *testing.T) {
	list := NewList("Empty", []ListItem{})

	rendered := list.Render()
	if rendered != "No items" {
		t.Errorf("Empty list render = %s, want 'No items'", rendered)
	}
}

func TestListSetSize(t *testing.T) {
	list := NewList("Test", []ListItem{{Title: "Item"}})

	list.SetSize(80, 20)

	if list.Width != 80 {
		t.Errorf("Width = %d, want 80", list.Width)
	}
	if list.Height != 20 {
		t.Errorf("Height = %d, want 20", list.Height)
	}
}
