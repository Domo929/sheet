package views

import (
	"testing"

	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newFinalizeTestModel builds a creation model wired to the on-disk data files
// with a deterministic ability array, ready for finalizeCharacter().
func newFinalizeTestModel(t *testing.T) (*CharacterCreationModel, *data.Loader) {
	t.Helper()
	store, err := storage.NewCharacterStorage(t.TempDir())
	require.NoError(t, err)
	loader := data.NewLoader("../../../data")
	model := NewCharacterCreationModel(store, loader)
	model.abilityScores = [6]int{15, 14, 13, 12, 10, 8}
	return model, loader
}

func TestFinalizeInitializesFullCasterSpellSlots(t *testing.T) {
	model, loader := newFinalizeTestModel(t)
	wizard, err := loader.FindClassByName("Wizard")
	require.NoError(t, err)
	model.selectedClass = wizard

	model.finalizeCharacter()

	require.NotNil(t, model.character.Spellcasting, "wizard should have spellcasting initialized at creation")
	assert.Equal(t, models.Ability("intelligence"), model.character.Spellcasting.Ability)
	require.NotNil(t, model.character.Spellcasting.SpellSlots.GetSlot(1))
	assert.Equal(t, 2, model.character.Spellcasting.SpellSlots.GetSlot(1).Total,
		"wizard should have 2 first-level spell slots at level 1")
	assert.Nil(t, model.character.Spellcasting.PactMagic, "wizard should not have pact magic")
}

func TestFinalizeInitializesWarlockPactMagic(t *testing.T) {
	model, loader := newFinalizeTestModel(t)
	warlock, err := loader.FindClassByName("Warlock")
	require.NoError(t, err)
	model.selectedClass = warlock

	model.finalizeCharacter()

	require.NotNil(t, model.character.Spellcasting)
	require.NotNil(t, model.character.Spellcasting.PactMagic, "warlock should have pact magic slots")
	assert.Equal(t, 1, model.character.Spellcasting.PactMagic.Total)
	assert.Equal(t, 1, model.character.Spellcasting.PactMagic.SlotLevel)
	assert.Equal(t, 1, model.character.Spellcasting.PactMagic.Remaining)
}

func TestFinalizeNonCasterHasNoSpellcasting(t *testing.T) {
	model, loader := newFinalizeTestModel(t)
	fighter, err := loader.FindClassByName("Fighter")
	require.NoError(t, err)
	model.selectedClass = fighter

	model.finalizeCharacter()

	assert.Nil(t, model.character.Spellcasting, "fighter should not have spellcasting")
}

// TestFinalizePaladinHalfCasterSlotsAtLevel1 locks in the 2024 change that
// Paladins (and Rangers) are spellcasters from level 1 with two 1st-level slots.
func TestFinalizePaladinHalfCasterSlotsAtLevel1(t *testing.T) {
	model, loader := newFinalizeTestModel(t)
	paladin, err := loader.FindClassByName("Paladin")
	require.NoError(t, err)
	model.selectedClass = paladin

	model.finalizeCharacter()

	require.NotNil(t, model.character.Spellcasting, "paladin should have spellcasting at level 1 (2024)")
	assert.Equal(t, models.Ability("charisma"), model.character.Spellcasting.Ability)
	require.NotNil(t, model.character.Spellcasting.SpellSlots.GetSlot(1))
	assert.Equal(t, 2, model.character.Spellcasting.SpellSlots.GetSlot(1).Total,
		"level 1 paladin should have 2 first-level spell slots (2024)")
}

func TestFinalizeAppliesBackgroundOriginFeat(t *testing.T) {
	model, loader := newFinalizeTestModel(t)
	fighter, err := loader.FindClassByName("Fighter")
	require.NoError(t, err)
	model.selectedClass = fighter
	// Mock background granting Tavern Brawler (+1 Str/Con) with no ability bonuses,
	// so the only ability change comes from the feat itself.
	model.selectedBackground = &data.Background{Name: "Test", Feat: "Tavern Brawler"}

	model.finalizeCharacter()

	var found bool
	for _, f := range model.character.Features.Feats {
		if f.Name == "Tavern Brawler" {
			found = true
			assert.NotEmpty(t, f.Description, "feat description should be resolved from feats.json")
		}
	}
	assert.True(t, found, "Tavern Brawler feat should be recorded on the sheet")
	// Strength started at 15; the feat's first ASI option (Strength) adds +1.
	assert.Equal(t, 16, model.character.AbilityScores.Strength.Base)
}

func TestFinalizeAppliesMagicInitiateVariantFeat(t *testing.T) {
	model, loader := newFinalizeTestModel(t)
	fighter, err := loader.FindClassByName("Fighter")
	require.NoError(t, err)
	model.selectedClass = fighter
	// Background feats can include a parenthetical variant; the full label is kept
	// on the sheet while the lookup strips the parenthetical.
	model.selectedBackground = &data.Background{Name: "Test", Feat: "Magic Initiate (Cleric)"}

	model.finalizeCharacter()

	var found bool
	for _, f := range model.character.Features.Feats {
		if f.Name == "Magic Initiate (Cleric)" {
			found = true
			assert.NotEmpty(t, f.Description, "variant feat should resolve to the base feat's description")
		}
	}
	assert.True(t, found, "Magic Initiate (Cleric) should be recorded with the full label")
}

func TestFinalizePopulatesTraitsAndClassFeatures(t *testing.T) {
	model, loader := newFinalizeTestModel(t)
	human, err := loader.FindRaceByName("Human")
	require.NoError(t, err)
	model.selectedRace = human
	fighter, err := loader.FindClassByName("Fighter")
	require.NoError(t, err)
	model.selectedClass = fighter

	model.finalizeCharacter()

	assert.NotEmpty(t, model.character.Features.RacialTraits, "racial traits should be populated from the species")
	require.NotEmpty(t, model.character.Features.ClassFeatures, "level 1 class features should be populated")

	var names []string
	for _, f := range model.character.Features.ClassFeatures {
		names = append(names, f.Name)
		assert.Equal(t, 1, f.Level, "only level 1 features should be added at creation")
	}
	assert.Contains(t, names, "Second Wind", "Fighter should gain Second Wind at level 1")
}

// TestFinalizeOriginFeatASIChoiceConstitution verifies that when the origin feat
// offers an ability choice (Tavern Brawler: Strength or Constitution), the
// player's Review-step selection is applied instead of the default first option.
func TestFinalizeOriginFeatASIChoiceConstitution(t *testing.T) {
	model, loader := newFinalizeTestModel(t)
	fighter, err := loader.FindClassByName("Fighter")
	require.NoError(t, err)
	model.selectedClass = fighter
	model.selectedBackground = &data.Background{Name: "Test", Feat: "Tavern Brawler"}

	// Player toggled the choice to the second option (Constitution).
	opts, featName := model.originFeatASIOptions()
	require.Equal(t, "Tavern Brawler", featName)
	require.Equal(t, []string{"Strength", "Constitution"}, opts)
	model.originFeatASIChoice = 1

	model.finalizeCharacter()

	// CON started at 13 in the deterministic array (index 2); +1 from the feat.
	assert.Equal(t, 14, model.character.AbilityScores.Constitution.Base,
		"the chosen ability (Constitution) should get the feat ASI")
	assert.Equal(t, 15, model.character.AbilityScores.Strength.Base,
		"Strength should be unchanged when Constitution is chosen")
}

// TestOriginFeatASIOptionsNoneForSingleOption confirms the Review toggle is not
// offered for feats whose ASI has a single fixed ability, nor for feats with no
// ASI at all.
func TestOriginFeatASIOptionsNoneForSingleOption(t *testing.T) {
	model, _ := newFinalizeTestModel(t)

	model.selectedBackground = &data.Background{Name: "Test", Feat: "Actor"} // +1 Charisma only
	opts, _ := model.originFeatASIOptions()
	assert.Empty(t, opts, "single-option feat ASI should not present a choice")

	model.selectedBackground = &data.Background{Name: "Test", Feat: "Alert"} // no ASI
	opts, _ = model.originFeatASIOptions()
	assert.Empty(t, opts, "feat without ASI should not present a choice")
}
