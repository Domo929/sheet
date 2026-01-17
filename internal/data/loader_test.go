package data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewLoader(t *testing.T) {
	t.Run("with custom directory", func(t *testing.T) {
		loader := NewLoader("/custom/path")
		if loader.dataDir != "/custom/path" {
			t.Errorf("expected dataDir to be '/custom/path', got '%s'", loader.dataDir)
		}
	})

	t.Run("with empty directory", func(t *testing.T) {
		loader := NewLoader("")
		if loader.dataDir != "./data" {
			t.Errorf("expected dataDir to be './data', got '%s'", loader.dataDir)
		}
	})
}

func TestLoaderGetRaces(t *testing.T) {
	loader := NewLoader("../../data")

	races, err := loader.GetRaces()
	if err != nil {
		t.Fatalf("GetRaces() error = %v", err)
	}

	if races == nil {
		t.Fatal("GetRaces() returned nil")
	}

	if len(races.Races) == 0 {
		t.Error("GetRaces() returned empty races list")
	}

	// Verify some expected races exist
	expectedRaces := []string{"Human", "Elf", "Dwarf", "Halfling"}
	raceMap := make(map[string]bool)
	for _, race := range races.Races {
		raceMap[race.Name] = true
	}

	for _, expected := range expectedRaces {
		found := false
		for _, race := range races.Races {
			if race.Name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected race '%s' not found", expected)
		}
	}

	// Test caching - second call should use cache
	races2, err := loader.GetRaces()
	if err != nil {
		t.Fatalf("GetRaces() second call error = %v", err)
	}

	// Verify it's the same pointer (cached)
	if races != races2 {
		t.Error("GetRaces() did not return cached data")
	}
}

func TestLoaderGetClasses(t *testing.T) {
	loader := NewLoader("../../data")

	classes, err := loader.GetClasses()
	if err != nil {
		t.Fatalf("GetClasses() error = %v", err)
	}

	if classes == nil {
		t.Fatal("GetClasses() returned nil")
	}

	if len(classes.Classes) == 0 {
		t.Error("GetClasses() returned empty classes list")
	}

	// Verify some expected classes exist
	expectedClasses := []string{"Barbarian", "Bard", "Cleric", "Druid", "Fighter", "Monk", "Paladin", "Ranger", "Rogue", "Sorcerer", "Warlock", "Wizard"}
	for _, expected := range expectedClasses {
		found := false
		for _, class := range classes.Classes {
			if class.Name == expected {
				found = true
				// Verify essential fields are populated
				if class.HitDice == "" {
					t.Errorf("class '%s' has empty HitDice", class.Name)
				}
				if len(class.SavingThrowProficiencies) == 0 {
					t.Errorf("class '%s' has no saving throw proficiencies", class.Name)
				}
				break
			}
		}
		if !found {
			t.Errorf("expected class '%s' not found", expected)
		}
	}

	// Test caching
	classes2, err := loader.GetClasses()
	if err != nil {
		t.Fatalf("GetClasses() second call error = %v", err)
	}

	if classes != classes2 {
		t.Error("GetClasses() did not return cached data")
	}
}

func TestLoaderGetSpells(t *testing.T) {
	loader := NewLoader("../../data")

	spells, err := loader.GetSpells()
	if err != nil {
		t.Fatalf("GetSpells() error = %v", err)
	}

	if spells == nil {
		t.Fatal("GetSpells() returned nil")
	}

	if len(spells.Spells) == 0 {
		t.Error("GetSpells() returned empty spells list")
	}

	// Verify some expected spells exist
	expectedSpells := []string{"Fireball", "Magic Missile", "Cure Wounds"}
	for _, expected := range expectedSpells {
		found := false
		for _, spell := range spells.Spells {
			if spell.Name == expected {
				found = true
				// Verify essential fields are populated
				if spell.School == "" {
					t.Errorf("spell '%s' has empty School", spell.Name)
				}
				if spell.Description == "" {
					t.Errorf("spell '%s' has empty Description", spell.Name)
				}
				break
			}
		}
		if !found {
			t.Errorf("expected spell '%s' not found", expected)
		}
	}

	// Test caching
	spells2, err := loader.GetSpells()
	if err != nil {
		t.Fatalf("GetSpells() second call error = %v", err)
	}

	if spells != spells2 {
		t.Error("GetSpells() did not return cached data")
	}
}

