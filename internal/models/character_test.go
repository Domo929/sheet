package models

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCharacter(t *testing.T) {
	c := NewCharacter("char-1", "Gandalf", "Human", "Wizard")

	assert.Equal(t, "char-1", c.ID)
	assert.Equal(t, "Gandalf", c.Info.Name)
	assert.Equal(t, "Human", c.Info.Race)
	assert.Equal(t, "Wizard", c.Info.Class)
	assert.Equal(t, 1, c.Info.Level)
	assert.False(t, c.CreatedAt.IsZero(), "CreatedAt should be set")
}

func TestCharacterGetSkillModifier(t *testing.T) {
	c := NewCharacter("char-1", "Test", "Human", "Rogue")
	c.AbilityScores = NewAbilityScoresFromValues(10, 16, 14, 12, 10, 14) // DEX 16 = +3
	c.Skills.SetProficiency(SkillStealth, Expertise)

	// Stealth with expertise: +3 (DEX) + 4 (2x prof at level 1) = +7
	mod := c.GetSkillModifier(SkillStealth)
	assert.Equal(t, 7, mod, "Stealth modifier (+3 DEX + 4 expertise)")

	// Athletics without proficiency: +0 (STR)
	mod = c.GetSkillModifier(SkillAthletics)
	assert.Equal(t, 0, mod, "Athletics modifier")
}

func TestCharacterGetSavingThrowModifier(t *testing.T) {
	c := NewCharacter("char-1", "Test", "Human", "Fighter")
	c.AbilityScores = NewAbilityScoresFromValues(16, 14, 14, 10, 12, 8)
	c.SavingThrows.SetProficiency(AbilityStrength, true)
	c.SavingThrows.SetProficiency(AbilityConstitution, true)

	// STR save proficient: +3 (STR) + 2 (prof) = +5
	mod := c.GetSavingThrowModifier(AbilityStrength)
	assert.Equal(t, 5, mod, "STR save")

	// WIS save not proficient: +1 (WIS) = +1
	mod = c.GetSavingThrowModifier(AbilityWisdom)
	assert.Equal(t, 1, mod, "WIS save")
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
	assert.Equal(t, 15, dc, "Spell save DC")

	// Spell attack bonus: 4 (INT) + 3 (prof) = +7
	atk := c.GetSpellAttackBonus()
	assert.Equal(t, 7, atk, "Spell attack")
}

func TestCharacterSpellcastingNil(t *testing.T) {
	c := NewCharacter("char-1", "Test", "Human", "Fighter")

	assert.Equal(t, 0, c.GetSpellSaveDC(), "Non-caster should have 0 spell save DC")
	assert.Equal(t, 0, c.GetSpellAttackBonus(), "Non-caster should have 0 spell attack")
	assert.False(t, c.IsSpellcaster(), "Fighter should not be a spellcaster")
}

func TestCharacterDamageAndHealing(t *testing.T) {
	c := NewCharacter("char-1", "Test", "Human", "Fighter")
	c.CombatStats.HitPoints = NewHitPoints(20)

	c.TakeDamage(8)
	assert.Equal(t, 12, c.CombatStats.HitPoints.Current, "HP after damage")

	c.Heal(5)
	assert.Equal(t, 17, c.CombatStats.HitPoints.Current, "HP after heal")
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
	assert.Equal(t, 30, c.CombatStats.HitPoints.Current, "HP after rest")
	assert.Equal(t, 0, c.CombatStats.HitPoints.Temporary, "Temp HP after rest")

	// Hit dice should recover (half of 5 = 2)
	assert.Equal(t, 2, c.CombatStats.HitDice.Remaining, "Hit dice after rest")

	// Exhaustion should decrease by 1
	assert.Equal(t, 1, c.CombatStats.ExhaustionLevel, "Exhaustion after rest")

	// Death saves should reset
	assert.Equal(t, 0, c.CombatStats.DeathSaves.Successes, "Death saves successes should reset after long rest")
	assert.Equal(t, 0, c.CombatStats.DeathSaves.Failures, "Death saves failures should reset after long rest")

	// Spell slots should be restored
	assert.Equal(t, 4, c.Spellcasting.SpellSlots.Level1.Remaining, "Level 1 slots")
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
	assert.Equal(t, 2, c.Spellcasting.PactMagic.Remaining, "Pact slots after short rest")
}

