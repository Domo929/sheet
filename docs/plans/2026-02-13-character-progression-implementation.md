# Character Progression Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add XP tracking, level-up triggers (including long rest prompts), and a guided multi-step level-up wizard with HP increase, subclass selection, ASI/feat choices, spell slot updates, and new feature summary.

**Architecture:** Multi-step wizard view (`LevelUpModel`) following the existing `CharacterCreationModel` step-based pattern. Changes are staged during the wizard and only applied on confirmation. XP input and level-up triggers are added to the main sheet. A new `feats.json` data file provides feat data for ASI/feat selection.

**Tech Stack:** Go, Bubble Tea (TUI framework), lipgloss (styling), bubbles/key (key bindings), math/rand (dice rolling)

**Existing Patterns to Follow:**
- Step-based wizard: see `internal/ui/views/character_creation.go` (CreationStep enum, currentStep field)
- Numeric input modal: see `main_sheet.go` HPInputMode / handleHPInput pattern
- View navigation: see `internal/ui/model.go` (OpenInventoryMsg → create model → switch view)
- Data loading: see `internal/data/loader.go` (GetX/loadXUnsafe pattern with mutex)
- Key bindings: see `main_sheet.go` mainSheetKeyMap struct

**Git Workflow:** All work on branch `feature/character-progression`. Commits should be logically grouped. PR against `main` when complete.

---

### Task 1: Feat Data Types and Loader

**Files:**
- Modify: `internal/data/types.go` — add Feat types
- Modify: `internal/data/loader.go` — add GetFeats, loadFeatsUnsafe, FindFeatByName
- Modify: `internal/data/loader_test.go` — add feat loader test
- Create: `data/feats.json` — feat database (parallel data work, large file)

**Step 1: Add feat types to types.go**

Add to `internal/data/types.go` after the existing `ConditionData` struct:

```go
// Feat represents a feat available for character selection.
type Feat struct {
	Name         string     `json:"name"`
	Category     string     `json:"category"`              // "General", "Fighting", "Magic", etc.
	Prerequisite string     `json:"prerequisite,omitempty"` // Human-readable prerequisite text
	Repeatable   bool       `json:"repeatable,omitempty"`   // Whether the feat can be taken multiple times
	Description  string     `json:"description"`
	Effects      FeatEffect `json:"effects"`
}

// FeatEffect represents the mechanical effects of a feat.
type FeatEffect struct {
	AbilityScoreIncrease *FeatASI `json:"abilityScoreIncrease,omitempty"` // +1 to a choice of abilities
	InitiativeBonus      int      `json:"initiativeBonus,omitempty"`
	SpeedBonus           int      `json:"speedBonus,omitempty"`
	HPPerLevel           int      `json:"hpPerLevel,omitempty"` // Tough feat: +2 HP per level
	ACBonus              int      `json:"acBonus,omitempty"`
}

// FeatASI represents an ability score increase granted by a feat.
type FeatASI struct {
	Options []string `json:"options"` // Which abilities can be increased
	Amount  int      `json:"amount"`  // How much to increase by (usually 1)
}

// FeatData contains all feat data.
type FeatData struct {
	Feats []Feat `json:"feats"`
}
```

**Step 2: Add feat loader methods to loader.go**

Add `feats *FeatData` field to the `Loader` struct. Add `GetFeats()`, `loadFeatsUnsafe()`, and `FindFeatByName()` methods following the exact same pattern as `GetConditions()`/`loadConditionsUnsafe()`/`FindConditionByName()`. Add `l.feats = nil` to `ClearCache()`. Add the `loadFeatsUnsafe` call to `LoadAll()`.

**Step 3: Create data/feats.json**

Create `data/feats.json` with 2024 5e SRD feats. Include at minimum:
- **Origin feats** (level 1): Alert, Crafter, Healer, Lucky, Magic Initiate, Musician, Savage Attacker, Skilled, Tavern Brawler, Tough
- **General feats** (level 4+): Actor, Athlete, Charger, Chef, Crossbow Expert, Crusher, Defensive Duelist, Dual Wielder, Durable, Elemental Adept, Fey Touched, Fighting Initiate, Great Weapon Master, Heavily Armored, Heavy Armor Master, Inspiring Leader, Keen Mind, Lightly Armored, Linguist, Mage Slayer, Medium Armor Master, Mobile, Moderately Armored, Mounted Combatant, Observant, Piercer, Polearm Master, Poisoner, Resilient, Ritual Caster, Sentinel, Shadow Touched, Sharpshooter, Shield Master, Skill Expert, Skulker, Slasher, Speedy, Spell Sniper, Telekinetic, Telepathic, War Caster, Weapon Master

