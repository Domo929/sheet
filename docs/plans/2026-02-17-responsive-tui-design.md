# Responsive TUI Design

## Goal

Make the TUI adapt gracefully to different terminal sizes by using proportional layouts, vertical scrolling, list pagination, and breakpoint-driven layout changes.

## Architecture

The approach is progressive layout adaptation: views detect available dimensions and switch between layout modes at defined breakpoints. A shared scrollable viewport utility handles vertical overflow. The existing component library (`Panel`, `List`, `Box`) is extended with dimension-awareness rather than replaced.

## Width Breakpoints

| Width | Layout Mode |
|-------|-------------|
| < 60  | "Terminal too small" message |
| 60–79 | Compact: single-column stacked layouts |
| 80–119 | Standard: two-column side-by-side |
| 120+ | Wide: two-column with roll history |

## Height Breakpoints

| Height | Behavior |
|--------|----------|
| < 20   | "Terminal too small" message |
| 20–29  | Compact: abbreviated panels, scrolling |
| 30+    | Full: all panels shown |

## Changes by View

### 1. Minimum Size Guard (model.go)

Add a check in `Model.View()` that displays a "resize your terminal" message when width < 60 or height < 20. This prevents panics from negative width calculations and provides a clear user message.

### 2. Main Sheet (main_sheet.go)

**Width adaptation:**
- **Wide (≥80):** Current two-column layout with proportional right column. Change `leftWidth` from hardcoded 38 to `min(38, width/3)` so it compresses slightly at medium widths.
- **Compact (<80):** Stack all panels vertically in a single column using full width. Order: Header → Abilities/Saves → Combat → Skills → Proficiencies → Actions → Footer.

**Height adaptation:**
- Wrap `mainContent` (the area between header and footer) in a scrollable viewport that tracks a `scrollOffset` field. When content height exceeds available height (total height minus header minus footer), show scroll indicators and allow scrolling with PgUp/PgDn/Ctrl+U/Ctrl+D.
- The scrolling only activates for the main content area — individual focused panels (abilities, skills, combat) still handle their own cursor navigation.

### 3. Inventory (inventory.go)

**Width adaptation:**
- Replace hardcoded `equipWidth=30, currencyWidth=25` with proportional splits:
  - Equipment: 25% of width (min 24)
  - Currency: 20% of width (min 20)
  - Items: remaining width
- **Compact (<80):** Stack panels vertically, each using full width. Only show the focused panel fully; show collapsed summaries for unfocused panels.

### 4. Spellbook (spellbook.go)

**Width adaptation:**
- Already uses proportional 1/3 splits — add minimum column width of 25 chars.
- **Compact (<80):** Switch to two panels (list + details), hiding spell slots panel. Show slot info inline in the header instead.
- **Very compact (<60):** Already blocked by minimum size guard.

### 5. List Component (components/list.go)

Make `Render()` respect the `Height` field:
- Calculate visible item count from Height (subtract title lines)
- Track a `ScrollOffset` field on the List struct
- Show only items from `ScrollOffset` to `ScrollOffset + visibleCount`
- Auto-adjust `ScrollOffset` when `SelectedIndex` moves outside visible range
- Add indicators: `↑ N more` at top and `↓ N more` at bottom when items are clipped

Update `MoveUp()`/`MoveDown()` to adjust `ScrollOffset` when cursor moves outside visible window.

### 6. Character Selection (character_selection.go)

The view already calculates `listHeight = m.height - 10` and sets it on the list, but `List.Render()` ignores it. Once List pagination is implemented (change 5), this view benefits automatically.

### 7. Character Creation (character_creation.go)

No structural changes needed — it uses text inputs that already work at any width. Just ensure the view respects terminal height by limiting visible content.

## Implementation Order

1. **Minimum size guard** — prevents crashes, quick win
2. **List pagination** — enables all list-using views to handle overflow
3. **Main sheet width breakpoints** — biggest visual impact
4. **Main sheet vertical scrolling** — handles height overflow
5. **Inventory proportional widths** — prevents negative width bug
6. **Spellbook compact mode** — handles narrow terminals
7. **Final polish and testing** — verify all views at 60×20, 80×24, 120×40, 200×50

## Testing Strategy

Each change should be tested by:
1. Unit tests for layout logic (e.g., list pagination math, width calculations)
2. Manual testing at multiple terminal sizes using `printf '\e[8;24;80t'` (or similar resize)
3. Verify no panics at minimum supported size (60×20)
