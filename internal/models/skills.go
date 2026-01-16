package models

// ProficiencyLevel represents the level of proficiency in a skill.
type ProficiencyLevel int

const (
	NotProficient ProficiencyLevel = iota
	Proficient
	Expertise
)

// Skill represents a single skill with its proficiency level.
type Skill struct {
	Proficiency ProficiencyLevel `json:"proficiency"`
}

// SkillName identifies a D&D 5e skill.
type SkillName string

const (
	SkillAcrobatics     SkillName = "acrobatics"
	SkillAnimalHandling SkillName = "animalHandling"
	SkillArcana         SkillName = "arcana"
	SkillAthletics      SkillName = "athletics"
	SkillDeception      SkillName = "deception"
	SkillHistory        SkillName = "history"
	SkillInsight        SkillName = "insight"
	SkillIntimidation   SkillName = "intimidation"
	SkillInvestigation  SkillName = "investigation"
	SkillMedicine       SkillName = "medicine"
	SkillNature         SkillName = "nature"
	SkillPerception     SkillName = "perception"
	SkillPerformance    SkillName = "performance"
	SkillPersuasion     SkillName = "persuasion"
	SkillReligion       SkillName = "religion"
	SkillSleightOfHand  SkillName = "sleightOfHand"
	SkillStealth        SkillName = "stealth"
	SkillSurvival       SkillName = "survival"
)

// AllSkills returns a slice of all D&D 5e skill names.
func AllSkills() []SkillName {
	return []SkillName{
		SkillAcrobatics,
		SkillAnimalHandling,
		SkillArcana,
		SkillAthletics,
		SkillDeception,
		SkillHistory,
		SkillInsight,
		SkillIntimidation,
		SkillInvestigation,
		SkillMedicine,
		SkillNature,
		SkillPerception,
		SkillPerformance,
		SkillPersuasion,
		SkillReligion,
		SkillSleightOfHand,
		SkillStealth,
		SkillSurvival,
	}
}

// SkillAbilityMap maps each skill to its associated ability.
var SkillAbilityMap = map[SkillName]Ability{
	SkillAcrobatics:     AbilityDexterity,
	SkillAnimalHandling: AbilityWisdom,
	SkillArcana:         AbilityIntelligence,
	SkillAthletics:      AbilityStrength,
	SkillDeception:      AbilityCharisma,
	SkillHistory:        AbilityIntelligence,
	SkillInsight:        AbilityWisdom,
	SkillIntimidation:   AbilityCharisma,
	SkillInvestigation:  AbilityIntelligence,
	SkillMedicine:       AbilityWisdom,
	SkillNature:         AbilityIntelligence,
	SkillPerception:     AbilityWisdom,
	SkillPerformance:    AbilityCharisma,
	SkillPersuasion:     AbilityCharisma,
	SkillReligion:       AbilityIntelligence,
	SkillSleightOfHand:  AbilityDexterity,
	SkillStealth:        AbilityDexterity,
	SkillSurvival:       AbilityWisdom,
}

// Skills contains all 18 D&D 5e skills.
type Skills struct {
	Acrobatics     Skill `json:"acrobatics"`
	AnimalHandling Skill `json:"animalHandling"`
	Arcana         Skill `json:"arcana"`
	Athletics      Skill `json:"athletics"`
	Deception      Skill `json:"deception"`
	History        Skill `json:"history"`
	Insight        Skill `json:"insight"`
	Intimidation   Skill `json:"intimidation"`
	Investigation  Skill `json:"investigation"`
	Medicine       Skill `json:"medicine"`
	Nature         Skill `json:"nature"`
	Perception     Skill `json:"perception"`
	Performance    Skill `json:"performance"`
	Persuasion     Skill `json:"persuasion"`
	Religion       Skill `json:"religion"`
	SleightOfHand  Skill `json:"sleightOfHand"`
	Stealth        Skill `json:"stealth"`
	Survival       Skill `json:"survival"`
}

// NewSkills creates a new Skills with all skills at NotProficient.
func NewSkills() Skills {
	return Skills{}
}

// Get returns the Skill for the given skill name.
func (s *Skills) Get(name SkillName) *Skill {
	switch name {
	case SkillAcrobatics:
		return &s.Acrobatics
	case SkillAnimalHandling:
		return &s.AnimalHandling
	case SkillArcana:
		return &s.Arcana
	case SkillAthletics:
		return &s.Athletics
	case SkillDeception:
		return &s.Deception
	case SkillHistory:
		return &s.History
	case SkillInsight:
		return &s.Insight
	case SkillIntimidation:
		return &s.Intimidation
	case SkillInvestigation:
		return &s.Investigation
	case SkillMedicine:
		return &s.Medicine
	case SkillNature:
		return &s.Nature
	case SkillPerception:
		return &s.Perception
	case SkillPerformance:
		return &s.Performance
	case SkillPersuasion:
		return &s.Persuasion
	case SkillReligion:
		return &s.Religion
	case SkillSleightOfHand:
		return &s.SleightOfHand
	case SkillStealth:
		return &s.Stealth
	case SkillSurvival:
		return &s.Survival
	default:
		return nil
	}
}

// SetProficiency sets the proficiency level for a skill.
func (s *Skills) SetProficiency(name SkillName, level ProficiencyLevel) {
	skill := s.Get(name)
	if skill != nil {
		skill.Proficiency = level
	}
}

// GetAbility returns the ability associated with a skill.
func GetSkillAbility(name SkillName) Ability {
	return SkillAbilityMap[name]
}

// CalculateModifier calculates the total modifier for a skill check.
func CalculateSkillModifier(skill *Skill, abilityMod int, proficiencyBonus int) int {
	if skill == nil {
		return abilityMod
	}

	switch skill.Proficiency {
	case Proficient:
		return abilityMod + proficiencyBonus
	case Expertise:
		return abilityMod + (proficiencyBonus * 2)
	default:
		return abilityMod
	}
}
