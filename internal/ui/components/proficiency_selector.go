package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProficiencySelector is a component for selecting proficiencies from a list.
type ProficiencySelector struct {
	title     string
	options   []string
	selected  map[int]bool
	maxSelect int
	cursor    int
	width     int
	height    int
	focused   bool

	// Styles
	titleStyle    lipgloss.Style
	optionStyle   lipgloss.Style
	selectedStyle lipgloss.Style
	cursorStyle   lipgloss.Style
	helpStyle     lipgloss.Style
}

// NewProficiencySelector creates a new proficiency selector.
func NewProficiencySelector(title string, options []string, maxSelect int) ProficiencySelector {
	return ProficiencySelector{
		title:         title,
		options:       options,
		selected:      make(map[int]bool),
		maxSelect:     maxSelect,
		cursor:        0,
		width:         60,
		height:        20,
		focused:       true,
		titleStyle:    lipgloss.NewStyle().Bold(true).Underline(true),
		optionStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		selectedStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("green")).Bold(true),
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
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if ps.cursor > 0 {
				ps.cursor--
			}
		case "down", "j":
			if ps.cursor < len(ps.options)-1 {
				ps.cursor++
			}
		case " ":
			// Toggle selection with space
			if ps.selected[ps.cursor] {
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
		if ps.selected[i] {
			checkbox = "[✓]"
			style = ps.selectedStyle
		}

		line := fmt.Sprintf("%s %s %s", cursor, checkbox, option)
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

// SetWidth sets the width of the component.
func (ps *ProficiencySelector) SetWidth(width int) {
	ps.width = width
}

// SetHeight sets the height of the component.
func (ps *ProficiencySelector) SetHeight(height int) {
	ps.height = height
}
