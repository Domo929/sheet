package models

// IsConcentrating reports whether the character is currently concentrating on a
// spell.
func (c *Character) IsConcentrating() bool {
	return c.Concentration != ""
}

// StartConcentration begins concentration on the named spell. Because only one
// spell can be concentrated on at a time (2024 rules), any existing
// concentration is replaced; the name of the dropped spell is returned (or ""
// if nothing was replaced).
func (c *Character) StartConcentration(spellName string) string {
	dropped := ""
	if c.Concentration != "" && c.Concentration != spellName {
		dropped = c.Concentration
	}
	c.Concentration = spellName
	c.MarkUpdated()
	return dropped
}

// BreakConcentration ends any active concentration and returns the name of the
// spell that was dropped (or "" if the character was not concentrating).
func (c *Character) BreakConcentration() string {
	dropped := c.Concentration
	c.Concentration = ""
	if dropped != "" {
		c.MarkUpdated()
	}
	return dropped
}

// ConcentrationSaveDC returns the Constitution saving throw DC required to
// maintain concentration after taking the given amount of damage: 10 or half
// the damage taken, whichever is higher (2024 rules).
func ConcentrationSaveDC(damage int) int {
	dc := damage / 2
	if dc < 10 {
		dc = 10
	}
	return dc
}
