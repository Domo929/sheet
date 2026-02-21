# Responsive TUI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make the TUI adapt gracefully to different terminal sizes using proportional layouts, vertical scrolling, list pagination, and breakpoint-driven layout changes.

**Architecture:** Views detect available dimensions and switch between layout modes at defined breakpoints (compact <80 cols, standard 80–119, wide 120+). A minimum size guard prevents crashes. The `List` component gains pagination support. Individual views switch from hardcoded widths to proportional calculations.

**Tech Stack:** Go, Bubble Tea (`charmbracelet/bubbletea`), Lipgloss (`charmbracelet/lipgloss`), testify for assertions

---

### Task 1: Minimum Size Guard

Add a terminal-too-small message in `Model.View()` when the terminal is below minimum usable dimensions (60 wide × 20 tall). This prevents panics from negative-width math and gives users a clear message.

**Files:**
- Modify: `internal/ui/model.go:383-418` (View function)
- Modify: `internal/ui/model.go:30-31` (add constants)
- Test: `internal/ui/model_test.go`

**Step 1: Write the failing test**

Open `internal/ui/model_test.go` and add:

```go
func TestModelViewTooSmallTerminal(t *testing.T) {
	// Create a minimal model for testing
	m := Model{
		width:  50,
		height: 15,
	}

	view := m.View()
	assert.Contains(t, view, "Terminal too small", "Should show too-small message for 50x15")
	assert.Contains(t, view, "60", "Should mention minimum width")
	assert.Contains(t, view, "20", "Should mention minimum height")
}

func TestModelViewMinimumSizeOK(t *testing.T) {
	m := Model{
		width:  60,
		height: 20,
	}

	view := m.View()
	assert.NotContains(t, view, "Terminal too small", "Should NOT show too-small message at 60x20")
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/ -run TestModelViewTooSmall -v`
Expected: FAIL — the test expects "Terminal too small" but the current code doesn't produce it.

**Step 3: Implement the minimum size guard**

In `internal/ui/model.go`, add constants near the existing `minRollHistoryWidth`:

```go
// Minimum terminal dimensions for usable display.
const (
	minTerminalWidth  = 60
	minTerminalHeight = 20
)
```

Then in the `View()` method (around line 383), add the guard right after the quitting and error checks:

```go
func (m Model) View() string {
	// Return empty view when quitting to avoid flashing content
	if m.quitting {
		return ""
	}

	if m.err != nil {
		return "Error: " + m.err.Error() + "\n\nPress q to quit."
	}

	// Minimum terminal size guard
	if m.width > 0 && m.height > 0 && (m.width < minTerminalWidth || m.height < minTerminalHeight) {
		return fmt.Sprintf(
			"Terminal too small (%dx%d).\n\nMinimum size: %dx%d\nPlease resize your terminal.",
			m.width, m.height, minTerminalWidth, minTerminalHeight,
		)
	}

	// Route to appropriate view renderer
	// ... rest of switch statement unchanged
```

Note: You'll need to add `"fmt"` to the imports in `model.go` if it's not already there.

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/ui/ -run TestModelView -v`
Expected: PASS

**Step 5: Run all tests**

Run: `go test ./... -count=1`
Expected: All PASS

**Step 6: Commit**

```bash
git add internal/ui/model.go internal/ui/model_test.go
git commit -m "feat: add minimum terminal size guard to prevent layout crashes"
```

---

### Task 2: List Component Pagination

Make `components.List.Render()` respect its `Height` field by showing only a window of items with scroll indicators. This is the foundation that makes character selection and other list-based views work at any height.

**Files:**
- Modify: `internal/ui/components/list.go` (add `ScrollOffset` field, update `Render`, `MoveUp`, `MoveDown`)
- Test: `internal/ui/components/list_test.go`

**Step 1: Write the failing tests**

Add these tests to `internal/ui/components/list_test.go`:

```go
func TestListRenderWithPagination(t *testing.T) {
	// Create a list with 10 items
	items := make([]ListItem, 10)
	for i := range items {
		items[i] = ListItem{Title: fmt.Sprintf("Item %d", i+1)}
	}

	list := NewList("", items) // No title to simplify height calculation
	list.Height = 5            // Only room for 5 items

	rendered := list.Render()

	// Should contain first 5 items
	assert.Contains(t, rendered, "Item 1")
	assert.Contains(t, rendered, "Item 5")

	// Should NOT contain item 6+
	assert.NotContains(t, rendered, "Item 6")

	// Should show "more below" indicator
	assert.Contains(t, rendered, "↓")
}

