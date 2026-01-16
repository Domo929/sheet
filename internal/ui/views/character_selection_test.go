package views

import (
	"testing"

	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewCharacterSelectionModel(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)

	if model == nil {
		t.Fatal("Expected model to be created")
	}

	if model.storage != store {
		t.Error("Storage not set correctly")
	}

	if !model.loading {
		t.Error("Model should start in loading state")
	}

	if model.confirmingDelete {
		t.Error("Model should not start in delete confirmation state")
	}
}

func TestCharacterSelectionInit(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)

	cmd := model.Init()
	if cmd == nil {
		t.Error("Init should return a command to load character list")
	}

	// Execute the command and check the message
	msg := cmd()
	listMsg, ok := msg.(CharacterListLoadedMsg)
	if !ok {
		t.Fatalf("Expected CharacterListLoadedMsg, got %T", msg)
	}

	if listMsg.Err != nil {
		t.Errorf("Expected no error loading empty directory, got: %v", listMsg.Err)
	}

	if len(listMsg.Characters) != 0 {
		t.Errorf("Expected empty character list, got %d characters", len(listMsg.Characters))
	}
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

	if model.loading {
		t.Error("Loading should be false after receiving character list")
	}

	if len(model.characters) != 1 {
		t.Errorf("Expected 1 character, got %d", len(model.characters))
	}

	if model.characters[0].Name != "Test Hero" {
		t.Errorf("Expected character name 'Test Hero', got '%s'", model.characters[0].Name)
	}
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
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	if model.list.SelectedIndex == initialSelection {
		t.Error("Expected selection to move down")
	}

	// Move up
	prevSelection := model.list.SelectedIndex
	model.Update(tea.KeyMsg{Type: tea.KeyUp})
	if model.list.SelectedIndex == prevSelection {
		t.Error("Expected selection to move up")
	}
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
	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel

	if cmd == nil {
		t.Fatal("Expected command to be returned for loading character")
	}

	// Execute command and check message
	msg := cmd()
	loadMsg, ok := msg.(CharacterLoadedMsg)
	if !ok {
		t.Fatalf("Expected CharacterLoadedMsg, got %T", msg)
	}

	if loadMsg.Path != path {
		t.Errorf("Expected path %s, got %s", path, loadMsg.Path)
	}
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
	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	model = updatedModel

	// Should be in confirmation mode
	if !model.confirmingDelete {
		t.Error("Expected model to be in delete confirmation state")
	}
	if model.deleteTarget != "Test Hero" {
		t.Errorf("Expected delete target 'Test Hero', got '%s'", model.deleteTarget)
	}

	// Press 'y' to confirm
	updatedModel, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	model = updatedModel

	if cmd == nil {
		t.Fatal("Expected command to be returned for deleting character")
	}

	// Execute command
	msg := cmd()
	deleteMsg, ok := msg.(CharacterDeletedMsg)
	if !ok {
		t.Fatalf("Expected CharacterDeletedMsg, got %T", msg)
	}

	if deleteMsg.Name != "Test Hero" {
		t.Errorf("Expected deleted character name 'Test Hero', got '%s'", deleteMsg.Name)
	}
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
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	model = updatedModel

	if !model.confirmingDelete {
		t.Error("Expected model to be in delete confirmation state")
	}

	// Press 'n' to cancel
	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	model = updatedModel

	if model.confirmingDelete {
		t.Error("Expected model to exit delete confirmation state")
	}
	if model.deleteTarget != "" {
		t.Error("Expected delete target to be cleared")
	}
	if cmd != nil {
		t.Error("Expected no command when canceling delete")
	}
}

func TestCharacterSelectionQuitAction(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)
	model.loading = false

	// Press 'q' to quit
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if cmd == nil {
		t.Fatal("Expected quit command to be returned")
	}

	// Verify it's a quit command - tea.Quit() returns a QuitMsg
	msg := cmd()
	if msg == nil {
		t.Error("Expected quit message")
	}
}

func TestCharacterSelectionNewCharacterKey(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)
	model.loading = false

	// Press 'n' to create new character
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	if cmd == nil {
		t.Fatal("Expected command for starting character creation")
	}

	// Execute command and check message
	msg := cmd()
	_, ok := msg.(StartCharacterCreationMsg)
	if !ok {
		t.Fatalf("Expected StartCharacterCreationMsg, got %T", msg)
	}
}

func TestCharacterSelectionQuitKey(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)

	// Press 'q' key
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if cmd == nil {
		t.Fatal("Expected quit command")
	}

	msg := cmd()
	if msg == nil {
		t.Error("Expected quit message")
	}
}

func TestCharacterSelectionWindowResize(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)

	// Send window size message
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updatedModel

	if model.width != 80 {
		t.Errorf("Expected width 80, got %d", model.width)
	}

	if model.height != 24 {
		t.Errorf("Expected height 24, got %d", model.height)
	}
}

func TestCharacterSelectionView(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)
	model.width = 80
	model.height = 24

	// Test loading state
	view := model.View()
	if view == "" {
		t.Error("View should not be empty")
	}

	// Test empty character list
	model.loading = false
	view = model.View()
	if view == "" {
		t.Error("View should not be empty for no characters")
	}

	// Test with characters
	model.characters = []storage.CharacterInfo{
		{Name: "Hero 1", Race: "Human", Class: "Fighter", Level: 1},
	}
	model.updateList()
	view = model.View()
	if view == "" {
		t.Error("View should not be empty with characters")
	}
}

func TestCharacterSelectionErrorHandling(t *testing.T) {
	store, _ := storage.NewCharacterStorage(t.TempDir())
	model := NewCharacterSelectionModel(store)
	model.loading = false

	// Test load with no characters - press enter
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel

	if model.err == nil {
		t.Error("Expected error when loading with no characters")
	}

	// Test delete with no characters - press 'd'
	model.err = nil
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	model = updatedModel

	if model.err == nil {
		t.Error("Expected error when deleting with no characters")
	}
}
