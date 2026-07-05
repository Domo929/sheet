package models

// EncumbranceLevel classifies how heavily laden a character is under the 2024
// optional variant encumbrance rules (Player's Handbook "Encumbrance" sidebar).
type EncumbranceLevel int

const (
	// LoadNormal means carried weight is at or below Strength × 5: no penalty.
	LoadNormal EncumbranceLevel = iota
	// LoadEncumbered means carried weight exceeds Strength × 5: speed −10 ft.
	LoadEncumbered
	// LoadHeavilyEncumbered means carried weight exceeds Strength × 10: speed
	// −20 ft and disadvantage on Strength, Dexterity, and Constitution checks,
	// attack rolls, and saving throws.
	LoadHeavilyEncumbered
	// LoadOverCapacity means carried weight exceeds the maximum carrying
	// capacity of Strength × 15.
	LoadOverCapacity
)

// String returns a short human-readable label for the encumbrance level.
func (e EncumbranceLevel) String() string {
	switch e {
	case LoadEncumbered:
		return "Encumbered"
	case LoadHeavilyEncumbered:
		return "Heavily Encumbered"
	case LoadOverCapacity:
		return "Over Capacity"
	default:
		return "Unencumbered"
	}
}

// SpeedPenalty returns the walking-speed reduction in feet imposed by this
// encumbrance level under the variant rule.
func (e EncumbranceLevel) SpeedPenalty() int {
	switch e {
	case LoadEncumbered:
		return 10
	case LoadHeavilyEncumbered, LoadOverCapacity:
		return 20
	default:
		return 0
	}
}

// CarryingCapacity returns the maximum weight in pounds the character can carry:
// Strength score × 15.
func (c *Character) CarryingCapacity() float64 {
	return float64(c.AbilityScores.Strength.Total()) * 15.0
}

// PushDragLiftCapacity returns the maximum weight in pounds the character can
// push, drag, or lift: Strength score × 30.
func (c *Character) PushDragLiftCapacity() float64 {
	return float64(c.AbilityScores.Strength.Total()) * 30.0
}

// CarriedWeight returns the total weight of carried items in pounds. Coin weight
// is not counted, matching the D&D Beyond default.
func (c *Character) CarriedWeight() float64 {
	return c.Inventory.TotalWeight()
}

// Encumbrance classifies the character's current load under the 2024 variant
// encumbrance rules.
func (c *Character) Encumbrance() EncumbranceLevel {
	str := float64(c.AbilityScores.Strength.Total())
	w := c.CarriedWeight()
	switch {
	case w > str*15:
		return LoadOverCapacity
	case w > str*10:
		return LoadHeavilyEncumbered
	case w > str*5:
		return LoadEncumbered
	default:
		return LoadNormal
	}
}