func TestListRenderPaginationScrollsWithSelection(t *testing.T) {
	items := make([]ListItem, 10)
	for i := range items {
		items[i] = ListItem{Title: fmt.Sprintf("Item %d", i+1)}
	}

	list := NewList("", items)
	list.Height = 5

	// Move selection to item 7 (index 6)
	for i := 0; i < 6; i++ {
		list.MoveDown()
	}

	rendered := list.Render()

	// Should contain item 7 (the selected one)
	assert.Contains(t, rendered, "Item 7")

	// Should show "more above" indicator
	assert.Contains(t, rendered, "↑")
}

func TestListRenderNoPaginationWhenAllFit(t *testing.T) {
	items := []ListItem{
		{Title: "Item 1"},
		{Title: "Item 2"},
		{Title: "Item 3"},
	}

	list := NewList("", items)
	list.Height = 10 // Plenty of room

	rendered := list.Render()

	// Should contain all items
	assert.Contains(t, rendered, "Item 1")
	assert.Contains(t, rendered, "Item 3")

	// Should NOT show scroll indicators
	assert.NotContains(t, rendered, "↑")
	assert.NotContains(t, rendered, "↓")
}

func TestListRenderNoPaginationWhenHeightZero(t *testing.T) {
	items := make([]ListItem, 10)
	for i := range items {
		items[i] = ListItem{Title: fmt.Sprintf("Item %d", i+1)}
	}

	list := NewList("", items)
	// Height = 0 (default) means show all, for backwards compat

	rendered := list.Render()

	// Should contain all 10 items
	assert.Contains(t, rendered, "Item 1")
	assert.Contains(t, rendered, "Item 10")
}

func TestListScrollOffsetAdjustsOnMoveDown(t *testing.T) {
	items := make([]ListItem, 10)
	for i := range items {
		items[i] = ListItem{Title: fmt.Sprintf("Item %d", i+1)}
	}

	list := NewList("", items)
	list.Height = 3

	// Move down past visible window
	list.MoveDown() // index 1
	list.MoveDown() // index 2
	list.MoveDown() // index 3 — should scroll

	assert.Equal(t, 3, list.SelectedIndex)
	assert.True(t, list.ScrollOffset > 0, "ScrollOffset should increase when selection moves past visible area")
}

func TestListScrollOffsetAdjustsOnMoveUp(t *testing.T) {
	items := make([]ListItem, 10)
	for i := range items {
		items[i] = ListItem{Title: fmt.Sprintf("Item %d", i+1)}
	}

	list := NewList("", items)
	list.Height = 3
	list.SelectedIndex = 5
	list.ScrollOffset = 4

	// Move up past visible window top
	list.MoveUp() // index 4
	list.MoveUp() // index 3 — should scroll up

	assert.Equal(t, 3, list.SelectedIndex)
	assert.True(t, list.ScrollOffset <= 3, "ScrollOffset should decrease when selection moves above visible area")
}

