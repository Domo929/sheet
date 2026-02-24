package views

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// OpenNotesMsg signals to open the notes editor view.
type OpenNotesMsg struct {
	ReturnTo string // "sheet" or "charinfo"
}

// BackToCharacterInfoMsg signals to return to the character info view.
type BackToCharacterInfoMsg struct{}

// NotesMode represents the current mode of the notes editor.
type NotesMode int

const (
	NotesModeList   NotesMode = iota // Document list
	NotesModeEditor                  // Text editor (Task 4)
)

// NotesSortOrder represents how notes are sorted.
type NotesSortOrder int

const (
	NotesSortLastEdited   NotesSortOrder = iota // Most recently edited first
	NotesSortAlphabetical                       // Alphabetical by title
)

// NotesEditorModel is the model for the notes editor view.
type NotesEditorModel struct {
	character *models.Character
	storage   *storage.CharacterStorage
	width     int
	height    int
	keys      notesKeyMap
	returnTo  string // "sheet" or "charinfo"

	// Mode
	mode NotesMode

	// List mode state
	listCursor int
	sortOrder  NotesSortOrder
	sortedDocs []int // indices into character.Personality.Documents

	// Input overlay (new note / rename)
	inputMode   bool
	inputBuffer string
	inputAction string // "new" or "rename"

	// Delete confirmation
	confirmingDelete bool

	// Editor state (Task 4 stub fields)
	editingNote  *models.Note
	editorLines  []string
	cursorRow    int
	cursorCol    int
	scrollOffset int

	// Status
	statusMessage string
}

type notesKeyMap struct {
	Quit     key.Binding
	Back     key.Binding
	NewNote  key.Binding
	Delete   key.Binding
	Rename   key.Binding
	Sort     key.Binding
	Select   key.Binding
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
}

func defaultNotesKeyMap() notesKeyMap {
	return notesKeyMap{
		Quit:     key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
		Back:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		NewNote:  key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "new note")),
		Delete:   key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		Rename:   key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rename")),
		Sort:     key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "sort")),
		Select:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		PageUp:   key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up")),
		PageDown: key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdn", "page down")),
	}
}

// NewNotesEditorModel creates a new notes editor model.
func NewNotesEditorModel(char *models.Character, storage *storage.CharacterStorage, returnTo string) *NotesEditorModel {
	m := &NotesEditorModel{
		character: char,
		storage:   storage,
		keys:      defaultNotesKeyMap(),
		returnTo:  returnTo,
		mode:      NotesModeList,
		sortOrder: NotesSortLastEdited,
	}
	m.updateSortedDocs()
	return m
}

// Init initializes the model.
func (m *NotesEditorModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m *NotesEditorModel) Update(msg tea.Msg) (*NotesEditorModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		// Ctrl+C always quits
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}

		switch m.mode {
		case NotesModeList:
			return m.updateListMode(msg)
		case NotesModeEditor:
			return m.updateEditorMode(msg)
		}
	}

	return m, nil
}

