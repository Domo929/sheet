package views

import (
	"strings"
	"testing"

	"github.com/Domo929/sheet/internal/domain"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMainSheetModel(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)

	require.NotNil(t, model, "NewMainSheetModel returned nil")
	assert.Equal(t, char, model.character, "character not set correctly")
	assert.Equal(t, FocusAbilitiesAndSaves, model.focusArea, "expected initial focus area to be FocusAbilitiesAndSaves")
}

func TestMainSheetModelInit(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)

	cmd := model.Init()
	assert.Nil(t, cmd, "Init should return nil command")
}

func TestMainSheetModelUpdate(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)

	// Set initial size
	model, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	assert.Equal(t, 120, model.width, "expected width 120")
	assert.Equal(t, 40, model.height, "expected height 40")
}

func TestMainSheetModelTabNavigation(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)

	// Initial focus should be FocusAbilitiesAndSaves (0)
	assert.Equal(t, FocusAbilitiesAndSaves, model.focusArea, "expected initial focus to be FocusAbilitiesAndSaves (0)")

	// Tab to next panel
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, FocusSkills, model.focusArea, "expected focus to be FocusSkills (1) after tab")

	// Tab again
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, FocusCombat, model.focusArea, "expected focus to be FocusCombat (2) after tab")

	// Tab to Actions
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, FocusActions, model.focusArea, "expected focus to be FocusActions (3) after tab")

	// Tab wraps around
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, FocusAbilitiesAndSaves, model.focusArea, "expected focus to wrap to FocusAbilitiesAndSaves (0)")
}

func TestMainSheetModelShiftTabNavigation(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)

	// Set focus to Skills (1)
	model.focusArea = FocusSkills

	// Shift+Tab to previous panel
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	assert.Equal(t, FocusAbilitiesAndSaves, model.focusArea, "expected focus to be FocusAbilitiesAndSaves (0) after shift+tab")

	// Shift+Tab wraps around from 0 to 3 (Actions)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	assert.Equal(t, FocusActions, model.focusArea, "expected focus to wrap to FocusActions (3)")
}

func TestMainSheetModelView(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)
	model.width = 120
	model.height = 40

	view := model.View()

	// Check that character name is displayed
	assert.Contains(t, view, char.Info.Name, "view should contain character name")

	// Check that race and class are displayed
	assert.Contains(t, view, char.Info.Race, "view should contain character race")
	assert.Contains(t, view, char.Info.Class, "view should contain character class")

	// Check that section headers are displayed
	assert.Contains(t, view, "Abilities", "view should contain Abilities header")
	assert.Contains(t, view, "Skills", "view should contain Skills header")
	assert.Contains(t, view, "Combat", "view should contain Combat header")
	assert.Contains(t, view, "Saving Throws", "view should contain Saving Throws header")
}

func TestMainSheetModelViewNoCharacter(t *testing.T) {
	model := NewMainSheetModel(nil, nil)
	view := model.View()

	assert.Equal(t, "No character loaded", view)
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
	assert.Contains(t, view, "Strength", "view should contain Strength")
	assert.Contains(t, view, "Dexterity", "view should contain Dexterity")
	assert.Contains(t, view, "Constitution", "view should contain Constitution")
	assert.Contains(t, view, "Intelligence", "view should contain Intelligence")
	assert.Contains(t, view, "Wisdom", "view should contain Wisdom")
	assert.Contains(t, view, "Charisma", "view should contain Charisma")
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
	assert.Contains(t, view, "Stealth", "view should contain Stealth skill")
	assert.Contains(t, view, "Perception", "view should contain Perception skill")
	assert.Contains(t, view, "Acrobatics", "view should contain Acrobatics skill")

	// Check passive perception is displayed
	assert.Contains(t, view, "Passive Perception", "view should contain Passive Perception")

	// Check legend is present in header
	assert.Contains(t, view, "Proficient", "view should contain proficiency legend in header")
	assert.Contains(t, view, "Expertise", "view should contain expertise legend in header")
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
	assert.Contains(t, view, "HP:", "view should contain HP label")
	assert.Contains(t, view, "AC:", "view should contain AC label")
	assert.Contains(t, view, "Speed:", "view should contain Speed label")
	assert.Contains(t, view, "Initiative:", "view should contain Initiative label")
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
	assert.Contains(t, view, "Spellcasting", "view should contain Spellcasting section for spellcasters")
	assert.Contains(t, view, "Spell Save DC:", "view should contain Spell Save DC")
	assert.Contains(t, view, "Spell Attack:", "view should contain Spell Attack")
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
	assert.Contains(t, view, "Conditions", "view should contain Conditions section when character has conditions")
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
	assert.Contains(t, view, "Death Saves", "view should contain Death Saves section when HP is 0")
}

