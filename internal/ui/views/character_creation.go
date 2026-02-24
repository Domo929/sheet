package views

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/Domo929/sheet/internal/ui/components"
)

// CreationStep represents the current step in character creation.
type CreationStep int

const (
	StepBasicInfo CreationStep = iota
	StepRace
	StepClass
	StepBackground
	StepAbilities
	StepProficiencies
	StepEquipment
	StepPersonality
	StepReview
)

// CharacterCreationModel manages the character creation wizard.
type CharacterCreationModel struct {
	storage  *storage.CharacterStorage
	loader   *data.Loader
	width    int
	height   int
	
	// Current step
	currentStep CreationStep
	
	// Character being created
	character *models.Character
	
	// Basic Info step
	nameInput       components.TextInput
	playerNameInput components.TextInput
	progressionType models.ProgressionType
	progressionList components.ButtonGroup
	focusedField    int // 0=name, 1=playerName, 2=progression
	
	// Race step
	raceList       components.List
	selectedRace   *data.Race
	selectedSubtype int // -1 if none selected
	
	// Class step
	classList       components.List
	selectedClass   *data.Class
	skillSelections []int // Indices of selected skills
	
	// Background step
	backgroundList       components.List
	selectedBackground   *data.Background
	
	// Ability Score step
	abilityScoreMode         AbilityScoreMode // Manual, Standard Array, or Point Buy
	abilityScores            [6]int           // STR, DEX, CON, INT, WIS, CHA
	focusedAbility           AbilityIndex     // Which ability is currently focused (0-5)
	standardArrayValues      []int            // For standard array mode
	standardArrayUsed        [6]bool          // Track which standard array values are used
	focusedSection           CreationSection  // Background bonus or ability scores
	
	// Background bonus allocation (integrated into ability score step)
	backgroundBonusPattern   BonusPattern  // +2/+1 or +1/+1/+1
	backgroundBonus2Target   int           // Which ability gets +2 (index in background options)
	backgroundBonus1Target   int           // Which ability gets +1 (index in background options)
	focusedBonusField        BonusField    // pattern, +2 target, or +1 target
	backgroundBonusComplete  bool          // Whether background bonuses are fully allocated
	
	// Proficiency Selection step
	proficiencyManager ProficiencySelectionManager
	
	// Equipment step
	startingGold             int               // Starting gold pieces for the character
	equipmentChoices         []int             // Selected option index for each equipment choice (-1 = not selected)
	focusedEquipmentChoice   int               // Which equipment choice is currently focused (top-level index)
	focusedEquipmentOption   int               // Which option within the focused choice is highlighted
	equipmentConfirmed       bool              // Whether equipment has been reviewed and confirmed
	equipmentSubSelections   map[int]string    // Map of choice index to selected sub-item name for "any [category]" patterns
	
	// Sub-selection for "any [category]" items
	inEquipmentSubSelection  bool     // Whether we're in the sub-selection UI
	equipmentSubItems        []string // Available items for sub-selection
	focusedSubItem           int      // Which sub-item is focused

	// Personality step
	personalityTraits     []string
	personalityIdeals     []string
	personalityBonds      []string
	personalityFlaws      []string
	personalityBackstory  string
	personalityFocusField int // 0=traits, 1=ideals, 2=bonds, 3=flaws, 4=backstory
	personalityItemCursor int // cursor within the focused field's entries
	personalityEditing    bool
	personalityEditBuffer string

	// Navigation
	helpFooter components.HelpFooter
	buttons    components.ButtonGroup
	err        error
	quitting   bool
}

// AbilityScoreMode represents the method of ability score assignment.
type AbilityScoreMode int

const (
	AbilityModeManual AbilityScoreMode = iota
	AbilityModeStandardArray
	AbilityModePointBuy
)

// CreationSection represents which section is focused in the ability screen.
type CreationSection int

const (
	SectionBackgroundBonus CreationSection = iota
	SectionAbilityScores
)

// BonusPattern represents the background ability bonus allocation pattern.
type BonusPattern int

const (
	BonusPattern2Plus1     BonusPattern = iota // +2 to one ability, +1 to another
	BonusPattern1Plus1Plus1                    // +1 to three abilities
)

// BonusField represents which bonus field is focused.
type BonusField int

const (
	BonusFieldPattern BonusField = iota
	BonusFieldPlus2Target
	BonusFieldPlus1Target
)

// AbilityIndex represents an ability score index (0-5 for STR, DEX, CON, INT, WIS, CHA).
type AbilityIndex int

const (
	AbilityIndexSTR AbilityIndex = iota
	AbilityIndexDEX
	AbilityIndexCON
	AbilityIndexINT
	AbilityIndexWIS
	AbilityIndexCHA
)

// NewCharacterCreationModel creates a new character creation model.
func NewCharacterCreationModel(store *storage.CharacterStorage, loader *data.Loader) *CharacterCreationModel {
	nameInput := components.NewTextInput("Character Name", "Enter character name...")
	// Width will be set when window size is received
	
	playerNameInput := components.NewTextInput("Player Name", "Enter player name...")
	// Width will be set when window size is received
	
	progressionButtons := components.NewButtonGroup("XP Tracking", "Milestone")
	
	helpFooter := components.NewHelpFooter()
	
	navButtons := components.NewButtonGroup("Next", "Cancel")
	
	return &CharacterCreationModel{
		storage:             store,
		loader:              loader,
		currentStep:         StepBasicInfo,
		nameInput:           nameInput,
		playerNameInput:     playerNameInput,
		progressionList:     progressionButtons,
		progressionType:     models.ProgressionXP,
		focusedField:        0,
		helpFooter:          helpFooter,
		buttons:             navButtons,
		selectedSubtype:     -1,
		abilityScoreMode:    AbilityModePointBuy, // Default to point buy
		abilityScores:       [6]int{8, 8, 8, 8, 8, 8}, // Start at minimum for point buy
		standardArrayValues: models.StandardArray(),
		personalityTraits:   []string{""},
		personalityIdeals:   []string{""},
		personalityBonds:    []string{""},
		personalityFlaws:    []string{""},
		character: &models.Character{
			Info: models.CharacterInfo{
				Level: 1,
			},
		},
	}
}

// CharacterCreatedMsg is sent when character creation is complete.
type CharacterCreatedMsg struct {
	Character *models.Character
	Path      string
}

// CancelCharacterCreationMsg is sent when user cancels character creation.
type CancelCharacterCreationMsg struct{}

// Init initializes the character creation model.
func (m *CharacterCreationModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the character creation screen.
func (m *CharacterCreationModel) Update(msg tea.Msg) (*CharacterCreationModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Scale text input widths to terminal width
		// Use 80% of width or 60 chars max, whichever is smaller
		inputWidth := msg.Width * 4 / 5
		if inputWidth > 60 {
			inputWidth = 60
		}
		if inputWidth < 20 {
			inputWidth = 20
		}
		m.nameInput.Width = inputWidth
		m.playerNameInput.Width = inputWidth
		return m, nil
		
	case RaceDataLoadedMsg:
		m.raceList = components.NewList("Available Races", msg.Items)
		return m, nil
		
	case ClassDataLoadedMsg:
		m.classList = components.NewList("Available Classes", msg.Items)
		return m, nil
		
	case BackgroundDataLoadedMsg:
		m.backgroundList = components.NewList("Available Backgrounds", msg.Items)
		return m, nil
		
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	
	return m, nil
}

// handleKey processes keyboard input based on current step.
func (m *CharacterCreationModel) handleKey(msg tea.KeyPressMsg) (*CharacterCreationModel, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	}
	
	switch m.currentStep {
	case StepBasicInfo:
		return m.handleBasicInfoKeys(msg)
	case StepRace:
		return m.handleRaceKeys(msg)
	case StepClass:
		return m.handleClassKeys(msg)
	case StepBackground:
		return m.handleBackgroundKeys(msg)
	case StepAbilities:
		return m.handleAbilityKeys(msg)
	case StepProficiencies:
		return m.handleProficiencyKeys(msg)
	case StepEquipment:
		return m.handleEquipmentKeys(msg)
	case StepPersonality:
		return m.handlePersonalityKeys(msg)
	case StepReview:
		return m.handleReviewKeys(msg)
	default:
		return m, nil
	}
}

// handleBasicInfoKeys handles keys for the basic info step.
func (m *CharacterCreationModel) handleBasicInfoKeys(msg tea.KeyPressMsg) (*CharacterCreationModel, tea.Cmd) {
	switch msg.String() {
	case "tab":
		m.focusedField = (m.focusedField + 1) % 3
		return m, nil
		
	case "shift+tab":
		m.focusedField--
		if m.focusedField < 0 {
			m.focusedField = 2
		}
		return m, nil
		
	case "up":
		// Arrow keys always work for navigation
		if m.focusedField > 0 {
			m.focusedField--
		}
		return m, nil
		
	case "k":
		// Vim key only works when NOT in text field
		if m.focusedField == 2 && m.focusedField > 0 {
			m.focusedField--
			return m, nil
		}
		// Otherwise treat as regular text input
		
	case "down":
		// Arrow keys always work for navigation
		if m.focusedField < 2 {
			m.focusedField++
		}
		return m, nil
		
	case "j":
		// Vim key only works when NOT in text field
		if m.focusedField == 2 && m.focusedField < 2 {
			m.focusedField++
			return m, nil
		}
		// Otherwise treat as regular text input
		
	case "left":
		// Arrow keys for progression type selection
		if m.focusedField == 2 {
			m.progressionList.MoveLeft()
			m.updateProgressionType()
		}
		return m, nil
		
	case "h":
		// Vim key only works for progression type selection
		if m.focusedField == 2 {
			m.progressionList.MoveLeft()
			m.updateProgressionType()
			return m, nil
		}
		// Otherwise treat as regular text input
		
	case "right":
		// Arrow keys for progression type selection
		if m.focusedField == 2 {
			m.progressionList.MoveRight()
			m.updateProgressionType()
		}
		return m, nil
		
	case "l":
		// Vim key only works for progression type selection
		if m.focusedField == 2 {
			m.progressionList.MoveRight()
			m.updateProgressionType()
			return m, nil
		}
		// Otherwise treat as regular text input
		
	case "enter":
		// Move to next step if validation passes
		if m.validateBasicInfo() {
			m.applyBasicInfo()
			return m.moveToStep(StepRace)
		}
		return m, nil
		
	case "esc":
		// Cancel creation - return to character selection
		return m, func() tea.Msg {
			return CancelCharacterCreationMsg{}
		}
		
	case "backspace":
		// Handle backspace for text input fields
		if m.focusedField == 0 {
			if len(m.nameInput.Value) > 0 {
				m.nameInput.Value = m.nameInput.Value[:len(m.nameInput.Value)-1]
			}
		} else if m.focusedField == 1 {
			if len(m.playerNameInput.Value) > 0 {
				m.playerNameInput.Value = m.playerNameInput.Value[:len(m.playerNameInput.Value)-1]
			}
		}
		return m, nil
		
	case "space":
		// Handle space in text input fields
		if m.focusedField == 0 {
			m.nameInput.Value += " "
		} else if m.focusedField == 1 {
			m.playerNameInput.Value += " "
		}
		return m, nil
	}
	
	// Handle text input - only process actual character runes
	if msg.Text != "" && (m.focusedField == 0 || m.focusedField == 1) {
		char := msg.Text
		if m.focusedField == 0 {
			m.nameInput.Value += char
		} else if m.focusedField == 1 {
			m.playerNameInput.Value += char
		}
	}
	return m, nil
}