Each feat should have: name, category, prerequisite (string), description, and effects (where mechanically modelable).

**Step 4: Write test for feat loader**

Add to `internal/data/loader_test.go`:

```go
func TestGetFeats(t *testing.T) {
	loader := NewLoader("../../data")
	feats, err := loader.GetFeats()
	require.NoError(t, err)
	assert.NotNil(t, feats)
	assert.Greater(t, len(feats.Feats), 0, "Should have at least one feat")

	// Verify a known feat
	feat, err := loader.FindFeatByName("Alert")
	require.NoError(t, err)
	assert.Equal(t, "Alert", feat.Name)
	assert.NotEmpty(t, feat.Description)
}
```

**Step 5: Run tests**

Run: `go test ./internal/data/ -v`
Expected: All tests pass including the new feat loader test.

**Step 6: Commit**

```bash
git add internal/data/types.go internal/data/loader.go internal/data/loader_test.go data/feats.json
git commit -m "feat: add feat data types, loader, and feats.json database"
```

---

### Task 2: XP Input on Main Sheet

**Files:**
- Modify: `internal/ui/views/main_sheet.go` — add XP input mode, key bindings, handler

**Step 1: Add XP input mode and key bindings**

Add to the `HPInputMode` enum (or create a new enum):

```go
const (
	HPInputNone HPInputMode = iota
	HPInputDamage
	HPInputHeal
	HPInputTemp
	HPInputXP // Add this
)
```

Add to `mainSheetKeyMap` struct:

```go
AddXP    key.Binding
LevelUp  key.Binding
```

Add to `defaultMainSheetKeyMap()`:

```go
AddXP: key.NewBinding(
	key.WithKeys("x"),
	key.WithHelp("x", "add XP"),
),
LevelUp: key.NewBinding(
	key.WithKeys("L"),
	key.WithHelp("L", "level up"),
),
```

**Step 2: Add key handlers in Update()**

In the `switch` block after the existing `Rest` key case (around line 396), add:

```go
case key.Matches(msg, m.keys.AddXP):
	if m.character.Info.ProgressionType == models.ProgressionXP {
		m.hpInputMode = HPInputXP
		m.hpInputBuffer = ""
		m.statusMessage = "Add XP: _"
	} else {
		m.statusMessage = "Milestone progression — no XP tracking"
	}
	return m, nil
case key.Matches(msg, m.keys.LevelUp):
	if m.character.Info.Level >= 20 {
		m.statusMessage = "Already at max level (20)"
		return m, nil
	}
	if m.character.Info.ProgressionType == models.ProgressionXP && !m.character.Info.CanLevelUp() {
		m.statusMessage = fmt.Sprintf("Not enough XP to level up (need %d)", models.XPForNextLevel(m.character.Info.Level))
		return m, nil
	}
	return m, func() tea.Msg { return OpenLevelUpMsg{} }
```

**Step 3: Handle XP input in handleHPInput()**

Add a case for `HPInputXP` in `handleHPInput()`. When Enter is pressed with a valid number:

```go
case HPInputXP:
	xp, err := strconv.Atoi(m.hpInputBuffer)
	if err != nil || xp <= 0 {
		m.statusMessage = "Invalid XP amount"
		m.hpInputMode = HPInputNone
		m.hpInputBuffer = ""
		return m, nil
	}
	m.character.Info.AddXP(xp)
	m.saveCharacter()
	m.hpInputMode = HPInputNone
	m.hpInputBuffer = ""
	if m.character.Info.CanLevelUp() {
		m.statusMessage = fmt.Sprintf("Gained %d XP! Level up available — press L to level up", xp)
	} else {
		nextLevelXP := models.XPForNextLevel(m.character.Info.Level)
		m.statusMessage = fmt.Sprintf("Gained %d XP (total: %d / %d)", xp, m.character.Info.ExperiencePoints, nextLevelXP)
	}
	return m, nil
```

