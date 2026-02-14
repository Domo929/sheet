package views

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------------
// Enums
// ---------------------------------------------------------------------------

// LevelUpStep represents a step in the level-up wizard.
type LevelUpStep int

const (
	LevelUpStepHP        LevelUpStep = iota // Choose HP increase method
	LevelUpStepSubclass                     // Select a subclass (if applicable)
	LevelUpStepASI                          // Ability Score Improvement or Feat
	LevelUpStepFeatures                     // Review new class/subclass features
	LevelUpStepSpellSlots                   // Review spell slot changes
	LevelUpStepConfirm                      // Confirm and apply
)

// HPMethod represents how HP is gained on level up.
type HPMethod int

const (
	HPMethodRoll    HPMethod = iota // Roll the hit die
	HPMethodAverage                // Take the fixed average
)

// ASIMode represents the ASI/Feat tab selection.
type ASIMode int

const (
	ASIModeASI  ASIMode = iota // Ability Score Improvement
	ASIModeFeat                // Choose a Feat
)

// ASIPattern represents the ASI allocation pattern.
type ASIPattern int

const (
	ASIPatternPlus2     ASIPattern = iota // +2 to one ability
	ASIPatternPlus1Plus1                  // +1 to two abilities
)

// ---------------------------------------------------------------------------
// Key map
// ---------------------------------------------------------------------------

type levelUpKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Select key.Binding
	Back   key.Binding
	Tab    key.Binding
	Quit   key.Binding
}

