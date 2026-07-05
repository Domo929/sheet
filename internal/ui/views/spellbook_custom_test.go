package views

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/stretchr/testify/assert"
)

func newCustomSpellTestModel(t *testing.T) (*SpellbookModel, *models.Character) {
	t.Helper()
	char := models.NewCharacter("cs-1", "Mage", "Human", "Wizard")
	sc := models.NewSpellcasting(models.AbilityIntelligence)
	char.Spellcasting = &sc
	store, _ := storage.NewCharacterStorage(t.TempDir())
	loader := data.NewLoader("../../../data")
	m := NewSpellbookModel(char, store, loader)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 150, Height: 40})
	return m, char
}

func csType(m *SpellbookModel, s string) *SpellbookModel {
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

func csEnter(m *SpellbookModel) *SpellbookModel {
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	return m
}

func TestSpellbookCreateCustomSpell(t *testing.T) {
	m, char := newCustomSpellTestModel(t)

	// Enter homebrew mode via the 'C' key.
	m, _ = m.Update(tea.KeyPressMsg{Code: 'C', Text: "C"})
	assert.Equal(t, ModeCustomSpell, m.mode)

	m = csEnter(csType(m, "Arcane Zap")) // name
	m = csEnter(csType(m, "2"))          // level
	m = csEnter(csType(m, "Evocation"))  // school
	m = csEnter(csType(m, "Action"))     // casting time
	m = csEnter(csType(m, "60 feet"))    // range
	m = csEnter(csType(m, "V, S"))       // components
	m = csEnter(csType(m, "Instant"))    // duration
	m = csEnter(csType(m, "2d8"))        // damage
	m = csEnter(csType(m, "force"))      // damage type
	m = csEnter(csType(m, "Dexterity"))  // saving throw
	m = csEnter(csType(m, "Zaps a foe")) // description -> create

	assert.Equal(t, ModeSpellList, m.mode, "should return to list after creation")

	sc := char.Spellcasting
	assert.True(t, sc.IsCustomSpell("Arcane Zap"))
	cs := sc.FindCustomSpell("Arcane Zap")
	if assert.NotNil(t, cs) {
		assert.Equal(t, 2, cs.Level)
		assert.Equal(t, "Evocation", cs.School)
		assert.Equal(t, "2d8", cs.Damage)
		assert.Equal(t, "force", cs.DamageType)
		assert.Equal(t, "V, S", cs.Components)
	}

	// Known spell should be present and marked custom.
	found := false
	for _, k := range sc.KnownSpells {
		if k.Name == "Arcane Zap" {
			found = true
			assert.True(t, k.Custom)
		}
	}
	assert.True(t, found, "custom spell should be in known spells")
}

func TestSpellbookCreateCustomCantrip(t *testing.T) {
	m, char := newCustomSpellTestModel(t)

	m, _ = m.Update(tea.KeyPressMsg{Code: 'C', Text: "C"})
	m = csEnter(csType(m, "Spark")) // name
	m = csEnter(csType(m, "0"))     // level 0 = cantrip
	// Remaining fields left blank.
	for i := 0; i < 9; i++ {
		m = csEnter(m)
	}

	assert.Equal(t, ModeSpellList, m.mode)
	assert.Contains(t, char.Spellcasting.CantripsKnown, "Spark")
	assert.True(t, char.Spellcasting.IsCustomSpell("Spark"))
}

func TestSpellbookCustomSpellEscapeCancels(t *testing.T) {
	m, char := newCustomSpellTestModel(t)
	m, _ = m.Update(tea.KeyPressMsg{Code: 'C', Text: "C"})
	m = csType(m, "Half")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	assert.Equal(t, ModeSpellList, m.mode)
	assert.Empty(t, char.Spellcasting.CustomSpells)
}

func TestSpellbookCustomSpellEmptyNameRejected(t *testing.T) {
	m, _ := newCustomSpellTestModel(t)
	m, _ = m.Update(tea.KeyPressMsg{Code: 'C', Text: "C"})
	// Press enter with an empty name; should stay on step 0.
	m = csEnter(m)
	assert.Equal(t, 0, m.customStep, "empty name should not advance the form")
	assert.Equal(t, ModeCustomSpell, m.mode)
}

func TestCustomSpellToDataParsesComponents(t *testing.T) {
	cs := &models.CustomSpell{Name: "X", Level: 1, Components: "V, S, M", School: "Evocation"}
	sd := customSpellToData(cs)
	assert.Len(t, sd.Components, 3)
	assert.Equal(t, data.SpellSchool("Evocation"), sd.School)
}
