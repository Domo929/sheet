package ui

import (
	"testing"

	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/ui/views"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMainSheetQuitKey(t *testing.T) {
	// Create a model in the main sheet view state
	m, err := NewModel()
	require.NoError(t, err, "Failed to create model")

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
	require.NotNil(t, cmd, "Expected quit command, got nil")

	// Verify it's actually a quit command by checking the message it produces
	msg := cmd()
	_, ok := msg.(tea.QuitMsg)
	assert.True(t, ok, "Expected tea.QuitMsg, got %T", msg)
}

func TestMainSheetCtrlCQuit(t *testing.T) {
	m, err := NewModel()
	require.NoError(t, err, "Failed to create model")

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

	require.NotNil(t, cmd, "Expected quit command, got nil")

	msg := cmd()
	_, ok := msg.(tea.QuitMsg)
	assert.True(t, ok, "Expected tea.QuitMsg, got %T", msg)
}

func TestCancelCharacterCreation(t *testing.T) {
	m, err := NewModel()
	require.NoError(t, err, "Failed to create model")

	// Start character creation
	m.currentView = ViewCharacterCreation
	m.characterCreationModel = &views.CharacterCreationModel{}

	// Send cancel message
	updatedModel, _ := m.Update(views.CancelCharacterCreationMsg{})
	m = updatedModel.(Model)

	// Should return to character selection
	assert.Equal(t, ViewCharacterSelection, m.currentView, "Expected ViewCharacterSelection after cancel")

	// Creation model should be cleared
	assert.Nil(t, m.characterCreationModel, "Expected characterCreationModel to be nil after cancel")
}
