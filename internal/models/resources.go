package models

import "github.com/Domo929/sheet/internal/domain"

// ResourcePool is a limited-use class resource (e.g. Rage, Ki/Focus Points,
// Channel Divinity) tracked on the character sheet. Current is the remaining
// amount; Max is refilled on the appropriate rest.
type ResourcePool struct {
	Name     string `json:"name"`
	Current  int    `json:"current"`
	Max      int    `json:"max"`
	Recharge string `json:"recharge"` // "short" or "long"
}

// SyncResources rebuilds the character's class resource pools from the 2024
// rules for the current class and level, preserving the remaining Current
// amount of any pool that still exists (clamped to the new Max). It is safe to
// call repeatedly (on load, after level-up, etc.).
func (c *Character) SyncResources() {
	defs := domain.ClassResources(c.Info.Class, c.Info.Level, c.AbilityScores.Charisma.Modifier())

	existing := make(map[string]ResourcePool, len(c.Resources))
	for _, p := range c.Resources {
		existing[p.Name] = p
	}

	var out []ResourcePool
	for _, d := range defs {
		current := d.Max
		if prev, ok := existing[d.Name]; ok {
			current = prev.Current
			if current > d.Max {
				current = d.Max
			}
			if current < 0 {
				current = 0
			}
		}
		out = append(out, ResourcePool{
			Name:     d.Name,
			Current:  current,
			Max:      d.Max,
			Recharge: string(d.Recharge),
		})
	}
	c.Resources = out
}

// RestoreResources refills class resource pools for a rest. A long rest refills
// every pool; a short rest refills only pools that recharge on a short rest.
func (c *Character) RestoreResources(longRest bool) {
	for i := range c.Resources {
		if longRest || c.Resources[i].Recharge == string(domain.RechargeShortRest) {
			c.Resources[i].Current = c.Resources[i].Max
		}
	}
}

// SpendResource decreases the named pool by amount (not below 0). Returns false
// if the pool does not exist or lacks enough remaining.
func (c *Character) SpendResource(name string, amount int) bool {
	for i := range c.Resources {
		if c.Resources[i].Name == name {
			if c.Resources[i].Current < amount {
				return false
			}
			c.Resources[i].Current -= amount
			return true
		}
	}
	return false
}

// RestoreResource increases the named pool by amount (not above Max). Returns
// false if the pool does not exist.
func (c *Character) RestoreResource(name string, amount int) bool {
	for i := range c.Resources {
		if c.Resources[i].Name == name {
			c.Resources[i].Current += amount
			if c.Resources[i].Current > c.Resources[i].Max {
				c.Resources[i].Current = c.Resources[i].Max
			}
			return true
		}
	}
	return false
}
