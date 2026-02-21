# Codebase Consistency Fixes Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix all identified inconsistencies, bugs, and best-practice violations found in the deep codebase analysis (excluding CI/CD).

**Architecture:** The project is a D&D 5e character sheet TUI in Go using Bubble Tea (MVU pattern) with layered architecture: `domain` -> `models` -> `storage`/`data` -> `ui`. Changes span bug fixes, safety guards, atomic writes, dependency cleanup, DRY refactors, and test additions.

**Tech Stack:** Go 1.24, Bubble Tea, Lipgloss, Testify, github.com/Domo929/roll v1.0.0

---

### Task 1: Fix GetEquipment() nil-return bug in loader.go

**Priority:** P0 CRITICAL

**Context:** `GetEquipment()` at `internal/data/loader.go:492` returns `l.equipment` *before* `loadEquipmentUnsafe()` executes, so the first call returns nil data with a nil error. Every other getter (e.g., `GetFeats()` at lines 321-342) correctly separates the load call from the return. The fix also needs a double-check after acquiring the write lock (same as `GetFeats`).

**Files:**
- Modify: `internal/data/loader.go:481-493`
- Modify: `internal/data/loader_test.go` (verify existing test passes)

**Step 1: Fix GetEquipment implementation**

Replace lines 481-493 in `internal/data/loader.go` with:

```go
// GetEquipment returns all equipment data, loading it if necessary.
func (l *Loader) GetEquipment() (*Equipment, error) {
	l.mu.RLock()
	if l.equipment != nil {
		defer l.mu.RUnlock()
		return l.equipment, nil
	}
	l.mu.RUnlock()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Double-check after acquiring write lock
	if l.equipment != nil {
		return l.equipment, nil
	}

	if err := l.loadEquipmentUnsafe(); err != nil {
		return nil, err
	}

	return l.equipment, nil
}
```

**Step 2: Run existing loader tests to verify fix**

Run: `cd /home/dcupo/Software/sheet && go test ./internal/data/ -run TestLoaderGetEquipment -v`
Expected: PASS - equipment is now correctly loaded and returned

**Step 3: Commit**

```bash
git add internal/data/loader.go
git commit -m "fix: GetEquipment() nil-return bug — load before return

GetEquipment() was returning l.equipment before loadEquipmentUnsafe()
executed (Go evaluates return values left-to-right). Added double-check
after write lock acquisition, matching the pattern used by GetFeats()."
```

---

### Task 2: Guard skillNameToKey() against empty string panic

**Priority:** P0 CRITICAL

**Context:** `skillNameToKey()` at `internal/ui/views/proficiency_selection.go:358` does `normalized[:1]` which panics on an empty string. The fix adds an early return. The silent fallback to `SkillAcrobatics` on unknown input is also noted but kept for now since it's a display-only path.

**Files:**
- Modify: `internal/ui/views/proficiency_selection.go:354-399`
- Modify: `internal/ui/views/proficiency_selection_test.go` (add test)

**Step 1: Write the failing test**

Add to `internal/ui/views/proficiency_selection_test.go`:

```go
func TestSkillNameToKeyEmptyString(t *testing.T) {
	// Should not panic on empty string
	assert.NotPanics(t, func() {
		result := skillNameToKey("")
		assert.Equal(t, models.SkillAcrobatics, result, "empty string should return default")
	})
}

func TestSkillNameToKeyValidSkills(t *testing.T) {
	tests := []struct {
		input    string
		expected models.SkillName
	}{
		{"Acrobatics", models.SkillAcrobatics},
		{"Animal Handling", models.SkillAnimalHandling},
		{"Sleight of Hand", models.SkillSleightOfHand},
		{"Stealth", models.SkillStealth},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, skillNameToKey(tt.input))
		})
	}
}
```

**Step 2: Run test to verify it fails (panics)**

Run: `cd /home/dcupo/Software/sheet && go test ./internal/ui/views/ -run TestSkillNameToKeyEmptyString -v`
Expected: FAIL with panic: runtime error: index out of range

