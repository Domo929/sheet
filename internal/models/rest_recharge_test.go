package models

import "testing"

func TestRestoreResources_ShortRestPartialAndFull(t *testing.T) {
	c := &Character{Resources: []ResourcePool{
		{Name: "Rage", Current: 0, Max: 4, Recharge: "long", ShortRestRecovery: 1},
		{Name: "Focus Points", Current: 0, Max: 5, Recharge: "short"},
		{Name: "Sorcery Points", Current: 0, Max: 5, Recharge: "long"},
	}}

	c.RestoreResources(false) // short rest

	got := map[string]int{}
	for _, p := range c.Resources {
		got[p.Name] = p.Current
	}
	if got["Rage"] != 1 {
		t.Errorf("Rage after short rest = %d, want 1 (partial recovery)", got["Rage"])
	}
	if got["Focus Points"] != 5 {
		t.Errorf("Focus Points after short rest = %d, want 5 (full)", got["Focus Points"])
	}
	if got["Sorcery Points"] != 0 {
		t.Errorf("Sorcery Points after short rest = %d, want 0 (long only)", got["Sorcery Points"])
	}
}

func TestRestoreResources_ShortRestRecoveryCapsAtMax(t *testing.T) {
	c := &Character{Resources: []ResourcePool{
		{Name: "Rage", Current: 4, Max: 4, Recharge: "long", ShortRestRecovery: 1},
	}}
	c.RestoreResources(false)
	if c.Resources[0].Current != 4 {
		t.Errorf("Rage = %d, want 4 (capped at max)", c.Resources[0].Current)
	}
}

func TestRestoreResources_LongRestRefillsEverything(t *testing.T) {
	c := &Character{Resources: []ResourcePool{
		{Name: "Rage", Current: 1, Max: 4, Recharge: "long", ShortRestRecovery: 1},
		{Name: "Sorcery Points", Current: 0, Max: 5, Recharge: "long"},
	}}
	c.RestoreResources(true)
	for _, p := range c.Resources {
		if p.Current != p.Max {
			t.Errorf("%s after long rest = %d, want %d", p.Name, p.Current, p.Max)
		}
	}
}

func TestChannelDivinity_ShortRestRegainsOnePerRest(t *testing.T) {
	c := NewCharacter("id", "Cleric Test", "human", "Cleric")
	c.Info.Level = 6
	c.SyncResources()

	cd := findPool(c, "Channel Divinity")
	if cd == nil {
		t.Fatal("expected Channel Divinity pool for level 6 cleric")
	}
	if cd.Max != 3 {
		t.Fatalf("Channel Divinity Max = %d, want 3", cd.Max)
	}

	// Spend all uses.
	c.SpendResource("Channel Divinity", 3)
	if findPool(c, "Channel Divinity").Current != 0 {
		t.Fatal("expected 0 remaining after spending all")
	}

	c.ShortRest()
	if got := findPool(c, "Channel Divinity").Current; got != 1 {
		t.Errorf("after first short rest = %d, want 1", got)
	}
	c.ShortRest()
	if got := findPool(c, "Channel Divinity").Current; got != 2 {
		t.Errorf("after second short rest = %d, want 2", got)
	}
	c.LongRest()
	if got := findPool(c, "Channel Divinity").Current; got != 3 {
		t.Errorf("after long rest = %d, want 3 (full)", got)
	}
}

func findPool(c *Character, name string) *ResourcePool {
	for i := range c.Resources {
		if c.Resources[i].Name == name {
			return &c.Resources[i]
		}
	}
	return nil
}
