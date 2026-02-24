package views

import (
	"testing"

	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpellbookModel_ModeSwitch(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Wizard", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:        models.AbilityIntelligence,
		SpellSlots:     models.NewSpellSlots(),
		KnownSpells:    []models.KnownSpell{},
		PreparesSpells: true,
		MaxPrepared:    5,
	}
	sc.AddSpell("Magic Missile", 1)
	sc.PrepareSpell("Magic Missile", true)
	char.Spellcasting = sc

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)

	// Should start in spell list mode
	assert.Equal(t, ModeSpellList, model.mode, "Should start in spell list mode")

	// Should show only prepared spells in spell list mode
	displaySpells := model.getDisplaySpells()
	assert.Len(t, displaySpells, 1, "Should show 1 prepared spell in spell list mode")

	// Switch to preparation mode
	model.mode = ModePreparation

	// Should show all spells in preparation mode
	displaySpells = model.getDisplaySpells()
	assert.Len(t, displaySpells, 1, "Should show 1 spell in preparation mode")
}

func TestSpellbookModel_DisplayOnlyPreparedInSpellListMode(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Wizard", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:        models.AbilityIntelligence,
		SpellSlots:     models.NewSpellSlots(),
		KnownSpells:    []models.KnownSpell{},
		PreparesSpells: true,
		MaxPrepared:    5,
	}
	sc.AddSpell("Magic Missile", 1)
	sc.AddSpell("Fireball", 3)
	sc.AddSpell("Shield", 1)
	sc.PrepareSpell("Magic Missile", true)
	sc.PrepareSpell("Shield", true)
	// Fireball is not prepared
	char.Spellcasting = sc

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)
	model.mode = ModeSpellList

	// Should only show 2 prepared spells
	displaySpells := model.getDisplaySpells()
	assert.Len(t, displaySpells, 2, "Should show 2 prepared spells")

	// Switch to preparation mode
	model.mode = ModePreparation

	// Should show all 3 spells
	displaySpells = model.getDisplaySpells()
	assert.Len(t, displaySpells, 3, "Should show all 3 spells in preparation mode")
}

func TestSpellbookModel_SpellSelectionAndRemoval(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Wizard", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:        models.AbilityIntelligence,
		SpellSlots:     models.NewSpellSlots(),
		KnownSpells:    []models.KnownSpell{},
		PreparesSpells: true,
		MaxPrepared:    5,
	}
	sc.AddSpell("Magic Missile", 1)
	sc.AddSpell("Invisibility", 2)
	sc.AddSpell("Misty Step", 2)
	char.Spellcasting = sc

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)
	model.mode = ModePreparation

	// Get display spells (should be sorted)
	displaySpells := model.getDisplaySpells()

	t.Logf("Display spell order:")
	for i, spell := range displaySpells {
		t.Logf("  [%d] %s (Level %d)", i, spell.Name, spell.Level)
	}

	// Verify order: Magic Missile (L1), then Invisibility (L2), then Misty Step (L2)
	require.Len(t, displaySpells, 3, "Should have 3 spells")

	assert.Equal(t, "Magic Missile", displaySpells[0].Name, "First spell should be Magic Missile")
	assert.Equal(t, "Invisibility", displaySpells[1].Name, "Second spell should be Invisibility")
	assert.Equal(t, "Misty Step", displaySpells[2].Name, "Third spell should be Misty Step")

	// Select Misty Step (index 2)
	model.spellCursor = 2
	model.removeSpellName = displaySpells[model.spellCursor].Name

	assert.Equal(t, "Misty Step", model.removeSpellName, "Should be removing Misty Step")

	// Perform removal
	model.performRemoveSpell()

	// Should now have 2 spells
	assert.Len(t, sc.KnownSpells, 2, "Should have 2 spells after removal")

	// Verify Misty Step is gone
	for _, spell := range sc.KnownSpells {
		assert.NotEqual(t, "Misty Step", spell.Name, "Misty Step should have been removed")
	}
}

