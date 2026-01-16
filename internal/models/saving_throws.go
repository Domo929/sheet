package models

// SavingThrow represents proficiency in a saving throw.
type SavingThrow struct {
	Proficient bool `json:"proficient"`
}

// SavingThrows contains all six D&D 5e saving throws.
type SavingThrows struct {
	Strength     SavingThrow `json:"strength"`
	Dexterity    SavingThrow `json:"dexterity"`
	Constitution SavingThrow `json:"constitution"`
	Intelligence SavingThrow `json:"intelligence"`
	Wisdom       SavingThrow `json:"wisdom"`
	Charisma     SavingThrow `json:"charisma"`
}

// NewSavingThrows creates a new SavingThrows with no proficiencies.
func NewSavingThrows() SavingThrows {
	return SavingThrows{}
}

// Get returns the SavingThrow for the given ability.
func (s *SavingThrows) Get(ability Ability) *SavingThrow {
	switch ability {
	case AbilityStrength:
		return &s.Strength
	case AbilityDexterity:
		return &s.Dexterity
	case AbilityConstitution:
		return &s.Constitution
	case AbilityIntelligence:
		return &s.Intelligence
	case AbilityWisdom:
		return &s.Wisdom
	case AbilityCharisma:
		return &s.Charisma
	default:
		return nil
	}
}

// SetProficiency sets the proficiency for a saving throw.
func (s *SavingThrows) SetProficiency(ability Ability, proficient bool) {
	save := s.Get(ability)
	if save != nil {
		save.Proficient = proficient
	}
}

// IsProficient returns whether the character is proficient in the given saving throw.
func (s *SavingThrows) IsProficient(ability Ability) bool {
	save := s.Get(ability)
	if save == nil {
		return false
	}
	return save.Proficient
}

// CalculateSavingThrowModifier calculates the total modifier for a saving throw.
func CalculateSavingThrowModifier(save *SavingThrow, abilityMod int, proficiencyBonus int) int {
	if save == nil {
		return abilityMod
	}

	if save.Proficient {
		return abilityMod + proficiencyBonus
	}
	return abilityMod
}
