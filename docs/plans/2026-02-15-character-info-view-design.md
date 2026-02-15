# Phase 11: Character Info, Notes & Proficiencies — Design

## Overview

Phase 11 adds three independent deliverables to the character sheet TUI:

1. **Proficiencies Panel** — A new read-only panel on the main sheet showing armor, weapon, tool, and language proficiencies
2. **Character Info View** — A full view for personality (traits, ideals, bonds, flaws, backstory) with editing, plus a features display grouped by source
3. **Notes Editor** — A multi-document notebook with a document list and full-screen editor
4. **Character Creation Personality Step** — An optional personality/backstory step added to the creation wizard

These are designed as independent pieces that can be built and merged separately.

---

## Part 1: Proficiencies Panel on Main Sheet

### Location

Left column, inserted between Abilities & Saving Throws and Skills. The left column order becomes:

1. Abilities & Saving Throws
2. Proficiencies (new)
3. Skills

### Layout

```
┌──── Proficiencies ───────────────┐
│ Armor: Light, Medium, Shields   │
│ Weapons: Simple, Martial        │
│ Tools: Smith's Tools            │
│ Languages: Common, Dwarvish     │
└──────────────────────────────────┘
```

### Behavior

- **Read-only** — proficiencies are set during character creation and level-up, not edited from the main sheet
- **No focus state** — purely informational. Tab cycling continues between the existing 4 panels (Abilities, Skills, Combat, Actions)
- **Compact rendering** — each category on one line, comma-separated. If values overflow panel width, they wrap to the next line
- **Omit empty categories** — if a category has no proficiencies, the row is omitted entirely

### Implementation

- Add `renderProficiencies()` method to `MainSheetModel`
- Insert its output between `renderAbilitiesAndSaves()` and `renderSkills()` in the left column layout
- No new key bindings, state, or messages needed

### Files

- Modify: `internal/ui/views/main_sheet.go`

---

## Part 2: Character Info View

### Access

Press `c` from the main sheet to open. Press `Esc` to return to the main sheet.

### Layout — Two Panel Design

```
┌──── Personality ─────────────────────┐  ┌──── Features ────────────────────────────┐
│                                      │  │                                          │
│ Traits:                              │  │ [Racial] [Class] [Subclass] [Feats]      │
│   • Loyal to a fault                 │  │                                          │
│   • Always has a joke ready          │  │ ── Racial Traits (Dwarf) ──              │
│                                      │  │ > Darkvision                             │
│ Ideals:                              │  │   Dwarven Resilience                     │
│   • Justice above all                │  │   Stonecunning                           │
│                                      │  │                                          │
│ Bonds:                               │  │ ──────────────────────────────────────── │
│   • My family's smithy               │  │ Darkvision                               │
│                                      │  │ You can see in dim light within 60 feet  │
│ Flaws:                               │  │ of you as if it were bright light, and   │
│   • Too trusting of authority        │  │ in darkness as if it were dim light.     │
│                                      │  │                                          │
│ Backstory:                           │  │ Source: Dwarf                            │
│   Born in the mountain halls of...   │  │                                          │
│   (truncated, press Enter to expand) │  │                                          │
│                                      │  │                                          │
│ [e]dit  [a]dd  [d]elete             │  │                                          │
└──────────────────────────────────────┘  └──────────────────────────────────────────┘
```

### Left Panel — Personality

**Display:**
- Personality traits, ideals, bonds, flaws displayed as bulleted lists under headers
- Backstory shown below, word-wrapped and truncated to ~3 lines with "press Enter to expand" if longer

**Editing (modal overlay pattern):**
- `e` — edit the currently focused item (opens modal text input overlay, prepopulated with existing text)
- `a` — add a new item to the currently focused section (opens empty modal text input)
- `d` — delete the currently focused item (with confirmation)
- `Up/Down` — navigate between items across all sections
- `Enter` on backstory — opens full expanded view of backstory text; editing backstory uses `e`

**Edit modal:**
- Same pattern as HP input / spell search — a centered lipgloss box with a text buffer
- `Enter` confirms, `Esc` cancels
- For backstory editing (multiline): `Ctrl+S` saves (since `Enter` adds a newline)

### Right Panel — Features