// updateListMode handles key messages in list mode.
func (m *NotesEditorModel) updateListMode(msg tea.KeyPressMsg) (*NotesEditorModel, tea.Cmd) {
	// Handle input overlay mode (new note or rename)
	if m.inputMode {
		return m.handleInputMode(msg)
	}

	// Handle delete confirmation
	if m.confirmingDelete {
		return m.handleDeleteConfirm(msg)
	}

	// Normal list mode key handling
	m.statusMessage = ""

	switch {
	case key.Matches(msg, m.keys.Back):
		return m, m.goBack()

	case key.Matches(msg, m.keys.Up):
		if m.listCursor > 0 {
			m.listCursor--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.listCursor < len(m.sortedDocs)-1 {
			m.listCursor++
		}
		return m, nil

	case key.Matches(msg, m.keys.PageUp):
		m.listCursor -= 10
		if m.listCursor < 0 {
			m.listCursor = 0
		}
		return m, nil

	case key.Matches(msg, m.keys.PageDown):
		m.listCursor += 10
		if m.listCursor >= len(m.sortedDocs) {
			m.listCursor = len(m.sortedDocs) - 1
		}
		if m.listCursor < 0 {
			m.listCursor = 0
		}
		return m, nil

	case key.Matches(msg, m.keys.Select):
		if len(m.sortedDocs) > 0 {
			m.enterEditorMode()
		}
		return m, nil

	case key.Matches(msg, m.keys.NewNote):
		m.inputMode = true
		m.inputBuffer = ""
		m.inputAction = "new"
		m.statusMessage = "Enter note title:"
		return m, nil

	case key.Matches(msg, m.keys.Delete):
		if len(m.sortedDocs) > 0 {
			docIdx := m.sortedDocs[m.listCursor]
			doc := m.character.Personality.Documents[docIdx]
			m.confirmingDelete = true
			m.statusMessage = fmt.Sprintf("Delete '%s'? (y/n)", doc.Title)
		}
		return m, nil

	case key.Matches(msg, m.keys.Rename):
		if len(m.sortedDocs) > 0 {
			docIdx := m.sortedDocs[m.listCursor]
			doc := m.character.Personality.Documents[docIdx]
			m.inputMode = true
			m.inputBuffer = doc.Title
			m.inputAction = "rename"
			m.statusMessage = "Rename note:"
		}
		return m, nil

	case key.Matches(msg, m.keys.Sort):
		if m.sortOrder == NotesSortLastEdited {
			m.sortOrder = NotesSortAlphabetical
			m.statusMessage = "Sorted alphabetically"
		} else {
			m.sortOrder = NotesSortLastEdited
			m.statusMessage = "Sorted by last edited"
		}
		m.updateSortedDocs()
		return m, nil
	}

	return m, nil
}

// handleInputMode handles key messages when in input overlay mode.
func (m *NotesEditorModel) handleInputMode(msg tea.KeyPressMsg) (*NotesEditorModel, tea.Cmd) {
	switch msg.Code {
	case tea.KeyEscape:
		m.inputMode = false
		m.inputBuffer = ""
		m.inputAction = ""
		m.statusMessage = ""
		return m, nil

	case tea.KeyEnter:
		title := strings.TrimSpace(m.inputBuffer)
		if title == "" {
			m.statusMessage = "Title cannot be empty"
			return m, nil
		}

		if m.inputAction == "new" {
			m.character.Personality.AddNote(title)
			m.saveCharacter()
			m.updateSortedDocs()
			// Set cursor to the new note
			for i, idx := range m.sortedDocs {
				if m.character.Personality.Documents[idx].Title == title {
					m.listCursor = i
					break
				}
			}
			m.statusMessage = fmt.Sprintf("Created '%s'", title)
			// Enter editor mode for the new note
			m.enterEditorMode()
		} else if m.inputAction == "rename" {
			if len(m.sortedDocs) > 0 && m.listCursor < len(m.sortedDocs) {
				docIdx := m.sortedDocs[m.listCursor]
				m.character.Personality.Documents[docIdx].Title = title
				m.character.Personality.Documents[docIdx].UpdatedAt = time.Now()
				m.saveCharacter()
				m.updateSortedDocs()
				m.statusMessage = fmt.Sprintf("Renamed to '%s'", title)
			}
		}

		m.inputMode = false
		m.inputBuffer = ""
		m.inputAction = ""
		return m, nil

	case tea.KeyBackspace:
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}
		return m, nil

	default:
		if msg.Text != "" {
			m.inputBuffer += msg.Text
		}
		return m, nil
	}
}

