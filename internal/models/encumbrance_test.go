package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func encTestChar(str int) *Character {
	c := NewCharacter("enc-1", "Test", "Human", "Fighter")
	c.AbilityScores.Strength = AbilityScore{Base: str}
	return c
}

func addWeight(c *Character, name string, weight float64, qty int) {
	item := NewItem(name, name, ItemTypeGeneral)
	item.Weight = weight
	item.Quantity = qty
	c.Inventory.AddItem(item)
}

func TestCarryingCapacity(t *testing.T) {
	assert.Equal(t, 225.0, encTestChar(15).CarryingCapacity())
	assert.Equal(t, 450.0, encTestChar(15).PushDragLiftCapacity())
	// Temporary Strength bonus counts toward capacity.
	c := encTestChar(15)
	c.AbilityScores.Strength.Temporary = 4 // e.g. a Giant's Might effect
	assert.Equal(t, 285.0, c.CarryingCapacity())
}

func TestCarriedWeight(t *testing.T) {
	c := encTestChar(10)
	addWeight(c, "Greatclub", 10, 1)
	addWeight(c, "Rations", 2, 5)
	assert.Equal(t, 20.0, c.CarriedWeight())
}

func TestEncumbranceTiers(t *testing.T) {
	// STR 10 -> thresholds 50 / 100 / 150.
	tests := []struct {
		weight float64
		want   EncumbranceLevel
	}{
		{0, LoadNormal},
		{50, LoadNormal},       // exactly 5x is not yet encumbered
		{50.1, LoadEncumbered}, // just over 5x
		{100, LoadEncumbered},  // exactly 10x still only encumbered
		{100.1, LoadHeavilyEncumbered},
		{150, LoadHeavilyEncumbered}, // exactly 15x is at capacity, not over
		{150.1, LoadOverCapacity},
	}
	for _, tt := range tests {
		c := encTestChar(10)
		addWeight(c, "Load", tt.weight, 1)
		assert.Equalf(t, tt.want, c.Encumbrance(),
			"weight %.1f with STR 10", tt.weight)
	}
}

func TestEncumbranceLevelStringAndPenalty(t *testing.T) {
	assert.Equal(t, "Unencumbered", LoadNormal.String())
	assert.Equal(t, "Encumbered", LoadEncumbered.String())
	assert.Equal(t, "Heavily Encumbered", LoadHeavilyEncumbered.String())
	assert.Equal(t, "Over Capacity", LoadOverCapacity.String())

	assert.Equal(t, 0, LoadNormal.SpeedPenalty())
	assert.Equal(t, 10, LoadEncumbered.SpeedPenalty())
	assert.Equal(t, 20, LoadHeavilyEncumbered.SpeedPenalty())
	assert.Equal(t, 20, LoadOverCapacity.SpeedPenalty())
}
