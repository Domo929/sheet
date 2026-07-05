package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCharacterDefensesFromRacialTraits(t *testing.T) {
	char := NewCharacter("id", "Gr01", "Tiefling", "Wizard")
	char.Features.AddRacialTrait("Infernal Legacy", "Tiefling", "You have Resistance to Fire damage.")

	def := char.Defenses()
	assert.Equal(t, []string{"Fire"}, def.Resistances)
	assert.Empty(t, def.Immunities)
	assert.Empty(t, def.Vulnerabilities)
	assert.False(t, def.Empty())
}

func TestCharacterDefensesFromFeats(t *testing.T) {
	char := NewCharacter("id", "Hero", "Human", "Fighter")
	char.Features.Feats = append(char.Features.Feats,
		NewFeature("Some Feat", "Feat", "You have Resistance to Necrotic damage."))

	def := char.Defenses()
	assert.Equal(t, []string{"Necrotic"}, def.Resistances)
}

func TestCharacterDefensesDragonbornAncestry(t *testing.T) {
	char := NewCharacter("id", "Ssur", "Dragonborn", "Paladin")
	char.Info.Subrace = "Red"
	// The generic ancestry trait text should be ignored; the concrete
	// resistance comes from the chosen ancestry.
	char.Features.AddRacialTrait("Draconic Ancestry", "Dragonborn",
		"You have Resistance to the damage type determined by your Draconic Ancestry.")

	def := char.Defenses()
	assert.Equal(t, []string{"Fire"}, def.Resistances)
}

func TestCharacterDefensesExcludesClassFeatures(t *testing.T) {
	char := NewCharacter("id", "Grog", "Human", "Barbarian")
	// Rage grants conditional resistance; it must NOT appear as a permanent
	// defense.
	char.Features.AddClassFeature("Rage", "Barbarian 1",
		"While raging you have Resistance to Bludgeoning, Piercing, and Slashing damage.", 1, "")

	def := char.Defenses()
	assert.True(t, def.Empty(), "class features must be excluded from passive defenses")
}

func TestCharacterDefensesEmpty(t *testing.T) {
	char := NewCharacter("id", "Plain", "Human", "Fighter")
	assert.True(t, char.Defenses().Empty())
}

func TestWeaponMasteryLimit(t *testing.T) {
	fighter := NewCharacter("id", "F", "Human", "Fighter")
	fighter.Info.Level = 1
	assert.Equal(t, 3, fighter.WeaponMasteryLimit())
	fighter.Info.Level = 10
	assert.Equal(t, 5, fighter.WeaponMasteryLimit())

	wizard := NewCharacter("id", "W", "Human", "Wizard")
	assert.Equal(t, 0, wizard.WeaponMasteryLimit())
}

func TestToggleWeaponMastery(t *testing.T) {
	char := NewCharacter("id", "Rogue", "Human", "Rogue") // limit 2
	char.Info.Level = 5

	assert.False(t, char.HasWeaponMastery("Shortsword"))

	assert.True(t, char.ToggleWeaponMastery("Shortsword"))
	assert.True(t, char.HasWeaponMastery("Shortsword"))
	assert.True(t, char.HasWeaponMastery("shortsword"), "match is case-insensitive")

	assert.True(t, char.ToggleWeaponMastery("Dagger"))
	assert.Len(t, char.MasteredWeapons, 2)

	// At the limit (2), a third weapon is refused.
	assert.False(t, char.ToggleWeaponMastery("Rapier"))
	assert.False(t, char.HasWeaponMastery("Rapier"))
	assert.Len(t, char.MasteredWeapons, 2)

	// Removing frees a slot.
	assert.True(t, char.ToggleWeaponMastery("Dagger"))
	assert.False(t, char.HasWeaponMastery("Dagger"))
	assert.True(t, char.ToggleWeaponMastery("Rapier"))
	assert.True(t, char.HasWeaponMastery("Rapier"))
}

func TestToggleWeaponMasteryNoLimit(t *testing.T) {
	char := NewCharacter("id", "W", "Human", "Wizard") // limit 0
	assert.False(t, char.ToggleWeaponMastery("Dagger"))
	assert.Empty(t, char.MasteredWeapons)
}
