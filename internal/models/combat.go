package models

// HitPoints tracks the character's hit points.
type HitPoints struct {
	Maximum   int `json:"maximum"`
	Current   int `json:"current"`
	Temporary int `json:"temporary"`
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

// Heal increases current HP by the given amount, up to maximum.
// Healing does not affect temporary HP.
func (hp *HitPoints) Heal(amount int) {
	if amount <= 0 {
		return
	}

	hp.Current += amount
	if hp.Current > hp.Maximum {
		hp.Current = hp.Maximum
	}
}

// AddTemporaryHP sets temporary HP if the new value is higher than current temp HP.
// Temp HP doesn't stack per 5e rules.
func (hp *HitPoints) AddTemporaryHP(amount int) {
	if amount > hp.Temporary {
		hp.Temporary = amount
	}
}

// IsUnconscious returns true if current HP is 0.
func (hp *HitPoints) IsUnconscious() bool {
	return hp.Current <= 0
}

// HitDice tracks hit dice available for short rest healing.
type HitDice struct {
	Total     int    `json:"total"`
	Remaining int    `json:"remaining"`
	DieType   string `json:"dieType"` // e.g., "d8", "d10", "d12"
}

// NewHitDice creates HitDice for a character of the given level with the specified die type.
func NewHitDice(level int, dieType string) HitDice {
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
func NewCombatStats(maxHP int, hitDieType string, level int, speed int) CombatStats {
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
