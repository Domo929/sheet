package data

import "testing"

func TestSpellRequiresConcentration(t *testing.T) {
	cases := []struct {
		duration string
		want     bool
	}{
		{"Concentration, up to 1 minute", true},
		{"Concentration, up to 1 hour", true},
		{"concentration, up to 10 minutes", true},
		{"Instantaneous", false},
		{"1 minute", false},
		{"8 hours", false},
		{"Until dispelled", false},
	}
	for _, c := range cases {
		s := SpellData{Duration: c.duration}
		if got := s.RequiresConcentration(); got != c.want {
			t.Errorf("RequiresConcentration(%q) = %v, want %v", c.duration, got, c.want)
		}
	}
}
