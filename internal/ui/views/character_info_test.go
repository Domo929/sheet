package views

import (
	"testing"

	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCharacterInfo_FeaturesCategorySwitch(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Fighter", "Human", "Fighter")
	char.Features.AddRacialTrait("Darkvision", "Human", "See in dim light")
	char.Features.AddClassFeature("Action Surge", "Fighter 2", "Extra action", 2, "")
	char.Features.AddFeat("Alert", "Can't be surprised")

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusFeatures

	assert.Equal(t, FeatureCategoryRacial, m.featureCategory)
	features := m.getFeaturesForCategory()
	assert.Equal(t, 1, len(features))
	assert.Equal(t, "Darkvision", features[0].Name)

	// Switch to Class
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, FeatureCategoryClass, m.featureCategory)

	// Switch to Feats
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, FeatureCategoryFeats, m.featureCategory)
}

func TestCharacterInfo_TabSwitchFocus(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)

	assert.Equal(t, CharInfoFocusPersonality, m.focus)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, CharInfoFocusFeatures, m.focus)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, CharInfoFocusPersonality, m.focus)
}

func TestCharacterInfo_FeatureNavigation(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	char.Features.AddRacialTrait("Darkvision", "Human", "Desc1")
	char.Features.AddRacialTrait("Fey Ancestry", "Elf", "Desc2")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusFeatures

	assert.Equal(t, 0, m.featureCursor)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, m.featureCursor)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, m.featureCursor) // clamped
}

func TestCharacterInfo_BackToSheet(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	msg := cmd()
	_, ok := msg.(BackToSheetMsg)
	assert.True(t, ok)
}

func TestCharacterInfo_OpenNotes(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	msg := cmd()
	notesMsg, ok := msg.(OpenNotesMsg)
	assert.True(t, ok)
	assert.Equal(t, "charinfo", notesMsg.ReturnTo)
}

func TestCharacterInfo_CategoryDoesNotWrapLeft(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusFeatures

	assert.Equal(t, FeatureCategoryRacial, m.featureCategory)
	// Left arrow should not wrap
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	assert.Equal(t, FeatureCategoryRacial, m.featureCategory)
}

func TestCharacterInfo_CategoryDoesNotWrapRight(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusFeatures

	// Go all the way right
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, FeatureCategoryFeats, m.featureCategory)

	// Should not wrap past Feats
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, FeatureCategoryFeats, m.featureCategory)
}

func TestCharacterInfo_SubclassFeaturesSplit(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Elf", "Fighter")
	char.Info.Subclass = "Champion"
	char.Features.AddClassFeature("Action Surge", "Fighter 2", "Extra action", 2, "")
	char.Features.AddClassFeature("Improved Critical", "Champion 3", "Crit on 19-20", 3, "")
	char.Features.AddClassFeature("Extra Attack", "Fighter 5", "Attack twice", 5, "")

	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)

	// Class features should exclude Champion features
	m.featureCategory = FeatureCategoryClass
	classFeatures := m.getFeaturesForCategory()
	assert.Equal(t, 2, len(classFeatures))
	for _, f := range classFeatures {
		assert.NotContains(t, f.Source, "Champion")
	}

	// Subclass features should only include Champion features
	m.featureCategory = FeatureCategorySubclass
	subclassFeatures := m.getFeaturesForCategory()
	assert.Equal(t, 1, len(subclassFeatures))
	assert.Equal(t, "Improved Critical", subclassFeatures[0].Name)
}

func TestCharacterInfo_ShiftTabSwitchesFocus(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)

	assert.Equal(t, CharInfoFocusPersonality, m.focus)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	assert.Equal(t, CharInfoFocusFeatures, m.focus)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	assert.Equal(t, CharInfoFocusPersonality, m.focus)
}

func TestCharacterInfo_FeatureCursorResetsOnCategorySwitch(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	char.Features.AddRacialTrait("Darkvision", "Human", "See in dim light")
	char.Features.AddRacialTrait("Fey Ancestry", "Elf", "Resistance")
	char.Features.AddFeat("Alert", "Can't be surprised")

	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusFeatures

	// Move cursor down in racial
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, m.featureCursor)

	// Switch to Class — cursor should reset
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, 0, m.featureCursor)
}

func TestCharacterInfo_ViewRendersWithoutCrash(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Character", "Elf", "Wizard")
	char.Personality.AddTrait("Curious about everything")
	char.Personality.AddTrait("Always fidgeting")
	char.Personality.AddIdeal("Knowledge is power")
	char.Personality.AddBond("My mentor's library")
	char.Personality.AddFlaw("Can't resist a puzzle")
	char.Personality.Backstory = "Born in a small village on the edge of the Silverwood forest."
	char.Features.AddRacialTrait("Darkvision", "Elf", "See in dim light within 60 feet")
	char.Features.AddRacialTrait("Fey Ancestry", "Elf", "Advantage on saving throws against being charmed")
	char.Features.AddClassFeature("Arcane Recovery", "Wizard 1", "Recover spell slots on short rest", 1, "")
	char.Features.AddFeat("Alert", "Can't be surprised")

	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.width = 120
	m.height = 40

	// Should not panic
	view := m.View()
	assert.Contains(t, view, "Test Character")
	assert.Contains(t, view, "Personality")
	assert.Contains(t, view, "Features")
	assert.Contains(t, view, "Darkvision")
}

func TestCharacterInfo_EmptyCharacterView(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.width = 120
	m.height = 40

	view := m.View()
	assert.Contains(t, view, "Personality")
	assert.Contains(t, view, "Features")
	assert.Contains(t, view, "(none)")
}

func TestCharacterInfo_LeftRightDoNothingWhenPersonalityFocused(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusPersonality

	originalCategory := m.featureCategory
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, originalCategory, m.featureCategory)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	assert.Equal(t, originalCategory, m.featureCategory)
}

func TestCharacterInfo_GetFeaturesForCategory_NoSubclass(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	char.Features.AddClassFeature("Action Surge", "Fighter 2", "Extra action", 2, "")

	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)

	// No subclass set — subclass category should return empty
	m.featureCategory = FeatureCategorySubclass
	features := m.getFeaturesForCategory()
	assert.Equal(t, 0, len(features))

	// Class category should include all class features
	m.featureCategory = FeatureCategoryClass
	features = m.getFeaturesForCategory()
	assert.Equal(t, 1, len(features))
}

func TestCharacterInfo_CtrlCQuits(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.NotNil(t, cmd)
}
