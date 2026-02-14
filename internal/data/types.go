package data

import (
	"github.com/Domo929/sheet/internal/domain"
)

// CastingTime represents the casting time of a spell.
type CastingTime string

const (
	CastingTimeAction      CastingTime = "A"
	CastingTimeBonusAction CastingTime = "BA"
	CastingTimeReaction    CastingTime = "R"
	CastingTimeOneMinute   CastingTime = "1 minute"
)

// SpellSchool represents the school of magic for a spell.
type SpellSchool string

const (
	SchoolAbjuration    SpellSchool = "Abjuration"
	SchoolConjuration   SpellSchool = "Conjuration"
	SchoolDivination    SpellSchool = "Divination"
	SchoolEnchantment   SpellSchool = "Enchantment"
	SchoolEvocation     SpellSchool = "Evocation"
	SchoolIllusion      SpellSchool = "Illusion"
	SchoolNecromancy    SpellSchool = "Necromancy"
	SchoolTransmutation SpellSchool = "Transmutation"
)

// SpellComponent represents a spell component type.
type SpellComponent string

const (
	ComponentVerbal   SpellComponent = "V"
	ComponentSomatic  SpellComponent = "S"
	ComponentMaterial SpellComponent = "M"
)

// ComponentsToStrings converts a slice of SpellComponent to a slice of string.
func ComponentsToStrings(c []SpellComponent) []string {
	result := make([]string, len(c))
	for i, comp := range c {
		result[i] = string(comp)
	}
	return result
}

// EquipmentCategory represents the category of an equipment item.
type EquipmentCategory string

const (
	CategoryWeapon EquipmentCategory = "weapon"
	CategoryArmor  EquipmentCategory = "armor"
	CategoryPack   EquipmentCategory = "pack"
	CategoryGear   EquipmentCategory = "gear"
	CategoryTool   EquipmentCategory = "tool"
)

// EquipmentChoiceType represents the type of equipment choice.
type EquipmentChoiceType string

const (
	EquipmentChoiceFixed  EquipmentChoiceType = "fixed"
	EquipmentChoiceSelect EquipmentChoiceType = "choice"
)

// WeaponType represents whether a weapon is simple or martial.
type WeaponType string

const (
	WeaponTypeSimple  WeaponType = "simple"
	WeaponTypeMartial WeaponType = "martial"
)

// WeaponStyle represents the style of a weapon (melee or ranged).
type WeaponStyle string

const (
	WeaponStyleMelee  WeaponStyle = "melee"
	WeaponStyleRanged WeaponStyle = "ranged"
)

// ArmorCategory represents the category of armor.
type ArmorCategory string

const (
	ArmorCategoryLight  ArmorCategory = "light"
	ArmorCategoryMedium ArmorCategory = "medium"
	ArmorCategoryHeavy  ArmorCategory = "heavy"
	ArmorCategoryShield ArmorCategory = "shield"
)

// FeatCategory represents the category of a feat.
type FeatCategory string

const (
	FeatCategoryOrigin   FeatCategory = "Origin"
	FeatCategoryGeneral  FeatCategory = "General"
	FeatCategoryFighting FeatCategory = "Fighting"
	FeatCategoryEpicBoon FeatCategory = "Epic Boon"
)

// Race represents a playable race in D&D 5e.
type Race struct {
	Name         string    `json:"name"`
	CreatureType string    `json:"creatureType"`
	Size         string    `json:"size"`
	Speed        int       `json:"speed"`
	Traits       []Trait   `json:"traits"`
	Languages    []string  `json:"languages"`
	Spells       []Spell   `json:"spells,omitempty"`
	Subtypes     []Subtype `json:"subtypes,omitempty"`
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
	Name                     string            `json:"name"`
	HitDice                  string            `json:"hitDice"`
	PrimaryAbility           []string          `json:"primaryAbility"`
	SavingThrowProficiencies []string          `json:"savingThrowProficiencies"`
	ArmorProficiencies       []string          `json:"armorProficiencies"`
	WeaponProficiencies      []string          `json:"weaponProficiencies"`
	ToolProficiencies        []string          `json:"toolProficiencies"`
	SkillChoices             SkillChoices      `json:"skillChoices"`
	StartingEquipment        []EquipmentChoice `json:"startingEquipment"`
	Spellcaster              bool              `json:"spellcaster"`
	SpellcastingAbility      string            `json:"spellcastingAbility,omitempty"`
	SpellSlots               []SpellSlot       `json:"spellSlots,omitempty"`
	RitualCaster             bool              `json:"ritualCaster,omitempty"`             // Can cast ritual spells
	RitualCasterUnprepared   bool              `json:"ritualCasterUnprepared,omitempty"`   // Can cast rituals without preparing (Wizard)
	Features                 []Feature         `json:"features"`
	Subclasses               []Subclass        `json:"subclasses,omitempty"`
}

