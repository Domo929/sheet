package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActionEconomyResetTurn(t *testing.T) {
	a := ActionEconomy{ActionUsed: true, BonusActionUsed: true, ReactionUsed: true, MovementUsed: 25}
	a.ResetTurn()
	assert.False(t, a.ActionUsed)
	assert.False(t, a.BonusActionUsed)
	assert.False(t, a.ReactionUsed)
	assert.Equal(t, 0, a.MovementUsed)
}

func TestActionEconomyMovement(t *testing.T) {
	a := ActionEconomy{}
	a.UseMovement(15, 30)
	assert.Equal(t, 15, a.MovementUsed)
	assert.Equal(t, 15, a.RemainingMovement(30))

	// Clamps at speed.
	a.UseMovement(50, 30)
	assert.Equal(t, 30, a.MovementUsed)
	assert.Equal(t, 0, a.RemainingMovement(30))

	// Refund clamps at zero.
	a.UseMovement(-100, 30)
	assert.Equal(t, 0, a.MovementUsed)
	assert.Equal(t, 30, a.RemainingMovement(30))
}

func TestRestResetsTurnState(t *testing.T) {
	c := NewCharacter("ae-1", "Test", "Human", "Fighter")

	c.TurnState.ActionUsed = true
	c.TurnState.MovementUsed = 20
	c.ShortRest()
	assert.False(t, c.TurnState.ActionUsed, "short rest resets action economy")
	assert.Equal(t, 0, c.TurnState.MovementUsed)

	c.TurnState.ReactionUsed = true
	c.TurnState.MovementUsed = 10
	c.LongRest()
	assert.False(t, c.TurnState.ReactionUsed, "long rest resets action economy")
	assert.Equal(t, 0, c.TurnState.MovementUsed)
}
