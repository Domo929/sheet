package domain

import "testing"

func TestClassResources_PartialShortRestRecovery(t *testing.T) {
	cases := []struct {
		class string
		level int
	}{
		{"Barbarian", 5},
		{"Cleric", 6},
		{"Paladin", 11},
	}
	for _, tc := range cases {
		r, ok := findResource(ClassResources(tc.class, tc.level, 3), map1Name(tc.class))
		if !ok {
			t.Fatalf("%s should have its channeled/rage resource", tc.class)
		}
		if r.Recharge != RechargeLongRest {
			t.Errorf("%s %s: recharge = %q, want long", tc.class, r.Name, r.Recharge)
		}
		if r.ShortRestRecovery != 1 {
			t.Errorf("%s %s: ShortRestRecovery = %d, want 1", tc.class, r.Name, r.ShortRestRecovery)
		}
	}
}

// map1Name returns the partial-recovery resource name for a class.
func map1Name(class string) string {
	if class == "Barbarian" {
		return "Rage"
	}
	return "Channel Divinity"
}

func TestClassResources_ShortRechargePoolsHaveNoPartialField(t *testing.T) {
	// Fully short-rest-recharging pools should not set ShortRestRecovery.
	res := ClassResources("Fighter", 10, 0)
	for _, r := range res {
		if r.Recharge == RechargeShortRest && r.ShortRestRecovery != 0 {
			t.Errorf("%s: short-recharge pool should have ShortRestRecovery 0, got %d", r.Name, r.ShortRestRecovery)
		}
	}
}

func TestClassResources_LayOnHandsNoShortRecovery(t *testing.T) {
	r, ok := findResource(ClassResources("Paladin", 5, 2), "Lay on Hands")
	if !ok {
		t.Fatal("expected Lay on Hands")
	}
	if r.Recharge != RechargeLongRest || r.ShortRestRecovery != 0 {
		t.Errorf("Lay on Hands: recharge=%q recovery=%d, want long/0", r.Recharge, r.ShortRestRecovery)
	}
}
