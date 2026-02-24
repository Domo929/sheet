package components

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// stripANSI removes ANSI escape sequences from a string for test assertions.
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;:]*m`)

func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func TestNewPanel(t *testing.T) {
	p := NewPanel("Title", "Content", 40, 10)
	assert.Equal(t, "Title", p.Title)
	assert.Equal(t, "Content", p.Content)
	assert.Equal(t, 40, p.Width)
	assert.Equal(t, 10, p.Height)
}

func TestPanelRender(t *testing.T) {
	t.Run("with title", func(t *testing.T) {
		p := NewPanel("Title", "Content", 40, 10)
		result := p.Render()
		stripped := stripANSI(result)
		assert.Contains(t, stripped, "Title")
		assert.Contains(t, stripped, "Content")
	})

	t.Run("without title", func(t *testing.T) {
		p := NewPanel("", "Content", 40, 10)
		result := p.Render()
		assert.Contains(t, result, "Content")
	})

	t.Run("zero dimensions", func(t *testing.T) {
		p := NewPanel("Title", "Content", 0, 0)
		result := p.Render()
		assert.Contains(t, result, "Content")
	})
}

func TestBox(t *testing.T) {
	result := Box("Hello", 20)
	assert.Contains(t, result, "Hello")
}

func TestDefaultPanelStyle(t *testing.T) {
	style := DefaultPanelStyle()
	assert.NotEmpty(t, style.Render("test"))
}

func TestJoinHorizontal(t *testing.T) {
	result := JoinHorizontal(2, "A", "B", "C")
	assert.Contains(t, result, "A")
	assert.Contains(t, result, "B")
	assert.Contains(t, result, "C")
}

func TestJoinVertical(t *testing.T) {
	result := JoinVertical("Line1", "Line2")
	assert.Contains(t, result, "Line1")
	assert.Contains(t, result, "Line2")
}
