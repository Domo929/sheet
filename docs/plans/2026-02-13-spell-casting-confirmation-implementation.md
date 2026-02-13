# Spell Casting Confirmation Modal Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add casting confirmation modal with complete spell information (damage, save DC, description) before any spell is cast, and enhance spell details display with damage/save information.

**Architecture:** Add new mode `ModeConfirmCast` to spellbook state machine. When user presses 'c', enter modal mode and display spell information with slot selection overlay. Enhance spell details rendering to show damage and saving throw fields. Make header stats bold for visibility.

**Tech Stack:** Go, Bubble Tea (Charmbracelet), lipgloss for styling

---

## Task 1: Add Damage and Saving Throw to Spell Details Panel

**Files:**
- Modify: `internal/ui/views/spellbook.go` (renderSpellDetails function around line 457)
- Test: Manual testing with existing character

**Step 1: Add helper function to calculate spell save DC**

Add this function before `renderSpellDetails`:

```go
// getSpellSaveDC calculates the spell save DC for the character.
func (m *SpellbookModel) getSpellSaveDC() int {
	if m.character == nil || m.character.Spellcasting == nil {
		return 10
	}

	sc := m.character.Spellcasting
	abilityMod := m.character.GetAbilityModifier(sc.Ability)
	profBonus := m.character.Info.ProficiencyBonus

	return models.CalculateSpellSaveDC(abilityMod, profBonus)
}
```

**Step 2: Update renderSpellDetails to show damage and save**

Modify the `renderSpellDetails` function to add damage and saving throw lines after the Duration line (around line 480):

```go
lines = append(lines, fmt.Sprintf("Casting Time: %s", spell.CastingTime))
lines = append(lines, fmt.Sprintf("Range: %s", spell.Range))
lines = append(lines, fmt.Sprintf("Components: %s", strings.Join(spell.Components, ", ")))
lines = append(lines, fmt.Sprintf("Duration: %s", spell.Duration))

// Add damage if present
if spell.Damage != "" {
	damageInfo := spell.Damage
	if spell.DamageType != "" {
		damageInfo = fmt.Sprintf("%s %s", damageInfo, spell.DamageType)
	}
	lines = append(lines, fmt.Sprintf("Damage: %s", damageInfo))
}

// Add saving throw if present
if spell.SavingThrow != "" {
	saveDC := m.getSpellSaveDC()
	lines = append(lines, fmt.Sprintf("Saving Throw: %s DC %d", spell.SavingThrow, saveDC))
}

lines = append(lines, "")
```

**Step 3: Test the changes**

Run the application and navigate to spellbook:
```bash
go run ./cmd/sheet
# Load character with spells (e.g., Elara)
# Press 's' to open spellbook
# Select Burning Hands - should see "Damage: 3d6 fire" and "Saving Throw: Dexterity DC 15"
```

**Step 4: Commit**

```bash
git add internal/ui/views/spellbook.go
git commit -m "feat(spellbook): add damage and saving throw to spell details

- Add getSpellSaveDC() helper to calculate DC
- Show damage with type in spell details panel
- Show saving throw with calculated DC
- Display between Duration and Description"
```

---

## Task 2: Make Header Statistics Bold and Prominent

**Files:**
- Modify: `internal/ui/views/spellbook.go` (View function around line 120)

**Step 1: Find header rendering code**

Search for where the spellcasting ability/DC/attack stats are displayed in the View function. It should be near the top of the rendered output.

**Step 2: Apply bold styling to header stats**

Find the line(s) that render the header with spellcasting stats and apply bold styling:

```go
// Example of what to look for and modify:
// Before:
headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

// After:
headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255"))
```

Or if stats are rendered individually, wrap each in bold:

```go
abilityText := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Spellcasting: %s", sc.Ability))
dcText := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("DC %d", saveDC))
attackText := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Attack +%d", attackBonus))
preparedText := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Prepared: %d/%d", numPrepared, sc.MaxPrepared))
```

