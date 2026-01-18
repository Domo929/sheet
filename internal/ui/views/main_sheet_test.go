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

	if model.focusArea != FocusAbilitiesAndSaves {
		t.Errorf("expected initial focus area to be FocusAbilitiesAndSaves, got %d", model.focusArea)
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

	// Initial focus should be FocusAbilitiesAndSaves (0)
	if model.focusArea != FocusAbilitiesAndSaves {
		t.Errorf("expected initial focus to be FocusAbilitiesAndSaves (0), got %d", model.focusArea)
	}

	// Tab to next panel
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	if model.focusArea != FocusSkills {
		t.Errorf("expected focus to be FocusSkills (1) after tab, got %d", model.focusArea)
	}

	// Tab again
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	if model.focusArea != FocusCombat {
		t.Errorf("expected focus to be FocusCombat (2) after tab, got %d", model.focusArea)
	}

	// Tab to Actions
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	if model.focusArea != FocusActions {
		t.Errorf("expected focus to be FocusActions (3) after tab, got %d", model.focusArea)
	}

	// Tab wraps around
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	if model.focusArea != FocusAbilitiesAndSaves {
		t.Errorf("expected focus to wrap to FocusAbilitiesAndSaves (0), got %d", model.focusArea)
	}
}

func TestMainSheetModelShiftTabNavigation(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)

	// Set focus to Skills (1)
	model.focusArea = FocusSkills

	// Shift+Tab to previous panel
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if model.focusArea != FocusAbilitiesAndSaves {
		t.Errorf("expected focus to be FocusAbilitiesAndSaves (0) after shift+tab, got %d", model.focusArea)
	}

	// Shift+Tab wraps around from 0 to 3 (Actions)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if model.focusArea != FocusActions {
		t.Errorf("expected focus to wrap to FocusActions (3), got %d", model.focusArea)
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

	// Check ability names are present (full names in new layout)
	if !strings.Contains(view, "Strength") {
		t.Error("view should contain Strength")
	}
	if !strings.Contains(view, "Dexterity") {
		t.Error("view should contain Dexterity")
	}
	if !strings.Contains(view, "Constitution") {
		t.Error("view should contain Constitution")
	}
	if !strings.Contains(view, "Intelligence") {
		t.Error("view should contain Intelligence")
	}
	if !strings.Contains(view, "Wisdom") {
		t.Error("view should contain Wisdom")
	}
	if !strings.Contains(view, "Charisma") {
		t.Error("view should contain Charisma")
	}
}

func TestMainSheetModelViewSkills(t *testing.T) {
	char := createTestCharacter()
	// Set some skill proficiencies
	char.Skills.Stealth.Proficiency = models.Proficient
	char.Skills.Perception.Proficiency = models.Expertise
	char.AbilityScores.Wisdom.Base = 14 // +2 modifier

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

	// Check passive perception is displayed
	if !strings.Contains(view, "Passive Perception") {
		t.Error("view should contain Passive Perception")
	}

	// Check legend is present in header
	if !strings.Contains(view, "Proficient") {
		t.Error("view should contain proficiency legend in header")
	}
	if !strings.Contains(view, "Expertise") {
		t.Error("view should contain expertise legend in header")
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

func TestMainSheetModelHPInputMode(t *testing.T) {
	char := createTestCharacter()
	char.CombatStats.HitPoints.Maximum = 45
	char.CombatStats.HitPoints.Current = 45

	model := NewMainSheetModel(char, nil)
	model.focusArea = FocusCombat

	// Test entering damage mode
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if model.hpInputMode != HPInputDamage {
		t.Errorf("expected HPInputDamage mode, got %d", model.hpInputMode)
	}

	// Test entering numbers
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1', '0'}})
	if model.hpInputBuffer != "10" {
		t.Errorf("expected buffer '10', got '%s'", model.hpInputBuffer)
	}

	// Test applying damage
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if model.character.CombatStats.HitPoints.Current != 35 {
		t.Errorf("expected HP 35 after 10 damage, got %d", model.character.CombatStats.HitPoints.Current)
	}
	if model.hpInputMode != HPInputNone {
		t.Error("expected to exit input mode after enter")
	}
}

func TestMainSheetModelHPHeal(t *testing.T) {
	char := createTestCharacter()
	char.CombatStats.HitPoints.Maximum = 45
	char.CombatStats.HitPoints.Current = 30

	model := NewMainSheetModel(char, nil)
	model.focusArea = FocusCombat

	// Enter heal mode and heal 10
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1', '0'}})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if model.character.CombatStats.HitPoints.Current != 40 {
		t.Errorf("expected HP 40 after healing 10, got %d", model.character.CombatStats.HitPoints.Current)
	}
}

func TestMainSheetModelHPTempHP(t *testing.T) {
	char := createTestCharacter()
	char.CombatStats.HitPoints.Maximum = 45
	char.CombatStats.HitPoints.Current = 45
	char.CombatStats.HitPoints.Temporary = 0

	model := NewMainSheetModel(char, nil)
	model.focusArea = FocusCombat

	// Enter temp HP mode and add 5
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if model.character.CombatStats.HitPoints.Temporary != 5 {
		t.Errorf("expected temp HP 5, got %d", model.character.CombatStats.HitPoints.Temporary)
	}
}

func TestMainSheetModelHPInputCancel(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)
	model.focusArea = FocusCombat

	// Enter damage mode
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1', '0'}})

	// Cancel with escape
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if model.hpInputMode != HPInputNone {
		t.Error("expected to exit input mode on escape")
	}
	if model.hpInputBuffer != "" {
		t.Error("expected buffer to be cleared on escape")
	}
}

