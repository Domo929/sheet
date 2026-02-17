package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Domo929/sheet/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestStorage creates a temporary storage for testing.
func createTestStorage(t *testing.T) (*CharacterStorage, string) {
	t.Helper()

	tmpDir := t.TempDir()
	storage, err := NewCharacterStorage(tmpDir)
	require.NoError(t, err, "Failed to create test storage")

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
	require.NoError(t, err, "NewCharacterStorage() error")
	require.NotNil(t, storage, "NewCharacterStorage() returned nil storage")

	assert.Equal(t, tmpDir, storage.GetBaseDir())

	// Check that directory was created
	_, err = os.Stat(tmpDir)
	assert.False(t, os.IsNotExist(err), "Character directory was not created")
}

func TestNewCharacterStorageDefault(t *testing.T) {
	storage, err := NewCharacterStorage("")
	require.NoError(t, err, "NewCharacterStorage(\"\") error")

	// Should use default directory
	homeDir, _ := os.UserHomeDir()
	expectedDir := filepath.Join(homeDir, ".dnd_sheet", "characters")

	assert.Equal(t, expectedDir, storage.GetBaseDir())
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
	require.NoError(t, err, "Save() error")
	assert.NotEmpty(t, path, "Save() returned empty path")

	// Load character
	loaded, err := storage.Load("TestHero")
	require.NoError(t, err, "Load() error")

	// Verify loaded character
	assert.Equal(t, "TestHero", loaded.Info.Name)
	assert.Equal(t, "Human", loaded.Info.Race)
	assert.Equal(t, "Fighter", loaded.Info.Class)
	assert.Equal(t, 5, loaded.Info.Level)
	assert.Equal(t, 16, loaded.AbilityScores.Strength.Base)
}

func TestSaveNilCharacter(t *testing.T) {
	storage, _ := createTestStorage(t)

	_, err := storage.Save(nil)
	assert.Error(t, err, "Save(nil) should return error")
}

func TestSaveEmptyName(t *testing.T) {
	storage, _ := createTestStorage(t)

	char := createTestCharacter("", "Human", "Fighter")

	_, err := storage.Save(char)
	assert.Equal(t, ErrInvalidCharacterName, err)
}

func TestLoadNonexistent(t *testing.T) {
	storage, _ := createTestStorage(t)

	_, err := storage.Load("Nonexistent")
	assert.Equal(t, ErrCharacterNotFound, err)
}

func TestLoadEmptyName(t *testing.T) {
	storage, _ := createTestStorage(t)

	_, err := storage.Load("")
	assert.Equal(t, ErrInvalidCharacterName, err)
}

func TestDelete(t *testing.T) {
	storage, _ := createTestStorage(t)

	// Create and save character
	char := createTestCharacter("ToDelete", "Elf", "Wizard")
	_, err := storage.Save(char)
	require.NoError(t, err, "Save() error")

	// Verify it exists
	assert.True(t, storage.Exists("ToDelete"), "Character should exist before delete")

	// Delete character
	err = storage.Delete("ToDelete")
	require.NoError(t, err, "Delete() error")

	// Verify it's gone
	assert.False(t, storage.Exists("ToDelete"), "Character should not exist after delete")
}

func TestDeleteNonexistent(t *testing.T) {
	storage, _ := createTestStorage(t)

	err := storage.Delete("Nonexistent")
	assert.Equal(t, ErrCharacterNotFound, err)
}

func TestDeleteEmptyName(t *testing.T) {
	storage, _ := createTestStorage(t)

	err := storage.Delete("")
	assert.Equal(t, ErrInvalidCharacterName, err)
}

func TestExists(t *testing.T) {
	storage, _ := createTestStorage(t)

	// Should not exist initially
	assert.False(t, storage.Exists("TestChar"), "Exists() = true, want false for nonexistent character")

	// Create and save character
	char := createTestCharacter("TestChar", "Dwarf", "Cleric")
	_, err := storage.Save(char)
	require.NoError(t, err, "Save() error")

	// Should exist now
	assert.True(t, storage.Exists("TestChar"), "Exists() = false, want true for saved character")
}

func TestExistsEmptyName(t *testing.T) {
	storage, _ := createTestStorage(t)

	assert.False(t, storage.Exists(""), "Exists(\"\") should return false")
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
		_, err := storage.Save(char)
		require.NoError(t, err, "Save() error")
	}

	// List characters
	list, err := storage.List()
	require.NoError(t, err, "List() error")

	assert.Len(t, list, 3)

	// Verify all characters are in the list
	names := make(map[string]bool)
	for _, info := range list {
		names[info.Name] = true
	}

	for _, char := range chars {
		assert.True(t, names[char.Info.Name], "Character %s not found in list", char.Info.Name)
	}
}