// handleRaceKeys handles keys for the race selection step.
func (m *CharacterCreationModel) handleRaceKeys(msg tea.KeyPressMsg) (*CharacterCreationModel, tea.Cmd) {
	return m.handleListSelectionKeys(msg, &m.raceList, m.selectCurrentRace, StepBasicInfo, StepClass)
}

// handleClassKeys handles keys for the class selection step.
func (m *CharacterCreationModel) handleClassKeys(msg tea.KeyPressMsg) (*CharacterCreationModel, tea.Cmd) {
	return m.handleListSelectionKeys(msg, &m.classList, m.selectCurrentClass, StepRace, StepBackground)
}

// handleBackgroundKeys handles keys for the background selection step.
func (m *CharacterCreationModel) handleBackgroundKeys(msg tea.KeyPressMsg) (*CharacterCreationModel, tea.Cmd) {
	return m.handleListSelectionKeys(msg, &m.backgroundList, func() bool {
		return m.selectCurrentBackground()
	}, StepClass, StepAbilities)
}

// handleAbilityKeys handles keys for the ability score assignment step.
func (m *CharacterCreationModel) handleAbilityKeys(msg tea.KeyPressMsg) (*CharacterCreationModel, tea.Cmd) {
	hasBackgroundBonuses := m.selectedBackground != nil && len(m.selectedBackground.AbilityScores.Options) > 0
	
	switch msg.String() {
	case "up", "k":
		if hasBackgroundBonuses && m.focusedSection == SectionBackgroundBonus {
			// In background bonus section - navigate fields
			if m.focusedBonusField > BonusFieldPattern {
				m.focusedBonusField--
			} else {
				// At top of bonus section - can't go higher
			}
		} else if m.focusedSection == SectionAbilityScores {
			// In ability score section - navigate abilities
			if m.focusedAbility > AbilityIndexSTR {
				m.focusedAbility--
			} else if hasBackgroundBonuses {
				// At top of ability section - move to bonus section
				m.focusedSection = SectionBackgroundBonus
				// Set bonus field to last valid field
				if m.backgroundBonusPattern == BonusPattern2Plus1 {
					m.focusedBonusField = BonusFieldPlus1Target // +1 target
				} else {
					m.focusedBonusField = BonusFieldPattern // pattern only
				}
			}
		}
		return m, nil
		
	case "down", "j":
		if hasBackgroundBonuses && m.focusedSection == SectionBackgroundBonus {
			// In background bonus section
			maxField := BonusFieldPattern
			if m.backgroundBonusPattern == BonusPattern2Plus1 {
				maxField = BonusFieldPlus1Target // pattern, +2, +1
			}
			
			if m.focusedBonusField < maxField {
				m.focusedBonusField++
			} else {
				// At bottom of bonus section - move to ability section
				m.focusedSection = SectionAbilityScores
				m.focusedAbility = AbilityIndexSTR
			}
		} else if m.focusedSection == SectionAbilityScores {
			// In ability score section
			if m.focusedAbility < AbilityIndexCHA {
				m.focusedAbility++
			}
		}
		return m, nil
		
	case "right", "l", "+", "=":
		if hasBackgroundBonuses && m.focusedSection == SectionBackgroundBonus {
			// Adjusting background bonuses
			return m.adjustBackgroundBonus(true), nil
		} else if m.focusedSection == SectionAbilityScores {
			// Adjusting ability scores
			return m.incrementAbility(), nil
		}
		return m, nil
		
	case "left", "h", "-", "_":
		if hasBackgroundBonuses && m.focusedSection == SectionBackgroundBonus {
			// Adjusting background bonuses
			return m.adjustBackgroundBonus(false), nil
		} else if m.focusedSection == SectionAbilityScores {
			// Adjusting ability scores
			return m.decrementAbility(), nil
		}
		return m, nil
		
	case "m":
		// Toggle mode (only works in ability score section)
		if m.focusedSection == SectionAbilityScores {
			switch m.abilityScoreMode {
			case AbilityModeManual:
				m.abilityScoreMode = AbilityModeStandardArray
				m.resetAbilityScores()
			case AbilityModeStandardArray:
				m.abilityScoreMode = AbilityModePointBuy
				m.resetAbilityScores()
			case AbilityModePointBuy:
				m.abilityScoreMode = AbilityModeManual
				m.resetAbilityScores()
			}
		}
		return m, nil
		
	case "enter":
		if hasBackgroundBonuses && m.focusedSection == SectionBackgroundBonus && !m.backgroundBonusComplete {
			// Confirm background bonuses
			if m.backgroundBonusPattern == BonusPattern1Plus1Plus1 {
				// +1/+1/+1 - auto-allocate
				m.backgroundBonusComplete = true
				m.focusedSection = SectionAbilityScores // Move to ability scores
				m.focusedAbility = 0
			} else {
				// +2/+1 - need both targets selected and different
				if m.backgroundBonus2Target >= 0 && m.backgroundBonus1Target >= 0 {
					if m.backgroundBonus2Target != m.backgroundBonus1Target {
						m.backgroundBonusComplete = true
						m.focusedSection = SectionAbilityScores // Move to ability scores
						m.focusedAbility = 0
					} else {
						m.err = fmt.Errorf("+2 and +1 must be assigned to different abilities")
					}
				} else {
					m.err = fmt.Errorf("must select abilities for both +2 and +1 bonuses")
				}
			}
		} else if m.focusedSection == SectionAbilityScores {
			// Move to proficiency selection
			if m.validateAbilityScores() {
				return m.moveToStep(StepProficiencies)
			}
		}
		return m, nil
		
	case "esc":
		// Go back to background selection
		return m.moveToStep(StepBackground)
	}
	
	return m, nil
}

// handleProficiencyKeys handles keys for the proficiency selection step.
func (m *CharacterCreationModel) handleProficiencyKeys(msg tea.KeyPressMsg) (*CharacterCreationModel, tea.Cmd) {
switch msg.String() {
case "enter":
// If proficiencies are complete, continue to equipment
if m.proficiencyManager.IsComplete() {
return m.moveToStep(StepEquipment)
}
// Otherwise, do nothing (user needs to complete selections)
return m, nil

case "tab":
// Try to move to next section
m.proficiencyManager.NextSection()
return m, nil

case "shift+tab":
// Try to move to previous section
m.proficiencyManager.PreviousSection()
return m, nil

case "esc":
// Go back to abilities
return m.moveToStep(StepAbilities)

default:
// Pass keys (including space for toggle) to the proficiency manager
m.proficiencyManager.Update(msg)
return m, nil
}
}

// renderProficiencySelection renders the proficiency selection step.
func (m *CharacterCreationModel) renderProficiencySelection() string {
	var content strings.Builder
	
	stepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)
	content.WriteString(stepStyle.Render("Step 6 of 9: Proficiency Selection"))
	content.WriteString("\n\n")
	
	content.WriteString(m.proficiencyManager.View())
	
	return content.String()
}

// handleEquipmentKeys handles keys for the starting equipment step.
func (m *CharacterCreationModel) handleEquipmentKeys(msg tea.KeyPressMsg) (*CharacterCreationModel, tea.Cmd) {
	if m.selectedClass == nil || len(m.selectedClass.StartingEquipment) == 0 {
		// No equipment to select
		switch msg.String() {
		case "enter":
			m.equipmentConfirmed = true
			m.startingGold = m.getStartingGold()
			return m.moveToStep(StepPersonality)
		case "esc":
			return m.moveToStep(StepProficiencies)
		}
		return m, nil
	}
	
	// Handle sub-selection mode (picking specific item from "any [category]")
	if m.inEquipmentSubSelection {
		switch msg.String() {
		case "up", "k":
			if m.focusedSubItem > 0 {
				m.focusedSubItem--
			}
			return m, nil
		case "down", "j":
			if m.focusedSubItem < len(m.equipmentSubItems)-1 {
				m.focusedSubItem++
			}
			return m, nil
		case "space", "enter":
			// Select this specific item (space or enter)
			if m.equipmentSubSelections == nil {
				m.equipmentSubSelections = make(map[int]string)
			}
			m.equipmentSubSelections[m.focusedEquipmentChoice] = m.equipmentSubItems[m.focusedSubItem]
			m.inEquipmentSubSelection = false
			m.equipmentSubItems = nil
			m.focusedSubItem = 0
			return m, nil
		case "esc":
			// Cancel sub-selection
			m.inEquipmentSubSelection = false
			m.equipmentSubItems = nil
			m.focusedSubItem = 0
			return m, nil
		}
		return m, nil
	}
	
	switch msg.String() {
	case "up", "k":
		// Move up through equipment items/options
		m.navigateEquipmentUp()
		return m, nil
		
	case "down", "j":
		// Move down through equipment items/options
		m.navigateEquipmentDown()
		return m, nil
		
	case "space":
		// Spacebar: Select current option and enter sub-selection if needed
		currentEquip := m.selectedClass.StartingEquipment[m.focusedEquipmentChoice]
		if currentEquip.Type == data.EquipmentChoiceSelect {
			// Select the focused option
			m.equipmentChoices[m.focusedEquipmentChoice] = m.focusedEquipmentOption
			
			// Check if this selection needs sub-selection
			if m.needsSubSelection() {
				m.enterSubSelection()
				return m, nil
			}
		}
		return m, nil
		
	case "enter":
		// Enter: Always proceed to personality if all choices made (never enters sub-selection)
		if m.allEquipmentChoicesMade() {
			m.equipmentConfirmed = true
			m.startingGold = m.getStartingGold()
			return m.moveToStep(StepPersonality)
		}
		return m, nil
		
	case "esc":
		// Go back to proficiencies
		return m.moveToStep(StepProficiencies)
	}
	
	return m, nil
}

// navigateEquipmentUp moves focus up through equipment choices and options.
func (m *CharacterCreationModel) navigateEquipmentUp() {
	if m.selectedClass == nil || len(m.selectedClass.StartingEquipment) == 0 {
		return
	}
	
	currentEquip := m.selectedClass.StartingEquipment[m.focusedEquipmentChoice]
	
	if currentEquip.Type == data.EquipmentChoiceSelect && m.focusedEquipmentOption > 0 {
		// Move up within current choice's options
		m.focusedEquipmentOption--
	} else if m.focusedEquipmentChoice > 0 {
		// Move to previous choice
		m.focusedEquipmentChoice--

		// Set focus to last option of previous choice (or 0 for fixed items)
		prevEquip := m.selectedClass.StartingEquipment[m.focusedEquipmentChoice]
		if prevEquip.Type == data.EquipmentChoiceSelect {
			m.focusedEquipmentOption = len(prevEquip.Options) - 1
		} else {
			m.focusedEquipmentOption = 0
		}
	}
}

