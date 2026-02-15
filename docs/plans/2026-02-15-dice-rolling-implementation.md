# Phase 12: Dice Rolling Integration — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Integrate the `github.com/Domo929/roll` library to provide animated dice rolling with a tumbling dice + color flash animation, advantage/disadvantage prompts, roll history column, custom dice roller, and Luck rolls across all active play views.

**Architecture:** A centralized `RollEngine` component in `internal/ui/components/` handles all dice rolling state: advantage prompts, tick-based animation, modal overlay rendering, and follow-up rolls. A `RollHistory` component tracks recent rolls and renders a toggleable right-side column. Views send `RequestRollMsg` messages and never call the roll library directly. The app model (`internal/ui/model.go`) owns both components and routes messages between views and the engine.

**Tech Stack:** Go, Bubble Tea (TUI framework), lipgloss (styling), bubbles/key (key bindings), `github.com/Domo929/roll` (dice library)

**Existing Patterns to Follow:**
- View navigation: see `internal/ui/model.go` (message routing in Update)
- Key maps: see `internal/ui/views/main_sheet.go` (mainSheetKeyMap struct + defaultMainSheetKeyMap())
- Modal overlay: see `main_sheet.go` renderCastConfirmationModal (lipgloss centered box)
- Component convention: see `internal/ui/components/input.go` (simple struct + Render() method)
- Tick-based animation: use `tea.Tick(duration, func(t time.Time) tea.Msg { return tickMsg(t) })`

**Git Workflow:** Work on branch `feature/dice-rolling`. Commits should be logically grouped. PR against `main` when complete.

---

### Task 1: Add Roll Library Dependency

**Files:**
- Modify: `go.mod`

**Step 1: Add the roll library dependency**

Since the roll library is local at `/home/dcupo/Software/roll`, add it as a local module replacement:

```bash
cd /home/dcupo/Software/sheet/.worktrees/dice-rolling  # or wherever the worktree is
go get github.com/Domo929/roll
```

If the module isn't published to a registry, use a replace directive:

```bash
go mod edit -replace github.com/Domo929/roll=/home/dcupo/Software/roll
go mod tidy
```

**Step 2: Verify the dependency works**

Create a quick test in a temporary file to verify the import works:

```bash
go build ./...
```

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add github.com/Domo929/roll dependency for dice rolling

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 2: Roll History Data Model

**Files:**
- Create: `internal/ui/components/roll_history.go`
- Create: `internal/ui/components/roll_history_test.go`

**Step 1: Create roll_history.go with types and data structure**

```go
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
	Entries    []RollHistoryEntry
	Visible    bool // whether the history column is shown
	ScrollPos  int  // scroll position for viewing history
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
```

**Step 2: Write tests in roll_history_test.go**

```go
package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRollHistory_Add(t *testing.T) {
	h := NewRollHistory()
	assert.Empty(t, h.Entries)
	assert.False(t, h.Visible)

	h.Add(RollHistoryEntry{Label: "Test Roll", Total: 15})
	assert.Len(t, h.Entries, 1)
	assert.True(t, h.Visible, "First add should make history visible")
	assert.Equal(t, "Test Roll", h.Entries[0].Label)
}

func TestRollHistory_AddNewestFirst(t *testing.T) {
	h := NewRollHistory()
	h.Add(RollHistoryEntry{Label: "First"})
	h.Add(RollHistoryEntry{Label: "Second"})
	assert.Equal(t, "Second", h.Entries[0].Label)
	assert.Equal(t, "First", h.Entries[1].Label)
}

func TestRollHistory_CapsAt50(t *testing.T) {
	h := NewRollHistory()
	for i := range 60 {
		h.Add(RollHistoryEntry{Label: fmt.Sprintf("Roll %d", i)})
	}
	assert.Len(t, h.Entries, 50)
	assert.Equal(t, "Roll 59", h.Entries[0].Label)
}

func TestRollHistory_Toggle(t *testing.T) {
	h := NewRollHistory()
	h.Add(RollHistoryEntry{Label: "Test"})
	assert.True(t, h.Visible)
	h.Toggle()
	assert.False(t, h.Visible)
	h.Toggle()
	assert.True(t, h.Visible)
}

func TestRollHistory_Clear(t *testing.T) {
	h := NewRollHistory()
	h.Add(RollHistoryEntry{Label: "Test"})
	h.Clear()
	assert.Empty(t, h.Entries)
	assert.False(t, h.Visible)
}

func TestRollTypeIcon(t *testing.T) {
	assert.Equal(t, "⚔", RollTypeIcon(RollAttack))
	assert.Equal(t, "💥", RollTypeIcon(RollDamage))
	assert.Equal(t, "🎯", RollTypeIcon(RollSkillCheck))
	assert.Equal(t, "🛡", RollTypeIcon(RollSavingThrow))
	assert.Equal(t, "🎲", RollTypeIcon(RollLuck))
}

func TestRollHistory_RenderEmpty(t *testing.T) {
	h := NewRollHistory()
	result := h.Render(25, 40)
	assert.Empty(t, result, "Hidden history should render empty")
}

func TestRollHistory_RenderVisible(t *testing.T) {
	h := NewRollHistory()
	h.Add(RollHistoryEntry{Label: "Longsword Attack", Expression: "1d20+7", Total: 19, RollType: RollAttack})
	result := h.Render(25, 40)
	assert.Contains(t, result, "Roll History")
	assert.Contains(t, result, "Longsword Attack")
}

func TestRollHistory_RenderTooNarrow(t *testing.T) {
	h := NewRollHistory()
	h.Add(RollHistoryEntry{Label: "Test"})
	result := h.Render(15, 40)
	assert.Empty(t, result, "Should not render if width < 20")
}
```

