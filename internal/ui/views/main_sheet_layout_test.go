package views

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Domo929/sheet/internal/domain"
	"github.com/Domo929/sheet/internal/models"
)

// findLineContaining returns the first ANSI-stripped rendered line that contains sub.
func findLineContaining(render, sub string) (string, bool) {
	for _, line := range strings.Split(render, "\n") {
		plain := ansi.Strip(line)
		if strings.Contains(plain, sub) {
			return plain, true
		}
	}
	return "", false
}

// maxLineWidth returns the widest ANSI-stripped rendered line.
func maxLineWidth(render string) int {
	max := 0
	for _, line := range strings.Split(render, "\n") {
		if w := ansi.StringWidth(line); w > max {
			max = w
		}
	}
	return max
}

// TestMainSheetAbilityRowsDoNotWrap guards the abilities/saves table against the
// regression where narrow two-column widths squeezed the left panel and either
// wrapped every ability row or truncated the save column. The test character has
// STR 16 with a +3 modifier and +3 save, so an intact row keeps both the score
// ("16") and the trailing save value ("○  +3") on the same line.
func TestMainSheetAbilityRowsDoNotWrap(t *testing.T) {
	for _, width := range []int{80, 90, 100, 120} {
		char := createTestCharacter() // STR 16
		model := NewMainSheetModel(char, nil)
		model, _ = model.Update(tea.WindowSizeMsg{Width: width, Height: 40})

		render := model.View()

		line, ok := findLineContaining(render, "Strength")
		require.True(t, ok, "width %d: expected a Strength row", width)

		// Isolate the abilities panel segment (between border pipes) so the right
		// column can't satisfy the assertion by accident.
		var segment string
		for _, seg := range strings.Split(line, "│") {
			if strings.Contains(seg, "Strength") {
				segment = seg
			}
		}
		require.NotEmpty(t, segment, "width %d: could not isolate abilities panel", width)
		assert.Contains(t, segment, "16",
			"width %d: Strength row lost its score (wrapped): %q", width, segment)
		assert.Regexp(t, `[○●]\s+\+3`, segment,
			"width %d: Strength row truncated the save column: %q", width, segment)

		assert.LessOrEqual(t, maxLineWidth(render), width,
			"width %d: a rendered line overflowed the terminal width", width)
	}
}

// TestMainSheetExhaustionDisplay verifies the leveled 2024 Exhaustion state is
// surfaced in the conditions panel with its mechanical penalty, and that the
// extra line never overflows the terminal at any supported width.
func TestMainSheetExhaustionDisplay(t *testing.T) {
	for _, width := range []int{80, 90, 100, 120} {
		char := createTestCharacter()
		char.CombatStats.ExhaustionLevel = 3

		model := NewMainSheetModel(char, nil)
		model, _ = model.Update(tea.WindowSizeMsg{Width: width, Height: 40})

		render := model.View()
		line, ok := findLineContaining(render, "Exhaustion 3")
		require.True(t, ok, "width %d: expected an Exhaustion level line", width)
		assert.Contains(t, line, "-6", "width %d: should show -2*level d20 penalty", width)

		assert.LessOrEqual(t, maxLineWidth(render), width,
			"width %d: a rendered line overflowed the terminal width", width)
	}
}
func TestMainSheetWeaponMasteryDisplay(t *testing.T) {
	for _, width := range []int{80, 90, 100, 120} {
		char := createTestCharacter() // Ranger L5, STR 16 (+3)
		char.Inventory.Items = append(char.Inventory.Items, models.Item{
			ID:         "longsword",
			Name:       "Longsword",
			Type:       models.ItemTypeWeapon,
			Quantity:   1,
			Damage:     "1d8",
			DamageType: domain.DamageSlashing,
			Mastery:    domain.MasterySap,
		})
		char.Inventory.Equipment.MainHand = &char.Inventory.Items[len(char.Inventory.Items)-1]

		model := NewMainSheetModel(char, nil)
		model, _ = model.Update(tea.WindowSizeMsg{Width: width, Height: 40})

		// The mastery label must reach the rendered attack description.
		var masteryShown bool
		for _, item := range model.getActionItems() {
			if item.Name == "Longsword" && strings.Contains(item.Description, "• Sap") {
				masteryShown = true
			}
		}
		require.True(t, masteryShown,
			"width %d: equipped Longsword should show its '• Sap' mastery in the actions panel", width)

		assert.LessOrEqual(t, maxLineWidth(model.View()), width,
			"width %d: a rendered line overflowed the terminal width", width)
	}
}
