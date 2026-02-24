# Bubble Tea & Lipgloss v2 Upgrade Design

**Date:** 2026-02-24
**Goal:** Upgrade from Bubble Tea v1.3.10 / Lipgloss v1.1.0 / Bubbles v0.21.0 to their v2.0.0 releases, fixing all breaking API changes and adopting selected new features.

**Approach:** Two-phase. Phase 1 handles all mandatory breaking API changes (mechanical migration). Phase 2 adopts new v2 features in focused commits.

---

## Current State

| Package | Current | Target |
|---------|---------|--------|
| `github.com/charmbracelet/bubbletea` | v1.3.10 | `charm.land/bubbletea/v2` v2.0.0 |
| `github.com/charmbracelet/lipgloss` | v1.1.0 | `charm.land/lipgloss/v2` v2.0.0 |
| `github.com/charmbracelet/bubbles` | v0.21.0 | `charm.land/bubbles/v2` v2.0.0 |

### Codebase Impact Summary

| Change Category | Occurrences | Files |
|-----------------|-------------|-------|
| `tea.KeyMsg` → `tea.KeyPressMsg` | ~311 | 22 |
| `lipgloss.Color()` calls | ~315 | 16 |
| `lipgloss.NewStyle()` calls | ~316 | 16 |
| `.Render()` calls | ~532 | 23 |
| `key.NewBinding` / `key.Matches` | ~198 | 6 |
| `tea.Cmd` references | ~108 | 13 |
| `tea.WindowSizeMsg` | ~28 | 17 |
| `msg.Type` field access | ~22 | 6 |
| `msg.Runes` field access | ~15 | 6 |
| `View() string` signatures | 14 | 13 |
| `Init() tea.Cmd` signatures | 9 | 9 |

---

## Phase 1: Mechanical API Migration

### 1.1 Import Path Changes

All files get updated imports:
- `github.com/charmbracelet/bubbletea` → `charm.land/bubbletea/v2`
- `github.com/charmbracelet/lipgloss` → `charm.land/lipgloss/v2`
- `github.com/charmbracelet/bubbles/key` → `charm.land/bubbles/v2/key`

### 1.2 go.mod Changes

Replace direct dependencies, run `go mod tidy` to clean up indirect deps.

### 1.3 Bubble Tea Breaking Changes

| v1 | v2 | Count |
|----|-----|-------|
| `case tea.KeyMsg:` | `case tea.KeyPressMsg:` | ~50 type switches |
| `tea.KeyMsg{...}` in tests | `tea.KeyPressMsg{...}` in tests | ~260 test constructions |
| `msg.Type` | `msg.Code` | 22 |
| `msg.Type == tea.KeyRunes` | `len(msg.Text) > 0` | ~10 |
| `msg.Runes` | `msg.Text` (now `string`) | 15 |
| `string(msg.Runes)` | `msg.Text` | ~10 |
| `msg.Runes[0]` | `rune(msg.Text[0])` | ~3 |
| `for _, r := range msg.Runes` | iterate `msg.Text` | ~3 |
| `View() string` | `View() tea.View` + wrap returns with `tea.NewView(...)` | 14 |
| `tea.WithAltScreen()` in NewProgram | `v.AltScreen = true` in root View() | 1 |
| `tea.WithMouseCellMotion()` in NewProgram | `v.MouseMode = tea.MouseModeCellMotion` in root View() | 1 |

### 1.4 Lipgloss Breaking Changes