func defaultLevelUpKeyMap() levelUpKeyMap {
	return levelUpKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back/cancel"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch mode"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

// LevelUpModel manages the level-up wizard.
type LevelUpModel struct {
	character *models.Character
	storage   *storage.CharacterStorage
	loader    *data.Loader
	keys      levelUpKeyMap
	width     int
	height    int
	rng       *rand.Rand

	// Class & feat data
	classData *data.Class
	featData  *data.FeatData

	// Level bookkeeping
	oldLevel int
	newLevel int

	// Steps
	steps       []LevelUpStep
	stepIndex   int
	currentStep LevelUpStep

	// Error / status
	err    error
	errMsg string

	// ---- HP step ----
	hpMethodCursor  int  // 0 = Roll, 1 = Average
	hpRolled        bool // true once a roll/average is locked in
	hpRollResult    int  // raw die result (before CON mod)
	stagedHPIncrease int // final HP gain (die + CON mod, min 1)

	// ---- Subclass step ----
	subclassCursor     int
	stagedSubclass     *data.Subclass
	stagedSubclassFeats []data.Feature

	// ---- ASI / Feat step ----
	asiMode           ASIMode
	asiPattern        ASIPattern
	asiAbilityCursor  int        // 0-5 index into abilities list
	asiSelected       [6]bool    // which abilities are toggled
	stagedASIChanges  [6]int     // delta per ability
	asiConfirmed      bool

	// Feat sub-mode
	featCursor        int
	featFilterText    string
	filteredFeats     []data.Feat
	stagedFeat        *data.Feat
	featASICursor     int    // cursor for choosing feat ASI ability
	featASIAbility    string // chosen ability for feat ASI
	inFeatASIPrompt   bool   // true when picking ability for feat ASI

	// ---- Features step ----
	stagedNewFeatures    []data.Feature // class features gained at newLevel
	featureScrollOffset  int

	// ---- Spell slots step ----
	stagedSpellSlots     *data.SpellSlot // new spell slot row at newLevel
	oldSpellSlots        *data.SpellSlot // spell slot row at oldLevel

	// ---- Confirmation ----
	quitting bool
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

// NewLevelUpModel creates a new level-up wizard model.
func NewLevelUpModel(character *models.Character, store *storage.CharacterStorage, loader *data.Loader) *LevelUpModel {
	m := &LevelUpModel{
		character: character,
		storage:   store,
		loader:    loader,
		keys:      defaultLevelUpKeyMap(),
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
		oldLevel:  character.Info.Level,
		newLevel:  character.Info.Level + 1,
	}

	// Load class data
	classData, err := loader.FindClassByName(character.Info.Class)
	if err != nil {
		m.errMsg = fmt.Sprintf("Failed to load class data: %v", err)
		// Provide a minimal step list so the wizard can still render an error
		m.steps = []LevelUpStep{LevelUpStepConfirm}
		m.currentStep = LevelUpStepConfirm
		return m
	}
	m.classData = classData

	// Load feat data (non-fatal if it fails)
	featData, _ := loader.GetFeats()
	m.featData = featData

	// Determine which steps are needed
	m.determineSteps()

	if len(m.steps) > 0 {
		m.currentStep = m.steps[0]
	}

	// Pre-populate staged class features at newLevel
	m.stagedNewFeatures = m.classFeaturesAtLevel(m.newLevel)

	// Pre-populate staged spell slots
	m.stagedSpellSlots = m.spellSlotRowForLevel(m.newLevel)
	m.oldSpellSlots = m.spellSlotRowForLevel(m.oldLevel)

	// Initialize filtered feats
	m.rebuildFilteredFeats()

	return m
}

// ---------------------------------------------------------------------------
// Step determination
// ---------------------------------------------------------------------------

func (m *LevelUpModel) determineSteps() {
	m.steps = nil

	// 1. Always HP
	m.steps = append(m.steps, LevelUpStepHP)

	// 2. Subclass: if class has subclasses, character has none, and this is the first subclass level
	if m.classData != nil && len(m.classData.Subclasses) > 0 && m.character.Info.Subclass == "" {
		firstSubLevel := m.firstSubclassLevel()
		if firstSubLevel > 0 && m.newLevel == firstSubLevel {
			m.steps = append(m.steps, LevelUpStepSubclass)
		}
	}

	// 3. ASI: if there's a feature named "Ability Score Improvement" at newLevel
	if m.hasFeatureAtLevel("Ability Score Improvement", m.newLevel) {
		m.steps = append(m.steps, LevelUpStepASI)
	}

	// 4. Features: if there are class features at newLevel (excluding ASI and subclass trigger)
	displayFeatures := m.displayableFeaturesAtLevel(m.newLevel)
	if len(displayFeatures) > 0 {
		m.steps = append(m.steps, LevelUpStepFeatures)
	}

	// 5. Spell slots: if class is a spellcaster and slots change
	if m.classData != nil && m.classData.Spellcaster && m.spellSlotsChanged() {
		m.steps = append(m.steps, LevelUpStepSpellSlots)
	}

	// 6. Always Confirm
	m.steps = append(m.steps, LevelUpStepConfirm)
}

// firstSubclassLevel returns the level at which the first subclass feature appears.
func (m *LevelUpModel) firstSubclassLevel() int {
	if m.classData == nil || len(m.classData.Subclasses) == 0 {
		return 0
	}
	sc := m.classData.Subclasses[0]
	if len(sc.Features) == 0 {
		return 0
	}
	return sc.Features[0].Level
}

// hasFeatureAtLevel checks if the class has a feature with the given name at the given level.
func (m *LevelUpModel) hasFeatureAtLevel(name string, level int) bool {
	if m.classData == nil {
		return false
	}
	for _, f := range m.classData.Features {
		if f.Level == level && f.Name == name {
			return true
		}
	}
	return false
}

// classFeaturesAtLevel returns all class features at the given level.
func (m *LevelUpModel) classFeaturesAtLevel(level int) []data.Feature {
	if m.classData == nil {
		return nil
	}
	var features []data.Feature
	for _, f := range m.classData.Features {
		if f.Level == level {
			features = append(features, f)
		}
	}
	return features
}

// displayableFeaturesAtLevel returns class features at the level excluding
// "Ability Score Improvement" and the subclass trigger feature.
func (m *LevelUpModel) displayableFeaturesAtLevel(level int) []data.Feature {
	all := m.classFeaturesAtLevel(level)
	var result []data.Feature
	for _, f := range all {
		if f.Name == "Ability Score Improvement" {
			continue
		}
		if m.isSubclassTriggerFeature(f) {
			continue
		}
		result = append(result, f)
	}
	return result
}

// isSubclassTriggerFeature checks if a feature is the trigger for subclass selection
// (e.g., "Martial Archetype", "Monastic Tradition", etc.).
func (m *LevelUpModel) isSubclassTriggerFeature(f data.Feature) bool {
	lower := strings.ToLower(f.Name)
	if strings.Contains(lower, "archetype") ||
		strings.Contains(lower, "subclass") ||
		strings.Contains(lower, "tradition") ||
		strings.Contains(lower, "oath") ||
		strings.Contains(lower, "patron") ||
		strings.Contains(lower, "origin") ||
		strings.Contains(lower, "domain") ||
		strings.Contains(lower, "circle") ||
		strings.Contains(lower, "conclave") ||
		strings.Contains(lower, "college") ||
		strings.Contains(lower, "path") ||
		strings.Contains(lower, "school") {
		// Only treat as trigger if at the first subclass level
		firstLevel := m.firstSubclassLevel()
		return firstLevel > 0 && f.Level == firstLevel
	}
	return false
}

// spellSlotRowForLevel finds the SpellSlot entry for the given class level.
func (m *LevelUpModel) spellSlotRowForLevel(level int) *data.SpellSlot {
	if m.classData == nil {
		return nil
	}
	for i := range m.classData.SpellSlots {
		if m.classData.SpellSlots[i].Level == level {
			return &m.classData.SpellSlots[i]
		}
	}
	return nil
}

// spellSlotsChanged returns true if spell slots differ between old and new level.
func (m *LevelUpModel) spellSlotsChanged() bool {
	oldSS := m.spellSlotRowForLevel(m.oldLevel)
	newSS := m.spellSlotRowForLevel(m.newLevel)
	if oldSS == nil && newSS == nil {
		return false
	}
	if oldSS == nil || newSS == nil {
		return true
	}
	for lvl := 1; lvl <= 9; lvl++ {
		if getSpellSlotCount(*oldSS, lvl) != getSpellSlotCount(*newSS, lvl) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Helper: spell slot count from data.SpellSlot
// ---------------------------------------------------------------------------

func getSpellSlotCount(ss data.SpellSlot, spellLevel int) int {
	switch spellLevel {
	case 1:
		return ss.First
	case 2:
		return ss.Second
	case 3:
		return ss.Third
	case 4:
		return ss.Fourth
	case 5:
		return ss.Fifth
	case 6:
		return ss.Sixth
	case 7:
		return ss.Seventh
	case 8:
		return ss.Eighth
	case 9:
		return ss.Ninth
	default:
		return 0
	}
}

// ---------------------------------------------------------------------------
// Word wrapping
// ---------------------------------------------------------------------------

// wordWrap wraps text to the given width, respecting word boundaries.
// It prepends the given indent to continuation lines.
func wordWrap(text string, width int, indent string) string {
	if width <= 0 {
		width = 80
	}
	// Account for indent width on first line is handled by caller
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) > width {
			lines = append(lines, currentLine)
			currentLine = word
		} else {
			currentLine += " " + word
		}
	}
	lines = append(lines, currentLine)

	return strings.Join(lines, "\n"+indent)
}

// descWidth returns the available width for description text, accounting for
// indentation. Falls back to 76 if terminal width is unknown.
func (m *LevelUpModel) descWidth(indent int) int {
	w := m.width - indent
	if w <= 20 {
		w = 76
	}
	return w
}

// ---------------------------------------------------------------------------
// Hit dice parsing
// ---------------------------------------------------------------------------

func (m *LevelUpModel) parseDieSize() int {
	if m.classData == nil {
		return 8
	}
	s := m.classData.HitDice
	s = strings.TrimSpace(s)
	// Handle "1d10" or "d10"
	idx := strings.Index(s, "d")
	if idx < 0 {
		return 8
	}
	n, err := strconv.Atoi(s[idx+1:])
	if err != nil || n <= 0 {
		return 8
	}
	return n
}

// ---------------------------------------------------------------------------
// Feat filtering
// ---------------------------------------------------------------------------

func (m *LevelUpModel) rebuildFilteredFeats() {
	if m.featData == nil {
		m.filteredFeats = nil
		return
	}
	filter := strings.ToLower(m.featFilterText)
	var result []data.Feat
	for _, f := range m.featData.Feats {
		if filter == "" || strings.Contains(strings.ToLower(f.Name), filter) {
			result = append(result, f)
		}
	}
	m.filteredFeats = result
	if m.featCursor >= len(m.filteredFeats) {
		m.featCursor = 0
	}
}

// ---------------------------------------------------------------------------
// Bubble Tea interface
// ---------------------------------------------------------------------------

// Init initialises the model.
func (m *LevelUpModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m *LevelUpModel) Update(msg tea.Msg) (*LevelUpModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Force quit
		if key.Matches(msg, m.keys.Quit) {
			m.quitting = true
			return m, tea.Quit
		}
		return m.handleKey(msg)
	}
	return m, nil
}

// handleKey dispatches key events based on current step.
func (m *LevelUpModel) handleKey(msg tea.KeyMsg) (*LevelUpModel, tea.Cmd) {
	switch m.currentStep {
	case LevelUpStepHP:
		return m.handleHPStepKey(msg)
	case LevelUpStepSubclass:
		return m.handleSubclassStepKey(msg)
	case LevelUpStepASI:
		return m.handleASIStepKey(msg)
	case LevelUpStepFeatures:
		return m.handleFeaturesStepKey(msg)
	case LevelUpStepSpellSlots:
		return m.handleSpellSlotsStepKey(msg)
	case LevelUpStepConfirm:
		return m.handleConfirmStepKey(msg)
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// Step navigation
// ---------------------------------------------------------------------------

func (m *LevelUpModel) advanceStep() {
	if m.stepIndex < len(m.steps)-1 {
		m.stepIndex++
		m.currentStep = m.steps[m.stepIndex]
	}
}

func (m *LevelUpModel) retreatStep() {
	if m.stepIndex > 0 {
		m.stepIndex--
		m.currentStep = m.steps[m.stepIndex]
	}
}

func (m *LevelUpModel) stepName(s LevelUpStep) string {
	switch s {
	case LevelUpStepHP:
		return "Hit Points"
	case LevelUpStepSubclass:
		return "Subclass"
	case LevelUpStepASI:
		return "Ability Score / Feat"
	case LevelUpStepFeatures:
		return "New Features"
	case LevelUpStepSpellSlots:
		return "Spell Slots"
	case LevelUpStepConfirm:
		return "Confirm"
	default:
		return "Unknown"
	}
}

// ---------------------------------------------------------------------------
// Ability constants list (ordered)
// ---------------------------------------------------------------------------

var abilityOrder = []models.Ability{
	models.AbilityStrength,
	models.AbilityDexterity,
	models.AbilityConstitution,
	models.AbilityIntelligence,
	models.AbilityWisdom,
	models.AbilityCharisma,
}

var abilityNames = []string{
	"Strength",
	"Dexterity",
	"Constitution",
	"Intelligence",
	"Wisdom",
	"Charisma",
}

// ---------------------------------------------------------------------------
// HP step handler
// ---------------------------------------------------------------------------

func (m *LevelUpModel) handleHPStepKey(msg tea.KeyMsg) (*LevelUpModel, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if !m.hpRolled && m.hpMethodCursor > 0 {
			m.hpMethodCursor--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if !m.hpRolled && m.hpMethodCursor < 1 {
			m.hpMethodCursor++
		}
		return m, nil

	case key.Matches(msg, m.keys.Select):
		if !m.hpRolled {
			// Lock in the HP method
			dieSize := m.parseDieSize()
			method := HPMethod(m.hpMethodCursor)
			switch method {
			case HPMethodRoll:
				m.hpRollResult = 1 + m.rng.Intn(dieSize)
			case HPMethodAverage:
				m.hpRollResult = (dieSize / 2) + 1
			}
			conMod := m.character.AbilityScores.Get(models.AbilityConstitution).Modifier()
			m.stagedHPIncrease = m.hpRollResult + conMod
			if m.stagedHPIncrease < 1 {
				m.stagedHPIncrease = 1
			}
			m.hpRolled = true
			return m, nil
		}
		// Already rolled — advance
		m.advanceStep()
		return m, nil

	case key.Matches(msg, m.keys.Back):
		if m.hpRolled {
			// Allow re-roll
			m.hpRolled = false
			m.hpRollResult = 0
			m.stagedHPIncrease = 0
			return m, nil
		}
		// Cancel level up
		return m, func() tea.Msg { return BackToSheetMsg{} }
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// Subclass step handler
// ---------------------------------------------------------------------------

func (m *LevelUpModel) handleSubclassStepKey(msg tea.KeyMsg) (*LevelUpModel, tea.Cmd) {
	if m.classData == nil {
		return m, nil
	}
	numSubclasses := len(m.classData.Subclasses)

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.subclassCursor > 0 {
			m.subclassCursor--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.subclassCursor < numSubclasses-1 {
			m.subclassCursor++
		}
		return m, nil

	case key.Matches(msg, m.keys.Select):
		if numSubclasses > 0 {
			sc := &m.classData.Subclasses[m.subclassCursor]
			m.stagedSubclass = sc
			// Collect subclass features at or below newLevel
			m.stagedSubclassFeats = nil
			for _, f := range sc.Features {
				if f.Level <= m.newLevel {
					m.stagedSubclassFeats = append(m.stagedSubclassFeats, f)
				}
			}
			m.advanceStep()
		}
		return m, nil

	case key.Matches(msg, m.keys.Back):
		m.retreatStep()
		return m, nil
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// ASI / Feat step handler
// ---------------------------------------------------------------------------

func (m *LevelUpModel) handleASIStepKey(msg tea.KeyMsg) (*LevelUpModel, tea.Cmd) {
	// Handle feat ASI ability prompt
	if m.inFeatASIPrompt {
		return m.handleFeatASIPromptKey(msg)
	}

	switch {
	case key.Matches(msg, m.keys.Tab):
		if m.asiMode == ASIModeASI {
			m.asiMode = ASIModeFeat
		} else {
			m.asiMode = ASIModeASI
		}
		// Reset selections when switching modes
		m.asiSelected = [6]bool{}
		m.stagedASIChanges = [6]int{}
		m.asiConfirmed = false
		m.stagedFeat = nil
		m.featASIAbility = ""
		return m, nil

	case key.Matches(msg, m.keys.Back):
		m.retreatStep()
		return m, nil
	}

	if m.asiMode == ASIModeASI {
		return m.handleASIModeKey(msg)
	}
	return m.handleFeatModeKey(msg)
}

func (m *LevelUpModel) handleASIModeKey(msg tea.KeyMsg) (*LevelUpModel, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.asiAbilityCursor > 0 {
			m.asiAbilityCursor--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.asiAbilityCursor < 5 {
			m.asiAbilityCursor++
		}
		return m, nil

	case key.Matches(msg, m.keys.Left):
		if m.asiPattern == ASIPatternPlus1Plus1 {
			m.asiPattern = ASIPatternPlus2
			m.asiSelected = [6]bool{}
			m.stagedASIChanges = [6]int{}
			m.asiConfirmed = false
		}
		return m, nil

	case key.Matches(msg, m.keys.Right):
		if m.asiPattern == ASIPatternPlus2 {
			m.asiPattern = ASIPatternPlus1Plus1
			m.asiSelected = [6]bool{}
			m.stagedASIChanges = [6]int{}
			m.asiConfirmed = false
		}
		return m, nil

	case key.Matches(msg, m.keys.Select):
		if m.asiConfirmed {
			m.advanceStep()
			return m, nil
		}

		idx := m.asiAbilityCursor
		ability := abilityOrder[idx]
		current := m.character.AbilityScores.Get(ability)

		// Check if already at 20
		if current.Base >= 20 {
			m.err = fmt.Errorf("%s is already at 20", abilityNames[idx])
			return m, nil
		}
		m.err = nil

		if m.asiPattern == ASIPatternPlus2 {
			// +2 to one ability
			if m.asiSelected[idx] {
				// Deselect
				m.asiSelected[idx] = false
				m.stagedASIChanges[idx] = 0
			} else {
				// Clear previous and select this one
				m.asiSelected = [6]bool{}
				m.stagedASIChanges = [6]int{}
				// Cap at 20
				delta := 2
				if current.Base+delta > 20 {
					delta = 20 - current.Base
				}
				m.asiSelected[idx] = true
				m.stagedASIChanges[idx] = delta
			}
			// Check if complete
			for _, s := range m.asiSelected {
				if s {
					m.asiConfirmed = true
					break
				}
			}
		} else {
			// +1/+1 pattern
			if m.asiSelected[idx] {
				// Deselect
				m.asiSelected[idx] = false
				m.stagedASIChanges[idx] = 0
				m.asiConfirmed = false
			} else {
				// Count currently selected
				count := 0
				for _, s := range m.asiSelected {
					if s {
						count++
					}
				}
				if count >= 2 {
					m.err = fmt.Errorf("already selected 2 abilities; deselect one first")
					return m, nil
				}
				delta := 1
				if current.Base+delta > 20 {
					delta = 20 - current.Base
				}
				m.asiSelected[idx] = true
				m.stagedASIChanges[idx] = delta
				count++
				if count == 2 {
					m.asiConfirmed = true
				}
			}
		}
		return m, nil
	}
	return m, nil
}

func (m *LevelUpModel) handleFeatModeKey(msg tea.KeyMsg) (*LevelUpModel, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.featCursor > 0 {
			m.featCursor--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.featCursor < len(m.filteredFeats)-1 {
			m.featCursor++
		}
		return m, nil

	case key.Matches(msg, m.keys.Select):
		if len(m.filteredFeats) > 0 && m.featCursor < len(m.filteredFeats) {
			feat := m.filteredFeats[m.featCursor]
			m.stagedFeat = &feat

			// If feat has ASI, prompt for ability
			if feat.Effects.AbilityScoreIncrease != nil && len(feat.Effects.AbilityScoreIncrease.Options) > 0 {
				m.inFeatASIPrompt = true
				m.featASICursor = 0
				return m, nil
			}

			// No ASI prompt needed, advance
			m.advanceStep()
		}
		return m, nil
	}

	// Handle text filtering for feat search
	if msg.Type == tea.KeyRunes {
		m.featFilterText += string(msg.Runes)
		m.rebuildFilteredFeats()
		return m, nil
	}
	if msg.Type == tea.KeyBackspace && len(m.featFilterText) > 0 {
		m.featFilterText = m.featFilterText[:len(m.featFilterText)-1]
		m.rebuildFilteredFeats()
		return m, nil
	}

	return m, nil
}

func (m *LevelUpModel) handleFeatASIPromptKey(msg tea.KeyMsg) (*LevelUpModel, tea.Cmd) {
	if m.stagedFeat == nil || m.stagedFeat.Effects.AbilityScoreIncrease == nil {
		m.inFeatASIPrompt = false
		return m, nil
	}
	options := m.stagedFeat.Effects.AbilityScoreIncrease.Options

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.featASICursor > 0 {
			m.featASICursor--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.featASICursor < len(options)-1 {
			m.featASICursor++
		}
		return m, nil

	case key.Matches(msg, m.keys.Select):
		if m.featASICursor < len(options) {
			m.featASIAbility = options[m.featASICursor]
			m.inFeatASIPrompt = false
			m.advanceStep()
		}
		return m, nil

	case key.Matches(msg, m.keys.Back):
		m.inFeatASIPrompt = false
		m.stagedFeat = nil
		return m, nil
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// Features step handler
// ---------------------------------------------------------------------------

func (m *LevelUpModel) handleFeaturesStepKey(msg tea.KeyMsg) (*LevelUpModel, tea.Cmd) {
	totalFeatures := len(m.stagedNewFeatures) + len(m.stagedSubclassFeats)

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.featureScrollOffset > 0 {
			m.featureScrollOffset--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.featureScrollOffset < totalFeatures-1 {
			m.featureScrollOffset++
		}
		return m, nil

	case key.Matches(msg, m.keys.Select):
		m.advanceStep()
		return m, nil

	case key.Matches(msg, m.keys.Back):
		m.retreatStep()
		return m, nil
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// Spell slots step handler
// ---------------------------------------------------------------------------

func (m *LevelUpModel) handleSpellSlotsStepKey(msg tea.KeyMsg) (*LevelUpModel, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Select):
		m.advanceStep()
		return m, nil

	case key.Matches(msg, m.keys.Back):
		m.retreatStep()
		return m, nil
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// Confirm step handler
// ---------------------------------------------------------------------------

func (m *LevelUpModel) handleConfirmStepKey(msg tea.KeyMsg) (*LevelUpModel, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Select):
		if m.errMsg != "" {
			// There was a load error; just go back
			return m, func() tea.Msg { return BackToSheetMsg{} }
		}
		m.applyLevelUp()
		return m, func() tea.Msg { return LevelUpCompleteMsg{} }

	case key.Matches(msg, m.keys.Back):
		if m.errMsg != "" {
			return m, func() tea.Msg { return BackToSheetMsg{} }
		}
		m.retreatStep()
		return m, nil
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// Apply level up
// ---------------------------------------------------------------------------

func (m *LevelUpModel) applyLevelUp() {
	char := m.character

	// 1. Increment level and update hit dice total
	char.LevelUp()

	// 2. HP increase
	char.CombatStats.HitPoints.Maximum += m.stagedHPIncrease
	char.CombatStats.HitPoints.Current += m.stagedHPIncrease

	// 3. Subclass
	if m.stagedSubclass != nil {
		char.Info.Subclass = m.stagedSubclass.Name
		for _, f := range m.stagedSubclassFeats {
			char.Features.AddClassFeature(
				f.Name,
				fmt.Sprintf("%s (%s)", char.Info.Class, m.stagedSubclass.Name),
				f.Description,
				f.Level,
			)
		}
	}

	// 4. ASI changes
	for i, delta := range m.stagedASIChanges {
		if delta > 0 {
			ability := abilityOrder[i]
			current := char.AbilityScores.Get(ability)
			newBase := current.Base + delta
			if newBase > 20 {
				newBase = 20
			}
			char.AbilityScores.SetBase(ability, newBase)
		}
	}

	// 5. Feat
	if m.stagedFeat != nil {
		char.Features.AddFeat(m.stagedFeat.Name, m.stagedFeat.Description)

		// Feat ASI
		if m.featASIAbility != "" && m.stagedFeat.Effects.AbilityScoreIncrease != nil {
			ability := models.Ability(m.featASIAbility)
			current := char.AbilityScores.Get(ability)
			amount := m.stagedFeat.Effects.AbilityScoreIncrease.Amount
			newBase := current.Base + amount
			if newBase > 20 {
				newBase = 20
			}
			char.AbilityScores.SetBase(ability, newBase)
		}

		// Feat effects: initiative, speed, AC
		effects := m.stagedFeat.Effects
		if effects.InitiativeBonus != 0 {
			char.CombatStats.Initiative += effects.InitiativeBonus
		}
		if effects.SpeedBonus != 0 {
			char.CombatStats.Speed += effects.SpeedBonus
		}
		if effects.ACBonus != 0 {
			char.CombatStats.ArmorClass += effects.ACBonus
		}

		// HPPerLevel: add hpPerLevel * newLevel to HP
		if effects.HPPerLevel > 0 {
			hpBonus := effects.HPPerLevel * m.newLevel
			char.CombatStats.HitPoints.Maximum += hpBonus
			char.CombatStats.HitPoints.Current += hpBonus
		}
	}

	// 6. Spell slots
	if m.stagedSpellSlots != nil && m.classData != nil && m.classData.Spellcaster {
		// Ensure Spellcasting struct exists
		if char.Spellcasting == nil {
			ability := models.Ability(strings.ToLower(m.classData.SpellcastingAbility))
			sc := models.NewSpellcasting(ability)
			char.Spellcasting = &sc
		}
		for lvl := 1; lvl <= 9; lvl++ {
			count := getSpellSlotCount(*m.stagedSpellSlots, lvl)
			if count > 0 {
				char.Spellcasting.SpellSlots.SetSlots(lvl, count)
			}
		}
	}

	// 7. Class features
	for _, f := range m.displayableFeaturesAtLevel(m.newLevel) {
		char.Features.AddClassFeature(
			f.Name,
			fmt.Sprintf("%s %d", char.Info.Class, m.newLevel),
			f.Description,
			m.newLevel,
		)
	}

	// 8. Mark updated and save
	char.MarkUpdated()
	if m.storage != nil {
		_ = m.storage.AutoSave(char)
	}
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

// View renders the level-up wizard.
func (m *LevelUpModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Styles
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	progressStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)

	// Title
	b.WriteString(titleStyle.Render(fmt.Sprintf("Level Up: Level %d → %d", m.oldLevel, m.newLevel)))
	b.WriteString("\n")

	// Progress
	b.WriteString(progressStyle.Render(fmt.Sprintf(
		"Step %d of %d: %s",
		m.stepIndex+1,
		len(m.steps),
		m.stepName(m.currentStep),
	)))
	b.WriteString("\n")
	b.WriteString("────────────────────────────────────────")
	b.WriteString("\n\n")

	// Load error
	if m.errMsg != "" {
		b.WriteString(errorStyle.Render(m.errMsg))
		b.WriteString("\n\n")
		b.WriteString("Press Enter or Esc to go back.")
		return b.String()
	}

	// Step-specific error
	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("⚠ %v", m.err)))
		b.WriteString("\n\n")
	}

	// Step content
	switch m.currentStep {
	case LevelUpStepHP:
		b.WriteString(m.viewHP())
	case LevelUpStepSubclass:
		b.WriteString(m.viewSubclass())
	case LevelUpStepASI:
		b.WriteString(m.viewASI())
	case LevelUpStepFeatures:
		b.WriteString(m.viewFeatures())
	case LevelUpStepSpellSlots:
		b.WriteString(m.viewSpellSlots())
	case LevelUpStepConfirm:
		b.WriteString(m.viewConfirm())
	}

	return b.String()
}

// ---------------------------------------------------------------------------
// View: HP step
// ---------------------------------------------------------------------------

func (m *LevelUpModel) viewHP() string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	valueStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))

	dieSize := m.parseDieSize()
	conMod := m.character.AbilityScores.Get(models.AbilityConstitution).Modifier()

	b.WriteString(headerStyle.Render("Choose HP Increase Method"))
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render(fmt.Sprintf("Hit Die: d%d  |  CON modifier: %+d", dieSize, conMod)))
	b.WriteString("\n\n")

	if !m.hpRolled {
		methods := []string{"Roll", "Average"}
		descriptions := []string{
			fmt.Sprintf("Roll 1d%d", dieSize),
			fmt.Sprintf("Take fixed value: %d", (dieSize/2)+1),
		}
		for i, method := range methods {
			prefix := "  "
			style := lipgloss.NewStyle()
			if i == m.hpMethodCursor {
				prefix = "▶ "
				style = selectedStyle
			}
			b.WriteString(prefix)
			b.WriteString(style.Render(method))
			b.WriteString(dimStyle.Render(fmt.Sprintf("  (%s)", descriptions[i])))
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("↑/↓: Navigate | Enter: Confirm | Esc: Cancel"))
	} else {
		methodName := "Roll"
		if HPMethod(m.hpMethodCursor) == HPMethodAverage {
			methodName = "Average"
		}
		b.WriteString(fmt.Sprintf("Method: %s\n", methodName))
		b.WriteString(fmt.Sprintf("Die result: %s\n", valueStyle.Render(fmt.Sprintf("%d", m.hpRollResult))))
		b.WriteString(fmt.Sprintf("CON modifier: %+d\n", conMod))
		b.WriteString(fmt.Sprintf("Total HP increase: %s\n", valueStyle.Render(fmt.Sprintf("%d", m.stagedHPIncrease))))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("Enter: Continue | Esc: Re-roll"))
	}

	return b.String()
}

// ---------------------------------------------------------------------------
// View: Subclass step
// ---------------------------------------------------------------------------

func (m *LevelUpModel) viewSubclass() string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	b.WriteString(headerStyle.Render("Choose Your Subclass"))
	b.WriteString("\n\n")

	if m.classData == nil || len(m.classData.Subclasses) == 0 {
		b.WriteString("No subclasses available.\n")
		return b.String()
	}

	for i, sc := range m.classData.Subclasses {
		prefix := "  "
		style := lipgloss.NewStyle()
		if i == m.subclassCursor {
			prefix = "▶ "
			style = selectedStyle
		}
		b.WriteString(prefix)
		b.WriteString(style.Render(sc.Name))
		b.WriteString("\n")
	}

	// Description panel for highlighted subclass
	b.WriteString("\n")
	b.WriteString("────────────────────────────────────────")
	b.WriteString("\n\n")

	highlighted := m.classData.Subclasses[m.subclassCursor]
	if highlighted.Description != "" {
		wrapped := wordWrap(highlighted.Description, m.descWidth(0), "")
		b.WriteString(descStyle.Render(wrapped))
		b.WriteString("\n\n")
	}

	// Show features at newLevel for highlighted subclass
	var levelFeatures []data.Feature
	for _, f := range highlighted.Features {
		if f.Level <= m.newLevel {
			levelFeatures = append(levelFeatures, f)
		}
	}
	if len(levelFeatures) > 0 {
		b.WriteString(headerStyle.Render("Features:"))
		b.WriteString("\n")
		for _, f := range levelFeatures {
			b.WriteString(fmt.Sprintf("  • %s (Level %d)\n", f.Name, f.Level))
			if f.Description != "" {
				wrapped := wordWrap(f.Description, m.descWidth(4), "    ")
				b.WriteString(dimStyle.Render(fmt.Sprintf("    %s", wrapped)))
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("↑/↓: Navigate | Enter: Select | Esc: Back"))

	return b.String()
}

// ---------------------------------------------------------------------------
// View: ASI step
// ---------------------------------------------------------------------------

func (m *LevelUpModel) viewASI() string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	disabledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	activeTabStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	inactiveTabStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	valueStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))

	// Tab bar
	asiTab := " ASI "
	featTab := " Feat "
	if m.asiMode == ASIModeASI {
		asiTab = activeTabStyle.Render("[ASI]")
		featTab = inactiveTabStyle.Render(" Feat ")
	} else {
		asiTab = inactiveTabStyle.Render(" ASI ")
		featTab = activeTabStyle.Render("[Feat]")
	}
	b.WriteString(asiTab + "  " + featTab)
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("Tab: switch mode"))
	b.WriteString("\n\n")

	if m.inFeatASIPrompt {
		return b.String() + m.viewFeatASIPrompt()
	}

	if m.asiMode == ASIModeASI {
		b.WriteString(m.viewASIMode(headerStyle, selectedStyle, dimStyle, disabledStyle, valueStyle))
	} else {
		b.WriteString(m.viewFeatMode(headerStyle, selectedStyle, dimStyle))
	}

	return b.String()
}

