package domain

import "strings"

// ResourceRecharge indicates when a limited-use class resource pool refills.
type ResourceRecharge string

const (
	// RechargeShortRest pools refill on a Short Rest or a Long Rest.
	RechargeShortRest ResourceRecharge = "short"
	// RechargeLongRest pools refill only on a Long Rest.
	RechargeLongRest ResourceRecharge = "long"
)

// ClassResource describes a limited-use class resource pool at a given level.
type ClassResource struct {
	Name     string
	Max      int
	Recharge ResourceRecharge
	// ShortRestRecovery is the number of expended uses regained on a Short
	// Rest for a long-recharge pool that partially recovers (e.g. 2024 Rage
	// and Channel Divinity regain one use per Short Rest). Zero for pools with
	// no partial short-rest recovery.
	ShortRestRecovery int
}

// ClassResources returns the limited-use resource pools granted by a class at
// the given level under the 2024 rules. chaMod supplies the Charisma modifier
// for resources that scale with it (Bardic Inspiration). Pools whose Max is
// zero at the given level are omitted.
//
// Coverage is intentionally limited to the well-defined base-class pools;
// subclass resources (e.g. Battle Master Superiority Dice) and pools that lack
// a simple fixed progression are left out.
func ClassResources(class string, level, chaMod int) []ClassResource {
	var out []ClassResource
	add := func(name string, max int, r ResourceRecharge) {
		if max > 0 {
			out = append(out, ClassResource{Name: name, Max: max, Recharge: r})
		}
	}
	// addPartial adds a long-recharge pool that regains a limited number of
	// uses on a Short Rest (2024 Rage and Channel Divinity: one use per Short
	// Rest, all on a Long Rest).
	addPartial := func(name string, max, shortRecovery int) {
		if max > 0 {
			out = append(out, ClassResource{
				Name:              name,
				Max:               max,
				Recharge:          RechargeLongRest,
				ShortRestRecovery: shortRecovery,
			})
		}
	}

	switch strings.ToLower(strings.TrimSpace(class)) {
	case "barbarian":
		addPartial("Rage", rageUses(level), 1)
	case "bard":
		bardic := chaMod
		if bardic < 1 {
			bardic = 1
		}
		add("Bardic Inspiration", bardic, RechargeShortRest)
	case "cleric":
		addPartial("Channel Divinity", clericChannelDivinity(level), 1)
	case "druid":
		add("Wild Shape", wildShapeUses(level), RechargeShortRest)
	case "fighter":
		add("Second Wind", secondWindUses(level), RechargeShortRest)
		add("Action Surge", actionSurgeUses(level), RechargeShortRest)
	case "monk":
		if level >= 2 {
			add("Focus Points", level, RechargeShortRest)
		}
	case "paladin":
		add("Lay on Hands", 5*level, RechargeLongRest)
		addPartial("Channel Divinity", paladinChannelDivinity(level), 1)
	case "sorcerer":
		if level >= 2 {
			add("Sorcery Points", level, RechargeLongRest)
		}
	case "wizard":
		add("Arcane Recovery", 1, RechargeLongRest)
	}

	return out
}

func rageUses(level int) int {
	switch {
	case level >= 17:
		return 6
	case level >= 12:
		return 5
	case level >= 6:
		return 4
	case level >= 3:
		return 3
	default:
		return 2
	}
}

func clericChannelDivinity(level int) int {
	switch {
	case level >= 18:
		return 4
	case level >= 6:
		return 3
	case level >= 2:
		return 2
	default:
		return 0
	}
}

func wildShapeUses(level int) int {
	switch {
	case level >= 17:
		return 4
	case level >= 6:
		return 3
	case level >= 2:
		return 2
	default:
		return 0
	}
}

func secondWindUses(level int) int {
	switch {
	case level >= 10:
		return 4
	case level >= 4:
		return 3
	default:
		return 2
	}
}

func actionSurgeUses(level int) int {
	switch {
	case level >= 17:
		return 2
	case level >= 2:
		return 1
	default:
		return 0
	}
}

func paladinChannelDivinity(level int) int {
	switch {
	case level >= 11:
		return 3
	case level >= 3:
		return 2
	default:
		return 0
	}
}
