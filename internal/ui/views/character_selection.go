package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/Domo929/sheet/internal/ui/components"
)

// CharacterSelectionModel manages the character selection screen.
type CharacterSelectionModel struct {
	storage     *storage.CharacterStorage
	characters  []storage.CharacterInfo
	list        components.List
	buttons     components.ButtonGroup
	helpFooter  components.HelpFooter
	width       int
	height      int
	err         error
	loading     bool
}

// NewCharacterSelectionModel creates a new character selection model.
func NewCharacterSelectionModel(store *storage.CharacterStorage) *CharacterSelectionModel {
	// Create buttons for actions
	buttons := components.NewButtonGroup("Load", "New", "Delete", "Quit")

	// Create help footer with appropriate bindings
	helpFooter := components.NewHelpFooter()

	return &CharacterSelectionModel{
		storage:    store,
		list:       components.NewList("Saved Characters", nil),
		buttons:    buttons,
		helpFooter: helpFooter,
		loading:    true,
	}
}

// CharacterLoadedMsg is sent when a character is successfully loaded.
type CharacterLoadedMsg struct {
	Path string
}

// StartCharacterCreationMsg is sent when user wants to create a new character.
type StartCharacterCreationMsg struct{}

// CharacterListLoadedMsg is sent when the character list is loaded.
type CharacterListLoadedMsg struct {
	Characters []storage.CharacterInfo
	Err        error
}

// CharacterDeletedMsg is sent when a character is deleted.
type CharacterDeletedMsg struct {
	Name string
	Err  error
}

// LoadCharacterListCmd loads the list of saved characters.
func LoadCharacterListCmd(store *storage.CharacterStorage) tea.Cmd {
	return func() tea.Msg {
		chars, err := store.List()
		return CharacterListLoadedMsg{
			Characters: chars,
			Err:        err,
		}
	}
}

// LoadCharacterCmd loads a specific character by path.
func LoadCharacterCmd(path string) tea.Cmd {
	return func() tea.Msg {
		return CharacterLoadedMsg{Path: path}
	}
}

// DeleteCharacterCmd deletes a character.
func DeleteCharacterCmd(store *storage.CharacterStorage, name string) tea.Cmd {
	return func() tea.Msg {
		err := store.Delete(name)
		return CharacterDeletedMsg{
			Name: name,
			Err:  err,
		}
	}
}

// Init initializes the character selection model.
func (m *CharacterSelectionModel) Init() tea.Cmd {
	return LoadCharacterListCmd(m.storage)
}

// Update handles messages for the character selection screen.
func (m *CharacterSelectionModel) Update(msg tea.Msg) (*CharacterSelectionModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case CharacterListLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.characters = msg.Characters
		m.updateList()
		return m, nil

	case CharacterDeletedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Reload character list after deletion
		return m, LoadCharacterListCmd(m.storage)

	case tea.KeyMsg:
		// Handle navigation keys
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if !m.loading && len(m.characters) > 0 {
				m.list.MoveUp()
			}
			return m, nil

		case "down", "j":
			if !m.loading && len(m.characters) > 0 {
				m.list.MoveDown()
			}
			return m, nil

		case "left", "h":
			m.buttons.MoveLeft()
			return m, nil

		case "right", "l":
			m.buttons.MoveRight()
			return m, nil

		case "enter", " ":
			return m.handleAction()

		case "tab":
			m.buttons.MoveRight()
			return m, nil

		case "shift+tab":
			m.buttons.MoveLeft()
			return m, nil
		}
	}

	return m, nil
}

// handleAction performs the selected action.
func (m *CharacterSelectionModel) handleAction() (*CharacterSelectionModel, tea.Cmd) {
	if m.loading {
		return m, nil
	}

	selectedIndex := m.buttons.SelectedIndex
	if selectedIndex < 0 || selectedIndex >= len(m.buttons.Buttons) {
		return m, nil
	}

	action := m.buttons.Buttons[selectedIndex].Label

	switch action {
	case "Load":
		if len(m.characters) == 0 {
			m.err = fmt.Errorf("no characters available to load")
			return m, nil
		}
		selected := m.list.Selected()
		if selected != nil {
			if path, ok := selected.Value.(string); ok {
				return m, LoadCharacterCmd(path)
			}
		}

	case "New":
		// Navigate to character creation view
		return m, func() tea.Msg {
			return StartCharacterCreationMsg{}
		}

	case "Delete":
		if len(m.characters) == 0 {
			m.err = fmt.Errorf("no characters available to delete")
			return m, nil
		}
		selected := m.list.Selected()
		if selected != nil {
			return m, DeleteCharacterCmd(m.storage, selected.Title)
		}

	case "Quit":
		return m, tea.Quit
	}

	return m, nil
}

// updateList updates the list component with current characters.
func (m *CharacterSelectionModel) updateList() {
	items := make([]components.ListItem, len(m.characters))
	for i, char := range m.characters {
		items[i] = components.ListItem{
			Title:       char.Name,
			Description: fmt.Sprintf("Level %d %s %s", char.Level, char.Race, char.Class),
			Value:       char.Path,
		}
	}
	m.list = components.NewList("Saved Characters", items)
}

// View renders the character selection screen.
func (m *CharacterSelectionModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Padding(1, 0)
	title := titleStyle.Render("D&D 5e Character Sheet")
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

	// Loading state
	if m.loading {
		content.WriteString("Loading characters...\n")
	} else if len(m.characters) == 0 {
		// No characters available
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true)
		content.WriteString(emptyStyle.Render("No saved characters found."))
		content.WriteString("\n")
		content.WriteString(emptyStyle.Render("Press 'New' to create a character."))
		content.WriteString("\n\n")
	} else {
		// Render character list
		listHeight := m.height - 12 // Reserve space for title, buttons, help
		if listHeight < 5 {
			listHeight = 5
		}
		m.list.Width = m.width - 4
		m.list.Height = listHeight
		content.WriteString(m.list.Render())
		content.WriteString("\n\n")
	}

	// Render action buttons
	content.WriteString(m.buttons.Render())
	content.WriteString("\n\n")

	// Render help footer
	content.WriteString(m.helpFooter.Render())

	return content.String()
}