func TestSpellbookModel_WindowResize(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Wizard", "Human", "Wizard")
	char.Spellcasting = &models.Spellcasting{
		Ability:     models.AbilityIntelligence,
		SpellSlots:  models.NewSpellSlots(),
		KnownSpells: []models.KnownSpell{},
	}

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)

	// Send window size message
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(msg)

	assert.Equal(t, 100, updatedModel.width, "Width should be 100")
	assert.Equal(t, 50, updatedModel.height, "Height should be 50")
}

func TestSpellbookModel_PrepareSpellInPreparationMode(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Wizard", "Human", "Wizard")
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

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)
	model.mode = ModePreparation

	// Test preparing a spell
	model = model.handlePrepareToggle()

	displaySpells := model.getDisplaySpells()
	assert.True(t, displaySpells[0].Prepared, "First spell should be prepared")

	// Test unpreparing
	model = model.handlePrepareToggle()

	displaySpells = model.getDisplaySpells()
	assert.False(t, displaySpells[0].Prepared, "First spell should be unprepared")
}

func TestSpellbookModel_CastPreparedSpell(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Wizard", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:        models.AbilityIntelligence,
		SpellSlots:     models.NewSpellSlots(),
		KnownSpells:    []models.KnownSpell{},
		PreparesSpells: true,
		MaxPrepared:    5,
	}
	sc.SpellSlots.SetSlots(1, 4) // 4 level 1 slots
	sc.AddSpell("Magic Missile", 1)
	sc.PrepareSpell("Magic Missile", true)
	char.Spellcasting = sc

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)
	model.mode = ModeSpellList // Spell list mode for casting

	// Check initial slots
	assert.Equal(t, 4, sc.SpellSlots.Level1.Remaining, "Should have 4 level 1 slots initially")

	// Try to cast the spell
	model = model.handleCastSpell()

	// Should enter confirmation modal mode
	assert.Equal(t, ModeConfirmCast, model.mode, "Should enter confirmation modal mode")
	assert.NotNil(t, model.castingSpell, "Should have casting spell set")
	assert.Equal(t, "Magic Missile", model.castingSpell.Name, "Should be casting Magic Missile")

	// Simulate Enter key to confirm cast
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	// Should return to spell list mode
	assert.Equal(t, ModeSpellList, model.mode, "Should return to spell list mode")
	assert.Contains(t, model.statusMessage, "Cast Magic Missile", "Should show cast message")

	// Spell slot should be consumed
	slot := sc.SpellSlots.GetSlot(1)
	assert.Equal(t, 3, slot.Remaining, "Should have consumed one level 1 slot")
}

func TestSpellbookModel_CancelCasting(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Wizard", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:        models.AbilityIntelligence,
		SpellSlots:     models.NewSpellSlots(),
		KnownSpells:    []models.KnownSpell{},
		PreparesSpells: true,
		MaxPrepared:    5,
	}
	sc.SpellSlots.SetSlots(1, 4)
	sc.AddSpell("Magic Missile", 1)
	sc.PrepareSpell("Magic Missile", true)
	char.Spellcasting = sc

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)
	model.mode = ModeSpellList

	// Start casting
	model = model.handleCastSpell()
	assert.Equal(t, ModeConfirmCast, model.mode)

	// Press Esc to cancel
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	// Should return to spell list
	assert.Equal(t, ModeSpellList, model.mode)
	assert.Contains(t, model.statusMessage, "cancelled", "Should show cancellation message")
	assert.Nil(t, model.castingSpell, "Should clear casting spell")

	// Spell slot should NOT be consumed
	slot := sc.SpellSlots.GetSlot(1)
	assert.Equal(t, 4, slot.Remaining, "Should not consume slot when cancelled")
}

func TestSpellbookModel_CastCantrip(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Wizard", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:        models.AbilityIntelligence,
		SpellSlots:     models.NewSpellSlots(),
		KnownSpells:    []models.KnownSpell{},
		CantripsKnown:  []string{"Fire Bolt"},
		PreparesSpells: true,
		MaxPrepared:    5,
	}
	char.Spellcasting = sc

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)
	model.mode = ModeSpellList

	// Start casting cantrip
	model = model.handleCastSpell()
	assert.Equal(t, ModeConfirmCast, model.mode)
	assert.Equal(t, 0, len(model.availableCastLevels), "Cantrips should have no slot levels")

	// Confirm cast
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	// Should cast without consuming resources
	assert.Equal(t, ModeSpellList, model.mode)
	assert.Contains(t, model.statusMessage, "no slot required")
}

