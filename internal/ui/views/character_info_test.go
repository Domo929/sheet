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

func TestCharacterInfo_AddTrait(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusPersonality
	// Cursor starts at 0 (Traits header), which is in the Traits section
	m.personalityCursor = 0
	m.personalitySection = PersonalitySectionTraits

	// Press 'a' to add
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	assert.True(t, m.editMode)
	assert.Equal(t, "add", m.editAction)
	assert.Equal(t, PersonalitySectionTraits, m.editSection)

	// Type trait text
	for _, r := range "Brave" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Confirm with Enter
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, m.editMode)
	assert.Contains(t, char.Personality.Traits, "Brave")
}

func TestCharacterInfo_EditTrait(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	char.Personality.AddTrait("Old trait")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusPersonality
	// Items: 0=Traits header, 1="Old trait", 2=Ideals header, ...
	m.personalityCursor = 1
	m.personalitySection = PersonalitySectionTraits

	// Press 'e' to edit
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	assert.True(t, m.editMode)
	assert.Equal(t, "edit", m.editAction)
	assert.Equal(t, "Old trait", m.editBuffer) // prepopulated

	// Clear and type new text
	m.editBuffer = ""
	for _, r := range "New trait" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, m.editMode)
	assert.Equal(t, "New trait", char.Personality.Traits[0])
}

func TestCharacterInfo_DeleteTrait(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	char.Personality.AddTrait("Brave")
	char.Personality.AddTrait("Curious")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusPersonality
	// Items: 0=Traits header, 1="Brave", 2="Curious", 3=Ideals header, ...
	m.personalityCursor = 1
	m.personalitySection = PersonalitySectionTraits

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	assert.True(t, m.confirmingDelete)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	assert.False(t, m.confirmingDelete)
	assert.Equal(t, 1, len(char.Personality.Traits))
	assert.Equal(t, "Curious", char.Personality.Traits[0])
}

func TestCharacterInfo_DeleteCancel(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	char.Personality.AddTrait("Keep me")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusPersonality
	// Items: 0=Traits header, 1="Keep me", 2=Ideals header, ...
	m.personalityCursor = 1
	m.personalitySection = PersonalitySectionTraits

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	assert.True(t, m.confirmingDelete)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	assert.False(t, m.confirmingDelete)
	assert.Equal(t, 1, len(char.Personality.Traits))
}

func TestCharacterInfo_EditCancelOnEsc(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	char.Personality.AddTrait("Original")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusPersonality
	// Items: 0=Traits header, 1="Original", 2=Ideals header, ...
	m.personalityCursor = 1
	m.personalitySection = PersonalitySectionTraits

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	assert.True(t, m.editMode)

	// Type something
	for _, r := range "Changed" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	// Cancel with Esc
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, m.editMode)
	assert.Equal(t, "Original", char.Personality.Traits[0]) // unchanged
}

func TestCharacterInfo_DeleteCancelOnEsc(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	char.Personality.AddTrait("Keep me")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusPersonality
	m.personalityCursor = 1
	m.personalitySection = PersonalitySectionTraits

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	assert.True(t, m.confirmingDelete)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, m.confirmingDelete)
	assert.Equal(t, 1, len(char.Personality.Traits))
}

func TestCharacterInfo_AddIdeal(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusPersonality
	// Items: 0=Traits header, 1=(none) traits, 2=Ideals header, ...
	m.personalityCursor = 2
	m.personalitySection = PersonalitySectionIdeals

	// Press 'a' to add ideal
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	assert.True(t, m.editMode)
	assert.Equal(t, PersonalitySectionIdeals, m.editSection)

	for _, r := range "Justice" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Contains(t, char.Personality.Ideals, "Justice")
}

func TestCharacterInfo_EditBackstory(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	char.Personality.Backstory = "Old story"
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusPersonality

	// Move cursor to backstory item
	items := m.getAllPersonalityItems()
	backstoryIdx := -1
	for i, item := range items {
		if item.section == PersonalitySectionBackstory && !item.header {
			backstoryIdx = i
			break
		}
	}
	require.NotEqual(t, -1, backstoryIdx)
	m.personalityCursor = backstoryIdx
	m.personalitySection = PersonalitySectionBackstory

	// Press 'e' to edit
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	assert.True(t, m.editMode)
	assert.Equal(t, PersonalitySectionBackstory, m.editSection)
	assert.Equal(t, "Old story", m.editBuffer)

	// Enter should add newline (not save)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.True(t, m.editMode) // still editing
	assert.Contains(t, m.editBuffer, "\n")

	// Clear and set new content, save with ctrl+s
	m.editBuffer = "New story"
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	assert.False(t, m.editMode)
	assert.Equal(t, "New story", char.Personality.Backstory)
}

