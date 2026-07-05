package export

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Domo929/sheet/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleCharacter() *models.Character {
	c := models.NewCharacter("exp-1", "Lyra", "Elf", "Wizard")
	c.Info.Level = 3
	c.Info.Subrace = "High Elf"
	c.Info.Subclass = "Evoker"
	c.Info.Background = "Sage"
	c.Info.Alignment = models.Alignment("Chaotic Good")
	c.AbilityScores = models.NewAbilityScoresFromValues(8, 16, 14, 17, 12, 10)
	c.CombatStats = models.NewCombatStats(20, 6, 3, 30)
	c.CombatStats.ArmorClass = 13
	c.SavingThrows.SetProficiency(models.AbilityIntelligence, true)
	c.SavingThrows.SetProficiency(models.AbilityWisdom, true)
	c.Skills.SetProficiency(models.SkillArcana, models.Proficient)
	c.Skills.SetProficiency(models.SkillHistory, models.Expertise)
	c.Proficiencies.Languages = []string{"Common", "Elvish"}
	c.Features.AddRacialTrait("Darkvision", "Elf", "See 60 ft in dim light.")

	sc := models.NewSpellcasting(models.AbilityIntelligence)
	sc.SpellSlots.SetSlots(1, 4)
	sc.SpellSlots.SetSlots(2, 2)
	sc.AddCantrip("Fire Bolt")
	sc.AddSpell("Magic Missile", 1)
	sc.AddSpell("Shield", 1)
	sc.KnownSpells[0].Prepared = true
	c.Spellcasting = &sc

	item := models.NewItem("spellbook", "Spellbook", models.ItemTypeGeneral)
	item.Weight = 3
	c.Inventory.AddItem(item)
	c.Inventory.Currency.Gold = 25

	c.Personality.Traits = []string{"Curious"}
	return c
}

func TestToMarkdownContainsSections(t *testing.T) {
	md := ToMarkdown(sampleCharacter())

	wants := []string{
		"# Lyra",
		"Level 3 Elf (High Elf) Wizard (Evoker)",
		"Background: Sage",
		"## Combat",
		"| Armor Class | 13 |",
		"## Ability Scores",
		"| Intelligence | 17 | +3 | +5 *(prof)* |", // +3 mod, +2 prof for L3
		"## Skills",
		"Arcana",
		"## Spellcasting",
		"Magic Missile",
		"Fire Bolt",
		"## Inventory",
		"Spellbook",
		"## Personality",
	}
	for _, w := range wants {
		assert.Contains(t, md, w, "markdown should contain %q", w)
	}
}

func TestToMarkdownOmitsSpellcastingForNonCaster(t *testing.T) {
	c := models.NewCharacter("f-1", "Grok", "Orc", "Barbarian")
	md := ToMarkdown(c)
	assert.NotContains(t, md, "## Spellcasting")
	assert.Contains(t, md, "# Grok")
}

func TestToMarkdownCompanions(t *testing.T) {
	c := models.NewCharacter("r-1", "Kael", "Human", "Ranger")
	// No companions: section omitted.
	assert.NotContains(t, ToMarkdown(c), "## Companions")

	c.AddCompanion(models.Companion{
		ID: "c1", Name: "Dire Wolf", Kind: models.CompanionPet,
		Size: "Large", Type: "Beast", AC: 14, MaxHP: 37, CurrentHP: 30,
		Speed: "50 ft.", Abilities: [6]int{17, 15, 15, 3, 12, 7},
		Attacks: []models.CompanionAttack{{Name: "Bite", Bonus: 5, Damage: "2d6 + 3 piercing"}},
	})
	md := ToMarkdown(c)
	assert.Contains(t, md, "## Companions & Summons")
	assert.Contains(t, md, "### Dire Wolf")
	assert.Contains(t, md, "30/37")
	assert.Contains(t, md, "Bite")
	assert.Contains(t, md, "+3", "STR 17 modifier should render")
}

func TestToJSONRoundTrips(t *testing.T) {
	c := sampleCharacter()
	data, err := ToJSON(c)
	require.NoError(t, err)

	var back models.Character
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, c.Info.Name, back.Info.Name)
	assert.Equal(t, c.Info.Level, back.Info.Level)
}

func TestWriteFiles(t *testing.T) {
	c := sampleCharacter()
	dir := filepath.Join(t.TempDir(), "out")

	mdPath, jsonPath, err := WriteFiles(c, dir)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "Lyra.md"), mdPath)
	assert.Equal(t, filepath.Join(dir, "Lyra.json"), jsonPath)

	mdBytes, err := os.ReadFile(mdPath)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(mdBytes), "# Lyra"))

	jsonBytes, err := os.ReadFile(jsonPath)
	require.NoError(t, err)
	var back models.Character
	require.NoError(t, json.Unmarshal(jsonBytes, &back))
	assert.Equal(t, "Lyra", back.Info.Name)
}

func TestSanitizeFilename(t *testing.T) {
	assert.Equal(t, "Sir_Reginald", SanitizeFilename("Sir Reginald"))
	assert.Equal(t, "AB", SanitizeFilename("A/B"))
	assert.Equal(t, "character", SanitizeFilename(""))
}
