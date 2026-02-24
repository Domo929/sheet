package ui

import (
	"testing"

	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/ui/views"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tea "charm.land/bubbletea/v2"
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
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})

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
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})

	require.NotNil(t, cmd, "Expected quit command, got nil")

	msg := cmd()
	_, ok := msg.(tea.QuitMsg)
	assert.True(t, ok, "Expected tea.QuitMsg, got %T", msg)
}

func TestModelViewTooSmallTerminal(t *testing.T) {
	// Create a minimal model for testing
	m := Model{
		width:  50,
		height: 15,
	}

	v := m.View()
	assert.Contains(t, v.Content, "Terminal too small", "Should show too-small message for 50x15")
	assert.Contains(t, v.Content, "60", "Should mention minimum width")
	assert.Contains(t, v.Content, "20", "Should mention minimum height")
}

func TestModelViewMinimumSizeOK(t *testing.T) {
	m := Model{
		width:  60,
		height: 20,
	}

	v := m.View()
	assert.NotContains(t, v.Content, "Terminal too small", "Should NOT show too-small message at 60x20")
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