// navigateEquipmentDown moves focus down through equipment choices and options.
func (m *CharacterCreationModel) navigateEquipmentDown() {
	if m.selectedClass == nil || len(m.selectedClass.StartingEquipment) == 0 {
		return
	}
	
	currentEquip := m.selectedClass.StartingEquipment[m.focusedEquipmentChoice]
	
	if currentEquip.Type == data.EquipmentChoiceSelect && m.focusedEquipmentOption < len(currentEquip.Options)-1 {
		// Move down within current choice's options
		m.focusedEquipmentOption++
	} else if m.focusedEquipmentChoice < len(m.selectedClass.StartingEquipment)-1 {
		// Move to next choice
		m.focusedEquipmentChoice++
		m.focusedEquipmentOption = 0
	}
}

// allEquipmentChoicesMade checks if all equipment choices have been selected.
func (m *CharacterCreationModel) allEquipmentChoicesMade() bool {
	if m.selectedClass == nil {
		return true
	}
	
	for i, equip := range m.selectedClass.StartingEquipment {
		if equip.Type == data.EquipmentChoiceSelect {
			// Check if a selection has been made
			if i >= len(m.equipmentChoices) || m.equipmentChoices[i] < 0 {
				return false
			}
			
			// Check if selected option needs sub-selection
			selectedIdx := m.equipmentChoices[i]
			if selectedIdx >= 0 && selectedIdx < len(equip.Options) {
				for _, item := range equip.Options[selectedIdx].Items {
					if m.needsItemSubSelection(item) {
						// Check if we have a sub-selection for this choice
						if m.equipmentSubSelections == nil || m.equipmentSubSelections[i] == "" {
							return false // Needs sub-selection but none made
						}
					}
				}
			}
		}
		// Fixed items don't need selection
	}
	return true
}

// needsItemSubSelection checks if an item needs a sub-selection (has a filter).
func (m *CharacterCreationModel) needsItemSubSelection(item data.EquipmentItem) bool {
	return item.Filter != nil
}

// needsSubSelection checks if the currently focused choice needs a sub-selection.
func (m *CharacterCreationModel) needsSubSelection() bool {
	if m.selectedClass == nil || m.focusedEquipmentChoice >= len(m.selectedClass.StartingEquipment) {
		return false
	}
	
	equip := m.selectedClass.StartingEquipment[m.focusedEquipmentChoice]
	if equip.Type != data.EquipmentChoiceSelect || m.focusedEquipmentChoice >= len(m.equipmentChoices) {
		return false
	}

	selectedIdx := m.equipmentChoices[m.focusedEquipmentChoice]
	if selectedIdx < 0 || selectedIdx >= len(equip.Options) {
		return false
	}

	// Check if any item in the selected option has a filter
	for _, item := range equip.Options[selectedIdx].Items {
		if m.needsItemSubSelection(item) {
			return true
		}
	}

	return false
}

// enterSubSelection enters the sub-selection mode for choosing a specific item.
func (m *CharacterCreationModel) enterSubSelection() {
	if m.selectedClass == nil || m.focusedEquipmentChoice >= len(m.selectedClass.StartingEquipment) {
		return
	}

	equip := m.selectedClass.StartingEquipment[m.focusedEquipmentChoice]
	if equip.Type != data.EquipmentChoiceSelect || m.focusedEquipmentChoice >= len(m.equipmentChoices) {
		return
	}
	
	selectedIdx := m.equipmentChoices[m.focusedEquipmentChoice]
	if selectedIdx < 0 || selectedIdx >= len(equip.Options) {
		return
	}
	
	// Find the item with a filter
	for _, item := range equip.Options[selectedIdx].Items {
		if m.needsItemSubSelection(item) {
			// Get available items based on filter
			m.equipmentSubItems = m.getItemsForFilter(item.Filter)
			if len(m.equipmentSubItems) > 0 {
				m.inEquipmentSubSelection = true
				m.focusedSubItem = 0
			}
			return
		}
	}
}

// getItemsForFilter returns a list of item names matching an equipment filter.
func (m *CharacterCreationModel) getItemsForFilter(filter *data.EquipmentFilter) []string {
	if filter == nil {
		return nil
	}
	
	equipment, err := m.loader.GetEquipment()
	if err != nil || equipment == nil {
		return nil
	}
	
	var items []string
	
	// Handle weapon filters
	if filter.WeaponType != "" {
		var weapons []data.Weapon
		
		if filter.WeaponType == data.WeaponTypeSimple {
			if filter.WeaponStyle == data.WeaponStyleMelee {
				weapons = equipment.Weapons.SimpleMelee
			} else if filter.WeaponStyle == data.WeaponStyleRanged {
				weapons = equipment.Weapons.SimpleRanged
			} else {
				weapons = equipment.Weapons.GetSimpleWeapons()
			}
		} else if filter.WeaponType == data.WeaponTypeMartial {
			if filter.WeaponStyle == data.WeaponStyleMelee {
				weapons = equipment.Weapons.MartialMelee
			} else if filter.WeaponStyle == data.WeaponStyleRanged {
				weapons = equipment.Weapons.MartialRanged
			} else {
				weapons = equipment.Weapons.GetMartialWeapons()
			}
		}
		
		for _, w := range weapons {
			items = append(items, w.Name)
		}
	}
	
	// Handle armor filters
	if filter.ArmorType != "" {
		var armorItems []data.ArmorItem
		
		switch filter.ArmorType {
		case data.ArmorCategoryLight:
			armorItems = equipment.Armor.Light
		case data.ArmorCategoryMedium:
			armorItems = equipment.Armor.Medium
		case data.ArmorCategoryHeavy:
			armorItems = equipment.Armor.Heavy
		case data.ArmorCategoryShield:
			armorItems = equipment.Armor.Shield
		}
		
		for _, a := range armorItems {
			items = append(items, a.Name)
		}
	}
	
	return items
}

// getFilterDisplayName generates a user-friendly display name from an equipment filter.
func (m *CharacterCreationModel) getFilterDisplayName(filter *data.EquipmentFilter) string {
	if filter == nil {
		return ""
	}
	
	// Build display string
	var parts []string
	
	if filter.WeaponType != "" {
		parts = append(parts, "any")
		parts = append(parts, string(filter.WeaponType))
		if filter.WeaponStyle != "" {
			parts = append(parts, string(filter.WeaponStyle))
		}
		parts = append(parts, "weapon")
	}

	if filter.ArmorType != "" {
		parts = append(parts, "any")
		parts = append(parts, string(filter.ArmorType))
		parts = append(parts, "armor")
	}
	
	return strings.Join(parts, " ")
}

// getSelectedEquipment returns all selected equipment items.
func (m *CharacterCreationModel) getSelectedEquipment() []data.EquipmentItem {
	var items []data.EquipmentItem
	
	if m.selectedClass == nil {
		return items
	}
	
	for i, equip := range m.selectedClass.StartingEquipment {
		if equip.Type == data.EquipmentChoiceFixed && equip.Item != nil {
			// Fixed item - add directly
			items = append(items, *equip.Item)
		} else if equip.Type == data.EquipmentChoiceSelect && i < len(m.equipmentChoices) {
			selectedIdx := m.equipmentChoices[i]
			if selectedIdx >= 0 && selectedIdx < len(equip.Options) {
				// Add all items from the selected option
				for _, item := range equip.Options[selectedIdx].Items {
					// If this item has a filter, use the sub-selection
					if m.needsItemSubSelection(item) {
						if m.equipmentSubSelections != nil && m.equipmentSubSelections[i] != "" {
							items = append(items, data.EquipmentItem{
								Name:     m.equipmentSubSelections[i],
								Quantity: item.Quantity,
								Category: item.Category,
							})
						}
						// If filter but no selection yet, skip (shouldn't happen if validation works)
					} else {
						// Normal item, add as-is
						items = append(items, item)
					}
				}
			}
		}
	}
	
	return items
}

// handleReviewKeys handles keys for the review step.
func (m *CharacterCreationModel) handleReviewKeys(msg tea.KeyPressMsg) (*CharacterCreationModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Save character
		return m.finalizeCharacter()
		
	case "esc":
		// Go back to personality
		return m.moveToStep(StepPersonality)
	}
	
	return m, nil
}

// handlePersonalityKeys handles keys for the personality step.
func (m *CharacterCreationModel) handlePersonalityKeys(msg tea.KeyPressMsg) (*CharacterCreationModel, tea.Cmd) {
	// If editing an entry, handle text input
	if m.personalityEditing {
		switch msg.Code {
		case tea.KeyEnter:
			if m.personalityFocusField == 4 {
				// Backstory: Enter inserts newline
				m.personalityEditBuffer += "\n"
				return m, nil
			}
			// Save the edit
			m.savePersonalityEdit()
			m.personalityEditing = false
			return m, nil
		case tea.KeyEscape:
			m.personalityEditing = false
			return m, nil
		case tea.KeyBackspace:
			if len(m.personalityEditBuffer) > 0 {
				m.personalityEditBuffer = m.personalityEditBuffer[:len(m.personalityEditBuffer)-1]
			}
			return m, nil
		default:
			if msg.Text != "" {
				m.personalityEditBuffer += msg.Text
			}
			return m, nil
		}
	}

	switch msg.String() {
	case "tab":
		if m.personalityFocusField == 4 {
			// Tab from last field advances to Review
			return m.moveToStep(StepReview)
		}
		m.personalityFocusField = (m.personalityFocusField + 1) % 5
		m.personalityItemCursor = 0
		return m, nil
	case "shift+tab":
		m.personalityFocusField = (m.personalityFocusField + 4) % 5
		m.personalityItemCursor = 0
		return m, nil
	case "up", "k":
		if m.personalityItemCursor > 0 {
			m.personalityItemCursor--
		}
		return m, nil
	case "down", "j":
		items := m.getPersonalityFieldItems()
		if m.personalityItemCursor < len(items) {
			m.personalityItemCursor++
		}
		return m, nil
	case "enter":
		items := m.getPersonalityFieldItems()
		if m.personalityFocusField < 4 && m.personalityItemCursor == len(items) {
			// On "+ Add another"
			m.addPersonalityEntry()
			return m, nil
		}
		if m.personalityFocusField == 4 {
			// Backstory: enter edit mode
			m.personalityEditing = true
			m.personalityEditBuffer = m.personalityBackstory
			return m, nil
		}
		if m.personalityFocusField < 4 && m.personalityItemCursor < len(items) {
			// Edit current item
			m.personalityEditing = true
			m.personalityEditBuffer = items[m.personalityItemCursor]
			return m, nil
		}
		// Advance to review
		return m.moveToStep(StepReview)
	case "d":
		m.deletePersonalityEntry()
		return m, nil
	case "esc":
		return m.moveToStep(StepEquipment)
	}

	// If on backstory and not editing, capture runes to start editing
	if m.personalityFocusField == 4 && msg.Text != "" {
		m.personalityEditing = true
		m.personalityEditBuffer = m.personalityBackstory + msg.Text
		return m, nil
	}

	// If on a list item, capture runes to start editing
	items := m.getPersonalityFieldItems()
	if m.personalityFocusField < 4 && m.personalityItemCursor < len(items) && msg.Text != "" {
		m.personalityEditing = true
		m.personalityEditBuffer = items[m.personalityItemCursor] + msg.Text
		return m, nil
	}

	return m, nil
}

func (m *CharacterCreationModel) getPersonalityFieldItems() []string {
	switch m.personalityFocusField {
	case 0:
		return m.personalityTraits
	case 1:
		return m.personalityIdeals
	case 2:
		return m.personalityBonds
	case 3:
		return m.personalityFlaws
	default:
		return nil
	}
}

