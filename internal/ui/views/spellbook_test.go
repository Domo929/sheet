package views

import (
	"testing"

	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSpellbookModel_Init(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Character", "Human", "Wizard")
	char.Spellcasting = &models.Spellcasting{
		Ability:        models.AbilityIntelligence,
		SpellSlots:     models.NewSpellSlots(),
		KnownSpells:    []models.KnownSpell{},
		PreparesSpells: true,
		MaxPrepared:    5,
	}

	store, _ := storage.NewCharacterStorage("")
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)

	cmd := model.Init()
	if cmd == nil {
		t.Error("Expected Init to return a command")
	}
}

func TestSpellbookModel_PrepareSpell(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Character", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:        models.AbilityIntelligence,
		SpellSlots:     models.NewSpellSlots(),
		KnownSpells:    []models.KnownSpell{},
		PreparesSpells: true,
		MaxPrepared:    2,
	}
	sc.AddSpell("Fireball", 3)
	sc.AddSpell("Magic Missile", 1)
	char.Spellcasting = sc

	store, _ := storage.NewCharacterStorage("")
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)

	// Test preparing a spell
	model = model.handlePrepareToggle()

	// Check if a spell was prepared (the flat list sorts by level then name)
	flatSpells := model.getFlatSpellList()
	if len(flatSpells) == 0 || !flatSpells[0].Prepared {
		t.Error("Expected first spell in sorted list to be prepared")
	}

	// Test unpreparing
	model = model.handlePrepareToggle()

	flatSpells = model.getFlatSpellList()
	if len(flatSpells) == 0 || flatSpells[0].Prepared {
		t.Error("Expected first spell in sorted list to be unprepared")
	}
}

func TestSpellbookModel_CastSpell(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Character", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:        models.AbilityIntelligence,
		SpellSlots:     models.NewSpellSlots(),
		KnownSpells:    []models.KnownSpell{},
		PreparesSpells: true,
		MaxPrepared:    5,
	}
	sc.SpellSlots.SetSlots(1, 3) // 3 level 1 slots
	sc.AddSpell("Magic Missile", 1)
	sc.PrepareSpell("Magic Missile", true)
	char.Spellcasting = sc

	store, _ := storage.NewCharacterStorage("")
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)

	// Check initial slots
	if sc.SpellSlots.Level1.Remaining != 3 {
		t.Errorf("Expected 3 level 1 slots, got %d", sc.SpellSlots.Level1.Remaining)
	}

	// Cast spell
	model = model.handleCastSpell()

	// Check slots after casting
	if sc.SpellSlots.Level1.Remaining != 2 {
		t.Errorf("Expected 2 level 1 slots after casting, got %d", sc.SpellSlots.Level1.Remaining)
	}
}

func TestSpellbookModel_FilterLevel(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Character", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:     models.AbilityIntelligence,
		SpellSlots:  models.NewSpellSlots(),
		KnownSpells: []models.KnownSpell{},
	}
	sc.AddSpell("Fireball", 3)
	sc.AddSpell("Magic Missile", 1)
	sc.AddSpell("Shield", 1)
	char.Spellcasting = sc

	store, _ := storage.NewCharacterStorage("")
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)

	// Initially all spells should be visible
	flatSpells := model.getFlatSpellList()
	if len(flatSpells) != 3 {
		t.Errorf("Expected 3 spells, got %d", len(flatSpells))
	}

	// Filter to level 1
	model.filterLevel = 1
	flatSpells = model.getFlatSpellList()
	if len(flatSpells) != 2 {
		t.Errorf("Expected 2 level 1 spells, got %d", len(flatSpells))
	}

	// Filter to level 3
	model.filterLevel = 3
	flatSpells = model.getFlatSpellList()
	if len(flatSpells) != 1 {
		t.Errorf("Expected 1 level 3 spell, got %d", len(flatSpells))
	}
}

