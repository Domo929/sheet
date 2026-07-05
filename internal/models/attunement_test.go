package models

import "testing"

func TestInventoryToggleAttunement(t *testing.T) {
	inv := NewInventory()
	ring := NewItem("r1", "Ring of Protection", ItemTypeMagicItem)
	ring.RequiresAttunement = true
	inv.AddItem(ring)

	if err := inv.ToggleAttunement("r1"); err != nil {
		t.Fatalf("attune failed: %v", err)
	}
	if inv.CountAttunedItems() != 1 {
		t.Fatalf("expected 1 attuned, got %d", inv.CountAttunedItems())
	}
	if err := inv.ToggleAttunement("r1"); err != nil {
		t.Fatalf("unattune failed: %v", err)
	}
	if inv.CountAttunedItems() != 0 {
		t.Errorf("expected 0 attuned after toggle off")
	}
}

func TestInventoryAttunementRequiresFlag(t *testing.T) {
	inv := NewInventory()
	sword := NewItem("s1", "Longsword", ItemTypeWeapon)
	inv.AddItem(sword)
	if err := inv.ToggleAttunement("s1"); err != ErrItemDoesNotRequireAttunement {
		t.Errorf("expected ErrItemDoesNotRequireAttunement, got %v", err)
	}
}

func TestInventoryAttunementMaxThree(t *testing.T) {
	inv := NewInventory()
	for _, id := range []string{"a", "b", "c", "d"} {
		it := NewItem(id, "Item"+id, ItemTypeMagicItem)
		it.RequiresAttunement = true
		inv.AddItem(it)
	}
	for _, id := range []string{"a", "b", "c"} {
		if err := inv.ToggleAttunement(id); err != nil {
			t.Fatalf("attune %s failed: %v", id, err)
		}
	}
	if err := inv.ToggleAttunement("d"); err != ErrMaxAttunementReached {
		t.Errorf("expected ErrMaxAttunementReached, got %v", err)
	}
}