// handleDeleteConfirm handles key messages during delete confirmation.
func (m *NotesEditorModel) handleDeleteConfirm(msg tea.KeyPressMsg) (*NotesEditorModel, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if len(m.sortedDocs) > 0 && m.listCursor < len(m.sortedDocs) {
			docIdx := m.sortedDocs[m.listCursor]
			doc := m.character.Personality.Documents[docIdx]
			m.character.Personality.DeleteNote(doc.ID)
			m.saveCharacter()
			m.updateSortedDocs()
			// Adjust cursor
			if m.listCursor >= len(m.sortedDocs) && m.listCursor > 0 {
				m.listCursor--
			}
			m.statusMessage = fmt.Sprintf("Deleted '%s'", doc.Title)
		}
		m.confirmingDelete = false
		return m, nil

	default:
		m.confirmingDelete = false
		m.statusMessage = "Delete cancelled"
		return m, nil
	}
}

// updateEditorMode handles key messages in editor mode.
func (m *NotesEditorModel) updateEditorMode(msg tea.KeyPressMsg) (*NotesEditorModel, tea.Cmd) {
	switch msg.Code {
	case tea.KeyEscape:
		m.exitEditorMode()
		return m, nil

	case tea.KeyEnter:
		// Split current line at cursor position
		line := m.editorLines[m.cursorRow]
		before := line[:m.cursorCol]
		after := line[m.cursorCol:]
		m.editorLines[m.cursorRow] = before
		// Insert new line after current row
		newLines := make([]string, len(m.editorLines)+1)
		copy(newLines, m.editorLines[:m.cursorRow+1])
		newLines[m.cursorRow+1] = after
		copy(newLines[m.cursorRow+2:], m.editorLines[m.cursorRow+1:])
		m.editorLines = newLines
		m.cursorRow++
		m.cursorCol = 0
		m.ensureCursorVisible()
		return m, nil

	case tea.KeyBackspace:
		if m.cursorCol > 0 {
			// Delete character before cursor
			line := m.editorLines[m.cursorRow]
			m.editorLines[m.cursorRow] = line[:m.cursorCol-1] + line[m.cursorCol:]
			m.cursorCol--
		} else if m.cursorRow > 0 {
			// Merge current line with previous line
			prevLine := m.editorLines[m.cursorRow-1]
			curLine := m.editorLines[m.cursorRow]
			m.cursorCol = len(prevLine)
			m.editorLines[m.cursorRow-1] = prevLine + curLine
			// Remove current line
			m.editorLines = append(m.editorLines[:m.cursorRow], m.editorLines[m.cursorRow+1:]...)
			m.cursorRow--
		}
		m.ensureCursorVisible()
		return m, nil

	case tea.KeyDelete:
		line := m.editorLines[m.cursorRow]
		if m.cursorCol < len(line) {
			// Delete character at cursor
			m.editorLines[m.cursorRow] = line[:m.cursorCol] + line[m.cursorCol+1:]
		} else if m.cursorRow < len(m.editorLines)-1 {
			// Merge with next line
			m.editorLines[m.cursorRow] = line + m.editorLines[m.cursorRow+1]
			m.editorLines = append(m.editorLines[:m.cursorRow+1], m.editorLines[m.cursorRow+2:]...)
		}
		return m, nil

	case tea.KeyUp:
		if m.cursorRow > 0 {
			m.cursorRow--
			if m.cursorCol > len(m.editorLines[m.cursorRow]) {
				m.cursorCol = len(m.editorLines[m.cursorRow])
			}
		}
		m.ensureCursorVisible()
		return m, nil

	case tea.KeyDown:
		if m.cursorRow < len(m.editorLines)-1 {
			m.cursorRow++
			if m.cursorCol > len(m.editorLines[m.cursorRow]) {
				m.cursorCol = len(m.editorLines[m.cursorRow])
			}
		}
		m.ensureCursorVisible()
		return m, nil

	case tea.KeyLeft:
		if m.cursorCol > 0 {
			m.cursorCol--
		} else if m.cursorRow > 0 {
			m.cursorRow--
			m.cursorCol = len(m.editorLines[m.cursorRow])
		}
		m.ensureCursorVisible()
		return m, nil

	case tea.KeyRight:
		line := m.editorLines[m.cursorRow]
		if m.cursorCol < len(line) {
			m.cursorCol++
		} else if m.cursorRow < len(m.editorLines)-1 {
			m.cursorRow++
			m.cursorCol = 0
		}
		m.ensureCursorVisible()
		return m, nil

	case tea.KeyHome:
		m.cursorCol = 0
		return m, nil

	case tea.KeyEnd:
		m.cursorCol = len(m.editorLines[m.cursorRow])
		return m, nil

	case tea.KeyPgUp:
		visibleHeight := m.height - 4
		if visibleHeight < 1 {
			visibleHeight = 1
		}
		m.cursorRow -= visibleHeight
		if m.cursorRow < 0 {
			m.cursorRow = 0
		}
		if m.cursorCol > len(m.editorLines[m.cursorRow]) {
			m.cursorCol = len(m.editorLines[m.cursorRow])
		}
		m.ensureCursorVisible()
		return m, nil

	case tea.KeyPgDown:
		visibleHeight := m.height - 4
		if visibleHeight < 1 {
			visibleHeight = 1
		}
		m.cursorRow += visibleHeight
		if m.cursorRow >= len(m.editorLines) {
			m.cursorRow = len(m.editorLines) - 1
		}
		if m.cursorCol > len(m.editorLines[m.cursorRow]) {
			m.cursorCol = len(m.editorLines[m.cursorRow])
		}
		m.ensureCursorVisible()
		return m, nil

	default:
		if msg.Text != "" {
			// Insert character(s) at cursor position
			line := m.editorLines[m.cursorRow]
			chars := msg.Text
			m.editorLines[m.cursorRow] = line[:m.cursorCol] + chars + line[m.cursorCol:]
			m.cursorCol += len(chars)
			m.ensureCursorVisible()
		}
		return m, nil
	}
}

