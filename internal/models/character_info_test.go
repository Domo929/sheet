package models

import "testing"

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
		got := ProficiencyBonusByLevel(tt.level)
		if got != tt.expected {
			t.Errorf("ProficiencyBonusByLevel(%d) = %d, want %d", tt.level, got, tt.expected)
		}
	}
}

func TestXPForNextLevel(t *testing.T) {
	tests := []struct {
		level    int
		expected int
	}{
		{1, 300},    // Need 300 XP for level 2
		{4, 6500},   // Need 6500 XP for level 5
		{10, 85000}, // Need 85000 XP for level 11
		{19, 355000}, // Need 355000 XP for level 20
		{20, 0},     // Max level, no more XP needed
	}

	for _, tt := range tests {
		got := XPForNextLevel(tt.level)
		if got != tt.expected {
			t.Errorf("XPForNextLevel(%d) = %d, want %d", tt.level, got, tt.expected)
		}
	}
}

func TestCharacterInfoCanLevelUp(t *testing.T) {
	ci := NewCharacterInfo("Test", "Human", "Fighter")

	// Level 1 with 0 XP cannot level up
	if ci.CanLevelUp() {
		t.Error("Should not be able to level up with 0 XP")
	}

	// Add enough XP for level 2
	ci.AddXP(300)
	if !ci.CanLevelUp() {
		t.Error("Should be able to level up with 300 XP")
	}

	// Level up
	ci.LevelUp()
	if ci.Level != 2 {
		t.Errorf("Level = %d, want 2", ci.Level)
	}

	// Should not be able to level up again (need 900 total for level 3)
	if ci.CanLevelUp() {
		t.Error("Should not be able to level up with 300 XP at level 2")
	}
}

func TestCharacterInfoMilestoneProgression(t *testing.T) {
	ci := NewCharacterInfo("Test", "Human", "Fighter")
	ci.ProgressionType = ProgressionMilestone

	// Milestone characters can't level up via XP check
	ci.AddXP(10000)
	if ci.CanLevelUp() {
		t.Error("Milestone characters should not use XP-based level up check")
	}

	// But they can still level up manually
	if !ci.LevelUp() {
		t.Error("LevelUp() should succeed")
	}
	if ci.Level != 2 {
		t.Errorf("Level = %d, want 2", ci.Level)
	}
}

func TestCharacterInfoMaxLevel(t *testing.T) {
	ci := NewCharacterInfo("Test", "Human", "Fighter")
	ci.Level = 20
	ci.ExperiencePoints = 500000

	if ci.CanLevelUp() {
		t.Error("Level 20 character should not be able to level up")
	}

	if ci.LevelUp() {
		t.Error("LevelUp() should fail at level 20")
	}
}

func TestCharacterInfoProficiencyBonus(t *testing.T) {
	ci := NewCharacterInfo("Test", "Human", "Fighter")

	if ci.ProficiencyBonus() != 2 {
		t.Errorf("Level 1 proficiency bonus = %d, want 2", ci.ProficiencyBonus())
	}

	ci.Level = 9
	if ci.ProficiencyBonus() != 4 {
		t.Errorf("Level 9 proficiency bonus = %d, want 4", ci.ProficiencyBonus())
	}
}

func TestPersonality(t *testing.T) {
	p := NewPersonality()

	p.AddTrait("I am always polite.")
	p.AddTrait("I love a good mystery.")
	p.AddIdeal("Knowledge is power.")
	p.AddBond("I seek to protect my hometown.")
	p.AddFlaw("I can't resist a pretty face.")

	if len(p.Traits) != 2 {
		t.Errorf("Traits count = %d, want 2", len(p.Traits))
	}
	if len(p.Ideals) != 1 {
		t.Errorf("Ideals count = %d, want 1", len(p.Ideals))
	}
	if len(p.Bonds) != 1 {
		t.Errorf("Bonds count = %d, want 1", len(p.Bonds))
	}
	if len(p.Flaws) != 1 {
		t.Errorf("Flaws count = %d, want 1", len(p.Flaws))
	}
}

func TestNewCharacterInfo(t *testing.T) {
	ci := NewCharacterInfo("Gandalf", "Human", "Wizard")

	if ci.Name != "Gandalf" {
		t.Errorf("Name = %s, want Gandalf", ci.Name)
	}
	if ci.Race != "Human" {
		t.Errorf("Race = %s, want Human", ci.Race)
	}
	if ci.Class != "Wizard" {
		t.Errorf("Class = %s, want Wizard", ci.Class)
	}
	if ci.Level != 1 {
		t.Errorf("Level = %d, want 1", ci.Level)
	}
	if ci.ProgressionType != ProgressionXP {
		t.Errorf("ProgressionType = %s, want xp", ci.ProgressionType)
	}
}
