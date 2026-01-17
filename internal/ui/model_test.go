package ui

import (
	"testing"

	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/ui/views"
	tea "github.com/charmbracelet/bubbletea"
)

func TestMainSheetQuitKey(t *testing.T) {
	// Create a model in the main sheet view state
	m, err := NewModel()
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Set up as if a character was loaded
	char := &models.Character{
		Info: models.CharacterInfo{
			Name: "Test Character",
		},
	}
	m.character = char
	m.mainSheetModel = views.NewMainSheetModel(char, m.storage)
	m.currentView = ViewMainSheet

	// Send 'q' key
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// Should return tea.Quit command
	if cmd == nil {
		t.Error("Expected quit command, got nil")
	}

	// Verify it's actually a quit command by checking the message it produces
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("Expected tea.QuitMsg, got %T", msg)
	}
}

func TestMainSheetCtrlCQuit(t *testing.T) {
	m, err := NewModel()
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	char := &models.Character{
		Info: models.CharacterInfo{
			Name: "Test Character",
		},
	}
	m.character = char
	m.mainSheetModel = views.NewMainSheetModel(char, m.storage)
	m.currentView = ViewMainSheet

	// Send Ctrl+C
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Error("Expected quit command, got nil")
	}

	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("Expected tea.QuitMsg, got %T", msg)
	}
}

func TestCancelCharacterCreation(t *testing.T) {
	m, err := NewModel()
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Start character creation
	m.currentView = ViewCharacterCreation
	m.characterCreationModel = &views.CharacterCreationModel{}

	// Send cancel message
	updatedModel, _ := m.Update(views.CancelCharacterCreationMsg{})
	m = updatedModel.(Model)

	// Should return to character selection
	if m.currentView != ViewCharacterSelection {
		t.Errorf("Expected ViewCharacterSelection after cancel, got %v", m.currentView)
	}

	// Creation model should be cleared
	if m.characterCreationModel != nil {
		t.Error("Expected characterCreationModel to be nil after cancel")
	}
}
