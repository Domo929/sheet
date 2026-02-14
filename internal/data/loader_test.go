package data

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoader(t *testing.T) {
	t.Run("with custom directory", func(t *testing.T) {
		loader := NewLoader("/custom/path")
		assert.Equal(t, "/custom/path", loader.dataDir)
	})

	t.Run("with empty directory", func(t *testing.T) {
		loader := NewLoader("")
		assert.Equal(t, "./data", loader.dataDir)
	})
}

func TestLoaderGetRaces(t *testing.T) {
	loader := NewLoader("../../data")

	races, err := loader.GetRaces()
	require.NoError(t, err, "GetRaces() error")
	require.NotNil(t, races, "GetRaces() returned nil")
	assert.NotEmpty(t, races.Races, "GetRaces() returned empty races list")

	// Verify some expected races exist
	expectedRaces := []string{"Human", "Elf", "Dwarf", "Halfling"}
	raceNames := make([]string, len(races.Races))
	for i, race := range races.Races {
		raceNames[i] = race.Name
	}

	for _, expected := range expectedRaces {
		assert.Contains(t, raceNames, expected, "expected race '%s' not found", expected)
	}

	// Test caching - second call should use cache
	races2, err := loader.GetRaces()
	require.NoError(t, err, "GetRaces() second call error")
	assert.Same(t, races, races2, "GetRaces() did not return cached data")
}

func TestLoaderGetClasses(t *testing.T) {
	loader := NewLoader("../../data")

	classes, err := loader.GetClasses()
	require.NoError(t, err, "GetClasses() error")
	require.NotNil(t, classes, "GetClasses() returned nil")
	assert.NotEmpty(t, classes.Classes, "GetClasses() returned empty classes list")

	// Verify some expected classes exist
	expectedClasses := []string{"Barbarian", "Bard", "Cleric", "Druid", "Fighter", "Monk", "Paladin", "Ranger", "Rogue", "Sorcerer", "Warlock", "Wizard"}
	classMap := make(map[string]*Class)
	for i := range classes.Classes {
		classMap[classes.Classes[i].Name] = &classes.Classes[i]
	}

	for _, expected := range expectedClasses {
		class, found := classMap[expected]
		assert.True(t, found, "expected class '%s' not found", expected)
		if found {
			assert.NotEmpty(t, class.HitDice, "class '%s' has empty HitDice", class.Name)
			assert.NotEmpty(t, class.SavingThrowProficiencies, "class '%s' has no saving throw proficiencies", class.Name)
		}
	}

	// Test caching
	classes2, err := loader.GetClasses()
	require.NoError(t, err, "GetClasses() second call error")
	assert.Same(t, classes, classes2, "GetClasses() did not return cached data")
}

func TestLoaderGetSpells(t *testing.T) {
	loader := NewLoader("../../data")

	spells, err := loader.GetSpells()
	require.NoError(t, err, "GetSpells() error")
	require.NotNil(t, spells, "GetSpells() returned nil")
	assert.NotEmpty(t, spells.Spells, "GetSpells() returned empty spells list")

	// Verify some expected spells exist
	expectedSpells := []string{"Fireball", "Magic Missile", "Cure Wounds"}
	spellMap := make(map[string]*SpellData)
	for i := range spells.Spells {
		spellMap[spells.Spells[i].Name] = &spells.Spells[i]
	}

	for _, expected := range expectedSpells {
		spell, found := spellMap[expected]
		assert.True(t, found, "expected spell '%s' not found", expected)
		if found {
			assert.NotEmpty(t, spell.School, "spell '%s' has empty School", spell.Name)
			assert.NotEmpty(t, spell.Description, "spell '%s' has empty Description", spell.Name)
		}
	}

	// Test caching
	spells2, err := loader.GetSpells()
	require.NoError(t, err, "GetSpells() second call error")
	assert.Same(t, spells, spells2, "GetSpells() did not return cached data")
}

