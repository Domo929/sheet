# Phase 12: Dice Rolling Integration — Design

## Overview

Integrate the `github.com/Domo929/roll` library into the character sheet TUI to provide animated dice rolling for weapon attacks, spell casting, skill checks, saving throws, hit dice, and custom rolls. Rolls display through animated modal overlays with a tumbling dice effect and color flashing. A toggleable roll history column on the right side tracks recent rolls across active play views.

---

## Roll Engine Core

### Architecture

A centralized `RollEngine` component lives in `internal/ui/components/roll_engine.go` and is owned by the app model. All views request rolls by sending a `RequestRollMsg` — the engine handles animation, display, and history. Views never call the roll library directly.

### States

1. **Idle** — no roll in progress
2. **AdvPrompt** — showing Normal/Advantage/Disadvantage picker (d20 rolls only)
3. **Animating** — tumbling dice with color flashing (~12 frames, decelerating)
4. **Showing** — final result displayed in modal overlay, waiting for dismissal

### Message Flow

```
View sends RequestRollMsg{Label, DiceExpr, Modifier, RollType, AdvPrompt, FollowUp}
  → If AdvPrompt: show [N]ormal / [A]dvantage / [D]isadvantage prompt → user picks
  → Animation starts (tick-based, ~1 second)
  → Animation completes → modal shows final result
  → If FollowUp (e.g., damage after attack):
      "Press Enter to roll damage, Esc to skip"
      → Enter: triggers follow-up RequestRollMsg → animates → shows result
      → Esc: skip, dismiss
  → Any key → RollCompleteMsg{Result, RollType} sent back to view
  → Roll added to history
```

### Roll Types

- `RollAttack` — d20 + attack bonus (weapon or spell attack)
- `RollDamage` — damage dice + damage modifier
- `RollSkillCheck` — d20 + skill modifier
- `RollSavingThrow` — d20 + save modifier
- `RollHitDice` — hit die + CON modifier (short rest healing)
- `RollLuck` — d20, no modifier, no advantage prompt
- `RollCustom` — arbitrary dice from the custom roller

### Animation Style — Tumbling Dice with Color Flash

- ~12 frames total, decelerating (80ms → 260ms between frames)
- All dice show simultaneously with rapidly changing random values
- Colors cycle through a bright palette each tick (magenta → cyan → yellow → white, cycling)
- Dice "land" one by one from left to right — landed dice settle to green with a rounded border
- Final frame: all dice green/landed, breakdown line shown

### Modal Overlay Rendering

Centered lipgloss box on top of the current view:

```
┌──── ⚔ Longsword Attack ────────────┐
│                                      │
│       ╭───╮  ╭───╮                  │
│       │ 14│  │  7│                  │
│       ╰───╯  ╰───╯                  │
│                                      │
│   (14) + 7 = 21                     │
│   Advantage: kept 14, dropped 7     │
│                                      │
│  Enter: roll damage • Esc: skip     │
└──────────────────────────────────────┘
```

- Nat 20: green/bold highlight on the die and total
- Nat 1: red/bold highlight on the die and total
- Luck rolls: purple/magenta coloring throughout

### Advantage/Disadvantage Prompt

For d20 rolls (attacks, skills, saves), a small prompt appears before rolling:

```
┌─ Roll Mode ──────────────────┐
│  [N]ormal                    │
│  [A]dvantage                 │
│  [D]isadvantage              │
└──────────────────────────────┘
```

Press N, A, or D to immediately trigger the roll. Esc cancels.

---

## Roll History

### Data Model

```go
type RollHistoryEntry struct {
    Label      string    // e.g., "Longsword Attack", "Perception Check", "Luck"
    RollType   RollType
    Expression string    // e.g., "1d20+7"
    Rolls      []int     // individual dice
    Modifier   int
    Total      int
    Advantage  bool      // was rolled with advantage
    Disadvantage bool    // was rolled with disadvantage
    NatCrit    bool      // natural 20 on d20
    NatFail    bool      // natural 1 on d20
    Timestamp  time.Time
}
```

### History Column (Right Side)

- **Hidden by default.** Appears after the first roll in a session.
- **Toggle:** Press `h` to show/hide (available on main sheet, spellbook, combat views).
- **Width:** ~25 columns. The active view's content shrinks to accommodate when visible.
- **Capacity:** Last 50 rolls, most recent at top.
- **Session-only:** Not persisted to character JSON. Cleared when leaving the character.

### Display Format

```
┌── Roll History ──────────┐
│ ⚔ Longsword Attack       │
│   1d20+7 → 19 (nat 12)  │
│                          │
│ 🎯 Perception Check      │
│   1d20+5 → 18 (ADV)     │
│                          │
│ 🎲 Luck                  │
│   1d20 → 14             │
│                          │
│ 💥 Longsword Damage      │
│   1d8+4 → 9             │
└──────────────────────────┘
```

- Nat 20s highlighted in green/bold
- Nat 1s highlighted in red/bold
- Luck rolls shown in purple/magenta
- Advantage/Disadvantage noted with (ADV)/(DIS) suffix

---

## Integration Points

### Weapon Attacks (Main Sheet — Actions Panel)