func (m *CharacterCreationModel) addPersonalityEntry() {
	switch m.personalityFocusField {
	case 0:
		m.personalityTraits = append(m.personalityTraits, "")
		m.personalityItemCursor = len(m.personalityTraits) - 1
	case 1:
		m.personalityIdeals = append(m.personalityIdeals, "")
		m.personalityItemCursor = len(m.personalityIdeals) - 1
	case 2:
		m.personalityBonds = append(m.personalityBonds, "")
		m.personalityItemCursor = len(m.personalityBonds) - 1
	case 3:
		m.personalityFlaws = append(m.personalityFlaws, "")
		m.personalityItemCursor = len(m.personalityFlaws) - 1
	}
	m.personalityEditing = true
	m.personalityEditBuffer = ""
}

func (m *CharacterCreationModel) deletePersonalityEntry() {
	items := m.getPersonalityFieldItems()
	if len(items) <= 1 || m.personalityItemCursor >= len(items) {
		return
	}
	switch m.personalityFocusField {
	case 0:
		m.personalityTraits = append(m.personalityTraits[:m.personalityItemCursor], m.personalityTraits[m.personalityItemCursor+1:]...)
	case 1:
		m.personalityIdeals = append(m.personalityIdeals[:m.personalityItemCursor], m.personalityIdeals[m.personalityItemCursor+1:]...)
	case 2:
		m.personalityBonds = append(m.personalityBonds[:m.personalityItemCursor], m.personalityBonds[m.personalityItemCursor+1:]...)
	case 3:
		m.personalityFlaws = append(m.personalityFlaws[:m.personalityItemCursor], m.personalityFlaws[m.personalityItemCursor+1:]...)
	}
	if m.personalityItemCursor >= len(m.getPersonalityFieldItems()) {
		m.personalityItemCursor = len(m.getPersonalityFieldItems()) - 1
	}
}

func (m *CharacterCreationModel) savePersonalityEdit() {
	text := m.personalityEditBuffer
	switch m.personalityFocusField {
	case 0:
		m.personalityTraits[m.personalityItemCursor] = text
	case 1:
		m.personalityIdeals[m.personalityItemCursor] = text
	case 2:
		m.personalityBonds[m.personalityItemCursor] = text
	case 3:
		m.personalityFlaws[m.personalityItemCursor] = text
	case 4:
		m.personalityBackstory = text
	}
}

// adjustBackgroundBonus handles left/right adjustments in the background bonus section.
func (m *CharacterCreationModel) adjustBackgroundBonus(increase bool) *CharacterCreationModel {
	if m.selectedBackground == nil {
		return m
	}
	
	options := m.selectedBackground.AbilityScores.Options
	points := m.selectedBackground.AbilityScores.Points
	canUse2Plus1 := len(options) >= 2 && points == 3
	canUse1Plus1Plus1 := len(options) >= 3 && points == 3
	
	if m.focusedBonusField == BonusFieldPattern {
		// Toggle pattern
		if increase {
			if m.backgroundBonusPattern == BonusPattern2Plus1 && canUse1Plus1Plus1 {
				m.backgroundBonusPattern = BonusPattern1Plus1Plus1
			} else if canUse2Plus1 {
				m.backgroundBonusPattern = BonusPattern2Plus1
			}
		} else {
			if m.backgroundBonusPattern == BonusPattern1Plus1Plus1 && canUse2Plus1 {
				m.backgroundBonusPattern = BonusPattern2Plus1
			} else if canUse1Plus1Plus1 {
				m.backgroundBonusPattern = BonusPattern1Plus1Plus1
			}
		}
	} else if m.focusedBonusField == BonusFieldPlus2Target {
		// Cycle +2 target (must be different from +1 target)
		if increase {
			newTarget := (m.backgroundBonus2Target + 1) % len(options)
			// Skip if same as +1 target
			if newTarget == m.backgroundBonus1Target && len(options) > 1 {
				newTarget = (newTarget + 1) % len(options)
			}
			m.backgroundBonus2Target = newTarget
		} else {
			newTarget := m.backgroundBonus2Target - 1
			if newTarget < 0 {
				newTarget = len(options) - 1
			}
			// Skip if same as +1 target
			if newTarget == m.backgroundBonus1Target && len(options) > 1 {
				newTarget--
				if newTarget < 0 {
					newTarget = len(options) - 1
				}
			}
			m.backgroundBonus2Target = newTarget
		}
	} else if m.focusedBonusField == BonusFieldPlus1Target {
		// Cycle +1 target (must be different from +2 target)
		if increase {
			newTarget := (m.backgroundBonus1Target + 1) % len(options)
			// Skip if same as +2 target
			if newTarget == m.backgroundBonus2Target && len(options) > 1 {
				newTarget = (newTarget + 1) % len(options)
			}
			m.backgroundBonus1Target = newTarget
		} else {
			newTarget := m.backgroundBonus1Target - 1
			if newTarget < 0 {
				newTarget = len(options) - 1
			}
			// Skip if same as +2 target
			if newTarget == m.backgroundBonus2Target && len(options) > 1 {
				newTarget--
				if newTarget < 0 {
					newTarget = len(options) - 1
				}
			}
			m.backgroundBonus1Target = newTarget
		}
	}
	
	return m
}

// handleListSelectionKeys is a generic handler for list-based selection steps (race, class, background).
func (m *CharacterCreationModel) handleListSelectionKeys(
	msg tea.KeyPressMsg,
	list *components.List,
	selectFunc func() bool,
	previousStep CreationStep,
	nextStep CreationStep,
) (*CharacterCreationModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		list.MoveUp()
		return m, nil
		
	case "down", "j":
		list.MoveDown()
		return m, nil
		
	case "enter":
		// Select item and move to next step
		if selectFunc() {
			return m.moveToStep(nextStep)
		}
		return m, nil
		
	case "esc":
		// Go back to previous step
		return m.moveToStep(previousStep)
	}
	
	return m, nil
}

// validateBasicInfo checks if basic info is complete.
func (m *CharacterCreationModel) validateBasicInfo() bool {
	if strings.TrimSpace(m.nameInput.Value) == "" {
		m.err = fmt.Errorf("character name is required")
		return false
	}
	if strings.TrimSpace(m.playerNameInput.Value) == "" {
		m.err = fmt.Errorf("player name is required")
		return false
	}
	m.err = nil
	return true
}

// applyBasicInfo applies the basic info to the character.
func (m *CharacterCreationModel) applyBasicInfo() {
	m.character.Info.Name = strings.TrimSpace(m.nameInput.Value)
	m.character.Info.PlayerName = strings.TrimSpace(m.playerNameInput.Value)
	m.character.Info.ProgressionType = m.progressionType
}

// updateProgressionType updates the progression type based on button selection.
func (m *CharacterCreationModel) updateProgressionType() {
	if m.progressionList.SelectedIndex == 0 {
		m.progressionType = models.ProgressionXP
	} else {
		m.progressionType = models.ProgressionMilestone
	}
}

// moveToStep moves to a specific creation step.
func (m *CharacterCreationModel) moveToStep(step CreationStep) (*CharacterCreationModel, tea.Cmd) {
	m.currentStep = step
	
	// Load data for the new step
	switch step {
	case StepRace:
		return m, m.loadRaces()
	case StepClass:
		return m, m.loadClasses()
	case StepBackground:
		return m, m.loadBackgrounds()
	case StepProficiencies:
		// Initialize proficiency manager
		m.proficiencyManager = NewProficiencySelectionManager(
			m.selectedClass,
			m.selectedBackground,
			m.selectedRace,
		)
	case StepEquipment:
		// Initialize equipment choices only if not already initialized
		if m.selectedClass != nil && m.equipmentChoices == nil {
			m.equipmentChoices = make([]int, len(m.selectedClass.StartingEquipment))
			for i, equip := range m.selectedClass.StartingEquipment {
				if equip.Type == data.EquipmentChoiceFixed {
					// Fixed items are automatically "selected" (0 = selected)
					m.equipmentChoices[i] = 0
				} else {
					// Choices start unselected
					m.equipmentChoices[i] = -1
				}
			}
			m.focusedEquipmentChoice = 0
			m.focusedEquipmentOption = 0
		}
	case StepPersonality:
		// No data loading needed
	}
	
	return m, nil
}

// loadRaces loads race data.
func (m *CharacterCreationModel) loadRaces() tea.Cmd {
	return func() tea.Msg {
		races, err := m.loader.GetRaces()
		if err != nil {
			return fmt.Errorf("failed to load races: %w", err)
		}
		
		items := make([]components.ListItem, len(races.Races))
		for i, race := range races.Races {
			desc := fmt.Sprintf("%s, %s, Speed %d ft", race.CreatureType, race.Size, race.Speed)
			items[i] = components.ListItem{
				Title:       race.Name,
				Description: desc,
				Value:       &races.Races[i],
			}
		}
		
		return RaceDataLoadedMsg{Items: items}
	}
}

// loadClasses loads class data.
func (m *CharacterCreationModel) loadClasses() tea.Cmd {
	return func() tea.Msg {
		classes, err := m.loader.GetClasses()
		if err != nil {
			return fmt.Errorf("failed to load classes: %w", err)
		}
		
		items := make([]components.ListItem, len(classes.Classes))
		for i, class := range classes.Classes {
			desc := fmt.Sprintf("Hit Die: %s, Primary: %s", class.HitDice, strings.Join(class.PrimaryAbility, "/"))
			items[i] = components.ListItem{
				Title:       class.Name,
				Description: desc,
				Value:       &classes.Classes[i],
			}
		}
		
		return ClassDataLoadedMsg{Items: items}
	}
}

// loadBackgrounds loads background data.
func (m *CharacterCreationModel) loadBackgrounds() tea.Cmd {
	return func() tea.Msg {
		backgrounds, err := m.loader.GetBackgrounds()
		if err != nil {
			return fmt.Errorf("failed to load backgrounds: %w", err)
		}
		
		items := make([]components.ListItem, len(backgrounds.Backgrounds))
		for i, bg := range backgrounds.Backgrounds {
			desc := fmt.Sprintf("Skills: %s", strings.Join(bg.SkillProficiencies, ", "))
			items[i] = components.ListItem{
				Title:       bg.Name,
				Description: desc,
				Value:       &backgrounds.Backgrounds[i],
			}
		}
		
		return BackgroundDataLoadedMsg{Items: items}
	}
}

// RaceDataLoadedMsg is sent when race data is loaded.
type RaceDataLoadedMsg struct {
	Items []components.ListItem
}

// ClassDataLoadedMsg is sent when class data is loaded.
type ClassDataLoadedMsg struct {
	Items []components.ListItem
}

// BackgroundDataLoadedMsg is sent when background data is loaded.
type BackgroundDataLoadedMsg struct {
	Items []components.ListItem
}

// selectCurrentRace selects the currently highlighted race.
func (m *CharacterCreationModel) selectCurrentRace() bool {
	selected := m.raceList.Selected()
	if selected == nil {
		m.err = fmt.Errorf("no race selected")
		return false
	}
	
	race, ok := selected.Value.(*data.Race)
	if !ok {
		m.err = fmt.Errorf("invalid race data")
		return false
	}
	
	m.selectedRace = race
	m.character.Info.Race = race.Name
	m.err = nil
	return true
}