func TestSpellbookModel_UpcastSpell(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Wizard", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:        models.AbilityIntelligence,
		SpellSlots:     models.NewSpellSlots(),
		KnownSpells:    []models.KnownSpell{},
		PreparesSpells: true,
		MaxPrepared:    5,
	}
	sc.SpellSlots.SetSlots(1, 0) // No level 1 slots
	sc.SpellSlots.SetSlots(2, 2) // 2 level 2 slots
	sc.SpellSlots.SetSlots(3, 1) // 1 level 3 slot
	sc.AddSpell("Magic Missile", 1)
	sc.PrepareSpell("Magic Missile", true)
	char.Spellcasting = sc

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)
	model.mode = ModeSpellList

	// Try to cast - should enter confirmation modal mode
	model = model.handleCastSpell()

	// Should be in confirmation modal mode
	assert.Equal(t, ModeConfirmCast, model.mode, "Should be in confirmation modal mode")

	// Should have 2 available levels (2 and 3)
	require.Len(t, model.availableCastLevels, 2, "Should have 2 available cast levels")

	assert.Equal(t, 2, model.availableCastLevels[0], "First available level should be 2")
	assert.Equal(t, 3, model.availableCastLevels[1], "Second available level should be 3")

	// Navigate to level 3 and cast
	model.castLevelCursor = 1 // Select level 3
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	// Check that level 3 slot was used
	assert.Equal(t, 0, sc.SpellSlots.Level3.Remaining, "Should have 0 level 3 slots after upcasting")

	// Level 2 slots should be untouched
	assert.Equal(t, 2, sc.SpellSlots.Level2.Remaining, "Should still have 2 level 2 slots")

	// Should be back in spell list mode
	assert.Equal(t, ModeSpellList, model.mode, "Should be back in spell list mode after casting")
}

func TestSpellbookModel_GetAvailableCastLevels(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Wizard", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:     models.AbilityIntelligence,
		SpellSlots:  models.NewSpellSlots(),
		KnownSpells: []models.KnownSpell{},
	}
	sc.SpellSlots.SetSlots(1, 0) // No level 1 slots
	sc.SpellSlots.SetSlots(2, 1) // 1 level 2 slot
	sc.SpellSlots.SetSlots(3, 0) // No level 3 slots
	sc.SpellSlots.SetSlots(4, 2) // 2 level 4 slots
	char.Spellcasting = sc

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)

	// For a level 1 spell, should return levels 2 and 4
	available := model.getAvailableCastLevels(1)
	require.Len(t, available, 2, "Should have 2 available levels for level 1 spell")
	assert.Equal(t, []int{2, 4}, available, "Should have levels 2 and 4 available")

	// For a level 3 spell, should return only level 4
	available = model.getAvailableCastLevels(3)
	require.Len(t, available, 1, "Should have 1 available level for level 3 spell")
	assert.Equal(t, []int{4}, available, "Should have level 4 available")

	// For a level 5 spell, should return nothing
	available = model.getAvailableCastLevels(5)
	assert.Empty(t, available, "Should have no available levels for level 5 spell")
}

