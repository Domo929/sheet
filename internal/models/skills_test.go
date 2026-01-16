package models

import "testing"

func TestAllSkillsCount(t *testing.T) {
	skills := AllSkills()
	if len(skills) != 18 {
		t.Errorf("AllSkills() returned %d skills, want 18", len(skills))
	}
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
			if got := GetSkillAbility(tt.skill); got != tt.ability {
				t.Errorf("GetSkillAbility(%s) = %s, want %s", tt.skill, got, tt.ability)
			}
		})
	}
}

func TestSkillsGet(t *testing.T) {
	skills := NewSkills()
	skills.Stealth.Proficiency = Expertise

	for _, skillName := range AllSkills() {
		skill := skills.Get(skillName)
		if skill == nil {
			t.Errorf("Skills.Get(%s) returned nil", skillName)
		}
	}

	// Check specific skill we modified
	stealth := skills.Get(SkillStealth)
	if stealth.Proficiency != Expertise {
		t.Errorf("Stealth proficiency = %d, want %d", stealth.Proficiency, Expertise)
	}
}

func TestSkillsGetInvalid(t *testing.T) {
	skills := NewSkills()
	if got := skills.Get(SkillName("invalid")); got != nil {
		t.Errorf("Skills.Get(invalid) = %v, want nil", got)
	}
}

func TestSkillsSetProficiency(t *testing.T) {
	skills := NewSkills()

	skills.SetProficiency(SkillPerception, Proficient)
	if skills.Perception.Proficiency != Proficient {
		t.Errorf("Perception proficiency = %d, want %d", skills.Perception.Proficiency, Proficient)
	}

	skills.SetProficiency(SkillStealth, Expertise)
	if skills.Stealth.Proficiency != Expertise {
		t.Errorf("Stealth proficiency = %d, want %d", skills.Stealth.Proficiency, Expertise)
	}
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
			if got != tt.expected {
				t.Errorf("CalculateSkillModifier() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestCalculateSkillModifierNilSkill(t *testing.T) {
	// Nil skill should just return ability modifier
	got := CalculateSkillModifier(nil, 3, 2)
	if got != 3 {
		t.Errorf("CalculateSkillModifier(nil, 3, 2) = %d, want 3", got)
	}
}