Also update the XP input display in the `handleHPInput` Enter/digit/backspace handling to show "Add XP: " prefix for `HPInputXP` mode (follow the same pattern as damage/heal).

**Step 4: Add long rest level-up prompt**

In `performLongRest()`, after the result string is built (around line 900), add:

```go
// Prompt level-up if XP threshold reached
if m.character.Info.CanLevelUp() {
	result.WriteString("\n\n★ You have enough XP to level up! Press L to level up")
}
```

**Step 5: Add message types**

In `internal/ui/views/main_sheet.go` (where `BackToSelectionMsg` etc. are defined), add:

```go
// OpenLevelUpMsg is sent when the user wants to level up.
type OpenLevelUpMsg struct{}

// LevelUpCompleteMsg is sent when a level-up is complete.
type LevelUpCompleteMsg struct{}
```

**Step 6: Update help footer**

In the help footer rendering, add `x: add XP` and `L: level up` to the displayed key hints.

**Step 7: Run tests**

Run: `go test ./internal/ui/views/ -v`
Expected: All existing tests pass. No new tests needed for key binding wiring (tested via integration later).

**Step 8: Commit**

```bash
git add internal/ui/views/main_sheet.go
git commit -m "feat(main-sheet): add XP input modal and level-up key bindings"
```

---

### Task 3: Level-Up Wizard — Core Structure and HP Step

**Files:**
- Create: `internal/ui/views/level_up.go`
- Create: `internal/ui/views/level_up_test.go`

**Step 1: Create the level-up model with step state machine**

Create `internal/ui/views/level_up.go` with:

```go
package views

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LevelUpStep represents the current step in the level-up wizard.
type LevelUpStep int

const (
	LevelUpStepHP LevelUpStep = iota
	LevelUpStepSubclass
	LevelUpStepASI
	LevelUpStepFeatures
	LevelUpStepSpellSlots
	LevelUpStepConfirm
)

// HPMethod represents how HP increase is determined.
type HPMethod int

const (
	HPMethodRoll    HPMethod = iota
	HPMethodAverage
)

// ASIMode represents whether the user is taking an ASI or a feat.
type ASIMode int

const (
	ASIModeASI  ASIMode = iota
	ASIModeFeat
)

// ASIPattern represents the ASI allocation pattern.
type ASIPattern int

const (
	ASIPatternPlus2 ASIPattern = iota // +2 to one ability
	ASIPatternPlus1 ASIPattern = 1    // +1 to two abilities
)

// LevelUpModel manages the level-up wizard.
type LevelUpModel struct {
	character *models.Character
	storage   *storage.CharacterStorage
	loader    *data.Loader
	width     int
	height    int
	keys      levelUpKeyMap

	// Class data (loaded on init)
	classData *data.Class
	featData  *data.FeatData

	// Step management
	currentStep LevelUpStep
	steps       []LevelUpStep // Ordered list of applicable steps
	stepIndex   int           // Index into steps slice

	// Level being gained
	newLevel int // The level the character will become

	// Staged changes (applied only on confirmation)
	stagedHPIncrease    int
	stagedSubclass      string
	stagedSubclassFeats []models.Feature
	stagedASIChanges    [6]int // Delta per ability (+1 or +2)
	stagedFeat          *data.Feat
	stagedFeatASIAbility int // Index of ability chosen for feat's ASI (-1 = none)
	stagedNewFeatures   []models.Feature
	stagedSpellSlots    map[int]int // level -> new slot count

	// HP step state
	hpMethod     HPMethod
	hpRollResult int
	hpCursorPos  int // 0=Roll, 1=Average

	// Subclass step state
	subclassCursor int
	subclassScroll int

	// ASI step state
	asiMode         ASIMode
	asiPattern      ASIPattern
	asiCursor       int    // Cursor for ability selection
	asiSelected     [6]bool // Which abilities have been selected
	asiModeCursor   int    // 0=ASI, 1=Feat

	// Feat step state
	featCursor     int
	featScroll     int
	featSearchTerm string
	filteredFeats  []data.Feat

	// Features step state
	featuresCursor int
	featuresScroll int

	// Status
	statusMessage string
	confirmingQuit bool
	rng           *rand.Rand
}
```

