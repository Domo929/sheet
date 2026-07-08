package models

import (
	"fmt"
	"strings"
)

// MulticlassRequirement describes the ability-score prerequisite that must be
// met to take a level in a class when multiclassing (2024 rules: a score of at
// least 13 in the class's primary ability). When RequireAll is true every
// listed ability must be 13+ (e.g., Monk needs Dexterity AND Wisdom); otherwise
// any one of them satisfies the requirement (e.g., Fighter needs Strength OR
// Dexterity).
type MulticlassRequirement struct {
	Abilities  []Ability
	RequireAll bool
}

// multiclassMinimum is the ability-score minimum required to multiclass (2024).
const multiclassMinimum = 13

// multiclassPrereqs maps a lowercased class name to its 2024 multiclassing
// ability prerequisite.
var multiclassPrereqs = map[string]MulticlassRequirement{
	"barbarian": {Abilities: []Ability{AbilityStrength}},
	"bard":      {Abilities: []Ability{AbilityCharisma}},
	"cleric":    {Abilities: []Ability{AbilityWisdom}},
	"druid":     {Abilities: []Ability{AbilityWisdom}},
	"fighter":   {Abilities: []Ability{AbilityStrength, AbilityDexterity}, RequireAll: false},
	"monk":      {Abilities: []Ability{AbilityDexterity, AbilityWisdom}, RequireAll: true},
	"paladin":   {Abilities: []Ability{AbilityStrength, AbilityCharisma}, RequireAll: true},
	"ranger":    {Abilities: []Ability{AbilityDexterity, AbilityWisdom}, RequireAll: true},
	"rogue":     {Abilities: []Ability{AbilityDexterity}},
	"sorcerer":  {Abilities: []Ability{AbilityCharisma}},
	"warlock":   {Abilities: []Ability{AbilityCharisma}},
	"wizard":    {Abilities: []Ability{AbilityIntelligence}},
	// Artificer (optional, non-PHB) uses Intelligence.
	"artificer": {Abilities: []Ability{AbilityIntelligence}},
}

// MulticlassPrerequisite returns the ability prerequisite for taking levels in
// the named class, and whether a prerequisite is known for that class.
func MulticlassPrerequisite(class string) (MulticlassRequirement, bool) {
	req, ok := multiclassPrereqs[strings.ToLower(strings.TrimSpace(class))]
	return req, ok
}