func TestLoaderGetBackgrounds(t *testing.T) {
	loader := NewLoader("../../data")

	backgrounds, err := loader.GetBackgrounds()
	require.NoError(t, err, "GetBackgrounds() error")
	require.NotNil(t, backgrounds, "GetBackgrounds() returned nil")
	assert.NotEmpty(t, backgrounds.Backgrounds, "GetBackgrounds() returned empty backgrounds list")

	// Verify some expected backgrounds exist
	expectedBackgrounds := []string{"Acolyte", "Criminal", "Noble", "Sage"}
	bgMap := make(map[string]*Background)
	for i := range backgrounds.Backgrounds {
		bgMap[backgrounds.Backgrounds[i].Name] = &backgrounds.Backgrounds[i]
	}

	for _, expected := range expectedBackgrounds {
		bg, found := bgMap[expected]
		assert.True(t, found, "expected background '%s' not found", expected)
		if found {
			assert.NotEmpty(t, bg.Description, "background '%s' has empty Description", bg.Name)
			assert.NotEmpty(t, bg.SkillProficiencies, "background '%s' has no skill proficiencies", bg.Name)
		}
	}

	// Test caching
	backgrounds2, err := loader.GetBackgrounds()
	require.NoError(t, err, "GetBackgrounds() second call error")
	assert.Same(t, backgrounds, backgrounds2, "GetBackgrounds() did not return cached data")
}

func TestLoaderGetConditions(t *testing.T) {
	loader := NewLoader("../../data")

	conditions, err := loader.GetConditions()
	require.NoError(t, err, "GetConditions() error")
	require.NotNil(t, conditions, "GetConditions() returned nil")
	assert.NotEmpty(t, conditions.Conditions, "GetConditions() returned empty conditions list")

	// Verify some expected conditions exist
	expectedConditions := []string{"Blinded", "Charmed", "Frightened", "Poisoned", "Stunned"}
	condMap := make(map[string]*Condition)
	for i := range conditions.Conditions {
		condMap[conditions.Conditions[i].Name] = &conditions.Conditions[i]
	}

	for _, expected := range expectedConditions {
		cond, found := condMap[expected]
		assert.True(t, found, "expected condition '%s' not found", expected)
		if found {
			assert.NotEmpty(t, cond.Description, "condition '%s' has empty Description", cond.Name)
		}
	}

	// Test caching
	conditions2, err := loader.GetConditions()
	require.NoError(t, err, "GetConditions() second call error")
	assert.Same(t, conditions, conditions2, "GetConditions() did not return cached data")
}

func TestLoaderLoadAll(t *testing.T) {
	loader := NewLoader("../../data")

	err := loader.LoadAll()
	require.NoError(t, err, "LoadAll() error")

	// Verify all data is cached
	assert.NotNil(t, loader.races, "LoadAll() did not cache races")
	assert.NotNil(t, loader.classes, "LoadAll() did not cache classes")
	assert.NotNil(t, loader.spells, "LoadAll() did not cache spells")
	assert.NotNil(t, loader.backgrounds, "LoadAll() did not cache backgrounds")
	assert.NotNil(t, loader.conditions, "LoadAll() did not cache conditions")
	assert.NotNil(t, loader.feats, "LoadAll() did not cache feats")
}

func TestLoaderFindRaceByName(t *testing.T) {
	loader := NewLoader("../../data")

	tests := []struct {
		name      string
		raceName  string
		wantError bool
	}{
		{"find Human", "Human", false},
		{"find Elf", "Elf", false},
		{"find Dwarf", "Dwarf", false},
		{"find nonexistent", "NonexistentRace", true},
		{"case sensitive", "human", true}, // Should not find "human" (lowercase)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			race, err := loader.FindRaceByName(tt.raceName)
			if tt.wantError {
				assert.Error(t, err, "FindRaceByName() expected error for '%s'", tt.raceName)
			} else {
				require.NoError(t, err, "FindRaceByName() error")
				require.NotNil(t, race, "FindRaceByName() returned nil race")
				assert.Equal(t, tt.raceName, race.Name)
			}
		})
	}
}

