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
