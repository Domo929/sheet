package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHelpFooter(t *testing.T) {
	bindings := []KeyBinding{
		{Key: "q", Description: "quit"},
		{Key: "enter", Description: "select"},
	}
	footer := NewHelpFooter(bindings...)
	assert.Len(t, footer.Bindings, 2)
}

func TestHelpFooterRender(t *testing.T) {
	t.Run("with bindings", func(t *testing.T) {
		footer := NewHelpFooter(
			KeyBinding{Key: "q", Description: "quit"},
			KeyBinding{Key: "enter", Description: "select"},
		)
		result := footer.Render()
		assert.Contains(t, result, "q")
		assert.Contains(t, result, "quit")
		assert.Contains(t, result, "enter")
		assert.Contains(t, result, "select")
	})

	t.Run("empty bindings", func(t *testing.T) {
		footer := NewHelpFooter()
		result := footer.Render()
		assert.Empty(t, result)
	})

	t.Run("with width", func(t *testing.T) {
		footer := HelpFooter{
			Bindings: []KeyBinding{{Key: "q", Description: "quit"}},
			Width:    80,
		}
		result := footer.Render()
		assert.Contains(t, result, "q")
	})
}

func TestCommonBindings(t *testing.T) {
	bindings := CommonBindings()
	assert.NotEmpty(t, bindings)
	keys := make([]string, len(bindings))
	for i, b := range bindings {
		keys[i] = b.Key
	}
	assert.Contains(t, keys, "q")
	assert.Contains(t, keys, "enter")
}

func TestNavigationBindings(t *testing.T) {
	bindings := NavigationBindings()
	assert.NotEmpty(t, bindings)
}

func TestListBindings(t *testing.T) {
	bindings := ListBindings()
	assert.NotEmpty(t, bindings)
}