**Step 3: Run tests**

```bash
go test ./internal/ui/components/ -run TestRollHistory -v
```

Expected: All tests pass.

**Step 4: Commit**

```bash
git add internal/ui/components/roll_history.go internal/ui/components/roll_history_test.go
git commit -m "feat: add RollHistory data model and column rendering

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 3: Roll Engine — Core State Machine and Messages

**Files:**
- Create: `internal/ui/components/roll_engine.go`
- Create: `internal/ui/components/roll_engine_test.go`

This task implements the roll engine state machine, message types, the advantage/disadvantage prompt, and the animated modal overlay with tumbling dice and color flashing. It also includes the custom dice roller overlay.

**Step 1: Create roll_engine.go**

The roll engine has these states: `Idle`, `AdvPrompt`, `CustomRoll`, `Animating`, `Showing`.

Key message types:
- `RequestRollMsg` — sent by views to request a roll
- `RollCompleteMsg` — sent by the engine when the user dismisses the result
- `rollTickMsg` — internal animation tick

The engine:
- Receives `RequestRollMsg` with `Label`, `DiceExpr` (string like "1d20"), `Modifier` (int), `RollType`, `AdvPrompt` (bool), and optional `FollowUp *RequestRollMsg`
- If `AdvPrompt` is true, enters `AdvPrompt` state showing N/A/D picker. N/A/D keys immediately trigger the roll.
- Calls `roll.RollString()` or `roll.RollAdvantage()`/`roll.RollDisadvantage()` from `github.com/Domo929/roll`
- Enters `Animating` state: ~12 frames, each showing random values for all dice, with colors cycling through a palette (magenta → cyan → yellow → white). Dice "land" from left to right (landed dice turn green). Frame delay starts at 80ms and increases by 15ms per frame.
- Enters `Showing` state: final result displayed. If `FollowUp` is set, shows "Enter: roll damage • Esc: skip". Otherwise "Press any key to continue".
- On dismissal, sends `RollCompleteMsg` back to the view.

The custom dice roller (`CustomRoll` state) shows a die-type picker (d4/d6/d8/d10/d12/d20/d100) with Left/Right to select, Up/Down to change quantity (1–100), Enter to roll, Esc to cancel.

Implement the full `RollEngine` struct with `Update(msg tea.Msg) (*RollEngine, tea.Cmd)` and `View(underlayWidth, underlayHeight int) string` methods. The `View()` method renders the appropriate modal overlay (adv prompt, animation, result, or custom roller) as a centered lipgloss box.

Important implementation details:
- Store `displayVals []int` for the tumbling animation (random values shown per frame)
- Store `displayColors []lipgloss.Color` cycling through `{"99", "14", "11", "15"}` (magenta, cyan, yellow, white)
- When a die "lands" (frame > maxFrames - (numDice - dieIndex) - 1), set it to the final value and color green ("10")
- Store `finalResult *roll.Result` after the roll completes
- The `RequestRollMsg.FollowUp` field allows chaining (attack → damage)
- For advantage: call `roll.RollAdvantage(modifier)`. For disadvantage: call `roll.RollDisadvantage(modifier)`.
- For normal d20: construct `roll.Dice{NumDice: 1, Sides: 20, Modifier: modifier}` and call `d.Roll()`
- For non-d20 rolls (damage, hit dice, custom): call `roll.RollString(expr)` where expr includes the modifier (e.g., "1d8+4")
- Nat crit/fail detection: for d20 rolls (Sides == 20), check if any value in `result.Kept` is 20 or 1

**Step 2: Write tests in roll_engine_test.go**

Test the state machine transitions:
- `TestRollEngine_RequestRollWithAdvPrompt` — verify AdvPrompt state on d20 roll request
- `TestRollEngine_RequestRollWithoutAdvPrompt` — verify direct Animating state for non-d20
- `TestRollEngine_AdvPromptNormal` — press N → Animating
- `TestRollEngine_AdvPromptAdvantage` — press A → Animating with advantage
- `TestRollEngine_AdvPromptDisadvantage` — press D → Animating with disadvantage
- `TestRollEngine_AdvPromptEsc` — press Esc → Idle
- `TestRollEngine_AnimationCompletes` — send enough ticks → Showing state
- `TestRollEngine_ShowingDismiss` — any key in Showing → RollCompleteMsg
- `TestRollEngine_ShowingWithFollowUp` — Enter in Showing with FollowUp → new Animating
- `TestRollEngine_ShowingSkipFollowUp` — Esc in Showing with FollowUp → RollCompleteMsg
- `TestRollEngine_CustomRollNavigation` — Left/Right changes die type, Up/Down changes quantity
- `TestRollEngine_CustomRollEnter` — Enter triggers roll with selected dice
- `TestRollEngine_CustomRollEsc` — Esc returns to Idle
- `TestRollEngine_CustomRollQuantityBounds` — quantity clamped to 1–100
- `TestRollEngine_IsActive` — verify IsActive() returns true when not Idle
- `TestRollEngine_ViewRendersModal` — verify View() produces non-empty string when active

**Step 3: Run tests**

```bash
go test ./internal/ui/components/ -run TestRollEngine -v
```

**Step 4: Commit**

```bash
git add internal/ui/components/roll_engine.go internal/ui/components/roll_engine_test.go
git commit -m "feat: add RollEngine state machine with animation and custom roller

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 4: Wire Roll Engine into App Model

