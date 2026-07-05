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
	"github.com/Domo929/sheet/internal/ui/components"
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

		// Until the weapon is chosen for Weapon Mastery, its mastery property
		// must NOT appear in the attack description (2024 rules: mastery applies
		// only to weapons you have selected).
		model := NewMainSheetModel(char, nil)
		model, _ = model.Update(tea.WindowSizeMsg{Width: width, Height: 40})
		for _, item := range model.getActionItems() {
			if item.Name == "Longsword" {
				assert.NotContains(t, item.Description, "• Sap",
					"width %d: unmastered Longsword should not show its mastery", width)
			}
		}

		// After choosing the weapon for mastery, the label must reach the
		// rendered attack description.
		char.MasteredWeapons = []string{"Longsword"}
		model = NewMainSheetModel(char, nil)
		model, _ = model.Update(tea.WindowSizeMsg{Width: width, Height: 40})

		var masteryShown bool
		for _, item := range model.getActionItems() {
			if item.Name == "Longsword" && strings.Contains(item.Description, "• Sap") {
				masteryShown = true
			}
		}
		require.True(t, masteryShown,
			"width %d: mastered Longsword should show its '• Sap' mastery in the actions panel", width)

		assert.LessOrEqual(t, maxLineWidth(model.View()), width,
			"width %d: a rendered line overflowed the terminal width", width)
	}
}

func TestWeaponMasterySelectionFlow(t *testing.T) {
	char := models.NewCharacter("fid", "Boromir", "Human", "Fighter") // limit 3 at L1
	char.Inventory.Items = append(char.Inventory.Items, models.Item{
		ID:         "greataxe",
		Name:       "Greataxe",
		Type:       models.ItemTypeWeapon,
		Quantity:   1,
		Damage:     "1d12",
		DamageType: domain.DamageSlashing,
		Mastery:    domain.MasteryCleave,
	})
	char.Inventory.Equipment.MainHand = &char.Inventory.Items[len(char.Inventory.Items)-1]

	model := NewMainSheetModel(char, nil)
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Focus the Actions panel and open the mastery selector with "m".
	model.focusArea = FocusActions
	model, _ = model.Update(tea.KeyPressMsg{Code: 'm', Text: "m"})
	require.True(t, model.masteryMode, "pressing m in the Actions panel should open the mastery selector")

	view := model.View()
	if _, ok := findLineContaining(view, "Weapon Mastery"); !ok {
		t.Fatal("mastery overlay header should render")
	}
	line, ok := findLineContaining(view, "] Greataxe")
	require.True(t, ok, "mastery overlay should list the Greataxe with a checkbox")
	assert.Contains(t, line, "[ ]", "weapon starts unmastered")

	// Toggle mastery on.
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.True(t, model.character.HasWeaponMastery("Greataxe"))
	line, ok = findLineContaining(model.View(), "] Greataxe")
	require.True(t, ok)
	assert.Contains(t, line, "[x]", "weapon should now be mastered")

	// Close the selector.
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.False(t, model.masteryMode)

	// Mastery now shows in the attack line.
	var shown bool
	for _, item := range model.getActionItems() {
		if item.Name == "Greataxe" && strings.Contains(item.Description, "• Cleave") {
			shown = true
		}
	}
	assert.True(t, shown, "a mastered weapon shows its property in the actions panel")

	assert.LessOrEqual(t, maxLineWidth(model.View()), 100)
}

func TestDefensesPanelRender(t *testing.T) {
	char := models.NewCharacter("tid", "Zar", "Dragonborn", "Fighter")
	char.Info.Subrace = "Red"
	char.Features.AddRacialTrait("Draconic Ancestry", "Dragonborn",
		"You have Resistance to the damage type determined by your Draconic Ancestry.")

	model := NewMainSheetModel(char, nil)
	for _, width := range []int{90, 100, 120} {
		model, _ = model.Update(tea.WindowSizeMsg{Width: width, Height: 40})
		view := model.View()
		line, ok := findLineContaining(view, "Resist:")
		require.True(t, ok, "width %d: the Defenses panel should show a Resist line", width)
		assert.Contains(t, line, "Fire", "width %d: Dragonborn (Red) should resist Fire", width)
		assert.LessOrEqual(t, maxLineWidth(view), width, "width %d: a rendered line overflowed", width)
	}
}

