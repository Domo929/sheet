package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlotTrackerUse(t *testing.T) {
	st := NewSlotTracker(3)

	for i := 0; i < 3; i++ {
		assert.True(t, st.Use(), "Use() should succeed with %d remaining", st.Remaining+1)
	}

	assert.False(t, st.Use(), "Use() should fail with 0 remaining")

	st.Restore()
	assert.Equal(t, 3, st.Remaining)
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
		require.NotNil(t, slot, "GetSlot(%d) returned nil", tt.level)
		assert.Equal(t, tt.expected, slot.Total, "Level %d Total", tt.level)
	}
}

func TestSpellSlotsGetInvalid(t *testing.T) {
	ss := NewSpellSlots()

	assert.Nil(t, ss.GetSlot(0), "GetSlot(0) should return nil")
	assert.Nil(t, ss.GetSlot(10), "GetSlot(10) should return nil")
}

func TestSpellSlotsUseSlot(t *testing.T) {
	ss := NewSpellSlots()
	ss.SetSlots(1, 2)

	// Use both slots
	assert.True(t, ss.UseSlot(1), "UseSlot(1) should succeed")
	assert.True(t, ss.UseSlot(1), "UseSlot(1) should succeed second time")
	assert.False(t, ss.UseSlot(1), "UseSlot(1) should fail when empty")

	// Invalid level
	assert.False(t, ss.UseSlot(0), "UseSlot(0) should fail")
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

	assert.Equal(t, 4, ss.Level1.Remaining)
	assert.Equal(t, 3, ss.Level2.Remaining)
}

func TestPactMagic(t *testing.T) {
	pm := NewPactMagic(5, 2) // 2 5th-level slots

	assert.Equal(t, 5, pm.SlotLevel)

	// Use slots
	assert.True(t, pm.Use(), "Use() should succeed")
	assert.True(t, pm.Use(), "Use() should succeed second time")
	assert.False(t, pm.Use(), "Use() should fail when empty")

	// Restore
	pm.Restore()
	assert.Equal(t, 2, pm.Remaining)
}

func TestSpellcastingSpells(t *testing.T) {
	sc := NewSpellcasting(AbilityWisdom)

	// Add cantrips
	sc.AddCantrip("Light")
	sc.AddCantrip("Sacred Flame")

	assert.Len(t, sc.CantripsKnown, 2)

	// Add spells
	sc.AddSpell("Cure Wounds", 1)
	sc.AddSpell("Bless", 1)
	sc.AddSpell("Spiritual Weapon", 2)

	assert.Len(t, sc.KnownSpells, 3)
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

	assert.Equal(t, 2, sc.CountPreparedSpells())

	prepared := sc.GetPreparedSpells()
	assert.Len(t, prepared, 2)

	// Unprepare
	sc.PrepareSpell("Bless", false)
	assert.Equal(t, 1, sc.CountPreparedSpells())

	// Prepare non-existent spell
	assert.False(t, sc.PrepareSpell("Nonexistent", true), "PrepareSpell should return false for unknown spell")
}

func TestCalculateSpellSaveDC(t *testing.T) {
	tests := []struct {
		name       string
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
			assert.Equal(t, tt.expected, dc)
		})
	}
}

func TestCalculateSpellAttackBonus(t *testing.T) {
	tests := []struct {
		name       string
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
			assert.Equal(t, tt.expected, bonus)
		})
	}
}
