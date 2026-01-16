package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
)

// ViewType represents the different views in the application.
type ViewType int

const (
	ViewCharacterSelection ViewType = iota
	ViewMainSheet
	ViewInventory
	ViewSpellbook
	ViewCharacterInfo
	ViewLevelUp
	ViewCombat
	ViewRest
)

// Model is the main application model that manages view routing.
type Model struct {
	currentView  ViewType
	width        int
	height       int
	character    *models.Character
	storage      *storage.CharacterStorage
	err          error
	
	// View-specific models (will be implemented in future phases)
	// characterSelectionModel interface{}
	// mainSheetModel interface{}
	// inventoryModel interface{}
	// etc.
}

// NewModel creates a new application model.
func NewModel() (Model, error) {
	// Initialize storage with default directory
	store, err := storage.NewCharacterStorage("")
	if err != nil {
		return Model{}, err
	}

	return Model{
		currentView: ViewCharacterSelection,
		storage:     store,
	}, nil
}

// Init initializes the model (required by Bubble Tea).
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model (required by Bubble Tea).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case NavigateMsg:
		m.currentView = msg.View
		return m, nil

	case LoadCharacterMsg:
		char, err := m.storage.Load(msg.CharacterName)
		if err != nil {
			m.err = err
			return m, nil
		}
		m.character = char
		m.currentView = ViewMainSheet
		return m, nil

	case ErrorMsg:
		m.err = msg.Err
		return m, nil
	}

	// Route to appropriate view handler
	return m.updateCurrentView(msg)
}

// updateCurrentView routes the message to the current view's update function.
func (m Model) updateCurrentView(msg tea.Msg) (tea.Model, tea.Cmd) {
	// This will be expanded as we add view-specific models
	switch m.currentView {
	case ViewCharacterSelection:
		// Will be implemented in character selection phase
		return m, nil
	case ViewMainSheet:
		// Will be implemented in main sheet phase
		return m, nil
	default:
		return m, nil
	}
}

// View renders the current view (required by Bubble Tea).
func (m Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n\nPress q to quit."
	}

	// Route to appropriate view renderer
	switch m.currentView {
	case ViewCharacterSelection:
		return m.renderCharacterSelection()
	case ViewMainSheet:
		return m.renderMainSheet()
	case ViewInventory:
		return m.renderInventory()
	case ViewSpellbook:
		return m.renderSpellbook()
	case ViewCharacterInfo:
		return m.renderCharacterInfo()
	case ViewLevelUp:
		return m.renderLevelUp()
	case ViewCombat:
		return m.renderCombat()
	case ViewRest:
		return m.renderRest()
	default:
		return "Unknown view"
	}
}

// View rendering stubs (will be implemented in future phases)
func (m Model) renderCharacterSelection() string {
	return "Character Selection View (TODO)\n\nPress q to quit."
}

func (m Model) renderMainSheet() string {
	if m.character == nil {
		return "No character loaded"
	}
	return "Main Character Sheet View (TODO)\n\nCharacter: " + m.character.Info.Name + "\nPress q to quit."
}

func (m Model) renderInventory() string {
	return "Inventory View (TODO)\n\nPress q to quit."
}

func (m Model) renderSpellbook() string {
	return "Spellbook View (TODO)\n\nPress q to quit."
}

func (m Model) renderCharacterInfo() string {
	return "Character Info View (TODO)\n\nPress q to quit."
}

func (m Model) renderLevelUp() string {
	return "Level Up View (TODO)\n\nPress q to quit."
}

func (m Model) renderCombat() string {
	return "Combat View (TODO)\n\nPress q to quit."
}

func (m Model) renderRest() string {
	return "Rest View (TODO)\n\nPress q to quit."
}
