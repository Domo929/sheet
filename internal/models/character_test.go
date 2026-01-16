package models

import (
	"encoding/json"
	"testing"
)

func TestNewCharacter(t *testing.T) {
	c := NewCharacter("char-1", "Gandalf", "Human", "Wizard")

	if c.ID != "char-1" {
		t.Errorf("ID = %s, want char-1", c.ID)
	}
	if c.Info.Name != "Gandalf" {
		t.Errorf("Name = %s, want Gandalf", c.Info.Name)
	}
	if c.Info.Race != "Human" {
		t.Errorf("Race = %s, want Human", c.Info.Race)
	}
	if c.Info.Class != "Wizard" {
		t.Errorf("Class = %s, want Wizard", c.Info.Class)
	}
	if c.Info.Level != 1 {
		t.Errorf("Level = %d, want 1", c.Info.Level)
	}
	if c.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestCharacterGetSkillModifier(t *testing.T) {
	c := NewCharacter("char-1", "Test", "Human", "Rogue")
	c.AbilityScores = NewAbilityScoresFromValues(10, 16, 14, 12, 10, 14) // DEX 16 = +3
	c.Skills.SetProficiency(SkillStealth, Expertise)

	// Stealth with expertise: +3 (DEX) + 4 (2x prof at level 1) = +7
	mod := c.GetSkillModifier(SkillStealth)
	if mod != 7 {
		t.Errorf("Stealth modifier = %d, want 7 (+3 DEX + 4 expertise)", mod)
	}

	// Athletics without proficiency: +0 (STR)
	mod = c.GetSkillModifier(SkillAthletics)
	if mod != 0 {
		t.Errorf("Athletics modifier = %d, want 0", mod)
	}
}

func TestCharacterGetSavingThrowModifier(t *testing.T) {
	c := NewCharacter("char-1", "Test", "Human", "Fighter")
	c.AbilityScores = NewAbilityScoresFromValues(16, 14, 14, 10, 12, 8)
	c.SavingThrows.SetProficiency(AbilityStrength, true)
	c.SavingThrows.SetProficiency(AbilityConstitution, true)

	// STR save proficient: +3 (STR) + 2 (prof) = +5
	mod := c.GetSavingThrowModifier(AbilityStrength)
	if mod != 5 {
		t.Errorf("STR save = %d, want 5", mod)
	}

	// WIS save not proficient: +1 (WIS) = +1
	mod = c.GetSavingThrowModifier(AbilityWisdom)
	if mod != 1 {
		t.Errorf("WIS save = %d, want 1", mod)
	}
}

func TestCharacterSpellcasting(t *testing.T) {
	c := NewCharacter("char-1", "Test", "Elf", "Wizard")
	c.AbilityScores = NewAbilityScoresFromValues(8, 14, 12, 18, 12, 10) // INT 18 = +4
	c.Info.Level = 5                                                    // Proficiency +3

	// Set up spellcasting
	sc := NewSpellcasting(AbilityIntelligence)
	c.Spellcasting = &sc

	// Spell save DC: 8 + 4 (INT) + 3 (prof) = 15
	dc := c.GetSpellSaveDC()
	if dc != 15 {
		t.Errorf("Spell save DC = %d, want 15", dc)
	}

	// Spell attack bonus: 4 (INT) + 3 (prof) = +7
	atk := c.GetSpellAttackBonus()
	if atk != 7 {
		t.Errorf("Spell attack = %d, want 7", atk)
	}
}

func TestCharacterSpellcastingNil(t *testing.T) {
	c := NewCharacter("char-1", "Test", "Human", "Fighter")

	if c.GetSpellSaveDC() != 0 {
		t.Error("Non-caster should have 0 spell save DC")
	}
	if c.GetSpellAttackBonus() != 0 {
		t.Error("Non-caster should have 0 spell attack")
	}
	if c.IsSpellcaster() {
		t.Error("Fighter should not be a spellcaster")
	}
}

func TestCharacterDamageAndHealing(t *testing.T) {
	c := NewCharacter("char-1", "Test", "Human", "Fighter")
	c.CombatStats.HitPoints = NewHitPoints(20)

	c.TakeDamage(8)
	if c.CombatStats.HitPoints.Current != 12 {
		t.Errorf("HP after damage = %d, want 12", c.CombatStats.HitPoints.Current)
	}

	c.Heal(5)
	if c.CombatStats.HitPoints.Current != 17 {
		t.Errorf("HP after heal = %d, want 17", c.CombatStats.HitPoints.Current)
	}
}

func TestCharacterLongRest(t *testing.T) {
	c := NewCharacter("char-1", "Test", "Elf", "Wizard")
	c.CombatStats.HitPoints = HitPoints{Maximum: 30, Current: 10, Temporary: 5}
	c.CombatStats.HitDice = HitDice{Total: 5, Remaining: 0, DieType: 6}
	c.CombatStats.ExhaustionLevel = 2
	c.CombatStats.DeathSaves = DeathSaves{Successes: 2, Failures: 1}

	// Set up spellcasting with some used slots
	sc := NewSpellcasting(AbilityIntelligence)
	sc.SpellSlots.SetSlots(1, 4)
	sc.SpellSlots.SetSlots(2, 2)
	sc.SpellSlots.UseSlot(1)
	sc.SpellSlots.UseSlot(1)
	c.Spellcasting = &sc

	c.LongRest()

	// HP should be restored
	if c.CombatStats.HitPoints.Current != 30 {
		t.Errorf("HP after rest = %d, want 30", c.CombatStats.HitPoints.Current)
	}
	if c.CombatStats.HitPoints.Temporary != 0 {
		t.Errorf("Temp HP after rest = %d, want 0", c.CombatStats.HitPoints.Temporary)
	}

	// Hit dice should recover (half of 5 = 2)
	if c.CombatStats.HitDice.Remaining != 2 {
		t.Errorf("Hit dice after rest = %d, want 2", c.CombatStats.HitDice.Remaining)
	}

	// Exhaustion should decrease by 1
	if c.CombatStats.ExhaustionLevel != 1 {
		t.Errorf("Exhaustion after rest = %d, want 1", c.CombatStats.ExhaustionLevel)
	}

	// Death saves should reset
	if c.CombatStats.DeathSaves.Successes != 0 || c.CombatStats.DeathSaves.Failures != 0 {
		t.Error("Death saves should reset after long rest")
	}

	// Spell slots should be restored
	if c.Spellcasting.SpellSlots.Level1.Remaining != 4 {
		t.Errorf("Level 1 slots = %d, want 4", c.Spellcasting.SpellSlots.Level1.Remaining)
	}
}

func TestCharacterShortRest(t *testing.T) {
	c := NewCharacter("char-1", "Test", "Tiefling", "Warlock")

	// Set up warlock pact magic
	pm := NewPactMagic(3, 2)
	pm.Use()
	sc := NewSpellcasting(AbilityCharisma)
	sc.PactMagic = &pm
	c.Spellcasting = &sc

	c.ShortRest()

	// Pact magic should be restored
	if c.Spellcasting.PactMagic.Remaining != 2 {
		t.Errorf("Pact slots after short rest = %d, want 2", c.Spellcasting.PactMagic.Remaining)
	}
}

func TestCharacterJSONRoundTrip(t *testing.T) {
	c := NewCharacter("char-1", "Gandalf", "Human", "Wizard")
	c.AbilityScores = NewAbilityScoresFromValues(10, 14, 12, 18, 16, 12)
	c.Skills.SetProficiency(SkillArcana, Proficient)
	c.SavingThrows.SetProficiency(AbilityIntelligence, true)
	c.Info.Level = 5

	// Serialize
	data, err := c.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}

	// Deserialize
	c2, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}

	// Verify
	if c2.Info.Name != "Gandalf" {
		t.Errorf("Name after round trip = %s, want Gandalf", c2.Info.Name)
	}
	if c2.AbilityScores.Intelligence.Base != 18 {
		t.Errorf("INT after round trip = %d, want 18", c2.AbilityScores.Intelligence.Base)
	}
	if c2.Skills.Arcana.Proficiency != Proficient {
		t.Error("Arcana proficiency should be preserved")
	}
}

