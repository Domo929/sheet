package views

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestInventoryBuyItem(t *testing.T) {
	char := models.NewCharacter("t1", "Test", "Human", "Fighter")
	char.Inventory.Currency = models.Currency{Gold: 50}
	store, _ := storage.NewCharacterStorage(t.TempDir())
	m := NewInventoryModel(char, store)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 150, Height: 40})
	m.focus = FocusItems

	m, _ = m.Update(tea.KeyPressMsg{Code: 'b', Text: "b"})
	assert.True(t, m.addingItem)
	assert.True(t, m.buyMode)

	m.searchResults = []models.Item{{ID: "longsword", Name: "Longsword", Type: models.ItemTypeWeapon, Value: models.Currency{Gold: 15}}}
	m.searchCursor = 0

	before := len(char.Inventory.Items)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Equal(t, before+1, len(char.Inventory.Items))
	assert.Equal(t, 35, char.Inventory.Currency.Gold)
	assert.False(t, m.buyMode)
}

func TestInventoryBuyItemInsufficient(t *testing.T) {
	char := models.NewCharacter("t1", "Test", "Human", "Fighter")
	char.Inventory.Currency = models.Currency{Gold: 5}
	store, _ := storage.NewCharacterStorage(t.TempDir())
	m := NewInventoryModel(char, store)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 150, Height: 40})
	m.focus = FocusItems

	m, _ = m.Update(tea.KeyPressMsg{Code: 'b', Text: "b"})
	m.searchResults = []models.Item{{ID: "plate", Name: "Plate Armor", Value: models.Currency{Gold: 1500}}}
	m.searchCursor = 0

	before := len(char.Inventory.Items)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Equal(t, before, len(char.Inventory.Items), "unaffordable item not added")
	assert.Equal(t, 5, char.Inventory.Currency.Gold)
	assert.True(t, m.addingItem, "search stays open on failed buy")
}

func TestInventorySellItem(t *testing.T) {
	char := models.NewCharacter("t1", "Test", "Human", "Fighter")
	sword := models.NewItem("s1", "Longsword", models.ItemTypeWeapon)
	sword.Value = models.Currency{Gold: 15}
	char.Inventory.AddItem(sword)
	store, _ := storage.NewCharacterStorage(t.TempDir())
	m := NewInventoryModel(char, store)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 150, Height: 40})
	m.focus = FocusItems
	m.itemCursor = 0

	m, _ = m.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	// Half of 15 gp = 750 cp.
	assert.Equal(t, 750, char.Inventory.Currency.TotalCopper())
	assert.Equal(t, 0, len(char.Inventory.Items), "single item removed after sale")
}
