# Bubble Tea & Lipgloss v2 Upgrade Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Upgrade from Bubble Tea v1 / Lipgloss v1 / Bubbles v0.21 to their v2.0 releases, migrating all breaking API changes and adopting new features (window title, cursor control, underline styles, hyperlinks).

**Architecture:** Two-phase approach. Phase 1 is a single atomic commit migrating all breaking changes (import paths, KeyMsg→KeyPressMsg, View()→tea.View, program options→View fields). Phase 2 adds new v2 features in separate commits. The project won't compile between the start and end of Phase 1 since import paths change globally.

**Tech Stack:** Go 1.24, charm.land/bubbletea/v2, charm.land/lipgloss/v2, charm.land/bubbles/v2, testify

---

## Key API Reference (v1 → v2)

### Struct/Type Changes
| v1 | v2 |
|----|-----|
| `tea.KeyMsg` (struct) | `tea.KeyPressMsg` (struct, alias of `Key`) |
| `msg.Type` (tea.KeyType) | `msg.Code` (rune) |
| `msg.Runes` ([]rune) | `msg.Text` (string) |
| `msg.Alt` (bool) | `msg.Mod` (ModifierKey) |
| `tea.KeyRunes` | Check `msg.Text != ""` |
| `View() string` | `View() tea.View` |

### Key Constants (unchanged names)
`tea.KeyEnter`, `tea.KeyEscape`, `tea.KeyTab`, `tea.KeyBackspace`, `tea.KeyUp`, `tea.KeyDown`, `tea.KeyLeft`, `tea.KeyRight`, `tea.KeyPgUp`, `tea.KeyPgDown` — all exist in v2 as rune constants used in `Code` field.

### Removed Constants
| v1 | v2 Equivalent |
|----|---------------|
| `tea.KeyEsc` | `tea.KeyEscape` |
| `tea.KeyShiftTab` | `tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}` |
| `tea.KeyCtrlC` | `tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}` |

### Test Construction Patterns
| v1 | v2 |
|----|-----|
| `tea.KeyMsg{Type: tea.KeyEnter}` | `tea.KeyPressMsg{Code: tea.KeyEnter}` |
| `tea.KeyMsg{Type: tea.KeyEsc}` | `tea.KeyPressMsg{Code: tea.KeyEscape}` |
| `tea.KeyMsg{Type: tea.KeyEscape}` | `tea.KeyPressMsg{Code: tea.KeyEscape}` |
| `tea.KeyMsg{Type: tea.KeyShiftTab}` | `tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}` |
| `tea.KeyMsg{Type: tea.KeyCtrlC}` | `tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}` |
| `tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}` | `tea.KeyPressMsg{Code: 'q', Text: "q"}` |
| `tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1', '0'}}` | `tea.KeyPressMsg{Code: '1', Text: "10"}` |

### msg.String() Changes
| v1 | v2 |
|----|-----|
| `" "` (space) | `"space"` |
| All other keys | Same as v1 |

### Import Paths
| v1 | v2 |
|----|-----|
| `github.com/charmbracelet/bubbletea` | `charm.land/bubbletea/v2` |
| `github.com/charmbracelet/lipgloss` | `charm.land/lipgloss/v2` |
| `github.com/charmbracelet/bubbles/key` | `charm.land/bubbles/v2/key` |

---

## Phase 1: Mechanical API Migration (Single Atomic Commit)

Phase 1 is split into tasks for organization, but they form ONE commit since the project can't compile between steps.

### Task 1: Update go.mod Dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Update dependencies**

```bash
# Remove old dependencies
go get github.com/charmbracelet/bubbletea@none
go get github.com/charmbracelet/lipgloss@none
go get github.com/charmbracelet/bubbles@none

# Add v2 dependencies
go get charm.land/bubbletea/v2@latest
go get charm.land/lipgloss/v2@latest
go get charm.land/bubbles/v2@latest
```

If that doesn't work cleanly, manually edit `go.mod` to replace the three direct dependencies, then run `go mod tidy`.

**Step 2: Verify go.mod looks correct**

The `require` block should have:
```
charm.land/bubbletea/v2 v2.0.0
charm.land/lipgloss/v2 v2.0.0
charm.land/bubbles/v2 v2.0.0
```