// selectCurrentClass selects the currently highlighted class.
func (m *CharacterCreationModel) selectCurrentClass() bool {
	selected := m.classList.Selected()
	if selected == nil {
		m.err = fmt.Errorf("no class selected")
		return false
	}
	
	class, ok := selected.Value.(*data.Class)
	if !ok {
		m.err = fmt.Errorf("invalid class data")
		return false
	}
	
	m.selectedClass = class
	m.character.Info.Class = class.Name
	m.err = nil
	return true
}

// selectCurrentBackground selects the currently highlighted background.
func (m *CharacterCreationModel) selectCurrentBackground() bool {
	selected := m.backgroundList.Selected()
	if selected == nil {
		m.err = fmt.Errorf("no background selected")
		return false
	}
	
	bg, ok := selected.Value.(*data.Background)
	if !ok {
		m.err = fmt.Errorf("invalid background data")
		return false
	}
	
	m.selectedBackground = bg
	m.character.Info.Background = bg.Name
	
	// Reset background bonus allocation for ability score step
	m.backgroundBonusPattern = BonusPattern2Plus1
	m.backgroundBonus2Target = -1
	m.backgroundBonus1Target = -1
	m.focusedBonusField = BonusFieldPattern
	m.backgroundBonusComplete = false
	
	m.err = nil
	return true
}

// resetAbilityScores resets ability scores based on current mode.
func (m *CharacterCreationModel) resetAbilityScores() {
	switch m.abilityScoreMode {
	case AbilityModeManual:
		m.abilityScores = [6]int{10, 10, 10, 10, 10, 10}
	case AbilityModeStandardArray:
		m.abilityScores = [6]int{0, 0, 0, 0, 0, 0}
		m.standardArrayUsed = [6]bool{false, false, false, false, false, false}
	case AbilityModePointBuy:
		m.abilityScores = [6]int{8, 8, 8, 8, 8, 8}
	}
}

// incrementAbility increases the focused ability score.
func (m *CharacterCreationModel) incrementAbility() *CharacterCreationModel {
	switch m.abilityScoreMode {
	case AbilityModeManual:
		if m.abilityScores[m.focusedAbility] < 20 {
			m.abilityScores[m.focusedAbility]++
		}
	case AbilityModeStandardArray:
		m.cycleStandardArrayValue(true)
	case AbilityModePointBuy:
		newScore := m.abilityScores[m.focusedAbility] + 1
		if newScore <= 15 {
			// Check if we have enough points
			oldCost := models.PointBuyCost(m.abilityScores[m.focusedAbility])
			newCost := models.PointBuyCost(newScore)
			totalPoints := m.calculateCurrentPointBuy()
			if totalPoints - oldCost + newCost <= 27 {
				m.abilityScores[m.focusedAbility] = newScore
			}
		}
	}
	return m
}

// decrementAbility decreases the focused ability score.
func (m *CharacterCreationModel) decrementAbility() *CharacterCreationModel {
	switch m.abilityScoreMode {
	case AbilityModeManual:
		if m.abilityScores[m.focusedAbility] > 3 {
			m.abilityScores[m.focusedAbility]--
		}
	case AbilityModeStandardArray:
		m.cycleStandardArrayValue(false)
	case AbilityModePointBuy:
		if m.abilityScores[m.focusedAbility] > 8 {
			m.abilityScores[m.focusedAbility]--
		}
	}
	return m
}

// cycleStandardArrayValue cycles through available standard array values.
func (m *CharacterCreationModel) cycleStandardArrayValue(forward bool) {
	current := m.abilityScores[m.focusedAbility]
	
	// Find current value index in standard array, or -1 if not set
	currentIdx := -1
	for i, val := range m.standardArrayValues {
		if val == current && m.standardArrayUsed[i] {
			currentIdx = i
			break
		}
	}
	
	// Mark current as unused if it was set
	if currentIdx != -1 {
		m.standardArrayUsed[currentIdx] = false
	}
	
	// Standard array is [15, 14, 13, 12, 10, 8]
	// Forward (right) should go 8->10->12->13->14->15 (decreasing indices)
	// Backward (left) should go 15->14->13->12->10->8 (increasing indices)
	
	// Find starting point for search
	start := 0
	if currentIdx != -1 {
		// Moving from a currently set value
		if forward {
			// Move to lower index (higher value)
			start = currentIdx - 1
			if start < 0 {
				start = len(m.standardArrayValues) - 1
			}
		} else {
			// Move to higher index (lower value)
			start = (currentIdx + 1) % len(m.standardArrayValues)
		}
	} else {
		// Starting from unset (0)
		if forward {
			// Start at the end (index 5, lowest value: 8)
			start = len(m.standardArrayValues) - 1
		} else {
			// Start at the beginning (index 0, highest value: 15)
			start = 0
		}
	}
	
	// Search for next unused value
	for i := 0; i < len(m.standardArrayValues); i++ {
		idx := start
		if forward {
			// Go backwards through array (decreasing index)
			idx = start - i
			if idx < 0 {
				idx += len(m.standardArrayValues)
			}
		} else {
			// Go forwards through array (increasing index)
			idx = (start + i) % len(m.standardArrayValues)
		}
		
		if !m.standardArrayUsed[idx] {
			m.abilityScores[m.focusedAbility] = m.standardArrayValues[idx]
			m.standardArrayUsed[idx] = true
			return
		}
	}
}

// calculateCurrentPointBuy returns the total point buy cost of current scores.
func (m *CharacterCreationModel) calculateCurrentPointBuy() int {
	total := 0
	for _, score := range m.abilityScores {
		total += models.PointBuyCost(score)
	}
	return total
}

// validateAbilityScores checks if ability scores are valid.
func (m *CharacterCreationModel) validateAbilityScores() bool {
	switch m.abilityScoreMode {
	case AbilityModeManual:
		for i, score := range m.abilityScores {
			if score < 3 || score > 20 {
				m.err = fmt.Errorf("ability scores must be between 3 and 20")
				return false
			}
			// Check that score is set (not 0)
			if score == 0 {
				m.err = fmt.Errorf("all ability scores must be assigned")
				return false
			}
			// Skip validation for unused abilities
			_ = i
		}
	case AbilityModeStandardArray:
		// Check that all values are used
		for i, used := range m.standardArrayUsed {
			if !used {
				m.err = fmt.Errorf("all standard array values must be assigned")
				return false
			}
			_ = i
		}
		// Check that all abilities have values
		for _, score := range m.abilityScores {
			if score == 0 {
				m.err = fmt.Errorf("all ability scores must be assigned")
				return false
			}
		}
	case AbilityModePointBuy:
		total := m.calculateCurrentPointBuy()
		if total > 27 {
			m.err = fmt.Errorf("point buy total exceeds 27 points (current: %d)", total)
			return false
		}
	}
	m.err = nil
	return true
}

// finalizeCharacter saves the character and returns to selection screen.
func (m *CharacterCreationModel) finalizeCharacter() (*CharacterCreationModel, tea.Cmd) {
	// Apply ability scores from assignment
	m.character.AbilityScores = models.AbilityScores{
		Strength:     models.AbilityScore{Base: m.abilityScores[0]},
		Dexterity:    models.AbilityScore{Base: m.abilityScores[1]},
		Constitution: models.AbilityScore{Base: m.abilityScores[2]},
		Intelligence: models.AbilityScore{Base: m.abilityScores[3]},
		Wisdom:       models.AbilityScore{Base: m.abilityScores[4]},
		Charisma:     models.AbilityScore{Base: m.abilityScores[5]},
	}
	
	// Apply background ability bonuses
	bonuses := m.getBackgroundBonuses()
	m.character.AbilityScores.Strength.Base += bonuses[0]
	m.character.AbilityScores.Dexterity.Base += bonuses[1]
	m.character.AbilityScores.Constitution.Base += bonuses[2]
	m.character.AbilityScores.Intelligence.Base += bonuses[3]
	m.character.AbilityScores.Wisdom.Base += bonuses[4]
	m.character.AbilityScores.Charisma.Base += bonuses[5]
	
	// Set saving throw proficiencies from class
	if m.selectedClass != nil {
		for _, save := range m.selectedClass.SavingThrowProficiencies {
			switch strings.ToLower(save) {
			case "str", "strength":
				m.character.SavingThrows.SetProficiency(models.AbilityStrength, true)
			case "dex", "dexterity":
				m.character.SavingThrows.SetProficiency(models.AbilityDexterity, true)
			case "con", "constitution":
				m.character.SavingThrows.SetProficiency(models.AbilityConstitution, true)
			case "int", "intelligence":
				m.character.SavingThrows.SetProficiency(models.AbilityIntelligence, true)
			case "wis", "wisdom":
				m.character.SavingThrows.SetProficiency(models.AbilityWisdom, true)
			case "cha", "charisma":
				m.character.SavingThrows.SetProficiency(models.AbilityCharisma, true)
			}
		}
	}
	
	// Set skill proficiencies from selections
	selectedSkills := m.proficiencyManager.GetAllSkills()
	for _, skillName := range selectedSkills {
		m.character.Skills.SetProficiency(skillNameToKey(skillName), models.Proficient)
	}
	
	// Set weapon, armor, and tool proficiencies from class
	if m.selectedClass != nil {
		m.character.Proficiencies.Weapons = append(m.character.Proficiencies.Weapons, m.selectedClass.WeaponProficiencies...)
		m.character.Proficiencies.Armor = append(m.character.Proficiencies.Armor, m.selectedClass.ArmorProficiencies...)
		m.character.Proficiencies.Tools = append(m.character.Proficiencies.Tools, m.selectedClass.ToolProficiencies...)
	}
	
	// Set combat stats
	hitDieType := 8 // default
	if m.selectedClass != nil {
		// Parse hit dice string like "d10" or "1d10"
		hitDiceStr := m.selectedClass.HitDice
		if strings.Contains(hitDiceStr, "d") {
			parts := strings.Split(hitDiceStr, "d")
			if len(parts) == 2 {
				var dieType int
				fmt.Sscanf(parts[1], "%d", &dieType)
				if dieType > 0 {
					hitDieType = dieType
				}
			}
		}
	}
	
	// Calculate starting HP: max hit die + CON modifier
	conMod := m.character.AbilityScores.Constitution.Modifier()
	startingHP := hitDieType + conMod
	if startingHP < 1 {
		startingHP = 1
	}
	
	// Get speed from race (default 30 if not specified)
	speed := 30
	if m.selectedRace != nil && m.selectedRace.Speed > 0 {
		speed = m.selectedRace.Speed
	}
	
	m.character.CombatStats = models.NewCombatStats(startingHP, hitDieType, 1, speed)
	m.character.CombatStats.ArmorClass = 10 + m.character.AbilityScores.Dexterity.Modifier() // Base AC
	
	// Add starting gold
	m.character.Inventory.Currency.AddGold(m.startingGold)
	
	// Add starting equipment
	selectedEquipment := m.getSelectedEquipment()
	for _, equipItem := range selectedEquipment {
		// Convert data.EquipmentItem to models.Item
		item := models.NewItem(
			fmt.Sprintf("%s-%d", strings.ToLower(strings.ReplaceAll(equipItem.Name, " ", "_")), len(m.character.Inventory.Items)),
			equipItem.Name,
			categoryToItemType(equipItem.Category),
		)
		item.Quantity = equipItem.Quantity
		
		// Look up weapon data from equipment database
		if equipItem.Category == data.CategoryWeapon {
			if weapon := m.lookupWeapon(equipItem.Name); weapon != nil {
				item.Damage = weapon.Damage
				item.DamageType = weapon.DamageType
				item.WeaponProps = weapon.Properties
				item.Weight = weapon.Weight
				item.VersatileDamage = weapon.VersatileDamage
				if weapon.Range != nil {
					item.RangeNormal = weapon.Range.Normal
					item.RangeLong = weapon.Range.Long
				}
				item.SubCategory = weapon.SubCategory
			}
		}
		
		// Look up armor data from equipment database
		if equipItem.Category == data.CategoryArmor {
			if armor := m.lookupArmor(equipItem.Name); armor != nil {
				item.Weight = armor.Weight
				item.StealthDisadvantage = armor.StealthDisadvantage
				// Parse AC from armor class string (e.g., "14 + Dex (max 2)")
				// For now, just store the base AC number
				if armor.ArmorClass != "" {
					var baseAC int
					fmt.Sscanf(armor.ArmorClass, "%d", &baseAC)
					item.ArmorClass = baseAC
				}
			}
		}
		
		m.character.Inventory.AddItem(item)
	}
	
	// Apply personality data (strip empty entries)
	for _, trait := range m.personalityTraits {
		if strings.TrimSpace(trait) != "" {
			m.character.Personality.AddTrait(strings.TrimSpace(trait))
		}
	}
	for _, ideal := range m.personalityIdeals {
		if strings.TrimSpace(ideal) != "" {
			m.character.Personality.AddIdeal(strings.TrimSpace(ideal))
		}
	}
	for _, bond := range m.personalityBonds {
		if strings.TrimSpace(bond) != "" {
			m.character.Personality.AddBond(strings.TrimSpace(bond))
		}
	}
	for _, flaw := range m.personalityFlaws {
		if strings.TrimSpace(flaw) != "" {
			m.character.Personality.AddFlaw(strings.TrimSpace(flaw))
		}
	}
	if strings.TrimSpace(m.personalityBackstory) != "" {
		m.character.Personality.Backstory = strings.TrimSpace(m.personalityBackstory)
	}

	// Save character
	path, err := m.storage.Save(m.character)
	if err != nil {
		m.err = fmt.Errorf("failed to save character: %w", err)
		return m, nil
	}
	
	return m, func() tea.Msg {
		return CharacterCreatedMsg{
			Character: m.character,
			Path:      path,
		}
	}
}

