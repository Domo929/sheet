package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Domo929/sheet/internal/models"
)

// createTestStorage creates a temporary storage for testing.
func createTestStorage(t *testing.T) (*CharacterStorage, string) {
	t.Helper()

	tmpDir := t.TempDir()
	storage, err := NewCharacterStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create test storage: %v", err)
	}

	return storage, tmpDir
}

// createTestCharacter creates a test character.
func createTestCharacter(name, race, class string) *models.Character {
	char := models.NewCharacter("test-id", name, race, class)
	return char
}

func TestNewCharacterStorage(t *testing.T) {
	tmpDir := t.TempDir()

	storage, err := NewCharacterStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewCharacterStorage() error = %v", err)
	}

	if storage == nil {
		t.Fatal("NewCharacterStorage() returned nil storage")
	}

	if storage.GetBaseDir() != tmpDir {
		t.Errorf("GetBaseDir() = %s, want %s", storage.GetBaseDir(), tmpDir)
	}

	// Check that directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Character directory was not created")
	}
}

func TestNewCharacterStorageDefault(t *testing.T) {
	storage, err := NewCharacterStorage("")
	if err != nil {
		t.Fatalf("NewCharacterStorage(\"\") error = %v", err)
	}

	// Should use default directory
	homeDir, _ := os.UserHomeDir()
	expectedDir := filepath.Join(homeDir, ".dnd_sheet", "characters")

	if storage.GetBaseDir() != expectedDir {
		t.Errorf("GetBaseDir() = %s, want %s", storage.GetBaseDir(), expectedDir)
	}
}

func TestSaveAndLoad(t *testing.T) {
	storage, _ := createTestStorage(t)

	// Create test character
	char := createTestCharacter("TestHero", "Human", "Fighter")
	char.AbilityScores.Strength.Base = 16
	char.AbilityScores.Dexterity.Base = 14
	char.Info.Level = 5

	// Save character
	path, err := storage.Save(char)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if path == "" {
		t.Error("Save() returned empty path")
	}

	// Load character
	loaded, err := storage.Load("TestHero")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify loaded character
	if loaded.Info.Name != "TestHero" {
		t.Errorf("Loaded character name = %s, want TestHero", loaded.Info.Name)
	}
	if loaded.Info.Race != "Human" {
		t.Errorf("Loaded character race = %s, want Human", loaded.Info.Race)
	}
	if loaded.Info.Class != "Fighter" {
		t.Errorf("Loaded character class = %s, want Fighter", loaded.Info.Class)
	}
	if loaded.Info.Level != 5 {
		t.Errorf("Loaded character level = %d, want 5", loaded.Info.Level)
	}
	if loaded.AbilityScores.Strength.Base != 16 {
		t.Errorf("Loaded character STR = %d, want 16", loaded.AbilityScores.Strength.Base)
	}
}

func TestSaveNilCharacter(t *testing.T) {
	storage, _ := createTestStorage(t)

	_, err := storage.Save(nil)
	if err == nil {
		t.Error("Save(nil) should return error")
	}
}

func TestSaveEmptyName(t *testing.T) {
	storage, _ := createTestStorage(t)

	char := createTestCharacter("", "Human", "Fighter")

	_, err := storage.Save(char)
	if err != ErrInvalidCharacterName {
		t.Errorf("Save() error = %v, want ErrInvalidCharacterName", err)
	}
}

func TestLoadNonexistent(t *testing.T) {
	storage, _ := createTestStorage(t)

	_, err := storage.Load("Nonexistent")
	if err != ErrCharacterNotFound {
		t.Errorf("Load() error = %v, want ErrCharacterNotFound", err)
	}
}

func TestLoadEmptyName(t *testing.T) {
	storage, _ := createTestStorage(t)

	_, err := storage.Load("")
	if err != ErrInvalidCharacterName {
		t.Errorf("Load() error = %v, want ErrInvalidCharacterName", err)
	}
}

func TestDelete(t *testing.T) {
	storage, _ := createTestStorage(t)

	// Create and save character
	char := createTestCharacter("ToDelete", "Elf", "Wizard")
	_, err := storage.Save(char)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify it exists
	if !storage.Exists("ToDelete") {
		t.Error("Character should exist before delete")
	}

	// Delete character
	err = storage.Delete("ToDelete")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's gone
	if storage.Exists("ToDelete") {
		t.Error("Character should not exist after delete")
	}
}

func TestDeleteNonexistent(t *testing.T) {
	storage, _ := createTestStorage(t)

	err := storage.Delete("Nonexistent")
	if err != ErrCharacterNotFound {
		t.Errorf("Delete() error = %v, want ErrCharacterNotFound", err)
	}
}

func TestDeleteEmptyName(t *testing.T) {
	storage, _ := createTestStorage(t)

	err := storage.Delete("")
	if err != ErrInvalidCharacterName {
		t.Errorf("Delete() error = %v, want ErrInvalidCharacterName", err)
	}
}

func TestExists(t *testing.T) {
	storage, _ := createTestStorage(t)

	// Should not exist initially
	if storage.Exists("TestChar") {
		t.Error("Exists() = true, want false for nonexistent character")
	}

	// Create and save character
	char := createTestCharacter("TestChar", "Dwarf", "Cleric")
	_, err := storage.Save(char)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Should exist now
	if !storage.Exists("TestChar") {
		t.Error("Exists() = false, want true for saved character")
	}
}

func TestExistsEmptyName(t *testing.T) {
	storage, _ := createTestStorage(t)

	if storage.Exists("") {
		t.Error("Exists(\"\") should return false")
	}
}

