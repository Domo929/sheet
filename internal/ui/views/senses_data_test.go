package views

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
)

// raceTraitFeatures flattens a race's traits (and an optional subtype's traits)
// into the models.Feature slice the senses parser consumes.
func raceTraitFeatures(t *testing.T, loader *data.Loader, raceName, subtypeName string) []models.Feature {
	t.Helper()
	race, err := loader.FindRaceByName(raceName)
	require.NoError(t, err)

	var feats []models.Feature
	for _, tr := range race.Traits {
		feats = append(feats, models.NewFeature(tr.Name, raceName, tr.Description))
	}
	for _, st := range race.Subtypes {
		if st.Name == subtypeName {
			for _, tr := range st.Traits {
				feats = append(feats, models.NewFeature(tr.Name, subtypeName, tr.Description))
			}
		}
	}
	return feats
}

// TestSensesFromEmbeddedRaceData guards against drift in the shipped race data
// wording that the senses/speed parser relies on.
func TestSensesFromEmbeddedRaceData(t *testing.T) {
	loader := data.NewEmbeddedLoader()
	require.NoError(t, loader.LoadAll())

	assert.Equal(t, 120, models.DeriveSensesFromTraits(raceTraitFeatures(t, loader, "Dwarf", "")).Darkvision)
	assert.Equal(t, 60, models.DeriveSensesFromTraits(raceTraitFeatures(t, loader, "Elf", "")).Darkvision)
	assert.Equal(t, 120, models.DeriveSensesFromTraits(raceTraitFeatures(t, loader, "Elf", "Drow")).Darkvision)
	assert.Zero(t, models.DeriveSensesFromTraits(raceTraitFeatures(t, loader, "Human", "")).Darkvision)

	// Dwarf Stonecunning tremorsense is a Bonus Action, not a permanent sense.
	assert.Zero(t, models.DeriveSensesFromTraits(raceTraitFeatures(t, loader, "Dwarf", "")).Tremorsense)

	// Wood Elf permanently walks faster.
	assert.Equal(t, 35, models.DeriveWalkSpeedOverride(raceTraitFeatures(t, loader, "Elf", "Wood Elf")))
}