func TestClassResourceOverlayFlow(t *testing.T) {
	char := models.NewCharacter("bid", "Grog", "Human", "Barbarian")
	char.Info.Level = 6 // Rage 4/4

	model := NewMainSheetModel(char, nil) // constructor syncs resources
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Resources should render in the Combat panel.
	line, ok := findLineContaining(model.View(), "Rage:")
	require.True(t, ok, "Combat panel should list the Rage resource")
	assert.Contains(t, line, "4/4")

	// Open the resource overlay from the Combat panel and spend one Rage.
	model.focusArea = FocusCombat
	model, _ = model.Update(tea.KeyPressMsg{Code: 'u', Text: "u"})
	require.True(t, model.resourceMode, "pressing u in Combat should open the resource overlay")

	model, _ = model.Update(tea.KeyPressMsg{Code: '-', Text: "-"})
	assert.Equal(t, 3, model.character.Resources[0].Current, "spending should decrement Rage")

	// Restore one and close.
	model, _ = model.Update(tea.KeyPressMsg{Code: '+', Text: "+"})
	assert.Equal(t, 4, model.character.Resources[0].Current)
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.False(t, model.resourceMode)

	assert.LessOrEqual(t, maxLineWidth(model.View()), 100)
}

func TestConcentrationDisplayAndManualBreak(t *testing.T) {
	char := models.NewCharacter("cid", "Wren", "Human", "Wizard")
	char.Info.Level = 5
	char.StartConcentration("Bless")

	model := NewMainSheetModel(char, nil)
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	line, ok := findLineContaining(model.View(), "Concentration")
	require.True(t, ok, "Combat panel should show a Concentration section")
	assert.NotEmpty(t, line)
	_, ok = findLineContaining(model.View(), "Bless")
	assert.True(t, ok, "should display the concentrated spell")

	// Manual break with capital C in the Combat panel.
	model.focusArea = FocusCombat
	model, _ = model.Update(tea.KeyPressMsg{Code: 'C', Text: "C"})
	assert.False(t, model.character.IsConcentrating(), "C should end concentration")
}

func TestConcentrationSaveOnDamage(t *testing.T) {
	char := models.NewCharacter("cid", "Wren", "Human", "Wizard")
	char.Info.Level = 5
	char.CombatStats.HitPoints.Maximum = 40
	char.CombatStats.HitPoints.Current = 40
	char.StartConcentration("Bless")

	model := NewMainSheetModel(char, nil)
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	model.focusArea = FocusCombat

	// Take 20 damage -> DC 10 CON save prompted.
	model, _ = model.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
	model, _ = model.Update(tea.KeyPressMsg{Code: '2', Text: "2"})
	model, _ = model.Update(tea.KeyPressMsg{Code: '0', Text: "0"})
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd, "damage while concentrating should prompt a save roll")
	assert.Equal(t, 10, model.concentrationSaveDC)
	assert.True(t, model.character.IsConcentrating(), "still concentrating until the save resolves")

	// A failing save breaks concentration.
	model, _ = model.Update(components.RollCompleteMsg{Entry: components.RollHistoryEntry{
		Label: "Concentration: Bless (DC 10)", Total: 7,
	}})
	assert.False(t, model.character.IsConcentrating(), "failed save breaks concentration")

	// Re-establish and pass a save: concentration holds.
	model.character.StartConcentration("Bless")
	model.concentrationSaveDC = 10
	model.concentrationSaveSpell = "Bless"
	model, _ = model.Update(components.RollCompleteMsg{Entry: components.RollHistoryEntry{
		Label: "Concentration: Bless (DC 10)", Total: 18,
	}})
	assert.True(t, model.character.IsConcentrating(), "passed save keeps concentration")
}

func TestConcentrationBreaksAtZeroHP(t *testing.T) {
	char := models.NewCharacter("cid", "Wren", "Human", "Wizard")
	char.Info.Level = 5
	char.CombatStats.HitPoints.Maximum = 10
	char.CombatStats.HitPoints.Current = 6
	char.StartConcentration("Bless")

	model := NewMainSheetModel(char, nil)
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	model.focusArea = FocusCombat

	// Massive damage drops to 0 -> concentration broken outright, no save roll.
	model, _ = model.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
	model, _ = model.Update(tea.KeyPressMsg{Code: '9', Text: "9"})
	model, _ = model.Update(tea.KeyPressMsg{Code: '9', Text: "9"})
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Nil(t, cmd, "no save roll when dropping to 0 HP")
	assert.False(t, model.character.IsConcentrating())
	assert.Equal(t, 0, model.character.CombatStats.HitPoints.Current)
}
