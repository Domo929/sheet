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
	storage           *storage.CharacterStorage
	characters        []storage.CharacterInfo
	list              components.List
	helpFooter        components.HelpFooter
	width             int
	height            int
	err               error
	loading           bool
	confirmingDelete  bool
	deleteTarget      string
	quitting          bool
}

// NewCharacterSelectionModel creates a new character selection model.
func NewCharacterSelectionModel(store *storage.CharacterStorage) *CharacterSelectionModel {
	// Create help footer with appropriate bindings
	helpFooter := components.NewHelpFooter()

	return &CharacterSelectionModel{
		storage:          store,
		list:             components.NewList("Saved Characters", nil),
		helpFooter:       helpFooter,
		loading:          true,
		confirmingDelete: false,
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
		// If confirming delete, handle y/n
		if m.confirmingDelete {
			switch msg.String() {
			case "y", "Y":
				// Confirm delete
				m.confirmingDelete = false
				return m, DeleteCharacterCmd(m.storage, m.deleteTarget)
			case "n", "N", "esc":
				// Cancel delete
				m.confirmingDelete = false
				m.deleteTarget = ""
				return m, nil
			}
			return m, nil
		}
		
		// Normal key handling
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
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

		case "enter":
			// Load selected character
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
			return m, nil
		
		case "n", "N":
			// New character
			return m, func() tea.Msg {
				return StartCharacterCreationMsg{}
			}
		
		case "d", "D":
			// Delete character (with confirmation)
			if len(m.characters) == 0 {
				m.err = fmt.Errorf("no characters available to delete")
				return m, nil
			}
			selected := m.list.Selected()
			if selected != nil {
				m.confirmingDelete = true
				m.deleteTarget = selected.Title
			}
			return m, nil
		}
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
	// Return empty view when quitting to prevent artifacts
	if m.quitting {
		return ""
	}
	
	if m.width == 0 || m.height == 0 {
		return "Initializing terminal display..."
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

	// Delete confirmation prompt
	if m.confirmingDelete {
		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(true)
		content.WriteString(confirmStyle.Render(fmt.Sprintf("Delete character '%s'?", m.deleteTarget)))
		content.WriteString("\n")
		content.WriteString("Press Y to confirm, N to cancel\n\n")
		return content.String()
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
		content.WriteString(emptyStyle.Render("Press 'n' to create a character."))
		content.WriteString("\n\n")
	} else {
		// Render character list
		listHeight := m.height - 10 // Reserve space for title and help
		if listHeight < 5 {
			listHeight = 5
		}
		m.list.Width = m.width - 4
		m.list.Height = listHeight
		content.WriteString(m.list.Render())
		content.WriteString("\n\n")
	}

	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	if len(m.characters) > 0 {
		content.WriteString(helpStyle.Render("↑/↓: Navigate | Enter: Load | n: New | d: Delete | q: Quit"))
	} else {
		content.WriteString(helpStyle.Render("n: New Character | q: Quit"))
	}

	return content.String()
}