func TestListPaginationWithTitle(t *testing.T) {
	items := make([]ListItem, 10)
	for i := range items {
		items[i] = ListItem{Title: fmt.Sprintf("Item %d", i+1)}
	}

	list := NewList("My Title", items)
	list.Height = 7 // Title takes 2 lines (title + blank), so ~5 items fit

	rendered := list.Render()

	// Should contain the title
	assert.Contains(t, rendered, "My Title")

	// Should NOT contain item 6+ (title takes 2 lines, leaving 5 for items)
	// Items 1-5 should be visible, but we may see at most 5 items
	assert.Contains(t, rendered, "Item 1")
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/ui/components/ -run TestListRenderWithPagination -v`
Expected: FAIL — current `Render()` shows all items regardless of Height.

**Step 3: Implement list pagination**

Modify `internal/ui/components/list.go`:

1. Add `ScrollOffset` field to the `List` struct:

```go
// List is a selectable list component.
type List struct {
	Items         []ListItem
	SelectedIndex int
	ScrollOffset  int
	Width         int
	Height        int
	Title         string
}
```

2. Update `MoveUp()` and `MoveDown()` to adjust `ScrollOffset`:

```go
// MoveUp moves the selection up.
func (l *List) MoveUp() {
	if l.SelectedIndex > 0 {
		l.SelectedIndex--
		l.ensureVisible()
	}
}

// MoveDown moves the selection down.
func (l *List) MoveDown() {
	if l.SelectedIndex < len(l.Items)-1 {
		l.SelectedIndex++
		l.ensureVisible()
	}
}

// visibleItemCount returns how many items can be displayed given the current Height.
// Returns len(Items) if Height is 0 (no pagination).
func (l *List) visibleItemCount() int {
	if l.Height <= 0 {
		return len(l.Items)
	}

	available := l.Height

	// Title takes 2 lines (title text + blank line)
	if l.Title != "" {
		available -= 2
	}

	// Reserve 1 line each for scroll indicators if needed
	hasAbove := l.ScrollOffset > 0
	hasBelow := l.ScrollOffset+available < len(l.Items)
	if hasAbove {
		available--
	}
	if hasBelow {
		available--
	}

	if available < 1 {
		available = 1
	}

	return available
}

// ensureVisible adjusts ScrollOffset so that SelectedIndex is within the visible window.
func (l *List) ensureVisible() {
	if l.Height <= 0 {
		return
	}

	visible := l.visibleItemCount()

	// If selected is above the visible window, scroll up
	if l.SelectedIndex < l.ScrollOffset {
		l.ScrollOffset = l.SelectedIndex
	}

	// If selected is below the visible window, scroll down
	if l.SelectedIndex >= l.ScrollOffset+visible {
		l.ScrollOffset = l.SelectedIndex - visible + 1
	}

	// Clamp scroll offset
	maxOffset := len(l.Items) - visible
	if maxOffset < 0 {
		maxOffset = 0
	}
	if l.ScrollOffset > maxOffset {
		l.ScrollOffset = maxOffset
	}
	if l.ScrollOffset < 0 {
		l.ScrollOffset = 0
	}
}
```

3. Replace the `Render()` method:

```go
// Render renders the list as a string.
func (l List) Render() string {
	if len(l.Items) == 0 {
		return "No items"
	}

	var b strings.Builder

	// Add title if present
	if l.Title != "" {
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99"))
		b.WriteString(titleStyle.Render(l.Title))
		b.WriteString("\n\n")
	}

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	indicatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))

	// Determine visible range
	l.ensureVisible()
	visibleCount := l.visibleItemCount()
	startIdx := l.ScrollOffset
	endIdx := startIdx + visibleCount
	if endIdx > len(l.Items) {
		endIdx = len(l.Items)
	}

	// "More above" indicator
	if startIdx > 0 {
		b.WriteString(indicatorStyle.Render(fmt.Sprintf("  ↑ %d more", startIdx)))
		b.WriteString("\n")
	}

	for i := startIdx; i < endIdx; i++ {
		item := l.Items[i]
		cursor := "  "
		style := normalStyle

		if i == l.SelectedIndex {
			cursor = "> "
			style = selectedStyle
		}

		line := fmt.Sprintf("%s%s", cursor, item.Title)
		if item.Description != "" {
			line = fmt.Sprintf("%s - %s", line, item.Description)
		}

		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	// "More below" indicator
	if endIdx < len(l.Items) {
		b.WriteString(indicatorStyle.Render(fmt.Sprintf("  ↓ %d more", len(l.Items)-endIdx)))
		b.WriteString("\n")
	}

	return b.String()
}
```

**Step 4: Run the pagination tests**

Run: `go test ./internal/ui/components/ -run TestList -v`
Expected: PASS

**Step 5: Run all tests**

Run: `go test ./... -count=1`
Expected: All PASS. The existing `TestListRender` and `TestListEmptyRender` tests should still pass since Height=0 means "show all" (backwards compatible).

**Step 6: Commit**

```bash
git add internal/ui/components/list.go internal/ui/components/list_test.go
git commit -m "feat: add pagination support to List component

