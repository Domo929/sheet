package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncResourcesPopulatesPools(t *testing.T) {
	char := NewCharacter("id", "Grog", "Human", "Barbarian")
	char.Info.Level = 6
	char.SyncResources()

	assert.Len(t, char.Resources, 1)
	assert.Equal(t, "Rage", char.Resources[0].Name)
	assert.Equal(t, 4, char.Resources[0].Max)
	assert.Equal(t, 4, char.Resources[0].Current, "new pools start full")
}

func TestSyncResourcesPreservesCurrentAndClamps(t *testing.T) {
	char := NewCharacter("id", "Grog", "Human", "Barbarian")
	char.Info.Level = 6
	char.SyncResources()

	// Spend two Rages, then a level change re-syncs.
	char.Resources[0].Current = 2

	// Level down to 3 (Rage max 3): current 2 is preserved (<= new max).
	char.Info.Level = 3
	char.SyncResources()
	assert.Equal(t, 3, char.Resources[0].Max)
	assert.Equal(t, 2, char.Resources[0].Current)

	// Now set current above a smaller max and re-sync to clamp.
	char.Resources[0].Current = 3
	char.Info.Level = 1 // Rage max 2
	char.SyncResources()
	assert.Equal(t, 2, char.Resources[0].Max)
	assert.Equal(t, 2, char.Resources[0].Current, "current clamps to new max")
}

func TestRestoreResourcesShortVsLong(t *testing.T) {
	char := NewCharacter("id", "Mona", "Human", "Fighter")
	char.Info.Level = 10
	char.SyncResources() // Second Wind (short), Action Surge (short)

	// Spend everything.
	for i := range char.Resources {
		char.Resources[i].Current = 0
	}

	// Short rest refills short-rest pools.
	char.RestoreResources(false)
	for _, r := range char.Resources {
		assert.Equal(t, r.Max, r.Current, "%s should refill on short rest", r.Name)
	}
}

func TestRestoreResourcesLongOnly(t *testing.T) {
	char := NewCharacter("id", "Zar", "Human", "Sorcerer")
	char.Info.Level = 5
	char.SyncResources() // Sorcery Points (long)
	char.Resources[0].Current = 0

	// Short rest does NOT refill a long-rest pool.
	char.RestoreResources(false)
	assert.Equal(t, 0, char.Resources[0].Current)

	// Long rest does.
	char.RestoreResources(true)
	assert.Equal(t, char.Resources[0].Max, char.Resources[0].Current)
}

func TestSpendAndRestoreResource(t *testing.T) {
	char := NewCharacter("id", "Grog", "Human", "Barbarian")
	char.Info.Level = 6
	char.SyncResources() // Rage 4/4

	assert.True(t, char.SpendResource("Rage", 1))
	assert.Equal(t, 3, char.Resources[0].Current)

	// Cannot overspend.
	assert.False(t, char.SpendResource("Rage", 10))
	assert.Equal(t, 3, char.Resources[0].Current)

	// Restore clamps to max.
	assert.True(t, char.RestoreResource("Rage", 10))
	assert.Equal(t, 4, char.Resources[0].Current)

	// Unknown pool.
	assert.False(t, char.SpendResource("Ki", 1))
}

func TestLongRestRestoresResources(t *testing.T) {
	char := NewCharacter("id", "Grog", "Human", "Barbarian")
	char.Info.Level = 6
	char.SyncResources()
	char.Resources[0].Current = 0

	char.LongRest()
	assert.Equal(t, char.Resources[0].Max, char.Resources[0].Current)
}
