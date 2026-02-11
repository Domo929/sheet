package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrencyTotalInGold(t *testing.T) {
	tests := []struct {
		name     string
		currency Currency
		expected int
	}{
		{"empty", Currency{}, 0},
		{"10 gold", Currency{Gold: 10}, 10},
		{"100 copper = 1 gold", Currency{Copper: 100}, 1},
		{"10 silver = 1 gold", Currency{Silver: 10}, 1},
		{"2 electrum = 1 gold", Currency{Electrum: 2}, 1},
		{"1 platinum = 10 gold", Currency{Platinum: 1}, 10},
		{"mixed", Currency{Copper: 50, Silver: 5, Gold: 2, Platinum: 1}, 13}, // 0.5 + 0.5 + 2 + 10
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gp, _ := tt.currency.TotalInGold()
			assert.Equal(t, tt.expected, gp)
		})
	}
}

func TestCurrencyAdd(t *testing.T) {
	c := NewCurrency()
	c.Add(10, 5, 2, 100, 1)

	assert.Equal(t, 10, c.Copper)
	assert.Equal(t, 5, c.Silver)
	assert.Equal(t, 2, c.Electrum)
	assert.Equal(t, 100, c.Gold)
	assert.Equal(t, 1, c.Platinum)
}

func TestSpendFromTotal(t *testing.T) {
	tests := []struct {
		name      string
		currency  Currency
		spend     int
		wantErr   bool
		wantGP    int
		wantPP    int
	}{
		{
			name:     "simple gold spend",
			currency: Currency{Gold: 50},
			spend:    20,
			wantGP:   0, // 50 GP - 20 GP = 30 GP = 3 PP + 0 GP
			wantPP:   3,
		},
		{
			name:     "spend from platinum",
			currency: Currency{Platinum: 3, Gold: 7},
			spend:    20,
			wantGP:   7, // 37 total GP - 20 = 17 GP = 1 PP + 7 GP
			wantPP:   1,
		},
		{
			name:     "exact spend",
			currency: Currency{Gold: 20},
			spend:    20,
			wantGP:   0,
			wantPP:   0,
		},
		{
			name:     "insufficient funds",
			currency: Currency{Gold: 10},
			spend:    20,
			wantErr:  true,
		},
		{
			name:     "spend from mixed currency",
			currency: Currency{Copper: 50, Silver: 50, Gold: 5, Platinum: 1},
			spend:    10, // total is 2050 CP = 20.5 GP, spend 10 leaves 1050 CP = 1 PP + 0 GP + 5 SP
			wantGP:   0,
			wantPP:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.currency.SpendFromTotal(tt.spend)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantGP, tt.currency.Gold)
			assert.Equal(t, tt.wantPP, tt.currency.Platinum)
		})
	}
}

func TestItemCharges(t *testing.T) {
	item := NewItem("wand-1", "Wand of Magic Missiles", ItemTypeMagicItem)
	item.Charges = 7
	item.MaxCharges = 7

	// Use some charges
	for i := 0; i < 5; i++ {
		assert.True(t, item.UseCharge(), "UseCharge() should succeed at %d charges", item.Charges+1)
	}

	assert.Equal(t, 2, item.Charges)

	// Recharge
	item.Recharge(10) // More than max
	assert.Equal(t, 7, item.Charges)

	// Use all charges
	for item.Charges > 0 {
		item.UseCharge()
	}

	assert.False(t, item.UseCharge(), "UseCharge() should fail at 0 charges")
}

func TestEquipmentSlots(t *testing.T) {
	equip := NewEquipment()

	sword := NewItem("sword-1", "Longsword", ItemTypeWeapon)
	shield := NewItem("shield-1", "Shield", ItemTypeShield)

	// Equip items
	equip.SetSlot(SlotMainHand, &sword)
	equip.SetSlot(SlotOffHand, &shield)

	// Verify equipped
	mainHand := equip.GetSlot(SlotMainHand)
	require.NotNil(t, mainHand, "Main hand should have Longsword")
	assert.Equal(t, "Longsword", mainHand.Name)

	offHand := equip.GetSlot(SlotOffHand)
	require.NotNil(t, offHand, "Off hand should have Shield")
	assert.Equal(t, "Shield", offHand.Name)

	// Unequip (replace with nil)
	previous := equip.SetSlot(SlotMainHand, nil)
	require.NotNil(t, previous, "SetSlot should return previously equipped item")
	assert.Equal(t, "Longsword", previous.Name)
	assert.Nil(t, equip.GetSlot(SlotMainHand), "Main hand should be empty after unequip")
}

func TestEquipmentCountAttunedItems(t *testing.T) {
	equip := NewEquipment()

	// Add non-attuned item
	sword := NewItem("sword-1", "Longsword", ItemTypeWeapon)
	equip.SetSlot(SlotMainHand, &sword)

	assert.Equal(t, 0, equip.CountAttunedItems())

	// Add attuned items
	ring1 := NewItem("ring-1", "Ring of Protection", ItemTypeMagicItem)
	ring1.Attuned = true
	equip.EquipRing(&ring1)

	cloak := NewItem("cloak-1", "Cloak of Displacement", ItemTypeMagicItem)
	cloak.Attuned = true
	equip.SetSlot(SlotCloak, &cloak)

	assert.Equal(t, 2, equip.CountAttunedItems())
}

func TestInventoryOperations(t *testing.T) {
	inv := NewInventory()

	// Add items
	sword := NewItem("sword-1", "Longsword", ItemTypeWeapon)
	sword.Weight = 3.0
	inv.AddItem(sword)

	potion := NewItem("potion-1", "Healing Potion", ItemTypeConsumable)
	potion.Quantity = 3
	potion.Weight = 0.5
	inv.AddItem(potion)

	// Verify items added
	assert.Len(t, inv.Items, 2)

	// Find item
	found := inv.FindItem("sword-1")
	require.NotNil(t, found, "FindItem should return the Longsword")
	assert.Equal(t, "Longsword", found.Name)

	// Find non-existent
	assert.Nil(t, inv.FindItem("nonexistent"), "FindItem should return nil for non-existent ID")

	// Total weight: 3.0 + (0.5 * 3) = 4.5
	assert.Equal(t, 4.5, inv.TotalWeight())

	// Remove item
	removed := inv.RemoveItem("sword-1")
	require.NotNil(t, removed, "RemoveItem should return the removed item")
	assert.Equal(t, "Longsword", removed.Name)
	assert.Len(t, inv.Items, 1)

	// Remove non-existent
	assert.Nil(t, inv.RemoveItem("nonexistent"), "RemoveItem should return nil for non-existent ID")
}