// abilityLabel returns a capitalized, human-readable label for an ability.
func abilityLabel(a Ability) string {
	s := string(a)
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// meetsRequirement reports whether the given ability scores satisfy a
// multiclass requirement, returning the abilities that fall short when they do
// not. For an OR requirement, the missing list is only populated when none of
// the options qualify (and then it lists all options).
func meetsRequirement(scores AbilityScores, req MulticlassRequirement) (bool, []Ability) {
	if len(req.Abilities) == 0 {
		return true, nil
	}
	if req.RequireAll {
		var missing []Ability
		for _, ab := range req.Abilities {
			if scores.Get(ab).Total() < multiclassMinimum {
				missing = append(missing, ab)
			}
		}
		return len(missing) == 0, missing
	}
	// OR semantics: any one qualifies.
	for _, ab := range req.Abilities {
		if scores.Get(ab).Total() >= multiclassMinimum {
			return true, nil
		}
	}
	return false, append([]Ability(nil), req.Abilities...)
}

// requirementText renders a prerequisite as human-readable text, e.g.
// "Strength or Dexterity 13" or "Dexterity and Wisdom 13".
func requirementText(req MulticlassRequirement) string {
	if len(req.Abilities) == 0 {
		return "no prerequisite"
	}
	labels := make([]string, 0, len(req.Abilities))
	for _, ab := range req.Abilities {
		labels = append(labels, abilityLabel(ab))
	}
	joiner := " or "
	if req.RequireAll {
		joiner = " and "
	}
	return fmt.Sprintf("%s %d", strings.Join(labels, joiner), multiclassMinimum)
}

// titleClass returns a class name with its first letter capitalized.
func titleClass(class string) string {
	s := strings.TrimSpace(strings.ToLower(class))
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// CanMulticlassInto reports whether this character satisfies the 2024
// multiclassing ability prerequisites to gain a level in the named class. Per
// the rules, the character must meet the primary-ability requirement of both
// the new class AND all of their current classes. When it returns false, the
// reason string explains which requirement was not met.
func (c *Character) CanMulticlassInto(class string) (bool, string) {
	scores := c.AbilityScores

	// New class requirement.
	if req, ok := MulticlassPrerequisite(class); ok {
		if met, _ := meetsRequirement(scores, req); !met {
			return false, fmt.Sprintf("Requires %s to multiclass into %s", requirementText(req), titleClass(class))
		}
	}

	// Current classes' requirements (you must also still qualify for what you have).
	for _, name := range c.currentClassNames() {
		if strings.EqualFold(name, class) {
			continue
		}
		if req, ok := MulticlassPrerequisite(name); ok {
			if met, _ := meetsRequirement(scores, req); !met {
				return false, fmt.Sprintf("Requires %s (from your %s levels) to multiclass", requirementText(req), titleClass(name))
			}
		}
	}
	return true, ""
}

// currentClassNames returns the character's current class names, falling back to
// the single-class Info.Class when no multiclass breakdown is present.
func (c *Character) currentClassNames() []string {
	if len(c.Classes) > 0 {
		names := make([]string, 0, len(c.Classes))
		for _, cl := range c.Classes {
			names = append(names, cl.Class)
		}
		return names
	}
	if c.Info.Class != "" {
		return []string{c.Info.Class}
	}
	return nil
}

// MulticlassProficiencyGrant is the reduced set of proficiencies a character
// gains when taking their first level in a class other than their initial
// class (2024 rules). Concrete Armor/Weapons/Tools proficiencies are granted
// automatically; Notes describe choices (skills, instruments) the player makes
// manually.
type MulticlassProficiencyGrant struct {
	Armor   []string
	Weapons []string
	Tools   []string
	Notes   []string
}

// IsEmpty reports whether the grant confers nothing.
func (g MulticlassProficiencyGrant) IsEmpty() bool {
	return len(g.Armor) == 0 && len(g.Weapons) == 0 && len(g.Tools) == 0 && len(g.Notes) == 0
}

// multiclassProfs maps a lowercased class name to the proficiencies gained when
// multiclassing into it. This is the canonical 5e multiclassing proficiency
// table preserved by the 2024 Player's Handbook. Skill and instrument choices
// are surfaced as Notes because they require a player decision.
var multiclassProfs = map[string]MulticlassProficiencyGrant{
	"barbarian": {Armor: []string{"Shields"}, Weapons: []string{"Simple Weapons", "Martial Weapons"}},
	"bard":      {Armor: []string{"Light Armor"}, Notes: []string{"One skill of your choice", "One Musical Instrument"}},
	"cleric":    {Armor: []string{"Light Armor", "Medium Armor", "Shields"}},
	"druid":     {Armor: []string{"Light Armor", "Medium Armor", "Shields"}},
	"fighter":   {Armor: []string{"Light Armor", "Medium Armor", "Shields"}, Weapons: []string{"Simple Weapons", "Martial Weapons"}},
	"monk":      {Weapons: []string{"Simple Weapons", "Shortswords"}},
	"paladin":   {Armor: []string{"Light Armor", "Medium Armor", "Shields"}, Weapons: []string{"Simple Weapons", "Martial Weapons"}},
	"ranger":    {Armor: []string{"Light Armor", "Medium Armor", "Shields"}, Weapons: []string{"Simple Weapons", "Martial Weapons"}, Notes: []string{"One skill from the Ranger list"}},
	"rogue":     {Armor: []string{"Light Armor"}, Tools: []string{"Thieves' Tools"}, Notes: []string{"One skill from the Rogue list"}},
	"sorcerer":  {},
	"warlock":   {Armor: []string{"Light Armor"}, Weapons: []string{"Simple Weapons"}},
	"wizard":    {},
	"artificer": {Armor: []string{"Light Armor", "Medium Armor", "Shields"}, Tools: []string{"Thieves' Tools", "Tinker's Tools"}},
}

// MulticlassProficiencies returns the proficiencies gained when multiclassing
// into the named class, and whether an entry is known for that class.
func MulticlassProficiencies(class string) (MulticlassProficiencyGrant, bool) {
	g, ok := multiclassProfs[strings.ToLower(strings.TrimSpace(class))]
	return g, ok
}

// GrantMulticlassProficiencies adds the concrete armor, weapon, and tool
// proficiencies gained from multiclassing into the named class to the
// character (deduplicated). Skill/instrument choices are not applied
// automatically. It returns the grant that was applied so callers can report
// any manual choices to the player.
func (c *Character) GrantMulticlassProficiencies(class string) MulticlassProficiencyGrant {
	grant, ok := MulticlassProficiencies(class)
	if !ok {
		return MulticlassProficiencyGrant{}
	}
	for _, a := range grant.Armor {
		c.Proficiencies.AddArmor(a)
	}
	for _, w := range grant.Weapons {
		c.Proficiencies.AddWeapon(w)
	}
	for _, t := range grant.Tools {
		c.Proficiencies.AddTool(t)
	}
	return grant
}
