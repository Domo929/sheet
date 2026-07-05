package models

import (
	"sort"
	"strings"

	"github.com/Domo929/sheet/internal/domain"
)

// DamageDefenses holds a character's damage resistances, immunities, and
// vulnerabilities as display-ready, title-cased damage type names.
type DamageDefenses struct {
	Resistances     []string
	Immunities      []string
	Vulnerabilities []string
}

// Empty reports whether there are no defenses of any kind.
func (d DamageDefenses) Empty() bool {
	return len(d.Resistances) == 0 && len(d.Immunities) == 0 && len(d.Vulnerabilities) == 0
}

// Defenses derives the character's damage resistances, immunities, and
// vulnerabilities from species traits and feats (permanent, passive effects).
//
// Class features are intentionally excluded: many are conditional (for example
// a Barbarian's Rage grants Resistance to Bludgeoning, Piercing, and Slashing
// only while raging), so treating them as always-on would be misleading.
// A Dragonborn's draconic ancestry Resistance is resolved from the chosen
// ancestry rather than its generic trait text.
func (c *Character) Defenses() DamageDefenses {
	descriptions := make([]string, 0, len(c.Features.RacialTraits)+len(c.Features.Feats))
	for _, t := range c.Features.RacialTraits {
		descriptions = append(descriptions, t.Description)
	}
	for _, f := range c.Features.Feats {
		descriptions = append(descriptions, f.Description)
	}

	res, imm, vuln := domain.ParseDamageDefenses(descriptions)

	if strings.EqualFold(strings.TrimSpace(c.Info.Race), "Dragonborn") {
		if dt, ok := domain.DraconicAncestryResistance(c.Info.Subrace); ok {
			res = appendUniqueType(res, dt)
		}
	}

	return DamageDefenses{
		Resistances:     titleDamageTypes(res),
		Immunities:      titleDamageTypes(imm),
		Vulnerabilities: titleDamageTypes(vuln),
	}
}

func appendUniqueType(list []domain.DamageType, dt domain.DamageType) []domain.DamageType {
	for _, existing := range list {
		if existing == dt {
			return list
		}
	}
	return append(list, dt)
}

func titleDamageTypes(types []domain.DamageType) []string {
	if len(types) == 0 {
		return nil
	}
	out := make([]string, 0, len(types))
	for _, t := range types {
		out = append(out, t.Title())
	}
	sort.Strings(out)
	return out
}
