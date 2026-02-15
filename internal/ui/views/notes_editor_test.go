package views

import (
	"testing"
	"time"

	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestNotesModel(t *testing.T) *NotesEditorModel {
	t.Helper()
	char := models.NewCharacter("test-id", "Test Character", "Human", "Fighter")
	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	return NewNotesEditorModel(char, store, "sheet")
}

func TestNotesEditor_EmptyList(t *testing.T) {
	m := newTestNotesModel(t)

	// Should be in list mode with no documents
	assert.Equal(t, NotesModeList, m.mode)
	assert.Equal(t, 0, len(m.sortedDocs))
	assert.Equal(t, 0, m.listCursor)

	// View should contain empty state message
	view := m.View()
	assert.Contains(t, view, "No notes yet")
	assert.Contains(t, view, "'a' to create")
}

func TestNotesEditor_CreateNote(t *testing.T) {
	m := newTestNotesModel(t)

	// Press 'a' to start creating
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	assert.True(t, m.inputMode)
	assert.Equal(t, "new", m.inputAction)

	// Type a title
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	assert.Equal(t, "Session 1", m.inputBuffer)

	// Press Enter to confirm
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, m.inputMode)

	// Should have created the document and entered editor mode
	assert.Equal(t, 1, len(m.character.Personality.Documents))
	assert.Equal(t, "Session 1", m.character.Personality.Documents[0].Title)
	assert.Equal(t, NotesModeEditor, m.mode)
	assert.NotNil(t, m.editingNote)
}

func TestNotesEditor_SortToggle(t *testing.T) {
	m := newTestNotesModel(t)

	// Add notes with different timestamps
	note1 := m.character.Personality.AddNote("Zebra Notes")
	note1.UpdatedAt = time.Now().Add(-2 * time.Hour)
	note2 := m.character.Personality.AddNote("Alpha Notes")
	note2.UpdatedAt = time.Now().Add(-1 * time.Hour)
	note3 := m.character.Personality.AddNote("Middle Notes")
	note3.UpdatedAt = time.Now()
	m.updateSortedDocs()

	// Default sort: last edited (most recent first)
	assert.Equal(t, NotesSortLastEdited, m.sortOrder)
	docs := m.character.Personality.Documents
	assert.Equal(t, "Middle Notes", docs[m.sortedDocs[0]].Title)
	assert.Equal(t, "Alpha Notes", docs[m.sortedDocs[1]].Title)
	assert.Equal(t, "Zebra Notes", docs[m.sortedDocs[2]].Title)

	// Press 's' to toggle to alphabetical
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	assert.Equal(t, NotesSortAlphabetical, m.sortOrder)
	assert.Equal(t, "Alpha Notes", docs[m.sortedDocs[0]].Title)
	assert.Equal(t, "Middle Notes", docs[m.sortedDocs[1]].Title)
	assert.Equal(t, "Zebra Notes", docs[m.sortedDocs[2]].Title)

	// Press 's' again to toggle back to last edited
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	assert.Equal(t, NotesSortLastEdited, m.sortOrder)
	assert.Equal(t, "Middle Notes", docs[m.sortedDocs[0]].Title)
}

func TestNotesEditor_DeleteNote(t *testing.T) {
	m := newTestNotesModel(t)

	// Add a note
	m.character.Personality.AddNote("To Delete")
	m.updateSortedDocs()
	assert.Equal(t, 1, len(m.sortedDocs))

	// Press 'd' to start delete
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	assert.True(t, m.confirmingDelete)
	assert.Contains(t, m.statusMessage, "Delete")
	assert.Contains(t, m.statusMessage, "To Delete")

	// Press 'y' to confirm
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	assert.False(t, m.confirmingDelete)
	assert.Equal(t, 0, len(m.character.Personality.Documents))
	assert.Equal(t, 0, len(m.sortedDocs))
	assert.Contains(t, m.statusMessage, "Deleted")
}

func TestNotesEditor_DeleteCancel(t *testing.T) {
	m := newTestNotesModel(t)

	// Add a note
	m.character.Personality.AddNote("Keep Me")
	m.updateSortedDocs()

	// Press 'd' to start delete
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	assert.True(t, m.confirmingDelete)

	// Press 'n' to cancel
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	assert.False(t, m.confirmingDelete)
	assert.Equal(t, 1, len(m.character.Personality.Documents))
	assert.Contains(t, m.statusMessage, "cancelled")
}

func TestNotesEditor_RenameNote(t *testing.T) {
	m := newTestNotesModel(t)

	// Add a note
	m.character.Personality.AddNote("Old Name")
	m.updateSortedDocs()

	// Press 'r' to start rename
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	assert.True(t, m.inputMode)
	assert.Equal(t, "rename", m.inputAction)
	assert.Equal(t, "Old Name", m.inputBuffer)

	// Clear and type new name
	// Backspace to clear
	for range len("Old Name") {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	}
	assert.Equal(t, "", m.inputBuffer)

	// Type new name
	for _, r := range "New Name" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	assert.Equal(t, "New Name", m.inputBuffer)

	// Press Enter to confirm
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, m.inputMode)
	assert.Equal(t, "New Name", m.character.Personality.Documents[0].Title)
	assert.Contains(t, m.statusMessage, "Renamed")
}

