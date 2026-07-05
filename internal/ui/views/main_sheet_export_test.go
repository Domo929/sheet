package views

import (
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Domo929/sheet/internal/storage"
)

// TestExportSheetKey verifies that pressing E writes Markdown and JSON exports
// next to the character storage directory and reports the destination.
func TestExportSheetKey(t *testing.T) {
	baseDir := filepath.Join(t.TempDir(), "characters")
	store, err := storage.NewCharacterStorage(baseDir)
	require.NoError(t, err)

	char := createTestCharacter() // "Aragorn"
	model := NewMainSheetModel(char, store)
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	model, _ = model.Update(tea.KeyPressMsg{Code: 'E', Text: "E"})

	exportDir := filepath.Join(filepath.Dir(baseDir), "exports")
	mdPath := filepath.Join(exportDir, "Aragorn.md")
	jsonPath := filepath.Join(exportDir, "Aragorn.json")

	assert.FileExists(t, mdPath, "pressing E should write a Markdown export")
	assert.FileExists(t, jsonPath, "pressing E should write a JSON export")
	assert.Contains(t, model.statusMessage, "Exported", "status should confirm the export")

	data, err := os.ReadFile(mdPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "# Aragorn")
}