**Step 3: Fix skillNameToKey**

Replace lines 354-358 in `internal/ui/views/proficiency_selection.go` with:

```go
// skillNameToKey converts a skill display name to a SkillName constant.
func skillNameToKey(name string) models.SkillName {
	if name == "" {
		return models.SkillAcrobatics
	}
	// Normalize the name by removing spaces and converting to camelCase
	normalized := strings.ReplaceAll(name, " ", "")
	normalized = strings.ToLower(normalized[:1]) + normalized[1:]
```

**Step 4: Run tests to verify they pass**

Run: `cd /home/dcupo/Software/sheet && go test ./internal/ui/views/ -run TestSkillNameToKey -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/ui/views/proficiency_selection.go internal/ui/views/proficiency_selection_test.go
git commit -m "fix: guard skillNameToKey() against empty string panic

Added early return for empty input to prevent index-out-of-range panic
on normalized[:1]. Added tests for empty and valid skill name inputs."
```

---

### Task 3: Make CharacterStorage.Save() atomic

**Priority:** P1 HIGH

**Context:** `Save()` at `internal/storage/character_storage.go:77-101` uses `os.Create()` directly, which truncates the file before writing. If the process crashes mid-write, the character file is corrupted. The codebase already has the correct pattern in `cmd/migrate/main.go:125-148` (temp file + `os.Rename`).

**Files:**
- Modify: `internal/storage/character_storage.go:77-101`
- Modify: `internal/storage/character_storage_test.go` (add atomicity test)

**Step 1: Write the failing test**

Add to `internal/storage/character_storage_test.go`:

```go
func TestSaveAtomicity(t *testing.T) {
	store := createTestStorage(t)
	char := createTestCharacter()

	// First save
	path, err := store.Save(char)
	require.NoError(t, err)

	// Verify no .tmp file remains after successful save
	_, err = os.Stat(path + ".tmp")
	assert.True(t, os.IsNotExist(err), "temp file should not exist after successful save")

	// Verify the saved file has valid content
	loaded, err := store.Load(char.Info.Name)
	require.NoError(t, err)
	assert.Equal(t, char.Info.Name, loaded.Info.Name)
}
```

**Step 2: Run test to verify it passes (baseline)**

Run: `cd /home/dcupo/Software/sheet && go test ./internal/storage/ -run TestSaveAtomicity -v`
Expected: PASS (the test itself should pass even before the fix, since it's testing behavior not implementation — but we're adding the atomicity guarantee)

**Step 3: Implement atomic Save**

Replace the `Save` method (lines 76-101) in `internal/storage/character_storage.go` with:

```go
// Save saves a character to disk using atomic write (temp file + rename).
// Returns the path where the character was saved.
func (cs *CharacterStorage) Save(character *models.Character) (string, error) {
	if character == nil {
		return "", errors.New("character cannot be nil")
	}

	if character.Info.Name == "" {
		return "", ErrInvalidCharacterName
	}

	path := cs.getCharacterPath(character.Info.Name)
	tempPath := path + ".tmp"

	// Write to temp file first
	file, err := os.Create(tempPath)
	if err != nil {
		return "", fmt.Errorf("failed to create temp character file: %w", err)
	}

	if err := character.WriteTo(file); err != nil {
		file.Close()
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to write character: %w", err)
	}

	if err := file.Close(); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to close temp character file: %w", err)
	}

	// Atomically replace the original file
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to finalize character save: %w", err)
	}

	return path, nil
}
```

**Step 4: Run all storage tests**

Run: `cd /home/dcupo/Software/sheet && go test ./internal/storage/ -v`
Expected: All PASS

**Step 5: Commit**

```bash
git add internal/storage/character_storage.go internal/storage/character_storage_test.go
git commit -m "fix: make Save() atomic with temp file + rename

Write to .tmp file first, then os.Rename to target path. Prevents
character data corruption if the process crashes mid-write. Pattern
matches cmd/migrate/main.go saveCharacter()."
```

