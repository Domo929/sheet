package views

import (
	"testing"

	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewProficiencySelectionManager(t *testing.T) {
	// Mock class with skill choices
	class := &data.Class{
		SkillChoices: data.SkillChoices{
			Count:   2,
			Options: []string{"Athletics", "Intimidation", "Survival"},
		},
	}

	// Mock background with fixed skills
	background := &data.Background{
		SkillProficiencies: []string{"Insight", "Religion"},
		ToolProficiency:    "Calligrapher's Supplies",
	}

	// Mock race with languages
	race := &data.Race{
		Languages: []string{"Common", "Elvish"},
	}

	psm := NewProficiencySelectionManager(class, background, race)

	// Check skill setup
	assert.Equal(t, 2, psm.skillsRequired, "expected 2 skills required")
	assert.Len(t, psm.skillOptions, 3, "expected 3 skill options")
	assert.Len(t, psm.backgroundSkills, 2, "expected 2 background skills")

	// Check tool setup
	assert.Equal(t, "Calligrapher's Supplies", psm.backgroundTool, "expected background tool 'Calligrapher's Supplies'")

	// Check language setup
	assert.Len(t, psm.racialLanguages, 2, "expected 2 racial languages")

	// Check initial section
	assert.Equal(t, ProfSectionSkills, psm.currentSection, "expected to start at skills section")

	// Check completion status
	assert.False(t, psm.IsComplete(), "proficiency selection should not be complete initially")
}

func TestProficiencySelectionNoSkillChoices(t *testing.T) {
	// Class with no skill choices
	class := &data.Class{
		SkillChoices: data.SkillChoices{
			Count:   0,
			Options: []string{},
		},
	}

	background := &data.Background{
		SkillProficiencies: []string{"Insight"},
		ToolProficiency:    "Tool",
	}

	race := &data.Race{
		Languages: []string{"Common"},
	}

	psm := NewProficiencySelectionManager(class, background, race)

	assert.Equal(t, 0, psm.skillsRequired, "expected 0 skills required")
	assert.True(t, psm.skillsComplete, "skills should be marked complete when no choices are required")
	assert.True(t, psm.IsComplete(), "proficiency selection should be complete when no choices are required")
}

func TestProficiencySelectionSectionNavigation(t *testing.T) {
	class := &data.Class{
		SkillChoices: data.SkillChoices{
			Count:   1,
			Options: []string{"Athletics"},
		},
	}

	background := &data.Background{
		SkillProficiencies: []string{},
		ToolProficiency:    "Tool",
	}

	race := &data.Race{
		Languages: []string{"Common"},
	}

	psm := NewProficiencySelectionManager(class, background, race)

	// Initially at skills
	assert.Equal(t, ProfSectionSkills, psm.currentSection, "should start at skills section")

	// Since tools and languages are auto-complete, NextSection should return false
	assert.False(t, psm.NextSection(), "NextSection should return false when no more sections with choices")

	// PreviousSection from skills should return false
	psm.currentSection = ProfSectionSkills
	assert.False(t, psm.PreviousSection(), "PreviousSection from first section should return false")
}

func TestApplyToCharacter(t *testing.T) {
	class := &data.Class{
		SkillChoices: data.SkillChoices{
			Count:   2,
			Options: []string{"Athletics", "Intimidation"},
		},
	}

	background := &data.Background{
		SkillProficiencies: []string{"Insight", "Religion"},
		ToolProficiency:    "Calligrapher's Supplies",
	}

	race := &data.Race{
		Languages: []string{"Common", "Elvish"},
	}

	psm := NewProficiencySelectionManager(class, background, race)

	// Simulate selecting skills
	psm.skillSelector.SetSelected([]int{0, 1}) // Select Athletics and Intimidation

	// Create a character
	char := models.NewCharacter("test-id", "Test Character", "Human", "Fighter")

	// Apply proficiencies
	psm.ApplyToCharacter(char)

	// Check class skill proficiencies were applied
	assert.Equal(t, models.Proficient, char.Skills.Get(models.SkillAthletics).Proficiency, "Athletics should be proficient")
	assert.Equal(t, models.Proficient, char.Skills.Get(models.SkillIntimidation).Proficiency, "Intimidation should be proficient")

	// Check background skill proficiencies were applied
	assert.Equal(t, models.Proficient, char.Skills.Get(models.SkillInsight).Proficiency, "Insight from background should be proficient")
	assert.Equal(t, models.Proficient, char.Skills.Get(models.SkillReligion).Proficiency, "Religion from background should be proficient")

	// Check tool proficiency
	assert.True(t, char.Proficiencies.HasTool("Calligrapher's Supplies"), "Character should have Calligrapher's Supplies tool proficiency")

	// Check languages
	assert.True(t, char.Proficiencies.HasLanguage("Common"), "Character should know Common")
	assert.True(t, char.Proficiencies.HasLanguage("Elvish"), "Character should know Elvish")
}

func TestSkillNameToKey(t *testing.T) {
	tests := []struct {
		input    string
		expected models.SkillName
	}{
		{"Athletics", models.SkillAthletics},
		{"Animal Handling", models.SkillAnimalHandling},
		{"Sleight Of Hand", models.SkillSleightOfHand},
		{"Acrobatics", models.SkillAcrobatics},
		{"Perception", models.SkillPerception},
	}

	for _, tt := range tests {
		result := skillNameToKey(tt.input)
		assert.Equal(t, tt.expected, result, "skillNameToKey(%q)", tt.input)
	}
}

func TestGetSelectedMethods(t *testing.T) {
	class := &data.Class{
		SkillChoices: data.SkillChoices{
			Count:   2,
			Options: []string{"Athletics", "Perception"},
		},
	}

	background := &data.Background{
		SkillProficiencies: []string{},
		ToolProficiency:    "Tool",
	}

	race := &data.Race{
		Languages: []string{"Common"},
	}

	psm := NewProficiencySelectionManager(class, background, race)

	// Select skills
	psm.skillSelector.SetSelected([]int{0, 1})

	selectedSkills := psm.GetSelectedSkills()
	assert.Len(t, selectedSkills, 2, "expected 2 selected skills")

	// Test with no selections required
	class.SkillChoices.Count = 0
	psm2 := NewProficiencySelectionManager(class, background, race)

	assert.Empty(t, psm2.GetSelectedSkills(), "expected empty selection when no skills required")
}
