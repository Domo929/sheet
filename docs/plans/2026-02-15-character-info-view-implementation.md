# Phase 11: Character Info, Notes & Proficiencies — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add proficiencies panel on the main sheet, a Character Info view for personality/features display and editing, a multi-document Notes editor, and a personality step in the character creation wizard.

**Architecture:** Four independent deliverables: (1) a `renderProficiencies()` method added to MainSheetModel, (2) a new `CharacterInfoModel` view with two-panel layout (personality editing + features display), (3) a new `NotesEditorModel` view with document list + full-screen text editor, and (4) a new `StepPersonality` in the character creation wizard. Each follows existing view patterns: key map, constructor, Init/Update/View methods, message-based navigation.

**Tech Stack:** Go, Bubble Tea (TUI framework), lipgloss (styling), bubbles/key (key bindings)

**Existing Patterns to Follow:**
- View navigation: see `internal/ui/model.go` (OpenInventoryMsg → create model → switch view)
- Key maps: see `internal/ui/views/main_sheet.go` (mainSheetKeyMap struct + defaultMainSheetKeyMap())
- Modal text editing: see `main_sheet.go` HPInputMode / handleHPInput pattern
- Two-panel layout: see `internal/ui/views/spellbook.go` (spell list + detail pane)
- Step wizard: see `internal/ui/views/character_creation.go` (CreationStep enum, moveToStep)
- Auto-save: see `internal/storage/character_storage.go` AutoSave()

**Git Workflow:** Work on branch `feature/character-info-view`. Commits should be logically grouped. PR against `main` when complete.

---

### Task 1: Add Proficiencies Panel to Main Sheet

**Files:**
- Modify: `internal/ui/views/main_sheet.go`

**Step 1: Add renderProficiencies() method**

Add a new method to `MainSheetModel` after `renderAbilitiesAndSaves()` (~line 1170). This method renders a compact panel showing armor, weapon, tool, and language proficiencies:

```go
// renderProficiencies renders the proficiencies panel.
func (m *MainSheetModel) renderProficiencies(width int) string {
	var content strings.Builder

	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	maxValueWidth := width - 14 // account for label + padding + border

	categories := []struct {
		label string
		items []string
	}{
		{"Armor", m.character.Proficiencies.Armor},
		{"Weapons", m.character.Proficiencies.Weapons},
		{"Tools", m.character.Proficiencies.Tools},
		{"Languages", m.character.Proficiencies.Languages},
	}

	for _, cat := range categories {
		if len(cat.items) == 0 {
			continue
		}
		label := labelStyle.Render(fmt.Sprintf("%-10s", cat.label+":"))
		value := valueStyle.Render(wrapText(strings.Join(cat.items, ", "), maxValueWidth))
		content.WriteString(label + " " + value + "\n")
	}

	// If all categories are empty, show a placeholder
	if content.Len() == 0 {
		content.WriteString(valueStyle.Render("None"))
	}

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(width)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	title := titleStyle.Render("Proficiencies")

	return panelStyle.Render(title + "\n" + strings.TrimRight(content.String(), "\n"))
}
```

Note: `wrapText` may need to be added as a helper if it doesn't already exist. Check if there's an existing word-wrap function in the file; the level-up view has one. If not, add a simple one:

```go
func wrapText(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}
	var result strings.Builder
	words := strings.Fields(text)
	lineLen := 0
	for i, word := range words {
		if i > 0 && lineLen+1+len(word) > maxWidth {
			result.WriteString("\n           ") // indent continuation to align with value
			lineLen = 11
		} else if i > 0 {
			result.WriteString(" ")
			lineLen++
		}
		result.WriteString(word)
		lineLen += len(word)
	}
	return result.String()
}
```

**Step 2: Insert proficiencies panel into the left column in View()**

In the `View()` method of `MainSheetModel`, at lines 991-994, change:

```go
// Left column: Abilities/Saves on top, Skills below
abilitiesAndSaves := m.renderAbilitiesAndSaves(leftWidth)
skills := m.renderSkills(leftWidth)
leftColumn := lipgloss.JoinVertical(lipgloss.Left, abilitiesAndSaves, skills)
```

To:

```go
// Left column: Abilities/Saves on top, Proficiencies, Skills below
abilitiesAndSaves := m.renderAbilitiesAndSaves(leftWidth)
proficiencies := m.renderProficiencies(leftWidth)
skills := m.renderSkills(leftWidth)
leftColumn := lipgloss.JoinVertical(lipgloss.Left, abilitiesAndSaves, proficiencies, skills)
```

**Step 3: Run tests**

Run: `go test ./internal/ui/views/ -v`
Expected: All existing tests pass.

**Step 4: Build and verify**

Run: `go build ./cmd/sheet/`
Expected: Compiles without errors.

**Step 5: Commit**

```bash
git add internal/ui/views/main_sheet.go
git commit -m "feat(main-sheet): add proficiencies panel to left column"
```

---

### Task 2: Note Data Model and Migration

