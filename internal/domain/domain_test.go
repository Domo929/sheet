package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActivationTypes(t *testing.T) {
	assert.Equal(t, ActivationType("action"), ActivationAction)
	assert.Equal(t, ActivationType("bonus"), ActivationBonus)
	assert.Equal(t, ActivationType("reaction"), ActivationReaction)
	assert.Equal(t, ActivationType(""), ActivationPassive)
}

func TestDamageTypes(t *testing.T) {
	expected := []DamageType{
		DamageAcid, DamageBludgeoning, DamageCold, DamageFire,
		DamageForce, DamageLightning, DamageNecrotic, DamagePiercing,
		DamagePoison, DamagePsychic, DamageRadiant, DamageSlashing,
		DamageThunder,
	}
	assert.Len(t, expected, 13, "should have 13 damage types")

	// Verify string values match D&D conventions
	assert.Equal(t, DamageType("fire"), DamageFire)
	assert.Equal(t, DamageType("bludgeoning"), DamageBludgeoning)
	assert.Equal(t, DamageType("radiant"), DamageRadiant)
}

func TestWeaponProperties(t *testing.T) {
	expected := []WeaponProperty{
		PropertyFinesse, PropertyLight, PropertyHeavy, PropertyReach,
		PropertyThrown, PropertyVersatile, PropertyTwoHanded,
		PropertyAmmunition, PropertyLoading,
	}
	assert.Len(t, expected, 9, "should have 9 weapon properties")

	// Verify string values
	assert.Equal(t, WeaponProperty("finesse"), PropertyFinesse)
	assert.Equal(t, WeaponProperty("two-handed"), PropertyTwoHanded)
	assert.Equal(t, WeaponProperty("ammunition"), PropertyAmmunition)
}
