package models

import "fmt"

// CompanionKind categorizes a companion stat block.
type CompanionKind string

const (
	CompanionPet       CompanionKind = "Companion"
	CompanionFamiliar  CompanionKind = "Familiar"
	CompanionSummon    CompanionKind = "Summon"
	CompanionWildShape CompanionKind = "Wild Shape"
	CompanionMount     CompanionKind = "Mount"
)

// AllCompanionKinds returns the selectable companion kinds in display order.
func AllCompanionKinds() []CompanionKind {
	return []CompanionKind{
		CompanionPet,
		CompanionFamiliar,
		CompanionSummon,
		CompanionWildShape,
		CompanionMount,
	}
}

// CompanionAttack is a single attack action on a companion stat block.
type CompanionAttack struct {
	Name   string `json:"name"`
	Bonus  int    `json:"bonus"`            // to-hit bonus
	Damage string `json:"damage,omitempty"` // e.g., "1d6 + 2 piercing"
	Notes  string `json:"notes,omitempty"`
}

// CompanionTrait is a named special ability or feature on a companion.
type CompanionTrait struct {
	Name string `json:"name"`
	Text string `json:"text,omitempty"`
}

// Companion is a stat block for a pet, familiar, summon, mount, or Wild Shape
// form tracked alongside the character.
type Companion struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Kind      CompanionKind `json:"kind"`
	Size      string        `json:"size,omitempty"`
	Type      string        `json:"type,omitempty"` // creature type, e.g., "Beast"
	AC        int           `json:"ac"`
	MaxHP     int           `json:"maxHP"`
	CurrentHP int           `json:"currentHP"`
	TempHP    int           `json:"tempHP,omitempty"`
	Speed     string        `json:"speed,omitempty"`

	// Abilities holds STR, DEX, CON, INT, WIS, CHA in that order.
	Abilities [6]int `json:"abilities"`

	Attacks []CompanionAttack `json:"attacks,omitempty"`
	Traits  []CompanionTrait  `json:"traits,omitempty"`
	Notes   string            `json:"notes,omitempty"`
}

// abilityLabels maps the Abilities array indices to their short labels.
var companionAbilityLabels = [6]string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}

// CompanionAbilityLabel returns the short label for ability index i (0-5).
func CompanionAbilityLabel(i int) string {
	if i < 0 || i >= len(companionAbilityLabels) {
		return "?"
	}
	return companionAbilityLabels[i]
}

// AbilityModifier returns the D&D ability modifier for a raw score.
func AbilityModifier(score int) int {
	// Floor division toward negative infinity for scores below 10.
	if score < 10 {
		return -((10 - score + 1) / 2)
	}
	return (score - 10) / 2
}

// Modifier returns the ability modifier for the ability at index i (0-5).
func (c *Companion) Modifier(i int) int {
	if i < 0 || i >= len(c.Abilities) {
		return 0
	}
	return AbilityModifier(c.Abilities[i])
}

// FormatModifier renders a modifier with an explicit sign, e.g. "+3" or "-1".
func FormatModifier(mod int) string {
	if mod >= 0 {
		return fmt.Sprintf("+%d", mod)
	}
	return fmt.Sprintf("%d", mod)
}

// Damage applies dmg to the companion, absorbing temp HP first and clamping
// current HP at 0.
func (c *Companion) Damage(dmg int) {
	if dmg <= 0 {
		return
	}
	if c.TempHP > 0 {
		if dmg <= c.TempHP {
			c.TempHP -= dmg
			return
		}
		dmg -= c.TempHP
		c.TempHP = 0
	}
	c.CurrentHP -= dmg
	if c.CurrentHP < 0 {
		c.CurrentHP = 0
	}
}

// Heal restores HP up to the companion's maximum.
func (c *Companion) Heal(amount int) {
	if amount <= 0 {
		return
	}
	c.CurrentHP += amount
	if c.CurrentHP > c.MaxHP {
		c.CurrentHP = c.MaxHP
	}
}

// AddCompanion appends a companion, initializing current HP to max if unset.
func (c *Character) AddCompanion(comp Companion) {
	if comp.CurrentHP == 0 {
		comp.CurrentHP = comp.MaxHP
	}
	c.Companions = append(c.Companions, comp)
}

// RemoveCompanion deletes the companion with the given ID, returning true if a
// companion was removed.
func (c *Character) RemoveCompanion(id string) bool {
	for i := range c.Companions {
		if c.Companions[i].ID == id {
			c.Companions = append(c.Companions[:i], c.Companions[i+1:]...)
			return true
		}
	}
	return false
}