List.Render() now respects the Height field. When Height > 0, only a
window of items is shown with scroll indicators. MoveUp/MoveDown auto-
adjust ScrollOffset to keep the selected item visible. Height=0 retains
the old behavior of showing all items."
```

---

### Task 3: Main Sheet Width Breakpoints

Make the main sheet switch between two-column (wide) and single-column (compact) layouts based on terminal width. At widths below 80, stack all panels vertically. Above 80, use the current side-by-side layout with proportional column widths.

**Files:**
- Modify: `internal/ui/views/main_sheet.go:1173-1264` (View function)
- Test: `internal/ui/views/main_sheet_test.go`

**Step 1: Write the failing test**

Add to `internal/ui/views/main_sheet_test.go`:

```go
func TestMainSheetCompactLayout(t *testing.T) {
	char := models.NewCharacter("test-1", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage(t.TempDir())

	m := NewMainSheetModel(char, store)

	// Simulate a narrow terminal
	sizeMsg := tea.WindowSizeMsg{Width: 70, Height: 30}
	m, _ = m.Update(sizeMsg)

	view := m.View()

	// In compact mode, all panels should still render
	assert.Contains(t, view, "Abilities", "Compact view should contain abilities")
	assert.Contains(t, view, "Skills", "Compact view should contain skills")
	// View should be non-empty and not panic
	assert.True(t, len(view) > 0, "Compact view should render content")
}

func TestMainSheetWideLayout(t *testing.T) {
	char := models.NewCharacter("test-1", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage(t.TempDir())

	m := NewMainSheetModel(char, store)

	// Simulate a wide terminal
	sizeMsg := tea.WindowSizeMsg{Width: 140, Height: 40}
	m, _ = m.Update(sizeMsg)

	view := m.View()

	// Wide view should render normally
	assert.Contains(t, view, "Abilities", "Wide view should contain abilities")
	assert.Contains(t, view, "Skills", "Wide view should contain skills")
	assert.True(t, len(view) > 0, "Wide view should render content")
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/ui/views/ -run TestMainSheetCompactLayout -v`
Expected: May PASS or FAIL depending on existing behavior — but the compact layout won't look right. If the test passes (since it just checks content exists), proceed to step 3 and ensure the visual behavior is correct.

**Step 3: Implement width breakpoints**

Replace the layout logic in `MainSheetModel.View()` (the section between header rendering and footer rendering, approximately lines 1191-1227):

```go
func (m *MainSheetModel) View() string {
	if m.character == nil {
		return "No character loaded"
	}

	// Calculate available width
	width := m.width
	if width == 0 {
		width = 140
	}
	height := m.height
	if height == 0 {
		height = 40
	}

	// Render sections
	header := m.renderHeader(width)

	// Width breakpoint: compact (<80) vs standard (≥80)
	const compactBreakpoint = 80

	var mainContent string

	if width < compactBreakpoint {
		// Compact layout: single column, all panels stacked vertically
		panelWidth := width - 2 // Leave small margin

		abilitiesAndSaves := m.renderAbilitiesAndSaves(panelWidth)
		combat := m.renderCombatStats(panelWidth)
		skills := m.renderSkills(panelWidth)
		proficiencies := m.renderProficiencies(panelWidth)
		actions := m.renderActions(panelWidth)

		mainContent = lipgloss.JoinVertical(lipgloss.Left,
			abilitiesAndSaves,
			combat,
			skills,
			proficiencies,
			actions,
		)
	} else {
		// Standard two-column layout
		// Left column: proportional, capped at 38
		leftWidth := width / 3
		if leftWidth > 38 {
			leftWidth = 38
		}
		if leftWidth < 28 {
			leftWidth = 28
		}

		historyReserved := 0
		if m.rollHistoryVisible {
			historyReserved = m.rollHistoryWidth
		}
		rightWidth := width - leftWidth - 4 - historyReserved // 4 for gap

		// Ensure minimum right column width
		if rightWidth < 35 {
			rightWidth = 35
		}

		// Left column: Abilities/Saves on top, Proficiencies, then Skills below
		abilitiesAndSaves := m.renderAbilitiesAndSaves(leftWidth)
		proficiencies := m.renderProficiencies(leftWidth)
		skills := m.renderSkills(leftWidth)
		leftColumn := lipgloss.JoinVertical(lipgloss.Left, abilitiesAndSaves, proficiencies, skills)

		// Right column: Combat on top, Actions below
		combat := m.renderCombatStats(rightWidth)
		actions := m.renderActions(rightWidth)
		rightColumn := lipgloss.JoinVertical(lipgloss.Left, combat, actions)

		// Join columns horizontally
		mainContent = lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftColumn,
			"  ",
			rightColumn,
		)
	}

	// Footer with navigation help
	footer := m.renderFooter(width)

	// Spell casting modal overlay
	if m.castingSpell != nil {
		castingModal := m.renderCastConfirmationModal()
		return lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			castingModal,
			footer,
		)
	}

	// Rest overlay if in rest mode
	if m.restMode != RestModeNone {
		restOverlay := m.renderRestOverlay(width)
		return lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			restOverlay,
			footer,
		)
	}

	// Join all sections vertically
	fullView := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		mainContent,
		footer,
	)

	return fullView
}
```

**Step 4: Run tests**

Run: `go test ./internal/ui/views/ -run TestMainSheet -v`
Expected: PASS

**Step 5: Run all tests**

Run: `go test ./... -count=1`
Expected: All PASS

**Step 6: Commit**

```bash
git add internal/ui/views/main_sheet.go internal/ui/views/main_sheet_test.go
git commit -m "feat: add compact single-column layout for main sheet at narrow widths

