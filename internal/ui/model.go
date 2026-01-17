package ui

import (
	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/Domo929/sheet/internal/ui/views"
	tea "github.com/charmbracelet/bubbletea"
)

// ViewType represents the different views in the application.
type ViewType int

const (
	ViewCharacterSelection ViewType = iota
	ViewCharacterCreation
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
	currentView ViewType
	width       int
	height      int
	character   *models.Character
	storage     *storage.CharacterStorage
	loader      *data.Loader
	err         error
	quitting    bool // Track if we're in the process of quitting

	// View-specific models
	characterSelectionModel *views.CharacterSelectionModel
	characterCreationModel  *views.CharacterCreationModel
	mainSheetModel          *views.MainSheetModel
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

	// Initialize data loader
	loader := data.NewLoader("./data")

	// Initialize character selection model
	charSelectionModel := views.NewCharacterSelectionModel(store)

	return Model{
		currentView:             ViewCharacterSelection,
		storage:                 store,
		loader:                  loader,
		characterSelectionModel: charSelectionModel,
	}, nil
}

// Init initializes the model (required by Bubble Tea).
func (m Model) Init() tea.Cmd {
	if m.characterSelectionModel != nil {
		return m.characterSelectionModel.Init()
	}
	return nil
}

// Update handles messages and updates the model (required by Bubble Tea).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Forward to current view
		return m.updateCurrentView(msg)

	case views.StartCharacterCreationMsg:
		// Create and initialize character creation model
		m.characterCreationModel = views.NewCharacterCreationModel(m.storage, m.loader)
		// Pass current window size to the new model
		if m.width > 0 && m.height > 0 {
			m.characterCreationModel, _ = m.characterCreationModel.Update(tea.WindowSizeMsg{
				Width:  m.width,
				Height: m.height,
			})
		}
		m.currentView = ViewCharacterCreation
		return m, m.characterCreationModel.Init()

	case views.CharacterCreatedMsg:
		// Character created successfully, return to selection screen
		m.currentView = ViewCharacterSelection
		m.characterCreationModel = nil
		// Reload character list
		if m.characterSelectionModel != nil {
			return m, m.characterSelectionModel.Init()
		}
		return m, nil

	case views.CancelCharacterCreationMsg:
		// User cancelled character creation, return to selection screen
		m.currentView = ViewCharacterSelection
		m.characterCreationModel = nil
		return m, nil

	case views.BackToSelectionMsg:
		// User wants to return to character selection from main sheet
		m.currentView = ViewCharacterSelection
		m.mainSheetModel = nil
		m.character = nil
		// Reload character list
		if m.characterSelectionModel != nil {
			return m, m.characterSelectionModel.Init()
		}
		return m, nil

	case views.CharacterLoadedMsg:
		// Load the character from storage
		char, err := m.storage.LoadByPath(msg.Path)
		if err != nil {
			m.err = err
			return m, nil
		}
		m.character = char
		m.mainSheetModel = views.NewMainSheetModel(char, m.storage)
		// Pass current window size to the new model
		if m.width > 0 && m.height > 0 {
			m.mainSheetModel, _ = m.mainSheetModel.Update(tea.WindowSizeMsg{
				Width:  m.width,
				Height: m.height,
			})
		}
		m.currentView = ViewMainSheet
		return m, m.mainSheetModel.Init()

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
		m.mainSheetModel = views.NewMainSheetModel(char, m.storage)
		// Pass current window size to the new model
		if m.width > 0 && m.height > 0 {
			m.mainSheetModel, _ = m.mainSheetModel.Update(tea.WindowSizeMsg{
				Width:  m.width,
				Height: m.height,
			})
		}
		m.currentView = ViewMainSheet
		return m, m.mainSheetModel.Init()

	case ErrorMsg:
		m.err = msg.Err
		return m, nil
	}

	// Route to appropriate view handler
	return m.updateCurrentView(msg)
}

// updateCurrentView routes the message to the current view's update function.
func (m Model) updateCurrentView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.currentView {
	case ViewCharacterSelection:
		if m.characterSelectionModel != nil {
			updatedModel, c := m.characterSelectionModel.Update(msg)
			m.characterSelectionModel = updatedModel
			cmd = c
		}
	case ViewCharacterCreation:
		if m.characterCreationModel != nil {
			updatedModel, c := m.characterCreationModel.Update(msg)
			m.characterCreationModel = updatedModel
			cmd = c
		}
	case ViewMainSheet:
		if m.mainSheetModel != nil {
			updatedModel, c := m.mainSheetModel.Update(msg)
			m.mainSheetModel = updatedModel
			cmd = c

			// Check for navigation back to selection
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				if keyMsg.String() == "q" || keyMsg.String() == "ctrl+c" {
					m.quitting = true
					return m, tea.Quit
				}
			}
		}
	default:
		// Other views not yet implemented
	}

	return m, cmd
}

// View renders the current view (required by Bubble Tea).
func (m Model) View() string {
	// Return empty view when quitting to avoid flashing content
	if m.quitting {
		return ""
	}

	if m.err != nil {
		return "Error: " + m.err.Error() + "\n\nPress q to quit."
	}

	// Route to appropriate view renderer
	switch m.currentView {
	case ViewCharacterSelection:
		return m.renderCharacterSelection()
	case ViewCharacterCreation:
		return m.renderCharacterCreation()
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
	if m.characterSelectionModel != nil {
		return m.characterSelectionModel.View()
	}
	return "Character Selection View (TODO)\n\nPress q to quit."
}

func (m Model) renderCharacterCreation() string {
	if m.characterCreationModel != nil {
		return m.characterCreationModel.View()
	}
	return "Character Creation View (TODO)\n\nPress q to quit."
}

func (m Model) renderMainSheet() string {
	if m.mainSheetModel != nil {
		return m.mainSheetModel.View()
	}
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