func TestLoaderGetBackgrounds(t *testing.T) {
	loader := NewLoader("../../data")

	backgrounds, err := loader.GetBackgrounds()
	if err != nil {
		t.Fatalf("GetBackgrounds() error = %v", err)
	}

	if backgrounds == nil {
		t.Fatal("GetBackgrounds() returned nil")
	}

	if len(backgrounds.Backgrounds) == 0 {
		t.Error("GetBackgrounds() returned empty backgrounds list")
	}

	// Verify some expected backgrounds exist
	expectedBackgrounds := []string{"Acolyte", "Criminal", "Noble", "Sage"}
	for _, expected := range expectedBackgrounds {
		found := false
		for _, bg := range backgrounds.Backgrounds {
			if bg.Name == expected {
				found = true
				// Verify essential fields are populated
				if bg.Description == "" {
					t.Errorf("background '%s' has empty Description", bg.Name)
				}
				if len(bg.SkillProficiencies) == 0 {
					t.Errorf("background '%s' has no skill proficiencies", bg.Name)
				}
				break
			}
		}
		if !found {
			t.Errorf("expected background '%s' not found", expected)
		}
	}

	// Test caching
	backgrounds2, err := loader.GetBackgrounds()
	if err != nil {
		t.Fatalf("GetBackgrounds() second call error = %v", err)
	}

	if backgrounds != backgrounds2 {
		t.Error("GetBackgrounds() did not return cached data")
	}
}

func TestLoaderGetConditions(t *testing.T) {
	loader := NewLoader("../../data")

	conditions, err := loader.GetConditions()
	if err != nil {
		t.Fatalf("GetConditions() error = %v", err)
	}

	if conditions == nil {
		t.Fatal("GetConditions() returned nil")
	}

	if len(conditions.Conditions) == 0 {
		t.Error("GetConditions() returned empty conditions list")
	}

	// Verify some expected conditions exist
	expectedConditions := []string{"Blinded", "Charmed", "Frightened", "Poisoned", "Stunned"}
	for _, expected := range expectedConditions {
		found := false
		for _, cond := range conditions.Conditions {
			if cond.Name == expected {
				found = true
				// Verify essential fields are populated
				if cond.Description == "" {
					t.Errorf("condition '%s' has empty Description", cond.Name)
				}
				break
			}
		}
		if !found {
			t.Errorf("expected condition '%s' not found", expected)
		}
	}

	// Test caching
	conditions2, err := loader.GetConditions()
	if err != nil {
		t.Fatalf("GetConditions() second call error = %v", err)
	}

	if conditions != conditions2 {
		t.Error("GetConditions() did not return cached data")
	}
}

func TestLoaderLoadAll(t *testing.T) {
	loader := NewLoader("../../data")

	err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	// Verify all data is cached
	if loader.races == nil {
		t.Error("LoadAll() did not cache races")
	}
	if loader.classes == nil {
		t.Error("LoadAll() did not cache classes")
	}
	if loader.spells == nil {
		t.Error("LoadAll() did not cache spells")
	}
	if loader.backgrounds == nil {
		t.Error("LoadAll() did not cache backgrounds")
	}
	if loader.conditions == nil {
		t.Error("LoadAll() did not cache conditions")
	}
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
				if err == nil {
					t.Errorf("FindRaceByName() expected error for '%s', got nil", tt.raceName)
				}
			} else {
				if err != nil {
					t.Errorf("FindRaceByName() error = %v", err)
				}
				if race == nil {
					t.Error("FindRaceByName() returned nil race")
				}
				if race != nil && race.Name != tt.raceName {
					t.Errorf("FindRaceByName() got name '%s', want '%s'", race.Name, tt.raceName)
				}
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
				if err == nil {
					t.Errorf("FindClassByName() expected error for '%s', got nil", tt.className)
				}
			} else {
				if err != nil {
					t.Errorf("FindClassByName() error = %v", err)
				}
				if class == nil {
					t.Error("FindClassByName() returned nil class")
				}
				if class != nil && class.Name != tt.className {
					t.Errorf("FindClassByName() got name '%s', want '%s'", class.Name, tt.className)
				}
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
				if err == nil {
					t.Errorf("FindSpellByName() expected error for '%s', got nil", tt.spellName)
				}
			} else {
				if err != nil {
					t.Errorf("FindSpellByName() error = %v", err)
				}
				if spell == nil {
					t.Error("FindSpellByName() returned nil spell")
				}
				if spell != nil && spell.Name != tt.spellName {
					t.Errorf("FindSpellByName() got name '%s', want '%s'", spell.Name, tt.spellName)
				}
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
				if err == nil {
					t.Errorf("FindBackgroundByName() expected error for '%s', got nil", tt.backgroundName)
				}
			} else {
				if err != nil {
					t.Errorf("FindBackgroundByName() error = %v", err)
				}
				if bg == nil {
					t.Error("FindBackgroundByName() returned nil background")
				}
				if bg != nil && bg.Name != tt.backgroundName {
					t.Errorf("FindBackgroundByName() got name '%s', want '%s'", bg.Name, tt.backgroundName)
				}
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
				if err == nil {
					t.Errorf("FindConditionByName() expected error for '%s', got nil", tt.conditionName)
				}
			} else {
				if err != nil {
					t.Errorf("FindConditionByName() error = %v", err)
				}
				if cond == nil {
					t.Error("FindConditionByName() returned nil condition")
				}
				if cond != nil && cond.Name != tt.conditionName {
					t.Errorf("FindConditionByName() got name '%s', want '%s'", cond.Name, tt.conditionName)
				}
			}
		})
	}
}

