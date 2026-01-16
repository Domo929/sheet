package data

// Race represents a playable race in D&D 5e.
type Race struct {
	Name         string     `json:"name"`
	CreatureType string     `json:"creatureType"`
	Size         string     `json:"size"`
	Speed        int        `json:"speed"`
	Traits       []Trait    `json:"traits"`
	Languages    []string   `json:"languages"`
	Spells       []Spell    `json:"spells,omitempty"`
	Subtypes     []Subtype  `json:"subtypes,omitempty"`
}

// Trait represents a racial or class trait.
type Trait struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Spell represents a spell granted by race or class.
type Spell struct {
	Level int    `json:"level"`
	Spell string `json:"spell"`
	Type  string `json:"type"`
}

// Subtype represents a subrace or variant.
type Subtype struct {
	Name   string  `json:"name"`
	Traits []Trait `json:"traits"`
}

// Class represents a playable class in D&D 5e.
type Class struct {
	Name                     string        `json:"name"`
	HitDice                  string        `json:"hitDice"`
	PrimaryAbility           []string      `json:"primaryAbility"`
	SavingThrowProficiencies []string      `json:"savingThrowProficiencies"`
	ArmorProficiencies       []string      `json:"armorProficiencies"`
	WeaponProficiencies      []string      `json:"weaponProficiencies"`
	ToolProficiencies        []string      `json:"toolProficiencies"`
	SkillChoices             SkillChoices  `json:"skillChoices"`
	StartingEquipment        []string      `json:"startingEquipment"`
	Spellcaster              bool          `json:"spellcaster"`
	SpellcastingAbility      string        `json:"spellcastingAbility,omitempty"`
	SpellSlots               []SpellSlot   `json:"spellSlots,omitempty"`
	Features                 []Feature     `json:"features"`
	Subclasses               []Subclass    `json:"subclasses,omitempty"`
}

// SpellSlot represents spell slots available at a given level.
type SpellSlot struct {
	Level int `json:"level"`
	First int `json:"1st"`
	Second int `json:"2nd"`
	Third int `json:"3rd"`
	Fourth int `json:"4th"`
	Fifth int `json:"5th"`
	Sixth int `json:"6th"`
	Seventh int `json:"7th"`
	Eighth int `json:"8th"`
	Ninth int `json:"9th"`
}

// SkillChoices represents the skill proficiency selection for a class.
type SkillChoices struct {
	Count   int      `json:"count"`
	Options []string `json:"options"`
}

// Feature represents a class feature.
type Feature struct {
	Level       int    `json:"level"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Subclass represents a class specialization.
type Subclass struct {
	Name     string    `json:"name"`
	Features []Feature `json:"features"`
}

// SpellData represents a spell in the game.
type SpellData struct {
	Name        string            `json:"name"`
	Level       int               `json:"level"`
	School      string            `json:"school"`
	CastingTime string            `json:"castingTime"`
	Range       string            `json:"range"`
	Components  []string          `json:"components"`
	Duration    string            `json:"duration"`
	Description string            `json:"description"`
	Classes     []string          `json:"classes"`
	Ritual      bool              `json:"ritual"`
	Damage      string            `json:"damage,omitempty"`
	DamageType  string            `json:"damageType,omitempty"`
	SavingThrow string            `json:"savingThrow,omitempty"`
	Scaling     map[string]string `json:"scaling,omitempty"`
}

// Background represents a character background.
type Background struct {
	Name               string            `json:"name"`
	Description        string            `json:"description"`
	AbilityScores      AbilityScoreBonus `json:"abilityScores"`
	SkillProficiencies []string          `json:"skillProficiencies"`
	ToolProficiency    string            `json:"toolProficiency"`
	Equipment          []string          `json:"equipment"`
	Feat               string            `json:"feat"`
}

// AbilityScoreBonus represents ability score options for backgrounds.
type AbilityScoreBonus struct {
	Options []string `json:"options"`
	Points  int      `json:"points"`
}

// Condition represents a game condition.
type Condition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// RaceData contains all race data.
type RaceData struct {
	Races []Race `json:"races"`
}

// ClassData contains all class data.
type ClassData struct {
	Classes []Class `json:"classes"`
}

// SpellDatabase contains all spell data.
type SpellDatabase struct {
	Spells []SpellData `json:"spells"`
}

// BackgroundData contains all background data.
type BackgroundData struct {
	Backgrounds []Background `json:"backgrounds"`
}

// ConditionData contains all condition data.
type ConditionData struct {
	Conditions []Condition `json:"conditions"`
}
