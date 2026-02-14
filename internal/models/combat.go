package models

import "github.com/Domo929/sheet/internal/domain"

// HitPoints tracks the character's hit points.
type HitPoints struct {
	Maximum          int `json:"maximum"`
	Current          int `json:"current"`
	Temporary        int `json:"temporary"`
	TemporaryMaximum int `json:"temporaryMaximum,omitempty"` // From spells like Aid
}

// NewHitPoints creates HitPoints with the given maximum (current set to max).
func NewHitPoints(maximum int) HitPoints {
	return HitPoints{
		Maximum: maximum,
		Current: maximum,
	}
}

// TakeDamage reduces current HP by the given amount, applying to temp HP first.
// Returns the actual damage taken (after temp HP absorption).
func (hp *HitPoints) TakeDamage(damage int) int {
	if damage <= 0 {
		return 0
	}

	// Apply to temporary HP first
	if hp.Temporary > 0 {
		if damage <= hp.Temporary {
			hp.Temporary -= damage
			return 0
		}
		damage -= hp.Temporary
		hp.Temporary = 0
	}

	// Apply remaining damage to current HP
	actualDamage := damage
	hp.Current -= damage
	if hp.Current < 0 {
		hp.Current = 0
	}

	return actualDamage
}

// Heal increases current HP by the given amount, up to effective maximum (including temporary).
// Healing does not affect temporary HP.
func (hp *HitPoints) Heal(amount int) {
	if amount <= 0 {
		return
	}

	hp.Current += amount
	effectiveMax := hp.Maximum + hp.TemporaryMaximum
	if hp.Current > effectiveMax {
		hp.Current = effectiveMax
	}
}

// AddTemporaryHP sets temporary HP if the new value is higher than current temp HP.
// Temp HP doesn't stack per 5e rules.
func (hp *HitPoints) AddTemporaryHP(amount int) {
	if amount > hp.Temporary {
		hp.Temporary = amount
	}
}

// AddTemporaryMaximum increases the maximum HP temporarily (e.g., Aid spell).
// Unlike temp HP, these increases stack with each casting.
func (hp *HitPoints) AddTemporaryMaximum(amount int) {
	if amount > 0 {
		hp.TemporaryMaximum += amount
		// Also increase current HP by the same amount
		hp.Current += amount
	}
}

// RemoveTemporaryMaximum removes temporary maximum HP.
// If current HP exceeds the new maximum, it's reduced to the new maximum.
func (hp *HitPoints) RemoveTemporaryMaximum(amount int) {
	if amount > 0 {
		hp.TemporaryMaximum -= amount
		if hp.TemporaryMaximum < 0 {
			hp.TemporaryMaximum = 0
		}
		effectiveMax := hp.Maximum + hp.TemporaryMaximum
		if hp.Current > effectiveMax {
			hp.Current = effectiveMax
		}
	}
}

// IsUnconscious returns true if current HP is 0.
func (hp *HitPoints) IsUnconscious() bool {
	return hp.Current <= 0
}

// HitDice tracks hit dice available for short rest healing.
type HitDice struct {
	Total     int `json:"total"`
	Remaining int `json:"remaining"`
	DieType   int `json:"dieType"` // e.g., 6, 8, 10, 12
}

// NewHitDice creates HitDice for a character of the given level with the specified die type.
func NewHitDice(level int, dieType int) HitDice {
	return HitDice{
		Total:     level,
		Remaining: level,
		DieType:   dieType,
	}
}

// Use spends a hit die if available. Returns true if successful.
func (hd *HitDice) Use() bool {
	if hd.Remaining > 0 {
		hd.Remaining--
		return true
	}
	return false
}

// RecoverOnLongRest recovers hit dice after a long rest (half of total, minimum 1).
func (hd *HitDice) RecoverOnLongRest() int {
	toRecover := hd.Total / 2
	if toRecover < 1 {
		toRecover = 1
	}

	recovered := 0
	for i := 0; i < toRecover && hd.Remaining < hd.Total; i++ {
		hd.Remaining++
		recovered++
	}
	return recovered
}

// DeathSaves tracks death saving throw successes and failures.
type DeathSaves struct {
	Successes int `json:"successes"`
	Failures  int `json:"failures"`
}

// NewDeathSaves creates a fresh DeathSaves tracker.
func NewDeathSaves() DeathSaves {
	return DeathSaves{}
}

// AddSuccess adds a death save success. Returns true if character is stabilized (3 successes).
func (ds *DeathSaves) AddSuccess() bool {
	ds.Successes++
	if ds.Successes > 3 {
		ds.Successes = 3
	}
	return ds.IsStabilized()
}

// AddFailure adds a death save failure. Returns true if character is dead (3 failures).
func (ds *DeathSaves) AddFailure() bool {
	ds.Failures++
	if ds.Failures > 3 {
		ds.Failures = 3
	}
	return ds.IsDead()
}

// AddCriticalFailure adds two failures (for rolling a natural 1). Returns true if dead.
func (ds *DeathSaves) AddCriticalFailure() bool {
	ds.Failures += 2
	if ds.Failures > 3 {
		ds.Failures = 3
	}
	return ds.IsDead()
}

// IsStabilized returns true if the character has 3 successes.
func (ds *DeathSaves) IsStabilized() bool {
	return ds.Successes >= 3
}

// IsDead returns true if the character has 3 failures.
func (ds *DeathSaves) IsDead() bool {
	return ds.Failures >= 3
}

// Reset resets death saves (when stabilized or healed).
func (ds *DeathSaves) Reset() {
	ds.Successes = 0
	ds.Failures = 0
}

// Condition represents an active condition affecting the character.
type Condition string

