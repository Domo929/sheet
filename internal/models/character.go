package models

import (
	"encoding/json"
	"time"
)

// Character is the main data structure representing a D&D 5e character.
type Character struct {
	// Metadata
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Core character information
	Info          CharacterInfo `json:"info"`
	AbilityScores AbilityScores `json:"abilityScores"`
	Skills        Skills        `json:"skills"`
	SavingThrows  SavingThrows  `json:"savingThrows"`
	CombatStats   CombatStats   `json:"combatStats"`
	Inventory     Inventory     `json:"inventory"`
	Spellcasting  *Spellcasting `json:"spellcasting,omitempty"`
	Features      Features      `json:"features"`
	Proficiencies Proficiencies `json:"proficiencies"`
	Personality   Personality   `json:"personality"`
}

// NewCharacter creates a new character with the given basic information.
func NewCharacter(id, name, race, class string) *Character {
	now := time.Now()
	return &Character{
		ID:            id,
		CreatedAt:     now,
		UpdatedAt:     now,
		Info:          NewCharacterInfo(name, race, class),
		AbilityScores: NewAbilityScores(),
		Skills:        NewSkills(),
		SavingThrows:  NewSavingThrows(),
		CombatStats:   NewCombatStats(10, 8, 1, 30), // Default values
		Inventory:     NewInventory(),
		Features:      NewFeatures(),
		Proficiencies: NewProficiencies(),
		Personality:   NewPersonality(),
	}
}

// MarkUpdated updates the UpdatedAt timestamp.
func (c *Character) MarkUpdated() {
	c.UpdatedAt = time.Now()
}

// GetProficiencyBonus returns the character's proficiency bonus.
func (c *Character) GetProficiencyBonus() int {
	return c.Info.ProficiencyBonus()
}

// GetSkillModifier calculates the total modifier for a skill check.
func (c *Character) GetSkillModifier(skillName SkillName) int {
	skill := c.Skills.Get(skillName)
	ability := GetSkillAbility(skillName)
	abilityMod := c.AbilityScores.GetModifier(ability)
	return CalculateSkillModifier(skill, abilityMod, c.GetProficiencyBonus())
}

// GetSavingThrowModifier calculates the total modifier for a saving throw.
func (c *Character) GetSavingThrowModifier(ability Ability) int {
	save := c.SavingThrows.Get(ability)
	abilityMod := c.AbilityScores.GetModifier(ability)
	return CalculateSavingThrowModifier(save, abilityMod, c.GetProficiencyBonus())
}

// GetSpellSaveDC returns the character's spell save DC.
func (c *Character) GetSpellSaveDC() int {
	if c.Spellcasting == nil {
		return 0
	}
	abilityMod := c.AbilityScores.GetModifier(c.Spellcasting.Ability)
	return CalculateSpellSaveDC(abilityMod, c.GetProficiencyBonus())
}

// GetSpellAttackBonus returns the character's spell attack bonus.
func (c *Character) GetSpellAttackBonus() int {
	if c.Spellcasting == nil {
		return 0
	}
	abilityMod := c.AbilityScores.GetModifier(c.Spellcasting.Ability)
	return CalculateSpellAttackBonus(abilityMod, c.GetProficiencyBonus())
}

// GetInitiative returns the character's initiative modifier.
func (c *Character) GetInitiative() int {
	return c.AbilityScores.GetModifier(AbilityDexterity)
}

// IsSpellcaster returns true if the character has spellcasting ability.
func (c *Character) IsSpellcaster() bool {
	return c.Spellcasting != nil
}

// TakeDamage applies damage to the character.
func (c *Character) TakeDamage(damage int) {
	c.CombatStats.HitPoints.TakeDamage(damage)
	c.MarkUpdated()
}

// Heal restores hit points to the character.
func (c *Character) Heal(amount int) {
	c.CombatStats.HitPoints.Heal(amount)
	c.MarkUpdated()
}

// ShortRest performs a short rest.
func (c *Character) ShortRest() {
	// Restore Warlock pact magic slots
	if c.Spellcasting != nil && c.Spellcasting.PactMagic != nil {
		c.Spellcasting.PactMagic.Restore()
	}
	c.MarkUpdated()
}

// LongRest performs a long rest.
func (c *Character) LongRest() {
	// Restore all HP
	c.CombatStats.HitPoints.Current = c.CombatStats.HitPoints.Maximum
	c.CombatStats.HitPoints.Temporary = 0

	// Restore hit dice (half of total, minimum 1)
	c.CombatStats.HitDice.RecoverOnLongRest()

	// Restore all spell slots
	if c.Spellcasting != nil {
		c.Spellcasting.SpellSlots.RestoreAll()
		if c.Spellcasting.PactMagic != nil {
			c.Spellcasting.PactMagic.Restore()
		}
	}

	// Reset death saves
	c.CombatStats.DeathSaves.Reset()

	// Reduce exhaustion by 1
	if c.CombatStats.ExhaustionLevel > 0 {
		c.CombatStats.ExhaustionLevel--
	}

	c.MarkUpdated()
}

// LevelUp increases the character's level and adjusts stats accordingly.
// Note: This does NOT automatically increase HP or spell slots, as those depend
// on class-specific rules that require external data (class features, hit dice rolls, etc.).
// Use this for milestone leveling or manual level increases.
// The caller should handle:
// - Rolling/taking average for HP increase (hit die + CON modifier)
// - Updating spell slot progression based on class
// - Adding new class features
// - Updating hit dice total
func (c *Character) LevelUp() bool {
	if !c.Info.LevelUp() {
		return false // Already at max level
	}

	// Update hit dice total (assume same die type)
	c.CombatStats.HitDice.Total = c.Info.Level

	c.MarkUpdated()
	return true
}

// ToJSON serializes the character to JSON.
func (c *Character) ToJSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

// FromJSON deserializes a character from JSON.
func FromJSON(data []byte) (*Character, error) {
	var c Character
	err := json.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Validate checks if the character data is valid.
func (c *Character) Validate() []string {
	errors := []string{}

	if c.ID == "" {
		errors = append(errors, "character ID is required")
	}
	if c.Info.Name == "" {
		errors = append(errors, "character name is required")
	}
	if c.Info.Race == "" {
		errors = append(errors, "character race is required")
	}
	if c.Info.Class == "" {
		errors = append(errors, "character class is required")
	}
	if c.Info.Level < 1 || c.Info.Level > 20 {
		errors = append(errors, "character level must be between 1 and 20")
	}

	return errors
}
