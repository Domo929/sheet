# Spell Casting Confirmation Modal & Enhanced Display

**Date:** 2026-02-13
**Status:** Approved
**Type:** Feature Enhancement

## Overview

Add a casting confirmation modal that appears before any spell is cast, displaying complete spell information including damage, saving throws, and spell save DC. Also enhance the spell details panel and make header statistics more prominent.

## Problem Statement

Currently:
- Spells cast immediately when pressing 'c' (for cantrips and single-slot spells)
- No opportunity to review spell details before casting
- Damage and saving throw information only visible in description text
- Spell save DC not displayed anywhere in spell details
- Header stats (ability, DC, attack bonus) are in a faded color and hard to read

Users need to see complete spell information (damage, save type, DC, description) before committing to casting, especially for resource management and tactical decision-making.

## Goals

1. Always show a confirmation modal before casting any spell
2. Display damage/healing and saving throw information prominently
3. Calculate and show spell save DC
4. Make character spellcasting stats in header more visible
5. Maintain consistency with existing UI patterns
6. Work without changes to spell data structure

## Design

### 1. Enhanced Spell Details Panel

Add damage and saving throw information to the right-side spell details panel.

**Current display:**
```
Burning Hands
Level 1 Evocation

Casting Time: 1 action
Range: Self (15-foot cone)
Components: V, S
Duration: Instantaneous

[Description...]
```

**New display:**
```
Burning Hands
Level 1 Evocation

Casting Time: 1 action
Range: Self (15-foot cone)
Components: V, S
Duration: Instantaneous
Damage: 3d6 fire
Saving Throw: Dexterity DC 15

[Description...]

At Higher Levels:
+1d6 per slot level
```

**Rules:**
- Show "Damage: [dice] [type]" line only if `spell.Damage` field exists
- Show "Saving Throw: [ability] DC [calculated]" line only if `spell.SavingThrow` field exists
- Calculate DC using: `8 + spellcasting ability modifier + proficiency bonus`
- Use existing `CalculateSpellSaveDC()` function from `models/spellcasting.go`
- Place these lines after Duration and before Description

### 2. Prominent Header Statistics

Make the top status bar spellcasting statistics bold and high-contrast.

**Current:** Faded gray text displaying ability, DC, attack bonus, prepared count

**New:** Bold, bright text for easy reading:
```
Spellcasting: Intelligence | DC 15 | Attack +7 | Prepared: 9/9
```

**Implementation:**
- Use `lipgloss.NewStyle().Bold(true)` for the stats
- Increase foreground color brightness
- Keep same layout, just enhance visibility

### 3. Casting Confirmation Modal

Always show a confirmation modal when user presses 'c' to cast a spell.

**Modal layout:**

```
┌─ Cast [Spell Name] ──────────────────────────────┐
│ Level [X] [School] [ritual indicator]            │
│                                                   │
│ Casting Time: [time]                             │
│ Range: [range]                                   │
│ Components: [components]                         │
│ Duration: [duration]                             │
│ Damage: [dice] [type]           (if present)     │
│ Saving Throw: [ability] DC [X]  (if present)     │
│                                                   │
│ [Word-wrapped description]                       │
│                                                   │
│ At Higher Levels:                (if applicable) │
│ [Word-wrapped upcast text]                       │
│                                                   │
│ ─────────────────────────────────                │
│                                                   │
│ Select Spell Slot Level:         (if not cantrip)│
│ > Level 1 Slot (3d6 damage) [2/4 remaining]      │
│   Level 2 Slot (4d6 damage) [3/3 remaining]      │
│                                                   │
│ ↑↓: select slot | Enter: cast | Esc: cancel      │
└───────────────────────────────────────────────────┘
```

**Content sections:**

1. **Header:** Spell name with level and school
2. **Basic info:** Casting time, range, components, duration
3. **Combat info:** Damage and saving throw (if applicable)
4. **Description:** Full spell description text
5. **Upcast info:** "At Higher Levels" section (if spell can be upcast)
6. **Separator line**
7. **Slot selection:** Available slot levels with calculated effects (if not cantrip)
8. **Help text:** Keyboard shortcuts

### 4. Modal Behavior by Spell Type

**Cantrips:**
- Show modal with spell information
- No slot selection section
- Help text: "Press Enter to cast | Esc: cancel"
- Enter immediately casts (no resource consumed)

**Spells with 1 available slot level:**
- Show modal with spell information
- Show single slot option (pre-selected, no cursor needed)
- Help text: "Enter: cast with Level X slot | Esc: cancel"
- Enter casts and consumes the slot

**Spells with multiple slot levels:**
- Show modal with spell information
- Show all available slot levels
- Arrow keys navigate between levels
- Selected level highlighted with cursor (">") and color
- Each level shows calculated upcast effect
- Help text: "↑↓: select | Enter: cast | Esc: cancel"
- Enter casts with selected slot level

### 5. User Flow

```
User presses 'c' on selected spell
  ↓
Check if spell can be cast
  ├─ Not prepared/no slots → Show error message
  └─ Can cast → Continue
      ↓
Enter ModeConfirmCast
  ↓
Render confirmation modal
  ↓
User interaction:
  ├─ Press Esc → Cancel, return to spell list
  ├─ Press ↑/↓ → Navigate slot levels (if multiple)
  └─ Press Enter → Cast spell
      ↓
      Consume slot (if applicable)
      ↓
      Show status message
      ↓
      Return to spell list
```

