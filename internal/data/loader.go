// Package data provides functionality for loading and caching game data from JSON files.
//
// The package includes structures for representing D&D 5e game data (races, classes,
// spells, backgrounds, and conditions) and a Loader type that handles loading and
// caching this data from JSON files.
//
// Basic usage:
//
//	loader := data.NewLoader("./data")
//
//	// Load all data at once
//	if err := loader.LoadAll(); err != nil {
//		log.Fatal(err)
//	}
//
//	// Or load data on demand
//	races, err := loader.GetRaces()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Find specific items
//	human, err := loader.FindRaceByName("Human")
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Thread Safety:
//
// The Loader type is safe for concurrent use. It uses a read-write mutex to
// ensure thread-safe access to cached data.
//
// Caching:
//
// Data is cached in memory after the first load. Subsequent calls to Get* methods
// will return the cached data without re-reading files. Use ClearCache() to force
// a reload of data from disk.
package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Loader handles loading and caching of game data from JSON files.
type Loader struct {
	dataDir string

	// Cached data
	races       *RaceData
	classes     *ClassData
	spells      *SpellDatabase
	backgrounds *BackgroundData
	conditions  *ConditionData

	// Mutex for thread-safe access
	mu sync.RWMutex
}

// NewLoader creates a new data loader with the specified data directory.
// If dataDir is empty, it defaults to "./data".
func NewLoader(dataDir string) *Loader {
	if dataDir == "" {
		dataDir = "./data"
	}
	return &Loader{
		dataDir: dataDir,
	}
}

// LoadAll loads all game data files at once.
// Returns an error if any file fails to load.
func (l *Loader) LoadAll() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Load races
	if err := l.loadRacesUnsafe(); err != nil {
		return fmt.Errorf("failed to load races: %w", err)
	}

	// Load classes
	if err := l.loadClassesUnsafe(); err != nil {
		return fmt.Errorf("failed to load classes: %w", err)
	}

	// Load spells
	if err := l.loadSpellsUnsafe(); err != nil {
		return fmt.Errorf("failed to load spells: %w", err)
	}

	// Load backgrounds
	if err := l.loadBackgroundsUnsafe(); err != nil {
		return fmt.Errorf("failed to load backgrounds: %w", err)
	}

	// Load conditions
	if err := l.loadConditionsUnsafe(); err != nil {
		return fmt.Errorf("failed to load conditions: %w", err)
	}

	return nil
}

// GetRaces returns all race data, loading it if necessary.
func (l *Loader) GetRaces() (*RaceData, error) {
	l.mu.RLock()
	if l.races != nil {
		defer l.mu.RUnlock()
		return l.races, nil
	}
	l.mu.RUnlock()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Check again in case another goroutine loaded it
	if l.races != nil {
		return l.races, nil
	}

	if err := l.loadRacesUnsafe(); err != nil {
		return nil, err
	}

	return l.races, nil
}

// GetClasses returns all class data, loading it if necessary.
func (l *Loader) GetClasses() (*ClassData, error) {
	l.mu.RLock()
	if l.classes != nil {
		defer l.mu.RUnlock()
		return l.classes, nil
	}
	l.mu.RUnlock()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Check again in case another goroutine loaded it
	if l.classes != nil {
		return l.classes, nil
	}

	if err := l.loadClassesUnsafe(); err != nil {
		return nil, err
	}

	return l.classes, nil
}

// GetSpells returns all spell data, loading it if necessary.
func (l *Loader) GetSpells() (*SpellDatabase, error) {
	l.mu.RLock()
	if l.spells != nil {
		defer l.mu.RUnlock()
		return l.spells, nil
	}
	l.mu.RUnlock()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Check again in case another goroutine loaded it
	if l.spells != nil {
		return l.spells, nil
	}

	if err := l.loadSpellsUnsafe(); err != nil {
		return nil, err
	}

	return l.spells, nil
}

// GetBackgrounds returns all background data, loading it if necessary.
func (l *Loader) GetBackgrounds() (*BackgroundData, error) {
	l.mu.RLock()
	if l.backgrounds != nil {
		defer l.mu.RUnlock()
		return l.backgrounds, nil
	}
	l.mu.RUnlock()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Check again in case another goroutine loaded it
	if l.backgrounds != nil {
		return l.backgrounds, nil
	}

	if err := l.loadBackgroundsUnsafe(); err != nil {
		return nil, err
	}

	return l.backgrounds, nil
}

// GetConditions returns all condition data, loading it if necessary.
func (l *Loader) GetConditions() (*ConditionData, error) {
	l.mu.RLock()
	if l.conditions != nil {
		defer l.mu.RUnlock()
		return l.conditions, nil
	}
	l.mu.RUnlock()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Check again in case another goroutine loaded it
	if l.conditions != nil {
		return l.conditions, nil
	}

	if err := l.loadConditionsUnsafe(); err != nil {
		return nil, err
	}

	return l.conditions, nil
}