And NOT have any `github.com/charmbracelet/bubbletea`, `lipgloss`, or `bubbles` entries.

---

### Task 2: Update Import Paths (All Files)

**Files:** Every `.go` file that imports charmbracelet packages.

**Step 1: Replace import paths globally**

In ALL `.go` files (source + test), replace:
- `"github.com/charmbracelet/bubbletea"` → `"charm.land/bubbletea/v2"`
- `"github.com/charmbracelet/lipgloss"` → `"charm.land/lipgloss/v2"`
- `"github.com/charmbracelet/bubbles/key"` → `"charm.land/bubbles/v2/key"`

Files to update (source):
- `cmd/sheet/main.go`
- `internal/ui/model.go`
- `internal/ui/navigation.go`
- `internal/ui/components/button.go`
- `internal/ui/components/help.go`
- `internal/ui/components/input.go`
- `internal/ui/components/list.go`
- `internal/ui/components/panel.go`
- `internal/ui/components/proficiency_selector.go`
- `internal/ui/components/roll_engine.go`
- `internal/ui/components/roll_helpers.go`
- `internal/ui/components/roll_history.go`
- `internal/ui/views/character_selection.go`
- `internal/ui/views/character_creation.go`
- `internal/ui/views/character_info.go`
- `internal/ui/views/main_sheet.go`
- `internal/ui/views/spellbook.go`
- `internal/ui/views/inventory.go`
- `internal/ui/views/notes_editor.go`
- `internal/ui/views/level_up.go`
- `internal/ui/views/proficiency_selection.go`

Files to update (tests):
- `internal/ui/model_test.go`
- `internal/ui/navigation_test.go`
- `internal/ui/components/list_test.go`
- `internal/ui/components/panel_test.go`
- `internal/ui/components/proficiency_selector_test.go`
- `internal/ui/components/roll_engine_test.go`
- `internal/ui/components/roll_helpers_test.go`
- `internal/ui/views/character_selection_test.go`
- `internal/ui/views/character_info_test.go`
- `internal/ui/views/main_sheet_test.go`
- `internal/ui/views/notes_editor_test.go`
- `internal/ui/views/spellbook_test.go`
- `internal/ui/views/inventory_test.go`
- `internal/ui/views/level_up_test.go`

---

### Task 3: Migrate tea.KeyMsg → tea.KeyPressMsg (Source Files)

**Files:** All source `.go` files that handle key messages.

For each file, apply these transformations:

#### 3a. Type switch cases

Replace all `case tea.KeyMsg:` with `case tea.KeyPressMsg:` in Update functions:
- `internal/ui/model.go:109` — `if keyMsg, ok := msg.(tea.KeyMsg)` → `if keyMsg, ok := msg.(tea.KeyPressMsg)`
- `internal/ui/model.go:346` — `if keyMsg, ok := msg.(tea.KeyMsg)` → `if keyMsg, ok := msg.(tea.KeyPressMsg)`
- `views/main_sheet.go` — `case tea.KeyMsg:` → `case tea.KeyPressMsg:`
- `views/inventory.go` — same
- `views/spellbook.go` — same
- `views/character_info.go` — same
- `views/notes_editor.go` — same
- `views/character_creation.go` — same
- `views/character_selection.go` — same
- `views/level_up.go` — same
- `components/roll_engine.go` — same
- `components/proficiency_selector.go` — same

Also update any function parameter types:
- Search for `func.*tea.KeyMsg` and replace with `tea.KeyPressMsg` in signatures

#### 3b. msg.Type → msg.Code

Replace all `switch msg.Type` with equivalent v2 patterns. For each `switch msg.Type { case tea.KeyX: ... case tea.KeyRunes: ... }` block:

**Pattern:** `switch msg.Type` → `switch msg.Code` for special key cases, and `tea.KeyRunes` case → check `msg.Text != ""`

Example transformation:
```go
// v1
switch msg.Type {
case tea.KeyEscape:
    // handle escape
case tea.KeyEnter:
    // handle enter
case tea.KeyRunes:
    for _, r := range msg.Runes {
        // process rune
    }
}

// v2
switch msg.Code {
case tea.KeyEscape:
    // handle escape
case tea.KeyEnter:
    // handle enter
default:
    if msg.Text != "" {
        for _, r := range msg.Text {
            // process rune
        }
    }
}
```