**Files:**
- Modify: `internal/ui/model.go`

**Step 1: Add rollEngine and rollHistory fields to Model struct**

After the existing view model fields (around line 46), add:

```go
// Roll engine and history (shared across views)
rollEngine  *components.RollEngine
rollHistory *components.RollHistory
```

Import `"github.com/Domo929/sheet/internal/ui/components"` if not already imported.

**Step 2: Initialize in NewModel()**

In the `NewModel()` function, add to the returned Model:

```go
rollEngine:  components.NewRollEngine(),
rollHistory: components.NewRollHistory(),
```

**Step 3: Handle roll messages in Update()**

Add cases in the `Update()` method's switch, before the default `updateCurrentView` call:

```go
case components.RequestRollMsg:
	// Forward to roll engine
	m.rollEngine, cmd = m.rollEngine.Update(msg)
	return m, cmd

case components.RollCompleteMsg:
	// Roll finished — views can handle if needed
	return m.updateCurrentView(msg)

case components.RollTickMsg:
	// Animation tick — forward to roll engine
	m.rollEngine, cmd = m.rollEngine.Update(msg)
	return m, cmd
```

**Step 4: Intercept key messages when roll engine is active**

In `Update()`, before routing to views, check if the roll engine is active and intercept keys:

```go
case tea.KeyMsg:
	// If roll engine is active, it handles all keys
	if m.rollEngine != nil && m.rollEngine.IsActive() {
		var engineCmd tea.Cmd
		m.rollEngine, engineCmd = m.rollEngine.Update(msg)
		// Check if roll completed
		// (RollCompleteMsg will come through as a tea.Cmd result)
		return m, engineCmd
	}
	// Otherwise fall through to normal handling...
```

This goes near the top of the `Update()` method, after `tea.WindowSizeMsg` but before other message handling.

**Step 5: Add roll history toggle key (`h`)**

In the key handling for active play views (main sheet, spellbook), handle `h`:

The simplest approach: add to the Update method, after the roll engine intercept but before view routing — when currentView is MainSheet, Spellbook, or Combat and the key is "h":

```go
// Handle roll history toggle on active play views
if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "h" {
	if m.currentView == ViewMainSheet || m.currentView == ViewSpellbook || m.currentView == ViewCombat {
		if m.rollHistory != nil {
			m.rollHistory.Toggle()
			return m, nil
		}
	}
}
```

Note: Be careful not to conflict with the main sheet `h` key which is currently bound to "heal" — this only applies when `focusArea == FocusCombat`. The `h` toggle for history should only trigger when focus is NOT on combat. Alternatively, use a different key. Check that the main sheet's key handling in its own Update takes priority (it does, since the main sheet's Update handles `h` internally for heal). The history toggle in model.go should only fire if the view's Update doesn't consume the key. This needs careful handling.