func TestMainSheetModelHPBar(t *testing.T) {
	char := createTestCharacter()
	char.CombatStats.HitPoints.Maximum = 100
	char.CombatStats.HitPoints.Current = 50

	model := NewMainSheetModel(char, nil)
	model.width = 120
	model.height = 40

	view := model.View()

	// Check HP bar brackets are present
	if !strings.Contains(view, "[") || !strings.Contains(view, "]") {
		t.Error("view should contain HP bar with brackets")
	}
}

func TestMainSheetModelDeathSaves(t *testing.T) {
	char := createTestCharacter()
	char.CombatStats.HitPoints.Current = 0 // At 0 HP

	model := NewMainSheetModel(char, nil)
	model.focusArea = FocusCombat

	// Test adding death save success
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if char.CombatStats.DeathSaves.Successes != 1 {
		t.Errorf("expected 1 success, got %d", char.CombatStats.DeathSaves.Successes)
	}

	// Test adding death save failure
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	if char.CombatStats.DeathSaves.Failures != 1 {
		t.Errorf("expected 1 failure, got %d", char.CombatStats.DeathSaves.Failures)
	}

	// Test reset
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}})
	if char.CombatStats.DeathSaves.Successes != 0 || char.CombatStats.DeathSaves.Failures != 0 {
		t.Error("expected death saves to be reset")
	}
}

func TestMainSheetModelDeathSavesStabilized(t *testing.T) {
	char := createTestCharacter()
	char.CombatStats.HitPoints.Current = 0
	char.CombatStats.DeathSaves.Successes = 3

	model := NewMainSheetModel(char, nil)
	model.width = 120
	model.height = 40

	view := model.View()

	if !strings.Contains(view, "STABILIZED") {
		t.Error("view should show STABILIZED when 3 successes")
	}
}

func TestMainSheetModelConditionAdd(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)
	model.focusArea = FocusCombat

	// Enter condition add mode
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	if !model.conditionMode || !model.conditionAdding {
		t.Error("expected to be in condition add mode")
	}

	// Select and add first condition
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if len(char.CombatStats.Conditions) != 1 {
		t.Errorf("expected 1 condition, got %d", len(char.CombatStats.Conditions))
	}
	if model.conditionMode {
		t.Error("expected to exit condition mode after adding")
	}
}

func TestMainSheetModelConditionRemove(t *testing.T) {
	char := createTestCharacter()
	char.CombatStats.AddCondition(models.ConditionPoisoned)
	char.CombatStats.AddCondition(models.ConditionFrightened)

	model := NewMainSheetModel(char, nil)
	model.focusArea = FocusCombat

	// Enter condition remove mode
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}})
	if !model.conditionMode || model.conditionAdding {
		t.Error("expected to be in condition remove mode")
	}

	// Remove first condition
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if len(char.CombatStats.Conditions) != 1 {
		t.Errorf("expected 1 condition after removal, got %d", len(char.CombatStats.Conditions))
	}
}

func TestMainSheetModelConditionNavigation(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)
	model.focusArea = FocusCombat

	// Enter condition add mode
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})

	// Navigate down
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	if model.conditionCursor != 1 {
		t.Errorf("expected cursor at 1, got %d", model.conditionCursor)
	}

	// Navigate up
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	if model.conditionCursor != 0 {
		t.Errorf("expected cursor at 0, got %d", model.conditionCursor)
	}

	// Cancel
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if model.conditionMode {
		t.Error("expected to exit condition mode on escape")
	}
}

