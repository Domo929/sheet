package views

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newFighterWizard builds a Fighter 5 / Wizard 3 multiclass character with a
// populated class breakdown and qualifying ability scores.
func newFighterWizard(t *testing.T) *models.Character {
	t.Helper()
	char := models.NewCharacter("mc-lvl", "Multi", "Human", "Fighter")
	char.AbilityScores = models.NewAbilityScoresFromValues(15, 13, 14, 15, 12, 10)
	char.Classes = []models.ClassLevel{
		{Class: "Fighter", Level: 5, HitDie: 10},
		{Class: "Wizard", Level: 3, HitDie: 6},
	}
	char.SyncPrimaryClass() // Info.Class=Fighter, Info.Level=8
	char.CombatStats.HitDice.DieType = 10
	char.CombatStats.HitDice.Total = 8
	char.CombatStats.HitDice.Remaining = 8
	char.CombatStats.HitPoints.Maximum = 50
	char.CombatStats.HitPoints.Current = 50
	return char
}

func TestLevelUpMulticlass_AddsClassStep(t *testing.T) {
	char := newFighterWizard(t)
	m := NewLevelUpModel(char, newTestStore(t), newTestLoader())

	require.True(t, m.multiclass, "should detect multiclass character")
	require.NotEmpty(t, m.steps)
	assert.Equal(t, LevelUpStepClass, m.steps[0], "class selection should be the first step")
	assert.Equal(t, LevelUpStepClass, m.currentStep)
	// Defaults to advancing the primary class.
	assert.Equal(t, "Fighter", m.levelClass)
	assert.Equal(t, 5, m.oldLevel)
	assert.Equal(t, 6, m.newLevel)
}

func TestLevelUpMulticlass_SelectSecondaryClass(t *testing.T) {
	char := newFighterWizard(t)
	m := NewLevelUpModel(char, newTestStore(t), newTestLoader())

	m = pressKey(m, tea.KeyDown)  // move cursor to Wizard
	m = pressKey(m, tea.KeyEnter) // select it

	assert.Equal(t, "Wizard", m.levelClass)
	assert.Equal(t, 3, m.oldLevel)
	assert.Equal(t, 4, m.newLevel)
	assert.NotEqual(t, LevelUpStepClass, m.currentStep, "should advance past the class step")
}

func TestLevelUpMulticlass_ApplyCasterClass(t *testing.T) {
	char := newFighterWizard(t)
	m := NewLevelUpModel(char, newTestStore(t), newTestLoader())

	m.applyChosenClass(1) // Wizard
	m.stagedHPIncrease = 4
	m.applyLevelUp()

	assert.Equal(t, 4, char.Classes[1].Level, "Wizard should advance to level 4")
	assert.Equal(t, 5, char.Classes[0].Level, "Fighter level should be unchanged")
	assert.Equal(t, 9, char.TotalLevel())
	assert.Equal(t, 9, char.Info.Level, "Info.Level should track total level")
	assert.Equal(t, "Fighter", char.Info.Class, "primary class should stay Fighter")

	// Combined caster level 4 (Wizard 4, Fighter contributes 0) -> slots 4/3.
	require.NotNil(t, char.Spellcasting)
	assert.Equal(t, 4, char.Spellcasting.SpellSlots.GetSlot(1).Total)
	assert.Equal(t, 3, char.Spellcasting.SpellSlots.GetSlot(2).Total)

	assert.Equal(t, 54, char.CombatStats.HitPoints.Maximum, "HP should increase by staged amount")
	assert.Equal(t, 9, char.CombatStats.HitDice.Total, "hit dice total should equal total level")
}

func TestLevelUpMulticlass_ApplyMartialClassLeavesSlots(t *testing.T) {
	char := newFighterWizard(t)
	m := NewLevelUpModel(char, newTestStore(t), newTestLoader())

	// Default selection is the primary (Fighter).
	m.stagedHPIncrease = 6
	m.applyLevelUp()

	assert.Equal(t, 6, char.Classes[0].Level, "Fighter should advance to level 6")
	assert.Equal(t, 9, char.TotalLevel())

	// Caster level still 3 (Wizard 3) -> slots 4/2 unchanged.
	require.NotNil(t, char.Spellcasting)
	assert.Equal(t, 4, char.Spellcasting.SpellSlots.GetSlot(1).Total)
	assert.Equal(t, 2, char.Spellcasting.SpellSlots.GetSlot(2).Total)
}

func TestLevelUpMulticlass_SubclassRecordedOnEntry(t *testing.T) {
	char := newFighterWizard(t)
	m := NewLevelUpModel(char, newTestStore(t), newTestLoader())

	m.applyChosenClass(1) // Wizard
	m.stagedSubclass = &data.Subclass{Name: "Evoker"}
	m.stagedHPIncrease = 3
	m.applyLevelUp()

	assert.Equal(t, "Evoker", char.Classes[1].Subclass, "subclass should be recorded on the Wizard entry")
	assert.Equal(t, "", char.Classes[0].Subclass, "Fighter entry should be untouched")
}

func TestLevelUpMulticlass_SingleClassUnaffected(t *testing.T) {
	char := newTestFighter() // level-1 Fighter, no multiclass breakdown
	m := NewLevelUpModel(char, newTestStore(t), newTestLoader())

	assert.False(t, m.multiclass)
	require.NotEmpty(t, m.steps)
	assert.NotEqual(t, LevelUpStepClass, m.steps[0], "single-class flow should not add a class step")
	assert.Equal(t, LevelUpStepHP, m.currentStep)
}

func TestLevelUpMulticlass_ClassViewRenders(t *testing.T) {
	char := newFighterWizard(t)
	m := NewLevelUpModel(char, newTestStore(t), newTestLoader())

	out := m.View()
	assert.Contains(t, out, "Choose a Class to Advance")
	assert.Contains(t, out, "Fighter")
	assert.Contains(t, out, "Wizard")
}