Below 80 columns, panels stack vertically in a single column.
Above 80, proportional two-column layout with leftWidth capped at 38
and floor at 28."
```

---

### Task 4: Main Sheet Vertical Scrolling

Add vertical scrolling to the main sheet's content area so it works at short terminal heights. Track a `scrollOffset` on the model and use PgUp/PgDn/Ctrl+U/Ctrl+D to scroll.

**Files:**
- Modify: `internal/ui/views/main_sheet.go` (add scrollOffset field, update View and Update)
- Test: `internal/ui/views/main_sheet_test.go`

**Step 1: Write the failing test**

Add to `internal/ui/views/main_sheet_test.go`:

```go
func TestMainSheetScrolling(t *testing.T) {
	char := models.NewCharacter("test-1", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage(t.TempDir())

	m := NewMainSheetModel(char, store)

	// Simulate a short terminal that will need scrolling
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 25}
	m, _ = m.Update(sizeMsg)

	// Initial scroll offset should be 0
	assert.Equal(t, 0, m.scrollOffset, "Initial scroll offset should be 0")

	// Page down should increase scroll offset
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	// scrollOffset may or may not change depending on content height
	// Just verify no panic and offset >= 0
	assert.GreaterOrEqual(t, m.scrollOffset, 0, "Scroll offset should be >= 0 after PgDn")
}

