package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllSkillsCount(t *testing.T) {
	skills := AllSkills()
	assert.Len(t, skills, 18)
}

func TestSkillAbilityMapping(t *testing.T) {
	tests := []struct {
		skill   SkillName
		ability Ability
	}{
		{SkillAcrobatics, AbilityDexterity},
		{SkillAnimalHandling, AbilityWisdom},
		{SkillArcana, AbilityIntelligence},
		{SkillAthletics, AbilityStrength},
		{SkillDeception, AbilityCharisma},
		{SkillHistory, AbilityIntelligence},
		{SkillInsight, AbilityWisdom},
		{SkillIntimidation, AbilityCharisma},
		{SkillInvestigation, AbilityIntelligence},
		{SkillMedicine, AbilityWisdom},
		{SkillNature, AbilityIntelligence},
		{SkillPerception, AbilityWisdom},
		{SkillPerformance, AbilityCharisma},
		{SkillPersuasion, AbilityCharisma},
		{SkillReligion, AbilityIntelligence},
		{SkillSleightOfHand, AbilityDexterity},
		{SkillStealth, AbilityDexterity},
		{SkillSurvival, AbilityWisdom},
	}

	for _, tt := range tests {
		t.Run(string(tt.skill), func(t *testing.T) {
			assert.Equal(t, tt.ability, GetSkillAbility(tt.skill))
		})
	}
}

func TestSkillsGet(t *testing.T) {
	skills := NewSkills()
	skills.Stealth.Proficiency = Expertise

	for _, skillName := range AllSkills() {
		skill := skills.Get(skillName)
		assert.NotNil(t, skill, "Skills.Get(%s) returned nil", skillName)
	}

	// Check specific skill we modified
	stealth := skills.Get(SkillStealth)
	assert.Equal(t, Expertise, stealth.Proficiency)
}

func TestSkillsGetInvalid(t *testing.T) {
	skills := NewSkills()
	assert.Nil(t, skills.Get(SkillName("invalid")))
}

func TestSkillsSetProficiency(t *testing.T) {
	skills := NewSkills()

	skills.SetProficiency(SkillPerception, Proficient)
	assert.Equal(t, Proficient, skills.Perception.Proficiency)

	skills.SetProficiency(SkillStealth, Expertise)
	assert.Equal(t, Expertise, skills.Stealth.Proficiency)
}

func TestCalculateSkillModifier(t *testing.T) {
	tests := []struct {
		name             string
		proficiency      ProficiencyLevel
		abilityMod       int
		proficiencyBonus int
		expected         int
	}{
		{"not proficient, +2 ability", NotProficient, 2, 2, 2},
		{"not proficient, -1 ability", NotProficient, -1, 3, -1},
		{"proficient, +3 ability, +2 prof", Proficient, 3, 2, 5},
		{"proficient, +0 ability, +4 prof", Proficient, 0, 4, 4},
		{"expertise, +2 ability, +2 prof", Expertise, 2, 2, 6},
		{"expertise, +3 ability, +3 prof", Expertise, 3, 3, 9},
		{"expertise, -1 ability, +2 prof", Expertise, -1, 2, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skill := &Skill{Proficiency: tt.proficiency}
			got := CalculateSkillModifier(skill, tt.abilityMod, tt.proficiencyBonus)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestCalculateSkillModifierNilSkill(t *testing.T) {
	// Nil skill should just return ability modifier
	got := CalculateSkillModifier(nil, 3, 2)
	assert.Equal(t, 3, got)
}