func TestSpellbookModel_Navigation(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Character", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:     models.AbilityIntelligence,
		SpellSlots:  models.NewSpellSlots(),
		KnownSpells: []models.KnownSpell{},
	}
	sc.AddSpell("Fireball", 3)
	sc.AddSpell("Magic Missile", 1)
	sc.AddSpell("Shield", 1)
	char.Spellcasting = sc

	store, _ := storage.NewCharacterStorage("")
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)
	model.height = 20 // Set a height for scrolling calculations

	// Initially at first spell
	if model.spellCursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", model.spellCursor)
	}

	// Move down
	model = model.handleDown()
	if model.spellCursor != 1 {
		t.Errorf("Expected cursor at 1, got %d", model.spellCursor)
	}

	// Move down again
	model = model.handleDown()
	if model.spellCursor != 2 {
		t.Errorf("Expected cursor at 2, got %d", model.spellCursor)
	}

	// Try to move down past end (should stay at 2)
	model = model.handleDown()
	if model.spellCursor != 2 {
		t.Errorf("Expected cursor to stay at 2, got %d", model.spellCursor)
	}

	// Move up
	model = model.handleUp()
	if model.spellCursor != 1 {
		t.Errorf("Expected cursor at 1, got %d", model.spellCursor)
	}

	// Move up again
	model = model.handleUp()
	if model.spellCursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", model.spellCursor)
	}

	// Try to move up past beginning (should stay at 0)
	model = model.handleUp()
	if model.spellCursor != 0 {
		t.Errorf("Expected cursor to stay at 0, got %d", model.spellCursor)
	}
}

func TestSpellbookModel_RemoveSpell(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Character", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:     models.AbilityIntelligence,
		SpellSlots:  models.NewSpellSlots(),
		KnownSpells: []models.KnownSpell{},
	}
	sc.AddSpell("Fireball", 3)
	sc.AddSpell("Magic Missile", 1)
	char.Spellcasting = sc

	store, _ := storage.NewCharacterStorage("")
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)

	// Initially 2 spells
	if len(sc.KnownSpells) != 2 {
		t.Errorf("Expected 2 spells, got %d", len(sc.KnownSpells))
	}

	// Remove first spell
	model.removeSpellName = sc.KnownSpells[0].Name
	model.performRemoveSpell()

	// Now should have 1 spell
	if len(sc.KnownSpells) != 1 {
		t.Errorf("Expected 1 spell after removal, got %d", len(sc.KnownSpells))
	}
}

func TestSpellbookModel_WindowResize(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Character", "Human", "Wizard")
	char.Spellcasting = &models.Spellcasting{
		Ability:     models.AbilityIntelligence,
		SpellSlots:  models.NewSpellSlots(),
		KnownSpells: []models.KnownSpell{},
	}

	store, _ := storage.NewCharacterStorage("")
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)

	// Send window size message
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(msg)

	if updatedModel.width != 100 {
		t.Errorf("Expected width 100, got %d", updatedModel.width)
	}
	if updatedModel.height != 50 {
		t.Errorf("Expected height 50, got %d", updatedModel.height)
	}
}

func TestSpellbookModel_PactMagic(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Warlock", "Human", "Warlock")
	sc := &models.Spellcasting{
		Ability:     models.AbilityCharisma,
		SpellSlots:  models.NewSpellSlots(),
		KnownSpells: []models.KnownSpell{},
		PactMagic:   &models.PactMagic{SlotLevel: 3, Total: 2, Remaining: 2},
	}
	sc.AddCantrip("Eldritch Blast")
	sc.AddSpell("Hex", 1)
	char.Spellcasting = sc

	store, _ := storage.NewCharacterStorage("")
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)

	// Cast Hex using pact magic (Hex is the only spell, so it's at index 0)
	model.spellCursor = 0 // Select Hex in the flat spell list
	initialPactSlots := sc.PactMagic.Remaining

	model = model.handleCastSpell()

	// Pact magic slot should be used
	if sc.PactMagic.Remaining != initialPactSlots-1 {
		t.Errorf("Expected pact magic slot to be used, had %d, now has %d", initialPactSlots, sc.PactMagic.Remaining)
	}
}
