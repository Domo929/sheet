package models

import "testing"

func TestFeaturesAdd(t *testing.T) {
	f := NewFeatures()

	f.AddRacialTrait("Darkvision", "Elf", "You can see in dim light within 60 feet as if it were bright light.")
	f.AddRacialTrait("Fey Ancestry", "Elf", "You have advantage on saving throws against being charmed.")

	f.AddClassFeature("Spellcasting", "Wizard 1", "You can cast wizard spells.", 1)
	f.AddClassFeature("Arcane Recovery", "Wizard 1", "You can recover spell slots on a short rest.", 1)
	f.AddClassFeature("Arcane Tradition", "Wizard 2", "You choose an arcane tradition.", 2)

	f.AddFeat("Alert", "+5 to initiative, can't be surprised.")

	if len(f.RacialTraits) != 2 {
		t.Errorf("RacialTraits count = %d, want 2", len(f.RacialTraits))
	}
	if len(f.ClassFeatures) != 3 {
		t.Errorf("ClassFeatures count = %d, want 3", len(f.ClassFeatures))
	}
	if len(f.Feats) != 1 {
		t.Errorf("Feats count = %d, want 1", len(f.Feats))
	}
}

func TestFeaturesAllFeatures(t *testing.T) {
	f := NewFeatures()

	f.AddRacialTrait("Darkvision", "Elf", "See in the dark.")
	f.AddClassFeature("Sneak Attack", "Rogue 1", "Extra damage.", 1)
	f.AddFeat("Lucky", "Reroll dice.")

	all := f.AllFeatures()
	if len(all) != 3 {
		t.Errorf("AllFeatures() count = %d, want 3", len(all))
	}
}

func TestFeatureLevel(t *testing.T) {
	f := NewFeatures()
	f.AddClassFeature("Extra Attack", "Fighter 5", "Attack twice.", 5)

	if f.ClassFeatures[0].Level != 5 {
		t.Errorf("Feature level = %d, want 5", f.ClassFeatures[0].Level)
	}
}

func TestProficienciesArmor(t *testing.T) {
	p := NewProficiencies()

	p.AddArmor("Light Armor")
	p.AddArmor("Medium Armor")
	p.AddArmor("Light Armor") // Duplicate

	if len(p.Armor) != 2 {
		t.Errorf("Armor count = %d, want 2", len(p.Armor))
	}

	if !p.HasArmor("Light Armor") {
		t.Error("Should have Light Armor proficiency")
	}
	if p.HasArmor("Heavy Armor") {
		t.Error("Should not have Heavy Armor proficiency")
	}
}

func TestProficienciesWeapons(t *testing.T) {
	p := NewProficiencies()

	p.AddWeapon("Simple Weapons")
	p.AddWeapon("Martial Weapons")
	p.AddWeapon("Simple Weapons") // Duplicate

	if len(p.Weapons) != 2 {
		t.Errorf("Weapons count = %d, want 2", len(p.Weapons))
	}

	if !p.HasWeapon("Martial Weapons") {
		t.Error("Should have Martial Weapons proficiency")
	}
}

func TestProficienciesTools(t *testing.T) {
	p := NewProficiencies()

	p.AddTool("Thieves' Tools")
	p.AddTool("Herbalism Kit")

	if len(p.Tools) != 2 {
		t.Errorf("Tools count = %d, want 2", len(p.Tools))
	}

	if !p.HasTool("Thieves' Tools") {
		t.Error("Should have Thieves' Tools proficiency")
	}
}

func TestProficienciesLanguages(t *testing.T) {
	p := NewProficiencies()

	p.AddLanguage("Common")
	p.AddLanguage("Elvish")
	p.AddLanguage("Common") // Duplicate

	if len(p.Languages) != 2 {
		t.Errorf("Languages count = %d, want 2", len(p.Languages))
	}

	if !p.HasLanguage("Elvish") {
		t.Error("Should know Elvish")
	}
	if p.HasLanguage("Dwarvish") {
		t.Error("Should not know Dwarvish")
	}
}
