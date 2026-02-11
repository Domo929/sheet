package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			assert.Equal(t, tt.expected, ability.Modifier(), "AbilityScore{Base: %d}.Modifier()", tt.score)
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
			assert.Equal(t, tt.expected, ability.Total())
		})
	}
}

func TestAbilityScoreModifierWithTemporary(t *testing.T) {
	ability := AbilityScore{Base: 14, Temporary: 2}
	// Total is 16, modifier should be +3
	assert.Equal(t, 3, ability.Modifier(), "AbilityScore{14, +2 temp}.Modifier()")
}

func TestNewAbilityScores(t *testing.T) {
	scores := NewAbilityScores()

	// All scores should be 10
	assert.Equal(t, 10, scores.Strength.Base, "Strength.Base")
	assert.Equal(t, 10, scores.Dexterity.Base, "Dexterity.Base")
	assert.Equal(t, 10, scores.Constitution.Base, "Constitution.Base")
	assert.Equal(t, 10, scores.Intelligence.Base, "Intelligence.Base")
	assert.Equal(t, 10, scores.Wisdom.Base, "Wisdom.Base")
	assert.Equal(t, 10, scores.Charisma.Base, "Charisma.Base")
}

func TestNewAbilityScoresFromValues(t *testing.T) {
	scores := NewAbilityScoresFromValues(15, 14, 13, 12, 10, 8)

	assert.Equal(t, 15, scores.Strength.Base, "Strength")
	assert.Equal(t, 14, scores.Dexterity.Base, "Dexterity")
	assert.Equal(t, 13, scores.Constitution.Base, "Constitution")
	assert.Equal(t, 12, scores.Intelligence.Base, "Intelligence")
	assert.Equal(t, 10, scores.Wisdom.Base, "Wisdom")
	assert.Equal(t, 8, scores.Charisma.Base, "Charisma")
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
			assert.Equal(t, tt.expected, score.Base, "Get(%s).Base", tt.ability)
		})
	}
}

func TestAbilityScoresGetInvalid(t *testing.T) {
	scores := NewAbilityScores()
	got := scores.Get(Ability("invalid"))
	// Should return default value (Base: 10)
	assert.Equal(t, 10, got.Base, "Get(invalid).Base")
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
			assert.Equal(t, tt.expected, scores.GetModifier(tt.ability), "GetModifier(%s)", tt.ability)
		})
	}
}

func TestStandardArray(t *testing.T) {
	arr := StandardArray()
	expected := []int{15, 14, 13, 12, 10, 8}

	require.Len(t, arr, len(expected), "StandardArray() length")
	assert.Equal(t, expected, arr)
}

func TestPointBuyCost(t *testing.T) {
	tests := []struct {
		score    int
		expected int
	}{
		{8, 0},
		{9, 1},
		{10, 2},
		{11, 3},
		{12, 4},
		{13, 5},
		{14, 7},
		{15, 9},
		{7, 0},   // Below minimum
		{16, 11}, // Above maximum (theoretical)
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.score)), func(t *testing.T) {
			assert.Equal(t, tt.expected, PointBuyCost(tt.score), "PointBuyCost(%d)", tt.score)
		})
	}
}

func TestValidatePointBuy(t *testing.T) {
	tests := []struct {
		name          string
		scores        AbilityScores
		expectedPts   int
		expectedValid bool
	}{
		{
			name:          "standard valid allocation (27 points)",
			scores:        NewAbilityScoresFromValues(15, 14, 13, 12, 10, 8),
			expectedPts:   27,
			expectedValid: true,
		},
		{
			name:          "all 13s (30 points, invalid)",
			scores:        NewAbilityScoresFromValues(13, 13, 13, 13, 13, 13),
			expectedPts:   30,
			expectedValid: false,
		},
		{
			name:          "balanced allocation (24 points)",
			scores:        NewAbilityScoresFromValues(13, 13, 13, 12, 12, 12),
			expectedPts:   27,
			expectedValid: true,
		},
		{
			name:          "all 10s (12 points)",
			scores:        NewAbilityScoresFromValues(10, 10, 10, 10, 10, 10),
			expectedPts:   12,
			expectedValid: true,
		},
		{
			name:          "invalid - score too low",
			scores:        NewAbilityScoresFromValues(7, 13, 13, 13, 13, 13),
			expectedPts:   0,
			expectedValid: false,
		},
		{
			name:          "invalid - score too high",
			scores:        NewAbilityScoresFromValues(16, 13, 13, 13, 13, 13),
			expectedPts:   0,
			expectedValid: false,
		},
		{
			name:          "min-max allocation",
			scores:        NewAbilityScoresFromValues(15, 15, 15, 8, 8, 8),
			expectedPts:   27,
			expectedValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPts, gotValid := ValidatePointBuy(tt.scores)
			assert.Equal(t, tt.expectedPts, gotPts, "ValidatePointBuy() points")
			assert.Equal(t, tt.expectedValid, gotValid, "ValidatePointBuy() valid")
		})
	}
}

func TestCalculatePointBuyTotal(t *testing.T) {
	scores := NewAbilityScoresFromValues(15, 14, 13, 12, 10, 8)
	total := CalculatePointBuyTotal(scores)
	assert.Equal(t, 27, total)
}