func TestSpellbookModel_CantripsAreSelectable(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Wizard", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:        models.AbilityIntelligence,
		SpellSlots:     models.NewSpellSlots(),
		KnownSpells:    []models.KnownSpell{},
		PreparesSpells: true,
		MaxPrepared:    5,
	}
	sc.AddCantrip("Fire Bolt")
	sc.AddCantrip("Mage Hand")
	sc.AddSpell("Magic Missile", 1)
	sc.PrepareSpell("Magic Missile", true)
	char.Spellcasting = sc

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)
	model.mode = ModeSpellList

	// Get display spells - should include cantrips
	displaySpells := model.getDisplaySpells()

	// Should have 3 spells total (2 cantrips + 1 prepared spell)
	require.Len(t, displaySpells, 3, "Should have 2 cantrips and 1 prepared spell")

	// First two should be cantrips (level 0)
	assert.Equal(t, 0, displaySpells[0].Level, "First spell should be a cantrip")
	assert.Equal(t, 0, displaySpells[1].Level, "Second spell should be a cantrip")
	assert.Equal(t, "Fire Bolt", displaySpells[0].Name, "First cantrip should be Fire Bolt")
	assert.Equal(t, "Mage Hand", displaySpells[1].Name, "Second cantrip should be Mage Hand")

	// Third should be Magic Missile
	assert.Equal(t, 1, displaySpells[2].Level, "Third spell should be level 1")
	assert.Equal(t, "Magic Missile", displaySpells[2].Name, "Third spell should be Magic Missile")

	// Test navigation to cantrips
	model.spellCursor = 0
	assert.Equal(t, "Fire Bolt", displaySpells[model.spellCursor].Name, "Cursor 0 should select Fire Bolt")

	model.spellCursor = 1
	assert.Equal(t, "Mage Hand", displaySpells[model.spellCursor].Name, "Cursor 1 should select Mage Hand")

	// Test casting a cantrip
	model.spellCursor = 0
	model = model.handleCastSpell()

	// Should enter confirmation modal
	assert.Equal(t, ModeConfirmCast, model.mode, "Should enter confirmation modal")
	assert.Equal(t, "Fire Bolt", model.castingSpell.Name, "Should be casting Fire Bolt")

	// Confirm cast
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	// Should return to spell list with success message
	assert.Equal(t, ModeSpellList, model.mode, "Should return to spell list")
	assert.Contains(t, model.statusMessage, "Fire Bolt", "Should cast Fire Bolt")
	assert.Contains(t, model.statusMessage, "no slot required", "Cantrips don't use slots")
}

func TestSpellbookModel_UpcastInformation(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Wizard", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:        models.AbilityIntelligence,
		SpellSlots:     models.NewSpellSlots(),
		KnownSpells:    []models.KnownSpell{},
		PreparesSpells: true,
		MaxPrepared:    5,
	}
	sc.SpellSlots.SetSlots(1, 0) // No level 1 slots
	sc.SpellSlots.SetSlots(2, 2) // 2 level 2 slots
	sc.SpellSlots.SetSlots(3, 1) // 1 level 3 slot
	sc.AddSpell("Burning Hands", 1)
	sc.PrepareSpell("Burning Hands", true)
	char.Spellcasting = sc

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)
	model.mode = ModeSpellList

	// Manually set up spell data to test upcast calculation
	model.selectedSpellData = &data.SpellData{
		Name:       "Burning Hands",
		Level:      1,
		School:     "Evocation",
		Damage:     "3d6",
		DamageType: "fire",
		Upcast:     "+1d6 per slot level",
	}

	// Get display spells
	displaySpells := model.getDisplaySpells()
	require.Len(t, displaySpells, 1, "Should have 1 spell")
	assert.Equal(t, "Burning Hands", displaySpells[0].Name)

	// Test upcast effect calculation
	model.castingSpell = &displaySpells[0]

	// Calculate at base level (level 1)
	effect := model.calculateUpcastEffect(1)
	assert.Contains(t, effect, "3d6", "Should show base damage at level 1")

	// Calculate upcast at level 2 (1 level above base)
	effect = model.calculateUpcastEffect(2)
	assert.Contains(t, effect, "4d6", "Should show 4d6 total for 1 level upcast (3+1)")
	assert.Contains(t, effect, "damage", "Should mention damage")

	// Calculate upcast at level 3 (2 levels above base)
	effect = model.calculateUpcastEffect(3)
	assert.Contains(t, effect, "5d6", "Should show 5d6 total for 2 level upcast (3+2)")
	assert.Contains(t, effect, "damage", "Should mention damage")
}

