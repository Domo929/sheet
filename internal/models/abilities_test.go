package models

import "testing"

func TestAbilityScoreModifier(t *testing.T) {
	tests := []struct {
		name     string
		score    int
		expected int
	}{
		{"score 1", 1, -5},
		{"score 2", 2, -4},
		{"score 3", 3, -4},
		{"score 8", 8, -1},
		{"score 9", 9, -1},
		{"score 10", 10, 0},
		{"score 11", 11, 0},
		{"score 12", 12, 1},
		{"score 13", 13, 1},
		{"score 14", 14, 2},
		{"score 15", 15, 2},
		{"score 16", 16, 3},
		{"score 17", 17, 3},
		{"score 18", 18, 4},
		{"score 19", 19, 4},
		{"score 20", 20, 5},
		{"score 22", 22, 6},
		{"score 24", 24, 7},
		{"score 30", 30, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ability := AbilityScore{Base: tt.score}
			if got := ability.Modifier(); got != tt.expected {
				t.Errorf("AbilityScore{Base: %d}.Modifier() = %d, want %d", tt.score, got, tt.expected)
			}
		})
	}
}

func TestAbilityScoreTotal(t *testing.T) {
	tests := []struct {
		name      string
		base      int
		temporary int
		expected  int
	}{
		{"base only", 14, 0, 14},
		{"with positive temp", 14, 2, 16},
		{"with negative temp", 14, -2, 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ability := AbilityScore{Base: tt.base, Temporary: tt.temporary}
			if got := ability.Total(); got != tt.expected {
				t.Errorf("AbilityScore.Total() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestAbilityScoreModifierWithTemporary(t *testing.T) {
	ability := AbilityScore{Base: 14, Temporary: 2}
	// Total is 16, modifier should be +3
	if got := ability.Modifier(); got != 3 {
		t.Errorf("AbilityScore{14, +2 temp}.Modifier() = %d, want 3", got)
	}
}

func TestNewAbilityScores(t *testing.T) {
	scores := NewAbilityScores()

	// All scores should be 10
	if scores.Strength.Base != 10 {
		t.Errorf("NewAbilityScores().Strength.Base = %d, want 10", scores.Strength.Base)
	}
	if scores.Dexterity.Base != 10 {
		t.Errorf("NewAbilityScores().Dexterity.Base = %d, want 10", scores.Dexterity.Base)
	}
	if scores.Constitution.Base != 10 {
		t.Errorf("NewAbilityScores().Constitution.Base = %d, want 10", scores.Constitution.Base)
	}
	if scores.Intelligence.Base != 10 {
		t.Errorf("NewAbilityScores().Intelligence.Base = %d, want 10", scores.Intelligence.Base)
	}
	if scores.Wisdom.Base != 10 {
		t.Errorf("NewAbilityScores().Wisdom.Base = %d, want 10", scores.Wisdom.Base)
	}
	if scores.Charisma.Base != 10 {
		t.Errorf("NewAbilityScores().Charisma.Base = %d, want 10", scores.Charisma.Base)
	}
}

func TestNewAbilityScoresFromValues(t *testing.T) {
	scores := NewAbilityScoresFromValues(15, 14, 13, 12, 10, 8)

	if scores.Strength.Base != 15 {
		t.Errorf("Strength = %d, want 15", scores.Strength.Base)
	}
	if scores.Dexterity.Base != 14 {
		t.Errorf("Dexterity = %d, want 14", scores.Dexterity.Base)
	}
	if scores.Constitution.Base != 13 {
		t.Errorf("Constitution = %d, want 13", scores.Constitution.Base)
	}
	if scores.Intelligence.Base != 12 {
		t.Errorf("Intelligence = %d, want 12", scores.Intelligence.Base)
	}
	if scores.Wisdom.Base != 10 {
		t.Errorf("Wisdom = %d, want 10", scores.Wisdom.Base)
	}
	if scores.Charisma.Base != 8 {
		t.Errorf("Charisma = %d, want 8", scores.Charisma.Base)
	}
}

func TestAbilityScoresGet(t *testing.T) {
	scores := NewAbilityScoresFromValues(15, 14, 13, 12, 10, 8)

	tests := []struct {
		ability  Ability
		expected int
	}{
		{AbilityStrength, 15},
		{AbilityDexterity, 14},
		{AbilityConstitution, 13},
		{AbilityIntelligence, 12},
		{AbilityWisdom, 10},
		{AbilityCharisma, 8},
	}

	for _, tt := range tests {
		t.Run(string(tt.ability), func(t *testing.T) {
			score := scores.Get(tt.ability)
			if score.Base != tt.expected {
				t.Errorf("Get(%s).Base = %d, want %d", tt.ability, score.Base, tt.expected)
			}
		})
	}
}

func TestAbilityScoresGetInvalid(t *testing.T) {
	scores := NewAbilityScores()
	got := scores.Get(Ability("invalid"))
	// Should return default value (Base: 10)
	if got.Base != 10 {
		t.Errorf("Get(invalid).Base = %d, want 10", got.Base)
	}
}

func TestAbilityScoresGetModifier(t *testing.T) {
	scores := NewAbilityScoresFromValues(16, 14, 12, 10, 8, 6)

	tests := []struct {
		ability  Ability
		expected int
	}{
		{AbilityStrength, 3},     // 16 -> +3
		{AbilityDexterity, 2},    // 14 -> +2
		{AbilityConstitution, 1}, // 12 -> +1
		{AbilityIntelligence, 0}, // 10 -> +0
		{AbilityWisdom, -1},      // 8 -> -1
		{AbilityCharisma, -2},    // 6 -> -2
	}

	for _, tt := range tests {
		t.Run(string(tt.ability), func(t *testing.T) {
			if got := scores.GetModifier(tt.ability); got != tt.expected {
				t.Errorf("GetModifier(%s) = %d, want %d", tt.ability, got, tt.expected)
			}
		})
	}
}

func TestStandardArray(t *testing.T) {
	arr := StandardArray()
	expected := []int{15, 14, 13, 12, 10, 8}

	if len(arr) != len(expected) {
		t.Fatalf("StandardArray() length = %d, want %d", len(arr), len(expected))
	}

	for i, v := range expected {
		if arr[i] != v {
			t.Errorf("StandardArray()[%d] = %d, want %d", i, arr[i], v)
		}
	}
}