- **Trigger:** Enter on a weapon in the Actions panel
- **Flow:** Advantage prompt → d20 + `getWeaponAttackBonus()` → result modal → "Enter to roll damage, Esc to skip" → damage dice (`weapon.Damage`) + `getWeaponDamageMod()` → result modal → any key dismisses
- **Data available:** `weapon.Damage` (string, e.g., "1d8"), `getWeaponAttackBonus()`, `getWeaponDamageMod()`, `weapon.DamageType`

### Spell Casting (Main Sheet + Spellbook)

- **Trigger:** Existing spell cast confirmation flow
- **Flow:** After slot consumption:
  - If spell requires spell attack: advantage prompt → d20 + `GetSpellAttackBonus()` → result → Enter for damage / Esc to skip
  - If spell has damage (save-based): roll damage dice directly (no attack roll — DC shown in result)
  - Upcast damage scaling uses existing `calculateUpcastEffect()` logic
- **Data available:** `spell.Damage`, `GetSpellAttackBonus()`, `getSpellSaveDC()`

### Skill Checks (Main Sheet — Skills Panel)

- **Trigger:** Enter on a focused skill
- **Flow:** Advantage prompt → d20 + `GetSkillModifier(skillName)` → result modal → any key dismisses
- **Label:** Skill name + "Check" (e.g., "Perception Check")

### Saving Throws (Main Sheet — Abilities/Saves Panel)

- **Trigger:** Enter on a focused saving throw
- **Flow:** Advantage prompt → d20 + `GetSavingThrowModifier(ability)` → result modal → any key dismisses
- **Label:** Ability name + "Saving Throw" (e.g., "DEX Saving Throw")

### Luck (Main Sheet — Skills Panel + Dedicated Key)

- **Location:** Displayed at the top of the Skills panel, visually separated from real skills. Rendered in purple/magenta with a 🎲 icon.
- **Trigger:** Enter on Luck in the Skills panel, or press `` ` `` (backtick) from the main sheet
- **Flow:** d20 straight roll — no modifier, no advantage prompt
- **Roll type:** `RollLuck` — uses purple/magenta coloring in modal and history
- **Not a real skill** — doesn't exist in the character model. Purely a UI convenience.

### Hit Dice (Short Rest)

- **Trigger:** During short rest hit dice spending
- **Flow:** Prompt "Roll or Take Average?" before spending each hit die
  - Roll: `1d{hitDieType}` + CON modifier via roll library with animation
  - Average: keeps current behavior (`dieType/2 + 1 + conMod`)
- **Roll type:** `RollHitDice`

### Custom Dice Roller (Overlay)

- **Trigger:** Press `/` from any active play view (main sheet, spellbook, combat)
- **Overlay:** Centered modal showing 7 die types in a row:
  ```
  ┌──── Custom Roll ──────────────────────┐
  │                                        │
  │  [d4]  [d6]  [d8] [d10] [d12] [d20] [d100]  │
  │                                        │
  │  Quantity: 3        ← / →  to change  │
  │                                        │
  │  Enter: roll • Esc: cancel            │
  └────────────────────────────────────────┘
  ```
- **Navigation:** Up/Down (or Left/Right) to select die type. Left/Right (or Up/Down on quantity row) to change quantity (1–100). Enter to roll, Esc to cancel.
- **Result:** Goes through standard animation → modal → history pipeline with `RollCustom` type
- **Label:** "Custom Roll" with expression (e.g., "3d8")

---

## Views with Roll History

Roll history column appears on **active play views only:**
- Main Sheet
- Spellbook
- Combat (when implemented)

Not shown on: Inventory, Character Info, Notes, Character Creation, Character Selection, Level Up.

---

## New Dependency

- `github.com/Domo929/roll` — dice parsing and rolling library (local at `/home/dcupo/Software/roll`)

## Files

### New Files

- `internal/ui/components/roll_engine.go` — RollEngine component (states, animation ticks, modal rendering, adv/disadv prompt, custom roll overlay)
- `internal/ui/components/roll_engine_test.go` — Tests
- `internal/ui/components/roll_history.go` — RollHistory data structure, column rendering
- `internal/ui/components/roll_history_test.go` — Tests

### Modified Files

- `internal/ui/model.go` — Add `rollEngine` and `rollHistory` fields, route `RequestRollMsg`/`RollCompleteMsg`, render overlay on top of current view, render history column alongside active play views, handle `h` toggle
- `internal/ui/views/main_sheet.go` — Add Enter handlers for skills/saves, update weapon attack flow to send `RequestRollMsg`, add Luck to skills panel rendering, add `/` and `` ` `` key bindings, add custom roll overlay state, update short rest to prompt roll vs average
- `internal/ui/views/spellbook.go` — Update spell casting to send `RequestRollMsg` for damage/attack rolls

## Edge Cases

- **Spell with no damage field:** No roll triggered, just "Cast {spell name}" status message as before
- **Weapon with no damage string:** Fall back to "1d4" (unarmed strike default)
- **Custom roll 100d100:** Roll library supports up to 10,000 dice, but animation only shows first ~10 dice visually with a "... and 90 more" note. Total still calculated from all dice.
- **Small terminal:** Modal overlay and history column gracefully degrade. History column hidden if terminal width < 80. Modal width capped to terminal width - 4.
- **Roll during animation:** Ignored — engine is locked during animation state
- **Nat 20 / Nat 1 detection:** Only on d20 rolls — check if any kept die shows 20 or 1
