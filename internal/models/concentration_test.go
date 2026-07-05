package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcentrationStartAndBreak(t *testing.T) {
	char := NewCharacter("id", "Wiz", "Human", "Wizard")
	assert.False(t, char.IsConcentrating())

	dropped := char.StartConcentration("Bless")
	assert.Equal(t, "", dropped, "nothing dropped when not previously concentrating")
	assert.True(t, char.IsConcentrating())
	assert.Equal(t, "Bless", char.Concentration)

	// Starting a new concentration replaces the old one.
	dropped = char.StartConcentration("Haste")
	assert.Equal(t, "Bless", dropped)
	assert.Equal(t, "Haste", char.Concentration)

	// Re-starting the same spell drops nothing.
	dropped = char.StartConcentration("Haste")
	assert.Equal(t, "", dropped)

	dropped = char.BreakConcentration()
	assert.Equal(t, "Haste", dropped)
	assert.False(t, char.IsConcentrating())

	// Breaking again is a no-op.
	assert.Equal(t, "", char.BreakConcentration())
}

func TestConcentrationSaveDC(t *testing.T) {
	cases := map[int]int{
		0:  10, // half of 0 -> min 10
		9:  10, // half rounds down to 4 -> min 10
		20: 10, // exactly 10
		21: 10, // half rounds down to 10
		22: 11, // half is 11
		30: 15,
		50: 25,
	}
	for damage, want := range cases {
		assert.Equal(t, want, ConcentrationSaveDC(damage), "damage %d", damage)
	}
}
