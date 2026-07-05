package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSpellDatabase2024 validates the expanded 2024 SRD 5.2 spell database.
func TestSpellDatabase2024(t *testing.T) {
	loader := NewLoader("../../data")

	spells, err := loader.GetSpells()
	require.NoError(t, err, "GetSpells() error")
	require.NotNil(t, spells)

	// The database was expanded to the full 2024 SRD 5.2 spell list.
	assert.GreaterOrEqual(t, len(spells.Spells), 300,
		"expected the expanded spell database to contain at least 300 spells")

	validComponents := map[string]bool{"V": true, "S": true, "M": true}
	seen := make(map[string]bool)

	for i := range spells.Spells {
		s := &spells.Spells[i]

		// No duplicate spell names.
		assert.False(t, seen[s.Name], "duplicate spell name: %s", s.Name)
		seen[s.Name] = true

		// Every spell must have the core fields populated for the UI to render.
		assert.NotEmpty(t, s.Name, "spell has empty Name")
		assert.NotEmpty(t, s.School, "spell %q has empty School", s.Name)
		assert.NotEmpty(t, s.Description, "spell %q has empty Description", s.Name)
		assert.NotEmpty(t, s.CastingTime, "spell %q has empty CastingTime", s.Name)
		assert.NotEmpty(t, s.Range, "spell %q has empty Range", s.Name)
		assert.NotEmpty(t, s.Duration, "spell %q has empty Duration", s.Name)
		assert.NotEmpty(t, s.Classes, "spell %q has no classes", s.Name)

		// Levels are cantrip (0) through 9th.
		assert.GreaterOrEqual(t, s.Level, 0, "spell %q has invalid level", s.Name)
		assert.LessOrEqual(t, s.Level, 9, "spell %q has invalid level", s.Name)

		// Components are limited to Verbal, Somatic, Material.
		for _, c := range s.Components {
			assert.True(t, validComponents[string(c)],
				"spell %q has invalid component %q", s.Name, c)
		}
	}

	// Representative spells spanning classes, levels, and 2024-specific rules.
	spellMap := make(map[string]*SpellData)
	for i := range spells.Spells {
		spellMap[spells.Spells[i].Name] = &spells.Spells[i]
	}
	for _, name := range []string{
		"Fireball", "Magic Missile", "Cure Wounds", "Wish",
		"Spirit Guardians", "Guiding Bolt", "Chromatic Orb", "Eldritch Blast",
		"Spike Growth", "Sunburst", "Divine Smite", "Aid",
		"Bestow Curse", "Vicious Mockery", "Flame Strike", "Circle of Death",
	} {
		assert.Contains(t, spellMap, name, "expected 2024 spell %q to be present", name)
	}

	// Spot-check 2024 mechanics: Divine Smite is a bonus-action spell in 2024.
	if ds, ok := spellMap["Divine Smite"]; ok {
		assert.Equal(t, "BA", string(ds.CastingTime), "Divine Smite should cast as a Bonus Action in 2024")
	}
}
