package models

// AbilityScore represents a single ability score with its base value and any temporary modifiers.
type AbilityScore struct {
	Base      int `json:"base"`
	Temporary int `json:"temporary,omitempty"`
}

// Total returns the total ability score (base + temporary).
func (a AbilityScore) Total() int {
	return a.Base + a.Temporary
}

// Modifier calculates the ability modifier using D&D 5e rules: (score - 10) / 2, rounded down.
// Go's integer division truncates toward zero, but D&D requires rounding toward negative infinity.
func (a AbilityScore) Modifier() int {
	diff := a.Total() - 10
	if diff >= 0 {
		return diff / 2
	}
	// For negative values, we need floor division
	return (diff - 1) / 2
}

// AbilityScores contains all six D&D 5e ability scores.
type AbilityScores struct {
	Strength     AbilityScore `json:"strength"`
	Dexterity    AbilityScore `json:"dexterity"`
	Constitution AbilityScore `json:"constitution"`
	Intelligence AbilityScore `json:"intelligence"`
	Wisdom       AbilityScore `json:"wisdom"`
	Charisma     AbilityScore `json:"charisma"`
}

// NewAbilityScores creates a new AbilityScores with all base values set to 10.
func NewAbilityScores() AbilityScores {
	return AbilityScores{
		Strength:     AbilityScore{Base: 10},
		Dexterity:    AbilityScore{Base: 10},
		Constitution: AbilityScore{Base: 10},
		Intelligence: AbilityScore{Base: 10},
		Wisdom:       AbilityScore{Base: 10},
		Charisma:     AbilityScore{Base: 10},
	}
}

// NewAbilityScoresFromValues creates AbilityScores from individual values.
func NewAbilityScoresFromValues(str, dex, con, intel, wis, cha int) AbilityScores {
	return AbilityScores{
		Strength:     AbilityScore{Base: str},
		Dexterity:    AbilityScore{Base: dex},
		Constitution: AbilityScore{Base: con},
		Intelligence: AbilityScore{Base: intel},
		Wisdom:       AbilityScore{Base: wis},
		Charisma:     AbilityScore{Base: cha},
	}
}

// StandardArray returns the D&D 5e standard array values.
func StandardArray() []int {
	return []int{15, 14, 13, 12, 10, 8}
}

// Ability represents which ability score to reference.
type Ability string

const (
	AbilityStrength     Ability = "strength"
	AbilityDexterity    Ability = "dexterity"
	AbilityConstitution Ability = "constitution"
	AbilityIntelligence Ability = "intelligence"
	AbilityWisdom       Ability = "wisdom"
	AbilityCharisma     Ability = "charisma"
)

// Get returns the AbilityScore for the given ability.
func (a AbilityScores) Get(ability Ability) AbilityScore {
	switch ability {
	case AbilityStrength:
		return a.Strength
	case AbilityDexterity:
		return a.Dexterity
	case AbilityConstitution:
		return a.Constitution
	case AbilityIntelligence:
		return a.Intelligence
	case AbilityWisdom:
		return a.Wisdom
	case AbilityCharisma:
		return a.Charisma
	default:
		return AbilityScore{Base: 10}
	}
}

// Set sets the AbilityScore for the given ability.
func (a *AbilityScores) Set(ability Ability, score AbilityScore) {
	switch ability {
	case AbilityStrength:
		a.Strength = score
	case AbilityDexterity:
		a.Dexterity = score
	case AbilityConstitution:
		a.Constitution = score
	case AbilityIntelligence:
		a.Intelligence = score
	case AbilityWisdom:
		a.Wisdom = score
	case AbilityCharisma:
		a.Charisma = score
	}
}

// SetBase sets the base value for the given ability.
func (a *AbilityScores) SetBase(ability Ability, base int) {
	score := a.Get(ability)
	score.Base = base
	a.Set(ability, score)
}

// SetTemporary sets the temporary modifier for the given ability.
func (a *AbilityScores) SetTemporary(ability Ability, temp int) {
	score := a.Get(ability)
	score.Temporary = temp
	a.Set(ability, score)
}

// GetModifier returns the modifier for the given ability.
func (a AbilityScores) GetModifier(ability Ability) int {
	return a.Get(ability).Modifier()
}