---

### Task 4: Replace go.mod local replace with roll v1.0.0

**Priority:** P1 HIGH

**Context:** `go.mod` line 38 has `replace github.com/Domo929/roll => /home/dcupo/Software/roll` which only works on the developer's machine. The `roll` module now has a published `v1.0.0` tag.

**Files:**
- Modify: `go.mod`
- Modify: `go.sum` (auto-updated by `go mod tidy`)

**Step 1: Remove replace directive and update dependency**

```bash
cd /home/dcupo/Software/sheet
# Remove the local replace directive
go mod edit -dropreplace github.com/Domo929/roll
# Add the specific version
go mod edit -require github.com/Domo929/roll@v1.0.0
# Tidy to update go.sum
go mod tidy
```

**Step 2: Verify the project builds**

Run: `cd /home/dcupo/Software/sheet && go build ./...`
Expected: Successful build with no errors

**Step 3: Run all tests to verify nothing broke**

Run: `cd /home/dcupo/Software/sheet && go test ./... -count=1`
Expected: All PASS

**Step 4: Commit**

```bash
git add go.mod go.sum
git commit -m "fix: replace local roll dependency with published v1.0.0

Remove non-portable local replace directive that only worked on one
developer's machine. Use published github.com/Domo929/roll v1.0.0."
```

---

### Task 5: Extract view-init helper to eliminate DRY violations in model.go

**Priority:** P2 MEDIUM

**Context:** `internal/ui/model.go` has 6+ repetitions of the same view initialization pattern: create model -> apply WindowSizeMsg -> set currentView -> return Init(). Extract a helper method to consolidate.

**Files:**
- Modify: `internal/ui/model.go`

**Step 1: Add helper type and method**

Add this interface and helper method to `internal/ui/model.go` (after the Model struct, before `Init()`):

```go
// viewModel is implemented by all sub-views that accept Init/Update/View.
type viewModel interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (viewModel, tea.Cmd)
}

// initView applies the current window size to a view and returns its Init command.
func (m *Model) initView(view tea.Model) tea.Cmd {
	if m.width > 0 && m.height > 0 {
		view.Update(tea.WindowSizeMsg{
			Width:  m.width,
			Height: m.height,
		})
	}
	return nil
}
```

Wait — the sub-view models don't share a common interface and use concrete types with different `Update()` signatures. A generic helper won't work cleanly here. Instead, extract a simpler helper that just sends the WindowSizeMsg:

```go
// sizeMsg returns a WindowSizeMsg if the model has known dimensions, nil otherwise.
func (m *Model) sizeMsg() *tea.WindowSizeMsg {
	if m.width > 0 && m.height > 0 {
		return &tea.WindowSizeMsg{Width: m.width, Height: m.height}
	}
	return nil
}
```

Then replace each repeated block. For example, `StartCharacterCreationMsg` (lines 106-116):

Before:
```go
case views.StartCharacterCreationMsg:
	m.characterCreationModel = views.NewCharacterCreationModel(m.storage, m.loader)
	if m.width > 0 && m.height > 0 {
		m.characterCreationModel, _ = m.characterCreationModel.Update(tea.WindowSizeMsg{
			Width:  m.width,
			Height: m.height,
		})
	}
	m.currentView = ViewCharacterCreation
	return m, m.characterCreationModel.Init()
```

After:
```go
case views.StartCharacterCreationMsg:
	m.characterCreationModel = views.NewCharacterCreationModel(m.storage, m.loader)
	if msg := m.sizeMsg(); msg != nil {
		m.characterCreationModel, _ = m.characterCreationModel.Update(*msg)
	}
	m.currentView = ViewCharacterCreation
	return m, m.characterCreationModel.Init()
```

