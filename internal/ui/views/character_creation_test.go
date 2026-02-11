package views

import (
	"testing"

	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestAbilityScoreManualMode(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	loader := data.NewLoader("../../../data")
	model := NewCharacterCreationModel(store, loader)

	// Set to manual mode
	model.abilityScoreMode = AbilityModeManual
	model.resetAbilityScores()

	// Check initial values (should be 10)
	for i, score := range model.abilityScores {
		assert.Equal(t, 10, score, "Expected initial score of 10 for ability %d", i)
	}

	// Test increment
	model.focusedAbility = 0 // STR
	model.incrementAbility()
	assert.Equal(t, 11, model.abilityScores[0], "Expected STR to be 11 after increment")

	// Test decrement
	model.decrementAbility()
	assert.Equal(t, 10, model.abilityScores[0], "Expected STR to be 10 after decrement")

	// Test upper bound (20)
	model.abilityScores[0] = 20
	model.incrementAbility()
	assert.Equal(t, 20, model.abilityScores[0], "Expected STR to stay at 20 (max)")

	// Test lower bound (3)
	model.abilityScores[0] = 3
	model.decrementAbility()
	assert.Equal(t, 3, model.abilityScores[0], "Expected STR to stay at 3 (min)")
}

func TestAbilityScorePointBuyMode(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	loader := data.NewLoader("../../../data")
	model := NewCharacterCreationModel(store, loader)

	// Set to point buy mode
	model.abilityScoreMode = AbilityModePointBuy
	model.resetAbilityScores()

	// Check initial values (should be 8)
	for i, score := range model.abilityScores {
		assert.Equal(t, 8, score, "Expected initial score of 8 for ability %d", i)
	}

	// Initial cost should be 0 (all at 8)
	cost := model.calculateCurrentPointBuy()
	assert.Equal(t, 0, cost, "Expected initial cost of 0")

	// Increment one ability
	model.focusedAbility = 0
	model.incrementAbility()
	assert.Equal(t, 9, model.abilityScores[0], "Expected STR to be 9")

	// Check cost increased
	cost = model.calculateCurrentPointBuy()
	assert.Equal(t, 1, cost, "Expected cost of 1")

	// Test upper bound (15)
	model.abilityScores[0] = 15
	model.incrementAbility()
	assert.Equal(t, 15, model.abilityScores[0], "Expected STR to stay at 15 (max)")

	// Test lower bound (8)
	model.abilityScores[0] = 8
	model.decrementAbility()
	assert.Equal(t, 8, model.abilityScores[0], "Expected STR to stay at 8 (min)")
}

func TestAbilityScoreStandardArrayMode(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	loader := data.NewLoader("../../../data")
	model := NewCharacterCreationModel(store, loader)

	// Set to standard array mode
	model.abilityScoreMode = AbilityModeStandardArray
	model.resetAbilityScores()

	// Check initial values (should be 0)
	for i, score := range model.abilityScores {
		assert.Equal(t, 0, score, "Expected initial score of 0 for ability %d", i)
	}

	// Standard array should be [15, 14, 13, 12, 10, 8]
	expected := []int{15, 14, 13, 12, 10, 8}
	assert.Len(t, model.standardArrayValues, len(expected), "Expected %d standard array values", len(expected))
	for i, val := range expected {
		assert.Equal(t, val, model.standardArrayValues[i], "Expected standard array[%d] to be %d", i, val)
	}

	// First increment from 0 should assign lowest value (8)
	model.focusedAbility = 0
	model.incrementAbility() // Should assign 8
	assert.Equal(t, 8, model.abilityScores[0], "Expected first increment to assign 8 (lowest)")

	// Next increment should go to next lowest (10)
	model.incrementAbility()
	assert.Equal(t, 10, model.abilityScores[0], "Expected second increment to assign 10")

	// Check that values are marked as used
	usedCount := 0
	for _, used := range model.standardArrayUsed {
		if used {
			usedCount++
		}
	}
	assert.Equal(t, 1, usedCount, "Expected 1 standard array value to be used")
}

func TestValidateAbilityScoresManual(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	loader := data.NewLoader("../../../data")
	model := NewCharacterCreationModel(store, loader)

	model.abilityScoreMode = AbilityModeManual

	// Valid scores
	model.abilityScores = [6]int{10, 12, 14, 8, 16, 9}
	assert.True(t, model.validateAbilityScores(), "Expected valid scores to pass validation")

	// Invalid: score too high
	model.abilityScores = [6]int{21, 12, 14, 8, 16, 9}
	assert.False(t, model.validateAbilityScores(), "Expected score of 21 to fail validation")

	// Invalid: score too low
	model.abilityScores = [6]int{2, 12, 14, 8, 16, 9}
	assert.False(t, model.validateAbilityScores(), "Expected score of 2 to fail validation")

	// Invalid: unset score (0)
	model.abilityScores = [6]int{0, 12, 14, 8, 16, 9}
	assert.False(t, model.validateAbilityScores(), "Expected unset score (0) to fail validation")
}

func TestValidateAbilityScoresPointBuy(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	loader := data.NewLoader("../../../data")
	model := NewCharacterCreationModel(store, loader)

	model.abilityScoreMode = AbilityModePointBuy

	// Valid: all 8s (0 points)
	model.abilityScores = [6]int{8, 8, 8, 8, 8, 8}
	assert.True(t, model.validateAbilityScores(), "Expected all 8s to pass validation")

	// Valid: 27 points exactly
	// 15, 15, 15, 8, 8, 8 = 9+9+9 = 27
	model.abilityScores = [6]int{15, 15, 15, 8, 8, 8}
	assert.True(t, model.validateAbilityScores(), "Expected 27 points to pass validation")

	// Invalid: over 27 points
	// 15, 15, 15, 15, 8, 8 = 9+9+9+9 = 36
	model.abilityScores = [6]int{15, 15, 15, 15, 8, 8}
	assert.False(t, model.validateAbilityScores(), "Expected over 27 points to fail validation")
}

func TestBackgroundBonusAllocation(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	loader := data.NewLoader("../../../data")
	model := NewCharacterCreationModel(store, loader)

	// Mock background with ability score options
	model.selectedBackground = &data.Background{
		Name:        "Test Background",
		Description: "Test",
		AbilityScores: data.AbilityScoreBonus{
			Options: []string{"int", "wis", "cha"},
			Points:  3,
		},
	}

	// Test +2/+1 pattern
	model.backgroundBonusPattern = 0
	model.backgroundBonus2Target = 0 // int
	model.backgroundBonus1Target = 1 // wis
	model.backgroundBonusComplete = true

	bonuses := model.getBackgroundBonuses()

	// Check int got +2 (index 3)
	assert.Equal(t, 2, bonuses[3], "Expected +2 to INT")

	// Check wis got +1 (index 4)
	assert.Equal(t, 1, bonuses[4], "Expected +1 to WIS")

	// Test +1/+1/+1 pattern
	model.backgroundBonusPattern = 1
	model.backgroundBonusComplete = true

	bonuses = model.getBackgroundBonuses()

	// Check first 3 options got +1 each
	assert.Equal(t, 1, bonuses[3], "Expected +1 to INT")
	assert.Equal(t, 1, bonuses[4], "Expected +1 to WIS")
	assert.Equal(t, 1, bonuses[5], "Expected +1 to CHA")
}

func TestAbilityScoreModeToggle(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	loader := data.NewLoader("../../../data")
	model := NewCharacterCreationModel(store, loader)

	// Start with default (Point Buy)
	assert.Equal(t, AbilityModePointBuy, model.abilityScoreMode, "Expected default mode to be PointBuy")

	// Cycle: Point Buy -> Manual
	model.abilityScoreMode = AbilityModePointBuy
	nextMode := AbilityModeManual
	model.abilityScoreMode = nextMode
	model.resetAbilityScores()

	assert.Equal(t, AbilityModeManual, model.abilityScoreMode, "Expected Manual mode")
	assert.Equal(t, 10, model.abilityScores[0], "Expected score of 10 in Manual mode")

	// Cycle: Manual -> Standard Array
	nextMode = AbilityModeStandardArray
	model.abilityScoreMode = nextMode
	model.resetAbilityScores()

	assert.Equal(t, AbilityModeStandardArray, model.abilityScoreMode, "Expected StandardArray mode")
	assert.Equal(t, 0, model.abilityScores[0], "Expected score of 0 in StandardArray mode")

	// Cycle: Standard Array -> Point Buy
	nextMode = AbilityModePointBuy
	model.abilityScoreMode = nextMode
	model.resetAbilityScores()

	assert.Equal(t, AbilityModePointBuy, model.abilityScoreMode, "Expected PointBuy mode")
	assert.Equal(t, 8, model.abilityScores[0], "Expected score of 8 in PointBuy mode")
}