func TestLoaderFindClassByName(t *testing.T) {
	loader := NewLoader("../../data")

	tests := []struct {
		name      string
		className string
		wantError bool
	}{
		{"find Barbarian", "Barbarian", false},
		{"find Wizard", "Wizard", false},
		{"find Rogue", "Rogue", false},
		{"find nonexistent", "NonexistentClass", true},
		{"case sensitive", "barbarian", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			class, err := loader.FindClassByName(tt.className)
			if tt.wantError {
				assert.Error(t, err, "FindClassByName() expected error for '%s'", tt.className)
			} else {
				require.NoError(t, err, "FindClassByName() error")
				require.NotNil(t, class, "FindClassByName() returned nil class")
				assert.Equal(t, tt.className, class.Name)
			}
		})
	}
}

func TestLoaderFindSpellByName(t *testing.T) {
	loader := NewLoader("../../data")

	tests := []struct {
		name      string
		spellName string
		wantError bool
	}{
		{"find Fireball", "Fireball", false},
		{"find Magic Missile", "Magic Missile", false},
		{"find nonexistent", "NonexistentSpell", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spell, err := loader.FindSpellByName(tt.spellName)
			if tt.wantError {
				assert.Error(t, err, "FindSpellByName() expected error for '%s'", tt.spellName)
			} else {
				require.NoError(t, err, "FindSpellByName() error")
				require.NotNil(t, spell, "FindSpellByName() returned nil spell")
				assert.Equal(t, tt.spellName, spell.Name)
			}
		})
	}
}

func TestLoaderFindBackgroundByName(t *testing.T) {
	loader := NewLoader("../../data")

	tests := []struct {
		name           string
		backgroundName string
		wantError      bool
	}{
		{"find Acolyte", "Acolyte", false},
		{"find Criminal", "Criminal", false},
		{"find nonexistent", "NonexistentBackground", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bg, err := loader.FindBackgroundByName(tt.backgroundName)
			if tt.wantError {
				assert.Error(t, err, "FindBackgroundByName() expected error for '%s'", tt.backgroundName)
			} else {
				require.NoError(t, err, "FindBackgroundByName() error")
				require.NotNil(t, bg, "FindBackgroundByName() returned nil background")
				assert.Equal(t, tt.backgroundName, bg.Name)
			}
		})
	}
}

func TestLoaderFindConditionByName(t *testing.T) {
	loader := NewLoader("../../data")

	tests := []struct {
		name          string
		conditionName string
		wantError     bool
	}{
		{"find Blinded", "Blinded", false},
		{"find Poisoned", "Poisoned", false},
		{"find nonexistent", "NonexistentCondition", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond, err := loader.FindConditionByName(tt.conditionName)
			if tt.wantError {
				assert.Error(t, err, "FindConditionByName() expected error for '%s'", tt.conditionName)
			} else {
				require.NoError(t, err, "FindConditionByName() error")
				require.NotNil(t, cond, "FindConditionByName() returned nil condition")
				assert.Equal(t, tt.conditionName, cond.Name)
			}
		})
	}
}

func TestLoaderGetFeats(t *testing.T) {
	loader := NewLoader("../../data")

	feats, err := loader.GetFeats()
	require.NoError(t, err, "GetFeats() error")
	require.NotNil(t, feats, "GetFeats() returned nil")
	assert.NotEmpty(t, feats.Feats, "GetFeats() returned empty feats list")

	// Verify some expected feats exist
	expectedFeats := []string{"Alert", "Lucky", "Tough", "Great Weapon Master", "Sentinel"}
	featMap := make(map[string]*Feat)
	for i := range feats.Feats {
		featMap[feats.Feats[i].Name] = &feats.Feats[i]
	}

	for _, expected := range expectedFeats {
		feat, found := featMap[expected]
		assert.True(t, found, "expected feat '%s' not found", expected)
		if found {
			assert.NotEmpty(t, feat.Description, "feat '%s' has empty Description", feat.Name)
			assert.NotEmpty(t, feat.Category, "feat '%s' has empty Category", feat.Name)
		}
	}

	// Verify origin feats have no prerequisite
	for _, feat := range feats.Feats {
		if feat.Category == FeatCategoryOrigin {
			assert.Empty(t, feat.Prerequisite, "Origin feat '%s' should have no prerequisite", feat.Name)
		}
	}

	// Test caching
	feats2, err := loader.GetFeats()
	require.NoError(t, err, "GetFeats() second call error")
	assert.Same(t, feats, feats2, "GetFeats() did not return cached data")
}

