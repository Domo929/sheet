package views

import (
	"strings"
	"testing"

	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newTestFighter creates a level-1 Fighter with sensible defaults.
func newTestFighter() *models.Character {
	char := models.NewCharacter("test-id", "Test Fighter", "Human", "Fighter")
	char.Info.Level = 1
	char.CombatStats.HitDice.DieType = 10
	char.CombatStats.HitPoints.Maximum = 12
	char.CombatStats.HitPoints.Current = 12
	return char
}

func newTestStore(t *testing.T) *storage.CharacterStorage {
	t.Helper()
	store, err := storage.NewCharacterStorage(t.TempDir())
	require.NoError(t, err, "failed to create character storage")
	return store
}

func newTestLoader() *data.Loader {
	return data.NewLoader("../../../data")
}

// containsStep returns true if the step slice contains the given step.
func containsStep(steps []LevelUpStep, target LevelUpStep) bool {
	for _, s := range steps {
		if s == target {
			return true
		}
	}
	return false
}

// pressKey sends a key message and returns the updated model.
func pressKey(m *LevelUpModel, keyType tea.KeyType) *LevelUpModel {
	updated, _ := m.Update(tea.KeyMsg{Type: keyType})
	return updated
}

// advancePastHP selects the Average HP method (cursor down then enter)
// and then presses Enter again to move past the HP result screen.
func advancePastHP(m *LevelUpModel) *LevelUpModel {
	// Select Average (cursor down to index 1, then enter to lock in)
	m = pressKey(m, tea.KeyDown)
	m = pressKey(m, tea.KeyEnter)
	// Press enter again to advance past the HP result display
	m = pressKey(m, tea.KeyEnter)
	return m
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestLevelUpModel_NewModel(t *testing.T) {
	char := newTestFighter()
	store := newTestStore(t)
	loader := newTestLoader()

	model := NewLevelUpModel(char, store, loader)

	require.NotNil(t, model, "NewLevelUpModel should not return nil")
	assert.Equal(t, 1, model.oldLevel, "oldLevel should be 1")
	assert.Equal(t, 2, model.newLevel, "newLevel should be 2")
	assert.Equal(t, LevelUpStepHP, model.currentStep, "first step should be HP")
	assert.True(t, containsStep(model.steps, LevelUpStepHP), "steps should contain HP")
	assert.True(t, containsStep(model.steps, LevelUpStepConfirm), "steps should contain Confirm")
	assert.Empty(t, model.errMsg, "errMsg should be empty for valid class")
}

func TestLevelUpModel_HPStepAverage(t *testing.T) {
	char := newTestFighter()
	char.AbilityScores.SetBase(models.AbilityConstitution, 14) // CON mod = +2
	char.CombatStats.HitDice.DieType = 10

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	require.Equal(t, LevelUpStepHP, model.currentStep, "should start on HP step")

	// Navigate cursor down to Average (index 1)
	model = pressKey(model, tea.KeyDown)
	assert.Equal(t, 1, model.hpMethodCursor, "cursor should be at Average (1)")

	// Press enter to confirm Average
	model = pressKey(model, tea.KeyEnter)

	// d10 average = (10/2)+1 = 6, + CON mod 2 = 8
	assert.True(t, model.hpRolled, "hpRolled should be true after confirming")
	assert.Equal(t, 8, model.stagedHPIncrease, "average HP increase for d10 + CON 14 should be 8")
	assert.Equal(t, 6, model.hpRollResult, "raw die result for d10 average should be 6")
}

func TestLevelUpModel_HPStepRoll(t *testing.T) {
	char := newTestFighter()
	char.AbilityScores.SetBase(models.AbilityConstitution, 10) // CON mod = 0
	char.CombatStats.HitDice.DieType = 10

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	require.Equal(t, LevelUpStepHP, model.currentStep)

	// Cursor is at Roll (index 0) by default — press enter to roll
	assert.Equal(t, 0, model.hpMethodCursor, "cursor should default to Roll (0)")
	model = pressKey(model, tea.KeyEnter)

	assert.True(t, model.hpRolled, "hpRolled should be true after rolling")
	assert.GreaterOrEqual(t, model.stagedHPIncrease, 1, "roll result should be >= 1")
	assert.LessOrEqual(t, model.stagedHPIncrease, 10, "roll result should be <= 10 (d10, CON mod 0)")
	assert.Equal(t, model.hpRollResult, model.stagedHPIncrease, "with CON mod 0, staged should equal raw roll")
}

func TestLevelUpModel_HPStepMinimum1(t *testing.T) {
	// Use Wizard (d6 hit die from class data) with very low CON to trigger min 1 clamp
	char := models.NewCharacter("test-wizard-min", "Test Wizard", "Elf", "Wizard")
	char.Info.Level = 1
	// CON 3 → modifier = (3-10-1)/2 = -4 (floor division toward negative infinity)
	char.AbilityScores.SetBase(models.AbilityConstitution, 3)
	char.CombatStats.HitPoints.Maximum = 4
	char.CombatStats.HitPoints.Current = 4

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	// Select Average: d6 average = (6/2)+1 = 4, + CON mod (-4) = 0 → clamped to 1
	model = pressKey(model, tea.KeyDown) // Average
	model = pressKey(model, tea.KeyEnter)

	assert.True(t, model.hpRolled)
	assert.Equal(t, 4, model.hpRollResult, "raw average for d6 should be 4")
	assert.Equal(t, 1, model.stagedHPIncrease, "HP increase should be clamped to minimum 1")
}

func TestLevelUpModel_SubclassStep(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 2 // leveling to 3 = Fighter subclass level
	char.CombatStats.HitDice.DieType = 10
	char.AbilityScores.SetBase(models.AbilityConstitution, 10)

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	require.Equal(t, 3, model.newLevel)
	require.True(t, containsStep(model.steps, LevelUpStepSubclass),
		"steps should contain Subclass when leveling to 3 as Fighter without a subclass")

	// Advance past HP step
	model = advancePastHP(model)
	assert.Equal(t, LevelUpStepSubclass, model.currentStep, "should be on Subclass step after HP")

	// Navigate the subclass list down one
	model = pressKey(model, tea.KeyDown)
	assert.Equal(t, 1, model.subclassCursor, "subclass cursor should be 1 after pressing down")

	// Select subclass
	model = pressKey(model, tea.KeyEnter)
	require.NotNil(t, model.stagedSubclass, "stagedSubclass should be set after selection")
	assert.NotEmpty(t, model.stagedSubclass.Name, "subclass name should not be empty")
}

func TestLevelUpModel_NoSubclassIfAlreadyHas(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 6 // leveling to 7
	char.Info.Subclass = "Champion"
	char.CombatStats.HitDice.DieType = 10

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	assert.Equal(t, 7, model.newLevel)
	assert.False(t, containsStep(model.steps, LevelUpStepSubclass),
		"steps should NOT contain Subclass when character already has one")
}

func TestLevelUpModel_ASIStep(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 3 // leveling to 4 = Fighter ASI level
	char.CombatStats.HitDice.DieType = 10

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	assert.Equal(t, 4, model.newLevel)
	assert.True(t, containsStep(model.steps, LevelUpStepASI),
		"steps should contain ASI when leveling to 4 as Fighter")
}

func TestLevelUpModel_ASIPlusTwo(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 3 // leveling to 4
	char.Info.Subclass = "Champion"
	char.CombatStats.HitDice.DieType = 10
	char.AbilityScores.SetBase(models.AbilityConstitution, 10)
	char.AbilityScores.SetBase(models.AbilityStrength, 14)

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	require.True(t, containsStep(model.steps, LevelUpStepASI))

	// Navigate to ASI step by advancing past previous steps
	for model.currentStep != LevelUpStepASI {
		if model.currentStep == LevelUpStepHP {
			model = advancePastHP(model)
		} else {
			// Generic advance for other steps (features, etc.)
			model = pressKey(model, tea.KeyEnter)
		}
	}
	require.Equal(t, LevelUpStepASI, model.currentStep, "should be on ASI step")

	// Default ASI pattern is +2 to one ability, default mode is ASI (not feat)
	assert.Equal(t, ASIModeASI, model.asiMode, "default ASI mode should be ASI")
	assert.Equal(t, ASIPatternPlus2, model.asiPattern, "default ASI pattern should be +2")

	// Cursor starts at index 0 (Strength) — press Enter to select +2 STR
	assert.Equal(t, 0, model.asiAbilityCursor, "cursor should start at ability 0 (Strength)")
	model = pressKey(model, tea.KeyEnter)

	assert.Equal(t, 2, model.stagedASIChanges[0], "stagedASIChanges[0] (Strength) should be +2")
	assert.True(t, model.asiConfirmed, "asiConfirmed should be true after selecting an ability")
}

func TestLevelUpModel_CancelDoesNotApply(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 1
	char.CombatStats.HitPoints.Maximum = 12
	char.CombatStats.HitPoints.Current = 12

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	require.Equal(t, LevelUpStepHP, model.currentStep)

	// Press Esc immediately on HP step (not rolled yet — triggers cancel)
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Verify cancel produces a BackToSheetMsg command
	require.NotNil(t, cmd, "Esc on HP step should return a command")
	msg := cmd()
	_, isBack := msg.(BackToSheetMsg)
	assert.True(t, isBack, "command should produce BackToSheetMsg")

	// Character should be unchanged
	assert.Equal(t, 1, char.Info.Level, "character level should remain 1 after cancel")
	assert.Equal(t, 12, char.CombatStats.HitPoints.Maximum, "HP max should remain 12 after cancel")
}

func TestLevelUpModel_ConfirmAppliesChanges(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 1
	char.CombatStats.HitPoints.Maximum = 12
	char.CombatStats.HitPoints.Current = 12
	char.AbilityScores.SetBase(models.AbilityConstitution, 14) // CON mod +2
	char.CombatStats.HitDice.DieType = 10

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	// Select Average HP: d10 average=6 + CON mod 2 = 8
	model = advancePastHP(model)

	// Navigate through any remaining steps until we reach Confirm
	for model.currentStep != LevelUpStepConfirm {
		model = pressKey(model, tea.KeyEnter)
	}
	require.Equal(t, LevelUpStepConfirm, model.currentStep, "should be on Confirm step")

	// Confirm
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Verify a LevelUpCompleteMsg was produced
	require.NotNil(t, cmd, "confirming level-up should return a command")
	msg := cmd()
	_, isComplete := msg.(LevelUpCompleteMsg)
	assert.True(t, isComplete, "command should produce LevelUpCompleteMsg")

	// Verify character was updated
	assert.Equal(t, 2, char.Info.Level, "character level should be 2 after confirm")
	assert.Equal(t, 20, char.CombatStats.HitPoints.Maximum,
		"HP max should be 12 + 8 = 20 after confirm")
	assert.Equal(t, 20, char.CombatStats.HitPoints.Current,
		"HP current should also increase by 8")
	assert.Equal(t, 2, char.CombatStats.HitDice.Total,
		"hit dice total should match new level")
}

func TestLevelUpModel_MaxLevelHandled(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 20

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	// Model should still be created (newLevel=21 is set, but LevelUp will fail)
	require.NotNil(t, model)
	assert.Equal(t, 20, model.oldLevel)
	assert.Equal(t, 21, model.newLevel)

	// The wizard should still have at least a Confirm step
	assert.True(t, len(model.steps) > 0, "steps should not be empty even at max level")
}

func TestLevelUpModel_MilestoneNoXPCheck(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 1
	char.Info.ProgressionType = models.ProgressionMilestone
	char.Info.ExperiencePoints = 0

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	require.NotNil(t, model)
	assert.Empty(t, model.errMsg, "should not have an error for milestone progression")
	assert.Equal(t, LevelUpStepHP, model.currentStep,
		"wizard should start at HP step regardless of XP for milestone characters")
}

func TestLevelUpModel_StepsForNonCaster(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 1 // leveling to 2

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	assert.False(t, containsStep(model.steps, LevelUpStepSpellSlots),
		"Fighter (non-caster) should NOT have SpellSlots step")
}

func TestLevelUpModel_StepsForCaster(t *testing.T) {
	// Create a Wizard — spellcaster with changing slots at level 2
	char := models.NewCharacter("test-wizard", "Test Wizard", "Elf", "Wizard")
	char.Info.Level = 1
	char.CombatStats.HitDice.DieType = 6

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	assert.Equal(t, 2, model.newLevel)
	assert.True(t, containsStep(model.steps, LevelUpStepSpellSlots),
		"Wizard (caster) should have SpellSlots step when leveling to 2")
}

func TestLevelUpModel_ViewRenders(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 1

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	view := model.View()
	assert.NotEmpty(t, view, "View() should return a non-empty string")
	assert.True(t, strings.Contains(view, "Level Up"),
		"View() should contain 'Level Up'")
}

func TestLevelUpModel_HPStepReroll(t *testing.T) {
	char := newTestFighter()
	char.AbilityScores.SetBase(models.AbilityConstitution, 10)
	char.CombatStats.HitDice.DieType = 10

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	// Roll HP
	model = pressKey(model, tea.KeyEnter)
	assert.True(t, model.hpRolled, "should be rolled after enter")
	firstResult := model.stagedHPIncrease

	// Press Esc to re-roll (goes back to method selection)
	model = pressKey(model, tea.KeyEsc)
	assert.False(t, model.hpRolled, "hpRolled should be false after Esc to re-roll")
	assert.Equal(t, 0, model.stagedHPIncrease, "stagedHPIncrease should reset to 0")

	// Roll again
	model = pressKey(model, tea.KeyEnter)
	assert.True(t, model.hpRolled, "should be rolled again")
	// Note: firstResult may equal the new result by chance, so just verify it's valid
	_ = firstResult
	assert.GreaterOrEqual(t, model.stagedHPIncrease, 1)
	assert.LessOrEqual(t, model.stagedHPIncrease, 10)
}

func TestLevelUpModel_SubclassStepBackRetreat(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 2 // leveling to 3
	char.AbilityScores.SetBase(models.AbilityConstitution, 10)
	char.CombatStats.HitDice.DieType = 10

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	require.True(t, containsStep(model.steps, LevelUpStepSubclass))

	// Get past HP
	model = advancePastHP(model)
	require.Equal(t, LevelUpStepSubclass, model.currentStep)

	// Press Esc on subclass step — should go back to HP
	model = pressKey(model, tea.KeyEsc)
	assert.Equal(t, LevelUpStepHP, model.currentStep,
		"Esc on Subclass should retreat to HP step")
}

func TestLevelUpModel_ASIPlusOnePlusOne(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 3 // leveling to 4
	char.Info.Subclass = "Champion"
	char.AbilityScores.SetBase(models.AbilityConstitution, 10)
	char.AbilityScores.SetBase(models.AbilityStrength, 14)
	char.AbilityScores.SetBase(models.AbilityDexterity, 12)
	char.CombatStats.HitDice.DieType = 10

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	// Navigate to ASI step
	for model.currentStep != LevelUpStepASI {
		if model.currentStep == LevelUpStepHP {
			model = advancePastHP(model)
		} else {
			model = pressKey(model, tea.KeyEnter)
		}
	}
	require.Equal(t, LevelUpStepASI, model.currentStep)

	// Switch to +1/+1 pattern (press Right)
	model = pressKey(model, tea.KeyRight)
	assert.Equal(t, ASIPatternPlus1Plus1, model.asiPattern, "should be +1/+1 pattern after Right")

	// Select Strength (index 0)
	model = pressKey(model, tea.KeyEnter)
	assert.Equal(t, 1, model.stagedASIChanges[0], "STR should get +1")
	assert.False(t, model.asiConfirmed, "should not be confirmed with only 1 selection")

	// Navigate down to Dexterity (index 1) and select
	model = pressKey(model, tea.KeyDown)
	model = pressKey(model, tea.KeyEnter)
	assert.Equal(t, 1, model.stagedASIChanges[1], "DEX should get +1")
	assert.True(t, model.asiConfirmed, "should be confirmed after selecting 2 abilities")
}

func TestLevelUpModel_ASITabToFeatMode(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 3 // leveling to 4
	char.Info.Subclass = "Champion"
	char.AbilityScores.SetBase(models.AbilityConstitution, 10)
	char.CombatStats.HitDice.DieType = 10

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	// Navigate to ASI step
	for model.currentStep != LevelUpStepASI {
		if model.currentStep == LevelUpStepHP {
			model = advancePastHP(model)
		} else {
			model = pressKey(model, tea.KeyEnter)
		}
	}
	require.Equal(t, LevelUpStepASI, model.currentStep)

	// Default mode should be ASI
	assert.Equal(t, ASIModeASI, model.asiMode)

	// Tab to Feat mode
	model = pressKey(model, tea.KeyTab)
	assert.Equal(t, ASIModeFeat, model.asiMode, "should switch to Feat mode after Tab")

	// Tab back to ASI mode
	model = pressKey(model, tea.KeyTab)
	assert.Equal(t, ASIModeASI, model.asiMode, "should switch back to ASI mode after Tab")
}

func TestLevelUpModel_FeaturesStepAdvance(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 1 // leveling to 2 → gains "Action Surge (one use)"

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	// Level 2 Fighter should have features
	if containsStep(model.steps, LevelUpStepFeatures) {
		// Navigate to Features step
		for model.currentStep != LevelUpStepFeatures {
			if model.currentStep == LevelUpStepHP {
				model = advancePastHP(model)
			} else {
				model = pressKey(model, tea.KeyEnter)
			}
		}
		require.Equal(t, LevelUpStepFeatures, model.currentStep)

		// Press Enter to advance past Features
		model = pressKey(model, tea.KeyEnter)
		assert.NotEqual(t, LevelUpStepFeatures, model.currentStep,
			"should advance past Features step on Enter")
	}
}

func TestLevelUpModel_SpellSlotsStepAdvance(t *testing.T) {
	char := models.NewCharacter("test-wizard", "Test Wizard", "Elf", "Wizard")
	char.Info.Level = 1
	char.CombatStats.HitDice.DieType = 6

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	require.True(t, containsStep(model.steps, LevelUpStepSpellSlots))

	// Navigate to SpellSlots step
	for model.currentStep != LevelUpStepSpellSlots {
		if model.currentStep == LevelUpStepHP {
			model = advancePastHP(model)
		} else {
			model = pressKey(model, tea.KeyEnter)
		}
	}
	require.Equal(t, LevelUpStepSpellSlots, model.currentStep)

	// Verify staged spell slots are populated
	assert.NotNil(t, model.stagedSpellSlots, "stagedSpellSlots should be set for Wizard")

	// Press Enter to advance
	model = pressKey(model, tea.KeyEnter)
	assert.NotEqual(t, LevelUpStepSpellSlots, model.currentStep,
		"should advance past SpellSlots step on Enter")
}

func TestLevelUpModel_WindowSizeMsg(t *testing.T) {
	char := newTestFighter()

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	model, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	assert.Equal(t, 120, model.width, "width should be set from WindowSizeMsg")
	assert.Equal(t, 40, model.height, "height should be set from WindowSizeMsg")
}

func TestLevelUpModel_ConfirmStepBackRetreats(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 1
	char.AbilityScores.SetBase(models.AbilityConstitution, 10)
	char.CombatStats.HitDice.DieType = 10

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	// Navigate to Confirm step
	model = advancePastHP(model)
	for model.currentStep != LevelUpStepConfirm {
		model = pressKey(model, tea.KeyEnter)
	}
	require.Equal(t, LevelUpStepConfirm, model.currentStep)

	// Remember the step index before going back
	prevStepIndex := model.stepIndex

	// Press Esc to go back
	model = pressKey(model, tea.KeyEsc)
	assert.Less(t, model.stepIndex, prevStepIndex,
		"stepIndex should decrease when pressing Esc on Confirm step")
}

func TestLevelUpModel_ConfirmAppliesSubclass(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 2 // leveling to 3
	char.AbilityScores.SetBase(models.AbilityConstitution, 10)
	char.CombatStats.HitDice.DieType = 10
	char.CombatStats.HitPoints.Maximum = 22
	char.CombatStats.HitPoints.Current = 22

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	require.True(t, containsStep(model.steps, LevelUpStepSubclass))

	// Advance past HP
	model = advancePastHP(model)
	require.Equal(t, LevelUpStepSubclass, model.currentStep)

	// Select first subclass (index 0)
	model = pressKey(model, tea.KeyEnter)
	selectedSubclass := model.stagedSubclass
	require.NotNil(t, selectedSubclass)

	// Navigate to Confirm and apply
	for model.currentStep != LevelUpStepConfirm {
		model = pressKey(model, tea.KeyEnter)
	}
	model = pressKey(model, tea.KeyEnter) // Confirm

	assert.Equal(t, 3, char.Info.Level, "level should be 3")
	assert.Equal(t, selectedSubclass.Name, char.Info.Subclass,
		"character subclass should be set to selected subclass")
}

func TestLevelUpModel_ConfirmAppliesASI(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 3 // leveling to 4
	char.Info.Subclass = "Champion"
	char.AbilityScores.SetBase(models.AbilityConstitution, 10)
	char.AbilityScores.SetBase(models.AbilityStrength, 14)
	char.CombatStats.HitDice.DieType = 10
	char.CombatStats.HitPoints.Maximum = 32
	char.CombatStats.HitPoints.Current = 32

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	require.True(t, containsStep(model.steps, LevelUpStepASI))

	// Navigate to ASI step
	for model.currentStep != LevelUpStepASI {
		if model.currentStep == LevelUpStepHP {
			model = advancePastHP(model)
		} else {
			model = pressKey(model, tea.KeyEnter)
		}
	}

	// Select +2 Strength (default pattern is +2, cursor at index 0 = Strength)
	model = pressKey(model, tea.KeyEnter) // Select Strength
	require.True(t, model.asiConfirmed)

	// Advance to Confirm and apply
	model = pressKey(model, tea.KeyEnter) // Advance past ASI
	for model.currentStep != LevelUpStepConfirm {
		model = pressKey(model, tea.KeyEnter)
	}
	model = pressKey(model, tea.KeyEnter) // Confirm

	assert.Equal(t, 4, char.Info.Level, "level should be 4")
	assert.Equal(t, 16, char.AbilityScores.Strength.Base,
		"Strength base should be 14 + 2 = 16 after ASI")
}

func TestLevelUpModel_ViewContainsStepInfo(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 1

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)
	model.width = 80
	model.height = 40

	view := model.View()

	// View should contain level info
	assert.Contains(t, view, "1", "view should reference old level")
	assert.Contains(t, view, "2", "view should reference new level")
	assert.Contains(t, view, "Hit Points", "view should contain the HP step name")
}

func TestLevelUpModel_HPMethodCursorBounds(t *testing.T) {
	char := newTestFighter()
	char.CombatStats.HitDice.DieType = 10

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	// Cursor starts at 0 (Roll)
	assert.Equal(t, 0, model.hpMethodCursor)

	// Press Up when already at 0 — should stay at 0
	model = pressKey(model, tea.KeyUp)
	assert.Equal(t, 0, model.hpMethodCursor, "cursor should not go below 0")

	// Press Down to Average (1)
	model = pressKey(model, tea.KeyDown)
	assert.Equal(t, 1, model.hpMethodCursor)

	// Press Down again — should stay at 1 (only 2 options)
	model = pressKey(model, tea.KeyDown)
	assert.Equal(t, 1, model.hpMethodCursor, "cursor should not exceed 1")
}

func TestLevelUpModel_NoSubclassAtLevel2(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 1 // leveling to 2 — not a subclass level for Fighter

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	assert.False(t, containsStep(model.steps, LevelUpStepSubclass),
		"Fighter leveling to 2 should NOT have Subclass step (subclass is at 3)")
}

func TestLevelUpModel_NoASIAtLevel2(t *testing.T) {
	char := newTestFighter()
	char.Info.Level = 1 // leveling to 2

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	assert.False(t, containsStep(model.steps, LevelUpStepASI),
		"Fighter leveling to 2 should NOT have ASI step (ASI is at 4)")
}

func TestLevelUpModel_InvalidClassError(t *testing.T) {
	char := models.NewCharacter("test-invalid", "Test", "Human", "NonexistentClass")
	char.Info.Level = 1

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	require.NotNil(t, model)
	assert.NotEmpty(t, model.errMsg, "should have an error message for invalid class")
	assert.Equal(t, LevelUpStepConfirm, model.currentStep,
		"should fall back to Confirm step on error")
}

func TestLevelUpModel_CasterConfirmAppliesSpellSlots(t *testing.T) {
	char := models.NewCharacter("test-wizard", "Test Wizard", "Elf", "Wizard")
	char.Info.Level = 1
	char.CombatStats.HitDice.DieType = 6
	char.CombatStats.HitPoints.Maximum = 8
	char.CombatStats.HitPoints.Current = 8
	char.AbilityScores.SetBase(models.AbilityConstitution, 10)

	store := newTestStore(t)
	loader := newTestLoader()
	model := NewLevelUpModel(char, store, loader)

	require.True(t, containsStep(model.steps, LevelUpStepSpellSlots))

	// Navigate through all steps to Confirm
	for model.currentStep != LevelUpStepConfirm {
		if model.currentStep == LevelUpStepHP {
			model = advancePastHP(model)
		} else if model.currentStep == LevelUpStepSubclass {
			// Select first subclass if needed
			model = pressKey(model, tea.KeyEnter)
		} else {
			model = pressKey(model, tea.KeyEnter)
		}
	}

	// Confirm the level up
	model = pressKey(model, tea.KeyEnter)

	assert.Equal(t, 2, char.Info.Level, "Wizard should be level 2")
	require.NotNil(t, char.Spellcasting, "Wizard should have spellcasting after level up")
}
