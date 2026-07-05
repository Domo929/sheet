package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func findResource(list []ClassResource, name string) (ClassResource, bool) {
	for _, r := range list {
		if r.Name == name {
			return r, true
		}
	}
	return ClassResource{}, false
}

func TestClassResourcesBarbarianRageScaling(t *testing.T) {
	cases := map[int]int{1: 2, 3: 3, 6: 4, 12: 5, 17: 6, 20: 6}
	for level, want := range cases {
		r, ok := findResource(ClassResources("Barbarian", level, 0), "Rage")
		assert.True(t, ok, "level %d should have Rage", level)
		assert.Equal(t, want, r.Max, "Rage uses at level %d", level)
		assert.Equal(t, RechargeLongRest, r.Recharge)
	}
}

func TestClassResourcesBardScalesWithCharisma(t *testing.T) {
	r, ok := findResource(ClassResources("Bard", 1, 3), "Bardic Inspiration")
	assert.True(t, ok)
	assert.Equal(t, 3, r.Max)
	assert.Equal(t, RechargeShortRest, r.Recharge)

	// Minimum of 1 even with a non-positive modifier.
	r, _ = findResource(ClassResources("Bard", 1, -1), "Bardic Inspiration")
	assert.Equal(t, 1, r.Max)
}

func TestClassResourcesMonkAndSorcererStartAtLevel2(t *testing.T) {
	_, ok := findResource(ClassResources("Monk", 1, 0), "Focus Points")
	assert.False(t, ok, "no Focus Points at level 1")
	r, ok := findResource(ClassResources("Monk", 5, 0), "Focus Points")
	assert.True(t, ok)
	assert.Equal(t, 5, r.Max)

	_, ok = findResource(ClassResources("Sorcerer", 1, 0), "Sorcery Points")
	assert.False(t, ok)
	r, ok = findResource(ClassResources("Sorcerer", 10, 0), "Sorcery Points")
	assert.True(t, ok)
	assert.Equal(t, 10, r.Max)
	assert.Equal(t, RechargeLongRest, r.Recharge)
}

func TestClassResourcesFighter(t *testing.T) {
	res := ClassResources("Fighter", 10, 0)
	sw, ok := findResource(res, "Second Wind")
	assert.True(t, ok)
	assert.Equal(t, 4, sw.Max)
	as, ok := findResource(res, "Action Surge")
	assert.True(t, ok)
	assert.Equal(t, 1, as.Max)

	// Action Surge only from level 2.
	_, ok = findResource(ClassResources("Fighter", 1, 0), "Action Surge")
	assert.False(t, ok)
}

func TestClassResourcesPaladinLayOnHands(t *testing.T) {
	r, ok := findResource(ClassResources("Paladin", 4, 2), "Lay on Hands")
	assert.True(t, ok)
	assert.Equal(t, 20, r.Max, "Lay on Hands pool is 5 x level")
	assert.Equal(t, RechargeLongRest, r.Recharge)

	// Channel Divinity begins at level 3.
	_, ok = findResource(ClassResources("Paladin", 2, 2), "Channel Divinity")
	assert.False(t, ok)
	cd, ok := findResource(ClassResources("Paladin", 11, 2), "Channel Divinity")
	assert.True(t, ok)
	assert.Equal(t, 3, cd.Max)
}

func TestClassResourcesNoneForResourcelessClasses(t *testing.T) {
	for _, class := range []string{"Rogue", "Warlock", "Ranger"} {
		assert.Empty(t, ClassResources(class, 20, 3), class)
	}
}
