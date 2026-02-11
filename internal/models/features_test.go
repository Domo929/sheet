package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFeaturesAdd(t *testing.T) {
	f := NewFeatures()

	f.AddRacialTrait("Darkvision", "Elf", "You can see in dim light within 60 feet as if it were bright light.")
	f.AddRacialTrait("Fey Ancestry", "Elf", "You have advantage on saving throws against being charmed.")

	f.AddClassFeature("Spellcasting", "Wizard 1", "You can cast wizard spells.", 1)
	f.AddClassFeature("Arcane Recovery", "Wizard 1", "You can recover spell slots on a short rest.", 1)
	f.AddClassFeature("Arcane Tradition", "Wizard 2", "You choose an arcane tradition.", 2)

	f.AddFeat("Alert", "+5 to initiative, can't be surprised.")

	assert.Len(t, f.RacialTraits, 2)
	assert.Len(t, f.ClassFeatures, 3)
	assert.Len(t, f.Feats, 1)
}

func TestFeaturesAllFeatures(t *testing.T) {
	f := NewFeatures()

	f.AddRacialTrait("Darkvision", "Elf", "See in the dark.")
	f.AddClassFeature("Sneak Attack", "Rogue 1", "Extra damage.", 1)
	f.AddFeat("Lucky", "Reroll dice.")

	all := f.AllFeatures()
	assert.Len(t, all, 3)
}

func TestFeatureLevel(t *testing.T) {
	f := NewFeatures()
	f.AddClassFeature("Extra Attack", "Fighter 5", "Attack twice.", 5)

	assert.Equal(t, 5, f.ClassFeatures[0].Level)
}

func TestProficienciesArmor(t *testing.T) {
	p := NewProficiencies()

	p.AddArmor("Light Armor")
	p.AddArmor("Medium Armor")
	p.AddArmor("Light Armor") // Duplicate

	assert.Len(t, p.Armor, 2)
	assert.True(t, p.HasArmor("Light Armor"), "Should have Light Armor proficiency")
	assert.False(t, p.HasArmor("Heavy Armor"), "Should not have Heavy Armor proficiency")
}

func TestProficienciesWeapons(t *testing.T) {
	p := NewProficiencies()

	p.AddWeapon("Simple Weapons")
	p.AddWeapon("Martial Weapons")
	p.AddWeapon("Simple Weapons") // Duplicate

	assert.Len(t, p.Weapons, 2)
	assert.True(t, p.HasWeapon("Martial Weapons"), "Should have Martial Weapons proficiency")
}

func TestProficienciesTools(t *testing.T) {
	p := NewProficiencies()

	p.AddTool("Thieves' Tools")
	p.AddTool("Herbalism Kit")

	assert.Len(t, p.Tools, 2)
	assert.True(t, p.HasTool("Thieves' Tools"), "Should have Thieves' Tools proficiency")
}

func TestProficienciesLanguages(t *testing.T) {
	p := NewProficiencies()

	p.AddLanguage("Common")
	p.AddLanguage("Elvish")
	p.AddLanguage("Common") // Duplicate

	assert.Len(t, p.Languages, 2)
	assert.True(t, p.HasLanguage("Elvish"), "Should know Elvish")
	assert.False(t, p.HasLanguage("Dwarvish"), "Should not know Dwarvish")
}