Apply this to these files:
- `views/inventory.go` — 3 `switch msg.Type` blocks (handleQuantityInput, handleAddItemInput, handleCurrencyInput)
- `views/main_sheet.go` — 4 `switch msg.Type` blocks (FocusAbilitiesAndSaves, FocusSkills, FocusActions, handleHPInput, handleConditionInput, handleCastingInput)
- `views/notes_editor.go` — 2 `switch msg.Type` blocks (handleInputMode, updateEditorMode)
- `views/character_info.go` — 2 `switch msg.Type` blocks (backstory multiline, single-line editing)
- `views/character_creation.go` — 1 `switch msg.Type` block (handlePersonalityKeys)

#### 3c. msg.Type == tea.KeyRunes inline checks

Replace `if msg.Type == tea.KeyRunes` with `if msg.Text != ""`:
- `views/level_up.go:880` — `if msg.Type == tea.KeyRunes` → `if msg.Text != ""`
- `views/level_up.go:885` — `if msg.Type == tea.KeyBackspace` → `if msg.Code == tea.KeyBackspace`
- `views/character_creation.go:400` — `if msg.Type == tea.KeyRunes && (...)` → `if msg.Text != "" && (...)`
- `views/character_creation.go:1056,1064` — same pattern
- `views/character_info.go:328` — `msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'y'` → `msg.Text == "y"`
- `views/character_info.go:332-333` — `msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'n', msg.Type == tea.KeyEsc:` → `msg.Text == "n", msg.Code == tea.KeyEscape:`

#### 3d. msg.Runes → msg.Text

Replace all `msg.Runes` usage:
- `string(msg.Runes)` → `msg.Text` (10 occurrences across 6 files)
- `for _, r := range msg.Runes` → `for _, r := range msg.Text` (3 occurrences)
- `msg.Runes[0] == 'y'` → `msg.Text == "y"` (2 occurrences in character_info.go)

#### 3e. tea.KeyEsc → tea.KeyEscape

Replace all `tea.KeyEsc` with `tea.KeyEscape`:
- `views/character_info.go:280,306,333` — 3 occurrences
- `views/character_creation.go:988` — 1 occurrence

#### 3f. msg.String() space key

Replace `case " ":` with `case "space":` in all `switch msg.String()` blocks:
- `views/character_creation.go:432,620,1228` — `case " ", "enter":` → `case "space", "enter":`
- `components/proficiency_selector.go:84` — `case " ", "enter":` → `case "space", "enter":`
- `views/notes_editor.go` — check if any `case " ":` exists

---

### Task 4: Migrate View() Signatures

**Files:** All files with `View()` methods.

#### 4a. Root Model — returns tea.View

In `internal/ui/model.go`, change `View() string` to `View() tea.View`:

```go
// v1
func (m Model) View() string {
    // ... rendering logic
    return someString
}

// v2
func (m Model) View() tea.View {
    // ... rendering logic
    content := someString // all existing render logic produces a string
    v := tea.NewView(content)
    v.AltScreen = true
    v.MouseMode = tea.MouseModeCellMotion
    return v
}
```

The root View() has multiple return points. Each `return someString` becomes:
```go
v := tea.NewView(someString)
v.AltScreen = true
v.MouseMode = tea.MouseModeCellMotion
return v
```

Create a helper to reduce repetition:
```go
func (m Model) newView(content string) tea.View {
    v := tea.NewView(content)
    v.AltScreen = true
    v.MouseMode = tea.MouseModeCellMotion
    return v
}
```

Then all return points become `return m.newView(someString)`.

#### 4b. Sub-views — keep returning string

Sub-view `View()` methods (MainSheetModel, SpellbookModel, InventoryModel, etc.) are NOT part of the `tea.Model` interface — they're called by the root model. Keep them returning `string`. No changes needed.

#### 4c. Update main.go

Remove program options that moved to View fields:

```go
// v1
p := tea.NewProgram(
    model,
    tea.WithAltScreen(),
    tea.WithMouseCellMotion(),
)

// v2
p := tea.NewProgram(model)
```

---

### Task 5: Migrate Test Files

**Files:** All `_test.go` files.

#### 5a. tea.KeyMsg → tea.KeyPressMsg in tests