**Files:**
- Modify: `internal/models/character_info.go` — add Note struct, update Personality
- Modify: `internal/models/character.go` — add migration helper
- Create: `internal/models/character_info_test.go` — add note tests (if file doesn't exist, add to existing test file)

**Step 1: Add Note struct and update Personality**

Add the `Note` struct to `internal/models/character_info.go` after the `Personality` struct (after line 134):

```go
// Note represents a single named note document.
type Note struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// NewNote creates a new note with the given title.
func NewNote(title string) Note {
	now := time.Now()
	return Note{
		ID:        fmt.Sprintf("note-%d", now.UnixNano()),
		Title:     title,
		Content:   "",
		CreatedAt: now,
		UpdatedAt: now,
	}
}
```

Update the `Personality` struct to add a `Notes` field (the old `Notes string` field stays for backwards compat during migration):

```go
type Personality struct {
	Traits    []string `json:"traits,omitempty"`
	Ideals    []string `json:"ideals,omitempty"`
	Bonds     []string `json:"bonds,omitempty"`
	Flaws     []string `json:"flaws,omitempty"`
	Backstory string   `json:"backstory,omitempty"`
	Notes     string   `json:"notes,omitempty"`       // Deprecated: kept for migration
	Documents []Note   `json:"documents,omitempty"`    // Multi-document notes
}
```

Add methods to Personality for managing notes:

```go
// AddNote creates a new note with the given title and returns it.
func (p *Personality) AddNote(title string) *Note {
	note := NewNote(title)
	p.Documents = append(p.Documents, note)
	return &p.Documents[len(p.Documents)-1]
}

// DeleteNote removes a note by ID.
func (p *Personality) DeleteNote(id string) bool {
	for i, n := range p.Documents {
		if n.ID == id {
			p.Documents = append(p.Documents[:i], p.Documents[i+1:]...)
			return true
		}
	}
	return false
}

// FindNote finds a note by ID and returns a pointer to it.
func (p *Personality) FindNote(id string) *Note {
	for i := range p.Documents {
		if p.Documents[i].ID == id {
			return &p.Documents[i]
		}
	}
	return nil
}

// MigrateNotes converts the deprecated Notes string to a Document if needed.
func (p *Personality) MigrateNotes() {
	if p.Notes != "" && len(p.Documents) == 0 {
		note := NewNote("Notes")
		note.Content = p.Notes
		p.Documents = append(p.Documents, note)
		p.Notes = "" // Clear deprecated field
	}
}
```

Add necessary imports (`fmt`, `time`) to the file if not present.

**Step 2: Add migration call to character loading**

In `internal/storage/character_storage.go`, in the `Load()` and `LoadByPath()` functions, after successfully decoding the character, add a migration call:

```go
// Migrate deprecated notes field
char.Personality.MigrateNotes()
```

Find these functions and add the migration call after the JSON decode but before returning the character.

**Step 3: Write tests for Note model**

Add tests to a test file (create `internal/models/character_info_test.go` if it doesn't exist, or add to existing test file):

```go
func TestNote_AddAndDelete(t *testing.T) {
	p := NewPersonality()

	// Add a note
	note := p.AddNote("Session 1")
	assert.NotEmpty(t, note.ID)
	assert.Equal(t, "Session 1", note.Title)
	assert.Equal(t, 1, len(p.Documents))

	// Find the note
	found := p.FindNote(note.ID)
	require.NotNil(t, found)
	assert.Equal(t, "Session 1", found.Title)

	// Delete the note
	ok := p.DeleteNote(note.ID)
	assert.True(t, ok)
	assert.Equal(t, 0, len(p.Documents))

	// Delete non-existent
	ok = p.DeleteNote("nonexistent")
	assert.False(t, ok)
}

func TestNote_MigrateNotes(t *testing.T) {
	p := NewPersonality()
	p.Notes = "Old notes content"

	p.MigrateNotes()

	assert.Empty(t, p.Notes, "deprecated field should be cleared")
	assert.Equal(t, 1, len(p.Documents))
	assert.Equal(t, "Notes", p.Documents[0].Title)
	assert.Equal(t, "Old notes content", p.Documents[0].Content)
}

func TestNote_MigrateNotes_NoOp(t *testing.T) {
	p := NewPersonality()
	p.Notes = ""

	p.MigrateNotes()

	assert.Equal(t, 0, len(p.Documents), "should not create doc from empty notes")
}

func TestNote_MigrateNotes_AlreadyMigrated(t *testing.T) {
	p := NewPersonality()
	p.Notes = "Old notes"
	p.Documents = []Note{{ID: "existing", Title: "Existing"}}

	p.MigrateNotes()

	assert.Equal(t, 1, len(p.Documents), "should not duplicate migration")
	assert.Equal(t, "Old notes", p.Notes, "should not clear if already has documents")
}
```

**Step 4: Run tests**

Run: `go test ./internal/models/ -v`
Expected: All tests pass.

**Step 5: Run all tests**

Run: `go test ./... -v`
Expected: All tests pass (including storage tests that might load characters).

**Step 6: Commit**

```bash
git add internal/models/character_info.go internal/models/character_info_test.go internal/storage/character_storage.go
git commit -m "feat: add multi-document Note model with migration from legacy notes"
```

---

### Task 3: Notes Editor View — Document List Mode

**Files:**
- Create: `internal/ui/views/notes_editor.go`
- Create: `internal/ui/views/notes_editor_test.go`

**Step 1: Create notes_editor.go with types, constructor, and document list mode**

Create `internal/ui/views/notes_editor.go`:

```go
package views

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	NotesModeEditor                  // Full-screen text editor
)

// NotesSortOrder represents the sort order for the document list.
type NotesSortOrder int

const (
	NotesSortLastEdited   NotesSortOrder = iota
	NotesSortAlphabetical
)

// NotesEditorModel manages the notes editor view.
type NotesEditorModel struct {
	character *models.Character
	storage   *storage.CharacterStorage
	width     int
	height    int
	keys      notesKeyMap
	returnTo  string // "sheet" or "charinfo"

	// Mode
	mode NotesMode

	// Document list state
	listCursor int
	sortOrder  NotesSortOrder
	sortedDocs []int // Indices into character.Personality.Documents

	// Input overlay state (for new note title, rename)
	inputMode   bool
	inputBuffer string
	inputAction string // "new" or "rename"

	// Delete confirmation
	confirmingDelete bool

	// Editor state
	editingNote   *models.Note
	editorLines   []string
	cursorRow     int
	cursorCol     int
	scrollOffset  int

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
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		NewNote: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "new note"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Rename: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rename"),
		),
		Sort: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "toggle sort"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "page down"),
		),
	}
}

// NewNotesEditorModel creates a new notes editor model.
func NewNotesEditorModel(char *models.Character, storage *storage.CharacterStorage, returnTo string) *NotesEditorModel {
	m := &NotesEditorModel{
		character: char,
		storage:   storage,
		keys:      defaultNotesKeyMap(),
		mode:      NotesModeList,
		returnTo:  returnTo,
		sortOrder: NotesSortLastEdited,
	}
	m.updateSortedDocs()
	return m
}

// Init initializes the model.
func (m *NotesEditorModel) Init() tea.Cmd {
	return nil
}

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

func (m *NotesEditorModel) saveCharacter() {
	_ = m.storage.AutoSave(m.character)
}

func (m *NotesEditorModel) goBack() tea.Cmd {
	if m.returnTo == "charinfo" {
		return func() tea.Msg { return BackToCharacterInfoMsg{} }
	}
	return func() tea.Msg { return BackToSheetMsg{} }
}
```

Then add the `Update()` method for document list mode and the `View()` method for the list. This is substantial — implement the full list mode handling:

- Window size message
- In input mode: handle runes, backspace, enter (create/rename), esc (cancel)
- In delete confirmation: y/n
- Normal list mode: up/down, enter (open), a (new), d (delete), r (rename), s (sort toggle), esc (back)

And the `View()` method for list mode showing: title bar, sort indicator, list of documents with cursor, empty state, and footer.

**Step 2: Write tests for document list mode**

Create `internal/ui/views/notes_editor_test.go`:

```go
package views

import (
	"testing"

	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestNotesModel(t *testing.T) *NotesEditorModel {
	char := models.NewCharacter("test-id", "Test Character", "Human", "Fighter")
	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	return NewNotesEditorModel(char, store, "sheet")
}

func TestNotesEditor_EmptyList(t *testing.T) {
	m := newTestNotesModel(t)
	assert.Equal(t, NotesModeList, m.mode)
	assert.Equal(t, 0, len(m.sortedDocs))
}

func TestNotesEditor_CreateNote(t *testing.T) {
	m := newTestNotesModel(t)

	// Press 'a' to start creating
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	assert.True(t, m.inputMode)

	// Type a title
	for _, r := range "Session 1" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Press Enter to create
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, m.inputMode)
	assert.Equal(t, 1, len(m.character.Personality.Documents))
	assert.Equal(t, "Session 1", m.character.Personality.Documents[0].Title)
	assert.Equal(t, NotesModeEditor, m.mode) // Should open editor
}

func TestNotesEditor_SortToggle(t *testing.T) {
	m := newTestNotesModel(t)
	assert.Equal(t, NotesSortLastEdited, m.sortOrder)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	assert.Equal(t, NotesSortAlphabetical, m.sortOrder)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	assert.Equal(t, NotesSortLastEdited, m.sortOrder)
}
```

**Step 3: Run tests**

Run: `go test ./internal/ui/views/ -run TestNotesEditor -v`
Expected: All tests pass.

**Step 4: Commit**

```bash
git add internal/ui/views/notes_editor.go internal/ui/views/notes_editor_test.go
git commit -m "feat: add notes editor view with document list mode"
```

---

### Task 4: Notes Editor View — Full-Screen Text Editor Mode

**Files:**
- Modify: `internal/ui/views/notes_editor.go`
- Modify: `internal/ui/views/notes_editor_test.go`

**Step 1: Implement editor mode in Update()**

Add the editor mode handling in `Update()`. When in `NotesModeEditor`:

- Regular character runes: insert at cursor position
- `Enter`: insert newline (split current line at cursor)
- `Backspace`: delete character before cursor (or merge with previous line)
- `Delete`: delete character at cursor (or merge with next line)
- Arrow keys: move cursor up/down/left/right with line wrapping
- `Home`/`End`: jump to start/end of current line
- `PgUp`/`PgDn`: move cursor up/down by visible page height
- `r`: open rename modal (input overlay for new title)
- `Esc`: save content back to the Note, update timestamp, auto-save character, return to list mode

The editor works with `editorLines []string` (one string per line). On entering editor mode, split `editingNote.Content` by `\n`. On save, join lines back with `\n`.

Add helper methods:

```go
func (m *NotesEditorModel) enterEditorMode(noteID string) {
	note := m.character.Personality.FindNote(noteID)
	if note == nil {
		return
	}
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

func (m *NotesEditorModel) saveEditorContent() {
	if m.editingNote == nil {
		return
	}
	m.editingNote.Content = strings.Join(m.editorLines, "\n")
	m.editingNote.UpdatedAt = time.Now()
	m.saveCharacter()
}

func (m *NotesEditorModel) exitEditorMode() {
	m.saveEditorContent()
	m.editingNote = nil
	m.editorLines = nil
	m.mode = NotesModeList
	m.updateSortedDocs()
}
```

**Step 2: Implement editor mode View()**

Render the full-screen editor:
- Title bar with note title
- Text content with visible cursor (render `█` at cursor position)
- Scroll the view so the cursor row is always visible (use `m.scrollOffset`)
- Footer: `Esc: save & back to list | PgUp/PgDn: scroll | r: rename`

```go
func (m *NotesEditorModel) renderEditor() string {
	// Title bar
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	title := titleStyle.Render(m.editingNote.Title)

	// Calculate visible area
	visibleHeight := m.height - 4 // title + border + footer
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	// Ensure cursor is visible
	if m.cursorRow < m.scrollOffset {
		m.scrollOffset = m.cursorRow
	}
	if m.cursorRow >= m.scrollOffset+visibleHeight {
		m.scrollOffset = m.cursorRow - visibleHeight + 1
	}

	// Render visible lines
	var content strings.Builder
	endLine := m.scrollOffset + visibleHeight
	if endLine > len(m.editorLines) {
		endLine = len(m.editorLines)
	}

	contentWidth := m.width - 4 // padding + border

	for i := m.scrollOffset; i < endLine; i++ {
		line := m.editorLines[i]
		if i == m.cursorRow {
			// Insert cursor character
			if m.cursorCol >= len(line) {
				line = line + "█"
			} else {
				line = line[:m.cursorCol] + "█" + line[m.cursorCol+1:]
			}
		}
		// Truncate if too long
		if len(line) > contentWidth {
			line = line[:contentWidth]
		}
		content.WriteString(line)
		if i < endLine-1 {
			content.WriteString("\n")
		}
	}

	// Pad remaining lines
	for i := endLine - m.scrollOffset; i < visibleHeight; i++ {
		content.WriteString("\n")
	}

	// Footer
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	footer := footerStyle.Render("Esc: save & back to list • PgUp/PgDn: scroll • r: rename")

	// Assemble
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(0, 1).
		Width(m.width - 2).
		Height(visibleHeight)

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		border.Render(content.String()),
		footer,
	)
}
```

**Step 3: Write editor mode tests**

Add to `notes_editor_test.go`:

```go
func TestNotesEditor_EditorTyping(t *testing.T) {
	m := newTestNotesModel(t)

	// Create a note and enter editor
	note := m.character.Personality.AddNote("Test Note")
	m.updateSortedDocs()
	m.enterEditorMode(note.ID)

	assert.Equal(t, NotesModeEditor, m.mode)

	// Type some text
	for _, r := range "Hello" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	assert.Equal(t, "Hello", m.editorLines[0])

	// Press Enter for newline
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, 2, len(m.editorLines))
	assert.Equal(t, "Hello", m.editorLines[0])
	assert.Equal(t, "", m.editorLines[1])

	// Type on second line
	for _, r := range "World" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	assert.Equal(t, "World", m.editorLines[1])
}

func TestNotesEditor_EditorBackspace(t *testing.T) {
	m := newTestNotesModel(t)

	note := m.character.Personality.AddNote("Test")
	note.Content = "Hello\nWorld"
	m.updateSortedDocs()
	m.enterEditorMode(note.ID)

	// Move to start of line 2
	m.cursorRow = 1
	m.cursorCol = 0

	// Backspace should merge with previous line
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	assert.Equal(t, 1, len(m.editorLines))
	assert.Equal(t, "HelloWorld", m.editorLines[0])
}

func TestNotesEditor_EditorSaveOnEsc(t *testing.T) {
	m := newTestNotesModel(t)

	note := m.character.Personality.AddNote("Test")
	m.updateSortedDocs()
	m.enterEditorMode(note.ID)

	for _, r := range "Saved content" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Press Esc to save and go back to list
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Equal(t, NotesModeList, m.mode)
	assert.Equal(t, "Saved content", m.character.Personality.Documents[0].Content)
}

func TestNotesEditor_PageUpDown(t *testing.T) {
	m := newTestNotesModel(t)
	m.height = 10 // small screen

	note := m.character.Personality.AddNote("Test")
	// Create 20 lines of content
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = fmt.Sprintf("Line %d", i+1)
	}
	note.Content = strings.Join(lines, "\n")
	m.updateSortedDocs()
	m.enterEditorMode(note.ID)

	// Cursor should be at end
	assert.Equal(t, 19, m.cursorRow)

	// Page up
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	assert.Less(t, m.cursorRow, 19)
}
```

**Step 4: Run tests**

Run: `go test ./internal/ui/views/ -run TestNotesEditor -v`
Expected: All tests pass.

**Step 5: Commit**

```bash
git add internal/ui/views/notes_editor.go internal/ui/views/notes_editor_test.go
git commit -m "feat: add full-screen text editor mode to notes view"
```

---

### Task 5: Wire Up Notes Editor in App Model and Main Sheet

**Files:**
- Modify: `internal/ui/model.go` — add notesEditorModel field, handle messages, route view
- Modify: `internal/ui/views/main_sheet.go` — add `n` key binding

**Step 1: Add ViewNotes to ViewType enum**

In `internal/ui/model.go`, add `ViewNotes` to the ViewType enum (after `ViewRest` at line 23):

```go
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
	ViewNotes // Add this
)
```

**Step 2: Add notesEditorModel field to Model struct**

In the Model struct (line 26-44), add:

```go
notesEditorModel *views.NotesEditorModel
```

**Step 3: Handle OpenNotesMsg and BackToCharacterInfoMsg**

In `Update()`, add message handlers after the existing `OpenLevelUpMsg` case (around line 167):

```go
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

case views.BackToCharacterInfoMsg:
	// Return from notes to character info view
	m.notesEditorModel = nil
	m.currentView = ViewCharacterInfo
	return m, nil
```

Update the `BackToSheetMsg` handler (line 125-131) to also clear `notesEditorModel`:

```go
case views.BackToSheetMsg:
	m.currentView = ViewMainSheet
	m.inventoryModel = nil
	m.spellbookModel = nil
	m.levelUpModel = nil
	m.notesEditorModel = nil    // Add this
	return m, nil
```

**Step 4: Add view routing for ViewNotes**

In `updateCurrentView()` (around line 281), add before the `default` case:

```go
case ViewNotes:
	if m.notesEditorModel != nil {
		updatedModel, c := m.notesEditorModel.Update(msg)
		m.notesEditorModel = updatedModel
		cmd = c
	}
```

Update `renderCharacterInfo()` to handle notes model, and add `renderNotes()`:

```go
func (m Model) renderNotes() string {
	if m.notesEditorModel != nil {
		return m.notesEditorModel.View()
	}
	return "Notes View (loading...)"
}
```

Add `ViewNotes` to the `View()` switch (after the `ViewRest` case):

```go
case ViewNotes:
	return m.renderNotes()
```

**Step 5: Add `n` key binding to main sheet**

In `internal/ui/views/main_sheet.go`:

Add to `mainSheetKeyMap` struct (around line 198):

```go
Notes key.Binding
```

Add to `defaultMainSheetKeyMap()` (around line 290):

```go
Notes: key.NewBinding(
	key.WithKeys("n"),
	key.WithHelp("n", "notes"),
),
```

Add key handler in `Update()` (after the `Info` key handler, around line 405):

```go
case key.Matches(msg, m.keys.Notes):
	return m, func() tea.Msg { return OpenNotesMsg{ReturnTo: "sheet"} }
```

Update the help footer string (line 2311) to include `n: notes`:

```go
help := "tab/shift+tab: navigate panels • i: inventory • s: spellbook • n: notes • x: add XP • L: level up • r: rest • esc: back • q: quit"
```

**Step 6: Run all tests**

Run: `go test ./... -v`
Expected: All tests pass.

**Step 7: Build and verify**

Run: `go build ./cmd/sheet/`
Expected: Compiles without errors.

**Step 8: Commit**

```bash
git add internal/ui/model.go internal/ui/views/main_sheet.go
git commit -m "feat: wire up notes editor view with navigation from main sheet"
```

---

### Task 6: Character Info View — Core Structure and Features Panel

**Files:**
- Create: `internal/ui/views/character_info.go`
- Create: `internal/ui/views/character_info_test.go`

**Step 1: Create character_info.go with types, constructor, and features panel**

Create `internal/ui/views/character_info.go` with the following structure:

```go
package views

import (
	"fmt"
	"strings"

	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// OpenCharacterInfoMsg signals to open the character info view.
type OpenCharacterInfoMsg struct{}

// CharacterInfoFocus represents which panel is focused.
type CharacterInfoFocus int

const (
	CharInfoFocusPersonality CharacterInfoFocus = iota
	CharInfoFocusFeatures
)

// FeatureCategory represents a category tab in the features panel.
type FeatureCategory int

const (
	FeatureCategoryRacial FeatureCategory = iota
	FeatureCategoryClass
	FeatureCategorySubclass
	FeatureCategoryFeats
)

// PersonalitySection represents which personality section the cursor is in.
type PersonalitySection int

const (
	PersonalitySectionTraits PersonalitySection = iota
	PersonalitySectionIdeals
	PersonalitySectionBonds
	PersonalitySectionFlaws
	PersonalitySectionBackstory
)

// CharacterInfoModel manages the character info view.
type CharacterInfoModel struct {
	character *models.Character
	storage   *storage.CharacterStorage
	width     int
	height    int
	keys      charInfoKeyMap

	// Focus
	focus CharacterInfoFocus

	// Features panel state
	featureCategory FeatureCategory
	featureCursor   int
	featureScroll   int

	// Personality panel state
	personalitySection PersonalitySection
	personalityCursor  int // index within current section
	personalityScroll  int

	// Edit modal state
	editMode    bool
	editBuffer  string
	editAction  string // "edit", "add"
	editSection PersonalitySection
	editIndex   int // index of item being edited (-1 for new)

	// Backstory expanded view
	backstoryExpanded bool

	// Delete confirmation
	confirmingDelete bool

	// Status
	statusMessage string
}
```

Add key map, constructor, Init(), Update(), and View() methods following the inventory/spellbook pattern.

The **features panel** (right side) should:
- Show category tabs at the top: `[Racial] [Class] [Subclass] [Feats]`
- Use Left/Right to switch categories
- Show a scrollable list of features in the selected category
- Show feature detail (description, source, level) in the bottom half of the panel
- Features data comes from `m.character.Features.RacialTraits`, `.ClassFeatures`, `.Feats`
- Class features sorted by level ascending
- Subclass features filtered from ClassFeatures where source contains the subclass name

Add a helper to get features for the current category:

```go
func (m *CharacterInfoModel) getFeaturesForCategory() []models.Feature {
	switch m.featureCategory {
	case FeatureCategoryRacial:
		return m.character.Features.RacialTraits
	case FeatureCategoryClass:
		var features []models.Feature
		subclass := m.character.Info.Subclass
		for _, f := range m.character.Features.ClassFeatures {
			// Exclude subclass features
			if subclass != "" && strings.Contains(f.Source, subclass) {
				continue
			}
			features = append(features, f)
		}
		return features
	case FeatureCategorySubclass:
		var features []models.Feature
		subclass := m.character.Info.Subclass
		if subclass == "" {
			return features
		}
		for _, f := range m.character.Features.ClassFeatures {
			if strings.Contains(f.Source, subclass) {
				features = append(features, f)
			}
		}
		return features
	case FeatureCategoryFeats:
		return m.character.Features.Feats
	}
	return nil
}
```

**Step 2: Write tests for features panel**

```go
func TestCharacterInfo_FeaturesCategorySwitch(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Fighter", "Human", "Fighter")
	char.Features.AddRacialTrait("Darkvision", "Human", "See in dim light")
	char.Features.AddClassFeature("Action Surge", "Fighter 2", "Extra action", 2, "")
	char.Features.AddFeat("Alert", "Can't be surprised")

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)

	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusFeatures

	// Should start on Racial
	assert.Equal(t, FeatureCategoryRacial, m.featureCategory)

	features := m.getFeaturesForCategory()
	assert.Equal(t, 1, len(features))
	assert.Equal(t, "Darkvision", features[0].Name)

	// Switch to Class
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, FeatureCategoryClass, m.featureCategory)

	// Switch to Feats
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, FeatureCategoryFeats, m.featureCategory)
}
```

**Step 3: Run tests**

Run: `go test ./internal/ui/views/ -run TestCharacterInfo -v`
Expected: All tests pass.

**Step 4: Commit**

```bash
git add internal/ui/views/character_info.go internal/ui/views/character_info_test.go
git commit -m "feat: add character info view with features panel"
```

---

### Task 7: Character Info View — Personality Panel with Editing

**Files:**
- Modify: `internal/ui/views/character_info.go`
- Modify: `internal/ui/views/character_info_test.go`

**Step 1: Implement personality panel display**

The personality panel (left side) displays:
- **Traits:** bulleted list of `m.character.Personality.Traits`
- **Ideals:** bulleted list of `m.character.Personality.Ideals`
- **Bonds:** bulleted list of `m.character.Personality.Bonds`
- **Flaws:** bulleted list of `m.character.Personality.Flaws`
- **Backstory:** word-wrapped text, truncated to ~3 lines with "(Enter to expand)" if longer

Each section header is bold. The cursor (when personality panel is focused) highlights the current item with `>` prefix.

Navigation: Up/Down moves between items across all sections. The `personalitySection` and `personalityCursor` track which section and which item within it the cursor is on.

**Step 2: Implement personality editing**

When personality panel is focused:
- `e` — opens edit modal prepopulated with current item text
- `a` — opens edit modal empty to add a new item to current section
- `d` — shows delete confirmation for current item
- `Enter` on backstory — toggles `backstoryExpanded` (full view)

Edit modal overlay (follows the HP input modal pattern):
- Shows a lipgloss-bordered box centered on screen
- Text buffer with cursor indicator (`█`)
- For single-line items (traits, ideals, bonds, flaws): `Enter` confirms, `Esc` cancels
- For backstory: `Enter` inserts newline, `Ctrl+S` saves, `Esc` cancels

On confirm:
- Edit mode: replace the item in the appropriate slice
- Add mode: append to the appropriate slice
- Call `m.saveCharacter()` after mutation

Delete confirmation:
- Shows "Delete this [trait/ideal/bond/flaw]? (y/n)"
- `y` removes the item, `n` cancels

```go
func (m *CharacterInfoModel) applyEdit() {
	text := strings.TrimSpace(m.editBuffer)
	if text == "" {
		return
	}

	switch m.editSection {
	case PersonalitySectionTraits:
		if m.editAction == "add" {
			m.character.Personality.AddTrait(text)
		} else {
			m.character.Personality.Traits[m.editIndex] = text
		}
	case PersonalitySectionIdeals:
		if m.editAction == "add" {
			m.character.Personality.AddIdeal(text)
		} else {
			m.character.Personality.Ideals[m.editIndex] = text
		}
	case PersonalitySectionBonds:
		if m.editAction == "add" {
			m.character.Personality.AddBond(text)
		} else {
			m.character.Personality.Bonds[m.editIndex] = text
		}
	case PersonalitySectionFlaws:
		if m.editAction == "add" {
			m.character.Personality.AddFlaw(text)
		} else {
			m.character.Personality.Flaws[m.editIndex] = text
		}
	case PersonalitySectionBackstory:
		m.character.Personality.Backstory = text
	}

	m.saveCharacter()
}
```

**Step 3: Write editing tests**

```go
func TestCharacterInfo_AddTrait(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)

	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusPersonality
	m.personalitySection = PersonalitySectionTraits

	// Press 'a' to add
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	assert.True(t, m.editMode)

	// Type trait text
	for _, r := range "Brave" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Confirm
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, m.editMode)
	assert.Contains(t, char.Personality.Traits, "Brave")
}

func TestCharacterInfo_DeleteTrait(t *testing.T) {
	char := models.NewCharacter("test-id", "Test", "Human", "Fighter")
	char.Personality.AddTrait("Brave")
	char.Personality.AddTrait("Curious")
	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)

	m := NewCharacterInfoModel(char, store)
	m.focus = CharInfoFocusPersonality
	m.personalitySection = PersonalitySectionTraits
	m.personalityCursor = 0

	// Press 'd' to delete
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	assert.True(t, m.confirmingDelete)

	// Confirm with 'y'
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	assert.False(t, m.confirmingDelete)
	assert.Equal(t, 1, len(char.Personality.Traits))
	assert.Equal(t, "Curious", char.Personality.Traits[0])
}
```

**Step 4: Run tests**

Run: `go test ./internal/ui/views/ -run TestCharacterInfo -v`
Expected: All tests pass.

**Step 5: Commit**

```bash
git add internal/ui/views/character_info.go internal/ui/views/character_info_test.go
git commit -m "feat(character-info): add personality panel with editing support"
```

---

### Task 8: Wire Up Character Info View in App Model

**Files:**
- Modify: `internal/ui/model.go` — add characterInfoModel field, handle messages, route view
- Modify: `internal/ui/views/main_sheet.go` — change `c` key to send OpenCharacterInfoMsg

**Step 1: Add characterInfoModel to Model struct**

In `internal/ui/model.go`, add to the Model struct:

```go
characterInfoModel *views.CharacterInfoModel
```

**Step 2: Handle OpenCharacterInfoMsg in Update()**

Add after the existing `OpenSpellbookMsg` case:

```go
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
```

**Step 3: Update BackToSheetMsg to clear characterInfoModel**

```go
case views.BackToSheetMsg:
	m.currentView = ViewMainSheet
	m.inventoryModel = nil
	m.spellbookModel = nil
	m.levelUpModel = nil
	m.notesEditorModel = nil
	m.characterInfoModel = nil  // Add this
	return m, nil
```

**Step 4: Add view routing for ViewCharacterInfo**

In `updateCurrentView()`, add:

```go
case ViewCharacterInfo:
	if m.characterInfoModel != nil {
		updatedModel, c := m.characterInfoModel.Update(msg)
		m.characterInfoModel = updatedModel
		cmd = c
	}
```

Update `renderCharacterInfo()`:

```go
func (m Model) renderCharacterInfo() string {
	if m.characterInfoModel != nil {
		return m.characterInfoModel.View()
	}
	return "Character Info View (loading...)"
}
```

**Step 5: Update main sheet `c` key handler**

In `internal/ui/views/main_sheet.go`, change the `Info` key handler (lines 403-405) from:

```go
case key.Matches(msg, m.keys.Info):
	m.statusMessage = "Character info view coming soon..."
	return m, nil
```

To:

```go
case key.Matches(msg, m.keys.Info):
	return m, func() tea.Msg { return OpenCharacterInfoMsg{} }
```

**Step 6: Add `n` key to character info view**

In `character_info.go`, add a `Notes` key binding and handler so pressing `n` from the Character Info view sends `OpenNotesMsg{ReturnTo: "charinfo"}`.

**Step 7: Handle BackToCharacterInfoMsg in model.go**

The `BackToCharacterInfoMsg` from notes view should return to the Character Info view:

```go
case views.BackToCharacterInfoMsg:
	m.notesEditorModel = nil
	m.currentView = ViewCharacterInfo
	return m, nil
```

**Step 8: Run all tests**

Run: `go test ./... -v`
Expected: All tests pass.

**Step 9: Build and verify**

Run: `go build ./cmd/sheet/`
Expected: Compiles without errors.

**Step 10: Commit**

```bash
git add internal/ui/model.go internal/ui/views/main_sheet.go internal/ui/views/character_info.go
git commit -m "feat: wire up character info view with navigation from main sheet"
```

---

### Task 9: Character Creation — Personality Step

**Files:**
- Modify: `internal/ui/views/character_creation.go`

**Step 1: Add StepPersonality to the CreationStep enum**

Change the enum (lines 18-27) to insert `StepPersonality` between `StepEquipment` and `StepReview`:

```go
const (
	StepBasicInfo CreationStep = iota
	StepRace
	StepClass
	StepBackground
	StepAbilities
	StepProficiencies
	StepEquipment
	StepPersonality  // Add this
	StepReview
)
```

**Step 2: Add personality step state to CharacterCreationModel**

Add fields to the struct:

```go
// Personality step
personalityTraits     []string
personalityIdeals     []string
personalityBonds      []string
personalityFlaws      []string
personalityBackstory  string
personalityFocusField int // 0=traits, 1=ideals, 2=bonds, 3=flaws, 4=backstory
personalityItemCursor int // cursor within the focused field's entries
personalityEditing    bool
personalityEditBuffer string
```

Initialize in the constructor (NewCharacterCreationModel):

```go
personalityTraits: []string{""},
personalityIdeals: []string{""},
personalityBonds:  []string{""},
personalityFlaws:  []string{""},
```

**Step 3: Update step flow**

In `handleEquipmentKeys()` (around line 593 and 664), change `StepReview` to `StepPersonality`:

```go
return m.moveToStep(StepPersonality) // was StepReview
```

Add a case to `moveToStep()`:

```go
case StepPersonality:
	// No data loading needed
```

**Step 4: Add handlePersonalityKeys()**

Route to it in Update() (add after the StepEquipment case, around line 258):

```go
case StepPersonality:
	return m.handlePersonalityKeys(msg)
```

Implement `handlePersonalityKeys()`:

```go
func (m *CharacterCreationModel) handlePersonalityKeys(msg tea.KeyMsg) (*CharacterCreationModel, tea.Cmd) {
	// If editing an entry, handle text input
	if m.personalityEditing {
		switch msg.Type {
		case tea.KeyEnter:
			if m.personalityFocusField == 4 {
				// Backstory: Enter inserts newline
				m.personalityEditBuffer += "\n"
				return m, nil
			}
			// Save the edit
			m.savePersonalityEdit()
			m.personalityEditing = false
			return m, nil
		case tea.KeyEsc:
			m.personalityEditing = false
			return m, nil
		case tea.KeyBackspace:
			if len(m.personalityEditBuffer) > 0 {
				m.personalityEditBuffer = m.personalityEditBuffer[:len(m.personalityEditBuffer)-1]
			}
			return m, nil
		case tea.KeyRunes:
			m.personalityEditBuffer += string(msg.Runes)
			return m, nil
		}
		return m, nil
	}

	switch msg.String() {
	case "tab":
		// Move to next field
		m.personalityFocusField = (m.personalityFocusField + 1) % 5
		m.personalityItemCursor = 0
		return m, nil
	case "shift+tab":
		// Move to previous field
		m.personalityFocusField = (m.personalityFocusField + 4) % 5
		m.personalityItemCursor = 0
		return m, nil
	case "up", "k":
		if m.personalityItemCursor > 0 {
			m.personalityItemCursor--
		}
		return m, nil
	case "down", "j":
		items := m.getPersonalityFieldItems()
		if m.personalityItemCursor < len(items) {
			// len(items) allows cursor on "+ Add another"
			m.personalityItemCursor++
		}
		return m, nil
	case "enter":
		items := m.getPersonalityFieldItems()
		if m.personalityItemCursor == len(items) {
			// On "+ Add another" - add new entry
			m.addPersonalityEntry()
			return m, nil
		}
		if m.personalityFocusField == 4 {
			// Backstory: enter edit mode
			m.personalityEditing = true
			m.personalityEditBuffer = m.personalityBackstory
			return m, nil
		}
		// Edit current item
		if m.personalityItemCursor < len(items) {
			m.personalityEditing = true
			m.personalityEditBuffer = items[m.personalityItemCursor]
			return m, nil
		}
		// Advance to review if at end
		return m.moveToStep(StepReview)
	case "d":
		m.deletePersonalityEntry()
		return m, nil
	case "esc":
		return m.moveToStep(StepEquipment)
	}

	// If on backstory and not editing, capture runes to start editing
	if m.personalityFocusField == 4 && msg.Type == tea.KeyRunes {
		m.personalityEditing = true
		m.personalityEditBuffer = m.personalityBackstory + string(msg.Runes)
		return m, nil
	}

	// If on a list item, capture runes to start editing
	items := m.getPersonalityFieldItems()
	if m.personalityFocusField < 4 && m.personalityItemCursor < len(items) && msg.Type == tea.KeyRunes {
		m.personalityEditing = true
		m.personalityEditBuffer = items[m.personalityItemCursor] + string(msg.Runes)
		return m, nil
	}

	return m, nil
}
```

Add helper methods:

```go
func (m *CharacterCreationModel) getPersonalityFieldItems() []string {
	switch m.personalityFocusField {
	case 0:
		return m.personalityTraits
	case 1:
		return m.personalityIdeals
	case 2:
		return m.personalityBonds
	case 3:
		return m.personalityFlaws
	default:
		return nil
	}
}

func (m *CharacterCreationModel) addPersonalityEntry() {
	switch m.personalityFocusField {
	case 0:
		m.personalityTraits = append(m.personalityTraits, "")
		m.personalityItemCursor = len(m.personalityTraits) - 1
	case 1:
		m.personalityIdeals = append(m.personalityIdeals, "")
		m.personalityItemCursor = len(m.personalityIdeals) - 1
	case 2:
		m.personalityBonds = append(m.personalityBonds, "")
		m.personalityItemCursor = len(m.personalityBonds) - 1
	case 3:
		m.personalityFlaws = append(m.personalityFlaws, "")
		m.personalityItemCursor = len(m.personalityFlaws) - 1
	}
	m.personalityEditing = true
	m.personalityEditBuffer = ""
}

func (m *CharacterCreationModel) deletePersonalityEntry() {
	items := m.getPersonalityFieldItems()
	if len(items) <= 1 || m.personalityItemCursor >= len(items) {
		return
	}
	switch m.personalityFocusField {
	case 0:
		m.personalityTraits = append(m.personalityTraits[:m.personalityItemCursor], m.personalityTraits[m.personalityItemCursor+1:]...)
	case 1:
		m.personalityIdeals = append(m.personalityIdeals[:m.personalityItemCursor], m.personalityIdeals[m.personalityItemCursor+1:]...)
	case 2:
		m.personalityBonds = append(m.personalityBonds[:m.personalityItemCursor], m.personalityBonds[m.personalityItemCursor+1:]...)
	case 3:
		m.personalityFlaws = append(m.personalityFlaws[:m.personalityItemCursor], m.personalityFlaws[m.personalityItemCursor+1:]...)
	}
	if m.personalityItemCursor >= len(m.getPersonalityFieldItems()) {
		m.personalityItemCursor = len(m.getPersonalityFieldItems()) - 1
	}
}

func (m *CharacterCreationModel) savePersonalityEdit() {
	text := m.personalityEditBuffer
	switch m.personalityFocusField {
	case 0:
		m.personalityTraits[m.personalityItemCursor] = text
	case 1:
		m.personalityIdeals[m.personalityItemCursor] = text
	case 2:
		m.personalityBonds[m.personalityItemCursor] = text
	case 3:
		m.personalityFlaws[m.personalityItemCursor] = text
	case 4:
		m.personalityBackstory = text
	}
}
```

**Step 5: Add renderPersonality()**

```go
func (m *CharacterCreationModel) renderPersonality() string {
	var content strings.Builder

	stepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
	content.WriteString(stepStyle.Render("Step 8 of 9: Personality"))
	content.WriteString("\n\n")

	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Italic(true)
	content.WriteString(infoStyle.Render("All fields are optional. You can fill these in later from the Character Info view."))
	content.WriteString("\n\n")

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))

	fields := []struct {
		label string
		items []string
		index int
	}{
		{"Personality Traits", m.personalityTraits, 0},
		{"Ideals", m.personalityIdeals, 1},
		{"Bonds", m.personalityBonds, 2},
		{"Flaws", m.personalityFlaws, 3},
	}

	for _, field := range fields {
		focused := m.personalityFocusField == field.index
		content.WriteString(headerStyle.Render(field.label + ":"))
		content.WriteString("\n")

		for i, item := range field.items {
			prefix := "  "
			if focused && i == m.personalityItemCursor {
				prefix = cursorStyle.Render("> ")
				if m.personalityEditing {
					content.WriteString(prefix + m.personalityEditBuffer + "█\n")
					continue
				}
			}
			if item == "" {
				content.WriteString(prefix + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("(empty)") + "\n")
			} else {
				content.WriteString(prefix + item + "\n")
			}
		}

		if focused && m.personalityItemCursor == len(field.items) {
			content.WriteString(cursorStyle.Render("> ") + "+ Add another\n")
		} else {
			content.WriteString("  + Add another\n")
		}
		content.WriteString("\n")
	}

	// Backstory
	bsFocused := m.personalityFocusField == 4
	content.WriteString(headerStyle.Render("Backstory:"))
	content.WriteString("\n")
	if bsFocused && m.personalityEditing {
		content.WriteString("  " + m.personalityEditBuffer + "█\n")
	} else if m.personalityBackstory == "" {
		content.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("(empty — press Enter to write)") + "\n")
	} else {
		content.WriteString("  " + m.personalityBackstory + "\n")
	}

	content.WriteString("\n")

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	content.WriteString(helpStyle.Render("Tab: next field • Enter: edit/add • d: delete • Esc: back"))

	return content.String()
}
```

**Step 6: Add personality rendering to View()**

In the `View()` switch (around line 1714), add:

```go
case StepPersonality:
	content.WriteString(m.renderPersonality())
```

**Step 7: Update step numbering**

Update all "Step X of Y" strings in the existing render functions to say "of 9" instead of "of 8" (since we added a step). Update the Equipment step to say "Step 7 of 9" and the Review step rendering to say "Step 9 of 9". Basic Info becomes "Step 1 of 9", etc.

Affected render functions: `renderBasicInfo`, `renderRaceSelection`, `renderClassSelection`, `renderBackgroundSelection`, `renderAbilityScores`, `renderProficiencySelection`, `renderEquipmentSelection`, `renderReview`.

**Step 8: Apply personality data in finalizeCharacter()**

In `finalizeCharacter()`, before saving, apply the personality data:

```go
// Apply personality data (strip empty entries)
for _, trait := range m.personalityTraits {
	if strings.TrimSpace(trait) != "" {
		m.character.Personality.AddTrait(strings.TrimSpace(trait))
	}
}
for _, ideal := range m.personalityIdeals {
	if strings.TrimSpace(ideal) != "" {
		m.character.Personality.AddIdeal(strings.TrimSpace(ideal))
	}
}
for _, bond := range m.personalityBonds {
	if strings.TrimSpace(bond) != "" {
		m.character.Personality.AddBond(strings.TrimSpace(bond))
	}
}
for _, flaw := range m.personalityFlaws {
	if strings.TrimSpace(flaw) != "" {
		m.character.Personality.AddFlaw(strings.TrimSpace(flaw))
	}
}
if strings.TrimSpace(m.personalityBackstory) != "" {
	m.character.Personality.Backstory = strings.TrimSpace(m.personalityBackstory)
}
```

**Step 9: Run all tests**

Run: `go test ./... -v`
Expected: All tests pass.

**Step 10: Build and verify**

Run: `go build ./cmd/sheet/`
Expected: Compiles without errors.

**Step 11: Commit**

```bash
git add internal/ui/views/character_creation.go
git commit -m "feat(character-creation): add optional personality step before review"
```

---

### Task 10: Integration Testing and Polish

**Files:**
- Modify: various — polish rendering, fix edge cases
- Add tests as needed

**Step 1: Run full test suite**

Run: `go test ./... -v`
Expected: All tests pass.

**Step 2: Build and test binary**

Run: `go build ./cmd/sheet/`
Expected: Compiles without errors.

**Step 3: Test navigation flow manually (if possible)**

Verify these flows work:
- Main sheet → `c` → Character Info view → `Esc` → Main sheet
- Main sheet → `n` → Notes list → `a` → create note → edit → `Esc` → list → `Esc` → Main sheet
- Character Info → `n` → Notes list → `Esc` → Character Info
- Character creation → Equipment → Personality step → Review → (skip all personality) → Create
- Proficiencies panel appears on main sheet left column

**Step 4: Fix any issues found during testing**

Address edge cases:
- Empty character with no features/proficiencies
- Notes view with 0 documents
- Very long backstory text
- Small terminal size

**Step 5: Commit any fixes**

```bash
git add -A
git commit -m "fix: polish character info, notes, and proficiencies views"
```

---

### Task 11: Final PR

**Step 1: Run all tests one final time**

Run: `go test ./...`
Expected: All pass, no failures.

**Step 2: Push branch and create PR**

```bash
git push -u origin feature/character-info-view
gh pr create --title "feat: Add character info, notes & proficiencies (Phase 11)" --body "$(cat <<'EOF'
## Summary
- Add proficiencies panel to main sheet left column (armor, weapons, tools, languages)
- Add Character Info view (`c` key) with personality editing and features display grouped by source
- Add multi-document Notes editor (`n` key) with document list and full-screen text editor
- Add optional personality step to character creation wizard (after equipment, before review)
- Add Note data model with migration from legacy single-string notes field

## Test plan
- [ ] Unit tests for Note model (add, delete, find, migrate)
- [ ] Unit tests for notes editor (create, sort toggle, editor typing, backspace, save)
- [ ] Unit tests for character info features panel (category switching)
- [ ] Unit tests for character info personality editing (add, delete traits)
- [ ] All existing tests pass
- [ ] Manual testing: proficiencies panel visible on main sheet
- [ ] Manual testing: character info view navigation and editing
- [ ] Manual testing: notes create/edit/delete/rename/sort
- [ ] Manual testing: character creation personality step (skip through, fill in, etc.)

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

**Step 3: Wait for review**
