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

// typeRunes sends each character of s as an individual key press to the model.
func typeRunes(m *InventoryModel, s string) *InventoryModel {
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

func TestCustomItemCreationWeapon(t *testing.T) {
	char := models.NewCharacter("test-1", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage(t.TempDir())
	m := NewInventoryModel(char, store)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 150, Height: 40})

	before := len(char.Inventory.Items)

	// Open custom creator from the Items panel.
	m.focus = FocusItems
	m, _ = m.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	assert.True(t, m.customMode, "pressing c should open the custom creator")

	// Step 0: name.
	m = typeRunes(m, "Sunblade")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Equal(t, 1, m.customStep, "should advance to type step")

	// Step 1: choose Weapon (index 1).
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.True(t, m.customIsWeapon(), "should have weapon selected")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Equal(t, 2, m.customStep, "should advance to weight step")

	// Step 2: weight.
	m = typeRunes(m, "3")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Equal(t, 3, m.customStep, "weapon should advance to damage step")

	// Step 3: damage, then create.
	m = typeRunes(m, "1d8")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.False(t, m.customMode, "creator should close after creation")
	assert.Equal(t, before+1, len(char.Inventory.Items), "item should be added")

	created := char.Inventory.Items[len(char.Inventory.Items)-1]
	assert.Equal(t, "Sunblade", created.Name)
	assert.Equal(t, models.ItemTypeWeapon, created.Type)
	assert.Equal(t, 3.0, created.Weight)
	assert.Equal(t, "1d8", created.Damage)
	assert.True(t, created.Custom, "created item should be flagged homebrew")
}

func TestCustomItemGeneralSkipsDamage(t *testing.T) {
	char := models.NewCharacter("test-1", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage(t.TempDir())
	m := NewInventoryModel(char, store)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 150, Height: 40})

	m.focus = FocusItems
	m, _ = m.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})

	// Name, keep default type (General, index 0).
	m = typeRunes(m, "Mysterious Orb")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // accept General
	assert.Equal(t, 2, m.customStep, "should be on weight step")

	// Blank weight, Enter should create immediately (no damage step).
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.False(t, m.customMode, "general item should be created without damage step")

	created := char.Inventory.Items[len(char.Inventory.Items)-1]
	assert.Equal(t, "Mysterious Orb", created.Name)
	assert.Equal(t, models.ItemTypeGeneral, created.Type)
	assert.Equal(t, 0.0, created.Weight)
	assert.True(t, created.Custom)
}

func TestCustomItemEscapeCancels(t *testing.T) {
	char := models.NewCharacter("test-1", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage(t.TempDir())
	m := NewInventoryModel(char, store)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 150, Height: 40})

	before := len(char.Inventory.Items)
	m.focus = FocusItems
	m, _ = m.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	m = typeRunes(m, "Discarded")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	assert.False(t, m.customMode, "escape should close the creator")
	assert.Equal(t, before, len(char.Inventory.Items), "no item should be added on cancel")
}

func TestCustomItemEmptyNameRejected(t *testing.T) {
	char := models.NewCharacter("test-1", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage(t.TempDir())
	m := NewInventoryModel(char, store)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 150, Height: 40})

	m.focus = FocusItems
	m, _ = m.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	// Press Enter with an empty name; should stay on step 0.
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Equal(t, 0, m.customStep, "empty name should not advance")
	assert.True(t, m.customMode, "creator should still be open")
}
