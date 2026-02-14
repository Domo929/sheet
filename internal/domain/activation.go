package domain

// ActivationType represents the activation method for a feature or ability.
type ActivationType string

const (
	ActivationAction   ActivationType = "action"
	ActivationBonus    ActivationType = "bonus"
	ActivationReaction ActivationType = "reaction"
	ActivationPassive  ActivationType = ""
)
