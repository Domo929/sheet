package models

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	darkvisionRe  = regexp.MustCompile(`(?i)darkvision[^.]*?(\d+)\s*feet`)
	blindsightRe  = regexp.MustCompile(`(?i)blindsight[^.]*?(\d+)\s*feet`)
	tremorsenseRe = regexp.MustCompile(`(?i)tremorsense[^.]*?(\d+)\s*feet`)
	truesightRe   = regexp.MustCompile(`(?i)truesight[^.]*?(\d+)\s*feet`)
	speedRaiseRe  = regexp.MustCompile(`(?i)your speed increases to (\d+)\s*feet`)
)

// isTemporarySense reports whether a trait description grants a sense only
// temporarily / conditionally (e.g. Dwarf Stonecunning grants Tremorsense as a
// Bonus Action for 10 minutes). Such grants should not be recorded as permanent
// senses.
func isTemporarySense(desc string) bool {
	d := strings.ToLower(desc)
	for _, marker := range []string{"bonus action", "as an action", "for 1 minute", "for 10 minute", "until the", "when you"} {
		if strings.Contains(d, marker) {
			return true
		}
	}
	return false
}

func maxSenseRange(re *regexp.Regexp, desc string) int {
	best := 0
	for _, m := range re.FindAllStringSubmatch(desc, -1) {
		if n, err := strconv.Atoi(m[1]); err == nil && n > best {
			best = n
		}
	}
	return best
}

// DeriveSensesFromTraits scans trait descriptions for permanently-granted
// special senses and returns the resulting Senses (taking the greatest range
// found for each sense). Conditional or temporary grants are ignored.
func DeriveSensesFromTraits(traits []Feature) Senses {
	var s Senses
	for _, t := range traits {
		desc := t.Description
		if isTemporarySense(desc) {
			continue
		}
		if n := maxSenseRange(darkvisionRe, desc); n > s.Darkvision {
			s.Darkvision = n
		}
		if n := maxSenseRange(blindsightRe, desc); n > s.Blindsight {
			s.Blindsight = n
		}
		if n := maxSenseRange(tremorsenseRe, desc); n > s.Tremorsense {
			s.Tremorsense = n
		}
		if n := maxSenseRange(truesightRe, desc); n > s.Truesight {
			s.Truesight = n
		}
	}
	return s
}

// DeriveWalkSpeedOverride scans trait descriptions for a permanent walking-speed
// override (e.g. Wood Elf's "Your Speed increases to 35 feet") and returns the
// highest value found, or 0 if none.
func DeriveWalkSpeedOverride(traits []Feature) int {
	best := 0
	for _, t := range traits {
		for _, m := range speedRaiseRe.FindAllStringSubmatch(t.Description, -1) {
			if n, err := strconv.Atoi(m[1]); err == nil && n > best {
				best = n
			}
		}
	}
	return best
}

// SyncSenses fills in the character's special senses from its racial traits,
// never lowering a sense range that is already set (so senses gained from feats
// or items are preserved). Existing saved characters gain their species senses
// the next time the sheet is opened.
func (c *Character) SyncSenses() {
	derived := DeriveSensesFromTraits(c.Features.RacialTraits)
	if derived.Darkvision > c.CombatStats.Senses.Darkvision {
		c.CombatStats.Senses.Darkvision = derived.Darkvision
	}
	if derived.Blindsight > c.CombatStats.Senses.Blindsight {
		c.CombatStats.Senses.Blindsight = derived.Blindsight
	}
	if derived.Tremorsense > c.CombatStats.Senses.Tremorsense {
		c.CombatStats.Senses.Tremorsense = derived.Tremorsense
	}
	if derived.Truesight > c.CombatStats.Senses.Truesight {
		c.CombatStats.Senses.Truesight = derived.Truesight
	}
}
