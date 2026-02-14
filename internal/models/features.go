package models

// Feature represents a class feature, racial trait, or feat.
type Feature struct {
	Name        string `json:"name"`
	Source      string `json:"source"` // e.g., "Cleric 1", "Elf", "Alert (Feat)"
	Description string `json:"description"`
	Level       int    `json:"level,omitempty"`      // Level gained (for class features)
	Activation  string `json:"activation,omitempty"` // "action", "bonus", "reaction", or "" (passive)
}

// NewFeature creates a new feature.
func NewFeature(name, source, description string) Feature {
	return Feature{
		Name:        name,
		Source:      source,
		Description: description,
	}
}

// Features contains all character features organized by source.
type Features struct {
	RacialTraits  []Feature `json:"racialTraits,omitempty"`
	ClassFeatures []Feature `json:"classFeatures,omitempty"`
	Feats         []Feature `json:"feats,omitempty"`
}

// NewFeatures creates an empty features container.
func NewFeatures() Features {
	return Features{
		RacialTraits:  []Feature{},
		ClassFeatures: []Feature{},
		Feats:         []Feature{},
	}
}

// AddRacialTrait adds a racial trait.
func (f *Features) AddRacialTrait(name, source, description string) {
	f.RacialTraits = append(f.RacialTraits, NewFeature(name, source, description))
}

// AddClassFeature adds a class feature.
func (f *Features) AddClassFeature(name, source, description string, level int, activation string) {
	feature := NewFeature(name, source, description)
	feature.Level = level
	feature.Activation = activation
	f.ClassFeatures = append(f.ClassFeatures, feature)
}

// AddFeat adds a feat.
func (f *Features) AddFeat(name, description string) {
	f.Feats = append(f.Feats, NewFeature(name, "Feat", description))
}

// AllFeatures returns all features as a single slice.
func (f *Features) AllFeatures() []Feature {
	all := []Feature{}
	all = append(all, f.RacialTraits...)
	all = append(all, f.ClassFeatures...)
	all = append(all, f.Feats...)
	return all
}

// Proficiencies tracks all character proficiencies.
type Proficiencies struct {
	Armor     []string `json:"armor,omitempty"`
	Weapons   []string `json:"weapons,omitempty"`
	Tools     []string `json:"tools,omitempty"`
	Languages []string `json:"languages,omitempty"`
}

// NewProficiencies creates an empty proficiencies container.
func NewProficiencies() Proficiencies {
	return Proficiencies{
		Armor:     []string{},
		Weapons:   []string{},
		Tools:     []string{},
		Languages: []string{},
	}
}

// AddArmor adds an armor proficiency if not already present.
func (p *Proficiencies) AddArmor(armor string) {
	if !p.HasArmor(armor) {
		p.Armor = append(p.Armor, armor)
	}
}

// AddWeapon adds a weapon proficiency if not already present.
func (p *Proficiencies) AddWeapon(weapon string) {
	if !p.HasWeapon(weapon) {
		p.Weapons = append(p.Weapons, weapon)
	}
}

// AddTool adds a tool proficiency if not already present.
func (p *Proficiencies) AddTool(tool string) {
	if !p.HasTool(tool) {
		p.Tools = append(p.Tools, tool)
	}
}

// AddLanguage adds a language if not already present.
func (p *Proficiencies) AddLanguage(language string) {
	if !p.HasLanguage(language) {
		p.Languages = append(p.Languages, language)
	}
}

// HasArmor checks if character is proficient with the armor.
func (p *Proficiencies) HasArmor(armor string) bool {
	return containsString(p.Armor, armor)
}

// HasWeapon checks if character is proficient with the weapon.
func (p *Proficiencies) HasWeapon(weapon string) bool {
	return containsString(p.Weapons, weapon)
}

// HasTool checks if character is proficient with the tool.
func (p *Proficiencies) HasTool(tool string) bool {
	return containsString(p.Tools, tool)
}

// HasLanguage checks if character knows the language.
func (p *Proficiencies) HasLanguage(language string) bool {
	return containsString(p.Languages, language)
}

// containsString checks if a slice contains a string.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