func (m *LevelUpModel) viewASIMode(headerStyle, selectedStyle, dimStyle, disabledStyle, valueStyle lipgloss.Style) string {
	var b strings.Builder

	// Pattern selector
	patternLabel := "+2 to one ability"
	if m.asiPattern == ASIPatternPlus1Plus1 {
		patternLabel = "+1 to two abilities"
	}
	b.WriteString(headerStyle.Render("Ability Score Improvement"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Pattern: %s", valueStyle.Render(patternLabel)))
	b.WriteString("  ")
	b.WriteString(dimStyle.Render("(←/→ to change)"))
	b.WriteString("\n\n")

	for i, name := range abilityNames {
		ability := abilityOrder[i]
		current := m.character.AbilityScores.Get(ability)
		base := current.Base
		atMax := base >= 20

		prefix := "  "
		style := lipgloss.NewStyle()
		if i == m.asiAbilityCursor {
			prefix = "▶ "
			style = selectedStyle
		}

		if atMax {
			style = disabledStyle
		}

		line := fmt.Sprintf("%-14s  %2d", name, base)
		if m.asiSelected[i] {
			line += valueStyle.Render(fmt.Sprintf("  (+%d → %d)", m.stagedASIChanges[i], base+m.stagedASIChanges[i]))
		} else if atMax {
			line += dimStyle.Render("  (max)")
		}

		b.WriteString(prefix)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if m.asiConfirmed {
		b.WriteString(valueStyle.Render("✓ Selection confirmed"))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("Enter: Continue | Esc: Back"))
	} else {
		b.WriteString(dimStyle.Render("↑/↓: Navigate | Enter: Toggle | ←/→: Pattern | Esc: Back"))
	}

	return b.String()
}

func (m *LevelUpModel) viewFeatMode(headerStyle, selectedStyle, dimStyle lipgloss.Style) string {
	var b strings.Builder

	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	b.WriteString(headerStyle.Render("Choose a Feat"))
	b.WriteString("\n")

	// Search filter
	filterDisplay := m.featFilterText
	if filterDisplay == "" {
		filterDisplay = dimStyle.Render("(type to filter)")
	}
	b.WriteString(fmt.Sprintf("Filter: %s█", filterDisplay))
	b.WriteString("\n\n")

	if len(m.filteredFeats) == 0 {
		b.WriteString(dimStyle.Render("No feats match the filter."))
		b.WriteString("\n")
	} else {
		// Show a window of feats around the cursor
		windowSize := 10
		start := m.featCursor - windowSize/2
		if start < 0 {
			start = 0
		}
		end := start + windowSize
		if end > len(m.filteredFeats) {
			end = len(m.filteredFeats)
			start = end - windowSize
			if start < 0 {
				start = 0
			}
		}

		for i := start; i < end; i++ {
			feat := m.filteredFeats[i]
			prefix := "  "
			style := lipgloss.NewStyle()
			if i == m.featCursor {
				prefix = "▶ "
				style = selectedStyle
			}
			label := feat.Name
			if feat.Category != "" {
				label += dimStyle.Render(fmt.Sprintf(" [%s]", feat.Category))
			}
			b.WriteString(prefix)
			b.WriteString(style.Render(label))
			b.WriteString("\n")
		}

		// Description panel for highlighted feat
		if m.featCursor < len(m.filteredFeats) {
			feat := m.filteredFeats[m.featCursor]
			b.WriteString("\n")
			b.WriteString("────────────────────────────────────────")
			b.WriteString("\n")
			if feat.Prerequisite != "" {
				b.WriteString(dimStyle.Render(fmt.Sprintf("Prerequisite: %s", feat.Prerequisite)))
				b.WriteString("\n")
			}
			// Show word-wrapped description
			wrapped := wordWrap(feat.Description, m.descWidth(0), "")
			b.WriteString(descStyle.Render(wrapped))
			b.WriteString("\n")

			// Show effects
			if feat.Effects.AbilityScoreIncrease != nil {
				asi := feat.Effects.AbilityScoreIncrease
				b.WriteString(dimStyle.Render(fmt.Sprintf(
					"ASI: +%d to %s",
					asi.Amount,
					strings.Join(asi.Options, " or "),
				)))
				b.WriteString("\n")
			}
			if feat.Effects.HPPerLevel > 0 {
				b.WriteString(dimStyle.Render(fmt.Sprintf("HP: +%d per level", feat.Effects.HPPerLevel)))
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("↑/↓: Navigate | Enter: Select | Type: Filter | Esc: Back"))

	return b.String()
}

func (m *LevelUpModel) viewFeatASIPrompt() string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	b.WriteString(headerStyle.Render(fmt.Sprintf("Choose ability for %s", m.stagedFeat.Name)))
	b.WriteString("\n\n")

	options := m.stagedFeat.Effects.AbilityScoreIncrease.Options
	amount := m.stagedFeat.Effects.AbilityScoreIncrease.Amount

	for i, opt := range options {
		prefix := "  "
		style := lipgloss.NewStyle()
		if i == m.featASICursor {
			prefix = "▶ "
			style = selectedStyle
		}
		// Show current value
		ability := models.Ability(opt)
		current := m.character.AbilityScores.Get(ability)
		line := fmt.Sprintf("%-14s  %d → %d", opt, current.Base, current.Base+amount)
		b.WriteString(prefix)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("↑/↓: Navigate | Enter: Select | Esc: Cancel"))

	return b.String()
}

// ---------------------------------------------------------------------------
// View: Features step
// ---------------------------------------------------------------------------

func (m *LevelUpModel) viewFeatures() string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	b.WriteString(headerStyle.Render("New Features"))
	b.WriteString("\n\n")

	// Combine class features and subclass features
	var allFeatures []struct {
		name, desc, source string
		level              int
	}

	displayable := m.displayableFeaturesAtLevel(m.newLevel)
	for _, f := range displayable {
		allFeatures = append(allFeatures, struct {
			name, desc, source string
			level              int
		}{f.Name, f.Description, m.character.Info.Class, f.Level})
	}

	for _, f := range m.stagedSubclassFeats {
		source := ""
		if m.stagedSubclass != nil {
			source = m.stagedSubclass.Name
		}
		allFeatures = append(allFeatures, struct {
			name, desc, source string
			level              int
		}{f.Name, f.Description, source, f.Level})
	}

	if len(allFeatures) == 0 {
		b.WriteString(dimStyle.Render("No new features at this level."))
		b.WriteString("\n")
	} else {
		for i, f := range allFeatures {
			marker := "  "
			if i == m.featureScrollOffset {
				marker = "▶ "
			}
			b.WriteString(marker)
			b.WriteString(nameStyle.Render(f.name))
			if f.source != "" {
				b.WriteString(dimStyle.Render(fmt.Sprintf(" (%s, Level %d)", f.source, f.level)))
			}
			b.WriteString("\n")
			if f.desc != "" {
				wrapped := wordWrap(f.desc, m.descWidth(4), "    ")
				b.WriteString("    ")
				b.WriteString(descStyle.Render(wrapped))
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
	}

	b.WriteString(dimStyle.Render("↑/↓: Scroll | Enter: Continue | Esc: Back"))

	return b.String()
}

// ---------------------------------------------------------------------------
// View: Spell slots step
// ---------------------------------------------------------------------------

func (m *LevelUpModel) viewSpellSlots() string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	valueStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	changeStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13"))

	b.WriteString(headerStyle.Render("Spell Slot Changes"))
	b.WriteString("\n\n")

	levelNames := []string{"", "1st", "2nd", "3rd", "4th", "5th", "6th", "7th", "8th", "9th"}

	// Header
	b.WriteString(fmt.Sprintf("  %-8s  %5s  %5s\n", "Level", "Old", "New"))
	b.WriteString("  ────────  ─────  ─────\n")

	hasChanges := false
	for lvl := 1; lvl <= 9; lvl++ {
		oldCount := 0
		newCount := 0
		if m.oldSpellSlots != nil {
			oldCount = getSpellSlotCount(*m.oldSpellSlots, lvl)
		}
		if m.stagedSpellSlots != nil {
			newCount = getSpellSlotCount(*m.stagedSpellSlots, lvl)
		}

		// Only show levels that have slots (or will have them)
		if oldCount == 0 && newCount == 0 {
			continue
		}

		changed := oldCount != newCount
		if changed {
			hasChanges = true
		}

		oldStr := fmt.Sprintf("%d", oldCount)
		newStr := fmt.Sprintf("%d", newCount)

		if changed {
			b.WriteString(fmt.Sprintf("  %-8s  %5s  %5s",
				levelNames[lvl],
				oldStr,
				changeStyle.Render(newStr),
			))
			b.WriteString(valueStyle.Render(fmt.Sprintf("  (+%d)", newCount-oldCount)))
		} else {
			b.WriteString(fmt.Sprintf("  %-8s  %5s  %5s",
				levelNames[lvl],
				oldStr,
				dimStyle.Render(newStr),
			))
		}
		b.WriteString("\n")
	}

	if !hasChanges {
		b.WriteString(dimStyle.Render("  No spell slot changes at this level."))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("Enter: Continue | Esc: Back"))

	return b.String()
}

// ---------------------------------------------------------------------------
// View: Confirm step
// ---------------------------------------------------------------------------

func (m *LevelUpModel) viewConfirm() string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	valueStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	b.WriteString(headerStyle.Render("Confirm Level Up"))
	b.WriteString("\n\n")

	// Level
	b.WriteString(labelStyle.Render("Level: "))
	b.WriteString(valueStyle.Render(fmt.Sprintf("%d → %d", m.oldLevel, m.newLevel)))
	b.WriteString("\n\n")

	// HP
	b.WriteString(labelStyle.Render("HP Increase: "))
	b.WriteString(valueStyle.Render(fmt.Sprintf("+%d", m.stagedHPIncrease)))
	currentMax := m.character.CombatStats.HitPoints.Maximum
	b.WriteString(dimStyle.Render(fmt.Sprintf(" (max HP: %d → %d)", currentMax, currentMax+m.stagedHPIncrease)))
	b.WriteString("\n")

	// Subclass
	if m.stagedSubclass != nil {
		b.WriteString(labelStyle.Render("Subclass: "))
		b.WriteString(valueStyle.Render(m.stagedSubclass.Name))
		b.WriteString("\n")
		if len(m.stagedSubclassFeats) > 0 {
			for _, f := range m.stagedSubclassFeats {
				b.WriteString(dimStyle.Render(fmt.Sprintf("  • %s (Level %d)", f.Name, f.Level)))
				b.WriteString("\n")
			}
		}
	}

	// ASI
	hasASI := false
	for i, delta := range m.stagedASIChanges {
		if delta > 0 {
			if !hasASI {
				b.WriteString(labelStyle.Render("ASI: "))
				hasASI = true
			} else {
				b.WriteString(", ")
			}
			ability := abilityOrder[i]
			current := m.character.AbilityScores.Get(ability)
			b.WriteString(valueStyle.Render(fmt.Sprintf(
				"%s %d → %d",
				abilityNames[i],
				current.Base,
				current.Base+delta,
			)))
		}
	}
	if hasASI {
		b.WriteString("\n")
	}

	// Feat
	if m.stagedFeat != nil {
		b.WriteString(labelStyle.Render("Feat: "))
		b.WriteString(valueStyle.Render(m.stagedFeat.Name))
		if m.featASIAbility != "" {
			b.WriteString(dimStyle.Render(fmt.Sprintf(" (+%d %s)",
				m.stagedFeat.Effects.AbilityScoreIncrease.Amount,
				m.featASIAbility,
			)))
		}
		b.WriteString("\n")
		if m.stagedFeat.Effects.HPPerLevel > 0 {
			b.WriteString(dimStyle.Render(fmt.Sprintf(
				"  +%d HP per level (total: +%d)",
				m.stagedFeat.Effects.HPPerLevel,
				m.stagedFeat.Effects.HPPerLevel*m.newLevel,
			)))
			b.WriteString("\n")
		}
	}

	// New features
	displayable := m.displayableFeaturesAtLevel(m.newLevel)
	if len(displayable) > 0 {
		b.WriteString(labelStyle.Render("New Features: "))
		b.WriteString("\n")
		for _, f := range displayable {
			b.WriteString(dimStyle.Render(fmt.Sprintf("  • %s", f.Name)))
			b.WriteString("\n")
		}
	}

	// Spell slots
	if m.stagedSpellSlots != nil && m.classData != nil && m.classData.Spellcaster {
		hasChanges := false
		var slotChanges []string
		levelNames := []string{"", "1st", "2nd", "3rd", "4th", "5th", "6th", "7th", "8th", "9th"}
		for lvl := 1; lvl <= 9; lvl++ {
			oldCount := 0
			if m.oldSpellSlots != nil {
				oldCount = getSpellSlotCount(*m.oldSpellSlots, lvl)
			}
			newCount := getSpellSlotCount(*m.stagedSpellSlots, lvl)
			if newCount != oldCount {
				hasChanges = true
				slotChanges = append(slotChanges, fmt.Sprintf("%s: %d→%d", levelNames[lvl], oldCount, newCount))
			}
		}
		if hasChanges {
			b.WriteString(labelStyle.Render("Spell Slots: "))
			b.WriteString(valueStyle.Render(strings.Join(slotChanges, ", ")))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(headerStyle.Render("Press Enter to apply changes"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("Esc: Go back"))

	return b.String()
}

// ---------------------------------------------------------------------------
// Message types (BackToSheetMsg is in main_sheet.go)
// ---------------------------------------------------------------------------

// BackToSheetMsg is sent when the user wants to go back to the main sheet.
// (This may already be defined in main_sheet.go — kept here as a safeguard.
//  If there's a compile error due to duplicate, remove this one.)
// type BackToSheetMsg struct{}