func TestSpellbookModel_MagicMissileUpcast(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Wizard", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:        models.AbilityIntelligence,
		SpellSlots:     models.NewSpellSlots(),
		KnownSpells:    []models.KnownSpell{},
		PreparesSpells: true,
		MaxPrepared:    5,
	}
	sc.SpellSlots.SetSlots(1, 2)
	sc.SpellSlots.SetSlots(2, 1)
	sc.AddSpell("Magic Missile", 1)
	sc.PrepareSpell("Magic Missile", true)
	char.Spellcasting = sc

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)
	model.mode = ModeSpellList

	// Manually set up Magic Missile spell data
	model.selectedSpellData = &data.SpellData{
		Name:       "Magic Missile",
		Level:      1,
		School:     "Evocation",
		Damage:     "1d4+1",
		DamageType: "force",
		Upcast:     "+1 dart per slot level",
	}

	displaySpells := model.getDisplaySpells()
	require.Len(t, displaySpells, 1, "Should have 1 spell")
	model.castingSpell = &displaySpells[0]

	// Level 1: 3 darts base
	effect := model.calculateUpcastEffect(1)
	assert.Contains(t, effect, "3 darts", "Should show 3 darts at level 1")
	assert.Contains(t, effect, "1d4+1", "Should show damage per dart")

	// Level 2: 4 darts (3 + 1)
	effect = model.calculateUpcastEffect(2)
	assert.Contains(t, effect, "4 darts", "Should show 4 darts at level 2")
	assert.Contains(t, effect, "1d4+1", "Should show damage per dart")

	// Level 3: 5 darts (3 + 2)
	effect = model.calculateUpcastEffect(3)
	assert.Contains(t, effect, "5 darts", "Should show 5 darts at level 3")
	assert.Contains(t, effect, "1d4+1", "Should show damage per dart")
}

func TestSpellbookModel_CastRitualSpell(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Wizard", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:                models.AbilityIntelligence,
		SpellSlots:             models.NewSpellSlots(),
		KnownSpells:            []models.KnownSpell{},
		PreparesSpells:         true,
		MaxPrepared:            5,
		RitualCasterUnprepared: true, // Wizards can cast rituals without preparing
	}
	sc.SpellSlots.SetSlots(1, 2) // 2 level 1 slots
	sc.AddSpell("Detect Magic", 1)
	sc.KnownSpells[0].Ritual = true // Mark as ritual
	char.Spellcasting = sc

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)
	model.mode = ModeSpellList

	// Start casting ritual spell
	model = model.handleCastSpell()
	assert.Equal(t, ModeConfirmCast, model.mode)
	// Ritual spells should offer both: ritual (0 sentinel) and slot options
	require.GreaterOrEqual(t, len(model.availableCastLevels), 1, "Should have at least ritual option")
	assert.Equal(t, 0, model.availableCastLevels[0], "First option should be ritual (sentinel 0)")
	assert.Equal(t, 2, len(model.availableCastLevels), "Should have ritual option + level 1 slot")

	// Cursor defaults to 0 (ritual option), confirm cast
	assert.Equal(t, 0, model.castLevelCursor)
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	// Should cast as ritual without consuming spell slots
	assert.Equal(t, ModeSpellList, model.mode)
	assert.Contains(t, model.statusMessage, "as ritual")
	assert.Contains(t, model.statusMessage, "no slot required")
	assert.Contains(t, model.statusMessage, "10 extra minutes")

	// Spell slot should NOT be consumed
	slot := sc.SpellSlots.GetSlot(1)
	assert.Equal(t, 2, slot.Remaining, "Should not consume slot when casting as ritual")
}

func TestSpellbookModel_CastRitualSpellWithSlot(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Wizard", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:                models.AbilityIntelligence,
		SpellSlots:             models.NewSpellSlots(),
		KnownSpells:            []models.KnownSpell{},
		PreparesSpells:         true,
		MaxPrepared:            5,
		RitualCasterUnprepared: true,
	}
	sc.SpellSlots.SetSlots(1, 2) // 2 level 1 slots
	sc.AddSpell("Detect Magic", 1)
	sc.KnownSpells[0].Ritual = true
	char.Spellcasting = sc

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)
	model.mode = ModeSpellList

	// Start casting ritual spell
	model = model.handleCastSpell()
	assert.Equal(t, ModeConfirmCast, model.mode)

	// Move cursor to slot option (index 1 = level 1 slot)
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.Equal(t, 1, model.castLevelCursor)

	// Confirm cast with spell slot
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	// Should cast using a spell slot
	assert.Equal(t, ModeSpellList, model.mode)
	assert.Contains(t, model.statusMessage, "Detect Magic")
	assert.Contains(t, model.statusMessage, "level 1 slot")

	// Spell slot SHOULD be consumed
	slot := sc.SpellSlots.GetSlot(1)
	assert.Equal(t, 1, slot.Remaining, "Should consume slot when casting with slot")
}