| v1 | v2 | Impact |
|----|-----|--------|
| `lipgloss.Color("99")` (type) | `lipgloss.Color("99")` (function returning color.Color) | Syntax unchanged, semantics change. All ~315 calls work as-is. |
| `.Foreground(lipgloss.Color("X"))` | Same syntax | No change needed |
| `lipgloss.JoinVertical/Horizontal` | Same API | No change needed |
| `lipgloss.Place()` | Same API (whitespace options change but we don't use them) | No change needed |
| `lipgloss.Height()` / `lipgloss.Width()` | Same API | No change needed |
| `lipgloss.RoundedBorder()` etc. | Same API | No change needed |

### 1.5 Bubbles/Key Breaking Changes

| v1 | v2 | Impact |
|----|-----|--------|
| Import path | `charm.land/bubbles/v2/key` | 6 files |
| `key.NewBinding`, `key.Matches`, `key.Binding` | Same API | No change |
| `key.WithKeys`, `key.WithHelp` | Same API | No change |

### 1.6 Test Changes

All test files constructing key messages need updating:
- `tea.KeyMsg{Type: tea.KeyEnter}` → `tea.KeyPressMsg{Code: tea.KeyEnter}`
- `tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}` → `tea.KeyPressMsg{Code: 'x', Text: "x"}`
- `tea.KeyMsg{Type: tea.KeyEsc}` → `tea.KeyPressMsg{Code: tea.KeyEscape}` (v2 likely consolidates the constant name)
- `tea.WindowSizeMsg{Width: X, Height: Y}` — unchanged

### 1.7 Sub-View View() Pattern

Sub-views (MainSheetModel, SpellbookModel, etc.) return `string` from their `View()` methods but are NOT `tea.Model` implementations — they're called directly by the root Model. Strategy:

- **Option A:** Keep sub-view View() returning `string`. Only the root `Model.View()` returns `tea.View` and wraps the string with `tea.NewView()`.
- **Option B:** Change all sub-view View() to return `tea.View` too.

**Decision: Option A.** Only `Model.View()` implements the `tea.Model` interface. Sub-views are internal and can keep returning strings. This minimizes changes.

---

## Phase 2: New Feature Adoption

### 2A. Declarative View Configuration

The root `Model.View()` sets View fields instead of using program options:

```go
func (m Model) View() tea.View {
    content := // ... existing rendering logic
    v := tea.NewView(content)
    v.AltScreen = true
    v.MouseMode = tea.MouseModeCellMotion
    return v
}
```

`cmd/sheet/main.go` simplifies to:
```go
p := tea.NewProgram(model)
```

### 2B. Dynamic Window Title

The root `Model.View()` sets `v.WindowTitle` based on current state:

```go
v.WindowTitle = fmt.Sprintf("D&D 5e — %s", m.characterName())
```

Where `characterName()` returns the active character's name, or "Character Selection" at the selection screen.

### 2C. Cursor Control

Views with text input set `v.Cursor` for explicit cursor positioning:
- **Notes editor** — cursor at text insertion point
- **HP input mode** — cursor at end of input buffer
- **Character creation** — cursor at active text field
- **Inventory item search** — cursor at search input

Since sub-views return strings (not tea.View), cursor info must be propagated up. Options:
- Add a `CursorPosition() *tea.Cursor` method to sub-views that need it
- Root Model checks current view and calls it

### 2D. Underline Styles

Replace plain `Underline(true)` with `UnderlineStyle(lipgloss.UnderlineCurly)` for:
- Character name in header
- Focused panel titles
- Active tab indicators

Also use `UnderlineColor()` to match panel accent colors.

### 2E. Hyperlinks on Spell Names

If spell data contains source URLs, use `lipgloss.NewStyle().Hyperlink(url)` to make spell names clickable in terminals that support OSC 8. Gracefully degrades to plain text in unsupported terminals.

---

## Implementation Order

1. **Phase 1:** Single commit — all mechanical breaking changes at once (must be atomic since import path change affects every file)
2. **Phase 2A:** Declarative view config (part of Phase 1 since it's required)
3. **Phase 2B:** Window title
4. **Phase 2C:** Cursor control
5. **Phase 2D:** Underline styles
6. **Phase 2E:** Hyperlinks (only if spell data has URLs)

---

## Testing Strategy

- Run `go vet ./...` and `go test ./... -count=1` after each phase
- Run `go build ./...` to verify compilation
- Manual testing with `go run ./cmd/sheet/` to verify TUI renders correctly
- Test at different terminal sizes to verify responsive layout still works
