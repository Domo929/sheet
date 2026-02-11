package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProficiencyBonusByLevel(t *testing.T) {
	tests := []struct {
		level    int
		expected int
	}{
		{1, 2}, {2, 2}, {3, 2}, {4, 2},
		{5, 3}, {6, 3}, {7, 3}, {8, 3},
		{9, 4}, {10, 4}, {11, 4}, {12, 4},
		{13, 5}, {14, 5}, {15, 5}, {16, 5},
		{17, 6}, {18, 6}, {19, 6}, {20, 6},
		{0, 2},  // Edge case: level 0 should return +2
		{25, 6}, // Edge case: over 20 should cap at +6
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, ProficiencyBonusByLevel(tt.level), "ProficiencyBonusByLevel(%d)", tt.level)
	}
}

func TestXPForNextLevel(t *testing.T) {
	tests := []struct {
		level    int
		expected int
	}{
		{1, 300},     // Need 300 XP for level 2
		{4, 6500},    // Need 6500 XP for level 5
		{10, 85000},  // Need 85000 XP for level 11
		{19, 355000}, // Need 355000 XP for level 20
		{20, 0},      // Max level, no more XP needed
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, XPForNextLevel(tt.level), "XPForNextLevel(%d)", tt.level)
	}
}

func TestCharacterInfoCanLevelUp(t *testing.T) {
	ci := NewCharacterInfo("Test", "Human", "Fighter")

	// Level 1 with 0 XP cannot level up
	assert.False(t, ci.CanLevelUp(), "Should not be able to level up with 0 XP")

	// Add enough XP for level 2
	ci.AddXP(300)
	assert.True(t, ci.CanLevelUp(), "Should be able to level up with 300 XP")

	// Level up
	ci.LevelUp()
	assert.Equal(t, 2, ci.Level)

	// Should not be able to level up again (need 900 total for level 3)
	assert.False(t, ci.CanLevelUp(), "Should not be able to level up with 300 XP at level 2")
}

func TestCharacterInfoMilestoneProgression(t *testing.T) {
	ci := NewCharacterInfo("Test", "Human", "Fighter")
	ci.ProgressionType = ProgressionMilestone

	// Milestone characters can't level up via XP check
	ci.AddXP(10000)
	assert.False(t, ci.CanLevelUp(), "Milestone characters should not use XP-based level up check")

	// But they can still level up manually
	assert.True(t, ci.LevelUp(), "LevelUp() should succeed")
	assert.Equal(t, 2, ci.Level)
}

func TestCharacterInfoMaxLevel(t *testing.T) {
	ci := NewCharacterInfo("Test", "Human", "Fighter")
	ci.Level = 20
	ci.ExperiencePoints = 500000

	assert.False(t, ci.CanLevelUp(), "Level 20 character should not be able to level up")
	assert.False(t, ci.LevelUp(), "LevelUp() should fail at level 20")
}

func TestCharacterInfoProficiencyBonus(t *testing.T) {
	ci := NewCharacterInfo("Test", "Human", "Fighter")

	assert.Equal(t, 2, ci.ProficiencyBonus(), "Level 1 proficiency bonus")

	ci.Level = 9
	assert.Equal(t, 4, ci.ProficiencyBonus(), "Level 9 proficiency bonus")
}

func TestPersonality(t *testing.T) {
	p := NewPersonality()

	p.AddTrait("I am always polite.")
	p.AddTrait("I love a good mystery.")
	p.AddIdeal("Knowledge is power.")
	p.AddBond("I seek to protect my hometown.")
	p.AddFlaw("I can't resist a pretty face.")

	assert.Len(t, p.Traits, 2, "Traits count")
	assert.Len(t, p.Ideals, 1, "Ideals count")
	assert.Len(t, p.Bonds, 1, "Bonds count")
	assert.Len(t, p.Flaws, 1, "Flaws count")
}

func TestNewCharacterInfo(t *testing.T) {
	ci := NewCharacterInfo("Gandalf", "Human", "Wizard")

	assert.Equal(t, "Gandalf", ci.Name)
	assert.Equal(t, "Human", ci.Race)
	assert.Equal(t, "Wizard", ci.Class)
	assert.Equal(t, 1, ci.Level)
	assert.Equal(t, ProgressionXP, ci.ProgressionType)
}
