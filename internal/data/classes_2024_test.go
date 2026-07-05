package data

import (
	"testing"

	"github.com/Domo929/sheet/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// featureLevel returns the level of the first feature with the given name, or -1.
func featureLevel(c *Class, name string) int {
	for _, f := range c.Features {
		if f.Name == name {
			return f.Level
		}
	}
	return -1
}

func hasFeature(c *Class, name string) bool {
	return featureLevel(c, name) >= 0
}

// TestAllSubclassesChosenAtLevel3 verifies the 2024 rule that every class picks
// its subclass at level 3 (the first subclass feature is at level 3).
func TestAllSubclassesChosenAtLevel3(t *testing.T) {
	loader := NewLoader("../../data")
	classes, err := loader.GetClasses()
	require.NoError(t, err)

	for _, c := range classes.Classes {
		if len(c.Subclasses) == 0 {
			continue
		}
		sub := c.Subclasses[0]
		require.NotEmpty(t, sub.Features, "%s subclass %s has no features", c.Name, sub.Name)
		assert.Equal(t, 3, sub.Features[0].Level,
			"%s: first subclass feature should be at level 3 (2024 rules)", c.Name)
	}
}

// TestPaladinRangerSpellcastingAtLevel1 verifies the 2024 change moving
// Paladin and Ranger Spellcasting from level 2 to level 1, including slots.
func TestPaladinRangerSpellcastingAtLevel1(t *testing.T) {
	loader := NewLoader("../../data")

	for _, name := range []string{"Paladin", "Ranger"} {
		c, err := loader.FindClassByName(name)
		require.NoError(t, err)
		assert.Equal(t, 1, featureLevel(c, "Spellcasting"),
			"%s Spellcasting should be a level 1 feature in 2024", name)

		// Half-casters now have two 1st-level slots at character level 1.
		var l1 *SpellSlot
		for i := range c.SpellSlots {
			if c.SpellSlots[i].Level == 1 {
				l1 = &c.SpellSlots[i]
			}
		}
		require.NotNil(t, l1, "%s missing level 1 spell slot row", name)
		assert.Equal(t, 2, l1.First, "%s should have two 1st-level spell slots at level 1", name)
	}
}

// TestWeaponMasteryGrants verifies that exactly the five 2024 martial classes
// grant Weapon Mastery at level 1 with the correct counts, and that they carry
// the "Weapon Mastery" class feature.
func TestWeaponMasteryGrants(t *testing.T) {
	loader := NewLoader("../../data")
	classes, err := loader.GetClasses()
	require.NoError(t, err)

	want := map[string]int{
		"Barbarian": 2,
		"Fighter":   3,
		"Paladin":   2,
		"Ranger":    2,
		"Rogue":     2,
	}

	for _, c := range classes.Classes {
		count, expected := want[c.Name]
		if expected {
			require.NotNil(t, c.WeaponMastery, "%s should grant Weapon Mastery", c.Name)
			assert.Equal(t, 1, c.WeaponMastery.Level, "%s Weapon Mastery should be gained at level 1", c.Name)
			assert.Equal(t, count, c.WeaponMastery.Count, "%s Weapon Mastery count", c.Name)
			assert.Equal(t, 1, featureLevel(&c, "Weapon Mastery"),
				"%s should have a level 1 Weapon Mastery feature", c.Name)
		} else {
			assert.Nil(t, c.WeaponMastery, "%s should NOT grant Weapon Mastery", c.Name)
		}
	}
}

// TestLegacy2014FeaturesRemoved verifies the removal/replacement of 2014-only
// features at levels 1-3.
func TestLegacy2014FeaturesRemoved(t *testing.T) {
	loader := NewLoader("../../data")

	cases := []struct {
		class   string
		feature string
	}{
		{"Ranger", "Natural Explorer"},
		{"Ranger", "Primeval Awareness"},
		{"Paladin", "Divine Sense"},
		{"Paladin", "Divine Health"},
		{"Bard", "Song of Rest"},
		{"Sorcerer", "Sorcerous Origin"},
	}
	for _, tc := range cases {
		c, err := loader.FindClassByName(tc.class)
		require.NoError(t, err)
		assert.False(t, hasFeature(c, tc.feature),
			"%s should no longer have legacy feature %q (2024 rules)", tc.class, tc.feature)
	}

	// New 2024 level-1 "order"/identity features are present.
	present := []struct{ class, feature string }{
		{"Cleric", "Divine Order"},
		{"Druid", "Primal Order"},
		{"Sorcerer", "Innate Sorcery"},
		{"Warlock", "Eldritch Invocations"},
		{"Wizard", "Ritual Adept"},
	}
	for _, tc := range present {
		c, err := loader.FindClassByName(tc.class)
		require.NoError(t, err)
		assert.Equal(t, 1, featureLevel(c, tc.feature),
			"%s should have level 1 feature %q (2024 rules)", tc.class, tc.feature)
	}
}

// TestWeaponMasteryProperties verifies representative 2024 weapon → mastery
// assignments are present in the equipment data.
func TestWeaponMasteryProperties(t *testing.T) {
	loader := NewLoader("../../data")
	equipment, err := loader.GetEquipment()
	require.NoError(t, err)

	want := map[string]domain.WeaponMastery{
		"Longsword":      domain.MasterySap,
		"Greataxe":       domain.MasteryCleave,
		"Dagger":         domain.MasteryNick,
		"Pike":           domain.MasteryPush,
		"Rapier":         domain.MasteryVex,
		"Glaive":         domain.MasteryGraze,
		"Club":           domain.MasterySlow,
		"Quarterstaff":   domain.MasteryTopple,
		"Crossbow, Hand": domain.MasteryVex,
	}

	byName := map[string]domain.WeaponMastery{}
	for _, w := range equipment.Weapons.GetAllWeapons() {
		byName[w.Name] = w.Mastery
		assert.NotEmpty(t, w.Mastery, "weapon %q should have a mastery property (2024 rules)", w.Name)
	}

	for name, mastery := range want {
		assert.Equal(t, mastery, byName[name], "weapon %q mastery", name)
	}
}
