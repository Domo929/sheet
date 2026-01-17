package views

import (
	"strings"
	"testing"

	"github.com/Domo929/sheet/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewMainSheetModel(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)

	if model == nil {
		t.Fatal("NewMainSheetModel returned nil")
	}

	if model.character != char {
		t.Error("character not set correctly")
	}

	if model.focusArea != FocusAbilities {
		t.Errorf("expected initial focus area to be FocusAbilities, got %d", model.focusArea)
	}
}

func TestMainSheetModelInit(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)

	cmd := model.Init()
	if cmd != nil {
		t.Error("Init should return nil command")
	}
}

func TestMainSheetModelUpdate(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)

	// Set initial size
	model, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if model.width != 120 {
		t.Errorf("expected width 120, got %d", model.width)
	}
	if model.height != 40 {
		t.Errorf("expected height 40, got %d", model.height)
	}
}

func TestMainSheetModelTabNavigation(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)

	// Initial focus should be FocusAbilities
	if model.focusArea != FocusAbilities {
		t.Errorf("expected initial focus to be FocusAbilities (1), got %d", model.focusArea)
	}

	// Tab to next panel
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	if model.focusArea != FocusSkills {
		t.Errorf("expected focus to be FocusSkills (2) after tab, got %d", model.focusArea)
	}

	// Tab again
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	if model.focusArea != FocusSavingThrows {
		t.Errorf("expected focus to be FocusSavingThrows (3) after tab, got %d", model.focusArea)
	}

	// Tab wraps around
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	if model.focusArea != FocusHeader {
		t.Errorf("expected focus to wrap to FocusHeader (0), got %d", model.focusArea)
	}
}

func TestMainSheetModelShiftTabNavigation(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)

	// Set focus to Skills (2)
	model.focusArea = FocusSkills

	// Shift+Tab to previous panel
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if model.focusArea != FocusAbilities {
		t.Errorf("expected focus to be FocusAbilities (1) after shift+tab, got %d", model.focusArea)
	}

	// Shift+Tab again goes to FocusHeader (0)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if model.focusArea != FocusHeader {
		t.Errorf("expected focus to be FocusHeader (0) after shift+tab, got %d", model.focusArea)
	}

	// Shift+Tab wraps around from 0 to 4
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if model.focusArea != FocusCombatStats {
		t.Errorf("expected focus to wrap to FocusCombatStats (4), got %d", model.focusArea)
	}
}

func TestMainSheetModelView(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)
	model.width = 120
	model.height = 40

	view := model.View()

	// Check that character name is displayed
	if !strings.Contains(view, char.Info.Name) {
		t.Error("view should contain character name")
	}

	// Check that race and class are displayed
	if !strings.Contains(view, char.Info.Race) {
		t.Error("view should contain character race")
	}
	if !strings.Contains(view, char.Info.Class) {
		t.Error("view should contain character class")
	}

	// Check that section headers are displayed
	if !strings.Contains(view, "Abilities") {
		t.Error("view should contain Abilities header")
	}
	if !strings.Contains(view, "Skills") {
		t.Error("view should contain Skills header")
	}
	if !strings.Contains(view, "Combat") {
		t.Error("view should contain Combat header")
	}
	if !strings.Contains(view, "Saving Throws") {
		t.Error("view should contain Saving Throws header")
	}
}

func TestMainSheetModelViewNoCharacter(t *testing.T) {
	model := NewMainSheetModel(nil, nil)
	view := model.View()

	if view != "No character loaded" {
		t.Errorf("expected 'No character loaded', got %s", view)
	}
}

func TestMainSheetModelViewAbilityScores(t *testing.T) {
	char := createTestCharacter()
	// Set specific ability scores
	char.AbilityScores.Strength.Base = 16
	char.AbilityScores.Dexterity.Base = 14
	char.AbilityScores.Constitution.Base = 15

	model := NewMainSheetModel(char, nil)
	model.width = 120
	model.height = 40

	view := model.View()

	// Check ability abbreviations are present
	if !strings.Contains(view, "STR") {
		t.Error("view should contain STR")
	}
	if !strings.Contains(view, "DEX") {
		t.Error("view should contain DEX")
	}
	if !strings.Contains(view, "CON") {
		t.Error("view should contain CON")
	}
	if !strings.Contains(view, "INT") {
		t.Error("view should contain INT")
	}
	if !strings.Contains(view, "WIS") {
		t.Error("view should contain WIS")
	}
	if !strings.Contains(view, "CHA") {
		t.Error("view should contain CHA")
	}
}

