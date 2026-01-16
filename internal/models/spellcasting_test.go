package models

import "testing"

func TestSlotTrackerUse(t *testing.T) {
	st := NewSlotTracker(3)

	for i := 0; i < 3; i++ {
		if !st.Use() {
			t.Errorf("Use() should succeed with %d remaining", st.Remaining+1)
		}
	}

	if st.Use() {
		t.Error("Use() should fail with 0 remaining")
	}

	st.Restore()
	if st.Remaining != 3 {
		t.Errorf("Remaining after Restore = %d, want 3", st.Remaining)
	}
}

func TestSpellSlotsGetAndSet(t *testing.T) {
	ss := NewSpellSlots()

	// Set slots for various levels
	ss.SetSlots(1, 4)
	ss.SetSlots(2, 3)
	ss.SetSlots(3, 3)
	ss.SetSlots(5, 1)

	tests := []struct {
		level    int
		expected int
	}{
		{1, 4},
		{2, 3},
		{3, 3},
		{4, 0},
		{5, 1},
		{6, 0},
	}

	for _, tt := range tests {
		slot := ss.GetSlot(tt.level)
		if slot == nil {
			t.Errorf("GetSlot(%d) returned nil", tt.level)
			continue
		}
		if slot.Total != tt.expected {
			t.Errorf("Level %d Total = %d, want %d", tt.level, slot.Total, tt.expected)
		}
	}
}

func TestSpellSlotsGetInvalid(t *testing.T) {
	ss := NewSpellSlots()

	if ss.GetSlot(0) != nil {
		t.Error("GetSlot(0) should return nil")
	}
	if ss.GetSlot(10) != nil {
		t.Error("GetSlot(10) should return nil")
	}
}

func TestSpellSlotsUseSlot(t *testing.T) {
	ss := NewSpellSlots()
	ss.SetSlots(1, 2)

	// Use both slots
	if !ss.UseSlot(1) {
		t.Error("UseSlot(1) should succeed")
	}
	if !ss.UseSlot(1) {
		t.Error("UseSlot(1) should succeed second time")
	}
	if ss.UseSlot(1) {
		t.Error("UseSlot(1) should fail when empty")
	}

	// Invalid level
	if ss.UseSlot(0) {
		t.Error("UseSlot(0) should fail")
	}
}

func TestSpellSlotsRestoreAll(t *testing.T) {
	ss := NewSpellSlots()
	ss.SetSlots(1, 4)
	ss.SetSlots(2, 3)
	ss.SetSlots(3, 2)

	// Use some slots
	ss.UseSlot(1)
	ss.UseSlot(1)
	ss.UseSlot(2)

	// Restore all
	ss.RestoreAll()

	if ss.Level1.Remaining != 4 {
		t.Errorf("Level1 remaining = %d, want 4", ss.Level1.Remaining)
	}
	if ss.Level2.Remaining != 3 {
		t.Errorf("Level2 remaining = %d, want 3", ss.Level2.Remaining)
	}
}

func TestPactMagic(t *testing.T) {
	pm := NewPactMagic(5, 2) // 2 5th-level slots

	if pm.SlotLevel != 5 {
		t.Errorf("SlotLevel = %d, want 5", pm.SlotLevel)
	}

	// Use slots
	if !pm.Use() {
		t.Error("Use() should succeed")
	}
	if !pm.Use() {
		t.Error("Use() should succeed second time")
	}
	if pm.Use() {
		t.Error("Use() should fail when empty")
	}

	// Restore
	pm.Restore()
	if pm.Remaining != 2 {
		t.Errorf("Remaining after Restore = %d, want 2", pm.Remaining)
	}
}

func TestSpellcastingSpells(t *testing.T) {
	sc := NewSpellcasting(AbilityWisdom)

	// Add cantrips
	sc.AddCantrip("Light")
	sc.AddCantrip("Sacred Flame")

	if len(sc.CantripsKnown) != 2 {
		t.Errorf("CantripsKnown count = %d, want 2", len(sc.CantripsKnown))
	}

	// Add spells
	sc.AddSpell("Cure Wounds", 1)
	sc.AddSpell("Bless", 1)
	sc.AddSpell("Spiritual Weapon", 2)

	if len(sc.KnownSpells) != 3 {
		t.Errorf("KnownSpells count = %d, want 3", len(sc.KnownSpells))
	}
}

func TestSpellcastingPrepare(t *testing.T) {
	sc := NewSpellcasting(AbilityWisdom)
	sc.PreparesSpells = true
	sc.MaxPrepared = 4

	sc.AddSpell("Cure Wounds", 1)
	sc.AddSpell("Bless", 1)
	sc.AddSpell("Healing Word", 1)

	// Prepare some spells
	sc.PrepareSpell("Cure Wounds", true)
	sc.PrepareSpell("Bless", true)

	if sc.CountPreparedSpells() != 2 {
		t.Errorf("CountPreparedSpells() = %d, want 2", sc.CountPreparedSpells())
	}

	prepared := sc.GetPreparedSpells()
	if len(prepared) != 2 {
		t.Errorf("GetPreparedSpells() count = %d, want 2", len(prepared))
	}

	// Unprepare
	sc.PrepareSpell("Bless", false)
	if sc.CountPreparedSpells() != 1 {
		t.Errorf("CountPreparedSpells() after unprepare = %d, want 1", sc.CountPreparedSpells())
	}

	// Prepare non-existent spell
	if sc.PrepareSpell("Nonexistent", true) {
		t.Error("PrepareSpell should return false for unknown spell")
	}
}

func TestCalculateSpellSaveDC(t *testing.T) {
	tests := []struct {
		name      string
		abilityMod int
		profBonus  int
		expected   int
	}{
		{"level 1, +3 ability", 3, 2, 13},
		{"level 5, +4 ability", 4, 3, 15},
		{"level 17, +5 ability", 5, 6, 19},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := CalculateSpellSaveDC(tt.abilityMod, tt.profBonus)
			if dc != tt.expected {
				t.Errorf("CalculateSpellSaveDC(%d, %d) = %d, want %d",
					tt.abilityMod, tt.profBonus, dc, tt.expected)
			}
		})
	}
}

func TestCalculateSpellAttackBonus(t *testing.T) {
	tests := []struct {
		name      string
		abilityMod int
		profBonus  int
		expected   int
	}{
		{"level 1, +3 ability", 3, 2, 5},
		{"level 5, +4 ability", 4, 3, 7},
		{"level 17, +5 ability", 5, 6, 11},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bonus := CalculateSpellAttackBonus(tt.abilityMod, tt.profBonus)
			if bonus != tt.expected {
				t.Errorf("CalculateSpellAttackBonus(%d, %d) = %d, want %d",
					tt.abilityMod, tt.profBonus, bonus, tt.expected)
			}
		})
	}
}
