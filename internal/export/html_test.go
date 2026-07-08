package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Domo929/sheet/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToHTMLContainsSections(t *testing.T) {
	h := ToHTML(sampleCharacter())

	wants := []string{
		"<!DOCTYPE html>",
		"<title>Lyra — Character Sheet</title>",
		"@page", // print CSS
		"<h1>Lyra</h1>",
		"Level 3 Elf (High Elf) Wizard (Evoker)",
		"<h2>Combat</h2>",
		"Armor Class",
		"<h2>Ability Scores</h2>",
		"<h2>Skills</h2>",
		"<h2>Spellcasting</h2>",
		"Magic Missile",
		"<h2>Inventory</h2>",
		"Spellbook",
		"<h2>Personality</h2>",
	}
	for _, w := range wants {
		assert.Contains(t, h, w, "html should contain %q", w)
	}
}

func TestToHTMLEscapesUserContent(t *testing.T) {
	c := models.NewCharacter("x-1", "<script>alert(1)</script>", "Human", "Fighter")
	c.Personality.Backstory = "Loves <b>swords</b> & shields"
	h := ToHTML(c)

	assert.NotContains(t, h, "<script>alert(1)</script>", "raw script must not appear")
	assert.Contains(t, h, "&lt;script&gt;", "name should be HTML-escaped")
	assert.Contains(t, h, "swords&lt;/b&gt; &amp; shields", "backstory should be escaped")
}

func TestToHTMLOmitsSpellcastingForNonCaster(t *testing.T) {
	c := models.NewCharacter("f-1", "Grok", "Orc", "Barbarian")
	h := ToHTML(c)
	assert.NotContains(t, h, "<h2>Spellcasting</h2>")
	assert.Contains(t, h, "<h1>Grok</h1>")
}

func TestToHTMLShowsClassResources(t *testing.T) {
	c := models.NewCharacter("c-1", "Priest", "Human", "Cleric")
	c.Info.Level = 6
	c.SyncResources()
	h := ToHTML(c)
	assert.Contains(t, h, "<h2>Class Resources</h2>")
	assert.Contains(t, h, "Channel Divinity")
	assert.Contains(t, h, "Short", "partial short-rest recovery should be noted")
}

func TestWriteFilesEmitsHTML(t *testing.T) {
	c := sampleCharacter()
	dir := filepath.Join(t.TempDir(), "out")

	_, _, err := WriteFiles(c, dir)
	require.NoError(t, err)

	htmlPath := filepath.Join(dir, "Lyra.html")
	data, err := os.ReadFile(htmlPath)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(data), "<!DOCTYPE html>"))
	assert.Contains(t, string(data), "<h1>Lyra</h1>")
}
