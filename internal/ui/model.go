package ui

import (
	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/Domo929/sheet/internal/ui/components"
	"github.com/Domo929/sheet/internal/ui/views"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	ViewNotes
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
	inventoryModel          *views.InventoryModel
	spellbookModel          *views.SpellbookModel
	levelUpModel            *views.LevelUpModel
	notesEditorModel        *views.NotesEditorModel
	characterInfoModel      *views.CharacterInfoModel

	// Roll engine and history (shared across views)
	rollEngine  *components.RollEngine
	rollHistory *components.RollHistory
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
		rollEngine:              components.NewRollEngine(),
		rollHistory:             components.NewRollHistory(),
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
	// If roll engine is active, it handles all keys
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if m.rollEngine != nil && m.rollEngine.IsActive() {
			cmd := m.rollEngine.Update(keyMsg)
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update main sheet's roll history layout state
		if m.mainSheetModel != nil && m.rollHistory != nil {
			historyWidth := 0
			if m.rollHistory.Visible && m.width >= 80 {
				historyWidth = 27
			}
			m.mainSheetModel.SetRollHistoryState(m.rollHistory.Visible && m.width >= 80, historyWidth)
		}
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

	case views.BackToSheetMsg:
		// Return from inventory/spellbook/level-up to main sheet
		m.currentView = ViewMainSheet
		m.inventoryModel = nil
		m.spellbookModel = nil
		m.levelUpModel = nil
		m.notesEditorModel = nil
		m.characterInfoModel = nil
		return m, nil

	case views.OpenInventoryMsg:
		// Navigate to inventory view
		m.inventoryModel = views.NewInventoryModel(m.character, m.storage)
		if m.width > 0 && m.height > 0 {
			m.inventoryModel, _ = m.inventoryModel.Update(tea.WindowSizeMsg{
				Width:  m.width,
				Height: m.height,
			})
		}
		m.currentView = ViewInventory
		return m, m.inventoryModel.Init()

	case views.OpenSpellbookMsg:
		// Navigate to spellbook view
		m.spellbookModel = views.NewSpellbookModel(m.character, m.storage, m.loader)
		if m.width > 0 && m.height > 0 {
			m.spellbookModel, _ = m.spellbookModel.Update(tea.WindowSizeMsg{
				Width:  m.width,
				Height: m.height,
			})
		}
		m.currentView = ViewSpellbook
		return m, m.spellbookModel.Init()

	case views.OpenLevelUpMsg:
		// Navigate to level-up wizard
		m.levelUpModel = views.NewLevelUpModel(m.character, m.storage, m.loader)
		if m.width > 0 && m.height > 0 {
			m.levelUpModel, _ = m.levelUpModel.Update(tea.WindowSizeMsg{
				Width:  m.width,
				Height: m.height,
			})
		}
		m.currentView = ViewLevelUp
		return m, m.levelUpModel.Init()

	case views.OpenNotesMsg:
		m.notesEditorModel = views.NewNotesEditorModel(m.character, m.storage, msg.ReturnTo)
		if m.width > 0 && m.height > 0 {
			m.notesEditorModel, _ = m.notesEditorModel.Update(tea.WindowSizeMsg{
				Width:  m.width,
				Height: m.height,
			})
		}
		m.currentView = ViewNotes
		return m, m.notesEditorModel.Init()

	case views.OpenCharacterInfoMsg:
		m.characterInfoModel = views.NewCharacterInfoModel(m.character, m.storage)
		if m.width > 0 && m.height > 0 {
			m.characterInfoModel, _ = m.characterInfoModel.Update(tea.WindowSizeMsg{
				Width:  m.width,
				Height: m.height,
			})
		}
		m.currentView = ViewCharacterInfo
		return m, m.characterInfoModel.Init()

	case views.BackToCharacterInfoMsg:
		m.notesEditorModel = nil
		m.currentView = ViewCharacterInfo
		return m, nil

	case views.LevelUpCompleteMsg:
		// Return to main sheet after level-up, rebuild to reflect new stats
		m.currentView = ViewMainSheet
		m.levelUpModel = nil
		m.mainSheetModel = views.NewMainSheetModel(m.character, m.storage)
		if m.width > 0 && m.height > 0 {
			m.mainSheetModel, _ = m.mainSheetModel.Update(tea.WindowSizeMsg{
				Width:  m.width,
				Height: m.height,
			})
		}
		return m, m.mainSheetModel.Init()

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

	case components.RequestRollMsg:
		// Forward to roll engine
		cmd := m.rollEngine.Update(msg)
		return m, cmd

	case components.RollCompleteMsg:
		// Roll finished — add to history, then let views handle
		if m.rollHistory != nil {
			m.rollHistory.Add(msg.Entry)
		}
		return m.updateCurrentView(msg)

	case components.RollTickMsg:
		// Animation tick — forward to roll engine
		cmd := m.rollEngine.Update(msg)
		return m, cmd

	case components.OpenCustomRollMsg:
		if m.rollEngine != nil {
			m.rollEngine.OpenCustomRoll()
		}
		return m, nil

	case components.ToggleRollHistoryMsg:
		if m.rollHistory != nil {
			m.rollHistory.Toggle()
			// Update main sheet's layout to accommodate history column
			if m.mainSheetModel != nil {
				historyWidth := 0
				if m.rollHistory.Visible && m.width >= 80 {
					historyWidth = 27
				}
				m.mainSheetModel.SetRollHistoryState(m.rollHistory.Visible && m.width >= 80, historyWidth)
			}
		}
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
	case ViewInventory:
		if m.inventoryModel != nil {
			updatedModel, c := m.inventoryModel.Update(msg)
			m.inventoryModel = updatedModel
			cmd = c
		}
	case ViewSpellbook:
		if m.spellbookModel != nil {
			updatedModel, c := m.spellbookModel.Update(msg)
			m.spellbookModel = updatedModel
			cmd = c
		}
	case ViewLevelUp:
		if m.levelUpModel != nil {
			updatedModel, c := m.levelUpModel.Update(msg)
			m.levelUpModel = updatedModel
			cmd = c
		}
	case ViewNotes:
		if m.notesEditorModel != nil {
			updatedModel, c := m.notesEditorModel.Update(msg)
			m.notesEditorModel = updatedModel
			cmd = c
		}
	case ViewCharacterInfo:
		if m.characterInfoModel != nil {
			updatedModel, c := m.characterInfoModel.Update(msg)
			m.characterInfoModel = updatedModel
			cmd = c
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
		return m.compositeWithRollUI(m.renderMainSheet())
	case ViewInventory:
		return m.renderInventory()
	case ViewSpellbook:
		return m.compositeWithRollUI(m.renderSpellbook())
	case ViewCharacterInfo:
		return m.renderCharacterInfo()
	case ViewLevelUp:
		return m.renderLevelUp()
	case ViewCombat:
		return m.renderCombat()
	case ViewRest:
		return m.renderRest()
	case ViewNotes:
		return m.renderNotes()
	default:
		return "Unknown view"
	}
}

// compositeWithRollUI adds the roll history column and roll engine overlay to view content.
func (m Model) compositeWithRollUI(viewContent string) string {
	// Add roll history column if visible
	if m.rollHistory != nil && m.rollHistory.Visible && m.width >= 80 {
		historyWidth := 27
		historyCol := m.rollHistory.Render(historyWidth, m.height)
		if historyCol != "" {
			viewContent = lipgloss.JoinHorizontal(lipgloss.Top, viewContent, historyCol)
		}
	}

	// Overlay roll engine modal if active
	if m.rollEngine != nil && m.rollEngine.IsActive() {
		return m.rollEngine.View(m.width, m.height)
	}

	return viewContent
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
	if m.inventoryModel != nil {
		return m.inventoryModel.View()
	}
	return "Inventory View (loading...)"
}

func (m Model) renderSpellbook() string {
	if m.spellbookModel != nil {
		return m.spellbookModel.View()
	}
	return "Spellbook View (loading...)"
}

func (m Model) renderCharacterInfo() string {
	if m.characterInfoModel != nil {
		return m.characterInfoModel.View()
	}
	return "Character Info View (loading...)"
}

func (m Model) renderLevelUp() string {
	if m.levelUpModel != nil {
		return m.levelUpModel.View()
	}
	return "Level Up View (loading...)"
}

func (m Model) renderCombat() string {
	return "Combat View (TODO)\n\nPress q to quit."
}

func (m Model) renderRest() string {
	return "Rest View (TODO)\n\nPress q to quit."
}

func (m Model) renderNotes() string {
	if m.notesEditorModel != nil {
		return m.notesEditorModel.View()
	}
	return "Notes View (loading...)"
}