Include key map, constructor (`NewLevelUpModel`), `Init()`, `Update()`, and `View()` methods. The constructor should:
1. Load class data via `loader.FindClassByName(character.Info.Class)`
2. Load feat data via `loader.GetFeats()`
3. Determine applicable steps based on new level and class
4. Initialize the RNG with `rand.New(rand.NewSource(time.Now().UnixNano()))`
5. Pre-populate `stagedNewFeatures` from class features at the new level
6. Pre-populate `stagedSpellSlots` from class spell slot progression

The `determineSteps()` method should build the `steps` slice:
- Always include `LevelUpStepHP`
- Include `LevelUpStepSubclass` if:
  - Class has subclasses AND character has no subclass AND new level matches a subclass feature level
  - Check by looking for features named "*Subclass*" at the new level in class data
- Include `LevelUpStepASI` if new level is an ASI level (check features for "Ability Score Improvement")
- Include `LevelUpStepFeatures` if there are new features at this level
- Include `LevelUpStepSpellSlots` if class is a spellcaster and spell slots change at this level
- Always include `LevelUpStepConfirm`

**Step 2: Implement HP step**

The HP step handler:
- Up/Down or cursor to select between Roll and Average
- Enter on "Roll" performs a simulated dice roll: `1 + rng.Intn(dieSize)` where dieSize is parsed from class hitDice (e.g., "d10" → 10)
- Enter on "Average" calculates `(dieSize/2) + 1`
- HP gained = chosen value + CON modifier (minimum 1 total)
- Store in `stagedHPIncrease`
- Show preview: "HP: {current max} → {current max + gain} (+{gain})"
- Press Enter again to advance to next step

The HP step view:
```
Level Up: Level {old} → {new}
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 1 of N: Hit Points

Hit Die: d{X} ({className})
CON Modifier: +{mod}

> Roll (1d{X} + {mod})     [Result: {N}]    → +{total} HP
  Take Average ({avg} + {mod})              → +{total} HP

New Max HP: {old} → {new} (+{gain})

Enter: confirm | Esc: cancel level-up
```

**Step 3: Write tests for HP step**

Create `internal/ui/views/level_up_test.go`:

```go
func TestLevelUpModel_HPStepAverage(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Fighter", "Human", "Fighter")
	char.Info.Level = 1
	char.CombatStats.HitPoints.Maximum = 12
	char.CombatStats.HitDice.DieType = 10
	char.AbilityScores.Constitution.Score = 14 // +2 mod

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewLevelUpModel(char, store, loader)
	require.NotNil(t, model)

	// Should start on HP step
	assert.Equal(t, LevelUpStepHP, model.currentStep)

	// Select "Take Average" (cursor down)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, model.hpCursorPos) // Average selected

	// Confirm
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// For d10 + CON mod 2: average = 6, total = 8
	assert.Equal(t, 8, model.stagedHPIncrease)
}

func TestLevelUpModel_HPStepRoll(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Fighter", "Human", "Fighter")
	char.Info.Level = 1
	char.CombatStats.HitDice.DieType = 10
	char.AbilityScores.Constitution.Score = 10 // +0 mod

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewLevelUpModel(char, store, loader)

	// Roll (cursor already on Roll)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Result should be between 1 and 10 (die) + 0 (CON) = 1-10
	assert.GreaterOrEqual(t, model.stagedHPIncrease, 1)
	assert.LessOrEqual(t, model.stagedHPIncrease, 10)
}
```

**Step 4: Run tests**

Run: `go test ./internal/ui/views/ -run TestLevelUpModel -v`
Expected: All tests pass.

**Step 5: Commit**

```bash
git add internal/ui/views/level_up.go internal/ui/views/level_up_test.go
git commit -m "feat: add level-up wizard with HP increase step"
```

---

### Task 4: Level-Up Wizard — Subclass Selection Step

**Files:**
- Modify: `internal/ui/views/level_up.go`
- Modify: `internal/ui/views/level_up_test.go`

**Step 1: Implement subclass selection step**

Add the subclass step handler in `Update()`:
- Up/Down navigates subclass list
- Enter selects the highlighted subclass
- Sets `stagedSubclass` to the subclass name
- Populates `stagedSubclassFeats` with features from the selected subclass at or below the new level
- Advances to next step

