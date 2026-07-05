package models

// ActionEconomy tracks which of a creature's per-turn actions have been used and
// how much movement has been spent during the current turn. Under the 2024 rules
// a creature gets one Action, one Bonus Action, and one Reaction per round, plus
// movement up to its Speed.
type ActionEconomy struct {
	ActionUsed      bool `json:"actionUsed,omitempty"`
	BonusActionUsed bool `json:"bonusActionUsed,omitempty"`
	ReactionUsed    bool `json:"reactionUsed,omitempty"`
	MovementUsed    int  `json:"movementUsed,omitempty"`
}

// ResetTurn clears all per-turn action economy at the start of a new turn. A
// Reaction refreshes at the start of the creature's turn, so it is cleared too.
func (a *ActionEconomy) ResetTurn() {
	a.ActionUsed = false
	a.BonusActionUsed = false
	a.ReactionUsed = false
	a.MovementUsed = 0
}

// RemainingMovement returns the feet of movement left this turn for a given
// speed, clamped to zero.
func (a *ActionEconomy) RemainingMovement(speed int) int {
	rem := speed - a.MovementUsed
	if rem < 0 {
		return 0
	}
	return rem
}

// UseMovement spends feet of movement this turn, clamping the running total to
// the [0, speed] band.
func (a *ActionEconomy) UseMovement(feet, speed int) {
	a.MovementUsed += feet
	if a.MovementUsed < 0 {
		a.MovementUsed = 0
	}
	if a.MovementUsed > speed {
		a.MovementUsed = speed
	}
}