func TestMainSheetModelViewFooter(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)
	model.width = 120
	model.height = 40

	view := model.View()

	// Check help text is present
	assert.Contains(t, view, "quit", "view should contain quit help")
	assert.Contains(t, view, "inventory", "view should contain inventory help")
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
		assert.Equal(t, test.expected, result, "formatModifier(%d)", test.input)
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
	assert.Equal(t, HPInputDamage, model.hpInputMode, "expected HPInputDamage mode")

	// Test entering numbers
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1', '0'}})
	assert.Equal(t, "10", model.hpInputBuffer, "expected buffer '10'")

	// Test applying damage
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, 35, model.character.CombatStats.HitPoints.Current, "expected HP 35 after 10 damage")
	assert.Equal(t, HPInputNone, model.hpInputMode, "expected to exit input mode after enter")
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

	assert.Equal(t, 40, model.character.CombatStats.HitPoints.Current, "expected HP 40 after healing 10")
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

	assert.Equal(t, 5, model.character.CombatStats.HitPoints.Temporary, "expected temp HP 5")
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

	assert.Equal(t, HPInputNone, model.hpInputMode, "expected to exit input mode on escape")
	assert.Empty(t, model.hpInputBuffer, "expected buffer to be cleared on escape")
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
	assert.True(t, strings.Contains(view, "[") && strings.Contains(view, "]"), "view should contain HP bar with brackets")
}

func TestMainSheetModelDeathSaves(t *testing.T) {
	char := createTestCharacter()
	char.CombatStats.HitPoints.Current = 0 // At 0 HP

	model := NewMainSheetModel(char, nil)
	model.focusArea = FocusCombat

	// Test adding death save success
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	assert.Equal(t, 1, char.CombatStats.DeathSaves.Successes, "expected 1 success")

	// Test adding death save failure
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	assert.Equal(t, 1, char.CombatStats.DeathSaves.Failures, "expected 1 failure")

	// Test reset
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}})
	assert.Equal(t, 0, char.CombatStats.DeathSaves.Successes, "expected death saves successes to be reset")
	assert.Equal(t, 0, char.CombatStats.DeathSaves.Failures, "expected death saves failures to be reset")
}

func TestMainSheetModelDeathSavesStabilized(t *testing.T) {
	char := createTestCharacter()
	char.CombatStats.HitPoints.Current = 0
	char.CombatStats.DeathSaves.Successes = 3

	model := NewMainSheetModel(char, nil)
	model.width = 120
	model.height = 40

	view := model.View()

	assert.Contains(t, view, "STABILIZED", "view should show STABILIZED when 3 successes")
}

func TestMainSheetModelConditionAdd(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)
	model.focusArea = FocusCombat

	// Enter condition add mode
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	assert.True(t, model.conditionMode && model.conditionAdding, "expected to be in condition add mode")

	// Select and add first condition
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Len(t, char.CombatStats.Conditions, 1, "expected 1 condition")
	assert.False(t, model.conditionMode, "expected to exit condition mode after adding")
}

func TestMainSheetModelConditionRemove(t *testing.T) {
	char := createTestCharacter()
	char.CombatStats.AddCondition(models.ConditionPoisoned)
	char.CombatStats.AddCondition(models.ConditionFrightened)

	model := NewMainSheetModel(char, nil)
	model.focusArea = FocusCombat

	// Enter condition remove mode
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}})
	assert.True(t, model.conditionMode && !model.conditionAdding, "expected to be in condition remove mode")

	// Remove first condition
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Len(t, char.CombatStats.Conditions, 1, "expected 1 condition after removal")
}

