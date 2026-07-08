package models

import (
	"strings"
	"testing"
)

func mkScores(str, dex, con, intel, wis, cha int) AbilityScores {
	return NewAbilityScoresFromValues(str, dex, con, intel, wis, cha)
}

func TestMulticlassPrerequisite_Table(t *testing.T) {
	tests := []struct {
		class      string
		abilities  []Ability
		requireAll bool
	}{
		{"Fighter", []Ability{AbilityStrength, AbilityDexterity}, false},
		{"Monk", []Ability{AbilityDexterity, AbilityWisdom}, true},
		{"Paladin", []Ability{AbilityStrength, AbilityCharisma}, true},
		{"Ranger", []Ability{AbilityDexterity, AbilityWisdom}, true},
		{"Barbarian", []Ability{AbilityStrength}, false},
		{"Wizard", []Ability{AbilityIntelligence}, false},
	}
	for _, tt := range tests {
		req, ok := MulticlassPrerequisite(tt.class)
		if !ok {
			t.Errorf("%s: expected a known prerequisite", tt.class)
			continue
		}
		if req.RequireAll != tt.requireAll {
			t.Errorf("%s: RequireAll = %v, want %v", tt.class, req.RequireAll, tt.requireAll)
		}
		if len(req.Abilities) != len(tt.abilities) {
			t.Errorf("%s: abilities = %v, want %v", tt.class, req.Abilities, tt.abilities)
		}
	}
}

func TestMeetsRequirement_OrVsAnd(t *testing.T) {
	fighter, _ := MulticlassPrerequisite("Fighter") // Str OR Dex
	// Str 13, Dex 8 -> qualifies (OR).
	if ok, _ := meetsRequirement(mkScores(13, 8, 10, 10, 10, 10), fighter); !ok {
		t.Error("Fighter: Str 13 should satisfy OR requirement")
	}
	// Str 8, Dex 8 -> fails.
	if ok, missing := meetsRequirement(mkScores(8, 8, 10, 10, 10, 10), fighter); ok || len(missing) != 2 {
		t.Errorf("Fighter: Str/Dex 8 should fail with 2 options listed, got ok=%v missing=%v", ok, missing)
	}

	monk, _ := MulticlassPrerequisite("Monk") // Dex AND Wis
	// Dex 13, Wis 12 -> fails (needs both).
	if ok, missing := meetsRequirement(mkScores(10, 13, 10, 10, 12, 10), monk); ok || len(missing) != 1 {
		t.Errorf("Monk: Wis 12 should fail AND requirement with 1 missing, got ok=%v missing=%v", ok, missing)
	}
	// Dex 13, Wis 13 -> qualifies.
	if ok, _ := meetsRequirement(mkScores(10, 13, 10, 10, 13, 10), monk); !ok {
		t.Error("Monk: Dex 13 & Wis 13 should satisfy AND requirement")
	}
}

func TestCanMulticlassInto_ChecksBothClasses(t *testing.T) {
	// A Wizard (Int 15) with low Str/Dex cannot multiclass into Fighter.
	c := NewCharacter("id", "Test", "Human", "Wizard")
	c.AbilityScores = mkScores(9, 10, 12, 15, 11, 8)
	c.Info.Level = 5
	if ok, reason := c.CanMulticlassInto("Fighter"); ok {
		t.Errorf("expected Fighter multiclass to be blocked, reason=%q", reason)
	} else if !strings.Contains(reason, "Fighter") {
		t.Errorf("reason should mention Fighter, got %q", reason)
	}

	// Give them Dex 13: now Fighter qualifies (and Wizard Int already 13+).
	c.AbilityScores.SetBase(AbilityDexterity, 13)
	if ok, reason := c.CanMulticlassInto("Fighter"); !ok {
		t.Errorf("expected Fighter multiclass to be allowed, got reason=%q", reason)
	}
}

func TestCanMulticlassInto_CurrentClassPrereq(t *testing.T) {
	// A Fighter who qualified via Strength (Dex low, Int high) tries to add
	// Wizard. Wizard needs Int 13 (met), and the current Fighter still needs
	// Str or Dex 13 (Str 15 met) -> allowed.
	c := NewCharacter("id", "Test", "Human", "Fighter")
	c.AbilityScores = mkScores(15, 8, 12, 14, 10, 8)
	c.Info.Level = 4
	if ok, reason := c.CanMulticlassInto("Wizard"); !ok {
		t.Errorf("expected Wizard multiclass to be allowed, got reason=%q", reason)
	}

	// Drop Int below 13: Wizard now blocked by the new-class requirement.
	c.AbilityScores.SetBase(AbilityIntelligence, 11)
	if ok, _ := c.CanMulticlassInto("Wizard"); ok {
		t.Error("expected Wizard multiclass to be blocked when Int < 13")
	}
}

func TestGrantMulticlassProficiencies(t *testing.T) {
	c := NewCharacter("id", "Test", "Human", "Wizard")
	before := len(c.Proficiencies.Armor) + len(c.Proficiencies.Weapons)

	grant := c.GrantMulticlassProficiencies("Fighter")
	if grant.IsEmpty() {
		t.Fatal("Fighter grant should not be empty")
	}
	if !c.Proficiencies.HasArmor("Light Armor") || !c.Proficiencies.HasArmor("Shields") {
		t.Error("Fighter multiclass should grant Light Armor and Shields")
	}
	if !c.Proficiencies.HasWeapon("Martial Weapons") {
		t.Error("Fighter multiclass should grant Martial Weapons")
	}
	if len(c.Proficiencies.Armor)+len(c.Proficiencies.Weapons) <= before {
		t.Error("expected proficiencies to increase")
	}

	// Idempotent: granting again should not duplicate.
	n := len(c.Proficiencies.Armor)
	c.GrantMulticlassProficiencies("Fighter")
	if len(c.Proficiencies.Armor) != n {
		t.Error("re-granting should not duplicate armor proficiencies")
	}
}

func TestGrantMulticlassProficiencies_NoneClasses(t *testing.T) {
	c := NewCharacter("id", "Test", "Human", "Fighter")
	// Sorcerer and Wizard grant nothing on multiclass.
	if g := c.GrantMulticlassProficiencies("Sorcerer"); !g.IsEmpty() {
		t.Errorf("Sorcerer multiclass grant should be empty, got %+v", g)
	}
	if g := c.GrantMulticlassProficiencies("Wizard"); !g.IsEmpty() {
		t.Errorf("Wizard multiclass grant should be empty, got %+v", g)
	}
}
