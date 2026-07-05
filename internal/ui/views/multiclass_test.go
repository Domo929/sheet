package views

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/stretchr/testify/assert"
)

func mcType(m *MulticlassModel, s string) *MulticlassModel {
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

func mcEnter(m *MulticlassModel) *MulticlassModel {
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	return m
}

func mcKey(m *MulticlassModel, code rune) *MulticlassModel {
	m, _ = m.Update(tea.KeyPressMsg{Code: code})
	return m
}

func newMulticlassTestModel(t *testing.T, class string, level int) (*MulticlassModel, *models.Character) {
	t.Helper()
	char := models.NewCharacter("mc-1", "Hero", "Human", class)
	char.Info.Level = level
	store, _ := storage.NewCharacterStorage(t.TempDir())
	loader := data.NewLoader("../../../data")
	m := NewMulticlassModel(char, store, loader)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 150, Height: 40})
	return m, char
}

func indexOf(list []string, target string) int {
	for i, v := range list {
		if v == target {
			return i
		}
	}
	return -1
}

// selectClassInPicker moves the add-form picker to the named class.
func selectClassInPicker(m *MulticlassModel, class string) *MulticlassModel {
	idx := indexOf(m.addAvail, class)
	for i := 0; i < idx; i++ {
		m = mcKey(m, tea.KeyDown)
	}
	return m
}

func TestMulticlassSeedsFromInfo(t *testing.T) {
	m, char := newMulticlassTestModel(t, "Wizard", 5)
	assert.Len(t, char.Classes, 1, "should seed one class entry from Info")
	assert.Equal(t, "Wizard", char.Classes[0].Class)
	assert.Equal(t, 5, char.Classes[0].Level)
	assert.Greater(t, char.Classes[0].HitDie, 0, "hit die should be populated from loader")
	assert.False(t, char.IsMulticlass())
	_ = m
}

func TestMulticlassAddClassFlow(t *testing.T) {
	m, char := newMulticlassTestModel(t, "Wizard", 5)

	// Open the add form.
	m = mcKey(m, 'a')
	assert.Equal(t, multiclassModeAdd, m.mode)

	// Pick Cleric, then level 3, empty subclass -> create.
	m = selectClassInPicker(m, "Cleric")
	m = mcEnter(m)                                           // accept class -> level step
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace}) // clear default "1"
	m = mcType(m, "3")                                       // level 3
	m = mcEnter(m)                                           // -> subclass step
	m = mcEnter(m)                                           // empty subclass -> create

	assert.Equal(t, multiclassModeList, m.mode)
	assert.Len(t, char.Classes, 2)
	assert.Equal(t, "Cleric", char.Classes[1].Class)
	assert.Equal(t, 3, char.Classes[1].Level)
	assert.True(t, char.IsMulticlass())

	// Primary should stay Wizard; Info.Level should reflect total (8).
	assert.Equal(t, "Wizard", char.Info.Class)
	assert.Equal(t, 8, char.Info.Level)
	assert.Equal(t, 8, char.TotalLevel())

	// Combined caster level 8 -> slots 4/3/3/2.
	if assert.NotNil(t, char.Spellcasting) {
		assert.Equal(t, 4, char.Spellcasting.SpellSlots.GetSlot(1).Total)
		assert.Equal(t, 3, char.Spellcasting.SpellSlots.GetSlot(2).Total)
		assert.Equal(t, 3, char.Spellcasting.SpellSlots.GetSlot(3).Total)
		assert.Equal(t, 2, char.Spellcasting.SpellSlots.GetSlot(4).Total)
	}
}

func TestMulticlassAdjustLevel(t *testing.T) {
	m, char := newMulticlassTestModel(t, "Fighter", 5)

	// Increment primary level.
	m = mcKey(m, '+')
	assert.Equal(t, 6, char.Classes[0].Level)
	assert.Equal(t, 6, char.TotalLevel())
	assert.Equal(t, 6, char.Info.Level)

	// Decrement back.
	m = mcKey(m, '-')
	assert.Equal(t, 5, char.Classes[0].Level)

	// Cannot go below 1.
	for i := 0; i < 10; i++ {
		m = mcKey(m, '-')
	}
	assert.Equal(t, 1, char.Classes[0].Level)
}

func TestMulticlassRemoveClass(t *testing.T) {
	m, char := newMulticlassTestModel(t, "Wizard", 5)

	// Add Cleric.
	m = mcKey(m, 'a')
	m = selectClassInPicker(m, "Cleric")
	m = mcEnter(m)
	m = mcEnter(m) // default level 1
	m = mcEnter(m) // empty subclass -> create
	assert.Len(t, char.Classes, 2)

	// Select the Cleric entry and remove it.
	m = mcKey(m, tea.KeyDown) // cursor -> 1
	m = mcKey(m, 'x')
	assert.True(t, m.confirmingDelete)
	m = mcKey(m, 'y')

	assert.Len(t, char.Classes, 1)
	assert.Equal(t, "Wizard", char.Classes[0].Class)
	assert.False(t, char.IsMulticlass())
}

func TestMulticlassCannotRemoveLastClass(t *testing.T) {
	m, char := newMulticlassTestModel(t, "Wizard", 5)
	m = mcKey(m, 'x')
	assert.False(t, m.confirmingDelete, "should not offer to remove the only class")
	assert.Len(t, char.Classes, 1)
}

func TestMulticlassTotalLevelCap(t *testing.T) {
	m, char := newMulticlassTestModel(t, "Wizard", 20)
	// Adding another class at level 20 total should be refused.
	m = mcKey(m, 'a')
	assert.Equal(t, multiclassModeList, m.mode, "add should be refused at total level 20")
	assert.Len(t, char.Classes, 1)
}

func TestMulticlassViewRenders(t *testing.T) {
	m, _ := newMulticlassTestModel(t, "Wizard", 5)
	out := m.View()
	assert.Contains(t, out, "Classes & Multiclassing")
	assert.Contains(t, out, "Wizard")
}
