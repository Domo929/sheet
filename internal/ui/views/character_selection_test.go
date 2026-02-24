package views

import (
	"testing"

	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCharacterSelectionModel(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)

	require.NotNil(t, model, "Expected model to be created")
	assert.Equal(t, store, model.storage, "Storage not set correctly")
	assert.True(t, model.loading, "Model should start in loading state")
	assert.False(t, model.confirmingDelete, "Model should not start in delete confirmation state")
}

func TestCharacterSelectionInit(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)

	cmd := model.Init()
	require.NotNil(t, cmd, "Init should return a command to load character list")

	// Execute the command and check the message
	msg := cmd()
	listMsg, ok := msg.(CharacterListLoadedMsg)
	require.True(t, ok, "Expected CharacterListLoadedMsg, got %T", msg)

	assert.NoError(t, listMsg.Err, "Expected no error loading empty directory")
	assert.Empty(t, listMsg.Characters, "Expected empty character list")
}

func TestCharacterListLoadedMessage(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)

	// Create a test character
	char := &models.Character{
		Info: models.CharacterInfo{
			Name:  "Test Hero",
			Race:  "Human",
			Class: "Fighter",
			Level: 1,
		},
	}
	store.Save(char)

	// Simulate loading the character list
	msg := CharacterListLoadedMsg{
		Characters: []storage.CharacterInfo{
			{
				Name:  "Test Hero",
				Race:  "Human",
				Class: "Fighter",
				Level: 1,
			},
		},
		Err: nil,
	}

	updatedModel, _ := model.Update(msg)
	model = updatedModel

	assert.False(t, model.loading, "Loading should be false after receiving character list")
	assert.Len(t, model.characters, 1, "Expected 1 character")
	assert.Equal(t, "Test Hero", model.characters[0].Name, "Expected character name 'Test Hero'")
}

func TestCharacterSelectionNavigation(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)

	// Add some test characters
	model.characters = []storage.CharacterInfo{
		{Name: "Hero 1", Race: "Human", Class: "Fighter", Level: 1},
		{Name: "Hero 2", Race: "Elf", Class: "Wizard", Level: 2},
		{Name: "Hero 3", Race: "Dwarf", Class: "Cleric", Level: 3},
	}
	model.loading = false
	model.updateList()

	// Test list navigation
	initialSelection := model.list.SelectedIndex

	// Move down
	model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.NotEqual(t, initialSelection, model.list.SelectedIndex, "Expected selection to move down")

	// Move up
	prevSelection := model.list.SelectedIndex
	model.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	assert.NotEqual(t, prevSelection, model.list.SelectedIndex, "Expected selection to move up")
}

func TestCharacterSelectionLoadAction(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)

	// Create and save a test character
	char := &models.Character{
		Info: models.CharacterInfo{
			Name:  "Test Hero",
			Race:  "Human",
			Class: "Fighter",
			Level: 1,
		},
	}
	path, _ := store.Save(char)

	// Set up model with character list
	model.characters = []storage.CharacterInfo{
		{Name: "Test Hero", Race: "Human", Class: "Fighter", Level: 1, Path: path},
	}
	model.loading = false
	model.updateList()

	// Press enter to load selected character
	updatedModel, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model = updatedModel

	require.NotNil(t, cmd, "Expected command to be returned for loading character")

	// Execute command and check message
	msg := cmd()
	loadMsg, ok := msg.(CharacterLoadedMsg)
	require.True(t, ok, "Expected CharacterLoadedMsg, got %T", msg)

	assert.Equal(t, path, loadMsg.Path, "Expected path %s", path)
}

func TestCharacterSelectionDeleteAction(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)

	// Create and save a test character
	char := &models.Character{
		Info: models.CharacterInfo{
			Name:  "Test Hero",
			Race:  "Human",
			Class: "Fighter",
			Level: 1,
		},
	}
	store.Save(char)

	// Set up model with character list
	model.characters = []storage.CharacterInfo{
		{Name: "Test Hero", Race: "Human", Class: "Fighter", Level: 1},
	}
	model.loading = false
	model.updateList()

	// Press 'd' to initiate delete
	updatedModel, cmd := model.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
	model = updatedModel

	// Should be in confirmation mode
	assert.True(t, model.confirmingDelete, "Expected model to be in delete confirmation state")
	assert.Equal(t, "Test Hero", model.deleteTarget, "Expected delete target 'Test Hero'")

	// Press 'y' to confirm
	updatedModel, cmd = model.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	model = updatedModel

	require.NotNil(t, cmd, "Expected command to be returned for deleting character")

	// Execute command
	msg := cmd()
	deleteMsg, ok := msg.(CharacterDeletedMsg)
	require.True(t, ok, "Expected CharacterDeletedMsg, got %T", msg)

	assert.Equal(t, "Test Hero", deleteMsg.Name, "Expected deleted character name 'Test Hero'")
}

