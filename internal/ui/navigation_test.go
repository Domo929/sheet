package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	assert.NotEmpty(t, km.Up, "Up keys should not be empty")
	assert.NotEmpty(t, km.Down, "Down keys should not be empty")
	assert.NotEmpty(t, km.Quit, "Quit keys should not be empty")
}

func TestIsKey(t *testing.T) {
	keys := []string{"up", "k"}

	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	assert.True(t, IsKey(upMsg, keys), "Should match 'up' key")

	kMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	assert.True(t, IsKey(kMsg, keys), "Should match 'k' key")

	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	assert.False(t, IsKey(downMsg, keys), "Should not match 'down' key")
}

func TestNavigationHandler(t *testing.T) {
	nav := NewNavigationHandler()

	tests := []struct {
		name     string
		msg      tea.KeyMsg
		checkFn  func(tea.KeyMsg) bool
		expected bool
	}{
		{
			name:     "up key",
			msg:      tea.KeyMsg{Type: tea.KeyUp},
			checkFn:  nav.IsUp,
			expected: true,
		},
		{
			name:     "k key (vim up)",
			msg:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			checkFn:  nav.IsUp,
			expected: true,
		},
		{
			name:     "down key",
			msg:      tea.KeyMsg{Type: tea.KeyDown},
			checkFn:  nav.IsDown,
			expected: true,
		},
		{
			name:     "j key (vim down)",
			msg:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			checkFn:  nav.IsDown,
			expected: true,
		},
		{
			name:     "left key",
			msg:      tea.KeyMsg{Type: tea.KeyLeft},
			checkFn:  nav.IsLeft,
			expected: true,
		},
		{
			name:     "right key",
			msg:      tea.KeyMsg{Type: tea.KeyRight},
			checkFn:  nav.IsRight,
			expected: true,
		},
		{
			name:     "enter key",
			msg:      tea.KeyMsg{Type: tea.KeyEnter},
			checkFn:  nav.IsSelect,
			expected: true,
		},
		{
			name:     "escape key",
			msg:      tea.KeyMsg{Type: tea.KeyEsc},
			checkFn:  nav.IsBack,
			expected: true,
		},
		{
			name:     "q key",
			msg:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			checkFn:  nav.IsQuit,
			expected: true,
		},
		{
			name:     "tab key",
			msg:      tea.KeyMsg{Type: tea.KeyTab},
			checkFn:  nav.IsTab,
			expected: true,
		},
		{
			name:     "shift+tab key",
			msg:      tea.KeyMsg{Type: tea.KeyShiftTab},
			checkFn:  nav.IsShiftTab,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.checkFn(tt.msg)
			assert.Equal(t, tt.expected, result, "%s check returned unexpected result", tt.name)
		})
	}
}

func TestNavigationHandlerWrongKeys(t *testing.T) {
	nav := NewNavigationHandler()

	downMsg := tea.KeyMsg{Type: tea.KeyDown}

	// Down key should not match Up check
	assert.False(t, nav.IsUp(downMsg), "Down key should not match IsUp")

	// Down key should not match Left check
	assert.False(t, nav.IsLeft(downMsg), "Down key should not match IsLeft")

	// Down key should not match Quit check
	assert.False(t, nav.IsQuit(downMsg), "Down key should not match IsQuit")
}