// FindRaceByName finds a race by name (case-sensitive).
func (l *Loader) FindRaceByName(name string) (*Race, error) {
	races, err := l.GetRaces()
	if err != nil {
		return nil, err
	}

	for _, race := range races.Races {
		if race.Name == name {
			return &race, nil
		}
	}

	return nil, fmt.Errorf("race not found: %s", name)
}

// FindClassByName finds a class by name (case-sensitive).
func (l *Loader) FindClassByName(name string) (*Class, error) {
	classes, err := l.GetClasses()
	if err != nil {
		return nil, err
	}

	for _, class := range classes.Classes {
		if class.Name == name {
			return &class, nil
		}
	}

	return nil, fmt.Errorf("class not found: %s", name)
}

// FindSpellByName finds a spell by name (case-sensitive).
func (l *Loader) FindSpellByName(name string) (*SpellData, error) {
	spells, err := l.GetSpells()
	if err != nil {
		return nil, err
	}

	for _, spell := range spells.Spells {
		if spell.Name == name {
			return &spell, nil
		}
	}

	return nil, fmt.Errorf("spell not found: %s", name)
}

// FindBackgroundByName finds a background by name (case-sensitive).
func (l *Loader) FindBackgroundByName(name string) (*Background, error) {
	backgrounds, err := l.GetBackgrounds()
	if err != nil {
		return nil, err
	}

	for _, background := range backgrounds.Backgrounds {
		if background.Name == name {
			return &background, nil
		}
	}

	return nil, fmt.Errorf("background not found: %s", name)
}

// FindConditionByName finds a condition by name (case-sensitive).
func (l *Loader) FindConditionByName(name string) (*Condition, error) {
	conditions, err := l.GetConditions()
	if err != nil {
		return nil, err
	}

	for _, condition := range conditions.Conditions {
		if condition.Name == name {
			return &condition, nil
		}
	}

	return nil, fmt.Errorf("condition not found: %s", name)
}

// ClearCache clears all cached data, forcing a reload on next access.
func (l *Loader) ClearCache() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.races = nil
	l.classes = nil
	l.spells = nil
	l.backgrounds = nil
	l.conditions = nil
}

// Internal unsafe methods (must be called with lock held)

func (l *Loader) loadRacesUnsafe() error {
	path := filepath.Join(l.dataDir, "races.json")
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open races.json: %w", err)
	}
	defer f.Close()

	var races RaceData
	if err := json.NewDecoder(f).Decode(&races); err != nil {
		return fmt.Errorf("failed to parse races.json: %w", err)
	}

	if len(races.Races) == 0 {
		return fmt.Errorf("races.json contains no races")
	}

	l.races = &races
	return nil
}

func (l *Loader) loadClassesUnsafe() error {
	path := filepath.Join(l.dataDir, "classes.json")
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open classes.json: %w", err)
	}
	defer f.Close()

	var classes ClassData
	if err := json.NewDecoder(f).Decode(&classes); err != nil {
		return fmt.Errorf("failed to parse classes.json: %w", err)
	}

	if len(classes.Classes) == 0 {
		return fmt.Errorf("classes.json contains no classes")
	}

	l.classes = &classes
	return nil
}

func (l *Loader) loadSpellsUnsafe() error {
	path := filepath.Join(l.dataDir, "spells.json")
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open spells.json: %w", err)
	}
	defer f.Close()

	var spells SpellDatabase
	if err := json.NewDecoder(f).Decode(&spells); err != nil {
		return fmt.Errorf("failed to parse spells.json: %w", err)
	}

	if len(spells.Spells) == 0 {
		return fmt.Errorf("spells.json contains no spells")
	}

	l.spells = &spells
	return nil
}

func (l *Loader) loadBackgroundsUnsafe() error {
	path := filepath.Join(l.dataDir, "backgrounds.json")
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open backgrounds.json: %w", err)
	}
	defer f.Close()

	var backgrounds BackgroundData
	if err := json.NewDecoder(f).Decode(&backgrounds); err != nil {
		return fmt.Errorf("failed to parse backgrounds.json: %w", err)
	}

	if len(backgrounds.Backgrounds) == 0 {
		return fmt.Errorf("backgrounds.json contains no backgrounds")
	}

	l.backgrounds = &backgrounds
	return nil
}

func (l *Loader) loadConditionsUnsafe() error {
	path := filepath.Join(l.dataDir, "conditions.json")
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open conditions.json: %w", err)
	}
	defer f.Close()

	var conditions ConditionData
	if err := json.NewDecoder(f).Decode(&conditions); err != nil {
		return fmt.Errorf("failed to parse conditions.json: %w", err)
	}

	if len(conditions.Conditions) == 0 {
		return fmt.Errorf("conditions.json contains no conditions")
	}

	l.conditions = &conditions
	return nil
}