// categoryToItemType converts equipment category to ItemType.
func categoryToItemType(category data.EquipmentCategory) models.ItemType {
	switch category {
	case data.CategoryWeapon:
		return models.ItemTypeWeapon
	case data.CategoryArmor:
		return models.ItemTypeArmor
	case data.CategoryPack:
		return models.ItemTypeGeneral
	case data.CategoryTool:
		return models.ItemTypeTool
	default:
		return models.ItemTypeGeneral
	}
}

// getBackgroundBonuses returns an array of bonuses for each ability [STR, DEX, CON, INT, WIS, CHA].
func (m *CharacterCreationModel) getBackgroundBonuses() [6]int {
	bonuses := [6]int{0, 0, 0, 0, 0, 0}
	
	if m.selectedBackground == nil || len(m.selectedBackground.AbilityScores.Options) == 0 {
		return bonuses
	}
	
	options := m.selectedBackground.AbilityScores.Options
	abilityIndices := map[string]int{
		"str": 0, "dex": 1, "con": 2, "int": 3, "wis": 4, "cha": 5,
	}
	
	if m.backgroundBonusPattern == BonusPattern2Plus1 {
		// +2/+1 pattern
		if m.backgroundBonus2Target >= 0 && m.backgroundBonus2Target < len(options) {
			abilityKey := options[m.backgroundBonus2Target]
			if idx, ok := abilityIndices[abilityKey]; ok {
				bonuses[idx] += 2
			}
		}
		if m.backgroundBonus1Target >= 0 && m.backgroundBonus1Target < len(options) {
			abilityKey := options[m.backgroundBonus1Target]
			if idx, ok := abilityIndices[abilityKey]; ok {
				bonuses[idx] += 1
			}
		}
	} else {
		// +1/+1/+1 pattern - distribute to first 3 options
		for i := 0; i < 3 && i < len(options); i++ {
			abilityKey := options[i]
			if idx, ok := abilityIndices[abilityKey]; ok {
				bonuses[idx] += 1
			}
		}
	}
	
	return bonuses
}

// View renders the character creation screen.
func (m *CharacterCreationModel) View() string {
	// Return empty view when quitting to prevent artifacts
	if m.quitting {
		return ""
	}
	
	if m.width == 0 || m.height == 0 {
		return "Error: Terminal size not initialized. Please resize your terminal or restart the application."
	}
	
	var content strings.Builder
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Padding(1, 0)
	title := titleStyle.Render("Create New Character")
	content.WriteString(title)
	content.WriteString("\n\n")
	
	// Error message if present
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)
		content.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		content.WriteString("\n\n")
	}
	
	// Render current step
	switch m.currentStep {
	case StepBasicInfo:
		content.WriteString(m.renderBasicInfo())
	case StepRace:
		content.WriteString(m.renderRaceSelection())
	case StepClass:
		content.WriteString(m.renderClassSelection())
	case StepBackground:
		content.WriteString(m.renderBackgroundSelection())
	case StepAbilities:
		content.WriteString(m.renderAbilityScores())
	case StepProficiencies:
		content.WriteString(m.renderProficiencySelection())
	case StepEquipment:
		content.WriteString(m.renderEquipmentSelection())
	case StepPersonality:
		content.WriteString(m.renderPersonality())
	case StepReview:
		content.WriteString(m.renderReview())
	}
	
	return content.String()
}

// renderBasicInfo renders the basic info step.
func (m *CharacterCreationModel) renderBasicInfo() string {
	var content strings.Builder
	
	stepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)
	content.WriteString(stepStyle.Render("Step 1 of 9: Basic Information"))
	content.WriteString("\n\n")
	
	// Character name input
	m.nameInput.Focused = (m.focusedField == 0)
	content.WriteString(m.nameInput.Render())
	content.WriteString("\n\n")
	
	// Player name input
	m.playerNameInput.Focused = (m.focusedField == 1)
	content.WriteString(m.playerNameInput.Render())
	content.WriteString("\n\n")
	
	// Progression type selection
	labelStyle := lipgloss.NewStyle().Bold(true)
	content.WriteString(labelStyle.Render("Progression Type:"))
	content.WriteString("\n")
	
	// Set focus state on button group based on focused field
	m.progressionList.SetFocused(m.focusedField == 2)
	content.WriteString(m.progressionList.Render())
	content.WriteString("\n\n")
	
	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	content.WriteString(helpStyle.Render("Tab: Next field | Enter: Continue | Esc: Cancel"))
	
	return content.String()
}

// renderRaceSelection renders the race selection step.
func (m *CharacterCreationModel) renderRaceSelection() string {
	var content strings.Builder
	
	stepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)
	content.WriteString(stepStyle.Render("Step 2 of 9: Race Selection"))
	content.WriteString("\n\n")
	
	// Render race list
	m.raceList.Width = m.width - 4
	m.raceList.Height = m.height - 15
	content.WriteString(m.raceList.Render())
	content.WriteString("\n\n")
	
	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	content.WriteString(helpStyle.Render("↑/↓: Navigate | Enter: Select | Esc: Back"))
	
	return content.String()
}

// renderClassSelection renders the class selection step.
func (m *CharacterCreationModel) renderClassSelection() string {
	var content strings.Builder
	
	stepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)
	content.WriteString(stepStyle.Render("Step 3 of 9: Class Selection"))
	content.WriteString("\n\n")
	
	// Render class list
	m.classList.Width = m.width - 4
	m.classList.Height = m.height - 15
	content.WriteString(m.classList.Render())
	content.WriteString("\n\n")
	
	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	content.WriteString(helpStyle.Render("↑/↓: Navigate | Enter: Select | Esc: Back"))
	
	return content.String()
}

// renderBackgroundSelection renders the background selection step.
func (m *CharacterCreationModel) renderBackgroundSelection() string {
	var content strings.Builder
	
	stepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)
	content.WriteString(stepStyle.Render("Step 4 of 9: Background Selection"))
	content.WriteString("\n\n")
	
	// Show background list
	m.backgroundList.Width = m.width - 4
	m.backgroundList.Height = m.height - 15
	content.WriteString(m.backgroundList.Render())
	content.WriteString("\n\n")
	
	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	content.WriteString(helpStyle.Render("↑/↓: Navigate | Enter: Next | Esc: Back"))
	
	return content.String()
}

// renderBackgroundBonusSelection renders the background bonus allocation UI at the top of ability scores.
func (m *CharacterCreationModel) renderBackgroundBonusSelection() string {
	var content strings.Builder
	
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	
	content.WriteString(titleStyle.Render("Background Ability Score Bonuses"))
	content.WriteString("\n")
	content.WriteString(helpStyle.Render(fmt.Sprintf("From: %s", m.selectedBackground.Name)))
	
	// Show completion status
	if m.backgroundBonusComplete {
		completeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
		content.WriteString(" ")
		content.WriteString(completeStyle.Render("✓ Allocated"))
	}
	content.WriteString("\n\n")
	
	options := m.selectedBackground.AbilityScores.Options
	points := m.selectedBackground.AbilityScores.Points
	
	abilityFullNames := map[string]string{
		"str": "Strength", "dex": "Dexterity", "con": "Constitution",
		"int": "Intelligence", "wis": "Wisdom", "cha": "Charisma",
	}
	
	// Pattern selection
	canUse2Plus1 := len(options) >= 2 && points == 3
	_ = canUse2Plus1 // May be used for validation in future
	canUse1Plus1Plus1 := len(options) >= 3 && points == 3
	_ = canUse1Plus1Plus1 // May be used for validation in future
	
	var patternStr string
	if m.backgroundBonusPattern == BonusPattern2Plus1 {
		patternStr = "+2 / +1"
	} else {
		patternStr = "+1 / +1 / +1"
	}
	
	var lineStyle lipgloss.Style
	if m.focusedSection == SectionBackgroundBonus && m.focusedBonusField == BonusFieldPattern {
		lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true)
		content.WriteString("▶ ")
	} else {
		lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
		content.WriteString("  ")
	}
	
	content.WriteString(lineStyle.Render(fmt.Sprintf("Pattern: %s", patternStr)))
	content.WriteString("\n\n")
	
	// Ability selections (if +2/+1 pattern)
	if m.backgroundBonusPattern == BonusPattern2Plus1 && canUse2Plus1 {
		// +2 target
		plus2Ability := "None"
		if m.backgroundBonus2Target >= 0 && m.backgroundBonus2Target < len(options) {
			key := options[m.backgroundBonus2Target]
			plus2Ability = abilityFullNames[key]
		}
		
		if m.focusedSection == SectionBackgroundBonus && m.focusedBonusField == BonusFieldPlus2Target {
			lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true)
			content.WriteString("▶ ")
		} else {
			lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
			content.WriteString("  ")
		}
		content.WriteString(lineStyle.Render(fmt.Sprintf("+2 Bonus: %s", plus2Ability)))
		content.WriteString("\n")
		
		// +1 target
		plus1Ability := "None"
		if m.backgroundBonus1Target >= 0 && m.backgroundBonus1Target < len(options) {
			key := options[m.backgroundBonus1Target]
			plus1Ability = abilityFullNames[key]
		}
		
		if m.focusedSection == SectionBackgroundBonus && m.focusedBonusField == BonusFieldPlus1Target {
			lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true)
			content.WriteString("▶ ")
		} else {
			lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
			content.WriteString("  ")
		}
		content.WriteString(lineStyle.Render(fmt.Sprintf("+1 Bonus: %s", plus1Ability)))
		content.WriteString("\n")
	} else if m.backgroundBonusPattern == BonusPattern1Plus1Plus1 {
		// +1/+1/+1 auto-allocation
		content.WriteString(helpStyle.Render("Auto-allocated to: "))
		var allocated []string
		for i := 0; i < 3 && i < len(options); i++ {
			key := options[i]
			allocated = append(allocated, abilityFullNames[key])
		}
		content.WriteString(helpStyle.Render(strings.Join(allocated, ", ")))
		content.WriteString("\n")
	}
	
	return content.String()
}