// ensureCursorVisible adjusts scrollOffset so the cursor row is visible.
func (m *NotesEditorModel) ensureCursorVisible() {
	visibleHeight := m.height - 4
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	if m.cursorRow < m.scrollOffset {
		m.scrollOffset = m.cursorRow
	}
	if m.cursorRow >= m.scrollOffset+visibleHeight {
		m.scrollOffset = m.cursorRow - visibleHeight + 1
	}
}

// enterEditorMode switches to editor mode for the selected note (from list cursor).
func (m *NotesEditorModel) enterEditorMode() {
	if len(m.sortedDocs) == 0 || m.listCursor >= len(m.sortedDocs) {
		return
	}
	docIdx := m.sortedDocs[m.listCursor]
	note := &m.character.Personality.Documents[docIdx]
	m.enterEditorModeForNote(note)
}

// enterEditorModeByID switches to editor mode for the note with the given ID.
func (m *NotesEditorModel) enterEditorModeByID(noteID string) {
	note := m.character.Personality.FindNote(noteID)
	if note == nil {
		return
	}
	m.enterEditorModeForNote(note)
}

// enterEditorModeForNote sets up editor mode state for the given note.
func (m *NotesEditorModel) enterEditorModeForNote(note *models.Note) {
	m.editingNote = note
	m.mode = NotesModeEditor
	m.editorLines = strings.Split(note.Content, "\n")
	if len(m.editorLines) == 0 {
		m.editorLines = []string{""}
	}
	// Place cursor at end
	m.cursorRow = len(m.editorLines) - 1
	m.cursorCol = len(m.editorLines[m.cursorRow])
	m.scrollOffset = 0
	m.statusMessage = ""
}

// saveEditorContent saves the current editor content back to the note.
func (m *NotesEditorModel) saveEditorContent() {
	if m.editingNote == nil {
		return
	}
	m.editingNote.Content = strings.Join(m.editorLines, "\n")
	m.editingNote.UpdatedAt = time.Now()
	m.saveCharacter()
}

