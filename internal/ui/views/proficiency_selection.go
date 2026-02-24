package views

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/ui/components"
)

// ProficiencySection represents which proficiency type is being selected.
type ProficiencySection int

const (
	ProfSectionSkills ProficiencySection = iota
	ProfSectionTools
	ProfSectionLanguages
)

// Proficiency Selection Manager handles the proficiency selection step in character creation.
type ProficiencySelectionManager struct {
	// Current section being edited
	currentSection ProficiencySection

	// Skill proficiency selection
	skillSelector     components.ProficiencySelector
	skillOptions      []string
	skillsRequired    int
	classSkillOptions []string // Skills available from class
	backgroundSkills  []string // Fixed skills from background
	backgroundName    string   // Name of selected background

	// Tool proficiency selection
	toolSelector   components.ProficiencySelector
	toolOptions    []string
	toolsRequired  int
	backgroundTool string // Fixed tool from background

	// Language selection
	languageSelector  components.ProficiencySelector
	languageOptions   []string
	languagesRequired int
	racialLanguages   []string // Fixed languages from race

	// Completion tracking
	skillsComplete    bool
	toolsComplete     bool
	languagesComplete bool
}

// NewProficiencySelectionManager creates a new proficiency selection manager.
func NewProficiencySelectionManager(
	selectedClass *data.Class,
	selectedBackground *data.Background,
	selectedRace *data.Race,
) ProficiencySelectionManager {
	psm := ProficiencySelectionManager{
		currentSection: ProfSectionSkills,
	}

	// Initialize skill selection
	if selectedClass != nil {
		psm.classSkillOptions = selectedClass.SkillChoices.Options
		psm.skillsRequired = selectedClass.SkillChoices.Count
		psm.skillOptions = normalizeSkillNames(psm.classSkillOptions)
	}

	if selectedBackground != nil {
		psm.backgroundSkills = normalizeSkillNames(selectedBackground.SkillProficiencies)
		psm.backgroundName = selectedBackground.Name
	}

	if psm.skillsRequired > 0 {
		psm.skillSelector = components.NewProficiencySelector(
			"Choose Class Skill Proficiencies",
			psm.skillOptions,
			psm.skillsRequired,
		)
		// Set background skills as locked (pre-selected, non-toggleable)
		if len(psm.backgroundSkills) > 0 {
			label := "(From " + psm.backgroundName + ")"
			psm.skillSelector.SetLocked(psm.backgroundSkills, label)
		}
		psm.skillSelector.SetFocused(true)
	} else {
		psm.skillsComplete = true
	}

	// Initialize tool selection
	// For now, most backgrounds provide a specific tool, not a choice
	// We'll implement tool choices if needed
	if selectedBackground != nil {
		psm.backgroundTool = selectedBackground.ToolProficiency
	}
	psm.toolsComplete = true // No choices needed for Phase 4B

	// Initialize language selection
	// Collect racial languages
	if selectedRace != nil {
		psm.racialLanguages = selectedRace.Languages
	}

	// For Phase 4B, we'll assume no additional language choices needed
	// unless the race or background specifically grants them
	psm.languagesComplete = true

	return psm
}

// Update handles input for the proficiency selection manager.
func (psm *ProficiencySelectionManager) Update(msg interface{}) {
	// Handle window size messages for all selectors
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		if psm.skillsRequired > 0 {
			psm.skillSelector, _ = psm.skillSelector.Update(msg)
		}
		if psm.toolsRequired > 0 {
			psm.toolSelector, _ = psm.toolSelector.Update(msg)
		}
		if psm.languagesRequired > 0 {
			psm.languageSelector, _ = psm.languageSelector.Update(msg)
		}
		return
	}

	// Handle other messages only for the current section
	switch psm.currentSection {
	case ProfSectionSkills:
		if psm.skillsRequired > 0 {
			psm.skillSelector, _ = psm.skillSelector.Update(msg)
			psm.skillsComplete = psm.skillSelector.IsComplete()
		}
	case ProfSectionTools:
		if psm.toolsRequired > 0 {
			psm.toolSelector, _ = psm.toolSelector.Update(msg)
			psm.toolsComplete = psm.toolSelector.IsComplete()
		}
	case ProfSectionLanguages:
		if psm.languagesRequired > 0 {
			psm.languageSelector, _ = psm.languageSelector.Update(msg)
			psm.languagesComplete = psm.languageSelector.IsComplete()
		}
	}
}

