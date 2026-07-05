package models

import (
	"fmt"
	"strings"
)

// ClassLevel represents levels taken in a single class, used to describe a
// multiclassed character. HitDie is the die size (e.g., 10 for a d10).
type ClassLevel struct {
	Class    string `json:"class"`
	Subclass string `json:"subclass,omitempty"`
	Level    int    `json:"level"`
	HitDie   int    `json:"hitDie,omitempty"`
}

// CasterProgression describes how a class contributes to the multiclass
// spellcaster level.
type CasterProgression int

const (
	CasterNone  CasterProgression = iota
	CasterFull                    // Bard, Cleric, Druid, Sorcerer, Wizard
	CasterHalf                    // Paladin, Ranger
	CasterThird                   // Eldritch Knight, Arcane Trickster subclasses
	CasterPact                    // Warlock (separate Pact Magic)
)

// classCasterProgression returns the multiclass spell-slot progression for a
// class/subclass pair.
func classCasterProgression(class, subclass string) CasterProgression {
	switch strings.ToLower(strings.TrimSpace(class)) {
	case "bard", "cleric", "druid", "sorcerer", "wizard":
		return CasterFull
	case "paladin", "ranger":
		return CasterHalf
	case "warlock":
		return CasterPact
	case "fighter":
		if strings.EqualFold(strings.TrimSpace(subclass), "Eldritch Knight") {
			return CasterThird
		}
	case "rogue":
		if strings.EqualFold(strings.TrimSpace(subclass), "Arcane Trickster") {
			return CasterThird
		}
	}
	return CasterNone
}

// casterLevelContribution returns the number of effective caster levels a class
// entry contributes to the shared multiclass spell-slot table.
func casterLevelContribution(cl ClassLevel) int {
	switch classCasterProgression(cl.Class, cl.Subclass) {
	case CasterFull:
		return cl.Level
	case CasterHalf:
		return cl.Level / 2
	case CasterThird:
		return cl.Level / 3
	default:
		return 0
	}
}

// multiclassSlotTable maps a combined caster level (index 0-20) to the number of
// spell slots available at spell levels 1-9. This is the standard shared
// full-caster progression from the Player's Handbook multiclassing rules.
var multiclassSlotTable = [21][9]int{
	{},                          // 0 (no caster levels)
	{2},                         // 1
	{3},                         // 2
	{4, 2},                      // 3
	{4, 3},                      // 4
	{4, 3, 2},                   // 5
	{4, 3, 3},                   // 6
	{4, 3, 3, 1},                // 7
	{4, 3, 3, 2},                // 8
	{4, 3, 3, 3, 1},             // 9
	{4, 3, 3, 3, 2},             // 10
	{4, 3, 3, 3, 2, 1},          // 11
	{4, 3, 3, 3, 2, 1},          // 12
	{4, 3, 3, 3, 2, 1, 1},       // 13
	{4, 3, 3, 3, 2, 1, 1},       // 14
	{4, 3, 3, 3, 2, 1, 1, 1},    // 15
	{4, 3, 3, 3, 2, 1, 1, 1},    // 16
	{4, 3, 3, 3, 2, 1, 1, 1, 1}, // 17
	{4, 3, 3, 3, 3, 1, 1, 1, 1}, // 18
	{4, 3, 3, 3, 3, 2, 1, 1, 1}, // 19
	{4, 3, 3, 3, 3, 2, 2, 1, 1}, // 20
}

// MulticlassCasterLevel returns the combined spellcaster level across all class
// entries, using the standard multiclassing fractions (full, half, third).
// Warlock levels are excluded (Pact Magic is tracked separately).
func MulticlassCasterLevel(classes []ClassLevel) int {
	total := 0
	for _, cl := range classes {
		total += casterLevelContribution(cl)
	}
	if total > 20 {
		total = 20
	}
	return total
}

// MulticlassSpellSlots returns the number of spell slots per spell level (1-9)
// for a given combined caster level.
func MulticlassSpellSlots(casterLevel int) [9]int {
	if casterLevel < 0 {
		casterLevel = 0
	}
	if casterLevel > 20 {
		casterLevel = 20
	}
	return multiclassSlotTable[casterLevel]
}

// IsMulticlass reports whether the character has more than one class entry.
func (c *Character) IsMulticlass() bool {
	return len(c.Classes) > 1
}

// TotalLevel returns the character's total level: the sum of all class levels
// when multiclass data is present, otherwise the single-class Info.Level.
func (c *Character) TotalLevel() int {
	if len(c.Classes) == 0 {
		return c.Info.Level
	}
	total := 0
	for _, cl := range c.Classes {
		total += cl.Level
	}
	return total
}

// ClassSummary returns a human-readable class/level summary, e.g.
// "Fighter 5 / Wizard 3" for a multiclass character, or "Wizard 5" for a
// single-class character.
func (c *Character) ClassSummary() string {
	if len(c.Classes) == 0 {
		if c.Info.Level > 0 {
			return fmt.Sprintf("%s %d", c.Info.Class, c.Info.Level)
		}
		return c.Info.Class
	}
	parts := make([]string, 0, len(c.Classes))
	for _, cl := range c.Classes {
		parts = append(parts, fmt.Sprintf("%s %d", cl.Class, cl.Level))
	}
	return strings.Join(parts, " / ")
}

// SyncPrimaryClass keeps the single-class Info fields aligned with the first
// class entry and Info.Level aligned with the total level. This preserves
// backward compatibility for code paths that read Info.Class/Info.Level.
func (c *Character) SyncPrimaryClass() {
	if len(c.Classes) == 0 {
		return
	}
	c.Info.Class = c.Classes[0].Class
	c.Info.Subclass = c.Classes[0].Subclass
	c.Info.Level = c.TotalLevel()
}

// ApplyMulticlassSpellSlots recomputes the character's regular spell slots from
// the combined multiclass caster level and applies them, preserving the number
// of already-expended slots where possible. Pact Magic (Warlock) is left
// untouched. It is a no-op for single-class characters.
func (c *Character) ApplyMulticlassSpellSlots() {
	if !c.IsMulticlass() {
		return
	}
	casterLevel := MulticlassCasterLevel(c.Classes)
	if casterLevel == 0 {
		return
	}
	if c.Spellcasting == nil {
		sc := NewSpellcasting(Ability(""))
		c.Spellcasting = &sc
	}
	slots := MulticlassSpellSlots(casterLevel)
	for lvl := 1; lvl <= 9; lvl++ {
		total := slots[lvl-1]
		tracker := c.Spellcasting.SpellSlots.GetSlot(lvl)
		if tracker == nil {
			continue
		}
		used := tracker.Total - tracker.Remaining
		if used < 0 {
			used = 0
		}
		tracker.Total = total
		tracker.Remaining = total - used
		if tracker.Remaining < 0 {
			tracker.Remaining = 0
		}
		if tracker.Remaining > total {
			tracker.Remaining = total
		}
	}
}