**Step 3: Test header visibility**

```bash
go run ./cmd/sheet
# Load character with spellcasting
# Press 's' for spellbook
# Verify header stats are bold and easy to read
```

**Step 4: Commit**

```bash
git add internal/ui/views/spellbook.go
git commit -m "feat(spellbook): make header statistics bold and prominent

- Apply bold styling to spellcasting ability, DC, attack, and prepared count
- Improve readability of key stats in header"
```

---

## Task 3: Add ModeConfirmCast Mode Constant

**Files:**
- Modify: `internal/ui/views/spellbook.go` (SpellbookMode enum around line 18)

**Step 1: Add new mode constant**

Add `ModeConfirmCast` to the SpellbookMode enum:

```go
type SpellbookMode int

const (
	ModeSpellList SpellbookMode = iota // Viewing/casting spells
	ModePreparation                     // Preparing/unpreparing spells
	ModeAddSpell                        // Adding a new spell
	ModeSelectCastLevel                 // Selecting spell slot level for casting
	ModeConfirmCast                     // Confirming spell cast with details
)
```

**Step 2: Commit**

```bash
git add internal/ui/views/spellbook.go
git commit -m "feat(spellbook): add ModeConfirmCast mode constant

- Add new mode for casting confirmation modal
- Will replace ModeSelectCastLevel flow"
```

---

## Task 4: Create Casting Confirmation Modal Rendering Function

**Files:**
- Modify: `internal/ui/views/spellbook.go` (add new function after renderCastLevelOverlay)

**Step 1: Create renderCastConfirmationModal function**

Add this function after the existing `renderCastLevelOverlay` function (around line 1213):

