package views

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestInventoryAttuneToggle(t *testing.T) {
	char := models.NewCharacter("t1", "Test", "Human", "Wizard")
	ring := models.NewItem("r1", "Ring of Protection", models.ItemTypeMagicItem)
	ring.RequiresAttunement = true
	char.Inventory.AddItem(ring)

	store, _ := storage.NewCharacterStorage(t.TempDir())
	m := NewInventoryModel(char, store)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 150, Height: 40})
	m.focus = FocusItems
	m.itemCursor = 0

	m, _ = m.Update(tea.KeyPressMsg{Code: 't', Text: "t"})
	assert.Equal(t, 1, char.Inventory.CountAttunedItems())
	assert.True(t, char.Inventory.Items[0].Attuned)

	m, _ = m.Update(tea.KeyPressMsg{Code: 't', Text: "t"})
	assert.Equal(t, 0, char.Inventory.CountAttunedItems())
}

func TestInventoryAttuneNonAttunementItem(t *testing.T) {
	char := models.NewCharacter("t1", "Test", "Human", "Fighter")
	sword := models.NewItem("s1", "Longsword", models.ItemTypeWeapon)
	char.Inventory.AddItem(sword)

	store, _ := storage.NewCharacterStorage(t.TempDir())
	m := NewInventoryModel(char, store)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 150, Height: 40})
	m.focus = FocusItems
	m.itemCursor = 0

	m, _ = m.Update(tea.KeyPressMsg{Code: 't', Text: "t"})
	assert.Equal(t, 0, char.Inventory.CountAttunedItems())
	assert.Contains(t, m.statusMessage, "does not require attunement")
}