**Better approach:** Don't handle `h` in model.go. Instead, have each view (main_sheet, spellbook) send a `ToggleRollHistoryMsg{}` message when appropriate, and handle that in model.go. This avoids key conflicts. The views already know their own key context.

Add to model.go Update:
```go
case views.ToggleRollHistoryMsg:
	if m.rollHistory != nil {
		m.rollHistory.Toggle()
	}
	return m, nil
```

**Step 6: Render roll overlay and history in View()**

The roll engine overlay and history column need to be composited on top of / beside the current view. Modify the `View()` method:

For active play views (MainSheet, Spellbook, Combat), the layout becomes:

```go
case ViewMainSheet:
	viewContent := m.renderMainSheet()
	return m.compositeWithRollUI(viewContent)
case ViewSpellbook:
	viewContent := m.renderSpellbook()
	return m.compositeWithRollUI(viewContent)
```

Add a helper:

```go
func (m Model) compositeWithRollUI(viewContent string) string {
	// Add roll history column if visible
	if m.rollHistory != nil && m.rollHistory.Visible && m.width >= 80 {
		historyWidth := 27
		viewWidth := m.width - historyWidth - 1
		// Re-render the view at reduced width? Or just join horizontally
		historyCol := m.rollHistory.Render(historyWidth, m.height)
		viewContent = lipgloss.JoinHorizontal(lipgloss.Top, viewContent, historyCol)
	}

	// Overlay roll engine modal if active
	if m.rollEngine != nil && m.rollEngine.IsActive() {
		overlay := m.rollEngine.View(m.width, m.height)
		// Use lipgloss.Place to center the overlay
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("0")))
	}

	return viewContent
}
```

Note: The exact overlay rendering approach may need adjustment. The roll engine's `View()` returns the modal box; we need to render it centered over the existing view. One approach is to render the underlay first, then replace center lines with the overlay. A simpler approach for Bubble Tea: since we can only return one string, render the overlay as the entire view when the engine is active (the modal is full-screen-sized with the underlay visible around the edges).

**Step 7: Handle rollHistory addition when RollCompleteMsg arrives**

When the roll engine sends a `RollCompleteMsg`, add the entry to history:

```go
case components.RollCompleteMsg:
	if m.rollHistory != nil {
		m.rollHistory.Add(msg.Entry)
	}
	return m.updateCurrentView(msg)
```

**Step 8: Run tests and build**

```bash
go test ./... -v
go build ./cmd/sheet/
```

**Step 9: Commit**

```bash
git add internal/ui/model.go
git commit -m "feat: wire roll engine and history into app model

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 5: Add Skill and Save Cursors to Main Sheet

**Files:**
- Modify: `internal/ui/views/main_sheet.go`

Currently the Skills and Abilities/Saves panels have no per-item cursor — they just render lists. We need to add cursors so users can select a skill or saving throw and press Enter to roll.

**Step 1: Add cursor fields to MainSheetModel**

Add to the struct (around line 53):

```go
// Skill and save cursors for rolling
skillCursor int // 0 = Luck, 1-18 = the 18 skills (Acrobatics..Survival)
saveCursor  int // 0-5 = STR..CHA saving throws
```

**Step 2: Add key bindings**

Add to `mainSheetKeyMap` struct:

```go
Luck        key.Binding // backtick
CustomRoll  key.Binding // /
HistoryToggle key.Binding // H (capital)
```

Add to `defaultMainSheetKeyMap()`:

```go
Luck: key.NewBinding(
	key.WithKeys("`"),
	key.WithHelp("`", "luck"),
),
CustomRoll: key.NewBinding(
	key.WithKeys("/"),
	key.WithHelp("/", "custom roll"),
),
HistoryToggle: key.NewBinding(
	key.WithKeys("H"),
	key.WithHelp("H", "roll history"),
),
```

Note: Use capital `H` (Shift+H) for history toggle since lowercase `h` is already used for heal when combat is focused. The backtick `` ` `` is for luck.

**Step 3: Handle Up/Down/Enter for Skills panel**

In the `Update()` method, add handling when `m.focusArea == FocusSkills` (before the existing general key handling, similar to the FocusActions block around line 365):

```go
if m.focusArea == FocusSkills {
	switch msg.Type {
	case tea.KeyUp:
		if m.skillCursor > 0 {
			m.skillCursor--
		}
		return m, nil
	case tea.KeyDown:
		// 0 = Luck, 1-18 = 18 skills
		if m.skillCursor < 18 {
			m.skillCursor++
		}
		return m, nil
	case tea.KeyEnter:
		return m.handleSkillRoll()
	}
}
```

