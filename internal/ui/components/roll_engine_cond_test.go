package components

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// When a condition imposes disadvantage and the user picks Normal, the roll
// should be made with disadvantage.
func TestRollEngine_ConditionDisadvantageAppliesOnNormal(t *testing.T) {
	e := NewRollEngine()
	e.Update(RequestRollMsg{
		DiceExpr:  "1d20",
		AdvPrompt: true,
		Cond:      ConditionEffect{Disadvantage: true, Reason: "Poisoned"},
	})

	e.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})

	if !e.disadvantage {
		t.Error("want disadvantage applied from condition on Normal pick")
	}
	if e.advantage {
		t.Error("did not expect advantage")
	}
}

// A user-selected advantage cancels a condition disadvantage (5e stacking).
func TestRollEngine_ConditionDisadvantageCanceledByAdvantage(t *testing.T) {
	e := NewRollEngine()
	e.Update(RequestRollMsg{
		DiceExpr:  "1d20",
		AdvPrompt: true,
		Cond:      ConditionEffect{Disadvantage: true, Reason: "Poisoned"},
	})

	e.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})

	if e.advantage || e.disadvantage {
		t.Errorf("want normal roll after cancel, got adv=%v dis=%v", e.advantage, e.disadvantage)
	}
}

// A condition advantage combines with a user Normal pick to yield advantage.
func TestRollEngine_ConditionAdvantageAppliesOnNormal(t *testing.T) {
	e := NewRollEngine()
	e.Update(RequestRollMsg{
		DiceExpr:  "1d20",
		AdvPrompt: true,
		Cond:      ConditionEffect{Advantage: true, Reason: "Invisible"},
	})

	e.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})

	if !e.advantage || e.disadvantage {
		t.Errorf("want advantage, got adv=%v dis=%v", e.advantage, e.disadvantage)
	}
}

// A non-prompt roll with a condition disadvantage still applies it.
func TestRollEngine_ConditionDisadvantageWithoutPrompt(t *testing.T) {
	e := NewRollEngine()
	e.Update(RequestRollMsg{
		DiceExpr:  "1d20",
		AdvPrompt: false,
		Cond:      ConditionEffect{Disadvantage: true, Reason: "Prone"},
	})

	if !e.disadvantage {
		t.Error("want disadvantage applied on non-prompt roll")
	}
}

func TestRollEngine_CombineAdvStacking(t *testing.T) {
	cases := []struct {
		name             string
		condAdv, condDis bool
		userAdv, userDis bool
		wantAdv, wantDis bool
	}{
		{"none", false, false, false, false, false, false},
		{"cond disadv only", false, true, false, false, false, true},
		{"cond adv only", true, false, false, false, true, false},
		{"cond disadv + user adv cancel", false, true, true, false, false, false},
		{"cond adv + user disadv cancel", true, false, false, true, false, false},
		{"user disadv only", false, false, false, true, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewRollEngine()
			e.condEffect = ConditionEffect{Advantage: tc.condAdv, Disadvantage: tc.condDis}
			adv, dis := e.combineAdv(tc.userAdv, tc.userDis)
			if adv != tc.wantAdv || dis != tc.wantDis {
				t.Errorf("combineAdv = (%v,%v), want (%v,%v)", adv, dis, tc.wantAdv, tc.wantDis)
			}
		})
	}
}
