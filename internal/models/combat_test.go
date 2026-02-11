package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHitPointsTakeDamage(t *testing.T) {
	tests := []struct {
		name           string
		initial        HitPoints
		damage         int
		expectedHP     int
		expectedTemp   int
		expectedActual int
	}{
		{
			name:           "simple damage",
			initial:        HitPoints{Maximum: 20, Current: 20, Temporary: 0},
			damage:         5,
			expectedHP:     15,
			expectedTemp:   0,
			expectedActual: 5,
		},
		{
			name:           "damage absorbed by temp HP",
			initial:        HitPoints{Maximum: 20, Current: 20, Temporary: 10},
			damage:         5,
			expectedHP:     20,
			expectedTemp:   5,
			expectedActual: 0,
		},
		{
			name:           "damage partially absorbed by temp HP",
			initial:        HitPoints{Maximum: 20, Current: 20, Temporary: 3},
			damage:         8,
			expectedHP:     15,
			expectedTemp:   0,
			expectedActual: 5,
		},
		{
			name:           "damage exceeds current HP",
			initial:        HitPoints{Maximum: 20, Current: 5, Temporary: 0},
			damage:         10,
			expectedHP:     0,
			expectedTemp:   0,
			expectedActual: 10,
		},
		{
			name:           "zero damage",
			initial:        HitPoints{Maximum: 20, Current: 15, Temporary: 0},
			damage:         0,
			expectedHP:     15,
			expectedTemp:   0,
			expectedActual: 0,
		},
		{
			name:           "negative damage",
			initial:        HitPoints{Maximum: 20, Current: 15, Temporary: 0},
			damage:         -5,
			expectedHP:     15,
			expectedTemp:   0,
			expectedActual: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hp := tt.initial
			actual := hp.TakeDamage(tt.damage)
			assert.Equal(t, tt.expectedHP, hp.Current, "Current HP")
			assert.Equal(t, tt.expectedTemp, hp.Temporary, "Temporary HP")
			assert.Equal(t, tt.expectedActual, actual, "Actual damage")
		})
	}
}

func TestHitPointsHeal(t *testing.T) {
	tests := []struct {
		name       string
		initial    HitPoints
		heal       int
		expectedHP int
	}{
		{
			name:       "simple heal",
			initial:    HitPoints{Maximum: 20, Current: 10},
			heal:       5,
			expectedHP: 15,
		},
		{
			name:       "heal to max",
			initial:    HitPoints{Maximum: 20, Current: 15},
			heal:       10,
			expectedHP: 20,
		},
		{
			name:       "heal at max",
			initial:    HitPoints{Maximum: 20, Current: 20},
			heal:       5,
			expectedHP: 20,
		},
		{
			name:       "zero heal",
			initial:    HitPoints{Maximum: 20, Current: 10},
			heal:       0,
			expectedHP: 10,
		},
		{
			name:       "negative heal",
			initial:    HitPoints{Maximum: 20, Current: 10},
			heal:       -5,
			expectedHP: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hp := tt.initial
			hp.Heal(tt.heal)
			assert.Equal(t, tt.expectedHP, hp.Current, "Current HP")
		})
	}
}

func TestHitPointsAddTemporaryHP(t *testing.T) {
	hp := HitPoints{Maximum: 20, Current: 20, Temporary: 5}

	// Lower temp HP should not replace
	hp.AddTemporaryHP(3)
	assert.Equal(t, 5, hp.Temporary, "should not decrease")

	// Higher temp HP should replace
	hp.AddTemporaryHP(10)
	assert.Equal(t, 10, hp.Temporary)
}

func TestHitPointsIsUnconscious(t *testing.T) {
	hp := HitPoints{Maximum: 20, Current: 5}
	assert.False(t, hp.IsUnconscious(), "Should not be unconscious at 5 HP")

	hp.Current = 0
	assert.True(t, hp.IsUnconscious(), "Should be unconscious at 0 HP")
}

func TestHitDiceUse(t *testing.T) {
	hd := NewHitDice(5, 8)

	for i := 0; i < 5; i++ {
		assert.True(t, hd.Use(), "Use() should succeed when %d remaining", hd.Remaining+1)
	}

	assert.Equal(t, 0, hd.Remaining)
	assert.False(t, hd.Use(), "Use() should fail when 0 remaining")
}

