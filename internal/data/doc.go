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