// renderAbilityScores renders the ability score assignment step.
func (m *CharacterCreationModel) renderAbilityScores() string {
	var content strings.Builder
	
	stepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)
	content.WriteString(stepStyle.Render("Step 5 of 9: Ability Score Assignment"))
	content.WriteString("\n\n")
	
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	
	// Background ability bonus selection (if applicable)
	if m.selectedBackground != nil && len(m.selectedBackground.AbilityScores.Options) > 0 {
		content.WriteString(m.renderBackgroundBonusSelection())
		content.WriteString("\n\n")
		
		// Show separator
		separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		content.WriteString(separatorStyle.Render("─────────────────────────────────────────"))
		content.WriteString("\n\n")
	}
	
	// Mode selector
	modeStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	var modeName string
	switch m.abilityScoreMode {
	case AbilityModeManual:
		modeName = "Manual Entry"
	case AbilityModeStandardArray:
		modeName = "Standard Array"
	case AbilityModePointBuy:
		modeName = "Point Buy"
	}
	content.WriteString(modeStyle.Render(fmt.Sprintf("Mode: %s", modeName)))
	content.WriteString("\n")
	
	content.WriteString(helpStyle.Render("Press 'm' to change mode"))
	content.WriteString("\n\n")
	
	// Ability names
	abilityNames := []string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}
	abilityFullNames := []string{"Strength", "Dexterity", "Constitution", "Intelligence", "Wisdom", "Charisma"}
	
	// Background bonuses
	backgroundBonuses := m.getBackgroundBonuses()
	
	// Render ability scores
	for i := 0; i < 6; i++ {
		base := m.abilityScores[i]
		bonus := backgroundBonuses[i]
		final := base + bonus
		modifier := (final - 10) / 2
		modifierStr := fmt.Sprintf("%+d", modifier)
		
		// Style based on focus (only show focus if in ability score section)
		var lineStyle lipgloss.Style
		if m.focusedSection == SectionAbilityScores && AbilityIndex(i) == m.focusedAbility {
			lineStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("13")).
				Bold(true)
		} else {
			lineStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15"))
		}
		
		// Build line
		var line strings.Builder
		if m.focusedSection == SectionAbilityScores && AbilityIndex(i) == m.focusedAbility {
			line.WriteString("▶ ")
		} else {
			line.WriteString("  ")
		}
		
		line.WriteString(fmt.Sprintf("%-3s  %-14s  ", abilityNames[i], abilityFullNames[i]))
		
		// Show base score
		line.WriteString(fmt.Sprintf("%2d", base))
		
		// Show bonus if any
		if bonus > 0 {
			line.WriteString(fmt.Sprintf(" + %d", bonus))
		} else {
			line.WriteString("    ")
		}
		
		// Show final and modifier
		line.WriteString(fmt.Sprintf("  =  %2d  (%s)", final, modifierStr))
		
		content.WriteString(lineStyle.Render(line.String()))
		content.WriteString("\n")
	}
	
	content.WriteString("\n")
	
	// Mode-specific info
	switch m.abilityScoreMode {
	case AbilityModeManual:
		content.WriteString(helpStyle.Render("Range: 3-20 for each ability"))
		content.WriteString("\n")
	case AbilityModeStandardArray:
		content.WriteString(helpStyle.Render("Assign values: 15, 14, 13, 12, 10, 8"))
		content.WriteString("\n")
		// Show available values
		var available []int
		for i, val := range m.standardArrayValues {
			if !m.standardArrayUsed[i] {
				available = append(available, val)
			}
		}
		if len(available) > 0 {
			content.WriteString(helpStyle.Render(fmt.Sprintf("Available: %v", available)))
			content.WriteString("\n")
		}
	case AbilityModePointBuy:
		total := m.calculateCurrentPointBuy()
		remaining := 27 - total
		pointsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
		if remaining < 0 {
			pointsStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
		}
		content.WriteString(pointsStyle.Render(fmt.Sprintf("Points Used: %d / 27  (Remaining: %d)", total, remaining)))
		content.WriteString("\n")
		content.WriteString(helpStyle.Render("Range: 8-15 for each ability"))
		content.WriteString("\n")
	}
	
	// Background bonus info
	if m.selectedBackground != nil && len(m.selectedBackground.AbilityScores.Options) > 0 {
		content.WriteString("\n")
		bonusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		if m.backgroundBonusComplete {
			content.WriteString(bonusStyle.Render("Background Bonuses: Allocated ✓"))
		} else {
			content.WriteString(bonusStyle.Render("Background Bonuses: Use ↑/↓ to navigate sections"))
		}
		content.WriteString("\n")
	}
	
	content.WriteString("\n")
	
	// Unified help text
	content.WriteString(helpStyle.Render("↑/↓: Navigate | ←/→: Adjust | m: Change mode | Enter: Confirm/Finish | Esc: Back"))
	
	return content.String()
}

// renderEquipmentSelection renders the starting equipment step.
func (m *CharacterCreationModel) renderEquipmentSelection() string {
	var content strings.Builder
	
	// If we're in sub-selection mode, show that instead
	if m.inEquipmentSubSelection {
		return m.renderEquipmentSubSelection()
	}
	
	stepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)
	content.WriteString(stepStyle.Render("Step 7 of 9: Starting Equipment"))
	content.WriteString("\n\n")
	
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	focusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true)
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	
	content.WriteString(titleStyle.Render("Starting Equipment"))
	content.WriteString("\n\n")
	
	// Display class starting equipment
	if m.selectedClass != nil && len(m.selectedClass.StartingEquipment) > 0 {
		content.WriteString(helpStyle.Render("Select your starting equipment:"))
		content.WriteString("\n\n")
		
		for i, equip := range m.selectedClass.StartingEquipment {
			isFocused := (i == m.focusedEquipmentChoice)
			
			if equip.Type == data.EquipmentChoiceFixed && equip.Item != nil {
				// Fixed item - no choice needed
				prefix := "  "
				if isFocused {
					prefix = "▶ "
				}
				itemText := fmt.Sprintf("%d. %d× %s", i+1, equip.Item.Quantity, equip.Item.Name)
				if isFocused {
					content.WriteString(prefix + focusStyle.Render(itemText) + " ✓\n")
				} else {
					content.WriteString(prefix + itemText + " ✓\n")
				}
				
				// If it's a pack, show contents
				if equip.Item.Category == data.CategoryPack {
					packContents := m.getPackContents(equip.Item.Name)
					if len(packContents) > 0 {
						dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
						for _, packItem := range packContents {
							content.WriteString(fmt.Sprintf("      - %s\n", dimStyle.Render(packItem)))
						}
					}
				}
			} else if equip.Type == data.EquipmentChoiceSelect && len(equip.Options) > 0 {
				// Choice - show all options vertically
				choiceText := fmt.Sprintf("%d. Choose one:", i+1)
				content.WriteString("  " + choiceText + "\n")
				
				// Show options
				selectedIdx := -1
				if i < len(m.equipmentChoices) {
					selectedIdx = m.equipmentChoices[i]
				}
				
				for j, opt := range equip.Options {
					if len(opt.Items) > 0 {
						isOptionFocused := (i == m.focusedEquipmentChoice && j == m.focusedEquipmentOption)
						isOptionSelected := (j == selectedIdx)
						
						itemStrs := []string{}
						for _, item := range opt.Items {
							displayName := item.Name
							// If this item has a filter, show descriptive text with sub-selection
							if m.needsItemSubSelection(item) {
								filterName := m.getFilterDisplayName(item.Filter)
								if isOptionSelected && m.equipmentSubSelections != nil && m.equipmentSubSelections[i] != "" {
									// Show "any martial weapon [Longsword]" format
									displayName = fmt.Sprintf("%s [%s]", filterName, m.equipmentSubSelections[i])
								} else {
									// Just show "any martial weapon"
									displayName = filterName
								}
							}
							
							if item.Quantity > 1 {
								itemStrs = append(itemStrs, fmt.Sprintf("%d× %s", item.Quantity, displayName))
							} else {
								itemStrs = append(itemStrs, displayName)
							}
						}
						
						optionText := fmt.Sprintf("(%c) %s", 'a'+j, strings.Join(itemStrs, ", "))
						
						// Determine prefix and styling
						prefix := "     "
						if isOptionFocused {
							prefix = "   ▶ "
						}
						
						// Add checkmark if selected
						suffix := ""
						if isOptionSelected {
							// Check if it needs sub-selection
							needsSub := false
							for _, item := range opt.Items {
								if m.needsItemSubSelection(item) {
									needsSub = true
									break
								}
							}
							
							if needsSub && (m.equipmentSubSelections == nil || m.equipmentSubSelections[i] == "") {
								suffix = " [Press Space to choose]"
							} else {
								suffix = " ✓"
							}
						}
						
						// Apply styling
						if isOptionFocused {
							content.WriteString(prefix + focusStyle.Render(optionText+suffix) + "\n")
						} else if isOptionSelected {
							content.WriteString(prefix + selectedStyle.Render(optionText+suffix) + "\n")
						} else {
							content.WriteString(prefix + optionText + suffix + "\n")
						}
						
						// If this option contains a pack, show its contents
						for _, item := range opt.Items {
							if item.Category == data.CategoryPack && item.Name != "" {
								packContents := m.getPackContents(item.Name)
								if len(packContents) > 0 {
									dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
									for _, packItem := range packContents {
										content.WriteString(fmt.Sprintf("        - %s\n", dimStyle.Render(packItem)))
									}
								}
							}
						}
					}
				}
			}
		}
	} else {
		content.WriteString(helpStyle.Render("No starting equipment specified"))
		content.WriteString("\n")
	}
	
	content.WriteString("\n")
	
	// Starting gold
	startingGold := m.getStartingGold()
	content.WriteString(labelStyle.Render(fmt.Sprintf("Starting Gold: %d gp", startingGold)))
	content.WriteString("\n\n")
	
	// Help text
	allSelected := m.allEquipmentChoicesMade()
	if allSelected {
		content.WriteString(helpStyle.Render("↑/↓: Navigate | Enter: Continue to Review | Esc: Back"))
	} else {
		content.WriteString(helpStyle.Render("↑/↓: Navigate | Space: Select | Enter: Continue (when ready) | Esc: Back"))
	}
	
	return content.String()
}