func TestSpellbookCompactLayout(t *testing.T) {
	char := models.NewCharacter("test-1", "Test", "Elf", "Wizard")
	sc := models.NewSpellcasting(models.AbilityIntelligence)
	char.Spellcasting = &sc

	store, err := storage.NewCharacterStorage(t.TempDir())
	require.NoError(t, err)
	loader := data.NewLoader("../../../data")

	model := NewSpellbookModel(char, store, loader)

	// Load spell database so View() doesn't return early
	db, loadErr := loader.GetSpells()
	require.NoError(t, loadErr)
	model.spellDatabase = db

	// Simulate a narrow terminal (below compact breakpoint of 90)
	sizeMsg := tea.WindowSizeMsg{Width: 70, Height: 30}
	model, _ = model.Update(sizeMsg)

	view := model.View()

	// Should render without panic
	assert.True(t, len(view) > 0, "Compact spellbook should render")
	assert.Contains(t, view, "Spellbook", "Should contain spellbook header")
}

func TestSpellbookStandardLayout(t *testing.T) {
	char := models.NewCharacter("test-2", "Test", "Elf", "Wizard")
	sc := models.NewSpellcasting(models.AbilityIntelligence)
	char.Spellcasting = &sc

	store, err := storage.NewCharacterStorage(t.TempDir())
	require.NoError(t, err)
	loader := data.NewLoader("../../../data")

	model := NewSpellbookModel(char, store, loader)

	// Load spell database
	db, loadErr := loader.GetSpells()
	require.NoError(t, loadErr)
	model.spellDatabase = db

	// Simulate a wide terminal (above compact breakpoint of 90)
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 30}
	model, _ = model.Update(sizeMsg)

	view := model.View()

	// Should render without panic and include spell slots panel
	assert.True(t, len(view) > 0, "Standard spellbook should render")
	assert.Contains(t, view, "Spellbook", "Should contain spellbook header")
	assert.Contains(t, view, "Spell Slots", "Standard layout should include spell slots panel")
}

func TestSpellbookCompactLayoutHidesSlots(t *testing.T) {
	char := models.NewCharacter("test-3", "Test", "Elf", "Wizard")
	sc := models.NewSpellcasting(models.AbilityIntelligence)
	sc.SpellSlots.SetSlots(1, 4) // Add some slots so "Spell Slots" would appear
	char.Spellcasting = &sc

	store, err := storage.NewCharacterStorage(t.TempDir())
	require.NoError(t, err)
	loader := data.NewLoader("../../../data")

	model := NewSpellbookModel(char, store, loader)

	// Load spell database
	db, loadErr := loader.GetSpells()
	require.NoError(t, loadErr)
	model.spellDatabase = db

	// Simulate a narrow terminal
	sizeMsg := tea.WindowSizeMsg{Width: 70, Height: 30}
	model, _ = model.Update(sizeMsg)

	view := model.View()

	// Compact mode should hide the spell slots panel
	assert.NotContains(t, view, "Spell Slots", "Compact layout should hide spell slots panel")
}

func TestSpellbookCompactWithRollHistory(t *testing.T) {
	char := models.NewCharacter("test-4", "Test", "Elf", "Wizard")
	sc := models.NewSpellcasting(models.AbilityIntelligence)
	char.Spellcasting = &sc

	store, err := storage.NewCharacterStorage(t.TempDir())
	require.NoError(t, err)
	loader := data.NewLoader("../../../data")

	model := NewSpellbookModel(char, store, loader)

	// Load spell database
	db, loadErr := loader.GetSpells()
	require.NoError(t, loadErr)
	model.spellDatabase = db

	// Set a wide terminal but with roll history taking space
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 30}
	model, _ = model.Update(sizeMsg)
	model.SetRollHistoryState(true, 40) // 40 cols for roll history, leaving 80 available

	view := model.View()

	// With roll history visible, available width is 80 which is < 90 breakpoint
	// So it should use compact mode (no spell slots)
	assert.True(t, len(view) > 0, "Should render with roll history")
	assert.NotContains(t, view, "Spell Slots", "Should use compact layout when roll history reduces available width below breakpoint")
}