func TestHitDiceRecoverOnLongRest(t *testing.T) {
	tests := []struct {
		name              string
		total             int
		remaining         int
		expectedRecovered int
		expectedRemaining int
	}{
		{"level 6, 0 remaining", 6, 0, 3, 3},
		{"level 6, 3 remaining", 6, 3, 3, 6},
		{"level 6, 5 remaining", 6, 5, 1, 6},
		{"level 1, 0 remaining", 1, 0, 1, 1}, // minimum 1
		{"level 3, 0 remaining", 3, 0, 1, 1}, // 3/2 = 1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hd := HitDice{Total: tt.total, Remaining: tt.remaining, DieType: 8}
			recovered := hd.RecoverOnLongRest()
			assert.Equal(t, tt.expectedRecovered, recovered, "Recovered")
			assert.Equal(t, tt.expectedRemaining, hd.Remaining, "Remaining")
		})
	}
}

func TestDeathSaves(t *testing.T) {
	ds := NewDeathSaves()

	// Test successes
	ds.AddSuccess()
	ds.AddSuccess()
	assert.False(t, ds.IsStabilized(), "Should not be stabilized with 2 successes")

	stabilized := ds.AddSuccess()
	assert.True(t, stabilized, "Should be stabilized with 3 successes")
	assert.True(t, ds.IsStabilized(), "Should be stabilized with 3 successes")

	// Reset and test failures
	ds.Reset()
	assert.Equal(t, 0, ds.Successes, "Reset should clear successes")
	assert.Equal(t, 0, ds.Failures, "Reset should clear failures")

	ds.AddFailure()
	ds.AddFailure()
	assert.False(t, ds.IsDead(), "Should not be dead with 2 failures")

	dead := ds.AddFailure()
	assert.True(t, dead, "Should be dead with 3 failures")
	assert.True(t, ds.IsDead(), "Should be dead with 3 failures")
}

func TestDeathSavesCriticalFailure(t *testing.T) {
	ds := NewDeathSaves()
	ds.AddFailure()
	dead := ds.AddCriticalFailure()
	assert.True(t, dead, "Critical failure after 1 failure should result in death")
	assert.Equal(t, 3, ds.Failures)
}

func TestCombatStatsConditions(t *testing.T) {
	cs := NewCombatStats(20, 8, 1, 30)

	// Add condition
	cs.AddCondition(ConditionPoisoned)
	assert.True(t, cs.HasCondition(ConditionPoisoned), "Should have poisoned condition")

	// Adding same condition shouldn't duplicate
	cs.AddCondition(ConditionPoisoned)
	assert.Len(t, cs.Conditions, 1)

	// Add another condition
	cs.AddCondition(ConditionFrightened)
	assert.Len(t, cs.Conditions, 2)

	// Remove condition
	cs.RemoveCondition(ConditionPoisoned)
	assert.False(t, cs.HasCondition(ConditionPoisoned), "Should not have poisoned condition after removal")
	assert.True(t, cs.HasCondition(ConditionFrightened), "Should still have frightened condition")

	// Clear all
	cs.ClearConditions()
	assert.Empty(t, cs.Conditions, "Should have no conditions after clear")
}

func TestAllConditionsCount(t *testing.T) {
	conditions := AllConditions()
	assert.Len(t, conditions, 15)
}

