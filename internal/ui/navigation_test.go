package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	if len(km.Up) == 0 {
		t.Error("Up keys should not be empty")
	}
	if len(km.Down) == 0 {
		t.Error("Down keys should not be empty")
	}
	if len(km.Quit) == 0 {
		t.Error("Quit keys should not be empty")
	}
}

func TestIsKey(t *testing.T) {
	keys := []string{"up", "k"}

	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	if !IsKey(upMsg, keys) {
		t.Error("Should match 'up' key")
	}

	kMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	if !IsKey(kMsg, keys) {
		t.Error("Should match 'k' key")
	}

	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	if IsKey(downMsg, keys) {
		t.Error("Should not match 'down' key")
	}
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
			if result != tt.expected {
				t.Errorf("%s check returned %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestNavigationHandlerWrongKeys(t *testing.T) {
	nav := NewNavigationHandler()

	downMsg := tea.KeyMsg{Type: tea.KeyDown}

	// Down key should not match Up check
	if nav.IsUp(downMsg) {
		t.Error("Down key should not match IsUp")
	}

	// Down key should not match Left check
	if nav.IsLeft(downMsg) {
		t.Error("Down key should not match IsLeft")
	}

	// Down key should not match Quit check
	if nav.IsQuit(downMsg) {
		t.Error("Down key should not match IsQuit")
	}
}