```go
// renderCastConfirmationModal renders the spell casting confirmation modal.
func (m *SpellbookModel) renderCastConfirmationModal() string {
	if m.castingSpell == nil || m.selectedSpellData == nil {
		return ""
	}

	spell := m.selectedSpellData
	var lines []string

	// Title
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Cast %s", spell.Name)))
	lines = append(lines, "")

	// Level and school
	levelSchool := fmt.Sprintf("Level %d %s", spell.Level, spell.School)
	if spell.Level == 0 {
		levelSchool = fmt.Sprintf("%s cantrip", spell.School)
	}
	if spell.Ritual {
		levelSchool += " (ritual)"
	}
	lines = append(lines, levelSchool)
	lines = append(lines, "")

	// Basic spell info
	lines = append(lines, fmt.Sprintf("Casting Time: %s", spell.CastingTime))
	lines = append(lines, fmt.Sprintf("Range: %s", spell.Range))
	lines = append(lines, fmt.Sprintf("Components: %s", strings.Join(spell.Components, ", ")))
	lines = append(lines, fmt.Sprintf("Duration: %s", spell.Duration))

	// Damage if present
	if spell.Damage != "" {
		damageInfo := spell.Damage
		if spell.DamageType != "" {
			damageInfo = fmt.Sprintf("%s %s", damageInfo, spell.DamageType)
		}
		lines = append(lines, fmt.Sprintf("Damage: %s", damageInfo))
	}

	// Saving throw if present
	if spell.SavingThrow != "" {
		saveDC := m.getSpellSaveDC()
		lines = append(lines, fmt.Sprintf("Saving Throw: %s DC %d", spell.SavingThrow, saveDC))
	}

	lines = append(lines, "")

	// Description (word-wrapped)
	descLines := m.wordWrap(spell.Description, 60)
	lines = append(lines, descLines...)

	// Upcast information if available
	if spell.Upcast != "" && spell.Level > 0 && spell.Level < 9 {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Bold(true).Render("At Higher Levels:"))
		upcastLines := m.wordWrap(spell.Upcast, 60)
		lines = append(lines, upcastLines...)
	}

	// Separator before slot selection
	if m.castingSpell.Level > 0 && len(m.availableCastLevels) > 0 {
		lines = append(lines, "")
		lines = append(lines, strings.Repeat("─", 60))
		lines = append(lines, "")
	}

	// Slot selection (if not a cantrip)
	if m.castingSpell.Level > 0 {
		if len(m.availableCastLevels) == 0 {
			lines = append(lines, "No spell slots available")
		} else if len(m.availableCastLevels) == 1 {
			// Single slot option
			level := m.availableCastLevels[0]
			upcastInfo := m.calculateUpcastEffect(level)
			sc := m.character.Spellcasting

			isPactMagic := sc.PactMagic != nil && sc.PactMagic.SlotLevel == level
			if isPactMagic {
				lines = append(lines, fmt.Sprintf("Using: Pact Magic - Level %d%s", level, upcastInfo))
			} else {
				lines = append(lines, fmt.Sprintf("Using: Level %d Slot%s", level, upcastInfo))
			}
		} else {
			// Multiple slot options
			lines = append(lines, "Select Spell Slot Level:")
			lines = append(lines, "")

			sc := m.character.Spellcasting
			for i, level := range m.availableCastLevels {
				cursor := "  "
				if i == m.castLevelCursor {
					cursor = "> "
				}

				isPactMagic := sc.PactMagic != nil && sc.PactMagic.SlotLevel == level
				upcastInfo := m.calculateUpcastEffect(level)

				var line string
				if isPactMagic {
					remaining := sc.PactMagic.Remaining
					total := sc.PactMagic.Total
					line = fmt.Sprintf("%sPact Magic - Level %d%s [%d/%d remaining]", cursor, level, upcastInfo, remaining, total)
				} else {
					slot := sc.SpellSlots.GetSlot(level)
					if slot != nil {
						line = fmt.Sprintf("%sLevel %d Slot%s [%d/%d remaining]", cursor, level, upcastInfo, slot.Remaining, slot.Total)
					}
				}

				if i == m.castLevelCursor {
					line = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(line)
				}

				lines = append(lines, line)
			}
		}
	}

	// Help text
	lines = append(lines, "")
	if m.castingSpell.Level == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Enter: cast | Esc: cancel"))
	} else if len(m.availableCastLevels) <= 1 {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Enter: cast | Esc: cancel"))
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("↑↓: select slot | Enter: cast | Esc: cancel"))
	}

	content := strings.Join(lines, "\n")

	width := 70
	height := len(lines) + 2

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Render(content)
}
```

**Step 2: Commit**

```bash
git add internal/ui/views/spellbook.go
git commit -m "feat(spellbook): add casting confirmation modal rendering

- Create renderCastConfirmationModal function
- Display full spell info: name, level, stats, description
- Show damage and saving throw with calculated DC
- Include upcast information if applicable
- Show slot selection for leveled spells
- Support cantrips, single-slot, and multi-slot scenarios"
```

---

## Task 5: Modify handleCastSpell to Enter Modal Mode

**Files:**
- Modify: `internal/ui/views/spellbook.go` (handleCastSpell function around line 691)

**Step 1: Update handleCastSpell to always show modal**

Replace the existing `handleCastSpell` function:

```go
func (m *SpellbookModel) handleCastSpell() *SpellbookModel {
	if m.character == nil || m.character.Spellcasting == nil {
		m.statusMessage = "No spellcasting ability"
		return m
	}

	sc := m.character.Spellcasting

	// Check if a spell is selected
	displaySpells := m.getDisplaySpells()
	if m.spellCursor >= len(displaySpells) {
		return m
	}

	spell := displaySpells[m.spellCursor]

	// Check if spell is prepared (if applicable)
	if sc.PreparesSpells && !spell.Prepared && !spell.Ritual {
		m.statusMessage = fmt.Sprintf("%s is not prepared", spell.Name)
		return m
	}

	// Store casting spell and get available levels
	m.castingSpell = &spell

	// For cantrips, no slots needed
	if spell.Level == 0 {
		m.availableCastLevels = []int{}
	} else {
		// Find available spell slot levels
		m.availableCastLevels = m.getAvailableCastLevels(spell.Level)

		if len(m.availableCastLevels) == 0 {
			m.statusMessage = fmt.Sprintf("No spell slots available for %s", spell.Name)
			m.castingSpell = nil
			return m
		}
	}

	// Initialize cursor
	m.castLevelCursor = 0

	// Always enter confirmation modal mode
	m.mode = ModeConfirmCast
	return m
}
```