func TestCalculateAC(t *testing.T) {
	tests := []struct {
		name                string
		baseAC              int
		armorAC             int
		shieldBonus         int
		dexModifier         int
		additionalModifiers int
		maxDexBonus         int
		expected            int
	}{
		{
			name:                "unarmored with +2 DEX",
			baseAC:              10,
			armorAC:             0,
			shieldBonus:         0,
			dexModifier:         2,
			additionalModifiers: 0,
			maxDexBonus:         -1,
			expected:            12,
		},
		{
			name:                "light armor (no DEX limit) with +3 DEX",
			baseAC:              10,
			armorAC:             11,
			shieldBonus:         0,
			dexModifier:         3,
			additionalModifiers: 0,
			maxDexBonus:         -1,
			expected:            14,
		},
		{
			name:                "medium armor (max +2 DEX) with +3 DEX",
			baseAC:              10,
			armorAC:             14,
			shieldBonus:         0,
			dexModifier:         3,
			additionalModifiers: 0,
			maxDexBonus:         2,
			expected:            16,
		},
		{
			name:                "heavy armor (no DEX) with +2 DEX",
			baseAC:              10,
			armorAC:             18,
			shieldBonus:         0,
			dexModifier:         2,
			additionalModifiers: 0,
			maxDexBonus:         0,
			expected:            18,
		},
		{
			name:                "armor with shield",
			baseAC:              10,
			armorAC:             16,
			shieldBonus:         2,
			dexModifier:         1,
			additionalModifiers: 0,
			maxDexBonus:         2,
			expected:            19,
		},
		{
			name:                "armor with magic bonus",
			baseAC:              10,
			armorAC:             14,
			shieldBonus:         0,
			dexModifier:         2,
			additionalModifiers: 1,
			maxDexBonus:         2,
			expected:            17,
		},
		{
			name:                "full setup with all bonuses",
			baseAC:              10,
			armorAC:             15,
			shieldBonus:         2,
			dexModifier:         3,
			additionalModifiers: 2,
			maxDexBonus:         2,
			expected:            21,
		},
		{
			name:                "negative DEX modifier unarmored",
			baseAC:              10,
			armorAC:             0,
			shieldBonus:         0,
			dexModifier:         -1,
			additionalModifiers: 0,
			maxDexBonus:         -1,
			expected:            9,
		},
		{
			name:                "medium armor with +1 DEX (below cap)",
			baseAC:              10,
			armorAC:             14,
			shieldBonus:         0,
			dexModifier:         1,
			additionalModifiers: 0,
			maxDexBonus:         2,
			expected:            15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateAC(tt.baseAC, tt.armorAC, tt.shieldBonus, tt.dexModifier, tt.additionalModifiers, tt.maxDexBonus)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestCalculateAttackBonus(t *testing.T) {
	tests := []struct {
		name             string
		strMod           int
		dexMod           int
		proficient       bool
		proficiencyBonus int
		magicBonus       int
		weaponProps      []string
		useStrength      bool
		expected         int
	}{
		{
			name:             "melee weapon, proficient, no magic",
			strMod:           3,
			dexMod:           1,
			proficient:       true,
			proficiencyBonus: 2,
			magicBonus:       0,
			weaponProps:      []string{},
			useStrength:      false,
			expected:         5,
		},
		{
			name:             "finesse weapon using DEX",
			strMod:           2,
			dexMod:           4,
			proficient:       true,
			proficiencyBonus: 2,
			magicBonus:       0,
			weaponProps:      []string{string(PropertyFinesse)},
			useStrength:      false,
			expected:         6,
		},
		{
			name:             "finesse weapon using STR (higher)",
			strMod:           4,
			dexMod:           2,
			proficient:       true,
			proficiencyBonus: 2,
			magicBonus:       0,
			weaponProps:      []string{string(PropertyFinesse)},
			useStrength:      false,
			expected:         6,
		},
		{
			name:             "finesse weapon forced to use STR",
			strMod:           2,
			dexMod:           4,
			proficient:       true,
			proficiencyBonus: 2,
			magicBonus:       0,
			weaponProps:      []string{string(PropertyFinesse)},
			useStrength:      true,
			expected:         4,
		},
		{
			name:             "magic weapon +1",
			strMod:           3,
			dexMod:           1,
			proficient:       true,
			proficiencyBonus: 3,
			magicBonus:       1,
			weaponProps:      []string{},
			useStrength:      false,
			expected:         7,
		},
		{
			name:             "not proficient",
			strMod:           3,
			dexMod:           1,
			proficient:       false,
			proficiencyBonus: 2,
			magicBonus:       0,
			weaponProps:      []string{},
			useStrength:      false,
			expected:         3,
		},
		{
			name:             "finesse light weapon",
			strMod:           2,
			dexMod:           4,
			proficient:       true,
			proficiencyBonus: 2,
			magicBonus:       0,
			weaponProps:      []string{string(PropertyFinesse), string(PropertyLight)},
			useStrength:      false,
			expected:         6,
		},
		{
			name:             "magic weapon +3, high level",
			strMod:           5,
			dexMod:           2,
			proficient:       true,
			proficiencyBonus: 6,
			magicBonus:       3,
			weaponProps:      []string{},
			useStrength:      false,
			expected:         14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateAttackBonus(tt.strMod, tt.dexMod, tt.proficient, tt.proficiencyBonus, tt.magicBonus, tt.weaponProps, tt.useStrength)
			assert.Equal(t, tt.expected, got)
		})
	}
}