func TestCharacterValidate(t *testing.T) {
	// Valid character
	c := NewCharacter("char-1", "Test", "Human", "Fighter")
	errors := c.Validate()
	if len(errors) != 0 {
		t.Errorf("Valid character has errors: %v", errors)
	}

	// Invalid: missing ID
	c2 := NewCharacter("", "Test", "Human", "Fighter")
	errors = c2.Validate()
	if len(errors) == 0 {
		t.Error("Character without ID should have validation error")
	}

	// Invalid: missing name
	c3 := NewCharacter("id-1", "", "Human", "Fighter")
	errors = c3.Validate()
	if len(errors) == 0 {
		t.Error("Character without name should have validation error")
	}

	// Invalid: level out of range
	c4 := NewCharacter("id-1", "Test", "Human", "Fighter")
	c4.Info.Level = 25
	errors = c4.Validate()
	if len(errors) == 0 {
		t.Error("Character with level 25 should have validation error")
	}
}

func TestCharacterGetInitiative(t *testing.T) {
	c := NewCharacter("char-1", "Test", "Human", "Fighter")
	c.AbilityScores.Dexterity = AbilityScore{Base: 16} // +3 modifier

	if c.GetInitiative() != 3 {
		t.Errorf("Initiative = %d, want 3", c.GetInitiative())
	}
}

func TestCharacterJSONStructure(t *testing.T) {
	c := NewCharacter("char-1", "Test", "Human", "Fighter")

	data, _ := c.ToJSON()
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	// Check expected top-level keys exist
	expectedKeys := []string{"id", "createdAt", "updatedAt", "info", "abilityScores",
		"skills", "savingThrows", "combatStats", "inventory", "features", "proficiencies", "personality"}

	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("Missing expected key in JSON: %s", key)
		}
	}
}