func TestLoaderFindFeatByName(t *testing.T) {
	loader := NewLoader("../../data")

	tests := []struct {
		name      string
		featName  string
		wantError bool
	}{
		{"find Alert", "Alert", false},
		{"find Tough", "Tough", false},
		{"find Great Weapon Master", "Great Weapon Master", false},
		{"find nonexistent", "NonexistentFeat", true},
		{"case sensitive", "alert", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feat, err := loader.FindFeatByName(tt.featName)
			if tt.wantError {
				assert.Error(t, err, "FindFeatByName() expected error for '%s'", tt.featName)
			} else {
				require.NoError(t, err, "FindFeatByName() error")
				require.NotNil(t, feat, "FindFeatByName() returned nil feat")
				assert.Equal(t, tt.featName, feat.Name)
			}
		})
	}
}

func TestLoaderClearCache(t *testing.T) {
	loader := NewLoader("../../data")

	// Load all data
	err := loader.LoadAll()
	require.NoError(t, err, "LoadAll() error")

	// Verify data is cached
	require.NotNil(t, loader.races, "races not cached")

	// Clear cache
	loader.ClearCache()

	// Verify cache is cleared
	assert.Nil(t, loader.races, "ClearCache() did not clear races")
	assert.Nil(t, loader.classes, "ClearCache() did not clear classes")
	assert.Nil(t, loader.spells, "ClearCache() did not clear spells")
	assert.Nil(t, loader.backgrounds, "ClearCache() did not clear backgrounds")
	assert.Nil(t, loader.conditions, "ClearCache() did not clear conditions")
}

func TestLoaderInvalidDirectory(t *testing.T) {
	loader := NewLoader("/nonexistent/directory")

	_, err := loader.GetRaces()
	assert.Error(t, err, "GetRaces() with invalid directory should return error")
}

