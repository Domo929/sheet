package domain

import "strings"

// WeaponMastery represents the mastery property of a weapon in the 2024
// (5.5e) rules. Each weapon has exactly one mastery property, which a
// character can use only if a class feature grants them mastery with that
// weapon.
type WeaponMastery string

const (
	MasteryNone   WeaponMastery = ""
	MasteryCleave WeaponMastery = "cleave"
	MasteryGraze  WeaponMastery = "graze"
	MasteryNick   WeaponMastery = "nick"
	MasteryPush   WeaponMastery = "push"
	MasterySap    WeaponMastery = "sap"
	MasterySlow   WeaponMastery = "slow"
	MasteryTopple WeaponMastery = "topple"
	MasteryVex    WeaponMastery = "vex"
)

// Description returns a short rules summary of the mastery property.
func (m WeaponMastery) Description() string {
	switch m {
	case MasteryCleave:
		return "On a hit, make one extra attack against a second creature within 5 ft of the first and within reach (no ability modifier to the extra damage). Once per turn."
	case MasteryGraze:
		return "On a miss, deal damage equal to the ability modifier used for the attack (same damage type as the weapon)."
	case MasteryNick:
		return "Make the extra attack from the Light property as part of the Attack action instead of a Bonus Action. Once per turn."
	case MasteryPush:
		return "On a hit, push the target up to 10 ft straight away from you if it is Large or smaller."
	case MasterySap:
		return "On a hit, the target has Disadvantage on its next attack roll before the start of your next turn."
	case MasterySlow:
		return "On a hit, reduce the target's Speed by 10 ft until the start of your next turn (does not stack)."
	case MasteryTopple:
		return "On a hit, force a Constitution saving throw (DC 8 + ability modifier + proficiency bonus) or the target has the Prone condition."
	case MasteryVex:
		return "On a hit, you have Advantage on your next attack roll against the same target before the end of your next turn."
	default:
		return ""
	}
}

// Label returns the display label (title case) for the mastery property.
func (m WeaponMastery) Label() string {
	switch m {
	case MasteryCleave:
		return "Cleave"
	case MasteryGraze:
		return "Graze"
	case MasteryNick:
		return "Nick"
	case MasteryPush:
		return "Push"
	case MasterySap:
		return "Sap"
	case MasterySlow:
		return "Slow"
	case MasteryTopple:
		return "Topple"
	case MasteryVex:
		return "Vex"
	default:
		return ""
	}
}

// WeaponMasteryCount returns how many weapons a character of the given class and
// level may have Weapon Mastery with under the 2024 rules. Fighters scale with
// level (3 at 1st, 4 at 4th, 5 at 10th, 6 at 16th); Barbarians, Paladins,
// Rangers, and Rogues get 2; all other classes get none.
func WeaponMasteryCount(className string, level int) int {
	switch strings.ToLower(strings.TrimSpace(className)) {
	case "fighter":
		switch {
		case level >= 16:
			return 6
		case level >= 10:
			return 5
		case level >= 4:
			return 4
		default:
			return 3
		}
	case "barbarian", "paladin", "ranger", "rogue":
		return 2
	default:
		return 0
	}
}
