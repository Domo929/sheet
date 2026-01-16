package models

import "testing"

func TestNewSavingThrows(t *testing.T) {
	saves := NewSavingThrows()

	// All should be not proficient
	abilities := []Ability{
		AbilityStrength, AbilityDexterity, AbilityConstitution,
		AbilityIntelligence, AbilityWisdom, AbilityCharisma,
	}

	for _, ability := range abilities {
		if saves.IsProficient(ability) {
			t.Errorf("NewSavingThrows().IsProficient(%s) = true, want false", ability)
		}
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
		if save == nil {
			t.Errorf("SavingThrows.Get(%s) returned nil", ability)
		}
	}

	// Check specific one we modified
	if !saves.Get(AbilityWisdom).Proficient {
		t.Error("Wisdom save should be proficient")
	}
}

func TestSavingThrowsGetInvalid(t *testing.T) {
	saves := NewSavingThrows()
	if got := saves.Get(Ability("invalid")); got != nil {
		t.Errorf("SavingThrows.Get(invalid) = %v, want nil", got)
	}
}

func TestSavingThrowsSetProficiency(t *testing.T) {
	saves := NewSavingThrows()

	saves.SetProficiency(AbilityConstitution, true)
	if !saves.Constitution.Proficient {
		t.Error("Constitution save should be proficient after SetProficiency")
	}

	saves.SetProficiency(AbilityConstitution, false)
	if saves.Constitution.Proficient {
		t.Error("Constitution save should not be proficient after SetProficiency(false)")
	}
}

func TestSavingThrowsIsProficient(t *testing.T) {
	saves := NewSavingThrows()
	saves.Dexterity.Proficient = true

	if !saves.IsProficient(AbilityDexterity) {
		t.Error("IsProficient(Dexterity) should be true")
	}

	if saves.IsProficient(AbilityStrength) {
		t.Error("IsProficient(Strength) should be false")
	}

	// Invalid ability should return false
	if saves.IsProficient(Ability("invalid")) {
		t.Error("IsProficient(invalid) should be false")
	}
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
			save := &SavingThrow{Proficient: tt.proficient}
			got := CalculateSavingThrowModifier(save, tt.abilityMod, tt.proficiencyBonus)
			if got != tt.expected {
				t.Errorf("CalculateSavingThrowModifier() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestCalculateSavingThrowModifierNil(t *testing.T) {
	got := CalculateSavingThrowModifier(nil, 3, 2)
	if got != 3 {
		t.Errorf("CalculateSavingThrowModifier(nil, 3, 2) = %d, want 3", got)
	}
}