Replace ALL test key message constructions using these patterns:

**Special keys:**
```go
// v1                                          // v2
tea.KeyMsg{Type: tea.KeyEnter}                 tea.KeyPressMsg{Code: tea.KeyEnter}
tea.KeyMsg{Type: tea.KeyEsc}                   tea.KeyPressMsg{Code: tea.KeyEscape}
tea.KeyMsg{Type: tea.KeyEscape}                tea.KeyPressMsg{Code: tea.KeyEscape}
tea.KeyMsg{Type: tea.KeyTab}                   tea.KeyPressMsg{Code: tea.KeyTab}
tea.KeyMsg{Type: tea.KeyShiftTab}              tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
tea.KeyMsg{Type: tea.KeyUp}                    tea.KeyPressMsg{Code: tea.KeyUp}
tea.KeyMsg{Type: tea.KeyDown}                  tea.KeyPressMsg{Code: tea.KeyDown}
tea.KeyMsg{Type: tea.KeyLeft}                  tea.KeyPressMsg{Code: tea.KeyLeft}
tea.KeyMsg{Type: tea.KeyRight}                 tea.KeyPressMsg{Code: tea.KeyRight}
tea.KeyMsg{Type: tea.KeyPgUp}                  tea.KeyPressMsg{Code: tea.KeyPgUp}
tea.KeyMsg{Type: tea.KeyPgDown}                tea.KeyPressMsg{Code: tea.KeyPgDown}
tea.KeyMsg{Type: tea.KeyCtrlC}                 tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
tea.KeyMsg{Type: tea.KeyBackspace}             tea.KeyPressMsg{Code: tea.KeyBackspace}
```

**Character keys:**
```go
// v1                                                    // v2
tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}      tea.KeyPressMsg{Code: 'q', Text: "q"}
tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1', '0'}} tea.KeyPressMsg{Code: '1', Text: "10"}
tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}      tea.KeyPressMsg{Code: ' ', Text: " "}
```

Test files to update:
- `internal/ui/model_test.go`
- `internal/ui/navigation_test.go`
- `internal/ui/components/proficiency_selector_test.go`
- `internal/ui/components/roll_engine_test.go`
- `internal/ui/views/character_selection_test.go`
- `internal/ui/views/character_info_test.go`
- `internal/ui/views/main_sheet_test.go`
- `internal/ui/views/notes_editor_test.go`
- `internal/ui/views/spellbook_test.go`
- `internal/ui/views/level_up_test.go`

#### 5b. Update pressKey helper in level_up_test.go

```go
// v1
func pressKey(m *LevelUpModel, keyType tea.KeyType) *LevelUpModel {
    updated, _ := m.Update(tea.KeyMsg{Type: keyType})
    return updated
}

// v2
func pressKey(m *LevelUpModel, code rune) *LevelUpModel {
    updated, _ := m.Update(tea.KeyPressMsg{Code: code})
    return updated
}
```

Update all call sites: `pressKey(m, tea.KeyEsc)` → `pressKey(m, tea.KeyEscape)`, etc.

#### 5c. Update model_test.go View() assertions

The `TestModelViewTooSmallTerminal` and `TestModelViewMinimumSizeOK` tests call `m.View()` and check the string. In v2, `View()` returns `tea.View`. Update assertions:

```go
// v1
view := m.View()
assert.Contains(t, view, "Terminal too small")

// v2
v := m.View()
assert.Contains(t, v.Content, "Terminal too small")
```

Or use `tea.NewView("...").Content` — the `Content` field holds the string.

#### 5d. Update navigation_test.go

The `NavigationHandler` tests use `tea.KeyMsg`. Update all constructions and the `IsKey` function/helper if needed.

---

### Task 6: Update navigation.go Helper

**Files:**
- Modify: `internal/ui/navigation.go`

Read `navigation.go` to understand the `IsKey()` helper and `NavigationHandler`. Update `tea.KeyMsg` references to `tea.KeyPressMsg`. The helper likely extracts `msg.String()` — this should still work in v2 since `KeyPressMsg.String()` exists.

---

### Task 7: Verify Phase 1 Compiles and Passes

**Step 1: Run go vet**
```bash
go vet ./...
```
Expected: No issues.