**Step 2: Commit**

```bash
git add internal/ui/views/spellbook.go
git commit -m "feat(spellbook): always show confirmation modal when casting

- Modify handleCastSpell to enter ModeConfirmCast
- Check for prepared status and available slots
- Initialize casting state (spell, levels, cursor)
- Remove direct casting - all spells go through modal"
```

---

## Task 6: Add Modal Key Handling in Update Function

**Files:**
- Modify: `internal/ui/views/spellbook.go` (Update function around line 150)

**Step 1: Add ModeConfirmCast case to Update function**

Find the switch statement in the Update function that handles different modes. Add a new case for `ModeConfirmCast`:

```go
case tea.KeyMsg:
	switch m.mode {
	case ModeSpellList:
		// ... existing code ...

	case ModePreparation:
		// ... existing code ...

	case ModeAddSpell:
		// ... existing code ...

	case ModeSelectCastLevel:
		// ... existing code (this can be removed later or kept for now) ...

	case ModeConfirmCast:
		switch {
		case key.Matches(msg, m.keys.Back): // Esc
			// Cancel casting
			m.mode = ModeSpellList
			m.castingSpell = nil
			m.availableCastLevels = nil
			m.statusMessage = "Casting cancelled"
			return m, nil

		case key.Matches(msg, m.keys.Up):
			// Navigate up in slot selection
			if len(m.availableCastLevels) > 1 && m.castLevelCursor > 0 {
				m.castLevelCursor--
			}
			return m, nil

		case key.Matches(msg, m.keys.Down):
			// Navigate down in slot selection
			if len(m.availableCastLevels) > 1 && m.castLevelCursor < len(m.availableCastLevels)-1 {
				m.castLevelCursor++
			}
			return m, nil

		case key.Matches(msg, m.keys.Enter):
			// Confirm cast
			if m.castingSpell.Level == 0 {
				// Cantrip - no slot needed
				m.statusMessage = fmt.Sprintf("Cast %s (no slot required)", m.castingSpell.Name)
				m.mode = ModeSpellList
				m.castingSpell = nil
				return m, m.saveCharacter()
			} else if len(m.availableCastLevels) > 0 {
				// Cast with selected slot level
				selectedLevel := m.availableCastLevels[m.castLevelCursor]
				m.mode = ModeSpellList
				return m.castSpellAtLevel(m.castingSpell, selectedLevel), m.saveCharacter()
			} else {
				// No slots available (shouldn't reach here)
				m.statusMessage = "No spell slots available"
				m.mode = ModeSpellList
				m.castingSpell = nil
				return m, nil
			}
		}
	}
```

**Step 2: Commit**

```bash
git add internal/ui/views/spellbook.go
git commit -m "feat(spellbook): add key handling for confirmation modal

- Handle Esc to cancel casting
- Handle Up/Down to navigate slot levels
- Handle Enter to confirm and cast spell
- Support cantrips and leveled spells
- Return to spell list after action"
```

---

## Task 7: Add Modal Overlay Rendering in View Function

**Files:**
- Modify: `internal/ui/views/spellbook.go` (View function around line 350)

**Step 1: Add modal overlay rendering**

Find the View function's return statement. Before returning, check if we're in ModeConfirmCast and overlay the modal:

