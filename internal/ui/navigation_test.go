package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"

	tea "charm.land/bubbletea/v2"
)

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	assert.NotEmpty(t, km.Up, "Up keys should not be empty")
	assert.NotEmpty(t, km.Down, "Down keys should not be empty")
	assert.NotEmpty(t, km.Quit, "Quit keys should not be empty")
}

func TestIsKey(t *testing.T) {
	keys := []string{"up", "k"}

	upMsg := tea.KeyPressMsg{Code: tea.KeyUp}
	assert.True(t, IsKey(upMsg, keys), "Should match 'up' key")

	kMsg := tea.KeyPressMsg{Code: 'k', Text: "k"}
	assert.True(t, IsKey(kMsg, keys), "Should match 'k' key")

	downMsg := tea.KeyPressMsg{Code: tea.KeyDown}
	assert.False(t, IsKey(downMsg, keys), "Should not match 'down' key")
}

func TestNavigationHandler(t *testing.T) {
	nav := NewNavigationHandler()

	tests := []struct {
		name     string
		msg      tea.KeyPressMsg
		checkFn  func(tea.KeyPressMsg) bool
		expected bool
	}{
		{
			name:     "up key",
			msg:      tea.KeyPressMsg{Code: tea.KeyUp},
			checkFn:  nav.IsUp,
			expected: true,
		},
		{
			name:     "k key (vim up)",
			msg:      tea.KeyPressMsg{Code: 'k', Text: "k"},
			checkFn:  nav.IsUp,
			expected: true,
		},
		{
			name:     "down key",
			msg:      tea.KeyPressMsg{Code: tea.KeyDown},
			checkFn:  nav.IsDown,
			expected: true,
		},
		{
			name:     "j key (vim down)",
			msg:      tea.KeyPressMsg{Code: 'j', Text: "j"},
			checkFn:  nav.IsDown,
			expected: true,
		},
		{
			name:     "left key",
			msg:      tea.KeyPressMsg{Code: tea.KeyLeft},
			checkFn:  nav.IsLeft,
			expected: true,
		},
		{
			name:     "right key",
			msg:      tea.KeyPressMsg{Code: tea.KeyRight},
			checkFn:  nav.IsRight,
			expected: true,
		},
		{
			name:     "enter key",
			msg:      tea.KeyPressMsg{Code: tea.KeyEnter},
			checkFn:  nav.IsSelect,
			expected: true,
		},
		{
			name:     "escape key",
			msg:      tea.KeyPressMsg{Code: tea.KeyEscape},
			checkFn:  nav.IsBack,
			expected: true,
		},
		{
			name:     "q key",
			msg:      tea.KeyPressMsg{Code: 'q', Text: "q"},
			checkFn:  nav.IsQuit,
			expected: true,
		},
		{
			name:     "tab key",
			msg:      tea.KeyPressMsg{Code: tea.KeyTab},
			checkFn:  nav.IsTab,
			expected: true,
		},
		{
			name:     "shift+tab key",
			msg:      tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift},
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

	downMsg := tea.KeyPressMsg{Code: tea.KeyDown}

	// Down key should not match Up check
	assert.False(t, nav.IsUp(downMsg), "Down key should not match IsUp")

	// Down key should not match Left check
	assert.False(t, nav.IsLeft(downMsg), "Down key should not match IsLeft")

	// Down key should not match Quit check
	assert.False(t, nav.IsQuit(downMsg), "Down key should not match IsQuit")
}