Apply this same pattern to ALL instances:
- `StartCharacterCreationMsg` (line 107-116)
- `OpenInventoryMsg` (line 157-165)
- `OpenSpellbookMsg` (line 169-177)
- `OpenLevelUpMsg` (line 181-189)
- `OpenNotesMsg` (line 193-200)
- `OpenCharacterInfoMsg` (line 203-211)
- `LevelUpCompleteMsg` (line 222-229)
- `CharacterLoadedMsg` (line 241-248)
- `LoadCharacterMsg` (line 264-270)

**Step 2: Extract named constant for minimum roll history width**

Add near the top of `internal/ui/model.go` (after the ViewType constants):

```go
// minRollHistoryWidth is the minimum terminal width required to show the roll history column.
const minRollHistoryWidth = 80
```

Replace `m.width >= 80` at line 317 and `m.width >= 80` at line 439 with `m.width >= minRollHistoryWidth`.

**Step 3: Remove TODO stubs from render methods**

In `renderCharacterSelection()` (line 460), `renderCharacterCreation()` (line 467): remove the `(TODO)` text from the fallback strings — these fallbacks should never appear in practice since the models are always initialized before these views are rendered. Replace with more accurate messages:

```go
// In renderCharacterSelection:
return "Loading character selection..."

// In renderCharacterCreation:
return "Loading character creation..."
```

For `renderCombat()` (line 509) and `renderRest()` (line 513), these are genuinely unimplemented views. Keep the TODO but remove `\n\nPress q to quit.` since the parent handles quitting:

```go
func (m Model) renderCombat() string {
	return "Combat View — coming soon"
}

func (m Model) renderRest() string {
	return "Rest View — coming soon"
}
```

**Step 4: Run the full test suite**

Run: `cd /home/dcupo/Software/sheet && go test ./... -count=1`
Expected: All PASS (or at least no regressions — these are refactors)

**Step 5: Commit**

```bash
git add internal/ui/model.go
git commit -m "refactor: extract sizeMsg() helper and named constants in model.go

- Add sizeMsg() to eliminate repeated WindowSizeMsg construction
- Add minRollHistoryWidth constant to replace magic number 80
- Clean up TODO stubs and fallback render messages"
```

---

### Task 6: Add logging for silently skipped files in CharacterStorage.List()

**Priority:** P2 MEDIUM

**Context:** `List()` at `internal/storage/character_storage.go:203-249` silently skips files it can't read or parse (lines 226-228, 236-237). Add `log` import and log warnings so debugging is possible.

**Files:**
- Modify: `internal/storage/character_storage.go:203-249`

**Step 1: Add log import and warnings**

Add `"log"` to the imports in `internal/storage/character_storage.go`.

In the `List()` method, replace the silent `continue` statements:

Line 226-228 — change:
```go
		// Skip files we can't read
		continue
```
to:
```go
		log.Printf("warning: skipping unreadable character file %s: %v", path, err)
		continue
```

Lines 236-237 — change:
```go
		// Skip files we can't parse
		continue
```
to:
```go
		log.Printf("warning: skipping unparseable character file %s: %v", path, err)
		continue
```

**Step 2: Run storage tests**

Run: `cd /home/dcupo/Software/sheet && go test ./internal/storage/ -v`
Expected: All PASS

**Step 3: Commit**

```bash
git add internal/storage/character_storage.go
git commit -m "fix: log warnings for skipped files in CharacterStorage.List()

Previously silently swallowed errors when character files couldn't be
read or parsed. Now logs warnings so issues are discoverable."
```

---

### Task 7: Standardize error return patterns (bool -> error)

**Priority:** P2 MEDIUM

**Context:** `HitDice.Use()` and `SpellSlotTracker.Use()` and `PactMagic.Use()` return `bool` for resource-spending operations. `Character.LevelUp()` returns `bool`. These should return `error` for consistency and better error messages. However, `Use()` is called in many places (main_sheet.go, spellbook.go, spellcasting.go) and `LevelUp()` is called from level_up.go and tested in character_info_test.go.