// EquipmentChoice represents a starting equipment item or choice.
type EquipmentChoice struct {
	Type    EquipmentChoiceType `json:"type"` // "fixed" or "choice"
	Item    *EquipmentItem     `json:"item,omitempty"` // For fixed items
	Options []EquipmentOption  `json:"options,omitempty"` // For choices
}

// EquipmentItem represents a specific equipment item with quantity.
type EquipmentItem struct {
	Name     string           `json:"name"`              // "Greataxe", "Explorer's Pack", etc. (or empty if using filter)
	Quantity int              `json:"quantity"`          // Number of items
	Category EquipmentCategory `json:"category"`          // "weapon", "armor", "pack", "gear", "tool"
	Filter   *EquipmentFilter `json:"filter,omitempty"`  // Optional filter for dynamic selection (e.g., "any martial melee weapon")
}

// EquipmentFilter specifies criteria for dynamic equipment selection.
type EquipmentFilter struct {
	WeaponType   WeaponType    `json:"weaponType,omitempty"`   // "simple" or "martial"
	WeaponStyle  WeaponStyle   `json:"weaponStyle,omitempty"`  // "melee" or "ranged"
	ArmorType    ArmorCategory `json:"armorType,omitempty"`    // "light", "medium", "heavy", "shield"
}

// EquipmentOption represents one option in an equipment choice.
type EquipmentOption struct {
	Items []EquipmentItem `json:"items"` // One or more items in this option
}

// SpellSlot represents spell slots available at a given level.
type SpellSlot struct {
	Level   int `json:"level"`
	First   int `json:"1st"`
	Second  int `json:"2nd"`
	Third   int `json:"3rd"`
	Fourth  int `json:"4th"`
	Fifth   int `json:"5th"`
	Sixth   int `json:"6th"`
	Seventh int `json:"7th"`
	Eighth  int `json:"8th"`
	Ninth   int `json:"9th"`
}

// SkillChoices represents the skill proficiency selection for a class.
type SkillChoices struct {
	Count   int      `json:"count"`
	Options []string `json:"options"`
}

// Feature represents a class feature.
type Feature struct {
	Level       int                  `json:"level"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Activation  domain.ActivationType `json:"activation,omitempty"` // "action", "bonus", "reaction", or "" (passive)
}

// Subclass represents a class specialization.
type Subclass struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	URL         string    `json:"url,omitempty"`
	Features    []Feature `json:"features"`
}

// SpellData represents a spell in the game.
type SpellData struct {
	Name        string            `json:"name"`
	Level       int               `json:"level"`
	School      SpellSchool       `json:"school"`
	CastingTime CastingTime       `json:"castingTime"`
	Range       string            `json:"range"`
	Components  []SpellComponent  `json:"components"`
	Duration    string            `json:"duration"`
	Description string            `json:"description"`
	Classes     []string          `json:"classes"`
	Ritual      bool              `json:"ritual"`
	Damage      string            `json:"damage,omitempty"`
	DamageType  domain.DamageType `json:"damageType,omitempty"`
	SavingThrow string            `json:"savingThrow,omitempty"`
	Scaling     map[string]string `json:"scaling,omitempty"`
	Upcast      string            `json:"upcast,omitempty"`
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

// Feat represents a feat available for character selection.
type Feat struct {
	Name         string       `json:"name"`
	Category     FeatCategory `json:"category"`              // "Origin", "General", "Fighting", "Epic Boon"
	Prerequisite string       `json:"prerequisite,omitempty"` // Human-readable prerequisite text
	Repeatable   bool         `json:"repeatable,omitempty"`   // Whether the feat can be taken multiple times
	Description  string       `json:"description"`
	Effects      FeatEffect   `json:"effects"`
}

// FeatEffect represents the mechanical effects of a feat.
type FeatEffect struct {
	AbilityScoreIncrease *FeatASI `json:"abilityScoreIncrease,omitempty"` // +1 to a choice of abilities
	InitiativeBonus      int      `json:"initiativeBonus,omitempty"`
	SpeedBonus           int      `json:"speedBonus,omitempty"`
	HPPerLevel           int      `json:"hpPerLevel,omitempty"` // Tough feat: +2 HP per level
	ACBonus              int      `json:"acBonus,omitempty"`
}

// FeatASI represents an ability score increase granted by a feat.
type FeatASI struct {
	Options []string `json:"options"` // Which abilities can be increased
	Amount  int      `json:"amount"`  // How much to increase by (usually 1)
}

// FeatData contains all feat data.
type FeatData struct {
	Feats []Feat `json:"feats"`
}