func TestMainSheetModelConditionNavigation(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)
	model.focusArea = FocusCombat

	// Enter condition add mode
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})

	// Navigate down
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, model.conditionCursor, "expected cursor at 1")

	// Navigate up
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 0, model.conditionCursor, "expected cursor at 0")

	// Cancel
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	assert.False(t, model.conditionMode, "expected to exit condition mode on escape")
}

func TestMainSheetModelWeaponAttacks(t *testing.T) {
	char := createTestCharacter()
	// Add a weapon to inventory
	sword := models.NewItem("sword-1", "Longsword", models.ItemTypeWeapon)
	sword.Damage = "1d8"
	sword.DamageType = domain.DamageSlashing
	char.Inventory.AddItem(sword)

	// Equip the weapon to main hand
	char.Inventory.Equipment.MainHand = &sword

	model := NewMainSheetModel(char, nil)
	model.width = 120
	model.height = 40

	view := model.View()

	// Check weapon is displayed
	assert.Contains(t, view, "Attacks", "view should contain Attacks section when weapons present")
	assert.Contains(t, view, "Longsword", "view should contain weapon name")
	assert.Contains(t, view, "1d8", "view should contain weapon damage")
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
	sword.DamageType = domain.DamageSlashing
	sword.SubCategory = "Martial Melee Weapons"
	char.Inventory.AddItem(sword)

	model := NewMainSheetModel(char, nil)
	bonus := model.getWeaponAttackBonus(sword)
	// STR (+3) + Prof (+3) = +6
	assert.Equal(t, 6, bonus, "expected attack bonus 6")

	// Test finesse weapon (uses better of STR/DEX)
	dagger := models.NewItem("dagger-1", "Dagger", models.ItemTypeWeapon)
	dagger.Damage = "1d4"
	dagger.DamageType = domain.DamagePiercing
	dagger.SubCategory = "Simple Melee Weapons"
	dagger.WeaponProps = []domain.WeaponProperty{domain.PropertyFinesse, domain.PropertyLight, domain.PropertyThrown}

	bonus = model.getWeaponAttackBonus(dagger)
	// Better of STR (+3) or DEX (+2) = STR (+3) + Prof (+3) = +6
	assert.Equal(t, 6, bonus, "expected finesse attack bonus 6")

	// Test with higher DEX
	char.AbilityScores.Dexterity.Base = 18 // +4 modifier
	bonus = model.getWeaponAttackBonus(dagger)
	// Better of STR (+3) or DEX (+4) = DEX (+4) + Prof (+3) = +7
	assert.Equal(t, 7, bonus, "expected finesse attack bonus 7 with higher DEX")
}

func TestRestModeTransitions(t *testing.T) {
	char := createTestCharacter()
	model := NewMainSheetModel(char, nil)

	// Initially not in rest mode
	assert.Equal(t, RestModeNone, model.restMode, "expected initial rest mode to be RestModeNone")

	// Press 'r' to enter rest menu
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	assert.Equal(t, RestModeMenu, model.restMode, "expected rest mode RestModeMenu after 'r'")

	// Press 's' to go to short rest
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	assert.Equal(t, RestModeShort, model.restMode, "expected rest mode RestModeShort after 's'")

	// Press Esc to go back to menu
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	assert.Equal(t, RestModeMenu, model.restMode, "expected rest mode RestModeMenu after Esc")

	// Press 'l' to go to long rest
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	assert.Equal(t, RestModeLong, model.restMode, "expected rest mode RestModeLong after 'l'")

	// Press Esc to go back to menu
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	assert.Equal(t, RestModeMenu, model.restMode, "expected rest mode RestModeMenu after Esc")

	// Press Esc again to exit rest mode
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	assert.Equal(t, RestModeNone, model.restMode, "expected rest mode RestModeNone after second Esc")
}