**Step 2: Run all tests**
```bash
go test ./... -count=1
```
Expected: All PASS. Fix any compilation errors or test failures.

**Step 3: Run build**
```bash
go build ./...
```
Expected: Clean build.

**Step 4: Commit Phase 1**
```bash
git add -A
git commit -m "feat: upgrade to Bubble Tea v2, Lipgloss v2, Bubbles v2

Migrate all breaking API changes:
- Import paths: charm.land/{bubbletea,lipgloss,bubbles}/v2
- tea.KeyMsg → tea.KeyPressMsg (msg.Type→msg.Code, msg.Runes→msg.Text)
- View() string → View() tea.View with declarative view fields
- Program options (WithAltScreen, WithMouseCellMotion) → View fields
- tea.KeyEsc → tea.KeyEscape, space ' ' → 'space' in msg.String()
- All tests updated for v2 API"
```

---

## Phase 2: New Feature Adoption

### Task 8: Dynamic Window Title

**Files:**
- Modify: `internal/ui/model.go`

**Step 1: Update the `newView` helper** to set `WindowTitle`:

```go
func (m Model) newView(content string) tea.View {
    v := tea.NewView(content)
    v.AltScreen = true
    v.MouseMode = tea.MouseModeCellMotion
    v.WindowTitle = m.windowTitle()
    return v
}

func (m Model) windowTitle() string {
    switch m.currentView {
    case ViewCharacterSelection:
        return "D&D 5e Character Sheet"
    case ViewCharacterCreation:
        return "D&D 5e — New Character"
    default:
        if m.character != nil {
            return fmt.Sprintf("D&D 5e — %s", m.character.Info.Name)
        }
        return "D&D 5e Character Sheet"
    }
}
```

**Step 2: Run tests**
```bash
go test ./internal/ui/ -v
```
Expected: PASS

**Step 3: Commit**
```bash
git add internal/ui/model.go
git commit -m "feat: set dynamic window title based on current character

Shows 'D&D 5e — CharacterName' in the terminal title bar, updating
as the user switches characters. Falls back to 'D&D 5e Character Sheet'
at the selection screen."
```

---

### Task 9: Underline Styles for Focused Elements

**Files:**
- Modify: `internal/ui/views/main_sheet.go` (character name header, panel focus indicators)
- Modify: `internal/ui/components/panel.go` (focused panel title)

**Step 1: Update character name underline in main_sheet.go**

Find the `renderHeader` method's title style. Replace:
```go
// v1
titleStyle := lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("15")).
    MarginBottom(0).
    Underline(true)

// v2
titleStyle := lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("15")).
    MarginBottom(0).
    UnderlineStyle(lipgloss.UnderlineCurly).
    UnderlineColor(lipgloss.Color("99"))
```

**Step 2: Update focused panel title styling in relevant components**

Look for places where focused panels use `Underline(true)` and replace with `UnderlineStyle(lipgloss.UnderlineCurly)`. Check:
- `components/panel.go` — Panel.Render() focused state
- `views/spellbook.go` — focused panel border
- `views/inventory.go` — focused panel border

If underlines aren't used for focus indicators (borders are used instead), add curly underlines to focused section headers only where it enhances readability.

**Step 3: Run tests**
```bash
go test ./... -count=1
```
Expected: All PASS

**Step 4: Commit**
```bash
git add internal/ui/views/main_sheet.go internal/ui/components/panel.go
git commit -m "feat: use curly underlines for focused elements

Replace plain underlines with UnderlineStyle(lipgloss.UnderlineCurly)
for character name header and focused panel titles. Gracefully degrades
to regular underline on unsupported terminals."
```

---

### Task 10: Hyperlinks on Spell Names

**Files:**
- Modify: `internal/ui/views/spellbook.go`
- Possibly modify: `internal/data/spell_data.go` (check if URL data exists)

**Step 1: Check if spell data includes URLs**

Read `internal/data/spell_data.go` and the spell JSON data to see if there's a URL or source field. If not, we can construct URLs from spell names using a D&D reference site pattern like `https://www.dndbeyond.com/spells/SPELL-NAME-SLUG`.

**Step 2: Add hyperlink to spell name rendering**

In `spellbook.go`, find where spell names are rendered in the spell list and spell details panel. Add hyperlinks:

