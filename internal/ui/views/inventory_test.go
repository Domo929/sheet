package views

import (
	"testing"

	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

func TestInventoryNarrowWidth(t *testing.T) {
	char := models.NewCharacter("test-1", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage(t.TempDir())

	m := NewInventoryModel(char, store)

	// Simulate a narrow terminal
	sizeMsg := tea.WindowSizeMsg{Width: 65, Height: 30}
	m, _ = m.Update(sizeMsg)

	view := m.View()

	// Should render without panic
	assert.True(t, len(view) > 0, "Narrow inventory should render content")
	// Should contain inventory content
	assert.Contains(t, view, "Inventory", "Should contain inventory header")
}

func TestInventoryWideWidth(t *testing.T) {
	char := models.NewCharacter("test-1", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage(t.TempDir())

	m := NewInventoryModel(char, store)

	// Simulate a wide terminal
	sizeMsg := tea.WindowSizeMsg{Width: 150, Height: 40}
	m, _ = m.Update(sizeMsg)

	view := m.View()

	assert.True(t, len(view) > 0, "Wide inventory should render content")
	assert.Contains(t, view, "Inventory", "Should contain inventory header")
}
