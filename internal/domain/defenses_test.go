package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDamageDefensesSingleResistance(t *testing.T) {
	res, imm, vuln := ParseDamageDefenses([]string{
		"You have Resistance to Poison damage and Advantage on saving throws against poison.",
	})
	assert.Equal(t, []DamageType{DamagePoison}, res)
	assert.Empty(t, imm)
	assert.Empty(t, vuln)
}

func TestParseDamageDefensesMultipleTypesInOneClause(t *testing.T) {
	res, _, _ := ParseDamageDefenses([]string{
		"You have Resistance to Lightning and Thunder damage.",
	})
	assert.Equal(t, []DamageType{DamageLightning, DamageThunder}, res)
}

func TestParseDamageDefensesIgnoresGenericPhrasing(t *testing.T) {
	// Dragonborn's generic trait and the "chosen damage type" feat wording
	// must not be misread as concrete damage types.
	res, imm, vuln := ParseDamageDefenses([]string{
		"You have Resistance to the damage type determined by your Draconic Ancestry.",
		"You gain Resistance to the chosen damage type.",
	})
	assert.Empty(t, res)
	assert.Empty(t, imm)
	assert.Empty(t, vuln)
}

func TestParseDamageDefensesImmunityAndVulnerability(t *testing.T) {
	res, imm, vuln := ParseDamageDefenses([]string{
		"You have Immunity to Fire damage.",
		"You have Vulnerability to Cold damage.",
	})
	assert.Empty(t, res)
	assert.Equal(t, []DamageType{DamageFire}, imm)
	assert.Equal(t, []DamageType{DamageCold}, vuln)
}

func TestParseDamageDefensesDedupesAndSorts(t *testing.T) {
	res, _, _ := ParseDamageDefenses([]string{
		"You have Resistance to Fire damage.",
		"Resistance to Fire damage.",
		"You have Resistance to Acid damage.",
	})
	assert.Equal(t, []DamageType{DamageAcid, DamageFire}, res)
}

func TestParseDamageDefensesIgnoresUnrelatedDamageMentions(t *testing.T) {
	res, imm, vuln := ParseDamageDefenses([]string{
		"When you hit with this weapon you deal an extra 1d6 Fire damage.",
		"This creature takes 10 Necrotic damage at the start of its turn.",
	})
	assert.Empty(t, res)
	assert.Empty(t, imm)
	assert.Empty(t, vuln)
}

func TestDraconicAncestryResistance(t *testing.T) {
	cases := map[string]DamageType{
		"Red":    DamageFire,
		"gold":   DamageFire,
		"brass":  DamageFire,
		"White":  DamageCold,
		"silver": DamageCold,
		"Blue":   DamageLightning,
		"bronze": DamageLightning,
		"Black":  DamageAcid,
		"copper": DamageAcid,
		"Green":  DamagePoison,
	}
	for ancestry, want := range cases {
		got, ok := DraconicAncestryResistance(ancestry)
		assert.True(t, ok, "ancestry %q should resolve", ancestry)
		assert.Equal(t, want, got, "ancestry %q", ancestry)
	}

	_, ok := DraconicAncestryResistance("Chartreuse")
	assert.False(t, ok, "unknown ancestry should not resolve")
}

func TestDamageTypeTitle(t *testing.T) {
	assert.Equal(t, "Fire", DamageFire.Title())
	assert.Equal(t, "Bludgeoning", DamageBludgeoning.Title())
	assert.Equal(t, "", DamageType("").Title())
}

func TestWeaponMasteryCount(t *testing.T) {
	// Fighter scales with level.
	assert.Equal(t, 3, WeaponMasteryCount("Fighter", 1))
	assert.Equal(t, 3, WeaponMasteryCount("Fighter", 3))
	assert.Equal(t, 4, WeaponMasteryCount("Fighter", 4))
	assert.Equal(t, 4, WeaponMasteryCount("Fighter", 9))
	assert.Equal(t, 5, WeaponMasteryCount("Fighter", 10))
	assert.Equal(t, 5, WeaponMasteryCount("Fighter", 15))
	assert.Equal(t, 6, WeaponMasteryCount("Fighter", 16))
	assert.Equal(t, 6, WeaponMasteryCount("Fighter", 20))

	// Fixed-count martial classes.
	for _, class := range []string{"Barbarian", "Paladin", "Ranger", "Rogue"} {
		assert.Equal(t, 2, WeaponMasteryCount(class, 1), class)
		assert.Equal(t, 2, WeaponMasteryCount(class, 20), class)
	}

	// Case-insensitive.
	assert.Equal(t, 3, WeaponMasteryCount("fighter", 1))
	assert.Equal(t, 2, WeaponMasteryCount("rogue", 5))

	// Classes without Weapon Mastery.
	for _, class := range []string{"Wizard", "Cleric", "Monk", "Sorcerer", "Warlock", "Druid", "Bard", ""} {
		assert.Equal(t, 0, WeaponMasteryCount(class, 20), class)
	}
}
