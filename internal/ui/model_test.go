package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Domo929/sheet/internal/models"
)

func TestMainSheetQuitKey(t *testing.T) {
	// Create a model in the main sheet view state
	m, err := NewModel()
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	
	// Set up as if a character was loaded
	m.character = &models.Character{
		Info: models.CharacterInfo{
			Name: "Test Character",
		},
	}
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
	
	m.character = &models.Character{
		Info: models.CharacterInfo{
			Name: "Test Character",
		},
	}
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
