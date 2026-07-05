package models

import (
	"sort"
	"strings"

	"github.com/Domo929/sheet/internal/domain"
)

// WeaponMasteryLimit returns how many weapons this character may have Weapon
// Mastery with, based on class and level under the 2024 rules.
func (c *Character) WeaponMasteryLimit() int {
	return domain.WeaponMasteryCount(c.Info.Class, c.Info.Level)
}

// HasWeaponMastery reports whether the character has chosen Weapon Mastery with
// the named weapon.
func (c *Character) HasWeaponMastery(name string) bool {
	for _, w := range c.MasteredWeapons {
		if strings.EqualFold(w, name) {
			return true
		}
	}
	return false
}

// ToggleWeaponMastery adds or removes the named weapon from the character's
// mastered weapons. Adding is refused (returning false) when the character is
// already at its Weapon Mastery limit. Removing always succeeds.
func (c *Character) ToggleWeaponMastery(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	for i, w := range c.MasteredWeapons {
		if strings.EqualFold(w, name) {
			c.MasteredWeapons = append(c.MasteredWeapons[:i], c.MasteredWeapons[i+1:]...)
			return true
		}
	}
	if len(c.MasteredWeapons) >= c.WeaponMasteryLimit() {
		return false
	}
	c.MasteredWeapons = append(c.MasteredWeapons, name)
	sort.Strings(c.MasteredWeapons)
	return true
}
