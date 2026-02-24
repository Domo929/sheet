package components

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ProficiencySelector is a component for selecting proficiencies from a list.
type ProficiencySelector struct {
	title     string
	options   []string
	selected  map[int]bool
	locked    map[int]bool   // Items that are pre-selected and cannot be deselected
	labels    map[int]string // Labels to show next to locked items (e.g., "(From Acolyte)")
	maxSelect int
	cursor    int
	width     int
	height    int
	focused   bool

	// Styles
	titleStyle    lipgloss.Style
	optionStyle   lipgloss.Style
	selectedStyle lipgloss.Style
	lockedStyle   lipgloss.Style
	cursorStyle   lipgloss.Style
	helpStyle     lipgloss.Style
}

// NewProficiencySelector creates a new proficiency selector.
func NewProficiencySelector(title string, options []string, maxSelect int) ProficiencySelector {
	return ProficiencySelector{
		title:         title,
		options:       options,
		selected:      make(map[int]bool),
		locked:        make(map[int]bool),
		labels:        make(map[int]string),
		maxSelect:     maxSelect,
		cursor:        0,
		width:         60,
		height:        20,
		focused:       true,
		titleStyle:    lipgloss.NewStyle().Bold(true).UnderlineStyle(lipgloss.UnderlineCurly).UnderlineColor(lipgloss.Color("99")),
		optionStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		selectedStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("green")).Bold(true),
		lockedStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("cyan")),
		cursorStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("yellow")),
		helpStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true),
	}
}

// Update handles messages for the proficiency selector.
func (ps *ProficiencySelector) Update(msg tea.Msg) (ProficiencySelector, tea.Cmd) {
	if !ps.focused {
		return *ps, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Scale width to 80% of terminal width, with min 20 and max 80
		width := msg.Width * 4 / 5
		if width > 80 {
			width = 80
		}
		if width < 20 {
			width = 20
		}
		ps.width = width

		// Scale height to 60% of terminal height, with min 10 and max 30
		height := msg.Height * 3 / 5
		if height > 30 {
			height = 30
		}
		if height < 10 {
			height = 10
		}
		ps.height = height

	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if ps.cursor > 0 {
				ps.cursor--
			}
		case "down", "j":
			if ps.cursor < len(ps.options)-1 {
				ps.cursor++
			}
		case "space":
			// Toggle selection with space (but not for locked items)
			if ps.locked[ps.cursor] {
				// Cannot toggle locked items
			} else if ps.selected[ps.cursor] {
				delete(ps.selected, ps.cursor)
			} else if ps.SelectedCount() < ps.maxSelect {
				ps.selected[ps.cursor] = true
			}
		}
	}

	return *ps, nil
}