// renderEquipmentSubSelection renders the sub-selection UI for "any [category]" items.
func (m *CharacterCreationModel) renderEquipmentSubSelection() string {
	var content strings.Builder
	
	stepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)
	content.WriteString(stepStyle.Render("Step 7 of 9: Starting Equipment - Choose Specific Item"))
	content.WriteString("\n\n")
	
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	focusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true)
	
	content.WriteString(titleStyle.Render("Choose Your Item"))
	content.WriteString("\n\n")
	
	content.WriteString(helpStyle.Render("Select from the available items:"))
	content.WriteString("\n\n")
	
	// Display items in sub-selection
	for i, item := range m.equipmentSubItems {
		isFocused := (i == m.focusedSubItem)
		prefix := "  "
		if isFocused {
			prefix = "▶ "
		}
		
		if isFocused {
			content.WriteString(prefix + focusStyle.Render(item) + "\n")
		} else {
			content.WriteString(prefix + item + "\n")
		}
	}
	
	content.WriteString("\n")
	content.WriteString(helpStyle.Render("↑/↓: Navigate | Space/Enter: Select | Esc: Cancel"))
	
	return content.String()
}

// renderPersonality renders the personality step.
func (m *CharacterCreationModel) renderPersonality() string {
	var content strings.Builder

	stepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
	content.WriteString(stepStyle.Render("Step 8 of 9: Personality"))
	content.WriteString("\n\n")

	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Italic(true)
	content.WriteString(infoStyle.Render("All fields are optional. You can fill these in later from the Character Info view."))
	content.WriteString("\n\n")

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))

	fields := []struct {
		label string
		items []string
		index int
	}{
		{"Personality Traits", m.personalityTraits, 0},
		{"Ideals", m.personalityIdeals, 1},
		{"Bonds", m.personalityBonds, 2},
		{"Flaws", m.personalityFlaws, 3},
	}

	for _, field := range fields {
		focused := m.personalityFocusField == field.index
		content.WriteString(headerStyle.Render(field.label + ":"))
		content.WriteString("\n")

		for i, item := range field.items {
			prefix := "  "
			if focused && i == m.personalityItemCursor {
				prefix = cursorStyle.Render("> ")
				if m.personalityEditing {
					content.WriteString(prefix + m.personalityEditBuffer + "█\n")
					continue
				}
			}
			if item == "" {
				content.WriteString(prefix + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("(empty)") + "\n")
			} else {
				content.WriteString(prefix + item + "\n")
			}
		}

		if focused && m.personalityItemCursor == len(field.items) {
			content.WriteString(cursorStyle.Render("> ") + "+ Add another\n")
		} else {
			content.WriteString("  + Add another\n")
		}
		content.WriteString("\n")
	}

	// Backstory
	bsFocused := m.personalityFocusField == 4
	content.WriteString(headerStyle.Render("Backstory:"))
	content.WriteString("\n")
	if bsFocused && m.personalityEditing {
		content.WriteString("  " + m.personalityEditBuffer + "█\n")
	} else if m.personalityBackstory == "" {
		content.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("(empty — press Enter to write)") + "\n")
	} else {
		content.WriteString("  " + m.personalityBackstory + "\n")
	}

	content.WriteString("\n")

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	content.WriteString(helpStyle.Render("Tab: next field/continue • Enter: edit/add • d: delete • Esc: back"))

	return content.String()
}

// renderReview renders the final review step before saving.
func (m *CharacterCreationModel) renderReview() string {
	var content strings.Builder
	
	stepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)
	content.WriteString(stepStyle.Render("Step 9 of 9: Review & Confirm"))
	content.WriteString("\n\n")
	
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	
	content.WriteString(titleStyle.Render("Character Summary"))
	content.WriteString("\n\n")
	
	// Basic info
	content.WriteString(labelStyle.Render("Name: "))
	content.WriteString(m.character.Info.Name)
	content.WriteString("\n")
	
	content.WriteString(labelStyle.Render("Player: "))
	content.WriteString(m.character.Info.PlayerName)
	content.WriteString("\n")
	
	content.WriteString(labelStyle.Render("Progression: "))
	if m.character.Info.ProgressionType == models.ProgressionXP {
		content.WriteString("XP Tracking")
	} else {
		content.WriteString("Milestone")
	}
	content.WriteString("\n\n")
	
	// Race, Class, Background
	if m.selectedRace != nil {
		content.WriteString(labelStyle.Render("Race: "))
		content.WriteString(m.selectedRace.Name)
		if m.selectedSubtype >= 0 && m.selectedSubtype < len(m.selectedRace.Subtypes) {
			content.WriteString(fmt.Sprintf(" (%s)", m.selectedRace.Subtypes[m.selectedSubtype].Name))
		}
		content.WriteString("\n")
	}
	
	if m.selectedClass != nil {
		content.WriteString(labelStyle.Render("Class: "))
		content.WriteString(m.selectedClass.Name)
		content.WriteString(" (Level 1)\n")
	}
	
	if m.selectedBackground != nil {
		content.WriteString(labelStyle.Render("Background: "))
		content.WriteString(m.selectedBackground.Name)
		content.WriteString("\n\n")
	}
	
	// Ability Scores
	content.WriteString(labelStyle.Render("Ability Scores:"))
	content.WriteString("\n")
	
	abilityNames := []string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}
	backgroundBonuses := m.getBackgroundBonuses()
	
	for i := 0; i < 6; i++ {
		base := m.abilityScores[i]
		bonus := backgroundBonuses[i]
		final := base + bonus
		modifier := (final - 10) / 2
		
		content.WriteString(fmt.Sprintf("  %s: %d", abilityNames[i], final))
		if bonus > 0 {
			content.WriteString(fmt.Sprintf(" (%d + %d)", base, bonus))
		}
		content.WriteString(fmt.Sprintf(" [%+d]\n", modifier))
	}
	
	content.WriteString("\n")
	
	// Proficiencies
	content.WriteString(labelStyle.Render("Proficiencies:"))
	content.WriteString("\n")
	
	// Skills (class-selected + background-granted)
	allSkills := m.proficiencyManager.GetAllSkills()
	if len(allSkills) > 0 {
		content.WriteString("  Skills: ")
		content.WriteString(strings.Join(allSkills, ", "))
		content.WriteString("\n")
	}
	
	// Tools (selected + background-granted)
	allTools := m.proficiencyManager.GetAllTools()
	if len(allTools) > 0 {
		content.WriteString("  Tools: ")
		content.WriteString(strings.Join(allTools, ", "))
		content.WriteString("\n")
	}
	
	// Languages (selected + racial)
	allLanguages := m.proficiencyManager.GetAllLanguages()
	if len(allLanguages) > 0 {
		content.WriteString("  Languages: ")
		content.WriteString(strings.Join(allLanguages, ", "))
		content.WriteString("\n")
	}
	
	content.WriteString("\n")
	
	// Starting Equipment
	content.WriteString(labelStyle.Render("Starting Equipment:"))
	content.WriteString("\n")
	
	selectedEquipment := m.getSelectedEquipment()
	if len(selectedEquipment) > 0 {
		// Group items by type for better display
		equipmentByType := make(map[data.EquipmentCategory][]data.EquipmentItem)
		for _, item := range selectedEquipment {
			equipmentByType[item.Category] = append(equipmentByType[item.Category], item)
		}

		// Display in order: weapon, armor, gear, tool, pack (pack last so contents are at end)
		categories := []data.EquipmentCategory{data.CategoryWeapon, data.CategoryArmor, data.CategoryGear, data.CategoryTool, data.CategoryPack}
		for _, cat := range categories {
			items := equipmentByType[cat]
			if len(items) > 0 {
				for _, item := range items {
					if item.Quantity > 1 {
						content.WriteString(fmt.Sprintf("  • %d× %s\n", item.Quantity, item.Name))
					} else {
						content.WriteString(fmt.Sprintf("  • %s\n", item.Name))
					}

					// If it's a pack, show contents
					if cat == data.CategoryPack {
						packContents := m.getPackContents(item.Name)
						if len(packContents) > 0 {
							for _, packItem := range packContents {
								content.WriteString(fmt.Sprintf("      - %s\n", packItem))
							}
						}
					}
				}
			}
		}
	} else {
		content.WriteString("  (none)\n")
	}
	
	content.WriteString("\n")
	
	// Starting Gold
	startingGold := m.getStartingGold()
	content.WriteString(labelStyle.Render(fmt.Sprintf("Starting Gold: %d gp", startingGold)))
	content.WriteString("\n\n")
	
	content.WriteString(helpStyle.Render("Enter: Save Character | Esc: Back to Equipment"))
	
	return content.String()
}

// getStartingGold returns the starting gold for the character's class.
func (m *CharacterCreationModel) getStartingGold() int {
	if m.selectedClass == nil {
		return 0
	}
	
	// Standard D&D 5e starting gold by class
	startingGoldByClass := map[string]int{
		"Barbarian": 50,  // 2d4 × 10 gp (average)
		"Bard":      125, // 5d4 × 10 gp (average)
		"Cleric":    125, // 5d4 × 10 gp (average)
		"Druid":     50,  // 2d4 × 10 gp (average)
		"Fighter":   125, // 5d4 × 10 gp (average)
		"Monk":      13,  // 5d4 gp (average)
		"Paladin":   125, // 5d4 × 10 gp (average)
		"Ranger":    125, // 5d4 × 10 gp (average)
		"Rogue":     100, // 4d4 × 10 gp (average)
		"Sorcerer":  75,  // 3d4 × 10 gp (average)
		"Warlock":   100, // 4d4 × 10 gp (average)
		"Wizard":    100, // 4d4 × 10 gp (average)
	}
	
	if gold, ok := startingGoldByClass[m.selectedClass.Name]; ok {
		return gold
	}
	
	return 100 // Default
}

// getPackContents returns the contents of a pack by name, or nil if not found or not a pack.
func (m *CharacterCreationModel) getPackContents(packName string) []string {
	equipment, err := m.loader.GetEquipment()
	if err != nil || equipment == nil {
		return nil
	}
	
	for _, pack := range equipment.Packs {
		if pack.Name == packName {
			return pack.Contents
		}
	}
	
	return nil
}

// lookupWeapon finds a weapon by name in the equipment database.
func (m *CharacterCreationModel) lookupWeapon(name string) *data.Weapon {
	equipment, err := m.loader.GetEquipment()
	if err != nil || equipment == nil {
		return nil
	}
	
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	
	for _, w := range equipment.Weapons.GetAllWeapons() {
		if strings.ToLower(w.Name) == normalizedName {
			return &w
		}
	}
	
	return nil
}

// lookupArmor finds an armor item by name in the equipment database.
func (m *CharacterCreationModel) lookupArmor(name string) *data.ArmorItem {
	equipment, err := m.loader.GetEquipment()
	if err != nil || equipment == nil {
		return nil
	}
	
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	
	for _, a := range equipment.Armor.GetAllArmor() {
		if strings.ToLower(a.Name) == normalizedName {
			return &a
		}
	}
	
	return nil
}