```go
func (m *SpellbookModel) View() string {
	// ... existing code to build the main view ...

	// Combine panels
	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		spellListPanel,
		spellDetailsPanel,
		spellSlotsPanel,
	)

	// Add status bar
	view := lipgloss.JoinVertical(lipgloss.Left, header, mainView, statusBar)

	// Overlay confirmation modal if in ModeConfirmCast
	if m.mode == ModeConfirmCast {
		modal := m.renderCastConfirmationModal()
		view = m.overlayModal(view, modal)
	}

	return view
}
```

**Step 2: Add overlayModal helper function**

Add this helper function to position the modal over the main view:

```go
// overlayModal overlays a modal on top of the base view.
func (m *SpellbookModel) overlayModal(baseView, modal string) string {
	// Simple center overlay - place modal over the base view
	// For now, just append the modal (Bubble Tea will handle layering)
	// In a more sophisticated version, you'd calculate positioning

	// Split base into lines
	baseLines := strings.Split(baseView, "\n")
	modalLines := strings.Split(modal, "\n")

	// Calculate vertical centering
	baseHeight := len(baseLines)
	modalHeight := len(modalLines)
	topPadding := (baseHeight - modalHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	// For horizontal centering, calculate based on width
	modalWidth := 0
	if len(modalLines) > 0 {
		modalWidth = len([]rune(modalLines[0]))
	}
	baseWidth := m.width
	leftPadding := (baseWidth - modalWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}

	// Overlay the modal
	result := make([]string, baseHeight)
	copy(result, baseLines)

	for i, line := range modalLines {
		targetLine := topPadding + i
		if targetLine >= 0 && targetLine < baseHeight {
			// Replace the line with padded modal line
			padding := strings.Repeat(" ", leftPadding)
			result[targetLine] = padding + line
		}
	}

	return strings.Join(result, "\n")
}
```

**Step 3: Test the modal display**

```bash
go run ./cmd/sheet
# Load character with spells
# Press 's' for spellbook
# Select a spell and press 'c'
# Should see confirmation modal with spell details
# Test navigation with arrow keys (if multiple slots)
# Test Enter to cast
# Test Esc to cancel
```

**Step 4: Commit**

```bash
git add internal/ui/views/spellbook.go
git commit -m "feat(spellbook): overlay confirmation modal in view

- Add modal overlay rendering in View function
- Create overlayModal helper for centering
- Display modal when in ModeConfirmCast mode
- Position modal in center of screen"
```

---

## Task 8: Update Tests for New Modal Behavior

**Files:**
- Modify: `internal/ui/views/spellbook_test.go`

**Step 1: Update TestSpellbookModel_CastPreparedSpell test**

Find the test that checks casting behavior and update it to expect modal mode:

```go
func TestSpellbookModel_CastPreparedSpell(t *testing.T) {
	// ... existing setup ...

	// Try to cast the spell
	model = model.handleCastSpell()

	// Should enter confirmation modal mode
	assert.Equal(t, ModeConfirmCast, model.mode, "Should enter confirmation modal mode")
	assert.NotNil(t, model.castingSpell, "Should have casting spell set")
	assert.Equal(t, "Magic Missile", model.castingSpell.Name, "Should be casting Magic Missile")

	// Simulate Enter key to confirm cast
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should return to spell list mode
	assert.Equal(t, ModeSpellList, model.mode, "Should return to spell list mode")
	assert.Contains(t, model.statusMessage, "Cast Magic Missile", "Should show cast message")

	// Spell slot should be consumed
	slot := sc.SpellSlots.GetSlot(1)
	assert.Equal(t, 3, slot.Remaining, "Should have consumed one level 1 slot")
}
```

**Step 2: Add test for modal cancellation**

