package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
)

func main() {
	fmt.Println("D&D Character Migration Tool")
	fmt.Println("=============================")
	fmt.Println()

	// Initialize storage and loader
	store, err := storage.NewCharacterStorage("")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	loader := data.NewLoader("./data")
	if err := loader.LoadAll(); err != nil {
		log.Fatalf("Failed to load game data: %v", err)
	}

	// Get list of all character files
	baseDir := store.GetBaseDir()
	files, err := filepath.Glob(filepath.Join(baseDir, "*.json"))
	if err != nil {
		log.Fatalf("Failed to list character files: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("No characters found to migrate.")
		return
	}

	fmt.Printf("Found %d character(s) to migrate:\n\n", len(files))

	// Migrate each character
	for _, file := range files {
		fmt.Printf("Migrating: %s\n", filepath.Base(file))

		// Load character
		char, err := loadCharacterFromPath(file)
		if err != nil {
			fmt.Printf("  ERROR: Failed to load: %v\n", err)
			continue
		}

		// Perform migrations
		migrated := false

		// 1. Upgrade ritual caster flags from class
		if char.Spellcasting != nil {
			classData, err := loader.FindClassByName(char.Info.Class)
			if err == nil {
				if char.Spellcasting.RitualCaster != classData.RitualCaster ||
					char.Spellcasting.RitualCasterUnprepared != classData.RitualCasterUnprepared {

					char.Spellcasting.RitualCaster = classData.RitualCaster
					char.Spellcasting.RitualCasterUnprepared = classData.RitualCasterUnprepared
					migrated = true
					fmt.Printf("  - Set RitualCaster=%v, RitualCasterUnprepared=%v\n",
						classData.RitualCaster, classData.RitualCasterUnprepared)
				}
			}

			// 2. Upgrade spell ritual flags from spell database
			spellDB, err := loader.GetSpells()
			if err == nil {
				updatedCount := 0
				for i := range char.Spellcasting.KnownSpells {
					spell := &char.Spellcasting.KnownSpells[i]

					// Look up spell in database
					for _, dbSpell := range spellDB.Spells {
						if dbSpell.Name == spell.Name {
							if spell.Ritual != dbSpell.Ritual {
								spell.Ritual = dbSpell.Ritual
								updatedCount++
								migrated = true
							}
							break
						}
					}
				}
				if updatedCount > 0 {
					fmt.Printf("  - Updated ritual flags for %d spell(s)\n", updatedCount)
				}
			}
		}

		// Save if migrated
		if migrated {
			if err := saveCharacter(file, char); err != nil {
				fmt.Printf("  ERROR: Failed to save: %v\n", err)
			} else {
				fmt.Printf("  ✓ Migrated successfully\n")
			}
		} else {
			fmt.Printf("  - Already up to date\n")
		}
		fmt.Println()
	}

	fmt.Println("Migration complete!")
}

func loadCharacterFromPath(path string) (*models.Character, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return models.ReadFrom(file)
}

func saveCharacter(path string, char *models.Character) error {
	// Write to temp file first
	tempPath := path + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(char); err != nil {
		file.Close()
		os.Remove(tempPath)
		return err
	}

	if err := file.Close(); err != nil {
		os.Remove(tempPath)
		return err
	}

	// Replace original file
	return os.Rename(tempPath, path)
}
