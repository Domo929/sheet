package views

import (
	"testing"

	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/storage"
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
		if score != 10 {
			t.Errorf("Expected initial score of 10, got %d for ability %d", score, i)
		}
	}

	// Test increment
	model.focusedAbility = 0 // STR
	model.incrementAbility()
	if model.abilityScores[0] != 11 {
		t.Errorf("Expected STR to be 11 after increment, got %d", model.abilityScores[0])
	}

	// Test decrement
	model.decrementAbility()
	if model.abilityScores[0] != 10 {
		t.Errorf("Expected STR to be 10 after decrement, got %d", model.abilityScores[0])
	}

	// Test upper bound (20)
	model.abilityScores[0] = 20
	model.incrementAbility()
	if model.abilityScores[0] != 20 {
		t.Errorf("Expected STR to stay at 20 (max), got %d", model.abilityScores[0])
	}

	// Test lower bound (3)
	model.abilityScores[0] = 3
	model.decrementAbility()
	if model.abilityScores[0] != 3 {
		t.Errorf("Expected STR to stay at 3 (min), got %d", model.abilityScores[0])
	}
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
		if score != 8 {
			t.Errorf("Expected initial score of 8, got %d for ability %d", score, i)
		}
	}

	// Initial cost should be 0 (all at 8)
	cost := model.calculateCurrentPointBuy()
	if cost != 0 {
		t.Errorf("Expected initial cost of 0, got %d", cost)
	}

	// Increment one ability
	model.focusedAbility = 0
	model.incrementAbility()
	if model.abilityScores[0] != 9 {
		t.Errorf("Expected STR to be 9, got %d", model.abilityScores[0])
	}

	// Check cost increased
	cost = model.calculateCurrentPointBuy()
	if cost != 1 {
		t.Errorf("Expected cost of 1, got %d", cost)
	}

	// Test upper bound (15)
	model.abilityScores[0] = 15
	model.incrementAbility()
	if model.abilityScores[0] != 15 {
		t.Errorf("Expected STR to stay at 15 (max), got %d", model.abilityScores[0])
	}

	// Test lower bound (8)
	model.abilityScores[0] = 8
	model.decrementAbility()
	if model.abilityScores[0] != 8 {
		t.Errorf("Expected STR to stay at 8 (min), got %d", model.abilityScores[0])
	}
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
		if score != 0 {
			t.Errorf("Expected initial score of 0, got %d for ability %d", score, i)
		}
	}

	// Standard array should be [15, 14, 13, 12, 10, 8]
	expected := []int{15, 14, 13, 12, 10, 8}
	if len(model.standardArrayValues) != len(expected) {
		t.Errorf("Expected %d standard array values, got %d", len(expected), len(model.standardArrayValues))
	}
	for i, val := range expected {
		if model.standardArrayValues[i] != val {
			t.Errorf("Expected standard array[%d] to be %d, got %d", i, val, model.standardArrayValues[i])
		}
	}

	// First increment from 0 should assign lowest value (8)
	model.focusedAbility = 0
	model.incrementAbility() // Should assign 8
	if model.abilityScores[0] != 8 {
		t.Errorf("Expected first increment to assign 8 (lowest), got %d", model.abilityScores[0])
	}

	// Next increment should go to next lowest (10)
	model.incrementAbility()
	if model.abilityScores[0] != 10 {
		t.Errorf("Expected second increment to assign 10, got %d", model.abilityScores[0])
	}

	// Check that values are marked as used
	usedCount := 0
	for _, used := range model.standardArrayUsed {
		if used {
			usedCount++
		}
	}
	if usedCount != 1 {
		t.Errorf("Expected 1 standard array value to be used, got %d", usedCount)
	}
}

func TestValidateAbilityScoresManual(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	loader := data.NewLoader("../../../data")
	model := NewCharacterCreationModel(store, loader)

	model.abilityScoreMode = AbilityModeManual
	
	// Valid scores
	model.abilityScores = [6]int{10, 12, 14, 8, 16, 9}
	if !model.validateAbilityScores() {
		t.Error("Expected valid scores to pass validation")
	}

	// Invalid: score too high
	model.abilityScores = [6]int{21, 12, 14, 8, 16, 9}
	if model.validateAbilityScores() {
		t.Error("Expected score of 21 to fail validation")
	}

	// Invalid: score too low
	model.abilityScores = [6]int{2, 12, 14, 8, 16, 9}
	if model.validateAbilityScores() {
		t.Error("Expected score of 2 to fail validation")
	}

	// Invalid: unset score (0)
	model.abilityScores = [6]int{0, 12, 14, 8, 16, 9}
	if model.validateAbilityScores() {
		t.Error("Expected unset score (0) to fail validation")
	}
}

