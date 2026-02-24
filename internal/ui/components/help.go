package components

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// KeyBinding represents a keyboard shortcut and its description.
type KeyBinding struct {
	Key         string
	Description string
}

// HelpFooter displays keyboard shortcuts at the bottom of the screen.
type HelpFooter struct {
	Bindings []KeyBinding
	Width    int
}

// NewHelpFooter creates a new help footer.
func NewHelpFooter(bindings ...KeyBinding) HelpFooter {
	return HelpFooter{
		Bindings: bindings,
	}
}

// Render renders the help footer.
func (h HelpFooter) Render() string {
	if len(h.Bindings) == 0 {
		return ""
	}

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	var parts []string
	for _, binding := range h.Bindings {
		key := keyStyle.Render(binding.Key)
		desc := descStyle.Render(binding.Description)
		parts = append(parts, key+" "+desc)
	}

	content := strings.Join(parts, "  •  ")

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("236")).
		Padding(0, 1)

	if h.Width > 0 {
		style = style.Width(h.Width)
	}

	return style.Render(content)
}

// CommonBindings returns commonly used key bindings.
func CommonBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "↑/k", Description: "up"},
		{Key: "↓/j", Description: "down"},
		{Key: "←/h", Description: "left"},
		{Key: "→/l", Description: "right"},
		{Key: "enter", Description: "select"},
		{Key: "esc", Description: "back"},
		{Key: "q", Description: "quit"},
	}
}

// NavigationBindings returns navigation key bindings.
func NavigationBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "tab", Description: "next section"},
		{Key: "shift+tab", Description: "prev section"},
		{Key: "1-9", Description: "quick navigate"},
		{Key: "q", Description: "quit"},
	}
}

// ListBindings returns list navigation key bindings.
func ListBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "↑/k", Description: "up"},
		{Key: "↓/j", Description: "down"},
		{Key: "enter", Description: "select"},
		{Key: "esc/q", Description: "back"},
	}
}