const (
	ConditionBlinded       Condition = "blinded"
	ConditionCharmed       Condition = "charmed"
	ConditionDeafened      Condition = "deafened"
	ConditionExhaustion    Condition = "exhaustion"
	ConditionFrightened    Condition = "frightened"
	ConditionGrappled      Condition = "grappled"
	ConditionIncapacitated Condition = "incapacitated"
	ConditionInvisible     Condition = "invisible"
	ConditionParalyzed     Condition = "paralyzed"
	ConditionPetrified     Condition = "petrified"
	ConditionPoisoned      Condition = "poisoned"
	ConditionProne         Condition = "prone"
	ConditionRestrained    Condition = "restrained"
	ConditionStunned       Condition = "stunned"
	ConditionUnconscious   Condition = "unconscious"
)

// AllConditions returns all standard D&D 5e conditions.
func AllConditions() []Condition {
	return []Condition{
		ConditionBlinded, ConditionCharmed, ConditionDeafened, ConditionExhaustion,
		ConditionFrightened, ConditionGrappled, ConditionIncapacitated, ConditionInvisible,
		ConditionParalyzed, ConditionPetrified, ConditionPoisoned, ConditionProne,
		ConditionRestrained, ConditionStunned, ConditionUnconscious,
	}
}

// CombatStats contains all combat-related statistics.
type CombatStats struct {
	HitPoints       HitPoints   `json:"hitPoints"`
	HitDice         HitDice     `json:"hitDice"`
	ArmorClass      int         `json:"armorClass"`
	Initiative      int         `json:"initiative"`
	Speed           int         `json:"speed"`
	DeathSaves      DeathSaves  `json:"deathSaves"`
	Conditions      []Condition `json:"conditions,omitempty"`
	ExhaustionLevel int         `json:"exhaustionLevel,omitempty"`
}

// NewCombatStats creates combat stats with the given parameters.
func NewCombatStats(maxHP int, hitDieType int, level int, speed int) CombatStats {
	return CombatStats{
		HitPoints:  NewHitPoints(maxHP),
		HitDice:    NewHitDice(level, hitDieType),
		ArmorClass: 10, // Base AC
		Speed:      speed,
		DeathSaves: NewDeathSaves(),
		Conditions: []Condition{},
	}
}

// AddCondition adds a condition if not already present.
func (cs *CombatStats) AddCondition(condition Condition) {
	for _, c := range cs.Conditions {
		if c == condition {
			return
		}
	}
	cs.Conditions = append(cs.Conditions, condition)
}

// RemoveCondition removes a condition if present.
func (cs *CombatStats) RemoveCondition(condition Condition) {
	for i, c := range cs.Conditions {
		if c == condition {
			cs.Conditions = append(cs.Conditions[:i], cs.Conditions[i+1:]...)
			return
		}
	}
}

// HasCondition returns true if the character has the given condition.
func (cs *CombatStats) HasCondition(condition Condition) bool {
	for _, c := range cs.Conditions {
		if c == condition {
			return true
		}
	}
	return false
}

// ClearConditions removes all conditions.
func (cs *CombatStats) ClearConditions() {
	cs.Conditions = []Condition{}
}

// CalculateAC calculates the total AC from base AC, armor, shield, and modifiers.
// Base AC is typically 10 + DEX modifier when unarmored.
// With armor: use armor's AC + DEX modifier (capped based on armor type).
// Additional modifiers can come from magic items, spells, or class features.
func CalculateAC(baseAC int, armorAC int, shieldBonus int, dexModifier int, additionalModifiers int, maxDexBonus int) int {
	ac := baseAC

	// If wearing armor, use armor's AC instead of base
	if armorAC > 0 {
		ac = armorAC

		// Apply DEX modifier up to max allowed by armor
		if maxDexBonus == -1 {
			// No limit (e.g., light armor)
			ac += dexModifier
		} else if maxDexBonus > 0 {
			// Limited DEX (e.g., medium armor allows max +2)
			if dexModifier > maxDexBonus {
				ac += maxDexBonus
			} else {
				ac += dexModifier
			}
		}
		// maxDexBonus == 0 means no DEX bonus (heavy armor)
	} else {
		// No armor, use base AC with full DEX modifier
		ac += dexModifier
	}

	// Add shield bonus
	ac += shieldBonus

	// Add any additional modifiers (magic items, spells, etc.)
	ac += additionalModifiers

	return ac
}

// CalculateAttackBonus calculates the attack bonus for a weapon.
// strMod and dexMod are the character's STR and DEX modifiers.
// proficiencyBonus is added if the character is proficient with the weapon.
// magicBonus is from magic weapons (+1, +2, +3, etc.).
// weaponProps contains the weapon's properties (e.g., "finesse", "thrown").
// useStrength can force using STR even for ranged/finesse weapons.
func CalculateAttackBonus(strMod int, dexMod int, proficient bool, proficiencyBonus int, magicBonus int, weaponProps []domain.WeaponProperty, useStrength bool) int {
	abilityMod := strMod // Default to STR for melee weapons

	// Check if weapon has finesse property
	hasFinesse := false
	for _, prop := range weaponProps {
		if prop == domain.PropertyFinesse {
			hasFinesse = true
			break
		}
	}

	// If finesse weapon, use higher of STR or DEX (unless forced to use STR)
	if hasFinesse && !useStrength {
		if dexMod > strMod {
			abilityMod = dexMod
		}
	}

	attackBonus := abilityMod

	// Add proficiency bonus if proficient
	if proficient {
		attackBonus += proficiencyBonus
	}

	// Add magic bonus
	attackBonus += magicBonus

	return attackBonus
}