func TestListEmpty(t *testing.T) {
	storage, _ := createTestStorage(t)

	list, err := storage.List()
	require.NoError(t, err, "List() error")

	assert.Empty(t, list)
}

func TestRename(t *testing.T) {
	storage, _ := createTestStorage(t)

	// Create and save character
	char := createTestCharacter("OldName", "Halfling", "Rogue")
	_, err := storage.Save(char)
	require.NoError(t, err, "Save() error")

	// Rename
	err = storage.Rename("OldName", "NewName")
	require.NoError(t, err, "Rename() error")

	// Old name should not exist
	assert.False(t, storage.Exists("OldName"), "Old character name should not exist after rename")

	// New name should exist
	assert.True(t, storage.Exists("NewName"), "New character name should exist after rename")

	// Should be able to load with new name
	loaded, err := storage.Load("NewName")
	require.NoError(t, err, "Load() error")

	// Data should be preserved (though name in file still says OldName)
	assert.Equal(t, "Rogue", loaded.Info.Class)
}

func TestRenameNonexistent(t *testing.T) {
	storage, _ := createTestStorage(t)

	err := storage.Rename("Nonexistent", "NewName")
	assert.Equal(t, ErrCharacterNotFound, err)
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
	assert.Equal(t, ErrCharacterExists, err)
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
			assert.Equal(t, ErrInvalidCharacterName, err)
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
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAutoSave(t *testing.T) {
	storage, _ := createTestStorage(t)

	char := createTestCharacter("AutoSaved", "Tiefling", "Warlock")
	oldUpdatedAt := char.UpdatedAt

	// Ensure enough time elapses for the timestamp to differ.
	// Windows system clock granularity can be ~15ms, so a brief sleep
	// guarantees MarkUpdated() produces a distinct timestamp.
	time.Sleep(20 * time.Millisecond)

	// AutoSave should update the timestamp
	err := storage.AutoSave(char)
	require.NoError(t, err, "AutoSave() error")

	assert.True(t, char.UpdatedAt.After(oldUpdatedAt), "AutoSave() should update the UpdatedAt timestamp")

	// Verify character was saved
	assert.True(t, storage.Exists("AutoSaved"), "Character should exist after AutoSave()")
}

func TestLoadByPath(t *testing.T) {
	storage, tmpDir := createTestStorage(t)

	// Create and save character
	char := createTestCharacter("PathTest", "Dragonborn", "Paladin")
	path, err := storage.Save(char)
	require.NoError(t, err, "Save() error")

	// Load by path
	loaded, err := storage.LoadByPath(path)
	require.NoError(t, err, "LoadByPath() error")

	assert.Equal(t, "PathTest", loaded.Info.Name)

	// Try loading non-existent path
	_, err = storage.LoadByPath(filepath.Join(tmpDir, "nonexistent.json"))
	assert.Equal(t, ErrCharacterNotFound, err)
}

func TestSaveAtomicity(t *testing.T) {
	store, _ := createTestStorage(t)
	char := createTestCharacter("AtomicHero", "Human", "Fighter")

	// First save
	path, err := store.Save(char)
	require.NoError(t, err)

	// Verify no temp files remain after successful save
	matches, _ := filepath.Glob(path + ".tmp.*")
	assert.Empty(t, matches, "temp files should not exist after successful save")

	// Verify the saved file has valid content
	loaded, err := store.Load(char.Info.Name)
	require.NoError(t, err)
	assert.Equal(t, char.Info.Name, loaded.Info.Name)
}

func TestCharacterInfoList(t *testing.T) {
	storage, _ := createTestStorage(t)

	// Create character with specific level
	char := createTestCharacter("InfoTest", "Gnome", "Bard")
	char.Info.Level = 7
	storage.Save(char)

	// Get list
	list, err := storage.List()
	require.NoError(t, err, "List() error")
	require.Len(t, list, 1)

	info := list[0]
	assert.Equal(t, "InfoTest", info.Name)
	assert.Equal(t, "Gnome", info.Race)
	assert.Equal(t, "Bard", info.Class)
	assert.Equal(t, 7, info.Level)
	assert.NotEmpty(t, info.Path, "CharacterInfo.Path should not be empty")
}