```go
func TestSpellbookModel_CancelCasting(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Wizard", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:        models.AbilityIntelligence,
		SpellSlots:     models.NewSpellSlots(),
		KnownSpells:    []models.KnownSpell{},
		PreparesSpells: true,
		MaxPrepared:    5,
	}
	sc.SpellSlots.SetSlots(1, 4)
	sc.AddSpell("Magic Missile", 1)
	sc.PrepareSpell("Magic Missile", true)
	char.Spellcasting = sc

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)
	model.mode = ModeSpellList

	// Start casting
	model = model.handleCastSpell()
	assert.Equal(t, ModeConfirmCast, model.mode)

	// Press Esc to cancel
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Should return to spell list
	assert.Equal(t, ModeSpellList, model.mode)
	assert.Contains(t, model.statusMessage, "cancelled", "Should show cancellation message")
	assert.Nil(t, model.castingSpell, "Should clear casting spell")

	// Spell slot should NOT be consumed
	slot := sc.SpellSlots.GetSlot(1)
	assert.Equal(t, 4, slot.Remaining, "Should not consume slot when cancelled")
}
```

**Step 3: Add test for cantrip casting**

```go
func TestSpellbookModel_CastCantrip(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Wizard", "Human", "Wizard")
	sc := &models.Spellcasting{
		Ability:        models.AbilityIntelligence,
		SpellSlots:     models.NewSpellSlots(),
		KnownSpells:    []models.KnownSpell{},
		CantripsKnown:  []string{"Fire Bolt"},
		PreparesSpells: true,
		MaxPrepared:    5,
	}
	char.Spellcasting = sc

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewSpellbookModel(char, store, loader)
	model.mode = ModeSpellList

	// Start casting cantrip
	model = model.handleCastSpell()
	assert.Equal(t, ModeConfirmCast, model.mode)
	assert.Equal(t, 0, len(model.availableCastLevels), "Cantrips should have no slot levels")

	// Confirm cast
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should cast without consuming resources
	assert.Equal(t, ModeSpellList, model.mode)
	assert.Contains(t, model.statusMessage, "no slot required")
}
```

**Step 4: Run all tests**

```bash
go test ./internal/ui/views -v -run TestSpellbook
```

Expected: All tests should pass

**Step 5: Commit**

```bash
git add internal/ui/views/spellbook_test.go
git commit -m "test(spellbook): update tests for confirmation modal

- Update CastPreparedSpell test to expect modal mode
- Add test for casting cancellation with Esc
- Add test for cantrip casting through modal
- Verify slot consumption and status messages"
```

---

## Task 9: Manual Testing and Refinement

**Files:**
- Test: Manual testing with application

**Step 1: Test with Elara (wizard)**

```bash
go run ./cmd/sheet
# Select Elara the Wizard
# Press 's' for spellbook
```

Test scenarios:
1. Cast a cantrip (Fire Bolt) - verify modal shows, no slots
2. Cast a level 1 spell with multiple slots available - verify slot selection
3. Cast a spell with only one slot available - verify single slot display
4. Cancel casting with Esc - verify return to spell list
5. Navigate with arrow keys between slot levels - verify highlighting
6. Verify damage and save DC display correctly
7. Verify upcast effects show for each level

**Step 2: Test with Raven (warlock with pact magic)**

```bash
# Select Raven the Warlock
# Press 's' for spellbook
```

Test scenarios:
1. Cast a spell using pact magic - verify correct slot display
2. Verify pact magic slots shown with level and remaining

**Step 3: Test edge cases**

- Spell without damage (utility spell like Invisibility)
- Spell without save (attack spell like Magic Missile)
- Spell with both damage and save (like Burning Hands)
- Unprepared spell - verify error message
- No slots remaining - verify error message

**Step 4: Document any issues**

Create notes of any visual issues, bugs, or UX improvements needed.

---

## Task 10: Final Cleanup and Documentation

**Files:**
- Modify: `internal/ui/views/spellbook.go` (remove old code if any)
- Create: `docs/features/spellbook-casting.md` (user-facing documentation)

**Step 1: Remove obsolete ModeSelectCastLevel code (optional)**

The old `ModeSelectCastLevel` mode can be removed if it's no longer used, or kept for backward compatibility. Review the code and decide.

**Step 2: Add code comments**

