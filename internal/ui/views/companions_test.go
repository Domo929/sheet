package views

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/stretchr/testify/assert"
)

func typeStr(m *CompanionsModel, s string) *CompanionsModel {
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

func enter(m *CompanionsModel) *CompanionsModel {
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	return m
}

func newCompanionsTestModel(t *testing.T) (*CompanionsModel, *models.Character) {
	t.Helper()
	char := models.NewCharacter("test-1", "Ranger", "Human", "Ranger")
	store, _ := storage.NewCharacterStorage(t.TempDir())
	m := NewCompanionsModel(char, store)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 150, Height: 40})
	return m, char
}

func TestCompanionsAddFlow(t *testing.T) {
	m, char := newCompanionsTestModel(t)

	// Open the add form.
	m, _ = m.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	assert.Equal(t, companionModeAdd, m.mode)

	m = enter(typeStr(m, "Dire Wolf")) // name
	m = enter(m)                       // accept default kind (Companion)
	m = enter(typeStr(m, "Large Beast"))
	m = enter(typeStr(m, "14"))              // AC
	m = enter(typeStr(m, "37"))              // Max HP
	m = enter(typeStr(m, "50 ft."))          // Speed
	m = enter(typeStr(m, "17 15 15 3 12 7")) // abilities -> creates

	assert.Equal(t, companionModeList, m.mode, "should return to list after creation")
	assert.Len(t, char.Companions, 1)

	c := char.Companions[0]
	assert.Equal(t, "Dire Wolf", c.Name)
	assert.Equal(t, models.CompanionPet, c.Kind)
	assert.Equal(t, "Large", c.Size)
	assert.Equal(t, "Beast", c.Type)
	assert.Equal(t, 14, c.AC)
	assert.Equal(t, 37, c.MaxHP)
	assert.Equal(t, 37, c.CurrentHP, "current HP should initialize to max")
	assert.Equal(t, "50 ft.", c.Speed)
	assert.Equal(t, 17, c.Abilities[0])
	assert.Equal(t, 3, c.Modifier(0), "STR 17 -> +3")
}

func TestCompanionsAddAttack(t *testing.T) {
	m, char := newCompanionsTestModel(t)
	char.AddCompanion(models.Companion{ID: "c1", Name: "Wolf", MaxHP: 11, CurrentHP: 11})
	m.cursor = 0

	m, _ = m.Update(tea.KeyPressMsg{Code: 'w', Text: "w"})
	assert.Equal(t, companionModeAttack, m.mode)

	m = enter(typeStr(m, "Bite"))
	m = enter(typeStr(m, "5"))
	m = enter(typeStr(m, "2d6 + 3 piercing"))

	assert.Equal(t, companionModeList, m.mode)
	assert.Len(t, char.Companions[0].Attacks, 1)
	atk := char.Companions[0].Attacks[0]
	assert.Equal(t, "Bite", atk.Name)
	assert.Equal(t, 5, atk.Bonus)
	assert.Equal(t, "2d6 + 3 piercing", atk.Damage)
}

func TestCompanionsDamageAndHeal(t *testing.T) {
	m, char := newCompanionsTestModel(t)
	char.AddCompanion(models.Companion{ID: "c1", Name: "Bear", MaxHP: 34, CurrentHP: 34})
	m.cursor = 0

	// Damage 10.
	m, _ = m.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
	m = enter(typeStr(m, "10"))
	assert.Equal(t, 24, char.Companions[0].CurrentHP)

	// Heal 5.
	m, _ = m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})
	m = enter(typeStr(m, "5"))
	assert.Equal(t, 29, char.Companions[0].CurrentHP)
}

func TestCompanionsDeleteConfirm(t *testing.T) {
	m, char := newCompanionsTestModel(t)
	char.AddCompanion(models.Companion{ID: "c1", Name: "Wolf", MaxHP: 11, CurrentHP: 11})
	m.cursor = 0

	m, _ = m.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	assert.True(t, m.confirmingDelete)
	m, _ = m.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	assert.Len(t, char.Companions, 0)
}

func TestCompanionsRenderStatBlock(t *testing.T) {
	m, char := newCompanionsTestModel(t)
	char.AddCompanion(models.Companion{
		ID: "c1", Name: "Panther", Kind: models.CompanionWildShape,
		Size: "Medium", Type: "Beast", AC: 12, MaxHP: 13, CurrentHP: 13,
		Speed: "50 ft., climb 40 ft.", Abilities: [6]int{14, 15, 10, 3, 14, 7},
		Attacks: []models.CompanionAttack{{Name: "Bite", Bonus: 4, Damage: "1d6 + 2 piercing"}},
	})
	m.cursor = 0

	view := m.View()
	assert.Contains(t, view, "Panther")
	assert.Contains(t, view, "Bite")
	assert.Contains(t, view, "13/13")
}

func TestCompanionsEmptyState(t *testing.T) {
	m, _ := newCompanionsTestModel(t)
	view := m.View()
	assert.Contains(t, view, "No companions")
}
