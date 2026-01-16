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
func (s SavingThrows) Get(ability Ability) SavingThrow {
	switch ability {
	case AbilityStrength:
		return s.Strength
	case AbilityDexterity:
		return s.Dexterity
	case AbilityConstitution:
		return s.Constitution
	case AbilityIntelligence:
		return s.Intelligence
	case AbilityWisdom:
		return s.Wisdom
	case AbilityCharisma:
		return s.Charisma
	default:
		return SavingThrow{}
	}
}

// SetProficiency sets the proficiency for a saving throw.
func (s *SavingThrows) SetProficiency(ability Ability, proficient bool) {
	switch ability {
	case AbilityStrength:
		s.Strength.Proficient = proficient
	case AbilityDexterity:
		s.Dexterity.Proficient = proficient
	case AbilityConstitution:
		s.Constitution.Proficient = proficient
	case AbilityIntelligence:
		s.Intelligence.Proficient = proficient
	case AbilityWisdom:
		s.Wisdom.Proficient = proficient
	case AbilityCharisma:
		s.Charisma.Proficient = proficient
	}
}

// IsProficient returns whether the character is proficient in the given saving throw.
func (s SavingThrows) IsProficient(ability Ability) bool {
	return s.Get(ability).Proficient
}

// CalculateSavingThrowModifier calculates the total modifier for a saving throw.
func CalculateSavingThrowModifier(save SavingThrow, abilityMod int, proficiencyBonus int) int {
	if save.Proficient {
		return abilityMod + proficiencyBonus
	}
	return abilityMod
}