**Display:**
- Tab-style header row: `[Racial] [Class] [Subclass] [Feats]` — switch between categories
- Top half: scrollable list of features in the selected category
- Bottom half: detail pane showing the full description of the highlighted feature
- Class features sorted by level (e.g., "Fighting Style [Level 1]", "Action Surge [Level 2]")

**Behavior:**
- `Up/Down` — navigate feature list (when right panel focused)
- `Left/Right` — switch feature category tabs (when right panel focused)
- Read-only — features are added during character creation and level-up, not edited here

### Navigation

- `Tab` / `Shift+Tab` — switch focus between Personality (left) and Features (right) panels
- `Esc` — return to main sheet (sends `BackToSheetMsg`)
- `n` — open Notes Editor from here (sends `OpenNotesMsg`)

### Messages

- `OpenCharacterInfoMsg{}` — sent by main sheet `c` key
- Returns via existing `BackToSheetMsg`

### Files

- New: `internal/ui/views/character_info.go`
- New: `internal/ui/views/character_info_test.go`
- Modify: `internal/ui/model.go` — add `characterInfoModel` field, handle `OpenCharacterInfoMsg`, route view
- Modify: `internal/ui/views/main_sheet.go` — change `c` key to send `OpenCharacterInfoMsg`

---

## Part 3: Notes Editor — Multi-Document Notebook

### Access

Press `n` from the main sheet or Character Info view to open the Notes view.

### Mode 1: Document List

```
┌──── Notes ── Thalion Brightblade ──────────────────────────────────────┐
│                                                                        │
│  Sort: [Last Edited ▾] [A-Z]                                          │
│  ──────────────────────────────────────────────────────────────────     │
│                                                                        │
│  > Session 5 - The Dragon's Lair          edited 2 hours ago           │
│    Session 4 - Ruins of Thornwall         edited 3 days ago            │
│    Questions for DM                       edited 1 week ago            │
│    Shopping List                          edited 1 week ago            │
│    Session 3 - The Goblin Caves           edited 2 weeks ago           │
│    Character Goals                        edited 1 month ago           │
│    Session 2 - Road to Waterdeep          edited 1 month ago           │
│    Session 1 - The Tavern                 edited 2 months ago          │
│                                                                        │
│                                                                        │
│                                                                        │
│                                                                        │
│  Enter: open | a: new | d: delete | r: rename | s: sort | Esc: back  │
└────────────────────────────────────────────────────────────────────────┘
```

**Behavior:**
- Lists all notes for this character
- Default sort: **Last Edited** (most recent first). Press `s` to toggle to **Alphabetical (A-Z)** and back
- Each entry shows: title and relative "edited X ago" timestamp
- `Up/Down` — navigate list
- `Enter` — open selected note in the editor
- `a` — create new note (prompts for title via text input modal, then opens editor)
- `d` — delete selected note (with confirmation)
- `r` — rename selected note (text input modal)
- `s` — toggle sort order (Last Edited / Alphabetical)
- `Esc` — return to previous view (main sheet or character info)

### Mode 2: Document Editor

```
┌──── Session 5 - The Dragon's Lair ─────────────────────────────────────┐
│                                                                        │
│ The party entered the dragon's lair through the collapsed mine         │
│ shaft on the north side of Mount Hotenow.                              │
│                                                                        │
│ Key events:                                                            │
│ - Roper ambush in the entry tunnel (used 2 spell slots)                │
│ - Found the dragon's hoard behind a magically sealed door              │
│ - Dex save vs breath weapon — made it with Shield spell                │
│ - Dragon fled when reduced to ~30% HP                                  │
│                                                                        │
│ Loot:                                                                  │
│ - 1,200 GP                                                             │
│ - Potion of Fire Resistance                                            │
│ - +1 Shield (attuned to Mordin)█                                       │
│                                                                        │
│                                                                        │
│                                                                        │
│  Esc: save & back to list | PgUp/PgDn: scroll | r: rename             │
└────────────────────────────────────────────────────────────────────────┘
```