// View renders the current proficiency selection.
func (psm ProficiencySelectionManager) View() string {
	var b strings.Builder

	switch psm.currentSection {
	case ProfSectionSkills:
		if psm.skillsRequired > 0 {
			// Render selector without help text
			b.WriteString(psm.skillSelector.ViewWithoutHelp())
			
			// Help text at the end
			helpText := psm.skillSelector.HelpText()
			if helpText != "" {
				b.WriteString("\n")
				b.WriteString(helpText)
			}
		} else {
			b.WriteString("No skill choices needed from class.\n")
		}

	case ProfSectionTools:
		if psm.toolsRequired > 0 {
			b.WriteString(psm.toolSelector.View())
		} else {
			b.WriteString("Tool proficiency from background:\n")
			b.WriteString("  • " + psm.backgroundTool + "\n")
		}

	case ProfSectionLanguages:
		if psm.languagesRequired > 0 {
			b.WriteString(psm.languageSelector.View())
		} else {
			b.WriteString("Languages from race:\n")
			for _, lang := range psm.racialLanguages {
				b.WriteString("  • " + lang + "\n")
			}
		}
	}

	return b.String()
}

// IsComplete returns whether all proficiency selections are complete.
func (psm ProficiencySelectionManager) IsComplete() bool {
	return psm.skillsComplete && psm.toolsComplete && psm.languagesComplete
}

// NextSection moves to the next proficiency section.
func (psm *ProficiencySelectionManager) NextSection() bool {
	// Unfocus current section
	switch psm.currentSection {
	case ProfSectionSkills:
		psm.skillSelector.SetFocused(false)
		if !psm.toolsComplete {
			psm.currentSection = ProfSectionTools
			psm.toolSelector.SetFocused(true)
			return true
		}
		if !psm.languagesComplete {
			psm.currentSection = ProfSectionLanguages
			psm.languageSelector.SetFocused(true)
			return true
		}
		return false

	case ProfSectionTools:
		psm.toolSelector.SetFocused(false)
		if !psm.languagesComplete {
			psm.currentSection = ProfSectionLanguages
			psm.languageSelector.SetFocused(true)
			return true
		}
		return false

	case ProfSectionLanguages:
		psm.languageSelector.SetFocused(false)
		return false
	}

	return false
}

// PreviousSection moves to the previous proficiency section.
func (psm *ProficiencySelectionManager) PreviousSection() bool {
	// Unfocus current section
	switch psm.currentSection {
	case ProfSectionSkills:
		return false

	case ProfSectionTools:
		psm.toolSelector.SetFocused(false)
		psm.currentSection = ProfSectionSkills
		psm.skillSelector.SetFocused(true)
		return true

	case ProfSectionLanguages:
		psm.languageSelector.SetFocused(false)
		if psm.toolsRequired > 0 {
			psm.currentSection = ProfSectionTools
			psm.toolSelector.SetFocused(true)
		} else {
			psm.currentSection = ProfSectionSkills
			psm.skillSelector.SetFocused(true)
		}
		return true
	}

	return false
}

// GetSelectedSkills returns the skill names selected by the user.
func (psm ProficiencySelectionManager) GetSelectedSkills() []string {
	if psm.skillsRequired == 0 {
		return []string{}
	}
	return psm.skillSelector.GetSelectedOptions()
}

// GetSelectedTools returns the tool names selected by the user.
func (psm ProficiencySelectionManager) GetSelectedTools() []string {
	if psm.toolsRequired == 0 {
		return []string{}
	}
	return psm.toolSelector.GetSelectedOptions()
}

