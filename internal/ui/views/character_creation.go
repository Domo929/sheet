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
	abilitySelections    []string // Selected ability scores for background bonus
	
	// Ability Score step
	abilityScoreMode     AbilityScoreMode // Manual, Standard Array, or Point Buy
	abilityScores        [6]int           // STR, DEX, CON, INT, WIS, CHA
	focusedAbility       int              // Which ability is currently focused (0-5)
	standardArrayValues  []int            // For standard array mode
	standardArrayUsed    [6]bool          // Track which standard array values are used
	
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
		
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	
	return m, nil
}

// handleKey processes keyboard input based on current step.
func (m *CharacterCreationModel) handleKey(msg tea.KeyMsg) (*CharacterCreationModel, tea.Cmd) {
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
		
	case " ":
		// Handle space in text input fields
		if m.focusedField == 0 {
			m.nameInput.Value += " "
		} else if m.focusedField == 1 {
			m.playerNameInput.Value += " "
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
	return m.handleListSelectionKeys(msg, &m.raceList, m.selectCurrentRace, StepBasicInfo, StepClass)
}

// handleClassKeys handles keys for the class selection step.
func (m *CharacterCreationModel) handleClassKeys(msg tea.KeyMsg) (*CharacterCreationModel, tea.Cmd) {
	return m.handleListSelectionKeys(msg, &m.classList, m.selectCurrentClass, StepRace, StepBackground)
}

// handleBackgroundKeys handles keys for the background selection step.
func (m *CharacterCreationModel) handleBackgroundKeys(msg tea.KeyMsg) (*CharacterCreationModel, tea.Cmd) {
	return m.handleListSelectionKeys(msg, &m.backgroundList, func() bool {
		return m.selectCurrentBackground()
	}, StepClass, StepAbilities)
}

// handleAbilityKeys handles keys for the ability score assignment step.
func (m *CharacterCreationModel) handleAbilityKeys(msg tea.KeyMsg) (*CharacterCreationModel, tea.Cmd) {
	switch msg.String() {
	case "tab", "right", "l":
		// Move to next ability
		m.focusedAbility = (m.focusedAbility + 1) % 6
		return m, nil
		
	case "shift+tab", "left", "h":
		// Move to previous ability
		m.focusedAbility--
		if m.focusedAbility < 0 {
			m.focusedAbility = 5
		}
		return m, nil
		
	case "m":
		// Toggle mode: Manual -> Standard Array -> Point Buy -> Manual
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
		return m, nil
		
	case "up", "k", "+", "=":
		return m.incrementAbility(), nil
		
	case "down", "j", "-", "_":
		return m.decrementAbility(), nil
		
	case "enter":
		// Validate and move to finalize
		if m.validateAbilityScores() {
			return m.finalizeCharacter()
		}
		return m, nil
		
	case "esc":
		// Go back to background selection
		return m.moveToStep(StepBackground)
	}
	
	return m, nil
}

// handleListSelectionKeys is a generic handler for list-based selection steps (race, class, background).
func (m *CharacterCreationModel) handleListSelectionKeys(
	msg tea.KeyMsg,
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
	
	// Auto-allocate background ability bonuses (MVP: +2/+1 pattern)
	// TODO: Allow user to choose allocation in a future enhancement
	if len(bg.AbilityScores.Options) >= 2 && bg.AbilityScores.Points == 3 {
		// Common case: +2/+1 split
		m.abilitySelections = []string{
			bg.AbilityScores.Options[0],
			bg.AbilityScores.Options[0], // +2 to first option
			bg.AbilityScores.Options[1], // +1 to second option
		}
	} else if len(bg.AbilityScores.Options) >= 3 && bg.AbilityScores.Points == 3 {
		// Alternative: +1/+1/+1 split
		m.abilitySelections = []string{
			bg.AbilityScores.Options[0],
			bg.AbilityScores.Options[1],
			bg.AbilityScores.Options[2],
		}
	}
	
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
	
	// Find next available value
	start := 0
	if currentIdx != -1 {
		if forward {
			start = (currentIdx + 1) % len(m.standardArrayValues)
		} else {
			start = currentIdx - 1
			if start < 0 {
				start = len(m.standardArrayValues) - 1
			}
		}
	}
	
	// Search for next unused value
	for i := 0; i < len(m.standardArrayValues); i++ {
		idx := start
		if forward {
			idx = (start + i) % len(m.standardArrayValues)
		} else {
			idx = start - i
			if idx < 0 {
				idx += len(m.standardArrayValues)
			}
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
	
	// Apply background ability bonuses if any are selected
	for _, abilityName := range m.abilitySelections {
		switch abilityName {
		case "str":
			m.character.AbilityScores.Strength.Base++
		case "dex":
			m.character.AbilityScores.Dexterity.Base++
		case "con":
			m.character.AbilityScores.Constitution.Base++
		case "int":
			m.character.AbilityScores.Intelligence.Base++
		case "wis":
			m.character.AbilityScores.Wisdom.Base++
		case "cha":
			m.character.AbilityScores.Charisma.Base++
		}
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
	content.WriteString(stepStyle.Render("Step 4 of 5: Background Selection"))
	content.WriteString("\n\n")
	
	// Render background list
	m.backgroundList.Width = m.width - 4
	m.backgroundList.Height = m.height - 15
	content.WriteString(m.backgroundList.Render())
	content.WriteString("\n\n")
	
	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	content.WriteString(helpStyle.Render("↑/↓: Navigate | Enter: Next | Esc: Back"))
	
	return content.String()
}

// renderAbilityScores renders the ability score assignment step.
func (m *CharacterCreationModel) renderAbilityScores() string {
	var content strings.Builder
	
	stepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)
	content.WriteString(stepStyle.Render("Step 5 of 5: Ability Score Assignment"))
	content.WriteString("\n\n")
	
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
	
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	content.WriteString(helpStyle.Render("Press 'm' to change mode"))
	content.WriteString("\n\n")
	
	// Ability names
	abilityNames := []string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}
	abilityFullNames := []string{"Strength", "Dexterity", "Constitution", "Intelligence", "Wisdom", "Charisma"}
	
	// Background bonuses
	backgroundBonuses := [6]int{0, 0, 0, 0, 0, 0}
	if m.selectedBackground != nil {
		for _, abilityName := range m.abilitySelections {
			switch abilityName {
			case "str":
				backgroundBonuses[0]++
			case "dex":
				backgroundBonuses[1]++
			case "con":
				backgroundBonuses[2]++
			case "int":
				backgroundBonuses[3]++
			case "wis":
				backgroundBonuses[4]++
			case "cha":
				backgroundBonuses[5]++
			}
		}
	}
	
	// Render ability scores
	for i := 0; i < 6; i++ {
		base := m.abilityScores[i]
		bonus := backgroundBonuses[i]
		final := base + bonus
		modifier := (final - 10) / 2
		modifierStr := fmt.Sprintf("%+d", modifier)
		
		// Style based on focus
		var lineStyle lipgloss.Style
		if i == m.focusedAbility {
			lineStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("13")).
				Bold(true)
		} else {
			lineStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15"))
		}
		
		// Build line
		var line strings.Builder
		if i == m.focusedAbility {
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
		totalBonus := len(m.abilitySelections)
		maxBonus := m.selectedBackground.AbilityScores.Points
		content.WriteString(bonusStyle.Render(fmt.Sprintf("Background Bonuses: %d / %d points allocated", totalBonus, maxBonus)))
		content.WriteString("\n")
		
		// TODO: Add UI for selecting background bonuses
		if totalBonus < maxBonus {
			content.WriteString(helpStyle.Render(fmt.Sprintf("Note: You can allocate %d more ability point(s) from your background", maxBonus-totalBonus)))
			content.WriteString("\n")
		}
	}
	
	content.WriteString("\n")
	content.WriteString(helpStyle.Render("Tab/←/→: Change ability | ↑/↓/+/-: Adjust score | m: Change mode | Enter: Finish | Esc: Back"))
	
	return content.String()
}