func TestMainSheetScrollDoesNotGoBelowZero(t *testing.T) {
	char := models.NewCharacter("test-1", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage(t.TempDir())

	m := NewMainSheetModel(char, store)
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 25}
	m, _ = m.Update(sizeMsg)

	// Page up from 0 should stay at 0
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	assert.Equal(t, 0, m.scrollOffset, "Scroll offset should not go below 0")
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/ui/views/ -run TestMainSheetScroll -v`
Expected: FAIL — `m.scrollOffset` field doesn't exist yet.

**Step 3: Implement vertical scrolling**

3a. Add the `scrollOffset` field to `MainSheetModel` (near line 63):

```go
	// Scroll offset for main content area
	scrollOffset int
```

3b. Add scroll key bindings to the `mainSheetKeyMap` struct and `defaultMainSheetKeyMap()`:

In the struct (find the `mainSheetKeyMap` struct definition):
```go
	PageDown key.Binding
	PageUp   key.Binding
```

In `defaultMainSheetKeyMap()`:
```go
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("PgDn", "scroll down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("PgUp", "scroll up"),
		),
```

3c. Add scroll handling in `Update()`. Add this in the key matching section (after the existing `case key.Matches(...)` statements, around line 495-520):

```go
		case key.Matches(msg, m.keys.PageDown):
			m.scrollOffset += 10
			return m, nil
		case key.Matches(msg, m.keys.PageUp):
			m.scrollOffset -= 10
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}
			return m, nil
```

3d. In `View()`, after constructing `mainContent` and before building `fullView`, add scroll clamping and line-based viewport cropping:

```go
	// Apply vertical scrolling to main content
	contentLines := strings.Split(mainContent, "\n")
	headerHeight := lipgloss.Height(header)
	footerHeight := lipgloss.Height(footer)
	availableHeight := height - headerHeight - footerHeight

	if availableHeight > 0 && len(contentLines) > availableHeight {
		// Clamp scroll offset
		maxScroll := len(contentLines) - availableHeight
		if maxScroll < 0 {
			maxScroll = 0
		}
		if m.scrollOffset > maxScroll {
			m.scrollOffset = maxScroll
		}

		// Crop to visible window
		endIdx := m.scrollOffset + availableHeight
		if endIdx > len(contentLines) {
			endIdx = len(contentLines)
		}
		mainContent = strings.Join(contentLines[m.scrollOffset:endIdx], "\n")
	} else {
		// Content fits, reset scroll
		m.scrollOffset = 0
	}
```

You'll need to add `"strings"` to imports if not already present (it should be).

**Step 4: Run scroll tests**

Run: `go test ./internal/ui/views/ -run TestMainSheetScroll -v`
Expected: PASS

**Step 5: Run all tests**

Run: `go test ./... -count=1`
Expected: All PASS

**Step 6: Commit**

```bash
git add internal/ui/views/main_sheet.go internal/ui/views/main_sheet_test.go
git commit -m "feat: add vertical scrolling to main sheet view

PgUp/PgDn/Ctrl+U/Ctrl+D scroll the main content area when it exceeds
the terminal height. Content is line-cropped to fit the available space
between header and footer."
```

---

### Task 5: Inventory Proportional Widths

Replace the inventory's hardcoded column widths (30/25) with proportional calculations that adapt to terminal width. Add a compact stacked layout for narrow terminals.

**Files:**
- Modify: `internal/ui/views/inventory.go:900-938` (View function)
- Test: `internal/ui/views/inventory.go` (add test near existing tests, or in a new file)

**Step 1: Write the failing test**

If there is an existing inventory test file, add there. Otherwise create tests inline. Check for `internal/ui/views/inventory_test.go` — if it doesn't exist, the test goes in a new file. Add:

```go
// In internal/ui/views/inventory_test.go (create if needed)
package views