**DECISION:** After reviewing the call sites, changing `Use() bool` to `Use() error` would be a large, risky refactor touching 10+ files with no functional benefit (the caller always just checks the bool anyway). The `bool` return is idiomatic Go for "try" operations (like `map` lookups, channel receives). **Skip this change** — the current pattern is actually fine and consistent with Go idioms for "attempt" operations. Document this decision.

No code changes for this task.

---

### Task 8: Add tests for domain types

**Priority:** P3 LOW

**Context:** `internal/domain/activation.go`, `damage.go`, `weapon_property.go` are type/constant definitions with no tests. These are simple enums but testing that constants have expected values prevents accidental changes.

**Files:**
- Create: `internal/domain/domain_test.go`

**Step 1: Write tests**

Create `internal/domain/domain_test.go`:

```go
package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActivationTypes(t *testing.T) {
	assert.Equal(t, ActivationType("action"), ActivationAction)
	assert.Equal(t, ActivationType("bonus"), ActivationBonus)
	assert.Equal(t, ActivationType("reaction"), ActivationReaction)
	assert.Equal(t, ActivationType(""), ActivationPassive)
}

func TestDamageTypes(t *testing.T) {
	expected := []DamageType{
		DamageAcid, DamageBludgeoning, DamageCold, DamageFire,
		DamageForce, DamageLightning, DamageNecrotic, DamagePiercing,
		DamagePoison, DamagePsychic, DamageRadiant, DamageSlashing,
		DamageThunder,
	}
	assert.Len(t, expected, 13, "should have 13 damage types")

	// Verify string values match D&D conventions
	assert.Equal(t, DamageType("fire"), DamageFire)
	assert.Equal(t, DamageType("bludgeoning"), DamageBludgeoning)
	assert.Equal(t, DamageType("radiant"), DamageRadiant)
}

func TestWeaponProperties(t *testing.T) {
	expected := []WeaponProperty{
		PropertyFinesse, PropertyLight, PropertyHeavy, PropertyReach,
		PropertyThrown, PropertyVersatile, PropertyTwoHanded,
		PropertyAmmunition, PropertyLoading,
	}
	assert.Len(t, expected, 9, "should have 9 weapon properties")

	// Verify string values
	assert.Equal(t, WeaponProperty("finesse"), PropertyFinesse)
	assert.Equal(t, WeaponProperty("two-handed"), PropertyTwoHanded)
	assert.Equal(t, WeaponProperty("ammunition"), PropertyAmmunition)
}
```

**Step 2: Run tests**

Run: `cd /home/dcupo/Software/sheet && go test ./internal/domain/ -v`
Expected: All PASS

**Step 3: Commit**

```bash
git add internal/domain/domain_test.go
git commit -m "test: add tests for domain type constants

Verify activation types, damage types, and weapon properties have
expected string values. Prevents accidental constant value changes."
```

---

### Task 9: Add tests for panel.go and help.go components

**Priority:** P3 LOW

**Files:**
- Create: `internal/ui/components/panel_test.go`
- Create: `internal/ui/components/help_test.go`

**Step 1: Write panel tests**

Create `internal/ui/components/panel_test.go`:

```go
package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPanel(t *testing.T) {
	p := NewPanel("Title", "Content", 40, 10)
	assert.Equal(t, "Title", p.Title)
	assert.Equal(t, "Content", p.Content)
	assert.Equal(t, 40, p.Width)
	assert.Equal(t, 10, p.Height)
}

func TestPanelRender(t *testing.T) {
	t.Run("with title", func(t *testing.T) {
		p := NewPanel("Title", "Content", 40, 10)
		result := p.Render()
		assert.Contains(t, result, "Title")
		assert.Contains(t, result, "Content")
	})

	t.Run("without title", func(t *testing.T) {
		p := NewPanel("", "Content", 40, 10)
		result := p.Render()
		assert.Contains(t, result, "Content")
	})

	t.Run("zero dimensions", func(t *testing.T) {
		p := NewPanel("Title", "Content", 0, 0)
		result := p.Render()
		assert.Contains(t, result, "Content")
	})
}

func TestBox(t *testing.T) {
	result := Box("Hello", 20)
	assert.Contains(t, result, "Hello")
}

func TestDefaultPanelStyle(t *testing.T) {
	style := DefaultPanelStyle()
	// Just verify it doesn't panic and returns a style
	assert.NotEmpty(t, style.Render("test"))
}

func TestJoinHorizontal(t *testing.T) {
	result := JoinHorizontal(2, "A", "B", "C")
	assert.Contains(t, result, "A")
	assert.Contains(t, result, "B")
	assert.Contains(t, result, "C")
}

func TestJoinVertical(t *testing.T) {
	result := JoinVertical("Line1", "Line2")
	assert.Contains(t, result, "Line1")
	assert.Contains(t, result, "Line2")
}
```