func TestMainSheetModelViewSkills(t *testing.T) {
	char := createTestCharacter()
	// Set some skill proficiencies
	char.Skills.Stealth.Proficiency = models.Proficient
	char.Skills.Perception.Proficiency = models.Expertise

	model := NewMainSheetModel(char, nil)
	model.width = 120
	model.height = 40

	view := model.View()

	// Check some skill names are present
	if !strings.Contains(view, "Stealth") {
		t.Error("view should contain Stealth skill")
	}
	if !strings.Contains(view, "Perception") {
		t.Error("view should contain Perception skill")
	}
	if !strings.Contains(view, "Acrobatics") {
		t.Error("view should contain Acrobatics skill")
	}
}

func TestMainSheetModelViewCombatStats(t *testing.T) {
	char := createTestCharacter()
	char.CombatStats.HitPoints.Maximum = 28
	char.CombatStats.HitPoints.Current = 20
	char.CombatStats.ArmorClass = 16
	char.CombatStats.Speed = 30

	model := NewMainSheetModel(char, nil)
	model.width = 120
	model.height = 40

	view := model.View()

	// Check combat stat labels are present
	if !strings.Contains(view, "HP:") {
		t.Error("view should contain HP label")
	}
	if !strings.Contains(view, "AC:") {
		t.Error("view should contain AC label")
	}
	if !strings.Contains(view, "Speed:") {
		t.Error("view should contain Speed label")
	}
	if !strings.Contains(view, "Initiative:") {
		t.Error("view should contain Initiative label")
	}
}

func TestMainSheetModelViewSpellcasting(t *testing.T) {
	char := createTestCharacter()
	// Make character a spellcaster
	char.Spellcasting = &models.Spellcasting{
		Ability: models.AbilityWisdom,
	}
	char.AbilityScores.Wisdom.Base = 16 // +3 modifier

	model := NewMainSheetModel(char, nil)
	model.width = 120
	model.height = 40

	view := model.View()

	// Check spellcasting section is present
	if !strings.Contains(view, "Spellcasting") {
		t.Error("view should contain Spellcasting section for spellcasters")
	}
	if !strings.Contains(view, "Spell Save DC:") {
		t.Error("view should contain Spell Save DC")
	}
	if !strings.Contains(view, "Spell Attack:") {
		t.Error("view should contain Spell Attack")
	}
}

func TestMainSheetModelViewConditions(t *testing.T) {
	char := createTestCharacter()
	char.CombatStats.AddCondition(models.ConditionPoisoned)
	char.CombatStats.AddCondition(models.ConditionFrightened)

	model := NewMainSheetModel(char, nil)
	model.width = 120
	model.height = 40

	view := model.View()

	// Check conditions section is present
	if !strings.Contains(view, "Conditions") {
		t.Error("view should contain Conditions section when character has conditions")
	}
}

func TestMainSheetModelViewDeathSaves(t *testing.T) {
	char := createTestCharacter()
	// Set HP to 0 to trigger death saves display
	char.CombatStats.HitPoints.Current = 0
	char.CombatStats.DeathSaves.Successes = 2
	char.CombatStats.DeathSaves.Failures = 1

	model := NewMainSheetModel(char, nil)
	model.width = 120
	model.height = 40

	view := model.View()

	// Check death saves section is present
	if !strings.Contains(view, "Death Saves") {
		t.Error("view should contain Death Saves section when HP is 0")
	}
}

func TestMainSheetModelViewFooter(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)
	model.width = 120
	model.height = 40

	view := model.View()

	// Check help text is present
	if !strings.Contains(view, "quit") {
		t.Error("view should contain quit help")
	}
	if !strings.Contains(view, "inventory") {
		t.Error("view should contain inventory help")
	}
}

func TestFormatModifier(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "+0"},
		{1, "+1"},
		{5, "+5"},
		{-1, "-1"},
		{-3, "-3"},
	}

	for _, test := range tests {
		result := formatModifier(test.input)
		if result != test.expected {
			t.Errorf("formatModifier(%d) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

// Helper function to create a test character
func createTestCharacter() *models.Character {
	char := models.NewCharacter("test-id", "Aragorn", "Human", "Ranger")
	char.Info.Level = 5
	char.Info.Background = "Outlander"
	char.AbilityScores = models.NewAbilityScoresFromValues(16, 14, 14, 10, 14, 10)
	char.CombatStats = models.NewCombatStats(45, 10, 5, 30)
	char.CombatStats.ArmorClass = 15
	return char
}