```go
// Construct URL from spell name
slug := strings.ToLower(strings.ReplaceAll(spell.Name, " ", "-"))
url := fmt.Sprintf("https://www.dndbeyond.com/spells/%s", slug)

// Apply hyperlink style
nameStyle := lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("99")).
    Hyperlink(url)

nameRendered := nameStyle.Render(spell.Name)
```

This gracefully degrades — terminals that don't support OSC 8 hyperlinks just show the text without the link.

**Step 3: Run tests**
```bash
go test ./internal/ui/views/ -run TestSpellbook -v
go test ./... -count=1
```
Expected: All PASS

**Step 4: Commit**
```bash
git add internal/ui/views/spellbook.go
git commit -m "feat: add terminal hyperlinks to spell names

Spell names in the spellbook view now link to D&D Beyond when the
terminal supports OSC 8 hyperlinks. Gracefully degrades to plain
text in unsupported terminals."
```

---

### Task 11: Cursor Control for Text Input Views

**Files:**
- Modify: `internal/ui/model.go` (propagate cursor from sub-views)
- Modify: `internal/ui/views/main_sheet.go` (HP input cursor)
- Modify: `internal/ui/views/notes_editor.go` (text editing cursor)
- Modify: `internal/ui/views/character_creation.go` (field input cursor)
- Modify: `internal/ui/views/inventory.go` (item search cursor)

**Step 1: Add CursorInfo method to sub-views**

Define a simple interface/pattern for sub-views to report cursor position:

```go
// In model.go or a shared types file
type CursorProvider interface {
    CursorInfo() *tea.Cursor
}
```

**Step 2: Update root Model.View() to use cursor info**

In the `newView` helper, check if the current sub-view provides cursor info:

```go
func (m Model) newView(content string) tea.View {
    v := tea.NewView(content)
    v.AltScreen = true
    v.MouseMode = tea.MouseModeCellMotion
    v.WindowTitle = m.windowTitle()

    // Set cursor from current view if applicable
    if cursor := m.currentCursor(); cursor != nil {
        v.Cursor = cursor
    }
    return v
}

func (m Model) currentCursor() *tea.Cursor {
    switch m.currentView {
    case ViewMainSheet:
        if m.mainSheetModel != nil {
            return m.mainSheetModel.CursorInfo()
        }
    case ViewNotes:
        if m.notesEditorModel != nil {
            return m.notesEditorModel.CursorInfo()
        }
    // ... other views with text input
    }
    return nil
}
```

**Step 3: Implement CursorInfo() in views with text input**

For each view that has text input modes, add a `CursorInfo()` method that returns a `*tea.Cursor` when an input field is active, nil otherwise.

For example, in `main_sheet.go` (HP input):
```go
func (m *MainSheetModel) CursorInfo() *tea.Cursor {
    if m.hpInputMode != HPInputNone {
        // Return cursor at appropriate position
        // Position calculation depends on where the input renders
        return &tea.Cursor{
            Shape: tea.CursorBar,
            Blink: true,
        }
    }
    return nil
}
```

Note: Precise X/Y positioning requires knowing the exact render layout. Start with just shape/blink (cursor shows wherever the terminal thinks it is), and refine position calculation if needed.

**Step 4: Run tests**
```bash
go test ./... -count=1
```
Expected: All PASS

**Step 5: Commit**
```bash
git add internal/ui/model.go internal/ui/views/main_sheet.go internal/ui/views/notes_editor.go internal/ui/views/character_creation.go internal/ui/views/inventory.go
git commit -m "feat: explicit cursor control for text input views

Views with text input (HP, notes, character creation, item search)
now set cursor shape and blink via Bubble Tea v2's cursor API.
Bar cursor with blink for active text fields."
```

---

### Task 12: Final Verification and Polish

**Step 1: Run go vet**
```bash
go vet ./...
```

**Step 2: Run all tests**
```bash
go test ./... -count=1
```

**Step 3: Run build**
```bash
go build ./...
```

**Step 4: Manual smoke test**
```bash
go run ./cmd/sheet/
```
Verify:
- App launches in alt screen
- Character selection renders properly
- Loading a character shows the main sheet
- Navigation between views works
- Window title updates
- Responsive layout still works at different sizes

**Step 5: Commit any fixes**

If any fixes were needed, commit with descriptive messages.
