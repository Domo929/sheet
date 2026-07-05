package domain

import (
	"sort"
	"strings"
)

// knownDamageTypes is the set of canonical damage types recognised by the
// defenses parser.
var knownDamageTypes = map[string]DamageType{
	"acid":        DamageAcid,
	"bludgeoning": DamageBludgeoning,
	"cold":        DamageCold,
	"fire":        DamageFire,
	"force":       DamageForce,
	"lightning":   DamageLightning,
	"necrotic":    DamageNecrotic,
	"piercing":    DamagePiercing,
	"poison":      DamagePoison,
	"psychic":     DamagePsychic,
	"radiant":     DamageRadiant,
	"slashing":    DamageSlashing,
	"thunder":     DamageThunder,
}

// Title returns the damage type formatted for display, e.g. "Fire".
func (d DamageType) Title() string {
	s := string(d)
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// defenseWindowLimit bounds how far after a "resistance to"/"immunity to"
// keyword the parser will look for the closing " damage", so that an unrelated
// later sentence cannot be misread as part of the defense clause.
const defenseWindowLimit = 45

// ParseDamageDefenses extracts damage resistances, immunities, and
// vulnerabilities described in free-text trait/feat descriptions such as
// "You have Resistance to Fire damage" or "Resistance to Lightning and Thunder
// damage". Only canonical damage types are recognised, so generic phrasings
// like "Resistance to the damage type determined by your ancestry" are ignored
// (callers resolve those separately). Each returned slice is unique and sorted.
func ParseDamageDefenses(descriptions []string) (resistances, immunities, vulnerabilities []DamageType) {
	resSet := map[DamageType]bool{}
	immSet := map[DamageType]bool{}
	vulnSet := map[DamageType]bool{}

	for _, desc := range descriptions {
		lower := strings.ToLower(desc)
		scanDefenseClause(lower, "resistance to", resSet)
		scanDefenseClause(lower, "immunity to", immSet)
		scanDefenseClause(lower, "immune to", immSet)
		scanDefenseClause(lower, "vulnerability to", vulnSet)
		scanDefenseClause(lower, "vulnerable to", vulnSet)
	}

	return sortedTypes(resSet), sortedTypes(immSet), sortedTypes(vulnSet)
}

// scanDefenseClause finds every occurrence of keyword in text and records any
// canonical damage types appearing between the keyword and the following
// " damage" token into set.
func scanDefenseClause(text, keyword string, set map[DamageType]bool) {
	search := text
	offset := 0
	for {
		idx := strings.Index(search, keyword)
		if idx < 0 {
			return
		}
		rest := search[idx+len(keyword):]
		if end := strings.Index(rest, " damage"); end >= 0 && end <= defenseWindowLimit {
			for _, token := range strings.FieldsFunc(rest[:end], func(r rune) bool {
				return r < 'a' || r > 'z'
			}) {
				if dt, ok := knownDamageTypes[token]; ok {
					set[dt] = true
				}
			}
		}
		advance := idx + len(keyword)
		offset += advance
		search = search[advance:]
	}
}

func sortedTypes(set map[DamageType]bool) []DamageType {
	if len(set) == 0 {
		return nil
	}
	out := make([]DamageType, 0, len(set))
	for dt := range set {
		out = append(out, dt)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// draconicAncestryResistance maps a Dragonborn draconic ancestry (subrace) to
// the damage type it grants Resistance to under the 2024 rules. It mirrors the
// subtype damage types in data/races.json.
var draconicAncestryResistance = map[string]DamageType{
	"black":  DamageAcid,
	"blue":   DamageLightning,
	"brass":  DamageFire,
	"bronze": DamageLightning,
	"copper": DamageAcid,
	"gold":   DamageFire,
	"green":  DamagePoison,
	"red":    DamageFire,
	"silver": DamageCold,
	"white":  DamageCold,
}

// DraconicAncestryResistance returns the Resistance damage type granted by a
// Dragonborn's draconic ancestry, resolved from the chosen ancestry (subrace).
func DraconicAncestryResistance(ancestry string) (DamageType, bool) {
	dt, ok := draconicAncestryResistance[strings.ToLower(strings.TrimSpace(ancestry))]
	return dt, ok
}
