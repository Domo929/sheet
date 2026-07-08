package models

import "strings"

// RollCategory identifies the kind of d20 roll used when evaluating
// condition-driven advantage or disadvantage.
type RollCategory int

const (
	// RollCatAttack is an attack roll made by the affected creature.
	RollCatAttack RollCategory = iota
	// RollCatAbilityCheck is an ability check or skill check.
	RollCatAbilityCheck
	// RollCatSavingThrow is a saving throw.
	RollCatSavingThrow
)

// ConditionRollEffect reports whether the active conditions impose advantage
// and/or disadvantage on a d20 roll of the given category made BY the affected
// creature, along with a short human-readable reason listing the contributing
// conditions. The ability argument only matters for saving throws (for example,
// Restrained imposes disadvantage on Dexterity saving throws); pass an empty
// Ability for attack rolls and ability checks. Effects follow the 2024 (5.5e)
// condition rules.
//
// Only condition effects that apply to the affected creature's own d20 rolls
// and are not target-specific are modeled. Exhaustion is intentionally excluded
// because 2024 exhaustion is a flat -2 per level penalty to d20 tests handled
// separately, not advantage/disadvantage.
//
// Per the rules, holding both advantage and disadvantage from any combination
// of sources cancels to a normal roll; callers should combine this result with
// any situational advantage/disadvantage before rolling.
func ConditionRollEffect(conditions []Condition, cat RollCategory, ability Ability) (advantage, disadvantage bool, reason string) {
	var reasons []string
	seen := map[string]bool{}
	note := func(name string) {
		if !seen[name] {
			seen[name] = true
			reasons = append(reasons, name)
		}
	}

	for _, c := range conditions {
		switch c {
		case ConditionPoisoned:
			// Disadvantage on attack rolls and ability checks.
			if cat == RollCatAttack || cat == RollCatAbilityCheck {
				disadvantage = true
				note("Poisoned")
			}
		case ConditionFrightened:
			// Disadvantage on ability checks and attack rolls while the
			// source of fear is in sight.
			if cat == RollCatAttack || cat == RollCatAbilityCheck {
				disadvantage = true
				note("Frightened")
			}
		case ConditionProne:
			// Disadvantage on your own attack rolls.
			if cat == RollCatAttack {
				disadvantage = true
				note("Prone")
			}
		case ConditionBlinded:
			// Can't see: disadvantage on attack rolls.
			if cat == RollCatAttack {
				disadvantage = true
				note("Blinded")
			}
		case ConditionRestrained:
			// Disadvantage on attack rolls and on Dexterity saving throws.
			if cat == RollCatAttack {
				disadvantage = true
				note("Restrained")
			}
			if cat == RollCatSavingThrow && ability == AbilityDexterity {
				disadvantage = true
				note("Restrained")
			}
		case ConditionInvisible:
			// Advantage on attack rolls.
			if cat == RollCatAttack {
				advantage = true
				note("Invisible")
			}
		}
	}

	reason = strings.Join(reasons, ", ")
	return advantage, disadvantage, reason
}