**Text editing:**
- Full-screen text editor — the entire terminal minus border and footer
- Always in edit mode — cursor placed at end of existing text when opened
- Type to insert at cursor, `Backspace` to delete, `Enter` for newline
- Arrow keys to move cursor (up/down/left/right)
- `Home/End` — jump to start/end of line
- `PgUp/PgDn` — jump one screen of text up/down
- `r` — rename the document (text input modal)
- `Esc` — save and return to the document list
- Auto-scrolls to keep cursor visible

**Quick access flow:** Main sheet → `n` → document list → `Enter` on a note → edit → `Esc` (saves, back to list) → `Esc` (back to main sheet).

### Data Model

```go
// Note represents a single named note document.
type Note struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

Notes stored as `[]Note` on the character, replacing the existing `Personality.Notes string` field. Existing characters with a non-empty `Notes` string are migrated on load to a single Note titled "Notes".

### Files

- New: `internal/ui/views/notes_editor.go`
- New: `internal/ui/views/notes_editor_test.go`
- Modify: `internal/models/character_info.go` — add `Note` struct, change `Personality.Notes` from `string` to `[]Note`
- Modify: `internal/ui/model.go` — add `notesModel` field, handle `OpenNotesMsg`, route view
- Modify: `internal/ui/views/main_sheet.go` — add `n` key binding
- Modify: `internal/ui/views/character_info.go` — add `n` key binding
- Migration: handle old `Notes string` → `[]Note` on character load

---

## Part 4: Character Creation — Personality Step

### Location

New step added after Equipment (StepEquipment) and before Review (StepReview). Becomes Step 7 of 8.

### Layout

```
┌──── Create Character ── Step 7 of 8: Personality ──────────────────────┐
│                                                                        │
│  All fields are optional. You can fill these in later from the         │
│  Character Info view.                                                  │
│                                                                        │
│  Personality Traits:                                                   │
│  > 1. _                                                                │
│    + Add another trait                                                 │
│                                                                        │
│  Ideals:                                                               │
│    1. _                                                                │
│    + Add another ideal                                                 │
│                                                                        │
│  Bonds:                                                                │
│    1. _                                                                │
│    + Add another bond                                                  │
│                                                                        │
│  Flaws:                                                                │
│    1. _                                                                │
│    + Add another flaw                                                  │
│                                                                        │
│  Backstory:                                                            │
│    _                                                                   │
│                                                                        │
│  Tab: next field | Enter: add entry | d: delete entry | Esc: back     │
└────────────────────────────────────────────────────────────────────────┘
```

### Behavior

- **Entirely optional** — every field can be left blank. You can Tab/Enter through everything and advance to Review with no input
- Each field (traits, ideals, bonds, flaws) starts with one empty text input
- `Tab` / `Shift+Tab` — move between fields
- `Enter` on "+ Add another" — adds a new text input to that field
- `d` on an entry — deletes that entry (if there are multiple)
- Backstory is a multi-line text input (Enter inserts newline, Tab advances)
- On advancing to Review, empty entries are stripped — empty strings discarded, empty fields stored as empty slices
- Can return to this step from Review to make changes

### Files

- Modify: `internal/ui/views/character_creation.go` — add `StepPersonality` enum value, insert between `StepEquipment` and `StepReview`, add rendering and key handling

---

## Edge Cases

- **Empty proficiencies**: Omit the category row entirely (don't show "Armor: None")
- **No features in a category**: Show category tab but display "No features" in the list area
- **No notes**: Document list shows empty state with "Press 'a' to create your first note"
- **Long backstory**: Truncated in Character Info view with expand option; full editing via `e` key
- **Migration from old Notes field**: On character load, if `Notes` is a non-empty string (old format), convert to a single `Note` with title "Notes" and migrate the content
- **Terminal height**: Character Info view and Notes Editor should handle small terminals gracefully with scrolling

## Integration Points

### Messages (new)

- `OpenCharacterInfoMsg{}` — main sheet sends when `c` pressed
- `OpenNotesMsg{}` — main sheet or character info sends when `n` pressed
- `BackToCharacterInfoMsg{}` — notes view sends when returning to character info (if opened from there)

### Existing Messages (reused)

- `BackToSheetMsg{}` — character info and notes use this to return to main sheet

### View Routing (model.go)

- `ViewCharacterInfo` — already exists as a constant, needs handler wiring
- `ViewNotes` — new view constant (or reuse an unused slot)
