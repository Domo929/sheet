package views

import (
	"testing"

	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
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
	if psm.skillsRequired != 2 {
		t.Errorf("expected 2 skills required, got %d", psm.skillsRequired)
	}

	if len(psm.skillOptions) != 3 {
		t.Errorf("expected 3 skill options, got %d", len(psm.skillOptions))
	}

	if len(psm.backgroundSkills) != 2 {
		t.Errorf("expected 2 background skills, got %d", len(psm.backgroundSkills))
	}

	// Check tool setup
	if psm.backgroundTool != "Calligrapher's Supplies" {
		t.Errorf("expected background tool 'Calligrapher's Supplies', got '%s'", psm.backgroundTool)
	}

	// Check language setup
	if len(psm.racialLanguages) != 2 {
		t.Errorf("expected 2 racial languages, got %d", len(psm.racialLanguages))
	}

	// Check initial section
	if psm.currentSection != ProfSectionSkills {
		t.Error("expected to start at skills section")
	}

	// Check completion status
	if psm.IsComplete() {
		t.Error("proficiency selection should not be complete initially")
	}
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

	if psm.skillsRequired != 0 {
		t.Errorf("expected 0 skills required, got %d", psm.skillsRequired)
	}

	if !psm.skillsComplete {
		t.Error("skills should be marked complete when no choices are required")
	}

	if !psm.IsComplete() {
		t.Error("proficiency selection should be complete when no choices are required")
	}
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
	if psm.currentSection != ProfSectionSkills {
		t.Error("should start at skills section")
	}

	// Since tools and languages are auto-complete, NextSection should return false
	if psm.NextSection() {
		t.Error("NextSection should return false when no more sections with choices")
	}

	// PreviousSection from skills should return false
	psm.currentSection = ProfSectionSkills
	if psm.PreviousSection() {
		t.Error("PreviousSection from first section should return false")
	}
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
	if char.Skills.Get(models.SkillAthletics).Proficiency != models.Proficient {
		t.Error("Athletics should be proficient")
	}

	if char.Skills.Get(models.SkillIntimidation).Proficiency != models.Proficient {
		t.Error("Intimidation should be proficient")
	}

	// Check background skill proficiencies were applied
	if char.Skills.Get(models.SkillInsight).Proficiency != models.Proficient {
		t.Error("Insight from background should be proficient")
	}

	if char.Skills.Get(models.SkillReligion).Proficiency != models.Proficient {
		t.Error("Religion from background should be proficient")
	}

	// Check tool proficiency
	if !char.Proficiencies.HasTool("Calligrapher's Supplies") {
		t.Error("Character should have Calligrapher's Supplies tool proficiency")
	}

	// Check languages
	if !char.Proficiencies.HasLanguage("Common") {
		t.Error("Character should know Common")
	}

	if !char.Proficiencies.HasLanguage("Elvish") {
		t.Error("Character should know Elvish")
	}
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
		if result != tt.expected {
			t.Errorf("skillNameToKey(%q) = %v, want %v", tt.input, result, tt.expected)
		}
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
	if len(selectedSkills) != 2 {
		t.Errorf("expected 2 selected skills, got %d", len(selectedSkills))
	}

	// Test with no selections required
	class.SkillChoices.Count = 0
	psm2 := NewProficiencySelectionManager(class, background, race)

	if len(psm2.GetSelectedSkills()) != 0 {
		t.Error("expected empty selection when no skills required")
	}
}
