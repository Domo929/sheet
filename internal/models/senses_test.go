package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func traits(pairs ...string) []Feature {
	var out []Feature
	for i := 0; i+1 < len(pairs); i += 2 {
		out = append(out, NewFeature(pairs[i], "Test", pairs[i+1]))
	}
	return out
}

func TestDeriveSensesDarkvision(t *testing.T) {
	s := DeriveSensesFromTraits(traits(
		"Darkvision", "You have Darkvision with a range of 60 feet.",
	))
	assert.Equal(t, 60, s.Darkvision)
	assert.Zero(t, s.Blindsight)

	// Drow lineage upgrades darkvision; the greater range wins.
	s = DeriveSensesFromTraits(traits(
		"Darkvision", "You have Darkvision with a range of 60 feet.",
		"Drow Lineage", "The range of your Darkvision increases to 120 feet. You also know the Dancing Lights cantrip.",
	))
	assert.Equal(t, 120, s.Darkvision)
}

func TestDeriveSensesIgnoresTemporaryGrants(t *testing.T) {
	// Dwarf Stonecunning grants Tremorsense only as a Bonus Action.
	s := DeriveSensesFromTraits(traits(
		"Stonecunning", "As a Bonus Action, you gain Tremorsense with a range of 60 feet for 10 minutes.",
	))
	assert.Zero(t, s.Tremorsense, "temporary/conditional senses must not be recorded")
}

func TestDerivePermanentBlindsightAndTruesight(t *testing.T) {
	s := DeriveSensesFromTraits(traits(
		"Echolocation", "You have Blindsight with a range of 30 feet.",
		"Cosmic Sight", "You have Truesight with a range of 10 feet.",
	))
	assert.Equal(t, 30, s.Blindsight)
	assert.Equal(t, 10, s.Truesight)
}

func TestDeriveWalkSpeedOverride(t *testing.T) {
	assert.Equal(t, 35, DeriveWalkSpeedOverride(traits(
		"Wood Elf Lineage", "Your Speed increases to 35 feet. You also know the Druidcraft cantrip.",
	)))

	// Reducing a target's speed must not be read as a self override.
	assert.Zero(t, DeriveWalkSpeedOverride(traits(
		"Frost's Chill", "You can reduce the target's Speed by 10 feet until the start of your next turn.",
	)))

	// A Fly Speed mention should not be read as a walking-speed override.
	assert.Zero(t, DeriveWalkSpeedOverride(traits(
		"Radiant Soul", "You have a Fly Speed equal to your Speed.",
	)))
}

func TestSyncSensesPreservesHigherExisting(t *testing.T) {
	c := NewCharacter("id", "Node", "Elf", "Wizard")
	c.Features.AddRacialTrait("Darkvision", "Elf", "You have Darkvision with a range of 60 feet.")

	// A feat already granted superior darkvision; sync must not lower it.
	c.CombatStats.Senses.Darkvision = 90
	c.SyncSenses()
	assert.Equal(t, 90, c.CombatStats.Senses.Darkvision)

	// When unset, sync fills from traits.
	c.CombatStats.Senses.Darkvision = 0
	c.SyncSenses()
	assert.Equal(t, 60, c.CombatStats.Senses.Darkvision)
}

func TestSensesListAndIsEmpty(t *testing.T) {
	assert.True(t, Senses{}.IsEmpty())
	assert.Empty(t, Senses{}.List())

	s := Senses{Darkvision: 120, Tremorsense: 30}
	assert.False(t, s.IsEmpty())
	assert.Equal(t, []string{"Darkvision 120 ft", "Tremorsense 30 ft"}, s.List())
}