// exitEditorMode saves content and returns to list mode.
func (m *NotesEditorModel) exitEditorMode() {
	m.saveEditorContent()
	m.editingNote = nil
	m.editorLines = nil
	m.mode = NotesModeList
	m.updateSortedDocs()
}

// updateSortedDocs rebuilds the sorted document index list.
func (m *NotesEditorModel) updateSortedDocs() {
	docs := m.character.Personality.Documents
	m.sortedDocs = make([]int, len(docs))
	for i := range docs {
		m.sortedDocs[i] = i
	}

	switch m.sortOrder {
	case NotesSortLastEdited:
		sort.Slice(m.sortedDocs, func(i, j int) bool {
			return docs[m.sortedDocs[i]].UpdatedAt.After(docs[m.sortedDocs[j]].UpdatedAt)
		})
	case NotesSortAlphabetical:
		sort.Slice(m.sortedDocs, func(i, j int) bool {
			return strings.ToLower(docs[m.sortedDocs[i]].Title) < strings.ToLower(docs[m.sortedDocs[j]].Title)
		})
	}

	// Clamp cursor
	if m.listCursor >= len(m.sortedDocs) {
		m.listCursor = len(m.sortedDocs) - 1
	}
	if m.listCursor < 0 {
		m.listCursor = 0
	}
}

// saveCharacter persists the character to storage.
func (m *NotesEditorModel) saveCharacter() {
	_ = m.storage.AutoSave(m.character)
}

// goBack returns a command to navigate back to the previous view.
func (m *NotesEditorModel) goBack() tea.Cmd {
	if m.returnTo == "charinfo" {
		return func() tea.Msg { return BackToCharacterInfoMsg{} }
	}
	return func() tea.Msg { return BackToSheetMsg{} }
}

// View renders the notes editor view.
func (m *NotesEditorModel) View() string {
	if m.character == nil {
		return "No character loaded"
	}

	switch m.mode {
	case NotesModeEditor:
		return m.viewEditorMode()
	default:
		return m.viewListMode()
	}
}

// viewEditorMode renders the full-screen text editor view.
func (m *NotesEditorModel) viewEditorMode() string {
	width := m.width
	if width == 0 {
		width = 80
	}
	height := m.height
	if height == 0 {
		height = 24
	}

	title := ""
	if m.editingNote != nil {
		title = m.editingNote.Title
	}

	// Calculate visible content area (total height minus title bar, footer, borders)
	visibleHeight := height - 4
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	// Ensure scroll offset keeps cursor visible
	m.ensureCursorVisible()

	// Build content lines with cursor rendering
	var contentLines []string
	endRow := m.scrollOffset + visibleHeight
	if endRow > len(m.editorLines) {
		endRow = len(m.editorLines)
	}
	for row := m.scrollOffset; row < endRow; row++ {
		line := m.editorLines[row]
		if row == m.cursorRow {
			// Render cursor at position
			if m.cursorCol >= len(line) {
				line = line + "█"
			} else {
				line = line[:m.cursorCol] + "█" + line[m.cursorCol+1:]
			}
		}
		contentLines = append(contentLines, line)
	}

	// Pad with empty lines to fill visible height
	for len(contentLines) < visibleHeight {
		contentLines = append(contentLines, "")
	}

	content := strings.Join(contentLines, "\n")

	// Footer
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	footer := helpStyle.Render("Esc: save & back to list | PgUp/PgDn: scroll")

	// Build final view with border
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	titleBar := titleStyle.Render(fmt.Sprintf("──── %s ", title))

	result := titleBar + "\n" + content + "\n" + footer

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(0, 1).
		Width(width - 2)

	return boxStyle.Render(result)
}