**Step 4: Handle Up/Down/Enter for Abilities/Saves panel**

Similarly for `FocusAbilitiesAndSaves`:

```go
if m.focusArea == FocusAbilitiesAndSaves {
	switch msg.Type {
	case tea.KeyUp:
		if m.saveCursor > 0 {
			m.saveCursor--
		}
		return m, nil
	case tea.KeyDown:
		if m.saveCursor < 5 {
			m.saveCursor++
		}
		return m, nil
	case tea.KeyEnter:
		return m.handleSaveRoll()
	}
}
```

**Step 5: Add handleSkillRoll() and handleSaveRoll() stubs**

These will send `RequestRollMsg` — implement as stubs for now that set a status message. The actual roll integration is Task 7.

```go
func (m *MainSheetModel) handleSkillRoll() (*MainSheetModel, tea.Cmd) {
	if m.skillCursor == 0 {
		// Luck roll
		m.statusMessage = "Luck roll (dice integration coming...)"
		return m, nil
	}
	skills := models.AllSkills()
	skillName := skills[m.skillCursor-1]
	m.statusMessage = fmt.Sprintf("Roll %s check (dice integration coming...)", skillName)
	return m, nil
}

func (m *MainSheetModel) handleSaveRoll() (*MainSheetModel, tea.Cmd) {
	abilities := []models.Ability{
		models.AbilityStrength, models.AbilityDexterity, models.AbilityConstitution,
		models.AbilityIntelligence, models.AbilityWisdom, models.AbilityCharisma,
	}
	ability := abilities[m.saveCursor]
	m.statusMessage = fmt.Sprintf("Roll %s saving throw (dice integration coming...)", ability)
	return m, nil
}
```

**Step 6: Handle Luck, Custom Roll, and History Toggle keys**

In the general key handling section (around line 387):

```go
case key.Matches(msg, m.keys.Luck):
	return m.handleSkillRoll() // skillCursor=0 is Luck, but we need to force it
	// Better: send a Luck roll directly
case key.Matches(msg, m.keys.CustomRoll):
	return m, func() tea.Msg { return components.OpenCustomRollMsg{} }
case key.Matches(msg, m.keys.HistoryToggle):
	return m, func() tea.Msg { return ToggleRollHistoryMsg{} }
```

Add `ToggleRollHistoryMsg` type to the views messages:

```go
type ToggleRollHistoryMsg struct{}
```

**Step 7: Update renderSkills() to show cursor and Luck**

Modify `renderSkills()` (starting around line 1305) to:
1. Add a "Luck" entry at the top, rendered in purple/magenta with 🎲 icon, visually separated
2. Show a cursor indicator (`>`) next to the selected skill when `FocusSkills` is active

**Step 8: Update renderAbilitiesAndSaves() to show save cursor**

Modify `renderAbilitiesAndSaves()` (starting around line 1131) to show a cursor indicator next to the selected saving throw when `FocusAbilitiesAndSaves` is active.

**Step 9: Update help footer**

Update the help text (line 2412) to include the new key bindings:

```go
help := "tab: panels • i: inventory • s: spellbook • c: char info • n: notes • /: roll dice • `: luck • H: history • r: rest • esc: back • q: quit"
```

**Step 10: Run tests and build**

```bash
go test ./... -v
go build ./cmd/sheet/
```

**Step 11: Commit**

```bash
git add internal/ui/views/main_sheet.go
git commit -m "feat: add skill/save cursors, luck, and custom roll key bindings

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 6: Main Sheet Layout — History Column Integration

**Files:**
- Modify: `internal/ui/views/main_sheet.go`
- Modify: `internal/ui/model.go`

**Step 1: Pass roll history visibility to main sheet**

The main sheet's `View()` method (line 963) needs to know whether to shrink its layout to accommodate the history column. Add a field to `MainSheetModel`:

```go
rollHistoryVisible bool
rollHistoryWidth   int // width reserved for history column (0 if hidden)
```

When the model.go routes WindowSizeMsg or ToggleRollHistoryMsg, update these fields on the main sheet model.

**Step 2: Adjust main sheet View() layout**

In the `View()` method (around line 989), adjust the column widths when history is visible:

```go
historyReserved := 0
if m.rollHistoryVisible {
	historyReserved = 27
}

leftWidth := 38
rightWidth := width - leftWidth - 4 - historyReserved
```

This shrinks the right column (actions/combat) to make room for the history column that model.go will append.

**Step 3: Update model.go compositeWithRollUI**

In model.go, the `compositeWithRollUI` helper renders the history column beside the view and overlays the roll engine modal when active. This was outlined in Task 4 — ensure it works correctly by passing the adjusted width to the main sheet before rendering.

