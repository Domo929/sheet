package models

import "testing"

func TestPreparedSpellCountFullCaster(t *testing.T) {
	prepares, n := PreparedSpellCount("Wizard", 5, 3) // 3 + 5
	if !prepares || n != 8 {
		t.Fatalf("wizard L5 +3: prepares=%v n=%d", prepares, n)
	}
}

func TestPreparedSpellCountHalfCaster(t *testing.T) {
	prepares, n := PreparedSpellCount("Paladin", 5, 3) // 3 + 2
	if !prepares || n != 5 {
		t.Fatalf("paladin L5 +3: prepares=%v n=%d", prepares, n)
	}
}

func TestPreparedSpellCountKnownCaster(t *testing.T) {
	if prepares, n := PreparedSpellCount("Sorcerer", 10, 5); prepares || n != 0 {
		t.Fatalf("sorcerer: prepares=%v n=%d", prepares, n)
	}
}

func TestPreparedSpellCountMinimumOne(t *testing.T) {
	if _, n := PreparedSpellCount("Cleric", 1, -1); n != 1 { // max(1, 0)
		t.Fatalf("expected min 1, got %d", n)
	}
}

func TestRecomputePreparedLimit(t *testing.T) {
	c := NewCharacter("id", "Cler", "Human", "Cleric")
	c.Info.Level = 3
	c.AbilityScores.SetBase(AbilityWisdom, 16) // +3
	sc := NewSpellcasting(AbilityWisdom)
	c.Spellcasting = &sc
	c.RecomputePreparedLimit()
	if !c.Spellcasting.PreparesSpells || c.Spellcasting.MaxPrepared != 6 { // 3 + 3
		t.Fatalf("prepares=%v max=%d", c.Spellcasting.PreparesSpells, c.Spellcasting.MaxPrepared)
	}
}

func TestRecomputePreparedLimitKnownCaster(t *testing.T) {
	c := NewCharacter("id", "Sorc", "Human", "Sorcerer")
	sc := NewSpellcasting(AbilityCharisma)
	c.Spellcasting = &sc
	c.RecomputePreparedLimit()
	if c.Spellcasting.PreparesSpells {
		t.Fatal("sorcerer should not be gated as a prepared caster")
	}
}