func TestCharacterJSONRoundTrip(t *testing.T) {
	c := NewCharacter("char-1", "Gandalf", "Human", "Wizard")
	c.AbilityScores = NewAbilityScoresFromValues(10, 14, 12, 18, 16, 12)
	c.Skills.SetProficiency(SkillArcana, Proficient)
	c.SavingThrows.SetProficiency(AbilityIntelligence, true)
	c.Info.Level = 5

	// Serialize
	data, err := c.ToJSON()
	require.NoError(t, err, "ToJSON")

	// Deserialize
	c2, err := FromJSON(data)
	require.NoError(t, err, "FromJSON")

	// Verify
	assert.Equal(t, "Gandalf", c2.Info.Name, "Name after round trip")
	assert.Equal(t, 18, c2.AbilityScores.Intelligence.Base, "INT after round trip")
	assert.Equal(t, Proficient, c2.Skills.Arcana.Proficiency, "Arcana proficiency should be preserved")
}

func TestCharacterValidate(t *testing.T) {
	// Valid character
	c := NewCharacter("char-1", "Test", "Human", "Fighter")
	errors := c.Validate()
	assert.Empty(t, errors, "Valid character should have no errors")

	// Invalid: missing ID
	c2 := NewCharacter("", "Test", "Human", "Fighter")
	errors = c2.Validate()
	assert.NotEmpty(t, errors, "Character without ID should have validation error")

	// Invalid: missing name
	c3 := NewCharacter("id-1", "", "Human", "Fighter")
	errors = c3.Validate()
	assert.NotEmpty(t, errors, "Character without name should have validation error")

	// Invalid: level out of range
	c4 := NewCharacter("id-1", "Test", "Human", "Fighter")
	c4.Info.Level = 25
	errors = c4.Validate()
	assert.NotEmpty(t, errors, "Character with level 25 should have validation error")
}

func TestCharacterGetInitiative(t *testing.T) {
	c := NewCharacter("char-1", "Test", "Human", "Fighter")
	c.AbilityScores.Dexterity = AbilityScore{Base: 16} // +3 modifier

	assert.Equal(t, 3, c.GetInitiative())
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
		assert.Contains(t, result, key, "Missing expected key in JSON")
	}
}

func TestCharacterWriteTo(t *testing.T) {
	c := NewCharacter("char-1", "WriteTest", "Elf", "Wizard")
	c.AbilityScores.Strength.Base = 10
	c.AbilityScores.Intelligence.Base = 18
	c.Info.Level = 5

	// Write to buffer
	var buf bytes.Buffer
	_, err := c.WriteTo(&buf)
	require.NoError(t, err, "WriteTo()")

	// Verify we got JSON output
	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err, "Failed to unmarshal WriteTo output")

	// Verify some data
	assert.Equal(t, "char-1", result["id"])
}

func TestCharacterReadFrom(t *testing.T) {
	// Create test character JSON
	original := NewCharacter("char-2", "ReadTest", "Dwarf", "Cleric")
	original.AbilityScores.Wisdom.Base = 16
	original.AbilityScores.Constitution.Base = 15
	original.Info.Level = 3

	// Write to buffer
	var buf bytes.Buffer
	_, err := original.WriteTo(&buf)
	require.NoError(t, err, "WriteTo()")

	// Read back from buffer
	loaded, err := ReadFrom(&buf)
	require.NoError(t, err, "ReadFrom()")

	// Verify loaded character
	assert.Equal(t, "char-2", loaded.ID)
	assert.Equal(t, "ReadTest", loaded.Info.Name)
	assert.Equal(t, "Dwarf", loaded.Info.Race)
	assert.Equal(t, "Cleric", loaded.Info.Class)
	assert.Equal(t, 3, loaded.Info.Level)
	assert.Equal(t, 16, loaded.AbilityScores.Wisdom.Base)
	assert.Equal(t, 15, loaded.AbilityScores.Constitution.Base)
}

func TestCharacterWriteToReadFromRoundTrip(t *testing.T) {
	// Create complex character
	original := NewCharacter("char-3", "RoundTrip", "Tiefling", "Warlock")
	original.Info.Level = 7
	original.AbilityScores = NewAbilityScoresFromValues(10, 14, 12, 14, 10, 18)
	original.Skills.SetProficiency(SkillDeception, Proficient)
	original.Skills.SetProficiency(SkillPersuasion, Expertise)
	original.CombatStats.HitPoints.Maximum = 50
	original.CombatStats.HitPoints.Current = 35

	// Write to buffer
	var buf bytes.Buffer
	_, err := original.WriteTo(&buf)
	require.NoError(t, err, "WriteTo()")

	// Read back
	loaded, err := ReadFrom(&buf)
	require.NoError(t, err, "ReadFrom()")

	// Verify key fields match
	assert.Equal(t, original.ID, loaded.ID, "ID mismatch")
	assert.Equal(t, original.Info.Level, loaded.Info.Level, "Level mismatch")
	assert.Equal(t, original.AbilityScores.Charisma.Base, loaded.AbilityScores.Charisma.Base, "Charisma mismatch")
	assert.Equal(t, original.CombatStats.HitPoints.Maximum, loaded.CombatStats.HitPoints.Maximum, "Max HP mismatch")
}