func TestList(t *testing.T) {
	storage, _ := createTestStorage(t)

	// Create and save multiple characters
	chars := []*models.Character{
		createTestCharacter("Hero1", "Human", "Fighter"),
		createTestCharacter("Hero2", "Elf", "Wizard"),
		createTestCharacter("Hero3", "Dwarf", "Cleric"),
	}

	for _, char := range chars {
		if _, err := storage.Save(char); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
	}

	// List characters
	list, err := storage.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(list) != 3 {
		t.Errorf("List() returned %d characters, want 3", len(list))
	}

	// Verify all characters are in the list
	names := make(map[string]bool)
	for _, info := range list {
		names[info.Name] = true
	}

	for _, char := range chars {
		if !names[char.Info.Name] {
			t.Errorf("Character %s not found in list", char.Info.Name)
		}
	}
}

func TestListEmpty(t *testing.T) {
	storage, _ := createTestStorage(t)

	list, err := storage.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(list) != 0 {
		t.Errorf("List() returned %d characters, want 0", len(list))
	}
}

func TestRename(t *testing.T) {
	storage, _ := createTestStorage(t)

	// Create and save character
	char := createTestCharacter("OldName", "Halfling", "Rogue")
	_, err := storage.Save(char)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Rename
	err = storage.Rename("OldName", "NewName")
	if err != nil {
		t.Fatalf("Rename() error = %v", err)
	}

	// Old name should not exist
	if storage.Exists("OldName") {
		t.Error("Old character name should not exist after rename")
	}

	// New name should exist
	if !storage.Exists("NewName") {
		t.Error("New character name should exist after rename")
	}

	// Should be able to load with new name
	loaded, err := storage.Load("NewName")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Data should be preserved (though name in file still says OldName)
	if loaded.Info.Class != "Rogue" {
		t.Errorf("Loaded character class = %s, want Rogue", loaded.Info.Class)
	}
}

func TestRenameNonexistent(t *testing.T) {
	storage, _ := createTestStorage(t)

	err := storage.Rename("Nonexistent", "NewName")
	if err != ErrCharacterNotFound {
		t.Errorf("Rename() error = %v, want ErrCharacterNotFound", err)
	}
}

func TestRenameToExisting(t *testing.T) {
	storage, _ := createTestStorage(t)

	// Create two characters
	char1 := createTestCharacter("Char1", "Human", "Fighter")
	char2 := createTestCharacter("Char2", "Elf", "Wizard")

	storage.Save(char1)
	storage.Save(char2)

	// Try to rename Char1 to Char2 (should fail)
	err := storage.Rename("Char1", "Char2")
	if err != ErrCharacterExists {
		t.Errorf("Rename() error = %v, want ErrCharacterExists", err)
	}
}

func TestRenameEmptyNames(t *testing.T) {
	storage, _ := createTestStorage(t)

	tests := []struct {
		name    string
		oldName string
		newName string
	}{
		{"empty old name", "", "NewName"},
		{"empty new name", "OldName", ""},
		{"both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.Rename(tt.oldName, tt.newName)
			if err != ErrInvalidCharacterName {
				t.Errorf("Rename() error = %v, want ErrInvalidCharacterName", err)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"SimpleHero", "SimpleHero"},
		{"Hero With Spaces", "Hero_With_Spaces"},
		{"Hero/With\\Slashes", "HeroWithSlashes"},
		{"Hero:With*Invalid?Chars", "HeroWithInvalidChars"},
		{"Hero<>With|Pipes", "HeroWithPipes"},
		{"", "unnamed"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAutoSave(t *testing.T) {
	storage, _ := createTestStorage(t)

	char := createTestCharacter("AutoSaved", "Tiefling", "Warlock")
	oldUpdatedAt := char.UpdatedAt

	// AutoSave should update the timestamp
	err := storage.AutoSave(char)
	if err != nil {
		t.Fatalf("AutoSave() error = %v", err)
	}

	if char.UpdatedAt.Equal(oldUpdatedAt) {
		t.Error("AutoSave() should update the UpdatedAt timestamp")
	}

	// Verify character was saved
	if !storage.Exists("AutoSaved") {
		t.Error("Character should exist after AutoSave()")
	}
}

func TestLoadByPath(t *testing.T) {
	storage, tmpDir := createTestStorage(t)

	// Create and save character
	char := createTestCharacter("PathTest", "Dragonborn", "Paladin")
	path, err := storage.Save(char)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load by path
	loaded, err := storage.LoadByPath(path)
	if err != nil {
		t.Fatalf("LoadByPath() error = %v", err)
	}

	if loaded.Info.Name != "PathTest" {
		t.Errorf("Loaded character name = %s, want PathTest", loaded.Info.Name)
	}

	// Try loading non-existent path
	_, err = storage.LoadByPath(filepath.Join(tmpDir, "nonexistent.json"))
	if err != ErrCharacterNotFound {
		t.Errorf("LoadByPath() error = %v, want ErrCharacterNotFound", err)
	}
}

func TestCharacterInfoList(t *testing.T) {
	storage, _ := createTestStorage(t)

	// Create character with specific level
	char := createTestCharacter("InfoTest", "Gnome", "Bard")
	char.Info.Level = 7
	storage.Save(char)

	// Get list
	list, err := storage.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("List() returned %d characters, want 1", len(list))
	}

	info := list[0]
	if info.Name != "InfoTest" {
		t.Errorf("CharacterInfo.Name = %s, want InfoTest", info.Name)
	}
	if info.Race != "Gnome" {
		t.Errorf("CharacterInfo.Race = %s, want Gnome", info.Race)
	}
	if info.Class != "Bard" {
		t.Errorf("CharacterInfo.Class = %s, want Bard", info.Class)
	}
	if info.Level != 7 {
		t.Errorf("CharacterInfo.Level = %d, want 7", info.Level)
	}
	if info.Path == "" {
		t.Error("CharacterInfo.Path should not be empty")
	}
}