func TestLoaderInvalidJSON(t *testing.T) {
	// Create a temporary directory with invalid JSON
	tmpDir, err := os.MkdirTemp("", "test-data-*")
	require.NoError(t, err, "failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	// Write invalid JSON
	invalidJSON := []byte(`{"races": [invalid json]}`)
	err = os.WriteFile(filepath.Join(tmpDir, "races.json"), invalidJSON, 0644)
	require.NoError(t, err, "failed to write invalid json")

	loader := NewLoader(tmpDir)
	_, err = loader.GetRaces()
	assert.Error(t, err, "GetRaces() with invalid JSON should return error")
}

func TestLoaderEmptyData(t *testing.T) {
	// Create a temporary directory with empty data
	tmpDir, err := os.MkdirTemp("", "test-data-*")
	require.NoError(t, err, "failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	// Write empty races JSON
	emptyJSON := []byte(`{"races": []}`)
	err = os.WriteFile(filepath.Join(tmpDir, "races.json"), emptyJSON, 0644)
	require.NoError(t, err, "failed to write empty json")

	loader := NewLoader(tmpDir)
	_, err = loader.GetRaces()
	assert.Error(t, err, "GetRaces() with empty data should return error")
}

func TestLoaderGetEquipment(t *testing.T) {
	loader := NewLoader("../../data")

	equipment, err := loader.GetEquipment()
	require.NoError(t, err, "GetEquipment() error")
	require.NotNil(t, equipment, "GetEquipment() returned nil")

	// Test weapons
	t.Run("weapons", func(t *testing.T) {
		allWeapons := equipment.Weapons.GetAllWeapons()
		assert.NotEmpty(t, allWeapons, "GetEquipment() returned no weapons")

		// Verify weapon categories have items
		assert.NotEmpty(t, equipment.Weapons.SimpleMelee, "no simple melee weapons")
		assert.NotEmpty(t, equipment.Weapons.SimpleRanged, "no simple ranged weapons")
		assert.NotEmpty(t, equipment.Weapons.MartialMelee, "no martial melee weapons")
		assert.NotEmpty(t, equipment.Weapons.MartialRanged, "no martial ranged weapons")

		// Verify weapon fields
		expectedWeapons := []string{"Dagger", "Longsword", "Shortbow"}
		weaponMap := make(map[string]*Weapon)
		for i := range allWeapons {
			weaponMap[allWeapons[i].Name] = &allWeapons[i]
		}

		for _, expected := range expectedWeapons {
			weapon, found := weaponMap[expected]
			assert.True(t, found, "expected weapon '%s' not found", expected)
			if found {
				assert.NotEmpty(t, weapon.ID, "weapon '%s' has empty ID", weapon.Name)
				assert.NotEmpty(t, weapon.Category, "weapon '%s' has empty Category", weapon.Name)
				assert.NotEmpty(t, weapon.SubCategory, "weapon '%s' has empty SubCategory", weapon.Name)
			}
		}
	})

	// Test armor
	t.Run("armor", func(t *testing.T) {
		allArmor := equipment.Armor.GetAllArmor()
		assert.NotEmpty(t, allArmor, "GetEquipment() returned no armor")

		// Verify armor categories have items
		assert.NotEmpty(t, equipment.Armor.Light, "no light armor")
		assert.NotEmpty(t, equipment.Armor.Medium, "no medium armor")
		assert.NotEmpty(t, equipment.Armor.Heavy, "no heavy armor")
		assert.NotEmpty(t, equipment.Armor.Shield, "no shields")

		// Verify armor fields
		expectedArmor := []string{"Leather Armor", "Chain Mail", "Shield"}
		armorMap := make(map[string]*ArmorItem)
		for i := range allArmor {
			armorMap[allArmor[i].Name] = &allArmor[i]
		}

		for _, expected := range expectedArmor {
			armor, found := armorMap[expected]
			assert.True(t, found, "expected armor '%s' not found", expected)
			if found {
				assert.NotEmpty(t, armor.ID, "armor '%s' has empty ID", armor.Name)
				assert.NotEmpty(t, armor.Category, "armor '%s' has empty Category", armor.Name)
			}
		}
	})

	// Test packs
	t.Run("packs", func(t *testing.T) {
		assert.NotEmpty(t, equipment.Packs, "GetEquipment() returned no packs")

		expectedPacks := []string{"Burglar's Pack", "Explorer's Pack", "Dungeoneer's Pack"}
		packMap := make(map[string]*Pack)
		for i := range equipment.Packs {
			packMap[equipment.Packs[i].Name] = &equipment.Packs[i]
		}

		for _, expected := range expectedPacks {
			pack, found := packMap[expected]
			assert.True(t, found, "expected pack '%s' not found", expected)
			if found {
				assert.NotEmpty(t, pack.ID, "pack '%s' has empty ID", pack.Name)
				assert.NotEmpty(t, pack.Contents, "pack '%s' has no contents", pack.Name)
				assert.NotEmpty(t, pack.Category, "pack '%s' has empty Category", pack.Name)
			}
		}
	})

	// Test gear
	t.Run("gear", func(t *testing.T) {
		assert.NotEmpty(t, equipment.Gear, "GetEquipment() returned no gear")

		expectedGear := []string{"Backpack", "Bedroll", "Rope (50 feet)"}
		gearMap := make(map[string]*Item)
		for i := range equipment.Gear {
			gearMap[equipment.Gear[i].Name] = &equipment.Gear[i]
		}

		for _, expected := range expectedGear {
			item, found := gearMap[expected]
			assert.True(t, found, "expected gear '%s' not found", expected)
			if found {
				assert.NotEmpty(t, item.ID, "gear '%s' has empty ID", item.Name)
				assert.NotEmpty(t, item.Category, "gear '%s' has empty Category", item.Name)
			}
		}
	})

	// Test tools
	t.Run("tools", func(t *testing.T) {
		assert.NotEmpty(t, equipment.Tools, "GetEquipment() returned no tools")

		// Verify tools have required fields
		for _, tool := range equipment.Tools {
			assert.NotEmpty(t, tool.ID, "tool '%s' has empty ID", tool.Name)
			assert.NotEmpty(t, tool.Category, "tool '%s' has empty Category", tool.Name)
		}

		// Check for some expected tools
		toolNames := make([]string, len(equipment.Tools))
		for i, tool := range equipment.Tools {
			toolNames[i] = tool.Name
		}
		assert.Contains(t, toolNames, "Thieves' Tools", "expected tool 'Thieves' Tools' not found")
	})

	// Test services
	t.Run("services", func(t *testing.T) {
		assert.NotEmpty(t, equipment.Services, "GetEquipment() returned no services")

		for _, service := range equipment.Services {
			assert.NotEmpty(t, service.ID, "service '%s' has empty ID", service.Name)
			assert.NotEmpty(t, service.Category, "service '%s' has empty Category", service.Name)
		}
	})

	// Test food/drink/lodging
	t.Run("foodDrinkLodging", func(t *testing.T) {
		assert.NotEmpty(t, equipment.FoodDrinkLodging, "GetEquipment() returned no food/drink/lodging items")

		for _, item := range equipment.FoodDrinkLodging {
			assert.NotEmpty(t, item.ID, "food/drink/lodging item '%s' has empty ID", item.Name)
			assert.NotEmpty(t, item.Category, "food/drink/lodging item '%s' has empty Category", item.Name)
		}
	})

	// Test transportation
	t.Run("transportation", func(t *testing.T) {
		assert.NotEmpty(t, equipment.Transportation, "GetEquipment() returned no transportation items")

		for _, item := range equipment.Transportation {
			assert.NotEmpty(t, item.ID, "transportation item '%s' has empty ID", item.Name)
			assert.NotEmpty(t, item.Category, "transportation item '%s' has empty Category", item.Name)
		}
	})

	// Test caching
	t.Run("caching", func(t *testing.T) {
		equipment2, err := loader.GetEquipment()
		require.NoError(t, err, "GetEquipment() second call error")
		assert.Same(t, equipment, equipment2, "GetEquipment() did not return cached data")
	})
}

func TestEquipmentWeaponHelpers(t *testing.T) {
	loader := NewLoader("../../data")

	equipment, err := loader.GetEquipment()
	require.NoError(t, err, "GetEquipment() error")

	t.Run("GetSimpleWeapons", func(t *testing.T) {
		simple := equipment.Weapons.GetSimpleWeapons()
		assert.NotEmpty(t, simple, "GetSimpleWeapons() returned empty list")

		expectedCount := len(equipment.Weapons.SimpleMelee) + len(equipment.Weapons.SimpleRanged)
		assert.Len(t, simple, expectedCount)
	})

	t.Run("GetMartialWeapons", func(t *testing.T) {
		martial := equipment.Weapons.GetMartialWeapons()
		assert.NotEmpty(t, martial, "GetMartialWeapons() returned empty list")

		expectedCount := len(equipment.Weapons.MartialMelee) + len(equipment.Weapons.MartialRanged)
		assert.Len(t, martial, expectedCount)
	})

	t.Run("GetAllWeapons", func(t *testing.T) {
		all := equipment.Weapons.GetAllWeapons()
		assert.NotEmpty(t, all, "GetAllWeapons() returned empty list")

		expectedCount := len(equipment.Weapons.SimpleMelee) + len(equipment.Weapons.SimpleRanged) +
			len(equipment.Weapons.MartialMelee) + len(equipment.Weapons.MartialRanged)
		assert.Len(t, all, expectedCount)
	})
}

func TestEquipmentArmorHelpers(t *testing.T) {
	loader := NewLoader("../../data")

	equipment, err := loader.GetEquipment()
	require.NoError(t, err, "GetEquipment() error")

	t.Run("GetAllArmor", func(t *testing.T) {
		all := equipment.Armor.GetAllArmor()
		assert.NotEmpty(t, all, "GetAllArmor() returned empty list")

		expectedCount := len(equipment.Armor.Light) + len(equipment.Armor.Medium) +
			len(equipment.Armor.Heavy) + len(equipment.Armor.Shield)
		assert.Len(t, all, expectedCount)
	})
}