**Step 2: Write help tests**

Create `internal/ui/components/help_test.go`:

```go
package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHelpFooter(t *testing.T) {
	bindings := []KeyBinding{
		{Key: "q", Description: "quit"},
		{Key: "enter", Description: "select"},
	}
	footer := NewHelpFooter(bindings...)
	assert.Len(t, footer.Bindings, 2)
}

func TestHelpFooterRender(t *testing.T) {
	t.Run("with bindings", func(t *testing.T) {
		footer := NewHelpFooter(
			KeyBinding{Key: "q", Description: "quit"},
			KeyBinding{Key: "enter", Description: "select"},
		)
		result := footer.Render()
		assert.Contains(t, result, "q")
		assert.Contains(t, result, "quit")
		assert.Contains(t, result, "enter")
		assert.Contains(t, result, "select")
	})

	t.Run("empty bindings", func(t *testing.T) {
		footer := NewHelpFooter()
		result := footer.Render()
		assert.Empty(t, result)
	})

	t.Run("with width", func(t *testing.T) {
		footer := HelpFooter{
			Bindings: []KeyBinding{{Key: "q", Description: "quit"}},
			Width:    80,
		}
		result := footer.Render()
		assert.Contains(t, result, "q")
	})
}

func TestCommonBindings(t *testing.T) {
	bindings := CommonBindings()
	assert.NotEmpty(t, bindings)
	// Should include standard vim-like navigation
	keys := make([]string, len(bindings))
	for i, b := range bindings {
		keys[i] = b.Key
	}
	assert.Contains(t, keys, "q")
	assert.Contains(t, keys, "enter")
}

func TestNavigationBindings(t *testing.T) {
	bindings := NavigationBindings()
	assert.NotEmpty(t, bindings)
}

func TestListBindings(t *testing.T) {
	bindings := ListBindings()
	assert.NotEmpty(t, bindings)
}
```

**Step 3: Run component tests**

Run: `cd /home/dcupo/Software/sheet && go test ./internal/ui/components/ -v`
Expected: All PASS

**Step 4: Commit**

```bash
git add internal/ui/components/panel_test.go internal/ui/components/help_test.go
git commit -m "test: add tests for Panel and HelpFooter components

Cover Panel creation, rendering with/without titles, Box helper,
JoinHorizontal/JoinVertical. Cover HelpFooter rendering with various
binding configurations and preset binding functions."
```

---

### Task 10: Add tests for roll_helpers.go

**Priority:** P3 LOW

**Files:**
- Create: `internal/ui/components/roll_helpers_test.go`

**Step 1: Write tests**

Create `internal/ui/components/roll_helpers_test.go`:

```go
package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSpellRollCmd(t *testing.T) {
	t.Run("no damage returns nil", func(t *testing.T) {
		cmd := BuildSpellRollCmd("Mage Hand", "", "", "", 5, 0)
		assert.Nil(t, cmd)
	})

	t.Run("spell attack with damage", func(t *testing.T) {
		cmd := BuildSpellRollCmd("Fire Bolt", "1d10", "fire", "", 7, 0)
		require.NotNil(t, cmd)

		msg := cmd()
		rollMsg, ok := msg.(RequestRollMsg)
		require.True(t, ok)
		assert.Equal(t, "Fire Bolt Attack", rollMsg.Label)
		assert.Equal(t, "1d20", rollMsg.DiceExpr)
		assert.Equal(t, 7, rollMsg.Modifier)
		assert.Equal(t, RollAttack, rollMsg.RollType)
		assert.True(t, rollMsg.AdvPrompt)

		// Check follow-up damage roll
		require.NotNil(t, rollMsg.FollowUp)
		assert.Equal(t, "Fire Bolt Damage (fire)", rollMsg.FollowUp.Label)
		assert.Equal(t, "1d10", rollMsg.FollowUp.DiceExpr)
		assert.Equal(t, RollDamage, rollMsg.FollowUp.RollType)
	})

	t.Run("save-based spell with damage", func(t *testing.T) {
		cmd := BuildSpellRollCmd("Fireball", "8d6", "fire", "DEX", 0, 15)
		require.NotNil(t, cmd)

		msg := cmd()
		rollMsg, ok := msg.(RequestRollMsg)
		require.True(t, ok)
		assert.Contains(t, rollMsg.Label, "Fireball Damage")
		assert.Contains(t, rollMsg.Label, "DC 15")
		assert.Contains(t, rollMsg.Label, "DEX")
		assert.Equal(t, "8d6", rollMsg.DiceExpr)
		assert.Equal(t, 0, rollMsg.Modifier)
		assert.Equal(t, RollDamage, rollMsg.RollType)
	})
}

// Verify BuildSpellRollCmd returns proper tea.Cmd type
func TestBuildSpellRollCmdReturnType(t *testing.T) {
	cmd := BuildSpellRollCmd("Test", "1d6", "fire", "", 5, 0)
	require.NotNil(t, cmd)
	var _ tea.Cmd = cmd // compile-time check
}
```

**Step 2: Run tests**

Run: `cd /home/dcupo/Software/sheet && go test ./internal/ui/components/ -run TestBuildSpellRollCmd -v`
Expected: All PASS

**Step 3: Commit**

```bash
git add internal/ui/components/roll_helpers_test.go
git commit -m "test: add tests for BuildSpellRollCmd helper

Cover no-damage, spell-attack, and save-based spell paths. Verify
roll types, labels, modifiers, and follow-up damage rolls."
```

---

### Task 11: Convert ListItem.Value from interface{} to any

**Priority:** P3 LOW

**Context:** `ListItem.Value` at `internal/ui/components/list.go:14` uses `interface{}`. Go 1.18+ has `any` as a built-in alias. Full generics conversion (`ListItem[T]`) would require changing all callers (character_selection.go, character_creation.go, list_test.go) and doesn't add practical value since the callers use different types (string, *Race, *Class, *Background). Just change `interface{}` to `any` for modernization.

**Files:**
- Modify: `internal/ui/components/list.go:14`

**Step 1: Replace interface{} with any**

In `internal/ui/components/list.go` line 14, change:
```go
	Value       interface{}
```
to:
```go
	Value       any
```

**Step 2: Run tests**

Run: `cd /home/dcupo/Software/sheet && go test ./internal/ui/components/ -v`
Expected: All PASS (any is an alias for interface{})

**Step 3: Commit**

```bash
git add internal/ui/components/list.go
git commit -m "refactor: replace interface{} with any in ListItem

Go 1.18+ provides 'any' as a built-in alias for interface{}.
Full generic parameterization deferred — callers use heterogeneous
types (string, *Race, *Class, *Background)."
```

---

### Task 12: Final verification

**Step 1: Run complete test suite**

Run: `cd /home/dcupo/Software/sheet && go test ./... -count=1`
Expected: All PASS

**Step 2: Run go vet**

Run: `cd /home/dcupo/Software/sheet && go vet ./...`
Expected: No warnings

**Step 3: Verify build**

Run: `cd /home/dcupo/Software/sheet && go build ./...`
Expected: Clean build
