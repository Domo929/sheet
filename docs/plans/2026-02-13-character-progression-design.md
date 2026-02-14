# Phase 10: Character Progression — Design

## Overview

Add XP tracking, level-up triggers, and a guided level-up wizard to the character sheet TUI. Characters can gain XP from the main sheet, get prompted to level up (including after long rests), and walk through a multi-step wizard that handles HP increases, subclass selection, ASI/feat choices, spell slot updates, and new feature acquisition.

## XP Tracking & Level-Up Triggers

### Main Sheet Integration

- `x` key: Opens "Add XP" numeric input modal (same pattern as HP damage/heal). After adding XP, if `CanLevelUp()` is true, shows status message: "Level up available! Press L to level up."
- `L` key: Opens level-up wizard (`ViewLevelUp`).
  - XP characters: Only works if `CanLevelUp()` is true. Otherwise shows "Not enough XP to level up."
  - Milestone characters: Always available (DM decides when to level up).
  - Level 20: Shows "Already at max level."

### Long Rest Prompt

After completing a long rest, if `CanLevelUp()` is true, show: "You have enough XP to level up! Press L to level up." The rest completes normally; the prompt is non-blocking.

### Multi-Level Catch-Up

If a character gains enough XP for multiple levels, they level up one at a time. After each level-up completes, the status message re-prompts if still eligible.

## Level-Up Wizard

### Architecture

`LevelUpModel` in `internal/ui/views/level_up.go`, following the same step-based pattern as `CharacterCreationModel`. Changes are staged during the wizard and only applied on confirmation. Cancel at any point reverts everything.

### Dynamic Step Flow

Steps are determined at wizard initialization based on the new level and class:

1. **HP Increase** (always)
2. **Subclass Selection** (if at subclass-granting level and no subclass chosen yet)
3. **ASI or Feat** (if at ASI level: typically 4, 8, 12, 16, 19)
4. **New Features Summary** (always, if new features exist at this level)
5. **Spell Slot Update** (if spellcaster and slots changed)
6. **Confirmation** (always)

Navigation: Enter to advance, Esc to go back (or cancel from step 1). Up/Down for selections within a step.

### Step 1: HP Increase

- Display: Current max HP, hit die type (from class), CON modifier
- Two options: "Roll" (simulated `math/rand` roll, minimum 1) or "Take Average" (die/2 + 1)
- HP gained = roll_or_average + CON modifier (minimum 1 total)
- User can re-roll or switch methods before confirming
- Shows new max HP preview

### Step 2: Subclass Selection (conditional)

- Lists available subclasses from `classes.json` with descriptions
- Right panel shows subclass description and features gained at current level
- On selection, subclass name stored in `CharacterInfo.Subclass`
- Subclass features for current level added to `Features.ClassFeatures`
- Subclass level varies by class (data-driven from `classes.json` feature names)

### Step 3: ASI or Feat (conditional)

**ASI mode:**
- Choose: +2 to one ability OR +1 to two abilities
- Display current scores with preview of changes
- Abilities at 20 are greyed out (can't exceed 20)
- For +1/+1 mode: select two different abilities

**Feat mode:**
- Searchable scrollable list from `feats.json`
- Filtered by prerequisites (ineligible feats greyed out)
- Right panel shows full feat description
- If feat grants an ASI (most 2024 feats do), sub-step asks which ability to boost
- Selecting a feat adds it to `Features.Feats` and applies mechanical effects

### Step 4: New Features Summary

- Lists all class features gained at this level from `classes.json`
- Lists subclass features if applicable
- Each feature shows name and full description
- Scrollable for levels with many features
- Read-only review step

### Step 5: Spell Slot Update (conditional)

- Automatically calculated from `classes.json` spell slot progression
- Shows before/after comparison of spell slots per level
- Notes when new spell levels are unlocked
- Informational only — auto-applied, no user input
- Skipped if no spell slot changes at this level

### Step 6: Confirmation

Summary of all changes:
- Level: X → X+1
- HP: old max → new max (+N)
- Subclass: (if selected)
- ASI/Feat: (if taken)
- Spell slots: (if changed)
- New features: (list names)

Enter to confirm and apply. Esc to cancel entire level-up.

## Data Requirements

### New: `data/feats.json`

```json
{
  "feats": [
    {
      "name": "Alert",
      "category": "General",
      "prerequisite": "",
      "description": "...",
      "effects": {
        "abilityScoreIncrease": {
          "options": ["Dexterity", "Charisma", "Wisdom"],
          "amount": 1
        },
        "initiativeBonus": 5
      }
    }
  ]
}
```

Full 2024 5e SRD feats. Created as parallel data branch (`data/feats`).

### New Data Types

```go
type Feat struct {
    Name         string     `json:"name"`
    Category     string     `json:"category"`
    Prerequisite string     `json:"prerequisite,omitempty"`
    Description  string     `json:"description"`
    Effects      FeatEffect `json:"effects"`
}

type FeatEffect struct {
    AbilityScoreIncrease *FeatASI `json:"abilityScoreIncrease,omitempty"`
    InitiativeBonus      int      `json:"initiativeBonus,omitempty"`
    SpeedBonus           int      `json:"speedBonus,omitempty"`
    HPPerLevel           int      `json:"hpPerLevel,omitempty"`
}

type FeatASI struct {
    Options []string `json:"options"`
    Amount  int      `json:"amount"`
}

type FeatData struct {
    Feats []Feat `json:"feats"`
}
```

### Existing Data Used

- `classes.json`: features by level, subclasses, spell slots, hit dice
- Already loaded by `data.Loader`

## Integration Points

### `internal/ui/model.go`

- Add `levelUpModel *views.LevelUpModel` field
- Handle `OpenLevelUpMsg` → create and navigate to `ViewLevelUp`
- Handle `LevelUpCompleteMsg` → return to main sheet
- Handle `BackToSheetMsg` from level-up → return to main sheet (cancel)

### `internal/ui/views/main_sheet.go`

- `x` key: XP input modal (numeric input, same pattern as HP)
- `L` key: Send `OpenLevelUpMsg` if eligible
- Long rest completion: Check `CanLevelUp()` and show prompt
- XP input mode enum and handler methods

### `internal/ui/messages.go`

- `OpenLevelUpMsg{}`
- `LevelUpCompleteMsg{}`

### New Files

- `internal/ui/views/level_up.go` — Level-up wizard model
- `internal/ui/views/level_up_test.go` — Tests
- `internal/data/feats.go` — Feat types (or added to types.go and loader.go)
- `data/feats.json` — Feat database

## Edge Cases

- **Level 20:** Block level-up with message
- **Multiple levels at once:** One at a time, re-prompt after each
- **Subclass already selected:** Skip subclass step; still show subclass features for current level
- **Non-spellcaster:** Skip spell slot step
- **Feat prerequisites not met:** Grey out with explanation
- **Ability score at 20:** Prevent ASI from exceeding cap
- **Cancel mid-wizard:** No changes applied
- **Milestone + long rest:** No automatic level-up prompt (no XP to check); user presses L manually
