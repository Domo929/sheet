package models

import "testing"

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
			gp, cp := tt.currency.TotalInGold()
			_ = cp // Ignore remaining copper for these tests
			if gp != tt.expected {
				t.Errorf("TotalInGold() = %d GP, want %d GP", gp, tt.expected)
			}
		})
	}
}

func TestCurrencyAdd(t *testing.T) {
	c := NewCurrency()
	c.Add(10, 5, 2, 100, 1)

	if c.Copper != 10 {
		t.Errorf("Copper = %d, want 10", c.Copper)
	}
	if c.Silver != 5 {
		t.Errorf("Silver = %d, want 5", c.Silver)
	}
	if c.Electrum != 2 {
		t.Errorf("Electrum = %d, want 2", c.Electrum)
	}
	if c.Gold != 100 {
		t.Errorf("Gold = %d, want 100", c.Gold)
	}
	if c.Platinum != 1 {
		t.Errorf("Platinum = %d, want 1", c.Platinum)
	}
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
			if (err != nil) != tt.wantErr {
				t.Errorf("SpendFromTotal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.currency.Gold != tt.wantGP {
				t.Errorf("Gold = %d, want %d", tt.currency.Gold, tt.wantGP)
			}
			if tt.currency.Platinum != tt.wantPP {
				t.Errorf("Platinum = %d, want %d", tt.currency.Platinum, tt.wantPP)
			}
		})
	}
}

func TestItemCharges(t *testing.T) {
	item := NewItem("wand-1", "Wand of Magic Missiles", ItemTypeMagicItem)
	item.Charges = 7
	item.MaxCharges = 7

	// Use some charges
	for i := 0; i < 5; i++ {
		if !item.UseCharge() {
			t.Errorf("UseCharge() should succeed at %d charges", item.Charges+1)
		}
	}

	if item.Charges != 2 {
		t.Errorf("Charges = %d, want 2", item.Charges)
	}

	// Recharge
	item.Recharge(10) // More than max
	if item.Charges != 7 {
		t.Errorf("Charges after recharge = %d, want 7", item.Charges)
	}

	// Use all charges
	for item.Charges > 0 {
		item.UseCharge()
	}

	if item.UseCharge() {
		t.Error("UseCharge() should fail at 0 charges")
	}
}

func TestEquipmentSlots(t *testing.T) {
	equip := NewEquipment()

	sword := NewItem("sword-1", "Longsword", ItemTypeWeapon)
	shield := NewItem("shield-1", "Shield", ItemTypeShield)

	// Equip items
	equip.SetSlot(SlotMainHand, &sword)
	equip.SetSlot(SlotOffHand, &shield)

	// Verify equipped
	if equip.GetSlot(SlotMainHand) == nil || equip.GetSlot(SlotMainHand).Name != "Longsword" {
		t.Error("Main hand should have Longsword")
	}
	if equip.GetSlot(SlotOffHand) == nil || equip.GetSlot(SlotOffHand).Name != "Shield" {
		t.Error("Off hand should have Shield")
	}

	// Unequip (replace with nil)
	previous := equip.SetSlot(SlotMainHand, nil)
	if previous == nil || previous.Name != "Longsword" {
		t.Error("SetSlot should return previously equipped item")
	}
	if equip.GetSlot(SlotMainHand) != nil {
		t.Error("Main hand should be empty after unequip")
	}
}

func TestEquipmentCountAttunedItems(t *testing.T) {
	equip := NewEquipment()

	// Add non-attuned item
	sword := NewItem("sword-1", "Longsword", ItemTypeWeapon)
	equip.SetSlot(SlotMainHand, &sword)

	if equip.CountAttunedItems() != 0 {
		t.Error("Should have 0 attuned items")
	}

	// Add attuned items
	ring1 := NewItem("ring-1", "Ring of Protection", ItemTypeMagicItem)
	ring1.Attuned = true
	equip.EquipRing(&ring1)

	cloak := NewItem("cloak-1", "Cloak of Displacement", ItemTypeMagicItem)
	cloak.Attuned = true
	equip.SetSlot(SlotCloak, &cloak)

	if equip.CountAttunedItems() != 2 {
		t.Errorf("CountAttunedItems() = %d, want 2", equip.CountAttunedItems())
	}
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
	if len(inv.Items) != 2 {
		t.Errorf("Items count = %d, want 2", len(inv.Items))
	}

	// Find item
	found := inv.FindItem("sword-1")
	if found == nil || found.Name != "Longsword" {
		t.Error("FindItem should return the Longsword")
	}

	// Find non-existent
	if inv.FindItem("nonexistent") != nil {
		t.Error("FindItem should return nil for non-existent ID")
	}

	// Total weight: 3.0 + (0.5 * 3) = 4.5
	if inv.TotalWeight() != 4.5 {
		t.Errorf("TotalWeight() = %f, want 4.5", inv.TotalWeight())
	}

	// Remove item
	removed := inv.RemoveItem("sword-1")
	if removed == nil || removed.Name != "Longsword" {
		t.Error("RemoveItem should return the removed item")
	}
	if len(inv.Items) != 1 {
		t.Errorf("Items count after remove = %d, want 1", len(inv.Items))
	}

	// Remove non-existent
	if inv.RemoveItem("nonexistent") != nil {
		t.Error("RemoveItem should return nil for non-existent ID")
	}
}