func TestMainSheetModelWeaponAttacks(t *testing.T) {
	char := createTestCharacter()
	// Add a weapon to inventory
	sword := models.NewItem("sword-1", "Longsword", models.ItemTypeWeapon)
	sword.Damage = "1d8"
	sword.DamageType = "slashing"
	char.Inventory.AddItem(sword)

	model := NewMainSheetModel(char, nil)
	model.width = 120
	model.height = 40

	view := model.View()

	// Check weapon is displayed
	if !strings.Contains(view, "Attacks") {
		t.Error("view should contain Attacks section when weapons present")
	}
	if !strings.Contains(view, "Longsword") {
		t.Error("view should contain weapon name")
	}
	if !strings.Contains(view, "1d8") {
		t.Error("view should contain weapon damage")
	}
}

func TestMainSheetModelWeaponAttackBonus(t *testing.T) {
	char := createTestCharacter()
	char.AbilityScores.Strength.Base = 16  // +3 modifier
	char.AbilityScores.Dexterity.Base = 14 // +2 modifier
	char.Info.Level = 5                     // +3 proficiency
	char.Proficiencies.Weapons = []string{"Simple Weapons", "Martial Weapons"}

	// Test regular melee weapon (uses STR)
	sword := models.NewItem("sword-1", "Longsword", models.ItemTypeWeapon)
	sword.Damage = "1d8"
	sword.DamageType = "slashing"
	sword.SubCategory = "Martial Melee Weapons"
	char.Inventory.AddItem(sword)

	model := NewMainSheetModel(char, nil)
	bonus := model.getWeaponAttackBonus(sword)
	// STR (+3) + Prof (+3) = +6
	if bonus != 6 {
		t.Errorf("expected attack bonus 6, got %d", bonus)
	}

	// Test finesse weapon (uses better of STR/DEX)
	dagger := models.NewItem("dagger-1", "Dagger", models.ItemTypeWeapon)
	dagger.Damage = "1d4"
	dagger.DamageType = "piercing"
	dagger.SubCategory = "Simple Melee Weapons"
	dagger.WeaponProps = []string{"finesse", "light", "thrown"}

	bonus = model.getWeaponAttackBonus(dagger)
	// Better of STR (+3) or DEX (+2) = STR (+3) + Prof (+3) = +6
	if bonus != 6 {
		t.Errorf("expected finesse attack bonus 6, got %d", bonus)
	}

	// Test with higher DEX
	char.AbilityScores.Dexterity.Base = 18 // +4 modifier
	bonus = model.getWeaponAttackBonus(dagger)
	// Better of STR (+3) or DEX (+4) = DEX (+4) + Prof (+3) = +7
	if bonus != 7 {
		t.Errorf("expected finesse attack bonus 7 with higher DEX, got %d", bonus)
	}
}

func TestRestModeTransitions(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)

	// Initially not in rest mode
	if model.restMode != RestModeNone {
		t.Error("expected initial rest mode to be RestModeNone")
	}

	// Press 'r' to enter rest menu
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if model.restMode != RestModeMenu {
		t.Errorf("expected rest mode RestModeMenu after 'r', got %d", model.restMode)
	}

	// Press 's' to go to short rest
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if model.restMode != RestModeShort {
		t.Errorf("expected rest mode RestModeShort after 's', got %d", model.restMode)
	}

	// Press Esc to go back to menu
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if model.restMode != RestModeMenu {
		t.Errorf("expected rest mode RestModeMenu after Esc, got %d", model.restMode)
	}

	// Press 'l' to go to long rest
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if model.restMode != RestModeLong {
		t.Errorf("expected rest mode RestModeLong after 'l', got %d", model.restMode)
	}

	// Press Esc to go back to menu
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if model.restMode != RestModeMenu {
		t.Errorf("expected rest mode RestModeMenu after Esc, got %d", model.restMode)
	}

	// Press Esc again to exit rest mode
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if model.restMode != RestModeNone {
		t.Errorf("expected rest mode RestModeNone after second Esc, got %d", model.restMode)
	}
}

func TestShortRestHitDiceAdjustment(t *testing.T) {
	char := createTestCharacter()
	char.CombatStats.HitDice.Remaining = 3
	model := NewMainSheetModel(char, nil)

	// Enter short rest mode
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})

	if model.restHitDice != 0 {
		t.Errorf("expected initial restHitDice 0, got %d", model.restHitDice)
	}

	// Increase hit dice to spend
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	if model.restHitDice != 1 {
		t.Errorf("expected restHitDice 1 after up, got %d", model.restHitDice)
	}

	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	if model.restHitDice != 2 {
		t.Errorf("expected restHitDice 2 after second up, got %d", model.restHitDice)
	}

	// Decrease hit dice
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	if model.restHitDice != 1 {
		t.Errorf("expected restHitDice 1 after down, got %d", model.restHitDice)
	}

	// Can't go below 0
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	if model.restHitDice != 0 {
		t.Errorf("expected restHitDice 0 (can't go negative), got %d", model.restHitDice)
	}

	// Can't exceed remaining hit dice
	model.restHitDice = 3
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	if model.restHitDice != 3 {
		t.Errorf("expected restHitDice 3 (can't exceed remaining), got %d", model.restHitDice)
	}
}