// GetSelectedLanguages returns the language names selected by the user.
func (psm ProficiencySelectionManager) GetSelectedLanguages() []string {
	if psm.languagesRequired == 0 {
		return []string{}
	}
	return psm.languageSelector.GetSelectedOptions()
}

// GetAllSkills returns all skill proficiencies (class-selected + background-granted).
func (psm ProficiencySelectionManager) GetAllSkills() []string {
	all := make([]string, 0)
	all = append(all, psm.GetSelectedSkills()...)
	all = append(all, psm.backgroundSkills...)
	return all
}

// GetAllTools returns all tool proficiencies (selected + background-granted).
func (psm ProficiencySelectionManager) GetAllTools() []string {
	all := make([]string, 0)
	all = append(all, psm.GetSelectedTools()...)
	if psm.backgroundTool != "" {
		all = append(all, psm.backgroundTool)
	}
	return all
}

// GetAllLanguages returns all language proficiencies (selected + racial).
func (psm ProficiencySelectionManager) GetAllLanguages() []string {
	all := make([]string, 0)
	all = append(all, psm.GetSelectedLanguages()...)
	all = append(all, psm.racialLanguages...)
	return all
}

// ApplyToCharacter applies the selected proficiencies to the character.
func (psm ProficiencySelectionManager) ApplyToCharacter(char *models.Character) {
	// Apply class skill selections
	selectedSkills := psm.GetSelectedSkills()
	for _, skillName := range selectedSkills {
		skillKey := skillNameToKey(skillName)
		if skill := char.Skills.Get(skillKey); skill != nil {
			skill.Proficiency = models.Proficient
		}
	}

	// Apply background skills
	for _, skillName := range psm.backgroundSkills {
		skillKey := skillNameToKey(skillName)
		if skill := char.Skills.Get(skillKey); skill != nil {
			skill.Proficiency = models.Proficient
		}
	}

	// Apply tool proficiencies
	if psm.backgroundTool != "" {
		char.Proficiencies.AddTool(psm.backgroundTool)
	}
	for _, tool := range psm.GetSelectedTools() {
		char.Proficiencies.AddTool(tool)
	}

	// Apply languages
	for _, lang := range psm.racialLanguages {
		char.Proficiencies.AddLanguage(lang)
	}
	for _, lang := range psm.GetSelectedLanguages() {
		char.Proficiencies.AddLanguage(lang)
	}
}

// Helper functions

// normalizeSkillNames converts skill names from data format to display format.
func normalizeSkillNames(skills []string) []string {
	normalized := make([]string, len(skills))
	for i, skill := range skills {
		normalized[i] = skill
	}
	return normalized
}

// skillNameToKey converts a skill display name to a SkillName constant.
func skillNameToKey(name string) models.SkillName {
	// Normalize the name by removing spaces and converting to camelCase
	normalized := strings.ReplaceAll(name, " ", "")
	if normalized == "" {
		return models.SkillAcrobatics // default fallback for empty/whitespace-only input
	}
	normalized = strings.ToLower(normalized[:1]) + normalized[1:]

	switch normalized {
	case "acrobatics":
		return models.SkillAcrobatics
	case "animalHandling":
		return models.SkillAnimalHandling
	case "arcana":
		return models.SkillArcana
	case "athletics":
		return models.SkillAthletics
	case "deception":
		return models.SkillDeception
	case "history":
		return models.SkillHistory
	case "insight":
		return models.SkillInsight
	case "intimidation":
		return models.SkillIntimidation
	case "investigation":
		return models.SkillInvestigation
	case "medicine":
		return models.SkillMedicine
	case "nature":
		return models.SkillNature
	case "perception":
		return models.SkillPerception
	case "performance":
		return models.SkillPerformance
	case "persuasion":
		return models.SkillPersuasion
	case "religion":
		return models.SkillReligion
	case "sleightOfHand":
		return models.SkillSleightOfHand
	case "stealth":
		return models.SkillStealth
	case "survival":
		return models.SkillSurvival
	default:
		return models.SkillAcrobatics // Default fallback
	}
}