func TestNotesEditor_NavigateList(t *testing.T) {
	m := newTestNotesModel(t)

	// Add 3 notes
	note1 := m.character.Personality.AddNote("Note 1")
	note1.UpdatedAt = time.Now().Add(-3 * time.Hour)
	note2 := m.character.Personality.AddNote("Note 2")
	note2.UpdatedAt = time.Now().Add(-2 * time.Hour)
	note3 := m.character.Personality.AddNote("Note 3")
	note3.UpdatedAt = time.Now().Add(-1 * time.Hour)
	m.updateSortedDocs()

	// Cursor should start at 0
	assert.Equal(t, 0, m.listCursor)

	// Move down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, m.listCursor)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 2, m.listCursor)

	// Should not go past last item
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 2, m.listCursor)

	// Move up
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 1, m.listCursor)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 0, m.listCursor)

	// Should not go below 0
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 0, m.listCursor)
}

func TestNotesEditor_BackToSheet(t *testing.T) {
	m := newTestNotesModel(t)
	// returnTo is "sheet" by default

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.NotNil(t, cmd)
	msg := cmd()
	_, isBack := msg.(BackToSheetMsg)
	assert.True(t, isBack, "should return BackToSheetMsg")
}

func TestNotesEditor_BackToCharInfo(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Character", "Human", "Fighter")
	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	m := NewNotesEditorModel(char, store, "charinfo")

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.NotNil(t, cmd)
	msg := cmd()
	_, isBack := msg.(BackToCharacterInfoMsg)
	assert.True(t, isBack, "should return BackToCharacterInfoMsg")
}

func TestNotesEditor_InputCancel(t *testing.T) {
	m := newTestNotesModel(t)

	// Press 'a' to start creating
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	assert.True(t, m.inputMode)

	// Type something
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	assert.Equal(t, "te", m.inputBuffer)

	// Press Esc to cancel
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, m.inputMode)
	assert.Equal(t, "", m.inputBuffer)
	assert.Equal(t, 0, len(m.character.Personality.Documents))
}

func TestNotesEditor_EmptyTitleRejected(t *testing.T) {
	m := newTestNotesModel(t)

	// Press 'a' to start creating
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	assert.True(t, m.inputMode)

	// Press Enter without typing anything
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// Should still be in input mode since title was empty
	assert.True(t, m.inputMode)
	assert.Contains(t, m.statusMessage, "empty")
}

func TestNotesEditor_WindowResize(t *testing.T) {
	m := newTestNotesModel(t)

	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
}

func TestNotesEditor_SelectEntersEditorMode(t *testing.T) {
	m := newTestNotesModel(t)

	// Add a note
	m.character.Personality.AddNote("Test Note")
	m.updateSortedDocs()

	// Press Enter to select
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, NotesModeEditor, m.mode)
	assert.NotNil(t, m.editingNote)
	assert.Equal(t, "Test Note", m.editingNote.Title)
}

func TestNotesEditor_SelectOnEmptyListDoesNothing(t *testing.T) {
	m := newTestNotesModel(t)

	// Press Enter with no notes
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, NotesModeList, m.mode)
}

func TestNotesEditor_RelativeTime(t *testing.T) {
	now := time.Now()

	assert.Equal(t, "just now", relativeTime(now))
	assert.Equal(t, "5 min ago", relativeTime(now.Add(-5*time.Minute)))
	assert.Equal(t, "3 hours ago", relativeTime(now.Add(-3*time.Hour)))
	assert.Equal(t, "2 days ago", relativeTime(now.Add(-2*24*time.Hour)))
	assert.Equal(t, "2 weeks ago", relativeTime(now.Add(-14*24*time.Hour)))
	assert.Equal(t, "2 months ago", relativeTime(now.Add(-60*24*time.Hour)))
}

func TestNotesEditor_DeleteAdjustsCursor(t *testing.T) {
	m := newTestNotesModel(t)

	// Add 3 notes
	note1 := m.character.Personality.AddNote("Note A")
	note1.UpdatedAt = time.Now().Add(-3 * time.Hour)
	note2 := m.character.Personality.AddNote("Note B")
	note2.UpdatedAt = time.Now().Add(-2 * time.Hour)
	note3 := m.character.Personality.AddNote("Note C")
	note3.UpdatedAt = time.Now().Add(-1 * time.Hour)
	m.updateSortedDocs()

	// Move cursor to last item
	m.listCursor = 2

	// Delete last item
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	// Cursor should be adjusted
	assert.Equal(t, 1, m.listCursor)
	assert.Equal(t, 2, len(m.sortedDocs))
}
