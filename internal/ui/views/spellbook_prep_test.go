package views

import (
	"testing"

	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestSpellbookPreparedLimitEnforced(t *testing.T) {
	char := models.NewCharacter("w1", "Mage", "Human", "Wizard")
	char.Info.Level = 1 // INT 10 -> MaxPrepared 1
	sc := models.NewSpellcasting(models.AbilityIntelligence)
	sc.AddSpell("Magic Missile", 1)
	sc.AddSpell("Shield", 1)
	char.Spellcasting = &sc

	store, _ := storage.NewCharacterStorage(t.TempDir())
	loader := data.NewLoader("../../../data")
	m := NewSpellbookModel(char, store, loader)

	// The constructor recomputed the prep limit for a Wizard.
	assert.True(t, char.Spellcasting.PreparesSpells)
	assert.Equal(t, 1, char.Spellcasting.MaxPrepared)

	m.mode = ModePreparation

	m.spellCursor = 0
	m.handlePrepareToggle()
	assert.Equal(t, 1, char.Spellcasting.CountPreparedSpells())

	m.spellCursor = 1
	m.handlePrepareToggle()
	assert.Equal(t, 1, char.Spellcasting.CountPreparedSpells(), "should not exceed MaxPrepared")
	assert.Contains(t, m.statusMessage, "Cannot prepare more than 1")
}

func TestSpellbookKnownCasterNotGated(t *testing.T) {
	char := models.NewCharacter("s1", "Sorc", "Human", "Sorcerer")
	sc := models.NewSpellcasting(models.AbilityCharisma)
	char.Spellcasting = &sc

	store, _ := storage.NewCharacterStorage(t.TempDir())
	loader := data.NewLoader("../../../data")
	_ = NewSpellbookModel(char, store, loader)

	assert.False(t, char.Spellcasting.PreparesSpells, "sorcerer should not be a prepared caster")
}