Add the subclass step view:
```
Step N of M: Choose Your Subclass
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

{className} Subclass (Level {N})

> Champion
  Battle Master
  Eldritch Knight
  ...

──────────────────────────
{Selected subclass description}

Features at Level {N}:
• {feature name}: {description}

Enter: select | ↑↓: navigate | Esc: back
```

**Step 2: Write subclass test**

```go
func TestLevelUpModel_SubclassStep(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Fighter", "Human", "Fighter")
	char.Info.Level = 2 // Leveling to 3 (subclass level for Fighter)
	char.CombatStats.HitDice.DieType = 10
	char.AbilityScores.Constitution.Score = 10

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewLevelUpModel(char, store, loader)

	// Skip HP step
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Roll/select HP
	// Should now be on subclass step (if Fighter has subclass at level 3)

	if model.currentStep == LevelUpStepSubclass {
		// Select first subclass
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		assert.NotEmpty(t, model.stagedSubclass)
	}
}
```

**Step 3: Run tests**

Run: `go test ./internal/ui/views/ -run TestLevelUpModel -v`
Expected: All tests pass.

**Step 4: Commit**

```bash
git add internal/ui/views/level_up.go internal/ui/views/level_up_test.go
git commit -m "feat(level-up): add subclass selection step"
```

---

### Task 5: Level-Up Wizard — ASI/Feat Step

**Files:**
- Modify: `internal/ui/views/level_up.go`
- Modify: `internal/ui/views/level_up_test.go`

**Step 1: Implement ASI mode**

Add the ASI step handler:
- Tab or Left/Right switches between ASI and Feat mode
- In ASI mode:
  - Up/Down selects ability
  - Left/Right switches between +2/one and +1/two patterns
  - Enter toggles ability selection
  - For +2 pattern: one ability gets +2 (max 20)
  - For +1/+1 pattern: two abilities each get +1 (max 20)
  - Abilities at 20 are skipped/greyed out
- When selections complete, sets `stagedASIChanges` and advances

ASI view:
```
Step N of M: Ability Score Improvement
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Mode: > ASI  |  Feat

Pattern: > +2 to one ability  |  +1 to two abilities

Select ability:        Current → New
> Strength              16 → 18  ✓
  Dexterity             14
  Constitution          15
  Intelligence          10
  Wisdom                12
  Charisma               8

Enter: select | Tab: switch mode | ↑↓: navigate | Esc: back
```

**Step 2: Implement Feat mode**

In Feat mode:
- Shows searchable list of feats from `featData`
- Type to filter by name (same search pattern as spellbook add spell)
- Up/Down navigates filtered list
- Right panel shows feat description
- Ineligible feats (prerequisites not met) are greyed out
- Enter selects feat
- If feat has `AbilityScoreIncrease`, sub-prompt to choose which ability
- Sets `stagedFeat` and optionally `stagedFeatASIAbility`

Feat view:
```
Step N of M: Ability Score Improvement
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Mode:   ASI  | > Feat

Search: ___

> Alert (+1 DEX/CHA/WIS)
  Great Weapon Master
  Sentinel
  Sharpshooter
  ...

────────────────────────────
Alert
You gain a +5 bonus to Initiative...

Prerequisites: None
Category: General

Enter: select | Tab: switch mode | ↑↓: navigate | Esc: back
```

**Step 3: Write ASI/Feat tests**

```go
func TestLevelUpModel_ASIStep(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Fighter", "Human", "Fighter")
	char.Info.Level = 3 // Leveling to 4 (ASI level)
	char.CombatStats.HitDice.DieType = 10
	char.AbilityScores.Strength.Score = 16
	char.AbilityScores.Constitution.Score = 14

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewLevelUpModel(char, store, loader)

	// Skip to ASI step (advance through HP, possibly subclass)
	// ... navigate to ASI step ...

	if model.currentStep == LevelUpStepASI {
		// Select +2 to STR (index 0)
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Select STR
		assert.Equal(t, 2, model.stagedASIChanges[0]) // STR +2
	}
}
```

**Step 4: Run tests**

Run: `go test ./internal/ui/views/ -run TestLevelUpModel -v`
Expected: All tests pass.

**Step 5: Commit**

```bash
git add internal/ui/views/level_up.go internal/ui/views/level_up_test.go
git commit -m "feat(level-up): add ASI and feat selection step"
```

---

### Task 6: Level-Up Wizard — Features Summary and Spell Slot Steps