// viewListMode renders the document list mode view.
func (m *NotesEditorModel) viewListMode() string {
	width := m.width
	if width == 0 {
		width = 80
	}
	height := m.height
	if height == 0 {
		height = 24
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	sortActiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	sortInactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	// Usable width inside borders
	innerWidth := width - 4
	if innerWidth < 20 {
		innerWidth = 20
	}

	var lines []string

	// Title
	charName := m.character.Info.Name
	lines = append(lines, titleStyle.Render(fmt.Sprintf("Notes ── %s", charName)))
	lines = append(lines, "")

	// Sort indicator
	var sortLine string
	if m.sortOrder == NotesSortLastEdited {
		sortLine = "Sort: " + sortActiveStyle.Render("[Last Edited ▾]") + " " + sortInactiveStyle.Render("[A-Z]")
	} else {
		sortLine = "Sort: " + sortInactiveStyle.Render("[Last Edited]") + " " + sortActiveStyle.Render("[A-Z ▾]")
	}
	lines = append(lines, sortLine)
	lines = append(lines, dimStyle.Render(strings.Repeat("─", innerWidth)))
	lines = append(lines, "")

	// Document list
	docs := m.character.Personality.Documents
	if len(m.sortedDocs) == 0 {
		lines = append(lines, dimStyle.Render("No notes yet. Press 'a' to create your first note."))
	} else {
		for i, docIdx := range m.sortedDocs {
			doc := docs[docIdx]
			prefix := "  "
			style := normalStyle
			if i == m.listCursor {
				prefix = "> "
				style = selectedStyle
			}

			timeStr := relativeTime(doc.UpdatedAt)
			titleText := doc.Title

			// Calculate available space for title and time
			timeDisplay := "edited " + timeStr
			// Ensure we have space: prefix(2) + title + gap(4) + timeDisplay
			maxTitleLen := innerWidth - len(prefix) - len(timeDisplay) - 4
			if maxTitleLen < 10 {
				maxTitleLen = 10
			}
			if len(titleText) > maxTitleLen {
				titleText = titleText[:maxTitleLen-3] + "..."
			}

			padding := innerWidth - len(prefix) - len(titleText) - len(timeDisplay)
			if padding < 2 {
				padding = 2
			}

			line := prefix + style.Render(titleText) + strings.Repeat(" ", padding) + dimStyle.Render(timeDisplay)
			lines = append(lines, line)
		}
	}

	lines = append(lines, "")

	// Input overlay
	if m.inputMode {
		var prompt string
		if m.inputAction == "new" {
			prompt = "New note title: "
		} else {
			prompt = "Rename: "
		}
		lines = append(lines, titleStyle.Render(prompt)+normalStyle.Render(m.inputBuffer+"_"))
		lines = append(lines, "")
	}

	// Delete confirmation overlay
	if m.confirmingDelete {
		lines = append(lines, titleStyle.Render(m.statusMessage))
		lines = append(lines, "")
	}

	// Status message (only if not showing overlays)
	if m.statusMessage != "" && !m.confirmingDelete && !m.inputMode {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
		lines = append(lines, statusStyle.Render(m.statusMessage))
		lines = append(lines, "")
	}

	// Footer
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	if m.inputMode {
		lines = append(lines, helpStyle.Render("Enter: confirm | Esc: cancel"))
	} else if m.confirmingDelete {
		lines = append(lines, helpStyle.Render("y: delete | n/Esc: cancel"))
	} else {
		lines = append(lines, helpStyle.Render("Enter: open | a: new | d: delete | r: rename | s: sort | Esc: back"))
	}

	content := strings.Join(lines, "\n")

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(0, 1).
		Width(width - 2)

	return boxStyle.Render(content)
}

// relativeTime returns a human-readable relative time string.
func relativeTime(t time.Time) string {
	dur := time.Since(t)
	switch {
	case dur < time.Minute:
		return "just now"
	case dur < time.Hour:
		return fmt.Sprintf("%d min ago", int(dur.Minutes()))
	case dur < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(dur.Hours()))
	case dur < 7*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(dur.Hours()/24))
	case dur < 30*24*time.Hour:
		return fmt.Sprintf("%d weeks ago", int(dur.Hours()/(24*7)))
	default:
		return fmt.Sprintf("%d months ago", int(dur.Hours()/(24*30)))
	}
}