Ensure all new functions have clear documentation comments explaining their purpose.

**Step 3: Create user documentation**

Create `docs/features/spellbook-casting.md`:

```markdown
# Spellbook Casting

## Overview

The spellbook provides a comprehensive spell management and casting interface for spellcasting characters.

## Casting Spells

1. Open spellbook with 's' key
2. Select a spell from the list (arrow keys or j/k)
3. Press 'c' to cast
4. Review the casting confirmation modal showing:
   - Spell name, level, and school
   - Casting time, range, components, duration
   - Damage (if applicable)
   - Saving throw and DC (if applicable)
   - Full description
   - Upcast effects (for leveled spells)
5. For leveled spells, select slot level with arrow keys
6. Press Enter to cast or Esc to cancel

## Spell Information

The spell details panel (right side) displays:
- Basic spell statistics
- **Damage**: Shows dice and damage type (e.g., "3d6 fire")
- **Saving Throw**: Shows ability and calculated DC (e.g., "Dexterity DC 15")
- Full description
- Upcast information (how the spell improves at higher levels)

## Spell Save DC

The spell save DC is automatically calculated based on:
- Spellcasting ability modifier
- Proficiency bonus
- Formula: 8 + ability modifier + proficiency bonus

## Cantrips

Cantrips can be cast without consuming spell slots. The confirmation modal will show all spell details but no slot selection.

## Pact Magic

Warlocks and other pact magic users will see their pact slots displayed separately from regular spell slots. Pact magic slots are shown with their level (e.g., "Pact Magic - Level 3").
```

**Step 4: Final commit**

```bash
git add docs/features/spellbook-casting.md internal/ui/views/spellbook.go
git commit -m "docs: add spellbook casting documentation and cleanup

- Add user-facing documentation for casting feature
- Clean up code comments
- Document spell information display and save DC calculation"
```

---

## Task 11: Build and Final Verification

**Files:**
- Build: Application binary

**Step 1: Clean build**

```bash
go build -o sheet ./cmd/sheet
```

Expected: Build should succeed with no errors

**Step 2: Run final verification**

Test the complete flow one more time:
1. Load character
2. Open spellbook
3. Cast various spell types
4. Verify all information displays correctly
5. Verify modal interactions work smoothly

**Step 3: Run all tests**

```bash
go test ./... -v
```

Expected: All tests pass

**Step 4: Final commit if needed**

```bash
# If any last-minute fixes were made:
git add .
git commit -m "fix: final adjustments for casting confirmation modal"
```

---

## Success Criteria Checklist

- [ ] Pressing 'c' always shows confirmation modal (even for cantrips)
- [ ] Modal displays complete spell information (name, stats, description)
- [ ] Damage shown with type (e.g., "3d6 fire")
- [ ] Saving throw shown with ability and calculated DC
- [ ] Spell save DC correctly calculated (8 + mod + prof)
- [ ] Header stats are bold and easy to read
- [ ] Cantrips work (no slot selection)
- [ ] Single-slot spells work (one option shown)
- [ ] Multi-slot spells work (arrow navigation)
- [ ] Upcast effects shown for each slot level
- [ ] Esc cancels casting
- [ ] Enter confirms and casts spell
- [ ] Slots consumed correctly after casting
- [ ] Status messages show casting confirmation
- [ ] All tests pass
- [ ] No regressions in existing functionality

---

## Notes

- **TDD:** Tests are updated after implementation to verify modal behavior
- **DRY:** Reuses existing functions (getAvailableCastLevels, calculateUpcastEffect, wordWrap)
- **YAGNI:** No spell categorization, just display what's available
- **Frequent commits:** Each task has its own commit
- **No data changes:** Works with existing SpellData structure

## Future Enhancements (Not in This Plan)

- Scroll support for very long spell descriptions
- Concentration indicator
- Spell range/area visualization
- Component explanations (hover/tooltip)
- Keyboard shortcut customization
