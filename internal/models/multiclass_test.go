package models

import "testing"

func TestMulticlassCasterLevel(t *testing.T) {
	cases := []struct {
		name    string
		classes []ClassLevel
		want    int
	}{
		{"Wizard5/Cleric5 full+full", []ClassLevel{{Class: "Wizard", Level: 5}, {Class: "Cleric", Level: 5}}, 10},
		{"Paladin6/Sorcerer4 half+full", []ClassLevel{{Class: "Paladin", Level: 6}, {Class: "Sorcerer", Level: 4}}, 7},
		{"EK Fighter6/Wizard4 third+full", []ClassLevel{{Class: "Fighter", Subclass: "Eldritch Knight", Level: 6}, {Class: "Wizard", Level: 4}}, 6},
		{"Ranger5/Rogue3 half+none", []ClassLevel{{Class: "Ranger", Level: 5}, {Class: "Rogue", Level: 3}}, 2},
		{"Warlock excluded", []ClassLevel{{Class: "Warlock", Level: 5}, {Class: "Sorcerer", Level: 3}}, 3},
		{"Barbarian none", []ClassLevel{{Class: "Barbarian", Level: 10}}, 0},
		{"plain Fighter is not a caster", []ClassLevel{{Class: "Fighter", Level: 6}}, 0},
	}
	for _, tc := range cases {
		if got := MulticlassCasterLevel(tc.classes); got != tc.want {
			t.Errorf("%s: caster level = %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestMulticlassSpellSlots(t *testing.T) {
	// Caster level 10 -> 4/3/3/3/2.
	got := MulticlassSpellSlots(10)
	want := [9]int{4, 3, 3, 3, 2, 0, 0, 0, 0}
	if got != want {
		t.Errorf("slots@10 = %v, want %v", got, want)
	}
	// Caster level 7 -> 4/3/3/1.
	got = MulticlassSpellSlots(7)
	want = [9]int{4, 3, 3, 1, 0, 0, 0, 0, 0}
	if got != want {
		t.Errorf("slots@7 = %v, want %v", got, want)
	}
	// Level 0 -> none.
	if MulticlassSpellSlots(0) != ([9]int{}) {
		t.Errorf("slots@0 should be empty")
	}
	// Overflow clamps to 20.
	if MulticlassSpellSlots(25) != MulticlassSpellSlots(20) {
		t.Errorf("caster level should clamp to 20")
	}
}

func TestTotalLevelAndSummary(t *testing.T) {
	c := NewCharacter("id", "Hero", "Human", "Fighter")
	c.Info.Level = 3
	// Single-class falls back to Info.
	if c.TotalLevel() != 3 {
		t.Errorf("single-class total level = %d, want 3", c.TotalLevel())
	}
	if c.IsMulticlass() {
		t.Error("should not be multiclass with no Classes")
	}

	c.Classes = []ClassLevel{{Class: "Fighter", Level: 5}, {Class: "Wizard", Level: 3}}
	if c.TotalLevel() != 8 {
		t.Errorf("multiclass total level = %d, want 8", c.TotalLevel())
	}
	if !c.IsMulticlass() {
		t.Error("should be multiclass with two Classes")
	}
	if got := c.ClassSummary(); got != "Fighter 5 / Wizard 3" {
		t.Errorf("ClassSummary = %q", got)
	}
}

func TestSyncPrimaryClass(t *testing.T) {
	c := NewCharacter("id", "Hero", "Human", "Fighter")
	c.Classes = []ClassLevel{{Class: "Wizard", Subclass: "Evoker", Level: 6}, {Class: "Fighter", Level: 2}}
	c.SyncPrimaryClass()
	if c.Info.Class != "Wizard" || c.Info.Subclass != "Evoker" {
		t.Errorf("primary class not synced: %s/%s", c.Info.Class, c.Info.Subclass)
	}
	if c.Info.Level != 8 {
		t.Errorf("Info.Level = %d, want 8", c.Info.Level)
	}
}

func TestApplyMulticlassSpellSlots(t *testing.T) {
	c := NewCharacter("id", "Hero", "Human", "Wizard")
	sc := NewSpellcasting(AbilityIntelligence)
	c.Spellcasting = &sc
	c.Classes = []ClassLevel{{Class: "Wizard", Level: 5}, {Class: "Cleric", Level: 5}}

	c.ApplyMulticlassSpellSlots()

	checks := map[int]int{1: 4, 2: 3, 3: 3, 4: 3, 5: 2, 6: 0}
	for lvl, want := range checks {
		if got := c.Spellcasting.SpellSlots.GetSlot(lvl).Total; got != want {
			t.Errorf("slot L%d total = %d, want %d", lvl, got, want)
		}
	}

	// Expend a level-1 slot, then re-apply; used count should be preserved.
	c.Spellcasting.SpellSlots.UseSlot(1)
	if rem := c.Spellcasting.SpellSlots.GetSlot(1).Remaining; rem != 3 {
		t.Fatalf("after expend, L1 remaining = %d, want 3", rem)
	}
	c.ApplyMulticlassSpellSlots()
	if rem := c.Spellcasting.SpellSlots.GetSlot(1).Remaining; rem != 3 {
		t.Errorf("re-apply should preserve expended slots: L1 remaining = %d, want 3", rem)
	}
}

func TestApplyMulticlassSpellSlotsNoopSingleClass(t *testing.T) {
	c := NewCharacter("id", "Hero", "Human", "Wizard")
	sc := NewSpellcasting(AbilityIntelligence)
	c.Spellcasting = &sc
	// Only one class entry -> not multiclass -> no-op.
	c.Classes = []ClassLevel{{Class: "Wizard", Level: 5}}
	c.ApplyMulticlassSpellSlots()
	if c.Spellcasting.SpellSlots.GetSlot(1).Total != 0 {
		t.Error("single-class apply should be a no-op")
	}
}
