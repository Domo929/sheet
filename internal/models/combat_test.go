package models

import "testing"

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
			if hp.Current != tt.expectedHP {
				t.Errorf("Current HP = %d, want %d", hp.Current, tt.expectedHP)
			}
			if hp.Temporary != tt.expectedTemp {
				t.Errorf("Temporary HP = %d, want %d", hp.Temporary, tt.expectedTemp)
			}
			if actual != tt.expectedActual {
				t.Errorf("Actual damage = %d, want %d", actual, tt.expectedActual)
			}
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
			if hp.Current != tt.expectedHP {
				t.Errorf("Current HP = %d, want %d", hp.Current, tt.expectedHP)
			}
		})
	}
}

func TestHitPointsAddTemporaryHP(t *testing.T) {
	hp := HitPoints{Maximum: 20, Current: 20, Temporary: 5}

	// Lower temp HP should not replace
	hp.AddTemporaryHP(3)
	if hp.Temporary != 5 {
		t.Errorf("Temporary HP = %d, want 5 (should not decrease)", hp.Temporary)
	}

	// Higher temp HP should replace
	hp.AddTemporaryHP(10)
	if hp.Temporary != 10 {
		t.Errorf("Temporary HP = %d, want 10", hp.Temporary)
	}
}

func TestHitPointsIsUnconscious(t *testing.T) {
	hp := HitPoints{Maximum: 20, Current: 5}
	if hp.IsUnconscious() {
		t.Error("Should not be unconscious at 5 HP")
	}

	hp.Current = 0
	if !hp.IsUnconscious() {
		t.Error("Should be unconscious at 0 HP")
	}
}

func TestHitDiceUse(t *testing.T) {
	hd := NewHitDice(5, "d8")

	for i := 0; i < 5; i++ {
		if !hd.Use() {
			t.Errorf("Use() should succeed when %d remaining", hd.Remaining+1)
		}
	}

	if hd.Remaining != 0 {
		t.Errorf("Remaining = %d, want 0", hd.Remaining)
	}

	if hd.Use() {
		t.Error("Use() should fail when 0 remaining")
	}
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
			hd := HitDice{Total: tt.total, Remaining: tt.remaining, DieType: "d8"}
			recovered := hd.RecoverOnLongRest()
			if recovered != tt.expectedRecovered {
				t.Errorf("Recovered = %d, want %d", recovered, tt.expectedRecovered)
			}
			if hd.Remaining != tt.expectedRemaining {
				t.Errorf("Remaining = %d, want %d", hd.Remaining, tt.expectedRemaining)
			}
		})
	}
}

func TestDeathSaves(t *testing.T) {
	ds := NewDeathSaves()

	// Test successes
	ds.AddSuccess()
	ds.AddSuccess()
	if ds.IsStabilized() {
		t.Error("Should not be stabilized with 2 successes")
	}

	stabilized := ds.AddSuccess()
	if !stabilized || !ds.IsStabilized() {
		t.Error("Should be stabilized with 3 successes")
	}

	// Reset and test failures
	ds.Reset()
	if ds.Successes != 0 || ds.Failures != 0 {
		t.Error("Reset should clear successes and failures")
	}

	ds.AddFailure()
	ds.AddFailure()
	if ds.IsDead() {
		t.Error("Should not be dead with 2 failures")
	}

	dead := ds.AddFailure()
	if !dead || !ds.IsDead() {
		t.Error("Should be dead with 3 failures")
	}
}

func TestDeathSavesCriticalFailure(t *testing.T) {
	ds := NewDeathSaves()
	ds.AddFailure()
	dead := ds.AddCriticalFailure()
	if !dead {
		t.Error("Critical failure after 1 failure should result in death")
	}
	if ds.Failures != 3 {
		t.Errorf("Failures = %d, want 3", ds.Failures)
	}
}

func TestCombatStatsConditions(t *testing.T) {
	cs := NewCombatStats(20, "d8", 1, 30)

	// Add condition
	cs.AddCondition(ConditionPoisoned)
	if !cs.HasCondition(ConditionPoisoned) {
		t.Error("Should have poisoned condition")
	}

	// Adding same condition shouldn't duplicate
	cs.AddCondition(ConditionPoisoned)
	if len(cs.Conditions) != 1 {
		t.Errorf("Conditions length = %d, want 1", len(cs.Conditions))
	}

	// Add another condition
	cs.AddCondition(ConditionFrightened)
	if len(cs.Conditions) != 2 {
		t.Errorf("Conditions length = %d, want 2", len(cs.Conditions))
	}

	// Remove condition
	cs.RemoveCondition(ConditionPoisoned)
	if cs.HasCondition(ConditionPoisoned) {
		t.Error("Should not have poisoned condition after removal")
	}
	if !cs.HasCondition(ConditionFrightened) {
		t.Error("Should still have frightened condition")
	}

	// Clear all
	cs.ClearConditions()
	if len(cs.Conditions) != 0 {
		t.Error("Should have no conditions after clear")
	}
}

func TestAllConditionsCount(t *testing.T) {
	conditions := AllConditions()
	if len(conditions) != 15 {
		t.Errorf("AllConditions() returned %d conditions, want 15", len(conditions))
	}
}