func TestShortRestHitDiceAdjustment(t *testing.T) {
	char := createTestCharacter()
	char.CombatStats.HitDice.Remaining = 3
	model := NewMainSheetModel(char, nil)

	// Enter short rest mode
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})

	assert.Equal(t, 0, model.restHitDice, "expected initial restHitDice 0")

	// Increase hit dice to spend
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 1, model.restHitDice, "expected restHitDice 1 after up")

	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 2, model.restHitDice, "expected restHitDice 2 after second up")

	// Decrease hit dice
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, model.restHitDice, "expected restHitDice 1 after down")

	// Can't go below 0
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 0, model.restHitDice, "expected restHitDice 0 (can't go negative)")

	// Can't exceed remaining hit dice
	model.restHitDice = 3
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 3, model.restHitDice, "expected restHitDice 3 (can't exceed remaining)")
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

	// Now we should be at the roll/average prompt
	assert.True(t, model.restRollPrompt, "expected roll prompt to be shown")

	// Choose average
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	// After confirming, rest transitions to result screen first
	assert.Equal(t, RestModeResult, model.restMode, "expected rest mode RestModeResult after confirming")

	// Result should contain rest summary
	assert.Contains(t, model.restResult, "SHORT REST COMPLETE", "expected rest result about short rest")

	// HP should have increased (average roll: 5+1=6, +2 CON = 8 per die, 2 dice = 16)
	assert.Greater(t, char.CombatStats.HitPoints.Current, 20, "expected HP to increase after short rest")

	// Hit dice should have decreased
	assert.Equal(t, 3, char.CombatStats.HitDice.Remaining, "expected 3 remaining hit dice")

	// Dismiss the result screen
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.Equal(t, RestModeNone, model.restMode, "expected rest mode RestModeNone after dismissing result")
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

	// After confirming, rest transitions to result screen first
	assert.Equal(t, RestModeResult, model.restMode, "expected rest mode RestModeResult after confirming")

	// Result should contain rest summary
	assert.Contains(t, model.restResult, "LONG REST COMPLETE", "expected rest result about long rest")

	// HP should be at maximum
	assert.Equal(t, 45, char.CombatStats.HitPoints.Current, "expected HP to be max (45)")

	// Hit dice should have recovered (half level, min 1 = 2)
	assert.GreaterOrEqual(t, char.CombatStats.HitDice.Remaining, 4, "expected at least 4 hit dice after recovery")

	// Dismiss the result screen
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.Equal(t, RestModeNone, model.restMode, "expected rest mode RestModeNone after dismissing result")
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
	assert.Contains(t, view, "Rest Options", "expected rest menu to show 'Rest Options'")
	assert.Contains(t, view, "Short Rest", "expected rest menu to show 'Short Rest'")
	assert.Contains(t, view, "Long Rest", "expected rest menu to show 'Long Rest'")

	// Test short rest screen
	model.restMode = RestModeShort
	view = model.View()
	assert.Contains(t, view, "Short Rest", "expected short rest screen to show 'Short Rest'")
	assert.Contains(t, view, "Hit Dice", "expected short rest screen to show 'Hit Dice'")

	// Test long rest screen
	model.restMode = RestModeLong
	view = model.View()
	assert.Contains(t, view, "Long Rest", "expected long rest screen to show 'Long Rest'")
}

func TestMainSheetCompactLayout(t *testing.T) {
	char := models.NewCharacter("test-1", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage(t.TempDir())

	m := NewMainSheetModel(char, store)

	// Simulate a narrow terminal
	sizeMsg := tea.WindowSizeMsg{Width: 70, Height: 30}
	m, _ = m.Update(sizeMsg)

	view := m.View()

	// In compact mode, all panels should still render
	assert.Contains(t, view, "Abilities", "Compact view should contain abilities")
	assert.Contains(t, view, "Skills", "Compact view should contain skills")
	// View should be non-empty and not panic
	assert.True(t, len(view) > 0, "Compact view should render content")
}

func TestMainSheetWideLayout(t *testing.T) {
	char := models.NewCharacter("test-1", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage(t.TempDir())

	m := NewMainSheetModel(char, store)

	// Simulate a wide terminal
	sizeMsg := tea.WindowSizeMsg{Width: 140, Height: 40}
	m, _ = m.Update(sizeMsg)

	view := m.View()

	// Wide view should render normally
	assert.Contains(t, view, "Abilities", "Wide view should contain abilities")
	assert.Contains(t, view, "Skills", "Wide view should contain skills")
	assert.True(t, len(view) > 0, "Wide view should render content")
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