// View renders the proficiency selector.
func (ps ProficiencySelector) View() string {
	var b strings.Builder

	// Title
	b.WriteString(ps.titleStyle.Render(ps.title))
	b.WriteString("\n")

	// Selection count
	selectedCount := ps.SelectedCount()
	countText := fmt.Sprintf("(%d/%d selected)", selectedCount, ps.maxSelect)
	if selectedCount >= ps.maxSelect {
		countText = ps.selectedStyle.Render(countText)
	} else {
		countText = ps.helpStyle.Render(countText)
	}
	b.WriteString(countText)
	b.WriteString("\n\n")

	// Options
	for i, option := range ps.options {
		cursor := "  "
		if ps.focused && i == ps.cursor {
			cursor = ps.cursorStyle.Render("> ")
		}

		checkbox := "[ ]"
		style := ps.optionStyle
		suffix := ""
		if ps.locked[i] {
			checkbox = "[✓]"
			style = ps.lockedStyle
			if label, ok := ps.labels[i]; ok {
				suffix = " " + label
			}
		} else if ps.selected[i] {
			checkbox = "[✓]"
			style = ps.selectedStyle
		}

		line := fmt.Sprintf("%s %s %s%s", cursor, checkbox, option, suffix)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	// Help text
	b.WriteString("\n")
	if ps.focused {
		help := "↑/↓: navigate • space: select/deselect • tab: next section"
		b.WriteString(ps.helpStyle.Render(help))
	}

	return b.String()
}

// ViewWithoutHelp renders the selector without the help text at the bottom.
func (ps ProficiencySelector) ViewWithoutHelp() string {
	var b strings.Builder

	// Title
	b.WriteString(ps.titleStyle.Render(ps.title))
	b.WriteString("\n")

	// Selection count
	selectedCount := ps.SelectedCount()
	countText := fmt.Sprintf("(%d/%d selected)", selectedCount, ps.maxSelect)
	if selectedCount >= ps.maxSelect {
		countText = ps.selectedStyle.Render(countText)
	} else {
		countText = ps.helpStyle.Render(countText)
	}
	b.WriteString(countText)
	b.WriteString("\n\n")

	// Options
	for i, option := range ps.options {
		cursor := "  "
		if ps.focused && i == ps.cursor {
			cursor = ps.cursorStyle.Render("> ")
		}

		checkbox := "[ ]"
		style := ps.optionStyle
		suffix := ""
		if ps.locked[i] {
			checkbox = "[✓]"
			style = ps.lockedStyle
			if label, ok := ps.labels[i]; ok {
				suffix = " " + label
			}
		} else if ps.selected[i] {
			checkbox = "[✓]"
			style = ps.selectedStyle
		}

		line := fmt.Sprintf("%s %s %s%s", cursor, checkbox, option, suffix)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	return b.String()
}

// HelpText returns the help text for the selector.
func (ps ProficiencySelector) HelpText() string {
	if ps.focused {
		return ps.helpStyle.Render("↑/↓: navigate • space: select/deselect • tab: next section")
	}
	return ""
}

// SetFocused sets the focused state of the selector.
func (ps *ProficiencySelector) SetFocused(focused bool) {
	ps.focused = focused
}

// IsFocused returns whether the selector is focused.
func (ps ProficiencySelector) IsFocused() bool {
	return ps.focused
}

// SelectedCount returns the number of selected items.
func (ps ProficiencySelector) SelectedCount() int {
	return len(ps.selected)
}

// IsComplete returns whether the selection is complete (reached maxSelect).
func (ps ProficiencySelector) IsComplete() bool {
	return ps.SelectedCount() >= ps.maxSelect
}

// GetSelectedIndices returns the indices of selected items.
func (ps ProficiencySelector) GetSelectedIndices() []int {
	indices := make([]int, 0, len(ps.selected))
	for i := range ps.selected {
		indices = append(indices, i)
	}
	return indices
}

// GetSelectedOptions returns the selected option strings.
func (ps ProficiencySelector) GetSelectedOptions() []string {
	options := make([]string, 0, len(ps.selected))
	for i := range ps.selected {
		options = append(options, ps.options[i])
	}
	return options
}

// SetSelected sets the selected indices.
func (ps *ProficiencySelector) SetSelected(indices []int) {
	ps.selected = make(map[int]bool)
	for _, i := range indices {
		if i >= 0 && i < len(ps.options) {
			ps.selected[i] = true
		}
	}
}

// SetLocked sets options as locked (pre-selected, non-toggleable) with a label.
func (ps *ProficiencySelector) SetLocked(optionNames []string, label string) {
	for i, opt := range ps.options {
		for _, locked := range optionNames {
			if strings.EqualFold(opt, locked) {
				ps.locked[i] = true
				ps.labels[i] = label
				break
			}
		}
	}
}

// GetLockedOptions returns the locked option strings.
func (ps ProficiencySelector) GetLockedOptions() []string {
	options := make([]string, 0, len(ps.locked))
	for i := range ps.locked {
		options = append(options, ps.options[i])
	}
	return options
}

// SetWidth sets the width of the component.
func (ps *ProficiencySelector) SetWidth(width int) {
	ps.width = width
}

// SetHeight sets the height of the component.
func (ps *ProficiencySelector) SetHeight(height int) {
	ps.height = height
}