import (
	"testing"

	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestInventoryNarrowWidth(t *testing.T) {
	char := models.NewCharacter("test-1", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage(t.TempDir())

	m := NewInventoryModel(char, store)

	// Simulate a narrow terminal
	sizeMsg := tea.WindowSizeMsg{Width: 65, Height: 30}
	m, _ = m.Update(sizeMsg)

	view := m.View()

	// Should render without panic
	assert.True(t, len(view) > 0, "Narrow inventory should render content")
	// Should contain inventory content
	assert.Contains(t, view, "Inventory", "Should contain inventory header")
}

func TestInventoryWideWidth(t *testing.T) {
	char := models.NewCharacter("test-1", "Test", "Human", "Fighter")
	store, _ := storage.NewCharacterStorage(t.TempDir())

	m := NewInventoryModel(char, store)

	// Simulate a wide terminal
	sizeMsg := tea.WindowSizeMsg{Width: 150, Height: 40}
	m, _ = m.Update(sizeMsg)

	view := m.View()

	assert.True(t, len(view) > 0, "Wide inventory should render content")
	assert.Contains(t, view, "Inventory", "Should contain inventory header")
}
```

**Step 2: Run tests to verify behavior**

Run: `go test ./internal/ui/views/ -run TestInventory -v`
Expected: The narrow test may panic or produce garbled output due to negative `itemsWidth`.

**Step 3: Implement proportional widths**

Replace the width calculation block in `InventoryModel.View()` (lines 917-924):

```go
	// Three-column layout: Equipment | Items | Currency
	// Use proportional widths with minimums
	const compactBreakpoint = 80

	if width < compactBreakpoint {
		// Compact layout: stack panels vertically
		panelWidth := width - 4

		var sections []string
		sections = append(sections, m.renderEquipment(panelWidth))
		sections = append(sections, m.renderItems(panelWidth))
		sections = append(sections, m.renderCurrency(panelWidth))

		columns := lipgloss.JoinVertical(lipgloss.Left, sections...)

		// Add item overlay
		if m.addingItem {
			overlay := m.renderAddItemOverlay(width)
			columns = lipgloss.JoinVertical(lipgloss.Left, columns, "", overlay)
		}

		footer := m.renderFooter(width)
		return lipgloss.JoinVertical(lipgloss.Left, header, "", columns, "", footer)
	}

	// Standard layout: three columns with proportional widths
	equipWidth := width * 25 / 100 // 25%
	if equipWidth < 24 {
		equipWidth = 24
	}
	currencyWidth := width * 20 / 100 // 20%
	if currencyWidth < 20 {
		currencyWidth = 20
	}
	itemsWidth := width - equipWidth - currencyWidth - 6 // borders and padding
	if itemsWidth < 20 {
		itemsWidth = 20
	}

	equipment := m.renderEquipment(equipWidth)
	items := m.renderItems(itemsWidth)
	currency := m.renderCurrency(currencyWidth)

	columns := lipgloss.JoinHorizontal(lipgloss.Top, equipment, items, currency)
```

Keep the rest of the function (add item overlay, footer) unchanged.

**Step 4: Run tests**

Run: `go test ./internal/ui/views/ -run TestInventory -v`
Expected: PASS — no more panics at narrow widths.

**Step 5: Run all tests**

Run: `go test ./... -count=1`
Expected: All PASS

**Step 6: Commit**

```bash
git add internal/ui/views/inventory.go internal/ui/views/inventory_test.go
git commit -m "feat: proportional column widths for inventory view

Replace hardcoded equipWidth=30, currencyWidth=25 with proportional
calculations (25%/20%/remaining). Below 80 columns, panels stack
vertically. Prevents negative itemsWidth at narrow terminals."
```

---

### Task 6: Spellbook Compact Mode

Make the spellbook switch from three-panel to two-panel layout at narrow widths, and ensure minimum column widths prevent unusable panels.

**Files:**
- Modify: `internal/ui/views/spellbook.go:339-406` (View function)
- Test: `internal/ui/views/spellbook_test.go`

**Step 1: Write the failing test**

Add to `internal/ui/views/spellbook_test.go`:

```go
func TestSpellbookCompactLayout(t *testing.T) {
	char := models.NewCharacter("test-1", "Test", "Elf", "Wizard")
	sc := models.NewSpellcasting(models.AbilityIntelligence)
	char.Spellcasting = &sc
	store, _ := storage.NewCharacterStorage(t.TempDir())
	loader := data.NewLoader("../../../data")

	m := NewSpellbookModel(char, store, loader)

	// Simulate a narrow terminal
	sizeMsg := tea.WindowSizeMsg{Width: 70, Height: 30}
	m, _ = m.Update(sizeMsg)

	view := m.View()

	// Should render without panic
	assert.True(t, len(view) > 0, "Compact spellbook should render")
	assert.Contains(t, view, "Spellbook", "Should contain spellbook header")
}
```

**Step 2: Run test to verify it fails or produces bad output**

Run: `go test ./internal/ui/views/ -run TestSpellbookCompactLayout -v`
Expected: May produce garbled output because each panel gets ~23 chars (70/3).

**Step 3: Implement compact mode**

Replace the panel layout section in `SpellbookModel.View()` (lines 358-382):

```go
	// Layout: Header at top, then panels
	const compactBreakpoint = 90

	// Subtract roll history width if visible
	availableWidth := m.width
	if m.rollHistoryVisible {
		availableWidth -= m.rollHistoryWidth
	}

	panelHeight := m.height - lipgloss.Height(header) - 2

	var panels string

	if availableWidth < compactBreakpoint {
		// Compact: two panels (list + details), spell slots info in header
		listWidth := availableWidth * 40 / 100 // 40%
		if listWidth < 25 {
			listWidth = 25
		}
		detailsWidth := availableWidth - listWidth

		spellListStyled := lipgloss.NewStyle().
			Width(listWidth).
			Height(panelHeight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12")).
			Render(spellList)

		spellDetailsStyled := lipgloss.NewStyle().
			Width(detailsWidth).
			Height(panelHeight).
			Border(lipgloss.RoundedBorder()).
			Render(spellDetails)

		panels = lipgloss.JoinHorizontal(lipgloss.Top, spellListStyled, spellDetailsStyled)
	} else {
		// Standard: three panels (list | details | slots)
		listWidth := availableWidth / 3
		detailsWidth := availableWidth / 3
		slotsWidth := availableWidth - listWidth - detailsWidth

		spellListStyled := lipgloss.NewStyle().
			Width(listWidth).
			Height(panelHeight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12")).
			Render(spellList)

		spellDetailsStyled := lipgloss.NewStyle().
			Width(detailsWidth).
			Height(panelHeight).
			Border(lipgloss.RoundedBorder()).
			Render(spellDetails)

		spellSlotsStyled := lipgloss.NewStyle().
			Width(slotsWidth).
			Height(panelHeight).
			Border(lipgloss.RoundedBorder()).
			Render(spellSlots)

		panels = lipgloss.JoinHorizontal(lipgloss.Top, spellListStyled, spellDetailsStyled, spellSlotsStyled)
	}
```

The rest of the function (footer, overlays) stays the same.

**Step 4: Run tests**

Run: `go test ./internal/ui/views/ -run TestSpellbook -v`
Expected: PASS

**Step 5: Run all tests**

Run: `go test ./... -count=1`
Expected: All PASS

**Step 6: Commit**

```bash
git add internal/ui/views/spellbook.go internal/ui/views/spellbook_test.go
git commit -m "feat: compact two-panel layout for spellbook at narrow widths

Below 90 columns, spellbook uses list+details panels only (40%/60%
split). Above 90, keeps the standard three-panel layout. Prevents
panels from being too narrow to read."
```

---

### Task 7: Final Verification and Polish

Run all tests, vet, and build. Verify no regressions. Fix any issues found.

**Files:**
- Potentially any file that needs minor fixes

**Step 1: Run go vet**

Run: `go vet ./...`
Expected: No issues

**Step 2: Run all tests**

Run: `go test ./... -count=1`
Expected: All PASS

**Step 3: Run build**

Run: `go build ./...`
Expected: Clean build with no errors

**Step 4: Commit any fixes**

If any fixes were needed, commit them with descriptive messages.

**Step 5: Final commit (if no fixes needed)**

No commit needed if everything passes clean.