### 6. Modal Display Logic

**New mode constant:** `ModeConfirmCast`

**State variables:**
- `castingSpell *models.KnownSpell` - The spell being cast
- `availableCastLevels []int` - Available slot levels for this spell
- `castLevelCursor int` - Selected slot level index
- `selectedSpellData *data.SpellData` - Full spell data (already exists)

**Key handling in ModeConfirmCast:**
- **Esc:** Cancel casting, return to `ModeSpellList`
- **Up:** Decrement `castLevelCursor` (if multiple levels)
- **Down:** Increment `castLevelCursor` (if multiple levels)
- **Enter:** Confirm cast at selected level

**Status messages after casting:**
- Leveled spells: "Cast [Spell Name] using Level [X] slot"
- Cantrips: "Cast [Spell Name] (no slot required)"
- Upcast spells: "Cast [Spell Name] at Level [X] ([calculated effect])"

### 7. Implementation Components

**Files to modify:**

1. **`internal/ui/views/spellbook.go`**
   - Add `ModeConfirmCast` constant to `SpellbookMode` enum
   - Modify `handleCastSpell()` to enter modal mode instead of casting directly
   - Create `renderCastConfirmationModal()` function
   - Update `renderSpellDetails()` to include Damage and Saving Throw lines
   - Update header rendering function to make stats bold
   - Add key event handlers for modal navigation
   - Reuse existing `calculateUpcastEffect()` for slot level display
   - Use existing `getAvailableCastLevels()` for slot options

2. **Reuse existing functions:**
   - `CalculateSpellSaveDC(abilityMod, proficiencyBonus)` from `models/spellcasting.go`
   - `castSpellAtLevel(spell, level)` for actual casting
   - `wordWrap(text, width)` for description formatting
   - `calculateUpcastEffect(level)` for upcast display

**No changes needed:**
- `internal/data/types.go` - SpellData structure has all necessary fields
- `data/spells.json` - existing data is sufficient
- `internal/models/spellcasting.go` - existing functions work as-is

### 8. UI Polish Details

**Modal styling:**
- Width: 60-70 characters (readable without being too wide)
- Border: Rounded border (`lipgloss.RoundedBorder()`)
- Padding: 1-2 spaces inside border
- Centered on screen

**Spell information formatting:**
- Bold spell name at top
- Consistent "Label: value" format for all stats
- Empty lines between major sections for readability
- Word-wrap description to modal width

**Slot selection styling:**
- Cursor indicator: `>` prefix for selected slot
- Highlight selected slot with bright color (`lipgloss.Color("12")`)
- Show remaining/total slots for each level
- Display calculated upcast effect inline with each level

**Edge cases:**
- Very long descriptions: Word-wrap within modal, don't scroll
- Spells with ritual tag: Show "(ritual)" after school
- Spells without damage/save: Simply omit those lines
- Cantrips in modal: Don't show slot selection section

## Technical Considerations

### Spell Save DC Calculation

The spell save DC depends on the character's spellcasting ability and proficiency bonus:

```
DC = 8 + ability modifier + proficiency bonus
```

The character already has:
- `Spellcasting.Ability` (e.g., "Intelligence")
- `Info.ProficiencyBonus` (calculated from level)
- Ability score in `AbilityScores` map

Implementation:
```go
abilityMod := character.GetAbilityModifier(spellcasting.Ability)
profBonus := character.Info.ProficiencyBonus
dc := models.CalculateSpellSaveDC(abilityMod, profBonus)
```

### Modal Rendering

Reuse the pattern from existing `renderCastLevelOverlay()` but expand to include full spell details.

Structure:
1. Build lines array with all spell information
2. Format slot selection (if applicable)
3. Join with newlines
4. Wrap in styled box with border
5. Return as string for overlay rendering

### Consistent Display Logic

The spell details should be shown identically in:
1. Right-side spell details panel (always visible)
2. Casting confirmation modal (when casting)

Consider extracting common formatting into a helper function like `formatSpellInfo(spell, showFull bool)` to avoid duplication.

## Success Criteria

1. ✅ Pressing 'c' always shows confirmation modal
2. ✅ Modal displays complete spell information
3. ✅ Damage and saving throw visible in spell details panel
4. ✅ Spell save DC correctly calculated and displayed
5. ✅ Header stats are bold and easy to read
6. ✅ Cantrips, single-slot, and multi-slot spells all work correctly
7. ✅ Esc cancels casting, Enter confirms
8. ✅ Arrow keys navigate slot levels (when multiple available)
9. ✅ Calculated upcast effects shown for each slot level
10. ✅ Status message confirms spell was cast

## Future Enhancements (Not in MVP)

- Spell categorization/filtering by type
- Concentration indicator if spell requires concentration
- Range/area visualization
- Quick reference for spell components (what V, S, M mean)
- Spell favorite/bookmark system
- Recently cast spells history

## Notes

- Design avoids complex spell categorization (attack vs save vs utility)
- Simply displays all available information for any spell type
- Keeps implementation simple and maintainable
- No data structure changes required
- Follows existing UI patterns and conventions
- Works as MVP to gather user feedback