**Files:**
- Modify: `internal/ui/views/level_up.go`
- Modify: `internal/ui/views/level_up_test.go`

**Step 1: Implement features summary step**

Features step handler:
- Read-only display of class features gained at the new level
- Up/Down scrolls if many features
- Enter advances to next step

Features view:
```
Step N of M: New Features
━━━━━━━━━━━━━━━━━━━━━━━━━

Features gained at Level {N}:

• {Feature Name}
  {Feature description, word-wrapped}

• {Feature Name}
  {Feature description}

Enter: continue | ↑↓: scroll | Esc: back
```

**Step 2: Implement spell slot update step**

Spell slot step handler:
- Read-only display of spell slot changes
- Automatically calculates from `classData.SpellSlots`
- Enter advances to confirmation step

Spell slot view:
```
Step N of M: Spell Slot Update
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Your spell slots have been updated:

Level 1:  3 → 4 slots
Level 2:  (new!) 2 slots

You can now cast Level 2 spells!

Enter: continue | Esc: back
```

**Step 3: Write tests**

```go
func TestLevelUpModel_FeaturesStep(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Fighter", "Human", "Fighter")
	char.Info.Level = 1
	char.CombatStats.HitDice.DieType = 10
	char.AbilityScores.Constitution.Score = 10

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewLevelUpModel(char, store, loader)
	assert.Greater(t, len(model.stagedNewFeatures), 0, "Should have features for level 2")
}
```

**Step 4: Run tests**

Run: `go test ./internal/ui/views/ -run TestLevelUpModel -v`
Expected: All tests pass.

**Step 5: Commit**

```bash
git add internal/ui/views/level_up.go internal/ui/views/level_up_test.go
git commit -m "feat(level-up): add features summary and spell slot update steps"
```

---

### Task 7: Level-Up Wizard — Confirmation and Apply

**Files:**
- Modify: `internal/ui/views/level_up.go`
- Modify: `internal/ui/views/level_up_test.go`

**Step 1: Implement confirmation step**

Confirmation step handler:
- Shows summary of all staged changes
- Enter applies all changes and sends `LevelUpCompleteMsg`
- Esc cancels (sends `BackToSheetMsg`)

Confirmation view:
```
Step N of N: Confirm Level Up
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Level: {old} → {new}
HP: {old max} → {new max} (+{gain})
Subclass: {name} (if selected)
ASI: {ability} +{N} (if taken)
Feat: {feat name} (if taken)
Spell Slots: Updated (if changed)
New Features:
  • {feature 1}
  • {feature 2}

Enter: confirm and apply | Esc: cancel level-up
```

**Step 2: Implement applyLevelUp() method**

`applyLevelUp()` takes all staged changes and applies them to the character:

```go
func (m *LevelUpModel) applyLevelUp() {
	char := m.character

	// 1. Increment level
	char.LevelUp()

	// 2. Apply HP increase
	char.CombatStats.HitPoints.Maximum += m.stagedHPIncrease
	char.CombatStats.HitPoints.Current += m.stagedHPIncrease

	// 3. Apply subclass
	if m.stagedSubclass != "" {
		char.Info.Subclass = m.stagedSubclass
		for _, feat := range m.stagedSubclassFeats {
			char.Features.ClassFeatures = append(char.Features.ClassFeatures, feat)
		}
	}

	// 4. Apply ASI changes
	abilities := []*models.AbilityScore{
		&char.AbilityScores.Strength,
		&char.AbilityScores.Dexterity,
		&char.AbilityScores.Constitution,
		&char.AbilityScores.Intelligence,
		&char.AbilityScores.Wisdom,
		&char.AbilityScores.Charisma,
	}
	for i, delta := range m.stagedASIChanges {
		if delta > 0 {
			abilities[i].Score += delta
			if abilities[i].Score > 20 {
				abilities[i].Score = 20
			}
		}
	}

	// 5. Apply feat
	if m.stagedFeat != nil {
		char.Features.AddFeat(m.stagedFeat.Name, m.stagedFeat.Description)
		// Apply feat ASI if applicable
		if m.stagedFeat.Effects.AbilityScoreIncrease != nil && m.stagedFeatASIAbility >= 0 {
			abilities[m.stagedFeatASIAbility].Score += m.stagedFeat.Effects.AbilityScoreIncrease.Amount
			if abilities[m.stagedFeatASIAbility].Score > 20 {
				abilities[m.stagedFeatASIAbility].Score = 20
			}
		}
	}

	// 6. Apply spell slot changes
	if char.Spellcasting != nil && len(m.stagedSpellSlots) > 0 {
		for level, count := range m.stagedSpellSlots {
			char.Spellcasting.SpellSlots.SetSlots(level, count)
		}
	}

	// 7. Add new class features
	for _, feat := range m.stagedNewFeatures {
		char.Features.ClassFeatures = append(char.Features.ClassFeatures, feat)
	}

	// 8. Recalculate HP for Tough feat (if applicable)
	if m.stagedFeat != nil && m.stagedFeat.Effects.HPPerLevel > 0 {
		char.CombatStats.HitPoints.Maximum += m.stagedFeat.Effects.HPPerLevel * m.newLevel
		char.CombatStats.HitPoints.Current += m.stagedFeat.Effects.HPPerLevel * m.newLevel
	}

	char.MarkUpdated()
}
```

