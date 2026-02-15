package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// RollType categorizes the purpose of a dice roll.
type RollType int

const (
	RollAttack RollType = iota
	RollDamage
	RollSkillCheck
	RollSavingThrow
	RollHitDice
	RollLuck
	RollCustom
)

// RollHistoryEntry records the result of a single dice roll.
type RollHistoryEntry struct {
	Label        string
	RollType     RollType
	Expression   string // e.g., "1d20+7"
	Rolls        []int  // individual dice results
	Kept         []int  // dice kept (for advantage/disadvantage)
	Dropped      []int  // dice dropped
	Modifier     int
	Total        int
	Advantage    bool
	Disadvantage bool
	NatCrit      bool // natural 20 on a d20
	NatFail      bool // natural 1 on a d20
	Timestamp    time.Time
}

const maxHistoryEntries = 50

// RollHistory stores a capped list of recent dice rolls.
type RollHistory struct {
	Entries   []RollHistoryEntry
	Visible   bool // whether the history column is shown
	ScrollPos int  // scroll position for viewing history
}

// NewRollHistory creates an empty roll history.
func NewRollHistory() *RollHistory {
	return &RollHistory{
		Entries: make([]RollHistoryEntry, 0),
	}
}

// Add appends a new roll to the history, capping at maxHistoryEntries.
// The first roll added automatically makes the history visible.
func (h *RollHistory) Add(entry RollHistoryEntry) {
	entry.Timestamp = time.Now()
	h.Entries = append([]RollHistoryEntry{entry}, h.Entries...)
	if len(h.Entries) > maxHistoryEntries {
		h.Entries = h.Entries[:maxHistoryEntries]
	}
	if !h.Visible {
		h.Visible = true
	}
}

// Toggle switches the history column visibility.
func (h *RollHistory) Toggle() {
	h.Visible = !h.Visible
}

// Clear removes all entries.
func (h *RollHistory) Clear() {
	h.Entries = h.Entries[:0]
	h.Visible = false
}

// RollTypeIcon returns the icon for a given roll type.
func RollTypeIcon(rt RollType) string {
	switch rt {
	case RollAttack:
		return "⚔"
	case RollDamage:
		return "💥"
	case RollSkillCheck:
		return "🎯"
	case RollSavingThrow:
		return "🛡"
	case RollHitDice:
		return "❤"
	case RollLuck:
		return "🎲"
	case RollCustom:
		return "🎲"
	default:
		return "🎲"
	}
}

// Render renders the roll history column at the given width and height.
func (h *RollHistory) Render(width, height int) string {
	if !h.Visible || width < 20 {
		return ""
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	natCritStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	natFailStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	luckStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	advStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))

	innerWidth := width - 4 // account for border padding

	var lines []string
	for _, entry := range h.Entries {
		icon := RollTypeIcon(entry.RollType)
		labelLine := fmt.Sprintf("%s %s", icon, entry.Label)
		if len(labelLine) > innerWidth {
			labelLine = labelLine[:innerWidth]
		}

		if entry.RollType == RollLuck {
			lines = append(lines, luckStyle.Render(labelLine))
		} else {
			lines = append(lines, titleStyle.Render(labelLine))
		}

		// Detail line: expression → total
		detail := fmt.Sprintf("  %s → %d", entry.Expression, entry.Total)

		// Add advantage/disadvantage indicator
		if entry.Advantage {
			detail += " " + advStyle.Render("(ADV)")
		} else if entry.Disadvantage {
			detail += " " + advStyle.Render("(DIS)")
		}

		// Apply nat crit/fail styling
		if entry.NatCrit {
			lines = append(lines, natCritStyle.Render(detail))
		} else if entry.NatFail {
			lines = append(lines, natFailStyle.Render(detail))
		} else if entry.RollType == RollLuck {
			lines = append(lines, luckStyle.Render(detail))
		} else {
			lines = append(lines, dimStyle.Render(detail))
		}

		lines = append(lines, "") // blank separator
	}

	// Truncate to fit height
	maxLines := height - 3 // account for border top/bottom + title
	if len(lines) > maxLines && maxLines > 0 {
		lines = lines[:maxLines]
	}

	content := strings.Join(lines, "\n")

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(width - 2).
		Padding(0, 1)

	header := titleStyle.Render("Roll History")
	return panelStyle.Render(header + "\n" + content)
}