func TestValidateAbilityScoresPointBuy(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	loader := data.NewLoader("../../../data")
	model := NewCharacterCreationModel(store, loader)

	model.abilityScoreMode = AbilityModePointBuy
	
	// Valid: all 8s (0 points)
	model.abilityScores = [6]int{8, 8, 8, 8, 8, 8}
	if !model.validateAbilityScores() {
		t.Error("Expected all 8s to pass validation")
	}

	// Valid: 27 points exactly
	// 15, 15, 15, 8, 8, 8 = 9+9+9 = 27
	model.abilityScores = [6]int{15, 15, 15, 8, 8, 8}
	if !model.validateAbilityScores() {
		t.Error("Expected 27 points to pass validation")
	}

	// Invalid: over 27 points
	// 15, 15, 15, 15, 8, 8 = 9+9+9+9 = 36
	model.abilityScores = [6]int{15, 15, 15, 15, 8, 8}
	if model.validateAbilityScores() {
		t.Error("Expected over 27 points to fail validation")
	}
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

	// Enter allocation mode
	model.allocatingBonuses = true
	model.focusedBonusSlot = 0
	model.abilitySelections = []string{}
	model.abilityBonusAmounts = []int{}

	// Allocate +2 to int
	model.incrementBackgroundBonus()
	model.incrementBackgroundBonus()

	if len(model.abilitySelections) != 1 {
		t.Fatalf("Expected 1 ability selection, got %d", len(model.abilitySelections))
	}
	if model.abilitySelections[0] != "int" {
		t.Errorf("Expected 'int', got %s", model.abilitySelections[0])
	}
	if model.abilityBonusAmounts[0] != 2 {
		t.Errorf("Expected +2 bonus, got %d", model.abilityBonusAmounts[0])
	}

	// Move to next ability (wis) and allocate +1
	model.focusedBonusSlot = 1
	model.incrementBackgroundBonus()

	if len(model.abilitySelections) != 2 {
		t.Fatalf("Expected 2 ability selections, got %d", len(model.abilitySelections))
	}
	if model.abilitySelections[1] != "wis" {
		t.Errorf("Expected 'wis', got %s", model.abilitySelections[1])
	}
	if model.abilityBonusAmounts[1] != 1 {
		t.Errorf("Expected +1 bonus, got %d", model.abilityBonusAmounts[1])
	}

	// Verify total is 3
	total := 0
	for _, amt := range model.abilityBonusAmounts {
		total += amt
	}
	if total != 3 {
		t.Errorf("Expected total of 3 points, got %d", total)
	}

	// Test decrement
	model.focusedBonusSlot = 0 // Back to int
	model.decrementBackgroundBonus()
	if model.abilityBonusAmounts[0] != 1 {
		t.Errorf("Expected +1 after decrement, got %d", model.abilityBonusAmounts[0])
	}
}

func TestAbilityScoreModeToggle(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	loader := data.NewLoader("../../../data")
	model := NewCharacterCreationModel(store, loader)

	// Start with default (Point Buy)
	if model.abilityScoreMode != AbilityModePointBuy {
		t.Errorf("Expected default mode to be PointBuy, got %v", model.abilityScoreMode)
	}

	// Cycle: Point Buy -> Manual
	model.abilityScoreMode = AbilityModePointBuy
	nextMode := AbilityModeManual
	model.abilityScoreMode = nextMode
	model.resetAbilityScores()
	
	if model.abilityScoreMode != AbilityModeManual {
		t.Errorf("Expected Manual mode, got %v", model.abilityScoreMode)
	}
	if model.abilityScores[0] != 10 {
		t.Errorf("Expected score of 10 in Manual mode, got %d", model.abilityScores[0])
	}

	// Cycle: Manual -> Standard Array
	nextMode = AbilityModeStandardArray
	model.abilityScoreMode = nextMode
	model.resetAbilityScores()
	
	if model.abilityScoreMode != AbilityModeStandardArray {
		t.Errorf("Expected StandardArray mode, got %v", model.abilityScoreMode)
	}
	if model.abilityScores[0] != 0 {
		t.Errorf("Expected score of 0 in StandardArray mode, got %d", model.abilityScores[0])
	}

	// Cycle: Standard Array -> Point Buy
	nextMode = AbilityModePointBuy
	model.abilityScoreMode = nextMode
	model.resetAbilityScores()
	
	if model.abilityScoreMode != AbilityModePointBuy {
		t.Errorf("Expected PointBuy mode, got %v", model.abilityScoreMode)
	}
	if model.abilityScores[0] != 8 {
		t.Errorf("Expected score of 8 in PointBuy mode, got %d", model.abilityScores[0])
	}
}