func TestLoaderClearCache(t *testing.T) {
	loader := NewLoader("../../data")

	// Load all data
	err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	// Verify data is cached
	if loader.races == nil {
		t.Fatal("races not cached")
	}

	// Clear cache
	loader.ClearCache()

	// Verify cache is cleared
	if loader.races != nil {
		t.Error("ClearCache() did not clear races")
	}
	if loader.classes != nil {
		t.Error("ClearCache() did not clear classes")
	}
	if loader.spells != nil {
		t.Error("ClearCache() did not clear spells")
	}
	if loader.backgrounds != nil {
		t.Error("ClearCache() did not clear backgrounds")
	}
	if loader.conditions != nil {
		t.Error("ClearCache() did not clear conditions")
	}
}

func TestLoaderInvalidDirectory(t *testing.T) {
	loader := NewLoader("/nonexistent/directory")

	_, err := loader.GetRaces()
	if err == nil {
		t.Error("GetRaces() with invalid directory should return error")
	}
}

func TestLoaderInvalidJSON(t *testing.T) {
	// Create a temporary directory with invalid JSON
	tmpDir, err := os.MkdirTemp("", "test-data-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write invalid JSON
	invalidJSON := []byte(`{"races": [invalid json]}`)
	err = os.WriteFile(filepath.Join(tmpDir, "races.json"), invalidJSON, 0644)
	if err != nil {
		t.Fatalf("failed to write invalid json: %v", err)
	}

	loader := NewLoader(tmpDir)
	_, err = loader.GetRaces()
	if err == nil {
		t.Error("GetRaces() with invalid JSON should return error")
	}
}

func TestLoaderEmptyData(t *testing.T) {
	// Create a temporary directory with empty data
	tmpDir, err := os.MkdirTemp("", "test-data-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write empty races JSON
	emptyJSON := []byte(`{"races": []}`)
	err = os.WriteFile(filepath.Join(tmpDir, "races.json"), emptyJSON, 0644)
	if err != nil {
		t.Fatalf("failed to write empty json: %v", err)
	}

	loader := NewLoader(tmpDir)
	_, err = loader.GetRaces()
	if err == nil {
		t.Error("GetRaces() with empty data should return error")
	}
}

func TestLoaderGetEquipment(t *testing.T) {
	loader := NewLoader("../../data")

	equipment, err := loader.GetEquipment()
	if err != nil {
		t.Fatalf("GetEquipment() error = %v", err)
	}

	if equipment == nil {
		t.Fatal("GetEquipment() returned nil")
	}

	// Test weapons
	t.Run("weapons", func(t *testing.T) {
		allWeapons := equipment.Weapons.GetAllWeapons()
		if len(allWeapons) == 0 {
			t.Error("GetEquipment() returned no weapons")
		}

		// Verify weapon categories have items
		if len(equipment.Weapons.SimpleMelee) == 0 {
			t.Error("no simple melee weapons")
		}
		if len(equipment.Weapons.SimpleRanged) == 0 {
			t.Error("no simple ranged weapons")
		}
		if len(equipment.Weapons.MartialMelee) == 0 {
			t.Error("no martial melee weapons")
		}
		if len(equipment.Weapons.MartialRanged) == 0 {
			t.Error("no martial ranged weapons")
		}

		// Verify weapon fields
		expectedWeapons := []string{"Dagger", "Longsword", "Shortbow"}
		for _, expected := range expectedWeapons {
			found := false
			for _, weapon := range allWeapons {
				if weapon.Name == expected {
					found = true
					if weapon.ID == "" {
						t.Errorf("weapon '%s' has empty ID", weapon.Name)
					}
					if weapon.Category == "" {
						t.Errorf("weapon '%s' has empty Category", weapon.Name)
					}
					if weapon.SubCategory == "" {
						t.Errorf("weapon '%s' has empty SubCategory", weapon.Name)
					}
					break
				}
			}
			if !found {
				t.Errorf("expected weapon '%s' not found", expected)
			}
		}
	})

	// Test armor
	t.Run("armor", func(t *testing.T) {
		allArmor := equipment.Armor.GetAllArmor()
		if len(allArmor) == 0 {
			t.Error("GetEquipment() returned no armor")
		}

		// Verify armor categories have items
		if len(equipment.Armor.Light) == 0 {
			t.Error("no light armor")
		}
		if len(equipment.Armor.Medium) == 0 {
			t.Error("no medium armor")
		}
		if len(equipment.Armor.Heavy) == 0 {
			t.Error("no heavy armor")
		}
		if len(equipment.Armor.Shield) == 0 {
			t.Error("no shields")
		}

		// Verify armor fields
		expectedArmor := []string{"Leather Armor", "Chain Mail", "Shield"}
		for _, expected := range expectedArmor {
			found := false
			for _, armor := range allArmor {
				if armor.Name == expected {
					found = true
					if armor.ID == "" {
						t.Errorf("armor '%s' has empty ID", armor.Name)
					}
					if armor.Category == "" {
						t.Errorf("armor '%s' has empty Category", armor.Name)
					}
					break
				}
			}
			if !found {
				t.Errorf("expected armor '%s' not found", expected)
			}
		}
	})

	// Test packs
	t.Run("packs", func(t *testing.T) {
		if len(equipment.Packs) == 0 {
			t.Error("GetEquipment() returned no packs")
		}

		expectedPacks := []string{"Burglar's Pack", "Explorer's Pack", "Dungeoneer's Pack"}
		for _, expected := range expectedPacks {
			found := false
			for _, pack := range equipment.Packs {
				if pack.Name == expected {
					found = true
					if pack.ID == "" {
						t.Errorf("pack '%s' has empty ID", pack.Name)
					}
					if len(pack.Contents) == 0 {
						t.Errorf("pack '%s' has no contents", pack.Name)
					}
					if pack.Category == "" {
						t.Errorf("pack '%s' has empty Category", pack.Name)
					}
					break
				}
			}
			if !found {
				t.Errorf("expected pack '%s' not found", expected)
			}
		}
	})

	// Test gear
	t.Run("gear", func(t *testing.T) {
		if len(equipment.Gear) == 0 {
			t.Error("GetEquipment() returned no gear")
		}

		expectedGear := []string{"Backpack", "Bedroll", "Rope (50 feet)"}
		for _, expected := range expectedGear {
			found := false
			for _, item := range equipment.Gear {
				if item.Name == expected {
					found = true
					if item.ID == "" {
						t.Errorf("gear '%s' has empty ID", item.Name)
					}
					if item.Category == "" {
						t.Errorf("gear '%s' has empty Category", item.Name)
					}
					break
				}
			}
			if !found {
				t.Errorf("expected gear '%s' not found", expected)
			}
		}
	})

	// Test tools
	t.Run("tools", func(t *testing.T) {
		if len(equipment.Tools) == 0 {
			t.Error("GetEquipment() returned no tools")
		}

		// Verify tools have required fields
		for _, tool := range equipment.Tools {
			if tool.ID == "" {
				t.Errorf("tool '%s' has empty ID", tool.Name)
			}
			if tool.Category == "" {
				t.Errorf("tool '%s' has empty Category", tool.Name)
			}
		}

		// Check for some expected tools
		expectedTools := []string{"Thieves' Tools"}
		for _, expected := range expectedTools {
			found := false
			for _, tool := range equipment.Tools {
				if tool.Name == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected tool '%s' not found", expected)
			}
		}
	})

	// Test services
	t.Run("services", func(t *testing.T) {
		if len(equipment.Services) == 0 {
			t.Error("GetEquipment() returned no services")
		}

		for _, service := range equipment.Services {
			if service.ID == "" {
				t.Errorf("service '%s' has empty ID", service.Name)
			}
			if service.Category == "" {
				t.Errorf("service '%s' has empty Category", service.Name)
			}
		}
	})

	// Test food/drink/lodging
	t.Run("foodDrinkLodging", func(t *testing.T) {
		if len(equipment.FoodDrinkLodging) == 0 {
			t.Error("GetEquipment() returned no food/drink/lodging items")
		}

		for _, item := range equipment.FoodDrinkLodging {
			if item.ID == "" {
				t.Errorf("food/drink/lodging item '%s' has empty ID", item.Name)
			}
			if item.Category == "" {
				t.Errorf("food/drink/lodging item '%s' has empty Category", item.Name)
			}
		}
	})

	// Test transportation
	t.Run("transportation", func(t *testing.T) {
		if len(equipment.Transportation) == 0 {
			t.Error("GetEquipment() returned no transportation items")
		}

		for _, item := range equipment.Transportation {
			if item.ID == "" {
				t.Errorf("transportation item '%s' has empty ID", item.Name)
			}
			if item.Category == "" {
				t.Errorf("transportation item '%s' has empty Category", item.Name)
			}
		}
	})

	// Test caching
	t.Run("caching", func(t *testing.T) {
		equipment2, err := loader.GetEquipment()
		if err != nil {
			t.Fatalf("GetEquipment() second call error = %v", err)
		}

		if equipment != equipment2 {
			t.Error("GetEquipment() did not return cached data")
		}
	})
}

func TestEquipmentWeaponHelpers(t *testing.T) {
	loader := NewLoader("../../data")

	equipment, err := loader.GetEquipment()
	if err != nil {
		t.Fatalf("GetEquipment() error = %v", err)
	}

	t.Run("GetSimpleWeapons", func(t *testing.T) {
		simple := equipment.Weapons.GetSimpleWeapons()
		if len(simple) == 0 {
			t.Error("GetSimpleWeapons() returned empty list")
		}

		expectedCount := len(equipment.Weapons.SimpleMelee) + len(equipment.Weapons.SimpleRanged)
		if len(simple) != expectedCount {
			t.Errorf("GetSimpleWeapons() returned %d weapons, expected %d", len(simple), expectedCount)
		}
	})

	t.Run("GetMartialWeapons", func(t *testing.T) {
		martial := equipment.Weapons.GetMartialWeapons()
		if len(martial) == 0 {
			t.Error("GetMartialWeapons() returned empty list")
		}

		expectedCount := len(equipment.Weapons.MartialMelee) + len(equipment.Weapons.MartialRanged)
		if len(martial) != expectedCount {
			t.Errorf("GetMartialWeapons() returned %d weapons, expected %d", len(martial), expectedCount)
		}
	})

	t.Run("GetAllWeapons", func(t *testing.T) {
		all := equipment.Weapons.GetAllWeapons()
		if len(all) == 0 {
			t.Error("GetAllWeapons() returned empty list")
		}

		expectedCount := len(equipment.Weapons.SimpleMelee) + len(equipment.Weapons.SimpleRanged) +
			len(equipment.Weapons.MartialMelee) + len(equipment.Weapons.MartialRanged)
		if len(all) != expectedCount {
			t.Errorf("GetAllWeapons() returned %d weapons, expected %d", len(all), expectedCount)
		}
	})
}

func TestEquipmentArmorHelpers(t *testing.T) {
	loader := NewLoader("../../data")

	equipment, err := loader.GetEquipment()
	if err != nil {
		t.Fatalf("GetEquipment() error = %v", err)
	}

	t.Run("GetAllArmor", func(t *testing.T) {
		all := equipment.Armor.GetAllArmor()
		if len(all) == 0 {
			t.Error("GetAllArmor() returned empty list")
		}

		expectedCount := len(equipment.Armor.Light) + len(equipment.Armor.Medium) +
			len(equipment.Armor.Heavy) + len(equipment.Armor.Shield)
		if len(all) != expectedCount {
			t.Errorf("GetAllArmor() returned %d items, expected %d", len(all), expectedCount)
		}
	})
}