**Step 4: Run tests and build**

```bash
go test ./... -v
go build ./cmd/sheet/
```

**Step 5: Commit**

```bash
git add internal/ui/views/main_sheet.go internal/ui/model.go
git commit -m "feat: integrate roll history column into main sheet layout

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 7: Skill Checks, Saving Throws, and Luck Rolls

**Files:**
- Modify: `internal/ui/views/main_sheet.go`

Replace the stub `handleSkillRoll()` and `handleSaveRoll()` with real `RequestRollMsg` sends.

**Step 1: Implement handleSkillRoll()**

```go
func (m *MainSheetModel) handleSkillRoll() (*MainSheetModel, tea.Cmd) {
	if m.skillCursor == 0 {
		// Luck: d20, no modifier, no advantage prompt
		return m, func() tea.Msg {
			return components.RequestRollMsg{
				Label:     "Luck",
				DiceExpr:  "1d20",
				Modifier:  0,
				RollType:  components.RollLuck,
				AdvPrompt: false,
			}
		}
	}

	skills := models.AllSkills()
	skillName := skills[m.skillCursor-1]
	mod := m.character.GetSkillModifier(skillName)

	return m, func() tea.Msg {
		return components.RequestRollMsg{
			Label:     string(skillName) + " Check",
			DiceExpr:  "1d20",
			Modifier:  mod,
			RollType:  components.RollSkillCheck,
			AdvPrompt: true,
		}
	}
}
```

**Step 2: Implement handleSaveRoll()**

```go
func (m *MainSheetModel) handleSaveRoll() (*MainSheetModel, tea.Cmd) {
	abilities := []models.Ability{
		models.AbilityStrength, models.AbilityDexterity, models.AbilityConstitution,
		models.AbilityIntelligence, models.AbilityWisdom, models.AbilityCharisma,
	}
	abilityNames := []string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}

	ability := abilities[m.saveCursor]
	mod := m.character.GetSavingThrowModifier(ability)

	return m, func() tea.Msg {
		return components.RequestRollMsg{
			Label:     abilityNames[m.saveCursor] + " Saving Throw",
			DiceExpr:  "1d20",
			Modifier:  mod,
			RollType:  components.RollSavingThrow,
			AdvPrompt: true,
		}
	}
}
```

**Step 3: Update Luck backtick key handler**

```go
case key.Matches(msg, m.keys.Luck):
	// Force luck roll regardless of current skill cursor
	return m, func() tea.Msg {
		return components.RequestRollMsg{
			Label:     "Luck",
			DiceExpr:  "1d20",
			Modifier:  0,
			RollType:  components.RollLuck,
			AdvPrompt: false,
		}
	}
```

**Step 4: Run tests and build**

```bash
go test ./... -v
go build ./cmd/sheet/
```

**Step 5: Commit**

```bash
git add internal/ui/views/main_sheet.go
git commit -m "feat: integrate dice rolling for skill checks, saving throws, and luck

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 8: Weapon Attack Rolls

**Files:**
- Modify: `internal/ui/views/main_sheet.go`

**Step 1: Update handleActionSelection() weapon case**

Replace the weapon case (lines 586–605) to send `RequestRollMsg` instead of setting a status message. The attack roll should have a follow-up damage roll:

```go
case ActionItemWeapon:
	if selectedItem.Weapon != nil {
		w := selectedItem.Weapon
		attackBonus := m.getWeaponAttackBonus(*w)
		damageMod := m.getWeaponDamageMod(*w)

		damageExpr := w.Damage
		if damageExpr == "" {
			damageExpr = "1d4" // unarmed strike fallback
		}

		return m, func() tea.Msg {
			return components.RequestRollMsg{
				Label:     w.Name + " Attack",
				DiceExpr:  "1d20",
				Modifier:  attackBonus,
				RollType:  components.RollAttack,
				AdvPrompt: true,
				FollowUp: &components.RequestRollMsg{
					Label:     w.Name + " Damage (" + string(w.DamageType) + ")",
					DiceExpr:  damageExpr,
					Modifier:  damageMod,
					RollType:  components.RollDamage,
					AdvPrompt: false,
				},
			}
		}
	}
	// Unarmed strike
	strMod := m.character.AbilityScores.Strength.Modifier()
	return m, func() tea.Msg {
		return components.RequestRollMsg{
			Label:     "Unarmed Strike Attack",
			DiceExpr:  "1d20",
			Modifier:  strMod + m.character.GetProficiencyBonus(),
			RollType:  components.RollAttack,
			AdvPrompt: true,
			FollowUp: &components.RequestRollMsg{
				Label:     "Unarmed Strike Damage (bludgeoning)",
				DiceExpr:  "1",
				Modifier:  strMod,
				RollType:  components.RollDamage,
				AdvPrompt: false,
			},
		}
	}
```