**Step 3: Write confirmation and apply test**

```go
func TestLevelUpModel_ConfirmAppliesChanges(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Fighter", "Human", "Fighter")
	char.Info.Level = 1
	char.Info.ExperiencePoints = 300 // Enough for level 2
	char.CombatStats.HitPoints.Maximum = 12
	char.CombatStats.HitPoints.Current = 12
	char.CombatStats.HitDice.DieType = 10
	char.AbilityScores.Constitution.Score = 14 // +2 mod

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewLevelUpModel(char, store, loader)

	// Navigate to average HP, then through all steps to confirmation
	// Select average (cursor down, enter)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Navigate through remaining steps to confirmation (Enter to advance each)
	for model.currentStep != LevelUpStepConfirm {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	}

	// Confirm
	oldLevel := char.Info.Level
	oldMaxHP := char.CombatStats.HitPoints.Maximum
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Verify changes applied
	assert.Equal(t, oldLevel+1, char.Info.Level)
	assert.Greater(t, char.CombatStats.HitPoints.Maximum, oldMaxHP)

	// Should produce a LevelUpCompleteMsg command
	assert.NotNil(t, cmd)
}
```

**Step 4: Run tests**

Run: `go test ./internal/ui/views/ -run TestLevelUpModel -v`
Expected: All tests pass.

**Step 5: Commit**

```bash
git add internal/ui/views/level_up.go internal/ui/views/level_up_test.go
git commit -m "feat(level-up): add confirmation step and apply level-up changes"
```

---

### Task 8: Wire Up Level-Up in App Model

**Files:**
- Modify: `internal/ui/model.go` — add levelUpModel field, handle messages, route view

**Step 1: Add level-up model field and view routing**

In `Model` struct, add:
```go
levelUpModel *views.LevelUpModel
```

**Step 2: Handle level-up messages in Update()**

Add cases in the `Update()` method:

```go
case views.OpenLevelUpMsg:
	m.levelUpModel = views.NewLevelUpModel(m.character, m.storage, m.loader)
	if m.width > 0 && m.height > 0 {
		m.levelUpModel, _ = m.levelUpModel.Update(tea.WindowSizeMsg{
			Width:  m.width,
			Height: m.height,
		})
	}
	m.currentView = ViewLevelUp
	return m, m.levelUpModel.Init()

case views.LevelUpCompleteMsg:
	// Return to main sheet after level-up
	m.currentView = ViewMainSheet
	m.levelUpModel = nil
	// Rebuild main sheet model to reflect new stats
	m.mainSheetModel = views.NewMainSheetModel(m.character, m.storage)
	if m.width > 0 && m.height > 0 {
		m.mainSheetModel, _ = m.mainSheetModel.Update(tea.WindowSizeMsg{
			Width:  m.width,
			Height: m.height,
		})
	}
	return m, m.mainSheetModel.Init()
```

**Step 3: Update BackToSheetMsg handler**

In the existing `BackToSheetMsg` handler, also clear `levelUpModel`:

```go
case views.BackToSheetMsg:
	m.currentView = ViewMainSheet
	m.inventoryModel = nil
	m.spellbookModel = nil
	m.levelUpModel = nil
	return m, nil
```

**Step 4: Route to level-up view in updateCurrentView()**

Add case in `updateCurrentView()`:

