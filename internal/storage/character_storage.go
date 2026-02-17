package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Domo929/sheet/internal/models"
)

// Error sentinels for storage operations
var (
	ErrCharacterNotFound    = errors.New("character not found")
	ErrInvalidCharacterName = errors.New("invalid character name")
	ErrCharacterExists      = errors.New("character already exists")
)

// CharacterStorage handles saving and loading characters to/from disk.
type CharacterStorage struct {
	baseDir string
}

// NewCharacterStorage creates a new character storage manager.
// If baseDir is empty, uses the default ~/.dnd_sheet/characters/ directory.
func NewCharacterStorage(baseDir string) (*CharacterStorage, error) {
	if baseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		baseDir = filepath.Join(homeDir, ".dnd_sheet", "characters")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create character directory: %w", err)
	}

	return &CharacterStorage{baseDir: baseDir}, nil
}

// GetBaseDir returns the base directory for character storage.
func (cs *CharacterStorage) GetBaseDir() string {
	return cs.baseDir
}

// sanitizeFilename removes or replaces invalid filename characters.
func sanitizeFilename(name string) string {
	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")

	// Remove or replace invalid characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "")
	}

	// Ensure it's not empty
	if name == "" {
		name = "unnamed"
	}

	return name
}

// getCharacterPath returns the full path for a character file.
func (cs *CharacterStorage) getCharacterPath(characterName string) string {
	filename := sanitizeFilename(characterName) + ".json"
	return filepath.Join(cs.baseDir, filename)
}

// Save saves a character to disk using atomic write (temp file + rename).
// Returns the path where the character was saved.
func (cs *CharacterStorage) Save(character *models.Character) (string, error) {
	if character == nil {
		return "", errors.New("character cannot be nil")
	}

	if character.Info.Name == "" {
		return "", ErrInvalidCharacterName
	}

	path := cs.getCharacterPath(character.Info.Name)

	// Write to a unique temp file in the same directory to avoid collisions
	file, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp.*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp character file: %w", err)
	}
	tempPath := file.Name()

	if err := character.WriteTo(file); err != nil {
		file.Close()
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to write character: %w", err)
	}

	// Sync to ensure data is flushed to disk before rename
	if err := file.Sync(); err != nil {
		file.Close()
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to sync temp character file: %w", err)
	}

	if err := file.Close(); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to close temp character file: %w", err)
	}

	// Atomically replace the original file
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to finalize character save: %w", err)
	}

	return path, nil
}

// Load loads a character from disk by name.
func (cs *CharacterStorage) Load(characterName string) (*models.Character, error) {
	if characterName == "" {
		return nil, ErrInvalidCharacterName
	}

	path := cs.getCharacterPath(characterName)

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, ErrCharacterNotFound
	}

	// Open file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open character file: %w", err)
	}
	defer file.Close()

	// Read character from file using decoder
	character, err := models.ReadFrom(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read character: %w", err)
	}

	// Migrate deprecated notes field
	character.Personality.MigrateNotes()

	return character, nil
}

// LoadByPath loads a character from a specific file path.
func (cs *CharacterStorage) LoadByPath(path string) (*models.Character, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, ErrCharacterNotFound
	}

	// Open file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open character file: %w", err)
	}
	defer file.Close()

	// Read character from file using decoder
	character, err := models.ReadFrom(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read character: %w", err)
	}

	// Migrate deprecated notes field
	character.Personality.MigrateNotes()

	return character, nil
}

// Delete deletes a character file by name.
func (cs *CharacterStorage) Delete(characterName string) error {
	if characterName == "" {
		return ErrInvalidCharacterName
	}

	path := cs.getCharacterPath(characterName)

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ErrCharacterNotFound
	}

	// Delete file
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete character file: %w", err)
	}

	return nil
}

// Exists checks if a character file exists.
func (cs *CharacterStorage) Exists(characterName string) bool {
	if characterName == "" {
		return false
	}

	path := cs.getCharacterPath(characterName)
	_, err := os.Stat(path)
	return err == nil
}

// CharacterInfo contains basic information about a saved character.
type CharacterInfo struct {
	Name  string
	Race  string
	Class string
	Level int
	Path  string
}

// List returns a list of all saved characters with their basic info.
func (cs *CharacterStorage) List() ([]CharacterInfo, error) {
	// Read directory
	entries, err := os.ReadDir(cs.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read character directory: %w", err)
	}

	var characters []CharacterInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .json files
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(cs.baseDir, entry.Name())

		// Read and parse the file to get basic info
		data, err := os.ReadFile(path)
		if err != nil {
			// Skip files we can't read
			continue
		}

		// Parse just the info we need
		var partialCharacter struct {
			Info models.CharacterInfo `json:"info"`
		}
		if err := json.Unmarshal(data, &partialCharacter); err != nil {
			// Skip files we can't parse
			continue
		}

		characters = append(characters, CharacterInfo{
			Name:  partialCharacter.Info.Name,
			Race:  partialCharacter.Info.Race,
			Class: partialCharacter.Info.Class,
			Level: partialCharacter.Info.Level,
			Path:  path,
		})
	}

	return characters, nil
}

// Rename renames a character file.
func (cs *CharacterStorage) Rename(oldName, newName string) error {
	if oldName == "" || newName == "" {
		return ErrInvalidCharacterName
	}

	oldPath := cs.getCharacterPath(oldName)
	newPath := cs.getCharacterPath(newName)

	// Check if old file exists
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return ErrCharacterNotFound
	}

	// Check if new file already exists
	if _, err := os.Stat(newPath); err == nil {
		return ErrCharacterExists
	}

	// Rename file
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename character file: %w", err)
	}

	return nil
}

// AutoSave is a convenience method that saves a character and marks it as updated.
// This is intended to be called after making changes to a character.
func (cs *CharacterStorage) AutoSave(character *models.Character) error {
	character.MarkUpdated()
	_, err := cs.Save(character)
	return err
}