Note: For unarmed strike damage, the roll library may not parse "1" as a valid dice expression. You may need to handle this as a special case (1 + STR mod, no roll needed) or use "1d1" as a workaround. Check the roll library's parser — it requires the `d` notation. Handle unarmed as a flat value result without animation, or use "1d1+mod" as a workaround.

**Step 2: Run tests and build**

```bash
go test ./... -v
go build ./cmd/sheet/
```

**Step 3: Commit**

```bash
git add internal/ui/views/main_sheet.go
git commit -m "feat: integrate dice rolling for weapon attacks with follow-up damage

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 9: Spell Casting Rolls

**Files:**
- Modify: `internal/ui/views/main_sheet.go`
- Modify: `internal/ui/views/spellbook.go`

**Step 1: Update main sheet castSpellAtLevel()**

After a spell slot is consumed (around the status message lines), check if the spell has damage and/or requires a spell attack roll:

```go
// After consuming slot successfully:
if spell.Damage != "" {
	// Does this spell require a spell attack?
	isSpellAttack := false // Determine from spell data — check if it has an attack roll
	// Heuristic: if spell has no SavingThrow, it might be a spell attack
	// For simplicity: if SavingThrow is empty and Damage is present, assume spell attack

	if spell.SavingThrow == "" && spell.Damage != "" {
		// Spell attack roll + damage follow-up
		attackBonus := m.character.GetSpellAttackBonus()
		damageExpr := spell.Damage // may need upcast adjustment
		return m, func() tea.Msg {
			return components.RequestRollMsg{
				Label:     spell.Name + " Attack",
				DiceExpr:  "1d20",
				Modifier:  attackBonus,
				RollType:  components.RollAttack,
				AdvPrompt: true,
				FollowUp: &components.RequestRollMsg{
					Label:    spell.Name + " Damage (" + spell.DamageType + ")",
					DiceExpr: damageExpr,
					Modifier: 0,
					RollType: components.RollDamage,
				},
			}
		}
	} else if spell.SavingThrow != "" && spell.Damage != "" {
		// Save-based spell — just roll damage, show DC in label
		saveDC := m.getSpellSaveDC()
		damageExpr := spell.Damage
		return m, func() tea.Msg {
			return components.RequestRollMsg{
				Label:    fmt.Sprintf("%s Damage (DC %d %s)", spell.Name, saveDC, spell.SavingThrow),
				DiceExpr: damageExpr,
				Modifier: 0,
				RollType: components.RollDamage,
			}
		}
	}
}
// Spells with no damage: keep existing status message behavior
```

**Step 2: Apply the same pattern in spellbook.go castSpellAtLevel()**

The spellbook view has its own `castSpellAtLevel()` function (lines 1358–1396). Apply the same logic there, checking the selected spell's damage and saving throw fields.

**Step 3: Run tests and build**

```bash
go test ./... -v
go build ./cmd/sheet/
```

**Step 4: Commit**

```bash
git add internal/ui/views/main_sheet.go internal/ui/views/spellbook.go
git commit -m "feat: integrate dice rolling for spell attacks and damage

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 10: Hit Dice Rolling During Short Rest

**Files:**
- Modify: `internal/ui/views/main_sheet.go`

**Step 1: Add rest hit dice mode**

Add a new mode to track whether the user chose to roll or take average for hit dice. Add a field:

```go
restUseAverage bool // true = use average, false = roll
```

**Step 2: Add roll/average prompt to short rest flow**

In the short rest input handling (around line 814), before the Enter key triggers `performShortRest()`, add a prompt step. When Enter is first pressed, if `restUseAverage` has not been chosen yet, show:

```
Roll or Take Average?
[R]oll  [A]verage
```

After the user picks R or A, then execute the rest with the chosen method.

**Step 3: Update performShortRest() for rolling**

When rolling is chosen, instead of using `avgRoll := (hd.DieType / 2) + 1`, send a `RequestRollMsg` for each hit die:

```go
if !m.restUseAverage {
	// Roll each hit die
	conMod := m.character.AbilityScores.Constitution.Modifier()
	return m, func() tea.Msg {
		return components.RequestRollMsg{
			Label:    "Hit Die",
			DiceExpr: fmt.Sprintf("1d%d", hd.DieType),
			Modifier: conMod,
			RollType: components.RollHitDice,
		}
	}
}
```

Handle `RollCompleteMsg` for `RollHitDice` to apply the healing result and continue spending remaining dice.

**Step 4: Run tests and build**

