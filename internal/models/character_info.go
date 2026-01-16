package models

// ProgressionType indicates how the character progresses in level.
type ProgressionType string

const (
	ProgressionXP        ProgressionType = "xp"
	ProgressionMilestone ProgressionType = "milestone"
)

// Alignment represents D&D alignment.
type Alignment string

const (
	AlignmentLawfulGood     Alignment = "lawfulGood"
	AlignmentNeutralGood    Alignment = "neutralGood"
	AlignmentChaoticGood    Alignment = "chaoticGood"
	AlignmentLawfulNeutral  Alignment = "lawfulNeutral"
	AlignmentTrueNeutral    Alignment = "trueNeutral"
	AlignmentChaoticNeutral Alignment = "chaoticNeutral"
	AlignmentLawfulEvil     Alignment = "lawfulEvil"
	AlignmentNeutralEvil    Alignment = "neutralEvil"
	AlignmentChaoticEvil    Alignment = "chaoticEvil"
)

// CharacterInfo contains basic character information.
type CharacterInfo struct {
	Name             string          `json:"name"`
	PlayerName       string          `json:"playerName,omitempty"`
	Race             string          `json:"race"`
	Subrace          string          `json:"subrace,omitempty"`
	Class            string          `json:"class"`
	Subclass         string          `json:"subclass,omitempty"`
	Level            int             `json:"level"`
	Background       string          `json:"background"`
	Alignment        Alignment       `json:"alignment,omitempty"`
	ExperiencePoints int             `json:"experiencePoints"`
	ProgressionType  ProgressionType `json:"progressionType"`
	Inspiration      bool            `json:"inspiration"`
}

// NewCharacterInfo creates a new character info with defaults.
func NewCharacterInfo(name, race, class string) CharacterInfo {
	return CharacterInfo{
		Name:            name,
		Race:            race,
		Class:           class,
		Level:           1,
		ProgressionType: ProgressionXP,
	}
}

// XPThresholds maps character level to required XP for that level.
var XPThresholds = map[int]int{
	1:  0,
	2:  300,
	3:  900,
	4:  2700,
	5:  6500,
	6:  14000,
	7:  23000,
	8:  34000,
	9:  48000,
	10: 64000,
	11: 85000,
	12: 100000,
	13: 120000,
	14: 140000,
	15: 165000,
	16: 195000,
	17: 225000,
	18: 265000,
	19: 305000,
	20: 355000,
}

// XPForNextLevel returns the XP required to reach the next level.
func XPForNextLevel(currentLevel int) int {
	if currentLevel >= 20 {
		return 0 // Max level
	}
	return XPThresholds[currentLevel+1]
}

// CanLevelUp returns true if the character has enough XP to level up.
func (ci *CharacterInfo) CanLevelUp() bool {
	if ci.ProgressionType != ProgressionXP {
		return false
	}
	if ci.Level >= 20 {
		return false
	}
	return ci.ExperiencePoints >= XPThresholds[ci.Level+1]
}

// AddXP adds experience points to the character.
func (ci *CharacterInfo) AddXP(amount int) {
	ci.ExperiencePoints += amount
}

// LevelUp increases the character level if possible.
func (ci *CharacterInfo) LevelUp() bool {
	if ci.Level >= 20 {
		return false
	}
	ci.Level++
	return true
}

// ProficiencyBonus returns the proficiency bonus for the character's level.
func (ci *CharacterInfo) ProficiencyBonus() int {
	return ProficiencyBonusByLevel(ci.Level)
}

// ProficiencyBonusByLevel returns the proficiency bonus for a given level.
func ProficiencyBonusByLevel(level int) int {
	if level < 1 {
		return 2
	}
	if level > 20 {
		level = 20
	}
	return (level-1)/4 + 2
}

// Personality contains character personality and roleplay information.
type Personality struct {
	Traits    []string `json:"traits,omitempty"`
	Ideals    []string `json:"ideals,omitempty"`
	Bonds     []string `json:"bonds,omitempty"`
	Flaws     []string `json:"flaws,omitempty"`
	Backstory string   `json:"backstory,omitempty"`
	Notes     string   `json:"notes,omitempty"`
}

// NewPersonality creates an empty personality.
func NewPersonality() Personality {
	return Personality{
		Traits: []string{},
		Ideals: []string{},
		Bonds:  []string{},
		Flaws:  []string{},
	}
}

// AddTrait adds a personality trait.
func (p *Personality) AddTrait(trait string) {
	p.Traits = append(p.Traits, trait)
}

// AddIdeal adds an ideal.
func (p *Personality) AddIdeal(ideal string) {
	p.Ideals = append(p.Ideals, ideal)
}

// AddBond adds a bond.
func (p *Personality) AddBond(bond string) {
	p.Bonds = append(p.Bonds, bond)
}

// AddFlaw adds a flaw.
func (p *Personality) AddFlaw(flaw string) {
	p.Flaws = append(p.Flaws, flaw)
}
