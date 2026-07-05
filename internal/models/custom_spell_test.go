package models

import "testing"

func TestAddCustomSpellLeveled(t *testing.T) {
	sc := NewSpellcasting(AbilityIntelligence)
	sc.AddCustomSpell(CustomSpell{Name: "Arcane Zap", Level: 2, Ritual: true})

	if !sc.IsCustomSpell("Arcane Zap") {
		t.Fatal("expected Arcane Zap to be a custom spell")
	}
	if len(sc.KnownSpells) != 1 {
		t.Fatalf("expected 1 known spell, got %d", len(sc.KnownSpells))
	}
	ks := sc.KnownSpells[0]
	if !ks.Custom || ks.Level != 2 || !ks.Ritual {
		t.Errorf("known spell not marked correctly: %+v", ks)
	}
	if cs := sc.FindCustomSpell("Arcane Zap"); cs == nil || cs.Level != 2 {
		t.Errorf("FindCustomSpell returned %+v", cs)
	}
}

func TestAddCustomSpellCantrip(t *testing.T) {
	sc := NewSpellcasting(AbilityCharisma)
	sc.AddCustomSpell(CustomSpell{Name: "Spark", Level: 0})

	if len(sc.CantripsKnown) != 1 || sc.CantripsKnown[0] != "Spark" {
		t.Fatalf("cantrip not added to CantripsKnown: %v", sc.CantripsKnown)
	}
	if len(sc.KnownSpells) != 0 {
		t.Errorf("cantrip should not be in KnownSpells")
	}
	if !sc.IsCustomSpell("Spark") {
		t.Errorf("Spark should be a custom spell")
	}
}

func TestAddCustomSpellReplacesDetail(t *testing.T) {
	sc := NewSpellcasting(AbilityWisdom)
	sc.AddCustomSpell(CustomSpell{Name: "Bolt", Level: 1, Damage: "1d6"})
	sc.AddCustomSpell(CustomSpell{Name: "Bolt", Level: 1, Damage: "3d6"})

	if len(sc.CustomSpells) != 1 {
		t.Fatalf("expected detail replaced, got %d custom spells", len(sc.CustomSpells))
	}
	if sc.CustomSpells[0].Damage != "3d6" {
		t.Errorf("detail not replaced: %s", sc.CustomSpells[0].Damage)
	}
	// Known spell should not be duplicated either.
	count := 0
	for _, k := range sc.KnownSpells {
		if k.Name == "Bolt" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 known Bolt, got %d", count)
	}
}

func TestRemoveCustomSpell(t *testing.T) {
	sc := NewSpellcasting(AbilityIntelligence)
	sc.AddCustomSpell(CustomSpell{Name: "Zap", Level: 1})
	sc.RemoveCustomSpell("Zap")
	if sc.IsCustomSpell("Zap") {
		t.Error("Zap detail should have been removed")
	}
}