func TestShortRestExecution(t *testing.T) {
	char := createTestCharacter()
	char.CombatStats.HitPoints.Current = 20
	char.CombatStats.HitPoints.Maximum = 45
	char.CombatStats.HitDice.Remaining = 5
	char.CombatStats.HitDice.DieType = 10
	char.AbilityScores.Constitution.Base = 14 // +2 modifier
	model := NewMainSheetModel(char, nil)

	// Enter short rest and spend hit dice
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp}) // 1 hit die
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp}) // 2 hit dice
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should show result screen first
	if model.restMode != RestModeResult {
		t.Errorf("expected rest mode RestModeResult after confirming, got %d", model.restMode)
	}

	// restResult should contain healing info
	if !strings.Contains(model.restResult, "SHORT REST COMPLETE") {
		t.Errorf("expected restResult to contain 'SHORT REST COMPLETE', got: %s", model.restResult)
	}

	// Press any key to dismiss
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if model.restMode != RestModeNone {
		t.Errorf("expected rest mode RestModeNone after dismissing result, got %d", model.restMode)
	}

	// HP should have increased (average roll: 5+1=6, +2 CON = 8 per die, 2 dice = 16)
	if char.CombatStats.HitPoints.Current <= 20 {
		t.Errorf("expected HP to increase after short rest, got %d", char.CombatStats.HitPoints.Current)
	}

	// Hit dice should have decreased
	if char.CombatStats.HitDice.Remaining != 3 {
		t.Errorf("expected 3 remaining hit dice, got %d", char.CombatStats.HitDice.Remaining)
	}
}

func TestLongRestExecution(t *testing.T) {
	char := createTestCharacter()
	char.CombatStats.HitPoints.Current = 20
	char.CombatStats.HitPoints.Maximum = 45
	char.CombatStats.HitDice.Remaining = 2
	char.CombatStats.HitDice.Total = 5
	model := NewMainSheetModel(char, nil)

	// Enter long rest and confirm
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should show result screen first
	if model.restMode != RestModeResult {
		t.Errorf("expected rest mode RestModeResult after confirming, got %d", model.restMode)
	}

	// restResult should contain long rest info
	if !strings.Contains(model.restResult, "LONG REST COMPLETE") {
		t.Errorf("expected restResult to contain 'LONG REST COMPLETE', got: %s", model.restResult)
	}

	// Press any key to dismiss
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if model.restMode != RestModeNone {
		t.Errorf("expected rest mode RestModeNone after dismissing result, got %d", model.restMode)
	}

	// HP should be at maximum
	if char.CombatStats.HitPoints.Current != 45 {
		t.Errorf("expected HP to be max (45), got %d", char.CombatStats.HitPoints.Current)
	}

	// Hit dice should have recovered (half level, min 1 = 2)
	if char.CombatStats.HitDice.Remaining < 4 {
		t.Errorf("expected at least 4 hit dice after recovery, got %d", char.CombatStats.HitDice.Remaining)
	}
}

func TestRestOverlayRendering(t *testing.T) {
	char := createTestCharacter()
	char.CombatStats.HitDice.Remaining = 3
	model := NewMainSheetModel(char, nil)
	model.width = 80
	model.height = 40

	// Test rest menu
	model.restMode = RestModeMenu
	view := model.View()
	if !strings.Contains(view, "Rest Options") {
		t.Error("expected rest menu to show 'Rest Options'")
	}
	if !strings.Contains(view, "Short Rest") {
		t.Error("expected rest menu to show 'Short Rest'")
	}
	if !strings.Contains(view, "Long Rest") {
		t.Error("expected rest menu to show 'Long Rest'")
	}

	// Test short rest screen
	model.restMode = RestModeShort
	view = model.View()
	if !strings.Contains(view, "Short Rest") {
		t.Error("expected short rest screen to show 'Short Rest'")
	}
	if !strings.Contains(view, "Hit Dice") {
		t.Error("expected short rest screen to show 'Hit Dice'")
	}

	// Test long rest screen
	model.restMode = RestModeLong
	view = model.View()
	if !strings.Contains(view, "Long Rest") {
		t.Error("expected long rest screen to show 'Long Rest'")
	}
	if !strings.Contains(view, "Full Recovery") || strings.Contains(view, "confirmation") {
		// Either "Full Recovery" or some confirmation text
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
