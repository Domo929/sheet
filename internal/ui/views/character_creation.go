package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	progressionType string // "xp" or "milestone"
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
	abilitySelections    []string // Selected ability scores for background bonus
	
	// Navigation
	helpFooter components.HelpFooter
	buttons    components.ButtonGroup
	err        error
}

// NewCharacterCreationModel creates a new character creation model.
func NewCharacterCreationModel(store *storage.CharacterStorage, loader *data.Loader) *CharacterCreationModel {
	nameInput := components.NewTextInput("Character Name", "Enter character name...")
	nameInput.Width = 40
	
	playerNameInput := components.NewTextInput("Player Name", "Enter player name...")
	playerNameInput.Width = 40
	
	progressionButtons := components.NewButtonGroup("XP Tracking", "Milestone")
	
	helpFooter := components.NewHelpFooter()
	
	navButtons := components.NewButtonGroup("Next", "Cancel")
	
	return &CharacterCreationModel{
		storage:         store,
		loader:          loader,
		currentStep:     StepBasicInfo,
		nameInput:       nameInput,
		playerNameInput: playerNameInput,
		progressionList: progressionButtons,
		progressionType: "xp",
		focusedField:    0,
		helpFooter:      helpFooter,
		buttons:         navButtons,
		selectedSubtype: -1,
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
		
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	
	return m, nil
}

// handleKey processes keyboard input based on current step.
func (m *CharacterCreationModel) handleKey(msg tea.KeyMsg) (*CharacterCreationModel, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
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
	default:
		return m, nil
	}
}

// handleBasicInfoKeys handles keys for the basic info step.
func (m *CharacterCreationModel) handleBasicInfoKeys(msg tea.KeyMsg) (*CharacterCreationModel, tea.Cmd) {
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
		// Cancel creation
		m.err = fmt.Errorf("character creation cancelled")
		return m, tea.Quit
		
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
	}
	
	// Handle text input - only process actual character runes
	if msg.Type == tea.KeyRunes && (m.focusedField == 0 || m.focusedField == 1) {
		char := string(msg.Runes)
		if m.focusedField == 0 {
			m.nameInput.Value += char
		} else if m.focusedField == 1 {
			m.playerNameInput.Value += char
		}
	}
	return m, nil
}

// handleRaceKeys handles keys for the race selection step.
func (m *CharacterCreationModel) handleRaceKeys(msg tea.KeyMsg) (*CharacterCreationModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.raceList.MoveUp()
		return m, nil
		
	case "down", "j":
		m.raceList.MoveDown()
		return m, nil
		
	case "enter":
		// Select race and move to class selection
		if m.selectCurrentRace() {
			return m.moveToStep(StepClass)
		}
		return m, nil
		
	case "esc":
		// Go back to basic info
		return m.moveToStep(StepBasicInfo)
	}
	
	return m, nil
}

// handleClassKeys handles keys for the class selection step.
func (m *CharacterCreationModel) handleClassKeys(msg tea.KeyMsg) (*CharacterCreationModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.classList.MoveUp()
		return m, nil
		
	case "down", "j":
		m.classList.MoveDown()
		return m, nil
		
	case "enter":
		// Select class and move to background selection
		if m.selectCurrentClass() {
			return m.moveToStep(StepBackground)
		}
		return m, nil
		
	case "esc":
		// Go back to race selection
		return m.moveToStep(StepRace)
	}
	
	return m, nil
}

// handleBackgroundKeys handles keys for the background selection step.
func (m *CharacterCreationModel) handleBackgroundKeys(msg tea.KeyMsg) (*CharacterCreationModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.backgroundList.MoveUp()
		return m, nil
		
	case "down", "j":
		m.backgroundList.MoveDown()
		return m, nil
		
	case "enter":
		// Select background and move to ability scores (future phase)
		if m.selectCurrentBackground() {
			// For now, finalize character creation
			return m.finalizeCharacter()
		}
		return m, nil
		
	case "esc":
		// Go back to class selection
		return m.moveToStep(StepClass)
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
	if m.progressionType == "xp" {
		m.character.Info.ProgressionType = models.ProgressionXP
	} else {
		m.character.Info.ProgressionType = models.ProgressionMilestone
	}
}

// updateProgressionType updates the progression type based on button selection.
func (m *CharacterCreationModel) updateProgressionType() {
	if m.progressionList.SelectedIndex == 0 {
		m.progressionType = "xp"
	} else {
		m.progressionType = "milestone"
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
	m.err = nil
	return true
}

// finalizeCharacter saves the character and returns to selection screen.
func (m *CharacterCreationModel) finalizeCharacter() (*CharacterCreationModel, tea.Cmd) {
	// For now, save with default ability scores
	// This will be replaced when ability score assignment is implemented
	m.character.AbilityScores = models.AbilityScores{
		Strength:     models.AbilityScore{Base: 10},
		Dexterity:    models.AbilityScore{Base: 10},
		Constitution: models.AbilityScore{Base: 10},
		Intelligence: models.AbilityScore{Base: 10},
		Wisdom:       models.AbilityScore{Base: 10},
		Charisma:     models.AbilityScore{Base: 10},
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

// View renders the character creation screen.
func (m *CharacterCreationModel) View() string {
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
	}
	
	return content.String()
}

// renderBasicInfo renders the basic info step.
func (m *CharacterCreationModel) renderBasicInfo() string {
	var content strings.Builder
	
	stepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)
	content.WriteString(stepStyle.Render("Step 1 of 4: Basic Information"))
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
	
	// Highlight progression buttons if focused
	if m.focusedField == 2 {
		focusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
		content.WriteString(focusStyle.Render(m.progressionList.Render()))
	} else {
		content.WriteString(m.progressionList.Render())
	}
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
	content.WriteString(stepStyle.Render("Step 2 of 4: Race Selection"))
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
	content.WriteString(stepStyle.Render("Step 3 of 4: Class Selection"))
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
	content.WriteString(stepStyle.Render("Step 4 of 4: Background Selection"))
	content.WriteString("\n\n")
	
	// Render background list
	m.backgroundList.Width = m.width - 4
	m.backgroundList.Height = m.height - 15
	content.WriteString(m.backgroundList.Render())
	content.WriteString("\n\n")
	
	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	content.WriteString(helpStyle.Render("↑/↓: Navigate | Enter: Finish | Esc: Back"))
	
	return content.String()
}
