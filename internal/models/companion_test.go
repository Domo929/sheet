package models

import "testing"

func TestAbilityModifier(t *testing.T) {
	cases := map[int]int{
		1:  -5,
		7:  -2,
		8:  -1,
		9:  -1,
		10: 0,
		11: 0,
		12: 1,
		15: 2,
		18: 4,
		20: 5,
	}
	for score, want := range cases {
		if got := AbilityModifier(score); got != want {
			t.Errorf("AbilityModifier(%d) = %d, want %d", score, got, want)
		}
	}
}

func TestFormatModifierModels(t *testing.T) {
	if got := FormatModifier(3); got != "+3" {
		t.Errorf("FormatModifier(3) = %q, want +3", got)
	}
	if got := FormatModifier(-1); got != "-1" {
		t.Errorf("FormatModifier(-1) = %q, want -1", got)
	}
	if got := FormatModifier(0); got != "+0" {
		t.Errorf("FormatModifier(0) = %q, want +0", got)
	}
}

func TestCompanionDamageClampsAtZero(t *testing.T) {
	c := Companion{MaxHP: 10, CurrentHP: 10}
	c.Damage(4)
	if c.CurrentHP != 6 {
		t.Fatalf("after 4 damage want 6, got %d", c.CurrentHP)
	}
	c.Damage(100)
	if c.CurrentHP != 0 {
		t.Fatalf("overkill should clamp to 0, got %d", c.CurrentHP)
	}
}

func TestCompanionTempHPAbsorbsDamage(t *testing.T) {
	c := Companion{MaxHP: 10, CurrentHP: 10, TempHP: 5}
	c.Damage(3)
	if c.TempHP != 2 || c.CurrentHP != 10 {
		t.Fatalf("temp HP should absorb first: temp=%d hp=%d", c.TempHP, c.CurrentHP)
	}
	c.Damage(4)
	if c.TempHP != 0 || c.CurrentHP != 8 {
		t.Fatalf("overflow should carry into HP: temp=%d hp=%d", c.TempHP, c.CurrentHP)
	}
}

func TestCompanionHealCapsAtMax(t *testing.T) {
	c := Companion{MaxHP: 10, CurrentHP: 3}
	c.Heal(4)
	if c.CurrentHP != 7 {
		t.Fatalf("heal 4 want 7, got %d", c.CurrentHP)
	}
	c.Heal(100)
	if c.CurrentHP != 10 {
		t.Fatalf("heal should cap at max, got %d", c.CurrentHP)
	}
}

func TestAddCompanionInitializesHP(t *testing.T) {
	ch := NewCharacter("id", "Hero", "Human", "Ranger")
	ch.AddCompanion(Companion{ID: "c1", Name: "Wolf", MaxHP: 11})
	if len(ch.Companions) != 1 {
		t.Fatalf("want 1 companion, got %d", len(ch.Companions))
	}
	if ch.Companions[0].CurrentHP != 11 {
		t.Errorf("current HP should default to max, got %d", ch.Companions[0].CurrentHP)
	}
}

func TestRemoveCompanion(t *testing.T) {
	ch := NewCharacter("id", "Hero", "Human", "Ranger")
	ch.AddCompanion(Companion{ID: "c1", Name: "Wolf", MaxHP: 11})
	ch.AddCompanion(Companion{ID: "c2", Name: "Bear", MaxHP: 34})

	if !ch.RemoveCompanion("c1") {
		t.Fatal("RemoveCompanion should report success")
	}
	if len(ch.Companions) != 1 || ch.Companions[0].ID != "c2" {
		t.Fatalf("wrong companion removed: %+v", ch.Companions)
	}
	if ch.RemoveCompanion("nope") {
		t.Error("removing a missing companion should return false")
	}
}
