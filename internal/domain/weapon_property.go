package domain

// WeaponProperty represents properties a weapon can have.
type WeaponProperty string

const (
	PropertyFinesse    WeaponProperty = "finesse"
	PropertyLight      WeaponProperty = "light"
	PropertyHeavy      WeaponProperty = "heavy"
	PropertyReach      WeaponProperty = "reach"
	PropertyThrown     WeaponProperty = "thrown"
	PropertyVersatile  WeaponProperty = "versatile"
	PropertyTwoHanded  WeaponProperty = "two-handed"
	PropertyAmmunition WeaponProperty = "ammunition"
	PropertyLoading    WeaponProperty = "loading"
)