```go
case ViewLevelUp:
	if m.levelUpModel != nil {
		updatedModel, c := m.levelUpModel.Update(msg)
		m.levelUpModel = updatedModel
		cmd = c
	}
```

**Step 5: Update renderLevelUp()**

Replace the stub:

```go
func (m Model) renderLevelUp() string {
	if m.levelUpModel != nil {
		return m.levelUpModel.View()
	}
	return "Level Up View (loading...)"
}
```

**Step 6: Run all tests**

Run: `go test ./... -v`
Expected: All tests pass.

**Step 7: Commit**

```bash
git add internal/ui/model.go
git commit -m "feat: wire up level-up wizard in app model routing"
```

---

### Task 9: Integration Testing and Polish

**Files:**
- Modify: `internal/ui/views/level_up.go` — polish rendering
- Modify: `internal/ui/views/level_up_test.go` — add edge case tests
- Modify: `internal/ui/model_test.go` — add routing tests

**Step 1: Add edge case tests**

```go
func TestLevelUpModel_MaxLevelBlocked(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Fighter", "Human", "Fighter")
	char.Info.Level = 20

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	// NewLevelUpModel should handle level 20 gracefully
	model := NewLevelUpModel(char, store, loader)
	// The model can be nil or have no steps
	// The main sheet should prevent this from being reached
	assert.NotNil(t, model)
}

func TestLevelUpModel_CancelDoesNotApply(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Fighter", "Human", "Fighter")
	char.Info.Level = 1
	char.CombatStats.HitPoints.Maximum = 12
	char.CombatStats.HitDice.DieType = 10

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	model := NewLevelUpModel(char, store, loader)

	// Cancel immediately with Esc
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Level should NOT have changed
	assert.Equal(t, 1, char.Info.Level)
	assert.Equal(t, 12, char.CombatStats.HitPoints.Maximum)

	// Should produce BackToSheetMsg
	assert.NotNil(t, cmd)
}

func TestLevelUpModel_MilestoneNoXPCheck(t *testing.T) {
	char := models.NewCharacter("test-id", "Test Fighter", "Human", "Fighter")
	char.Info.Level = 1
	char.Info.ProgressionType = models.ProgressionMilestone
	char.Info.ExperiencePoints = 0 // No XP
	char.CombatStats.HitDice.DieType = 10

	store, err := storage.NewCharacterStorage("")
	require.NoError(t, err)
	loader := data.NewLoader("../../data")

	// Should work fine for milestone — no XP check needed
	model := NewLevelUpModel(char, store, loader)
	assert.NotNil(t, model)
	assert.Equal(t, LevelUpStepHP, model.currentStep)
}
```

**Step 2: Run all tests**

Run: `go test ./... -v`
Expected: All tests pass.

**Step 3: Build and verify the binary compiles**

Run: `go build ./cmd/sheet/`
Expected: Compiles without errors.

**Step 4: Commit and prepare PR**

```bash
git add .
git commit -m "test: add level-up edge case and integration tests"
```

---

### Task 10: Final PR

**Step 1: Run all tests one final time**

Run: `go test ./...`
Expected: All pass, no failures.

**Step 2: Push branch and create PR**

```bash
git push -u origin feature/character-progression
gh pr create --title "feat: Add character progression (Phase 10)" --body "$(cat <<'EOF'
## Summary
- Add XP input modal (`x` key) on main sheet with level-up prompts
- Add level-up key (`L`) for both XP and milestone characters
- Add long rest level-up prompt when XP threshold reached
- Implement multi-step level-up wizard with:
  - HP increase (roll or average)
  - Subclass selection (data-driven from classes.json)
  - ASI (+2/one or +1/two) or Feat selection (from new feats.json)
  - New features summary display
  - Spell slot progression updates
  - Confirmation step with full change summary
- Add feats.json data file with 2024 5e SRD feats
- All changes staged and only applied on confirmation

## Test plan
- [ ] Unit tests for HP step (roll and average)
- [ ] Unit tests for subclass selection
- [ ] Unit tests for ASI/feat selection
- [ ] Unit tests for confirmation and apply
- [ ] Edge case tests (level 20, cancel, milestone)
- [ ] All existing tests pass
- [ ] Manual testing: create character, gain XP, level up through wizard

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

**Step 3: Wait for review**