```bash
go test ./... -v
go build ./cmd/sheet/
```

**Step 5: Commit**

```bash
git add internal/ui/views/main_sheet.go
git commit -m "feat: add roll/average prompt for hit dice during short rest

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 11: Custom Dice Roller Integration

**Files:**
- Modify: `internal/ui/model.go`
- Modify: `internal/ui/views/main_sheet.go`
- Modify: `internal/ui/views/spellbook.go`

**Step 1: Handle OpenCustomRollMsg in model.go**

Add `OpenCustomRollMsg` handling in model.go Update():

```go
case components.OpenCustomRollMsg:
	if m.rollEngine != nil {
		m.rollEngine.OpenCustomRoll()
	}
	return m, nil
```

**Step 2: Add `/` key binding to spellbook**

In `spellbook.go`, add a `/` key binding that sends `OpenCustomRollMsg{}` (same as main sheet).

**Step 3: Add `H` key binding to spellbook for history toggle**

Add a `ToggleRollHistoryMsg` send on `H` key in spellbook.

**Step 4: Update spellbook View to support history column**

Similar to main sheet, the spellbook view needs to account for history column width.

**Step 5: Run tests and build**

```bash
go test ./... -v
go build ./cmd/sheet/
```

**Step 6: Commit**

```bash
git add internal/ui/model.go internal/ui/views/main_sheet.go internal/ui/views/spellbook.go
git commit -m "feat: add custom dice roller and roll history to spellbook view

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 12: Integration Testing and Polish

**Files:**
- Modify: various — polish rendering, fix edge cases

**Step 1: Run full test suite**

```bash
go test ./... -v
```

Fix any failures.

**Step 2: Build and test binary**

```bash
go build ./cmd/sheet/
```

**Step 3: Test key flows manually (if possible)**

Verify these flows work:
- Main sheet → focus Skills → Enter on a skill → advantage prompt → roll → result → dismiss
- Main sheet → focus Saves → Enter → advantage prompt → roll → dismiss
- Main sheet → focus Actions → Enter on weapon → advantage prompt → attack roll → Enter for damage → damage roll → dismiss
- Main sheet → `` ` `` → Luck roll (no prompt) → dismiss
- Main sheet → `/` → custom roll overlay → pick d20, qty 1 → Enter → roll → dismiss
- Main sheet → `H` → history column appears/disappears
- Spellbook → cast spell with damage → roll → dismiss
- Short rest → spend hit die → Roll/Average prompt → roll → healing applied
- Roll history shows all rolls in chronological order
- Nat 20 highlighted in green, Nat 1 in red
- Luck rolls show in purple/magenta
- History column auto-appears after first roll

**Step 4: Fix edge cases**

- Empty character with no weapons/spells
- Terminal width < 80: history column auto-hidden
- Rolling during existing animation: ignored
- Spell with no damage: no roll triggered
- Custom roll 100d100: only show first ~10 dice visually

**Step 5: Commit fixes**

```bash
git add -A
git commit -m "fix: polish dice rolling integration, edge cases, and rendering

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 13: Final PR

**Step 1: Run all tests one final time**

```bash
go test ./...
```

Expected: All pass, no failures.

**Step 2: Push branch and create PR**

```bash
git push -u origin feature/dice-rolling
gh pr create --title "feat: Add dice rolling integration (Phase 12)" --body "$(cat <<'EOF'
## Summary
- Add centralized RollEngine with animated tumbling dice and color flash effects
- Add advantage/disadvantage prompt for all d20 rolls (attacks, skills, saves)
- Add roll history column (toggleable with H) on main sheet and spellbook
- Add Luck roll (backtick key + top of Skills panel) — d20 with no modifier
- Add custom dice roller (/ key) — pick from d4/d6/d8/d10/d12/d20/d100 with quantity 1-100
- Integrate rolling into weapon attacks (with follow-up damage roll), spell casting, skill checks, saving throws
- Add roll/average prompt for hit dice during short rest
- Nat 20/1 highlighting, advantage/disadvantage indicators in history

## Test plan
- [ ] Roll a skill check from the Skills panel (Enter)
- [ ] Roll a saving throw from the Abilities panel (Enter)
- [ ] Roll a weapon attack with follow-up damage (Enter on weapon)
- [ ] Roll Luck (backtick key)
- [ ] Open custom dice roller (/) and roll 3d8
- [ ] Toggle roll history (H)
- [ ] Cast a spell with damage and verify roll triggers
- [ ] Short rest with hit dice — verify roll/average prompt
- [ ] Verify nat 20/1 highlighting
- [ ] Verify advantage/disadvantage works (N/A/D prompt)

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```
