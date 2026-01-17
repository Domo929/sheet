package ui

import tea "github.com/charmbracelet/bubbletea"

// KeyMap defines common keyboard shortcuts.
type KeyMap struct {
	Up       []string
	Down     []string
	Left     []string
	Right    []string
	Select   []string
	Back     []string
	Quit     []string
	Help     []string
	Tab      []string
	ShiftTab []string
}

// DefaultKeyMap returns the default key mappings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:       []string{"up", "k"},
		Down:     []string{"down", "j"},
		Left:     []string{"left", "h"},
		Right:    []string{"right", "l"},
		Select:   []string{"enter", " "},
		Back:     []string{"esc"},
		Quit:     []string{"q", "ctrl+c"},
		Help:     []string{"?"},
		Tab:      []string{"tab"},
		ShiftTab: []string{"shift+tab"},
	}
}

// IsKey checks if a key message matches any of the given keys.
func IsKey(msg tea.KeyMsg, keys []string) bool {
	key := msg.String()
	for _, k := range keys {
		if key == k {
			return true
		}
	}
	return false
}

// HandleNavigation is a helper for handling common navigation keys.
type NavigationHandler struct {
	KeyMap KeyMap
}

// NewNavigationHandler creates a new navigation handler.
func NewNavigationHandler() NavigationHandler {
	return NavigationHandler{
		KeyMap: DefaultKeyMap(),
	}
}

// IsUp checks if the key is an up key.
func (n NavigationHandler) IsUp(msg tea.KeyMsg) bool {
	return IsKey(msg, n.KeyMap.Up)
}

// IsDown checks if the key is a down key.
func (n NavigationHandler) IsDown(msg tea.KeyMsg) bool {
	return IsKey(msg, n.KeyMap.Down)
}

// IsLeft checks if the key is a left key.
func (n NavigationHandler) IsLeft(msg tea.KeyMsg) bool {
	return IsKey(msg, n.KeyMap.Left)
}

// IsRight checks if the key is a right key.
func (n NavigationHandler) IsRight(msg tea.KeyMsg) bool {
	return IsKey(msg, n.KeyMap.Right)
}

// IsSelect checks if the key is a select key.
func (n NavigationHandler) IsSelect(msg tea.KeyMsg) bool {
	return IsKey(msg, n.KeyMap.Select)
}

// IsBack checks if the key is a back key.
func (n NavigationHandler) IsBack(msg tea.KeyMsg) bool {
	return IsKey(msg, n.KeyMap.Back)
}

// IsQuit checks if the key is a quit key.
func (n NavigationHandler) IsQuit(msg tea.KeyMsg) bool {
	return IsKey(msg, n.KeyMap.Quit)
}

// IsHelp checks if the key is a help key.
func (n NavigationHandler) IsHelp(msg tea.KeyMsg) bool {
	return IsKey(msg, n.KeyMap.Help)
}

// IsTab checks if the key is a tab key.
func (n NavigationHandler) IsTab(msg tea.KeyMsg) bool {
	return IsKey(msg, n.KeyMap.Tab)
}

// IsShiftTab checks if the key is a shift+tab key.
func (n NavigationHandler) IsShiftTab(msg tea.KeyMsg) bool {
	return IsKey(msg, n.KeyMap.ShiftTab)
}