func TestCharacterSelectionCancelDelete(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)

	char := &models.Character{
		Info: models.CharacterInfo{
			Name:  "Test Hero",
			Race:  "Human",
			Class: "Fighter",
			Level: 1,
		},
	}
	store.Save(char)

	model.characters = []storage.CharacterInfo{
		{Name: "Test Hero", Race: "Human", Class: "Fighter", Level: 1},
	}
	model.loading = false
	model.updateList()

	// Press 'd' to initiate delete
	updatedModel, _ := model.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
	model = updatedModel

	assert.True(t, model.confirmingDelete, "Expected model to be in delete confirmation state")

	// Press 'n' to cancel
	updatedModel, cmd := model.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	model = updatedModel

	assert.False(t, model.confirmingDelete, "Expected model to exit delete confirmation state")
	assert.Empty(t, model.deleteTarget, "Expected delete target to be cleared")
	assert.Nil(t, cmd, "Expected no command when canceling delete")
}

func TestCharacterSelectionQuitAction(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)
	model.loading = false

	// Press 'q' to initiate quit - should show confirmation
	updatedModel, cmd := model.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})

	assert.Nil(t, cmd, "Expected no command yet, should be confirming")
	require.True(t, updatedModel.confirmingQuit, "Expected to be in quit confirmation mode")

	// Press 'y' to confirm quit
	_, cmd = updatedModel.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})

	require.NotNil(t, cmd, "Expected quit command to be returned")

	// Verify it's a quit command - tea.Quit() returns a QuitMsg
	msg := cmd()
	assert.NotNil(t, msg, "Expected quit message")
}

func TestCharacterSelectionNewCharacterKey(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)
	model.loading = false

	// Press 'n' to create new character
	_, cmd := model.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})

	require.NotNil(t, cmd, "Expected command for starting character creation")

	// Execute command and check message
	msg := cmd()
	_, ok := msg.(StartCharacterCreationMsg)
	assert.True(t, ok, "Expected StartCharacterCreationMsg, got %T", msg)
}

func TestCharacterSelectionQuitKey(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)

	// Press 'q' key - should show confirmation
	updatedModel, cmd := model.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})

	assert.Nil(t, cmd, "Expected no command yet, should be confirming")
	require.True(t, updatedModel.confirmingQuit, "Expected to be in quit confirmation mode")

	// Press 'y' to confirm quit
	_, cmd = updatedModel.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})

	require.NotNil(t, cmd, "Expected quit command")

	msg := cmd()
	assert.NotNil(t, msg, "Expected quit message")
}

func TestCharacterSelectionWindowResize(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)

	// Send window size message
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updatedModel

	assert.Equal(t, 80, model.width, "Expected width 80")
	assert.Equal(t, 24, model.height, "Expected height 24")
}

func TestCharacterSelectionView(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)
	model.width = 80
	model.height = 24

	// Test loading state
	view := model.View()
	assert.NotEmpty(t, view, "View should not be empty")

	// Test empty character list
	model.loading = false
	view = model.View()
	assert.NotEmpty(t, view, "View should not be empty for no characters")

	// Test with characters
	model.characters = []storage.CharacterInfo{
		{Name: "Hero 1", Race: "Human", Class: "Fighter", Level: 1},
	}
	model.updateList()
	view = model.View()
	assert.NotEmpty(t, view, "View should not be empty with characters")
}

func TestCharacterSelectionErrorHandling(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)
	model.loading = false

	// Test load with no characters - press enter
	updatedModel, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model = updatedModel

	assert.NotNil(t, model.err, "Expected error when loading with no characters")

	// Test delete with no characters - press 'd'
	model.err = nil
	updatedModel, _ = model.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
	model = updatedModel

	assert.NotNil(t, model.err, "Expected error when deleting with no characters")
}
