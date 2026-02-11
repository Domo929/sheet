package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSavingThrows(t *testing.T) {
	saves := NewSavingThrows()

	// All should be not proficient
	abilities := []Ability{
		AbilityStrength, AbilityDexterity, AbilityConstitution,
		AbilityIntelligence, AbilityWisdom, AbilityCharisma,
	}

	for _, ability := range abilities {
		assert.False(t, saves.IsProficient(ability), "NewSavingThrows().IsProficient(%s) should be false", ability)
	}
}

func TestSavingThrowsGet(t *testing.T) {
	saves := NewSavingThrows()
	saves.Wisdom.Proficient = true

	abilities := []Ability{
		AbilityStrength, AbilityDexterity, AbilityConstitution,
		AbilityIntelligence, AbilityWisdom, AbilityCharisma,
	}

	for _, ability := range abilities {
		save := saves.Get(ability)
		// Just verify we can get all saves
		_ = save
	}

	// Check specific one we modified
	assert.True(t, saves.Get(AbilityWisdom).Proficient, "Wisdom save should be proficient")
}

func TestSavingThrowsGetInvalid(t *testing.T) {
	saves := NewSavingThrows()
	got := saves.Get(Ability("invalid"))
	// Should return default (not proficient)
	assert.False(t, got.Proficient, "SavingThrows.Get(invalid).Proficient should be false")
}

func TestSavingThrowsSetProficiency(t *testing.T) {
	saves := NewSavingThrows()

	saves.SetProficiency(AbilityConstitution, true)
	assert.True(t, saves.Constitution.Proficient, "Constitution save should be proficient after SetProficiency")

	saves.SetProficiency(AbilityConstitution, false)
	assert.False(t, saves.Constitution.Proficient, "Constitution save should not be proficient after SetProficiency(false)")
}

func TestSavingThrowsIsProficient(t *testing.T) {
	saves := NewSavingThrows()
	saves.Dexterity.Proficient = true

	assert.True(t, saves.IsProficient(AbilityDexterity), "IsProficient(Dexterity) should be true")
	assert.False(t, saves.IsProficient(AbilityStrength), "IsProficient(Strength) should be false")

	// Invalid ability should return false
	assert.False(t, saves.IsProficient(Ability("invalid")), "IsProficient(invalid) should be false")
}

func TestCalculateSavingThrowModifier(t *testing.T) {
	tests := []struct {
		name             string
		proficient       bool
		abilityMod       int
		proficiencyBonus int
		expected         int
	}{
		{"not proficient, +3 ability", false, 3, 2, 3},
		{"not proficient, -1 ability", false, -1, 4, -1},
		{"proficient, +2 ability, +2 prof", true, 2, 2, 4},
		{"proficient, +4 ability, +3 prof", true, 4, 3, 7},
		{"proficient, -1 ability, +2 prof", true, -1, 2, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			save := SavingThrow{Proficient: tt.proficient}
			got := CalculateSavingThrowModifier(save, tt.abilityMod, tt.proficiencyBonus)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestCalculateSavingThrowModifierDefault(t *testing.T) {
	// Test with default (not proficient) save
	save := SavingThrow{}
	got := CalculateSavingThrowModifier(save, 3, 2)
	assert.Equal(t, 3, got)
}
