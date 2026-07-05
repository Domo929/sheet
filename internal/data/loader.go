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
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	sheet "github.com/Domo929/sheet"
)

// Loader handles loading and caching of game data from JSON files.
type Loader struct {
	dataDir string

	// fsys, when non-nil, is the filesystem the loader reads data files from
	// (for example, the game data embedded into the binary at build time). When
	// it is set, dataDir is ignored. When it is nil, data is read from disk
	// relative to dataDir.
	fsys fs.FS

	// Cached data
	races       *RaceData
	classes     *ClassData
	spells      *SpellDatabase
	backgrounds *BackgroundData
	conditions  *ConditionData
	equipment   *Equipment
	feats       *FeatData

	// Mutex for thread-safe access
	mu sync.RWMutex
}

// NewLoader creates a new data loader that reads game data from disk, relative
// to the specified data directory. If dataDir is empty, it defaults to "./data".
//
// For production use, prefer NewEmbeddedLoader, which reads the game data bundled
// into the binary and therefore does not depend on the process's working
// directory. NewLoader remains useful for tests and for overriding the bundled
// data with files on disk.
func NewLoader(dataDir string) *Loader {
	if dataDir == "" {
		dataDir = "./data"
	}
	return &Loader{
		dataDir: dataDir,
	}
}

// NewEmbeddedLoader creates a data loader that reads the game data files
// embedded into the binary at build time. The resulting loader is independent
// of the current working directory, so the compiled executable can be run from
// anywhere without a separate data/ folder.
func NewEmbeddedLoader() *Loader {
	return &Loader{
		fsys: sheet.DataFS(),
	}
}

// openDataFile opens the named data file (e.g. "races.json") from the loader's
// embedded filesystem when one is configured, or from disk relative to dataDir
// otherwise. The caller is responsible for closing the returned reader.
func (l *Loader) openDataFile(name string) (io.ReadCloser, error) {
	if l.fsys != nil {
		return l.fsys.Open(name)
	}
	return os.Open(filepath.Join(l.dataDir, name))
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

	// Load equipment
	if err := l.loadEquipmentUnsafe(); err != nil {
		return fmt.Errorf("failed to load equipment: %w", err)
	}

	// Load feats
	if err := l.loadFeatsUnsafe(); err != nil {
		return fmt.Errorf("failed to load feats: %w", err)
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

// GetFeats returns all feat data, loading it if necessary.
func (l *Loader) GetFeats() (*FeatData, error) {
	l.mu.RLock()
	if l.feats != nil {
		defer l.mu.RUnlock()
		return l.feats, nil
	}
	l.mu.RUnlock()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Check again in case another goroutine loaded it
	if l.feats != nil {
		return l.feats, nil
	}

	if err := l.loadFeatsUnsafe(); err != nil {
		return nil, err
	}

	return l.feats, nil
}

// FindFeatByName finds a feat by name (case-sensitive).
func (l *Loader) FindFeatByName(name string) (*Feat, error) {
	feats, err := l.GetFeats()
	if err != nil {
		return nil, err
	}

	for _, feat := range feats.Feats {
		if feat.Name == name {
			return &feat, nil
		}
	}

	return nil, fmt.Errorf("feat not found: %s", name)
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
	l.equipment = nil
	l.feats = nil
}

// Internal unsafe methods (must be called with lock held)

func (l *Loader) loadRacesUnsafe() error {
	f, err := l.openDataFile("races.json")
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
	f, err := l.openDataFile("classes.json")
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
	f, err := l.openDataFile("spells.json")
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
	f, err := l.openDataFile("backgrounds.json")
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
	f, err := l.openDataFile("conditions.json")
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

// GetEquipment returns all equipment data, loading it if necessary.
func (l *Loader) GetEquipment() (*Equipment, error) {
	l.mu.RLock()
	if l.equipment != nil {
		defer l.mu.RUnlock()
		return l.equipment, nil
	}
	l.mu.RUnlock()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Double-check after acquiring write lock
	if l.equipment != nil {
		return l.equipment, nil
	}

	if err := l.loadEquipmentUnsafe(); err != nil {
		return nil, err
	}

	return l.equipment, nil
}

// loadEquipmentUnsafe loads equipment data without acquiring locks.
// Caller must hold the write lock.
func (l *Loader) loadEquipmentUnsafe() error {
	if l.equipment != nil {
		return nil
	}

	f, err := l.openDataFile("equipment.json")
	if err != nil {
		return fmt.Errorf("failed to open equipment.json: %w", err)
	}
	defer f.Close()

	var equipment Equipment
	if err := json.NewDecoder(f).Decode(&equipment); err != nil {
		return fmt.Errorf("failed to parse equipment.json: %w", err)
	}

	l.equipment = &equipment
	return nil
}

// loadFeatsUnsafe loads feat data without acquiring locks.
// Caller must hold the write lock.
func (l *Loader) loadFeatsUnsafe() error {
	if l.feats != nil {
		return nil
	}

	f, err := l.openDataFile("feats.json")
	if err != nil {
		return fmt.Errorf("failed to open feats.json: %w", err)
	}
	defer f.Close()

	var feats FeatData
	if err := json.NewDecoder(f).Decode(&feats); err != nil {
		return fmt.Errorf("failed to parse feats.json: %w", err)
	}

	if len(feats.Feats) == 0 {
		return fmt.Errorf("feats.json contains no feats")
	}

	l.feats = &feats
	return nil
}