func TestCharacterInfo_BackspaceInEditMode(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	char.Personality.AddTrait("AB")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusPersonality
	m.personalityCursor = 1
	m.personalitySection = PersonalitySectionTraits

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	assert.Equal(t, "AB", m.editBuffer)

	// Backspace removes last character
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	assert.Equal(t, "A", m.editBuffer)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	assert.Equal(t, "", m.editBuffer)

	// Backspace on empty buffer does nothing
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	assert.Equal(t, "", m.editBuffer)
}

func TestCharacterInfo_AddEmptyDoesNotSave(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusPersonality
	m.personalityCursor = 0
	m.personalitySection = PersonalitySectionTraits

	// Press 'a' to add, then immediately Enter (empty buffer)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, m.editMode)
	assert.Equal(t, 0, len(char.Personality.Traits)) // nothing added
}

func TestCharacterInfo_EditDoesNotWorkOnHeader(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusPersonality
	m.personalityCursor = 0 // header
	m.personalitySection = PersonalitySectionTraits

	// Press 'e' on header - should NOT enter edit mode
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	assert.False(t, m.editMode)
}

func TestCharacterInfo_DeleteDoesNotWorkOnHeader(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusPersonality
	m.personalityCursor = 0 // header
	m.personalitySection = PersonalitySectionTraits

	// Press 'd' on header - should NOT enter delete confirmation
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	assert.False(t, m.confirmingDelete)
}

func TestCharacterInfo_EditDoesNotWorkWhenFeaturesFocused(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	char.Personality.AddTrait("A trait")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusFeatures

	// Press 'e' when features focused - should NOT enter edit mode
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	assert.False(t, m.editMode)
}

func TestCharacterInfo_EditModalView(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	char.Personality.AddTrait("A trait")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusPersonality
	m.personalityCursor = 1
	m.personalitySection = PersonalitySectionTraits
	m.width = 120
	m.height = 40

	// Enter edit mode
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	assert.True(t, m.editMode)

	// View should contain the edit modal
	view := m.View()
	assert.Contains(t, view, "Edit Trait")
	assert.Contains(t, view, "█") // cursor indicator
}

func TestCharacterInfo_DeleteConfirmModalView(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	char.Personality.AddTrait("Doomed trait")
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusPersonality
	m.personalityCursor = 1
	m.personalitySection = PersonalitySectionTraits
	m.width = 120
	m.height = 40

	// Enter delete confirmation
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	assert.True(t, m.confirmingDelete)

	// View should contain the confirmation
	view := m.View()
	assert.Contains(t, view, "Delete this trait?")
}

func TestCharacterInfo_GetCurrentSectionItems(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	char.Personality.AddTrait("Brave")
	char.Personality.AddTrait("Curious")
	char.Personality.AddIdeal("Justice")
	char.Personality.Backstory = "A tale"
	store, _ := storage.NewCharacterStorage("")
	m := NewCharacterInfoModel(char, store)

	m.personalitySection = PersonalitySectionTraits
	items := m.getCurrentSectionItems()
	assert.Equal(t, []string{"Brave", "Curious"}, items)

	m.personalitySection = PersonalitySectionIdeals
	items = m.getCurrentSectionItems()
	assert.Equal(t, []string{"Justice"}, items)

	m.personalitySection = PersonalitySectionBonds
	items = m.getCurrentSectionItems()
	assert.Equal(t, 0, len(items))

	m.personalitySection = PersonalitySectionBackstory
	items = m.getCurrentSectionItems()
	assert.Equal(t, []string{"A tale"}, items)
}

func TestCharacterInfo_SectionName(t *testing.T) {
	assert.Equal(t, "Trait", sectionName(PersonalitySectionTraits))
	assert.Equal(t, "Ideal", sectionName(PersonalitySectionIdeals))
	assert.Equal(t, "Bond", sectionName(PersonalitySectionBonds))
	assert.Equal(t, "Flaw", sectionName(PersonalitySectionFlaws))
	assert.Equal(t, "Backstory", sectionName(PersonalitySectionBackstory))
}
