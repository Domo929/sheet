package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MainSheetModel is the model for the main character sheet view.
type MainSheetModel struct {
	character     *models.Character
	storage       *storage.CharacterStorage
	spellDatabase *data.SpellDatabase
	width         int
	height        int
	focusArea     FocusArea
	keys          mainSheetKeyMap
	statusMessage string

	// Quit confirmation
	confirmingQuit bool

	// HP input mode
	hpInputMode   HPInputMode
	hpInputBuffer string

	// Condition selection mode
	conditionMode    bool
	conditionCursor  int
	conditionAdding  bool // true = adding, false = removing

	// Action type selection (for Actions panel)
	selectedActionType ActionType
	actionCursor       int // Cursor for selecting actions within the current type

	// Spell casting modal
	castingSpell        *data.SpellData // Spell being cast
	castLevelCursor     int             // Cursor for slot level selection
	availableCastLevels []int           // Available slot levels for upcasting

	// Rest mode
	restMode       RestMode
	restHitDice    int    // Number of hit dice to spend on short rest
	restConfirming bool   // Confirming long rest
	restResult     string // Result summary to display
}

// HPInputMode represents the current HP modification mode.
type HPInputMode int

const (
	HPInputNone HPInputMode = iota
	HPInputDamage
	HPInputHeal
	HPInputTemp
)

// RestMode represents the current rest interaction mode.
type RestMode int

const (
	RestModeNone   RestMode = iota
	RestModeMenu            // Choosing between short/long rest
	RestModeShort           // Short rest - spending hit dice
	RestModeLong            // Long rest - confirming
	RestModeResult          // Showing rest results
)

// FocusArea represents which panel is currently focused.
type FocusArea int

const (
	FocusAbilitiesAndSaves FocusArea = iota
	FocusSkills
	FocusCombat
	FocusActions
)

const numFocusAreas = 4

// ActionType represents the type of action being viewed.
type ActionType int

const (
	ActionTypeAction ActionType = iota
	ActionTypeBonus
	ActionTypeReaction
	ActionTypeOther
)

const numActionTypes = 4

// ActionItemType represents different types of action items
type ActionItemType int

const (
	ActionItemWeapon ActionItemType = iota
	ActionItemSpell
	ActionItemStandard
)

// ActionItem represents a selectable action in the Actions panel
type ActionItem struct {
	Type        ActionItemType
	Name        string
	Description string
	// For weapons
	Weapon *models.Item
	// For spells
	Spell *data.SpellData
	// For standard actions
	StandardAction *StandardAction
}

// StandardAction represents a standard D&D 2024 action.
type StandardAction struct {
	Name        string
	Description string
	ActionType  ActionType
}

// Standard D&D 2024 actions
var standardActions = []StandardAction{
	// Actions
	{Name: "Attack", Description: "Make one attack with a weapon or Unarmed Strike", ActionType: ActionTypeAction},
	{Name: "Dash", Description: "Gain extra movement equal to your Speed", ActionType: ActionTypeAction},
	{Name: "Disengage", Description: "Your movement doesn't provoke Opportunity Attacks", ActionType: ActionTypeAction},
	{Name: "Dodge", Description: "Attacks against you have Disadvantage; DEX saves have Advantage", ActionType: ActionTypeAction},
	{Name: "Grapple", Description: "Grab a creature (Athletics vs Athletics/Acrobatics)", ActionType: ActionTypeAction},
	{Name: "Help", Description: "Give an ally Advantage on their next ability check or attack", ActionType: ActionTypeAction},
	{Name: "Hide", Description: "Make a Stealth check to become Hidden", ActionType: ActionTypeAction},
	{Name: "Influence", Description: "Make a Charisma check to alter a creature's attitude", ActionType: ActionTypeAction},
	{Name: "Magic", Description: "Cast a spell, use a magic item, or use a magical feature", ActionType: ActionTypeAction},
	{Name: "Ready", Description: "Prepare to take an action in response to a trigger", ActionType: ActionTypeAction},
	{Name: "Search", Description: "Make a Perception or Investigation check", ActionType: ActionTypeAction},
	{Name: "Shove", Description: "Push a creature 5 feet or knock it Prone", ActionType: ActionTypeAction},
	{Name: "Study", Description: "Make an Intelligence check to recall information", ActionType: ActionTypeAction},
	{Name: "Utilize", Description: "Use a nonmagical object", ActionType: ActionTypeAction},
	// Bonus Actions
	{Name: "Offhand Attack", Description: "Attack with a Light weapon in your other hand (no ability mod to damage)", ActionType: ActionTypeBonus},
	// Reactions
	{Name: "Opportunity Attack", Description: "Make one melee attack when a creature leaves your reach", ActionType: ActionTypeReaction},
}

// All D&D 5e conditions for selection
var allConditions = []models.Condition{
	models.ConditionBlinded,
	models.ConditionCharmed,
	models.ConditionDeafened,
	models.ConditionExhaustion,
	models.ConditionFrightened,
	models.ConditionGrappled,
	models.ConditionIncapacitated,
	models.ConditionInvisible,
	models.ConditionParalyzed,
	models.ConditionPetrified,
	models.ConditionPoisoned,
	models.ConditionProne,
	models.ConditionRestrained,
	models.ConditionStunned,
	models.ConditionUnconscious,
}

type mainSheetKeyMap struct {
	Quit           key.Binding
	ForceQuit      key.Binding
	Tab            key.Binding
	ShiftTab       key.Binding
	Inventory      key.Binding
	Spellbook      key.Binding
	Info           key.Binding
	Combat         key.Binding
	Rest           key.Binding
	Navigation     key.Binding
	Damage         key.Binding
	Heal           key.Binding
	TempHP         key.Binding
	DeathSuccess   key.Binding
	DeathFail      key.Binding
	DeathReset     key.Binding
	AddCondition   key.Binding
	RemCondition   key.Binding
	NextActionType key.Binding
	PrevActionType key.Binding
}

func defaultMainSheetKeyMap() mainSheetKeyMap {
	return mainSheetKeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "force quit"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next panel"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev panel"),
		),
		Inventory: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "inventory"),
		),
		Spellbook: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "spellbook"),
		),
		Info: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "character info"),
		),
		Combat: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "combat"),
		),
		Rest: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rest"),
		),
		Navigation: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back to selection"),
		),
		Damage: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "take damage"),
		),
		Heal: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "heal"),
		),
		TempHP: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "temp HP"),
		),
		DeathSuccess: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "death save success"),
		),
		DeathFail: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "death save failure"),
		),
		DeathReset: key.NewBinding(
			key.WithKeys("0"),
			key.WithHelp("0", "reset death saves"),
		),
		AddCondition: key.NewBinding(
			key.WithKeys("+", "="),
			key.WithHelp("+", "add condition"),
		),
		RemCondition: key.NewBinding(
			key.WithKeys("-", "_"),
			key.WithHelp("-", "remove condition"),
		),
		NextActionType: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→", "next action type"),
		),
		PrevActionType: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←", "prev action type"),
		),
	}
}

// NewMainSheetModel creates a new main sheet model.
func NewMainSheetModel(character *models.Character, storage *storage.CharacterStorage) *MainSheetModel {
	// Load spell database
	loader := data.NewLoader("./data")
	spellDB, _ := loader.GetSpells() // Ignore error for now, spells optional

	return &MainSheetModel{
		character:     character,
		storage:       storage,
		spellDatabase: spellDB,
		focusArea:     FocusAbilitiesAndSaves,
		keys:          defaultMainSheetKeyMap(),
	}
}

// Init initializes the model.
func (m *MainSheetModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m *MainSheetModel) Update(msg tea.Msg) (*MainSheetModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle quit confirmation
		if m.confirmingQuit {
			switch msg.String() {
			case "y", "Y", "enter":
				return m, tea.Quit
			default:
				m.confirmingQuit = false
				m.statusMessage = ""
				return m, nil
			}
		}

		// Ctrl+C always quits immediately
		if key.Matches(msg, m.keys.ForceQuit) {
			return m, tea.Quit
		}

		// Handle HP input mode
		if m.hpInputMode != HPInputNone {
			return m.handleHPInput(msg)
		}

		// Handle condition selection mode
		if m.conditionMode {
			return m.handleConditionInput(msg)
		}

		// Handle spell casting modal
		if m.castingSpell != nil {
			return m.handleCastingInput(msg)
		}

		// Handle rest mode
		if m.restMode != RestModeNone {
			return m.handleRestInput(msg)
		}

		// Handle action navigation when Actions are focused
		if m.focusArea == FocusActions {
			switch msg.Type {
			case tea.KeyUp:
				if m.actionCursor > 0 {
					m.actionCursor--
				}
				return m, nil
			case tea.KeyDown:
				actionItems := m.getActionItems()
				if m.actionCursor < len(actionItems)-1 {
					m.actionCursor++
				}
				return m, nil
			case tea.KeyEnter:
				return m.handleActionSelection()
			}
		}

		// Clear status message on any key press
		m.statusMessage = ""

		switch {
		case key.Matches(msg, m.keys.Quit):
			m.confirmingQuit = true
			m.statusMessage = "Quit? (y/n)"
			return m, nil
		case key.Matches(msg, m.keys.Tab):
			m.focusArea = (m.focusArea + 1) % numFocusAreas
			return m, nil
		case key.Matches(msg, m.keys.ShiftTab):
			if m.focusArea == 0 {
				m.focusArea = numFocusAreas - 1
			} else {
				m.focusArea--
			}
			return m, nil
		case key.Matches(msg, m.keys.Navigation):
			return m, func() tea.Msg { return BackToSelectionMsg{} }
		case key.Matches(msg, m.keys.Inventory):
			return m, func() tea.Msg { return OpenInventoryMsg{} }
		case key.Matches(msg, m.keys.Spellbook):
			return m, func() tea.Msg { return OpenSpellbookMsg{} }
		case key.Matches(msg, m.keys.Info):
			m.statusMessage = "Character info view coming soon..."
			return m, nil
		case key.Matches(msg, m.keys.Combat):
			m.statusMessage = "Combat tracker coming soon..."
			return m, nil
		case key.Matches(msg, m.keys.Rest):
			m.restMode = RestModeMenu
			return m, nil
		case key.Matches(msg, m.keys.Damage):
			if m.focusArea == FocusCombat {
				m.hpInputMode = HPInputDamage
				m.hpInputBuffer = ""
				return m, nil
			}
		case key.Matches(msg, m.keys.Heal):
			if m.focusArea == FocusCombat {
				m.hpInputMode = HPInputHeal
				m.hpInputBuffer = ""
				return m, nil
			}
		case key.Matches(msg, m.keys.TempHP):
			if m.focusArea == FocusCombat {
				m.hpInputMode = HPInputTemp
				m.hpInputBuffer = ""
				return m, nil
			}
		case key.Matches(msg, m.keys.DeathSuccess):
			if m.focusArea == FocusCombat && m.character.CombatStats.HitPoints.Current == 0 {
				if m.character.CombatStats.DeathSaves.Successes < 3 {
					m.character.CombatStats.DeathSaves.Successes++
					if m.character.CombatStats.DeathSaves.Successes >= 3 {
						m.statusMessage = "Stabilized!"
					} else {
						m.statusMessage = fmt.Sprintf("Death save success (%d/3)", m.character.CombatStats.DeathSaves.Successes)
					}
					m.saveCharacter()
				}
				return m, nil
			}
		case key.Matches(msg, m.keys.DeathFail):
			if m.focusArea == FocusCombat && m.character.CombatStats.HitPoints.Current == 0 {
				if m.character.CombatStats.DeathSaves.Failures < 3 {
					m.character.CombatStats.DeathSaves.Failures++
					if m.character.CombatStats.DeathSaves.Failures >= 3 {
						m.statusMessage = "Character has died!"
					} else {
						m.statusMessage = fmt.Sprintf("Death save failure (%d/3)", m.character.CombatStats.DeathSaves.Failures)
					}
					m.saveCharacter()
				}
				return m, nil
			}
		case key.Matches(msg, m.keys.DeathReset):
			if m.focusArea == FocusCombat {
				m.character.CombatStats.DeathSaves.Reset()
				m.statusMessage = "Death saves reset"
				m.saveCharacter()
				return m, nil
			}
		case key.Matches(msg, m.keys.AddCondition):
			if m.focusArea == FocusCombat {
				m.conditionMode = true
				m.conditionAdding = true
				m.conditionCursor = 0
				return m, nil
			}
		case key.Matches(msg, m.keys.RemCondition):
			if m.focusArea == FocusCombat && len(m.character.CombatStats.Conditions) > 0 {
				m.conditionMode = true
				m.conditionAdding = false
				m.conditionCursor = 0
				return m, nil
			}
		case key.Matches(msg, m.keys.NextActionType):
			if m.focusArea == FocusActions {
				m.selectedActionType = (m.selectedActionType + 1) % numActionTypes
				m.actionCursor = 0 // Reset cursor when changing tabs
				return m, nil
			}
		case key.Matches(msg, m.keys.PrevActionType):
			if m.focusArea == FocusActions {
				if m.selectedActionType == 0 {
					m.selectedActionType = numActionTypes - 1
				} else {
					m.selectedActionType--
				}
				m.actionCursor = 0 // Reset cursor when changing tabs
				return m, nil
			}
		}
	}

	return m, nil
}

// saveCharacter saves the character if storage is available.
func (m *MainSheetModel) saveCharacter() {
	if m.storage != nil {
		_, _ = m.storage.Save(m.character)
	}
}

// handleActionSelection handles when user presses Enter on an action
func (m *MainSheetModel) handleActionSelection() (*MainSheetModel, tea.Cmd) {
	actionItems := m.getActionItems()
	if m.actionCursor < 0 || m.actionCursor >= len(actionItems) {
		return m, nil
	}

	selectedItem := actionItems[m.actionCursor]

	switch selectedItem.Type {
	case ActionItemSpell:
		// Initiate spell casting modal
		if selectedItem.Spell != nil {
			spell := selectedItem.Spell

			// Check if character can cast spells
			if m.character.Spellcasting == nil {
				m.statusMessage = "This character cannot cast spells"
				return m, nil
			}

			// For cantrips, open modal immediately
			if spell.Level == 0 {
				m.castingSpell = spell
				m.availableCastLevels = nil
				m.castLevelCursor = 0
				return m, nil
			}

			// For leveled spells, check available slots
			availableLevels := m.getAvailableCastLevels(spell.Level)
			if len(availableLevels) == 0 {
				// Check if can cast as ritual
				if spell.Ritual {
					// Find the known spell to check ritual flag
					for _, ks := range m.character.Spellcasting.KnownSpells {
						if ks.Name == spell.Name && ks.Ritual {
							m.castingSpell = spell
							m.availableCastLevels = nil
							m.castLevelCursor = 0
							return m, nil
						}
					}
				}
				m.statusMessage = fmt.Sprintf("No spell slots available for %s", spell.Name)
				return m, nil
			}

			// Open casting modal with available slot levels
			m.castingSpell = spell
			m.availableCastLevels = availableLevels
			m.castLevelCursor = 0
			return m, nil
		}
		return m, nil

	case ActionItemWeapon:
		// For weapons, show detailed attack info
		if selectedItem.Weapon != nil {
			w := selectedItem.Weapon
			attackBonus := m.getWeaponAttackBonus(*w)
			damageMod := m.getWeaponDamageMod(*w)

			damageStr := w.Damage
			if damageMod != 0 {
				damageStr = fmt.Sprintf("%s%s", w.Damage, formatModifier(damageMod))
			}
			damageStr += " " + w.DamageType

			m.statusMessage = fmt.Sprintf("⚔ %s: Hit %s, Dmg %s", w.Name, formatModifier(attackBonus), damageStr)
		} else {
			// Unarmed strike
			m.statusMessage = fmt.Sprintf("⚔ Unarmed Strike: Hit %s, Dmg 1 bludgeoning",
				formatModifier(m.character.AbilityScores.Strength.Modifier()))
		}
		return m, nil

	case ActionItemStandard:
		// For standard actions, show the description
		if selectedItem.StandardAction != nil {
			m.statusMessage = fmt.Sprintf("📋 %s: %s", selectedItem.Name, selectedItem.StandardAction.Description)
		}
		return m, nil
	}

	return m, nil
}

// handleHPInput handles keyboard input when in HP modification mode.
func (m *MainSheetModel) handleHPInput(msg tea.KeyMsg) (*MainSheetModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.hpInputMode = HPInputNone
		m.hpInputBuffer = ""
		return m, nil

	case tea.KeyEnter:
		if m.hpInputBuffer == "" {
			m.hpInputMode = HPInputNone
			return m, nil
		}
		
		amount, err := strconv.Atoi(m.hpInputBuffer)
		if err != nil || amount < 0 {
			m.statusMessage = "Invalid number"
			m.hpInputMode = HPInputNone
			m.hpInputBuffer = ""
			return m, nil
		}

		// Apply the HP change
		switch m.hpInputMode {
		case HPInputDamage:
			m.character.CombatStats.HitPoints.TakeDamage(amount)
			m.statusMessage = fmt.Sprintf("Took %d damage", amount)
		case HPInputHeal:
			m.character.CombatStats.HitPoints.Heal(amount)
			m.statusMessage = fmt.Sprintf("Healed %d HP", amount)
		case HPInputTemp:
			m.character.CombatStats.HitPoints.AddTemporaryHP(amount)
			m.statusMessage = fmt.Sprintf("Gained %d temp HP", amount)
		}

		// Save character
		if m.storage != nil {
			_, _ = m.storage.Save(m.character)
		}

		m.hpInputMode = HPInputNone
		m.hpInputBuffer = ""
		return m, nil

	case tea.KeyBackspace:
		if len(m.hpInputBuffer) > 0 {
			m.hpInputBuffer = m.hpInputBuffer[:len(m.hpInputBuffer)-1]
		}
		return m, nil

	case tea.KeyRunes:
		// Only accept digits
		for _, r := range msg.Runes {
			if r >= '0' && r <= '9' {
				m.hpInputBuffer += string(r)
			}
		}
		return m, nil
	}

	return m, nil
}

// handleConditionInput handles keyboard input when in condition selection mode.
func (m *MainSheetModel) handleConditionInput(msg tea.KeyMsg) (*MainSheetModel, tea.Cmd) {
	var listLen int
	if m.conditionAdding {
		listLen = len(allConditions)
	} else {
		listLen = len(m.character.CombatStats.Conditions)
	}

	switch msg.Type {
	case tea.KeyEscape:
		m.conditionMode = false
		return m, nil

	case tea.KeyEnter:
		if m.conditionAdding {
			// Add the selected condition
			cond := allConditions[m.conditionCursor]
			m.character.CombatStats.AddCondition(cond)
			m.statusMessage = fmt.Sprintf("Added %s", cond)
		} else {
			// Remove the selected condition
			if len(m.character.CombatStats.Conditions) > 0 {
				cond := m.character.CombatStats.Conditions[m.conditionCursor]
				m.character.CombatStats.RemoveCondition(cond)
				m.statusMessage = fmt.Sprintf("Removed %s", cond)
			}
		}
		m.conditionMode = false
		m.saveCharacter()
		return m, nil

	case tea.KeyUp:
		if m.conditionCursor > 0 {
			m.conditionCursor--
		}
		return m, nil

	case tea.KeyDown:
		if m.conditionCursor < listLen-1 {
			m.conditionCursor++
		}
		return m, nil
	}

	return m, nil
}

// handleCastingInput handles keyboard input when in spell casting modal mode.
func (m *MainSheetModel) handleCastingInput(msg tea.KeyMsg) (*MainSheetModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.castingSpell = nil
		m.availableCastLevels = nil
		m.statusMessage = "Casting cancelled"
		return m, nil

	case tea.KeyEnter:
		// Cast cantrip (no slot needed)
		if m.castingSpell.Level == 0 {
			m.statusMessage = fmt.Sprintf("Cast %s (cantrip)", m.castingSpell.Name)
			m.castingSpell = nil
			m.availableCastLevels = nil
			return m, nil
		}

		// Cast as ritual (no slot needed)
		if m.castingSpell.Ritual && len(m.availableCastLevels) == 0 {
			m.statusMessage = fmt.Sprintf("Cast %s as ritual (10 minutes, no slot)", m.castingSpell.Name)
			m.castingSpell = nil
			m.availableCastLevels = nil
			return m, nil
		}

		// Cast with spell slot
		if len(m.availableCastLevels) > 0 && m.castLevelCursor < len(m.availableCastLevels) {
			selectedLevel := m.availableCastLevels[m.castLevelCursor]
			m.castSpellAtLevel(m.castingSpell, selectedLevel)
			return m, nil
		}

		return m, nil

	case tea.KeyUp:
		if len(m.availableCastLevels) > 1 && m.castLevelCursor > 0 {
			m.castLevelCursor--
		}
		return m, nil

	case tea.KeyDown:
		if len(m.availableCastLevels) > 1 && m.castLevelCursor < len(m.availableCastLevels)-1 {
			m.castLevelCursor++
		}
		return m, nil
	}

	return m, nil
}

// handleRestInput handles keyboard input when in rest mode.
func (m *MainSheetModel) handleRestInput(msg tea.KeyMsg) (*MainSheetModel, tea.Cmd) {
	switch m.restMode {
	case RestModeMenu:
		switch msg.String() {
		case "s", "1":
			// Short rest
			m.restMode = RestModeShort
			m.restHitDice = 0
			return m, nil
		case "l", "2":
			// Long rest
			m.restMode = RestModeLong
			return m, nil
		case "esc", "q":
			m.restMode = RestModeNone
			return m, nil
		}

	case RestModeShort:
		switch msg.String() {
		case "up", "k", "+", "=":
			// Add hit die to spend (if available)
			hd := m.character.CombatStats.HitDice
			if m.restHitDice < hd.Remaining {
				m.restHitDice++
			}
			return m, nil
		case "down", "j", "-":
			// Remove hit die to spend
			if m.restHitDice > 0 {
				m.restHitDice--
			}
			return m, nil
		case "enter":
			// Perform short rest with hit dice spending
			m.performShortRest()
			return m, nil
		case "esc":
			m.restMode = RestModeMenu
			m.restHitDice = 0
			return m, nil
		}

	case RestModeLong:
		switch msg.String() {
		case "enter", "y":
			// Perform long rest
			m.performLongRest()
			return m, nil
		case "esc", "n":
			m.restMode = RestModeMenu
			return m, nil
		}

	case RestModeResult:
		// Any key dismisses the result screen
		m.restMode = RestModeNone
		m.restResult = ""
		return m, nil
	}

	return m, nil
}

// performShortRest executes a short rest with hit dice spending.
func (m *MainSheetModel) performShortRest() {
	char := m.character
	hd := &char.CombatStats.HitDice
	hp := &char.CombatStats.HitPoints
	conMod := char.AbilityScores.Constitution.Modifier()
	diceSpent := m.restHitDice
	oldHP := hp.Current

	// Spend hit dice and heal
	totalHealed := 0
	healingPerDie := []int{}
	for i := 0; i < m.restHitDice && hd.Remaining > 0; i++ {
		if hd.Use() {
			// Use average roll + CON modifier (dice rolling to be added later)
			avgRoll := (hd.DieType / 2) + 1
			healing := avgRoll + conMod
			if healing < 1 {
				healing = 1
			}
			healingPerDie = append(healingPerDie, healing)
			hp.Heal(healing)
			totalHealed += healing
		}
	}

	// Call character's short rest method for other effects (warlock slots, etc)
	char.ShortRest()

	m.saveCharacter()
	m.restHitDice = 0

	// Build result summary
	var result strings.Builder
	result.WriteString("SHORT REST COMPLETE\n\n")
	
	if diceSpent > 0 {
		result.WriteString(fmt.Sprintf("Hit Dice Spent: %d (d%d)\n", diceSpent, hd.DieType))
		result.WriteString(fmt.Sprintf("CON Modifier: %+d\n\n", conMod))
		
		result.WriteString("Healing Breakdown:\n")
		for i, heal := range healingPerDie {
			avgRoll := (hd.DieType / 2) + 1
			result.WriteString(fmt.Sprintf("  Die %d: %d (avg) + %d (CON) = %d HP\n", i+1, avgRoll, conMod, heal))
		}
		result.WriteString(fmt.Sprintf("\nTotal Healed: %d HP\n", totalHealed))
		result.WriteString(fmt.Sprintf("HP: %d → %d\n", oldHP, hp.Current))
	} else {
		result.WriteString("No hit dice spent.\n")
	}
	
	result.WriteString(fmt.Sprintf("\nHit Dice Remaining: %d/%d", hd.Remaining, hd.Total))

	m.restResult = result.String()
	m.restMode = RestModeResult
}

// performLongRest executes a long rest.
func (m *MainSheetModel) performLongRest() {
	char := m.character
	hp := &char.CombatStats.HitPoints
	hd := &char.CombatStats.HitDice
	oldHP := hp.Current
	oldHitDice := hd.Remaining

	char.LongRest()

	newHP := hp.Current
	newHitDice := hd.Remaining

	m.saveCharacter()

	// Build result summary
	var result strings.Builder
	result.WriteString("LONG REST COMPLETE\n\n")
	
	hpRestored := newHP - oldHP
	if hpRestored > 0 {
		result.WriteString(fmt.Sprintf("HP Restored: +%d (now %d/%d)\n", hpRestored, newHP, hp.Maximum))
	} else {
		result.WriteString(fmt.Sprintf("HP: %d/%d (already at max)\n", newHP, hp.Maximum))
	}
	
	hitDiceRestored := newHitDice - oldHitDice
	if hitDiceRestored > 0 {
		result.WriteString(fmt.Sprintf("Hit Dice Recovered: +%d (now %d/%d)\n", hitDiceRestored, newHitDice, hd.Total))
	} else {
		result.WriteString(fmt.Sprintf("Hit Dice: %d/%d (already at max)\n", newHitDice, hd.Total))
	}
	
	result.WriteString("\nAll spell slots restored\n")
	result.WriteString("Death saves reset\n")
	result.WriteString("Exhaustion reduced by 1 level")

	m.restResult = result.String()
	m.restMode = RestModeResult
}

// View renders the main sheet.
func (m *MainSheetModel) View() string {
	if m.character == nil {
		return "No character loaded"
	}

	// Calculate available width
	width := m.width
	if width == 0 {
		width = 140
	}
	height := m.height
	if height == 0 {
		height = 40
	}

	// Render sections
	header := m.renderHeader(width)
	
	// Two-column layout with vertical stacking
	// Left column is sized to content (abilities table + skills list)
	// Right column gets remaining space
	
	// Left column width: "Constitution" (12) + score (3) + mod (4) + save icon+mod (6) + padding/border (6) = ~35
	// Skills: icon (1) + mod (4) + name (15) + ability (5) + padding/border (6) = ~33
	// Use the wider of the two, plus a small margin
	leftWidth := 38
	rightWidth := width - leftWidth - 4  // 4 for gap between columns

	// Ensure minimum width for right column
	if rightWidth < 45 {
		rightWidth = 45
	}

	// Left column: Abilities/Saves on top, Skills below
	abilitiesAndSaves := m.renderAbilitiesAndSaves(leftWidth)
	skills := m.renderSkills(leftWidth)
	leftColumn := lipgloss.JoinVertical(lipgloss.Left, abilitiesAndSaves, skills)

	// Right column: Combat on top, Actions below
	combat := m.renderCombatStats(rightWidth)
	actions := m.renderActions(rightWidth)
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, combat, actions)

	// Join columns horizontally
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftColumn,
		"  ",
		rightColumn,
	)

	// Footer with navigation help
	footer := m.renderFooter(width)

	// Spell casting modal overlay
	if m.castingSpell != nil {
		castingModal := m.renderCastConfirmationModal()
		return lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			castingModal,
			footer,
		)
	}

	// Rest overlay if in rest mode
	if m.restMode != RestModeNone {
		restOverlay := m.renderRestOverlay(width)
		return lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			restOverlay,
			footer,
		)
	}

	// Join all sections vertically
	// TODO: Make view more dynamic based on terminal height to avoid overflow
	fullView := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		mainContent,
		footer,
	)

	return fullView
}

func (m *MainSheetModel) renderHeader(width int) string {
	char := m.character

	// Title style - make character name very prominent
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).  // Bright white for maximum visibility
		MarginBottom(0).
		Underline(true)  // Add underline for emphasis

	// Info style
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))

	// Build header content
	name := titleStyle.Render(char.Info.Name)

	// Race and class line
	raceClass := fmt.Sprintf("%s %s %d",
		char.Info.Race,
		char.Info.Class,
		char.Info.Level,
	)

	// XP or Milestone
	var progression string
	if char.Info.ProgressionType == models.ProgressionMilestone {
		progression = "Milestone"
	} else {
		progression = fmt.Sprintf("XP: %d / %d", char.Info.ExperiencePoints, models.XPForNextLevel(char.Info.Level))
	}

	// Inspiration
	inspiration := ""
	if char.Info.Inspiration {
		inspiration = " ★ Inspired"
	}

	// Build header
	headerLeft := lipgloss.JoinVertical(
		lipgloss.Left,
		name,
		infoStyle.Render(raceClass),
	)

	// Proficiency legend icons
	profIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("76")).Render("●")
	expertIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("◆")
	legendStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	legend := legendStyle.Render(fmt.Sprintf("%s Proficient  %s Expertise", profIcon, expertIcon))

	headerRight := lipgloss.JoinVertical(
		lipgloss.Right,
		labelStyle.Render("Proficiency: ")+infoStyle.Render(fmt.Sprintf("+%d", char.GetProficiencyBonus())),
		labelStyle.Render(progression)+infoStyle.Render(inspiration),
		legend,
	)

	// Join header left and right
	headerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 2).
		Width(width - 2)

	headerContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		headerLeft,
		strings.Repeat(" ", max(1, width-lipgloss.Width(headerLeft)-lipgloss.Width(headerRight)-10)),
		headerRight,
	)

	return headerStyle.Render(headerContent)
}

func (m *MainSheetModel) renderAbilitiesAndSaves(width int) string {
	char := m.character

	// Styles
	isFocused := m.focusArea == FocusAbilitiesAndSaves
	borderColor := lipgloss.Color("240")
	if isFocused {
		borderColor = lipgloss.Color("99")
	}

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(width - 2)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99"))

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Bold(true)

	profIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("76")).Render("●")
	noProfIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("○")

	// Build combined ability scores and saving throws
	abilities := []struct {
		name    string
		short   string
		score   models.AbilityScore
		save    *models.SavingThrow
		ability models.Ability
	}{
		{"Strength", "STR", char.AbilityScores.Strength, &char.SavingThrows.Strength, models.AbilityStrength},
		{"Dexterity", "DEX", char.AbilityScores.Dexterity, &char.SavingThrows.Dexterity, models.AbilityDexterity},
		{"Constitution", "CON", char.AbilityScores.Constitution, &char.SavingThrows.Constitution, models.AbilityConstitution},
		{"Intelligence", "INT", char.AbilityScores.Intelligence, &char.SavingThrows.Intelligence, models.AbilityIntelligence},
		{"Wisdom", "WIS", char.AbilityScores.Wisdom, &char.SavingThrows.Wisdom, models.AbilityWisdom},
		{"Charisma", "CHA", char.AbilityScores.Charisma, &char.SavingThrows.Charisma, models.AbilityCharisma},
	}

	var lines []string
	lines = append(lines, titleStyle.Render("Abilities & Saving Throws"))
	
	// Header row - use plain formatting then style the whole line
	headerLine := fmt.Sprintf("%-12s %5s %4s %5s", "Ability", "Score", "Mod", "Save")
	lines = append(lines, headerStyle.Render(headerLine))

	for _, a := range abilities {
		mod := a.score.Modifier()
		modStr := formatModifier(mod)
		saveMod := char.GetSavingThrowModifier(a.ability)
		saveStr := formatModifier(saveMod)

		icon := noProfIcon
		if a.save.Proficient {
			icon = profIcon
		}

		// Format with fixed widths: name(12) score(5) mod(4) icon+save(5)
		line := fmt.Sprintf("%-12s %5d %4s %s%4s",
			a.name,
			a.score.Total(),
			modStr,
			icon,
			saveStr,
		)
		lines = append(lines, line)
	}

	return panelStyle.Render(strings.Join(lines, "\n"))
}

// renderAbilities is kept for backwards compatibility but now redirects
func (m *MainSheetModel) renderAbilities(width int) string {
	return m.renderAbilitiesAndSaves(width)
}

func (m *MainSheetModel) renderSkills(width int) string {
	char := m.character

	// Styles
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(width)

	if m.focusArea == FocusSkills {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("99"))
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))

	profIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("76")).Render("●")
	expertIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("◆")
	noProfIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("○")

	// Skill display names
	skillNames := map[models.SkillName]string{
		models.SkillAcrobatics:     "Acrobatics",
		models.SkillAnimalHandling: "Animal Handling",
		models.SkillArcana:         "Arcana",
		models.SkillAthletics:      "Athletics",
		models.SkillDeception:      "Deception",
		models.SkillHistory:        "History",
		models.SkillInsight:        "Insight",
		models.SkillIntimidation:   "Intimidation",
		models.SkillInvestigation:  "Investigation",
		models.SkillMedicine:       "Medicine",
		models.SkillNature:         "Nature",
		models.SkillPerception:     "Perception",
		models.SkillPerformance:    "Performance",
		models.SkillPersuasion:     "Persuasion",
		models.SkillReligion:       "Religion",
		models.SkillSleightOfHand:  "Sleight of Hand",
		models.SkillStealth:        "Stealth",
		models.SkillSurvival:       "Survival",
	}

	// Skill ability abbreviations
	skillAbilityAbbr := map[models.Ability]string{
		models.AbilityStrength:     "STR",
		models.AbilityDexterity:    "DEX",
		models.AbilityConstitution: "CON",
		models.AbilityIntelligence: "INT",
		models.AbilityWisdom:       "WIS",
		models.AbilityCharisma:     "CHA",
	}

	var lines []string
	lines = append(lines, titleStyle.Render("Skills"))

	// Passive skills - Perception, Investigation, Insight (aligned numbers)
	passivePerception := 10 + char.GetSkillModifier(models.SkillPerception)
	passiveInvestigation := 10 + char.GetSkillModifier(models.SkillInvestigation)
	passiveInsight := 10 + char.GetSkillModifier(models.SkillInsight)
	
	// Format passives with plain strings first
	lines = append(lines, labelStyle.Render(fmt.Sprintf("Passive Perception:    %2d", passivePerception)))
	lines = append(lines, labelStyle.Render(fmt.Sprintf("Passive Investigation: %2d", passiveInvestigation)))
	lines = append(lines, labelStyle.Render(fmt.Sprintf("Passive Insight:       %2d", passiveInsight)))
	lines = append(lines, "")

	for _, skillName := range models.AllSkills() {
		skill := char.Skills.Get(skillName)
		ability := models.GetSkillAbility(skillName)
		mod := char.GetSkillModifier(skillName)
		modStr := formatModifier(mod)

		icon := noProfIcon
		if skill.Proficiency == models.Proficient {
			icon = profIcon
		} else if skill.Proficiency == models.Expertise {
			icon = expertIcon
		}

		displayName := skillNames[skillName]
		abilityAbbr := skillAbilityAbbr[ability]

		// Format: icon name modifier (ability) - modifier between name and ability
		line := fmt.Sprintf("%s %-15s %3s %s",
			icon,
			displayName,
			modStr,
			labelStyle.Render(fmt.Sprintf("(%s)", abilityAbbr)),
		)
		lines = append(lines, line)
	}

	return panelStyle.Render(strings.Join(lines, "\n"))
}

func (m *MainSheetModel) renderCombatStats(width int) string {
	char := m.character

	// Styles
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(width - 2) // Account for border

	if m.focusArea == FocusCombat {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("99"))
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))

	valueStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("252"))

	var lines []string
	lines = append(lines, titleStyle.Render("Combat"))
	lines = append(lines, "")

	// HP with visual bar
	hp := char.CombatStats.HitPoints
	hpPercent := float64(hp.Current) / float64(hp.Maximum)
	hpColor := "76" // green
	if hpPercent < 0.25 {
		hpColor = "196" // red
	} else if hpPercent < 0.5 {
		hpColor = "214" // yellow
	}

	hpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(hpColor))
	hpLine := fmt.Sprintf("%s %s / %s",
		labelStyle.Render("HP:"),
		hpStyle.Render(fmt.Sprintf("%d", hp.Current)),
		valueStyle.Render(fmt.Sprintf("%d", hp.Maximum)),
	)
	if hp.Temporary > 0 {
		hpLine += labelStyle.Render(fmt.Sprintf(" (+%d temp)", hp.Temporary))
	}
	lines = append(lines, hpLine)

	// HP Bar - shows current HP + temp HP
	barWidth := width - 6
	if barWidth < 10 {
		barWidth = 10
	}

	// Calculate widths for current HP, temp HP, and empty
	totalEffectiveHP := hp.Current + hp.Temporary
	maxForBar := hp.Maximum
	if totalEffectiveHP > maxForBar {
		maxForBar = totalEffectiveHP // Extend bar if temp HP exceeds max
	}

	currentWidth := 0
	if maxForBar > 0 {
		currentWidth = int(float64(barWidth) * float64(hp.Current) / float64(maxForBar))
	}
	if currentWidth < 0 {
		currentWidth = 0
	}
	if currentWidth > barWidth {
		currentWidth = barWidth
	}

	tempWidth := 0
	if hp.Temporary > 0 && maxForBar > 0 {
		tempWidth = int(float64(barWidth) * float64(hp.Temporary) / float64(maxForBar))
		if tempWidth < 1 {
			tempWidth = 1 // Ensure at least 1 char if there's any temp HP
		}
	}
	if currentWidth+tempWidth > barWidth {
		tempWidth = barWidth - currentWidth
	}

	emptyWidth := barWidth - currentWidth - tempWidth
	if emptyWidth < 0 {
		emptyWidth = 0
	}

	barCurrent := lipgloss.NewStyle().Background(lipgloss.Color(hpColor)).Render(strings.Repeat(" ", currentWidth))
	barTemp := lipgloss.NewStyle().Background(lipgloss.Color("39")).Render(strings.Repeat(" ", tempWidth)) // Cyan/blue for temp HP
	barEmpty := lipgloss.NewStyle().Background(lipgloss.Color("238")).Render(strings.Repeat(" ", emptyWidth))
	lines = append(lines, fmt.Sprintf("[%s%s%s]", barCurrent, barTemp, barEmpty))

	// HP controls hint
	if m.focusArea == FocusCombat {
		hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Italic(true)
		lines = append(lines, hintStyle.Render("d: damage  h: heal  t: temp HP"))
	}
	lines = append(lines, "")

	// AC
	lines = append(lines, fmt.Sprintf("%s %s",
		labelStyle.Render("AC:"),
		valueStyle.Render(fmt.Sprintf("%d", char.CombatStats.ArmorClass)),
	))

	// Initiative
	init := char.GetInitiative()
	lines = append(lines, fmt.Sprintf("%s %s",
		labelStyle.Render("Initiative:"),
		valueStyle.Render(formatModifier(init)),
	))

	// Speed
	lines = append(lines, fmt.Sprintf("%s %s ft",
		labelStyle.Render("Speed:"),
		valueStyle.Render(fmt.Sprintf("%d", char.CombatStats.Speed)),
	))

	// Hit Dice
	hd := char.CombatStats.HitDice
	lines = append(lines, fmt.Sprintf("%s %s/%s d%d",
		labelStyle.Render("Hit Dice:"),
		valueStyle.Render(fmt.Sprintf("%d", hd.Remaining)),
		valueStyle.Render(fmt.Sprintf("%d", hd.Total)),
		hd.DieType,
	))

	// Death Saves (only show if relevant)
	if hp.Current == 0 {
		lines = append(lines, "")
		lines = append(lines, titleStyle.Render("Death Saves"))
		ds := char.CombatStats.DeathSaves
		successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("76"))
		failStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

		successes := strings.Repeat(successStyle.Render("●"), ds.Successes) +
			strings.Repeat(labelStyle.Render("○"), 3-ds.Successes)
		failures := strings.Repeat(failStyle.Render("●"), ds.Failures) +
			strings.Repeat(labelStyle.Render("○"), 3-ds.Failures)

		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Successes:"), successes))
		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Failures: "), failures))

		// Status indicator
		if ds.Successes >= 3 {
			stableStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("76")).Bold(true)
			lines = append(lines, stableStyle.Render("★ STABILIZED"))
		} else if ds.Failures >= 3 {
			deadStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
			lines = append(lines, deadStyle.Render("☠ DEAD"))
		} else if m.focusArea == FocusCombat {
			hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Italic(true)
			lines = append(lines, hintStyle.Render("1: success  2: fail  0: reset"))
		}
	}

	// Spellcasting info if applicable
	if char.IsSpellcaster() {
		lines = append(lines, "")
		lines = append(lines, titleStyle.Render("Spellcasting"))

		lines = append(lines, fmt.Sprintf("%s %s",
			labelStyle.Render("Spell Save DC:"),
			valueStyle.Render(fmt.Sprintf("%d", char.GetSpellSaveDC())),
		))
		lines = append(lines, fmt.Sprintf("%s %s",
			labelStyle.Render("Spell Attack:"),
			valueStyle.Render(formatModifier(char.GetSpellAttackBonus())),
		))

		// Show spell slots or pact magic
		sc := char.Spellcasting
		if sc.PactMagic != nil {
			// Warlock pact magic
			pm := sc.PactMagic
			slotStyle := valueStyle
			if pm.Remaining == 0 {
				slotStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
			}
			lines = append(lines, fmt.Sprintf("%s %s",
				labelStyle.Render(fmt.Sprintf("Pact Slots (L%d):", pm.SlotLevel)),
				slotStyle.Render(fmt.Sprintf("%d/%d", pm.Remaining, pm.Total)),
			))
		} else {
			// Regular spell slots - show only levels with slots
			hasSlots := false
			for level := 1; level <= 9; level++ {
				slot := sc.SpellSlots.GetSlot(level)
				if slot != nil && slot.Total > 0 {
					hasSlots = true
					slotStyle := valueStyle
					if slot.Remaining == 0 {
						slotStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
					}
					lines = append(lines, fmt.Sprintf("%s %s",
						labelStyle.Render(fmt.Sprintf("Level %d:", level)),
						slotStyle.Render(fmt.Sprintf("%d/%d", slot.Remaining, slot.Total)),
					))
				}
			}
			if !hasSlots {
				lines = append(lines, labelStyle.Render("No spell slots"))
			}
		}
	}

	// Active conditions (always show section with hint when focused)
	lines = append(lines, "")
	lines = append(lines, titleStyle.Render("Conditions"))
	if len(char.CombatStats.Conditions) > 0 {
		condStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
		for _, cond := range char.CombatStats.Conditions {
			lines = append(lines, condStyle.Render("• "+string(cond)))
		}
	} else {
		lines = append(lines, labelStyle.Render("None"))
	}
	if m.focusArea == FocusCombat {
		hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Italic(true)
		lines = append(lines, hintStyle.Render("+: add  -: remove"))
	}

	return panelStyle.Render(strings.Join(lines, "\n"))
}

func (m *MainSheetModel) renderActions(width int) string {
	char := m.character
	isFocused := m.focusArea == FocusActions

	borderColor := lipgloss.Color("240")
	if isFocused {
		borderColor = lipgloss.Color("99")
	}

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(width - 2)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	unselectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	var lines []string

	// Action type tabs
	actionTypes := []string{"Action", "Bonus", "Reaction", "Other"}
	var tabs []string
	for i, at := range actionTypes {
		if ActionType(i) == m.selectedActionType {
			tabs = append(tabs, selectedStyle.Render("["+at+"]"))
		} else {
			tabs = append(tabs, unselectedStyle.Render(" "+at+" "))
		}
	}
	lines = append(lines, strings.Join(tabs, " "))
	if isFocused {
		hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Italic(true)
		lines = append(lines, hintStyle.Render("←/→: switch type  ↑↓: select  Enter: use"))
	}

	// Get all action items for current type
	actionItems := m.getActionItems()

	// Ensure cursor is valid
	if m.actionCursor >= len(actionItems) {
		m.actionCursor = len(actionItems) - 1
	}
	if m.actionCursor < 0 {
		m.actionCursor = 0
	}

	// Render action items with cursor, grouped by category
	if len(actionItems) == 0 {
		if m.selectedActionType == ActionTypeOther {
			// Other tab - non-interactive items
			lines = append(lines, titleStyle.Render("Other"))
			lines = append(lines, fmt.Sprintf("  %s", valueStyle.Render("Object Interaction")))
			lines = append(lines, fmt.Sprintf("    %s", labelStyle.Render("Interact with one object for free")))
			lines = append(lines, fmt.Sprintf("  %s", valueStyle.Render("Movement")))
			lines = append(lines, fmt.Sprintf("    %s", labelStyle.Render(fmt.Sprintf("Move up to %d ft (can split)", char.CombatStats.Speed))))
			lines = append(lines, fmt.Sprintf("  %s", valueStyle.Render("Drop Prone")))
			lines = append(lines, fmt.Sprintf("    %s", labelStyle.Render("Fall prone (no cost)")))
			lines = append(lines, fmt.Sprintf("  %s", valueStyle.Render("Stand Up")))
			lines = append(lines, fmt.Sprintf("    %s", labelStyle.Render("Costs half your Speed")))
		} else {
			lines = append(lines, labelStyle.Render("  No actions available"))
		}
	} else {
		// Group items by type: Weapons, Spells, Standard Actions
		var weaponItems, spellItems, standardItems []ActionItem
		for _, item := range actionItems {
			switch item.Type {
			case ActionItemWeapon:
				weaponItems = append(weaponItems, item)
			case ActionItemSpell:
				spellItems = append(spellItems, item)
			case ActionItemStandard:
				standardItems = append(standardItems, item)
			}
		}

		currentIndex := 0

		// Render weapons section
		if len(weaponItems) > 0 {
			lines = append(lines, titleStyle.Render("Weapon Attacks"))
			for _, item := range weaponItems {
				cursor := "  "
				if isFocused && currentIndex == m.actionCursor {
					cursor = "> "
				}

				itemLine := fmt.Sprintf("%s%s - %s", cursor, valueStyle.Render(item.Name), labelStyle.Render(item.Description))

				if isFocused && currentIndex == m.actionCursor {
					itemLine = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(itemLine)
				}

				lines = append(lines, itemLine)
				currentIndex++
			}
		}

		// Render spells section
		if len(spellItems) > 0 {
			if len(weaponItems) > 0 {
				lines = append(lines, "")
			}
			lines = append(lines, titleStyle.Render("Spells"))
			for _, item := range spellItems {
				cursor := "  "
				if isFocused && currentIndex == m.actionCursor {
					cursor = "> "
				}

				itemLine := fmt.Sprintf("%s%s - %s", cursor, valueStyle.Render(item.Name), labelStyle.Render(item.Description))

				if isFocused && currentIndex == m.actionCursor {
					itemLine = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(itemLine)
				}

				lines = append(lines, itemLine)
				currentIndex++
			}
		}

		// Render standard actions section
		if len(standardItems) > 0 {
			if len(weaponItems) > 0 || len(spellItems) > 0 {
				lines = append(lines, "")
			}
			lines = append(lines, titleStyle.Render("Standard Actions"))
			for _, item := range standardItems {
				cursor := "  "
				if isFocused && currentIndex == m.actionCursor {
					cursor = "> "
				}

				itemLine := fmt.Sprintf("%s%s - %s", cursor, valueStyle.Render(item.Name), labelStyle.Render(item.Description))

				if isFocused && currentIndex == m.actionCursor {
					itemLine = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(itemLine)
				}

				lines = append(lines, itemLine)
				currentIndex++
			}
		}
	}

	return panelStyle.Render(strings.Join(lines, "\n"))
}

// getSpellsByCastingTime returns spells that match the given casting time.
// castingTime can be: "A" (Action), "BA" (Bonus Action), "R" (Reaction), etc.
func (m *MainSheetModel) getSpellsByCastingTime(castingTime string) []data.SpellData {
	if m.character == nil || m.character.Spellcasting == nil || m.spellDatabase == nil {
		return nil
	}

	var spells []data.SpellData
	sc := m.character.Spellcasting

	// Helper to check if spell is available
	isAvailable := func(spellName string, level int, spell data.SpellData) bool {
		// Cantrips are always available
		if level == 0 {
			for _, c := range sc.CantripsKnown {
				if c == spellName {
					return true
				}
			}
			return false
		}

		// For leveled spells, check if prepared (or always available for non-preparing classes)
		for _, ks := range sc.KnownSpells {
			if ks.Name == spellName {
				// Non-preparing classes: all known spells are available
				if !sc.PreparesSpells {
					return true
				}

				// Always prepared spells
				if ks.AlwaysPrepared {
					return true
				}

				// Prepared spells
				if ks.Prepared {
					return true
				}

				// Wizards can cast ritual spells from spellbook without preparing
				if spell.Ritual && sc.RitualCasterUnprepared {
					return true
				}

				return false
			}
		}
		return false
	}

	// Search through all spells in database
	for _, spell := range m.spellDatabase.Spells {
		if spell.CastingTime == castingTime && isAvailable(spell.Name, spell.Level, spell) {
			spells = append(spells, spell)
		}
	}

	return spells
}

// getActionItems returns all available actions for the current action type
func (m *MainSheetModel) getActionItems() []ActionItem {
	if m.character == nil {
		return nil
	}

	var items []ActionItem

	switch m.selectedActionType {
	case ActionTypeAction:
		// Unarmed Strike (always available)
		items = append(items, ActionItem{
			Type:        ActionItemWeapon,
			Name:        "Unarmed Strike",
			Description: fmt.Sprintf("Hit: %s, Dmg: 1 bludgeoning",
				formatModifier(m.character.AbilityScores.Strength.Modifier())),
		})

		// Equipped weapons
		weapons := m.getWeapons()
		for i := range weapons {
			w := &weapons[i]
			attackBonus := m.getWeaponAttackBonus(*w)
			damageMod := m.getWeaponDamageMod(*w)

			damageStr := w.Damage
			if damageMod != 0 {
				damageStr = fmt.Sprintf("%s%s", w.Damage, formatModifier(damageMod))
			}
			damageStr += " " + w.DamageType

			if w.RangeNormal > 0 {
				damageStr = fmt.Sprintf("%s (%d/%d ft)", damageStr, w.RangeNormal, w.RangeLong)
			}

			items = append(items, ActionItem{
				Type:        ActionItemWeapon,
				Name:        w.Name,
				Description: fmt.Sprintf("Hit: %s, Dmg: %s", formatModifier(attackBonus), damageStr),
				Weapon:      w,
			})
		}

		// Action spells (before standard actions in display order)
		actionSpells := m.getSpellsByCastingTime("A")
		for i := range actionSpells {
			spell := &actionSpells[i]
			spellInfo := fmt.Sprintf("Level %d", spell.Level)
			if spell.Level == 0 {
				spellInfo = "Cantrip"
			}
			items = append(items, ActionItem{
				Type:        ActionItemSpell,
				Name:        spell.Name,
				Description: spellInfo,
				Spell:       spell,
			})
		}

		// Standard actions (after spells in display order)
		for i := range standardActions {
			action := &standardActions[i]
			if action.ActionType == ActionTypeAction && action.Name != "Attack" {
				items = append(items, ActionItem{
					Type:           ActionItemStandard,
					Name:           action.Name,
					Description:    action.Description,
					StandardAction: action,
				})
			}
		}

	case ActionTypeBonus:
		// Bonus action spells (before standard actions in display order)
		bonusSpells := m.getSpellsByCastingTime("BA")
		for i := range bonusSpells {
			spell := &bonusSpells[i]
			spellInfo := fmt.Sprintf("Level %d", spell.Level)
			if spell.Level == 0 {
				spellInfo = "Cantrip"
			}
			items = append(items, ActionItem{
				Type:        ActionItemSpell,
				Name:        spell.Name,
				Description: spellInfo,
				Spell:       spell,
			})
		}

		// Standard bonus actions (after spells in display order)
		for i := range standardActions {
			action := &standardActions[i]
			if action.ActionType == ActionTypeBonus {
				items = append(items, ActionItem{
					Type:           ActionItemStandard,
					Name:           action.Name,
					Description:    action.Description,
					StandardAction: action,
				})
			}
		}

	case ActionTypeReaction:
		// Reaction spells (before standard actions in display order)
		reactionSpells := m.getSpellsByCastingTime("R")
		for i := range reactionSpells {
			spell := &reactionSpells[i]
			spellInfo := fmt.Sprintf("Level %d", spell.Level)
			if spell.Level == 0 {
				spellInfo = "Cantrip"
			}
			items = append(items, ActionItem{
				Type:        ActionItemSpell,
				Name:        spell.Name,
				Description: spellInfo,
				Spell:       spell,
			})
		}

		// Standard reactions (after spells in display order)
		for i := range standardActions {
			action := &standardActions[i]
			if action.ActionType == ActionTypeReaction {
				items = append(items, ActionItem{
					Type:           ActionItemStandard,
					Name:           action.Name,
					Description:    action.Description,
					StandardAction: action,
				})
			}
		}

	case ActionTypeOther:
		// Movement and other non-standard actions (not interactive for now)
		// Could add these later if needed
	}

	return items
}

func (m *MainSheetModel) renderRestOverlay(width int) string {
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1, 2).
		Width(width - 4)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	var lines []string

	switch m.restMode {
	case RestModeMenu:
		lines = append(lines, titleStyle.Render("⛺ Rest Options"))
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("[S] Short Rest - Spend hit dice to recover HP"))
		lines = append(lines, labelStyle.Render("[L] Long Rest  - Full recovery (8 hours)"))
		lines = append(lines, "")
		lines = append(lines, dimStyle.Render("Press Esc to cancel"))

	case RestModeShort:
		hd := m.character.CombatStats.HitDice
		hp := m.character.CombatStats.HitPoints
		conMod := m.character.AbilityScores.Constitution.Modifier()

		lines = append(lines, titleStyle.Render("⛺ Short Rest - Spend Hit Dice"))
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf(
			"%s %s%s%s",
			labelStyle.Render("Current HP:"),
			valueStyle.Render(fmt.Sprintf("%d", hp.Current)),
			dimStyle.Render("/"),
			fmt.Sprintf("%d", hp.Maximum),
		))
		lines = append(lines, fmt.Sprintf(
			"%s %s%s%s",
			labelStyle.Render("Hit Dice:"),
			valueStyle.Render(fmt.Sprintf("%d", hd.Remaining)),
			dimStyle.Render("/"),
			fmt.Sprintf("%d (d%d)", hd.Total, hd.DieType),
		))
		
		// Show CON modifier
		conModStr := fmt.Sprintf("%+d", conMod)
		lines = append(lines, fmt.Sprintf(
			"%s %s %s",
			labelStyle.Render("CON Modifier:"),
			valueStyle.Render(conModStr),
			dimStyle.Render("(added to each hit die roll)"),
		))
		lines = append(lines, "")

		// Show hit dice to spend
		lines = append(lines, fmt.Sprintf(
			"%s %s",
			labelStyle.Render("Hit Dice to Spend:"),
			valueStyle.Render(fmt.Sprintf("%d", m.restHitDice)),
		))

		// Show expected healing
		if m.restHitDice > 0 {
			avgPerDie := (hd.DieType / 2) + 1 + conMod
			if avgPerDie < 1 {
				avgPerDie = 1
			}
			expectedHealing := avgPerDie * m.restHitDice
			lines = append(lines, fmt.Sprintf(
				"%s %s %s",
				labelStyle.Render("Expected Healing:"),
				valueStyle.Render(fmt.Sprintf("~%d HP", expectedHealing)),
				dimStyle.Render(fmt.Sprintf("(avg %d per die)", avgPerDie)),
			))
		}
		lines = append(lines, "")
		lines = append(lines, dimStyle.Render("↑/↓: Adjust hit dice • Enter: Confirm • Esc: Back"))

	case RestModeLong:
		hp := m.character.CombatStats.HitPoints
		hd := m.character.CombatStats.HitDice

		lines = append(lines, titleStyle.Render("🌙 Long Rest - Full Recovery"))
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("A long rest requires 8 hours. After resting:"))
		lines = append(lines, "")

		// HP restoration
		hpToRestore := hp.Maximum - hp.Current
		if hpToRestore > 0 {
			lines = append(lines, fmt.Sprintf(
				"  • %s %s",
				labelStyle.Render("Restore HP:"),
				valueStyle.Render(fmt.Sprintf("+%d (to %d)", hpToRestore, hp.Maximum)),
			))
		} else {
			lines = append(lines, fmt.Sprintf(
				"  • %s %s",
				labelStyle.Render("HP:"),
				dimStyle.Render("Already at maximum"),
			))
		}

		// Hit dice recovery (half, min 1)
		hitDiceToRecover := hd.Total / 2
		if hitDiceToRecover < 1 {
			hitDiceToRecover = 1
		}
		if hd.Remaining < hd.Total {
			canRecover := hitDiceToRecover
			if hd.Remaining+canRecover > hd.Total {
				canRecover = hd.Total - hd.Remaining
			}
			lines = append(lines, fmt.Sprintf(
				"  • %s %s",
				labelStyle.Render("Recover Hit Dice:"),
				valueStyle.Render(fmt.Sprintf("+%d (to %d/%d)", canRecover, hd.Remaining+canRecover, hd.Total)),
			))
		} else {
			lines = append(lines, fmt.Sprintf(
				"  • %s %s",
				labelStyle.Render("Hit Dice:"),
				dimStyle.Render("Already at maximum"),
			))
		}

		// Spell slots restoration
		lines = append(lines, fmt.Sprintf("  • %s", labelStyle.Render("Restore all spell slots")))

		// Exhaustion reduction
		exhaustion := 0
		for _, cond := range m.character.CombatStats.Conditions {
			if strings.HasPrefix(string(cond), "Exhaustion") {
				exhaustion++
			}
		}
		if exhaustion > 0 {
			lines = append(lines, fmt.Sprintf(
				"  • %s %s",
				labelStyle.Render("Reduce exhaustion:"),
				valueStyle.Render("-1 level"),
			))
		}

		// Reset death saves
		lines = append(lines, fmt.Sprintf("  • %s", labelStyle.Render("Reset death saves")))

		lines = append(lines, "")
		lines = append(lines, dimStyle.Render("Enter/Y: Confirm • Esc/N: Cancel"))

	case RestModeResult:
		// Display the result summary
		lines = append(lines, titleStyle.Render("✓ Rest Complete"))
		lines = append(lines, "")
		for _, line := range strings.Split(m.restResult, "\n") {
			if strings.HasPrefix(line, "SHORT REST") || strings.HasPrefix(line, "LONG REST") {
				lines = append(lines, titleStyle.Render(line))
			} else if strings.HasPrefix(line, "  ") {
				lines = append(lines, dimStyle.Render(line))
			} else if strings.Contains(line, "+") || strings.Contains(line, "→") {
				lines = append(lines, valueStyle.Render(line))
			} else {
				lines = append(lines, labelStyle.Render(line))
			}
		}
		lines = append(lines, "")
		lines = append(lines, dimStyle.Render("Press any key to continue"))
	}

	return panelStyle.Render(strings.Join(lines, "\n"))
}

// renderCastConfirmationModal renders the spell casting confirmation modal.
func (m *MainSheetModel) renderCastConfirmationModal() string {
	if m.castingSpell == nil {
		return ""
	}

	spell := m.castingSpell
	var lines []string

	// Title
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Cast %s", spell.Name)))
	lines = append(lines, "")

	// Level and school
	levelSchool := fmt.Sprintf("Level %d %s", spell.Level, spell.School)
	if spell.Level == 0 {
		levelSchool = fmt.Sprintf("%s cantrip", spell.School)
	}
	if spell.Ritual {
		levelSchool += " (ritual)"
	}
	lines = append(lines, levelSchool)
	lines = append(lines, "")

	// Basic spell info
	castingTime := spell.CastingTime
	// Check if being cast as ritual (no slots available but spell has ritual tag)
	if spell.Ritual && len(m.availableCastLevels) == 0 {
		castingTime = "10 minutes (ritual)"
	}
	lines = append(lines, fmt.Sprintf("Casting Time: %s", castingTime))
	lines = append(lines, fmt.Sprintf("Range: %s", spell.Range))
	lines = append(lines, fmt.Sprintf("Components: %s", strings.Join(spell.Components, ", ")))
	lines = append(lines, fmt.Sprintf("Duration: %s", spell.Duration))

	// Damage if present
	if spell.Damage != "" {
		damageInfo := spell.Damage
		if spell.DamageType != "" {
			damageInfo = fmt.Sprintf("%s %s", damageInfo, spell.DamageType)
		}
		lines = append(lines, fmt.Sprintf("Damage: %s", damageInfo))
	}

	// Saving throw if present
	if spell.SavingThrow != "" {
		saveDC := m.getSpellSaveDC()
		lines = append(lines, fmt.Sprintf("Saving Throw: %s DC %d", spell.SavingThrow, saveDC))
	}

	lines = append(lines, "")

	// Description (word-wrapped)
	descLines := m.wordWrap(spell.Description, 60)
	lines = append(lines, descLines...)

	// Upcast information if available
	if spell.Upcast != "" && spell.Level > 0 && spell.Level < 9 {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Bold(true).Render("At Higher Levels:"))
		upcastLines := m.wordWrap(spell.Upcast, 60)
		lines = append(lines, upcastLines...)
	}

	// Separator before slot selection
	if spell.Level > 0 && len(m.availableCastLevels) > 0 {
		lines = append(lines, "")
		lines = append(lines, strings.Repeat("─", 60))
		lines = append(lines, "")
	}

	// Slot selection (if not a cantrip and not being cast as ritual)
	if spell.Level > 0 && len(m.availableCastLevels) > 0 {
		if len(m.availableCastLevels) == 1 {
			// Single slot option
			level := m.availableCastLevels[0]
			upcastInfo := m.calculateUpcastEffect(level, spell)
			sc := m.character.Spellcasting

			isPactMagic := sc.PactMagic != nil && sc.PactMagic.SlotLevel == level
			if isPactMagic {
				lines = append(lines, fmt.Sprintf("Using: Pact Magic - Level %d%s", level, upcastInfo))
			} else {
				lines = append(lines, fmt.Sprintf("Using: Level %d Slot%s", level, upcastInfo))
			}
		} else {
			// Multiple slot options
			lines = append(lines, "Select Spell Slot Level:")
			lines = append(lines, "")

			sc := m.character.Spellcasting
			for i, level := range m.availableCastLevels {
				cursor := "  "
				if i == m.castLevelCursor {
					cursor = "> "
				}

				isPactMagic := sc.PactMagic != nil && sc.PactMagic.SlotLevel == level
				upcastInfo := m.calculateUpcastEffect(level, spell)

				var line string
				if isPactMagic {
					remaining := sc.PactMagic.Remaining
					total := sc.PactMagic.Total
					line = fmt.Sprintf("%sPact Magic - Level %d%s [%d/%d remaining]", cursor, level, upcastInfo, remaining, total)
				} else {
					slot := sc.SpellSlots.GetSlot(level)
					if slot != nil {
						line = fmt.Sprintf("%sLevel %d Slot%s [%d/%d remaining]", cursor, level, upcastInfo, slot.Remaining, slot.Total)
					}
				}

				if i == m.castLevelCursor {
					line = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(line)
				}

				lines = append(lines, line)
			}
		}
	}

	// Help text
	lines = append(lines, "")
	if spell.Level == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Enter: cast cantrip | Esc: cancel"))
	} else if spell.Ritual && len(m.availableCastLevels) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Enter: cast as ritual (no slot) | Esc: cancel"))
	} else if len(m.availableCastLevels) <= 1 {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Enter: cast | Esc: cancel"))
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("↑↓: select slot | Enter: cast | Esc: cancel"))
	}

	content := strings.Join(lines, "\n")

	width := 70
	height := len(lines) + 2

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Render(content)
}

// castSpellAtLevel casts a spell at the given slot level, consuming the appropriate resource.
func (m *MainSheetModel) castSpellAtLevel(spell *data.SpellData, slotLevel int) {
	if m.character == nil || m.character.Spellcasting == nil {
		return
	}

	sc := m.character.Spellcasting

	// Try pact magic first if it matches the slot level
	if sc.PactMagic != nil && sc.PactMagic.SlotLevel == slotLevel && sc.PactMagic.Remaining > 0 {
		if sc.PactMagic.Use() {
			upcastMsg := ""
			if slotLevel > spell.Level {
				upcastMsg = fmt.Sprintf(" (upcast to level %d)", slotLevel)
			}
			m.statusMessage = fmt.Sprintf("Cast %s%s using pact magic", spell.Name, upcastMsg)
			m.castingSpell = nil
			m.availableCastLevels = nil
			m.saveCharacter()
			return
		}
	}

	// Try regular spell slot
	if sc.SpellSlots.UseSlot(slotLevel) {
		upcastMsg := ""
		if slotLevel > spell.Level {
			upcastMsg = fmt.Sprintf(" (upcast to level %d)", slotLevel)
		}
		m.statusMessage = fmt.Sprintf("Cast %s%s using level %d slot", spell.Name, upcastMsg, slotLevel)
		m.castingSpell = nil
		m.availableCastLevels = nil
		m.saveCharacter()
		return
	}

	m.statusMessage = "Failed to use spell slot"
}

func (m *MainSheetModel) renderFooter(width int) string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Width(width)

	help := "tab/shift+tab: navigate panels • i: inventory • s: spellbook • c: character info • r: rest • esc: back • q: quit"

	// Show condition selection if in condition mode
	if m.conditionMode {
		var title string
		var items []string
		if m.conditionAdding {
			title = "Add Condition (↑/↓ to select, Enter to add, Esc to cancel)"
			for i, cond := range allConditions {
				prefix := "  "
				if i == m.conditionCursor {
					prefix = "▶ "
				}
				items = append(items, prefix+string(cond))
			}
		} else {
			title = "Remove Condition (↑/↓ to select, Enter to remove, Esc to cancel)"
			for i, cond := range m.character.CombatStats.Conditions {
				prefix := "  "
				if i == m.conditionCursor {
					prefix = "▶ "
				}
				items = append(items, prefix+string(cond))
			}
		}

		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
		itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

		// Show only a window of items around the cursor
		start := m.conditionCursor - 3
		if start < 0 {
			start = 0
		}
		end := start + 7
		if end > len(items) {
			end = len(items)
		}
		visibleItems := items[start:end]

		return lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render(title),
			itemStyle.Render(strings.Join(visibleItems, "\n")),
		)
	}

	// Show HP input prompt if in input mode
	if m.hpInputMode != HPInputNone {
		var prompt string
		switch m.hpInputMode {
		case HPInputDamage:
			prompt = "Damage amount: "
		case HPInputHeal:
			prompt = "Heal amount: "
		case HPInputTemp:
			prompt = "Temp HP amount: "
		}
		inputStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true)
		inputLine := inputStyle.Render(prompt + m.hpInputBuffer + "█")
		helpLine := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("Enter to confirm • Esc to cancel")
		return lipgloss.JoinVertical(lipgloss.Left,
			inputLine,
			helpLine,
		)
	}

	// Show status message if present
	if m.statusMessage != "" {
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)
		return lipgloss.JoinVertical(lipgloss.Left,
			statusStyle.Render(m.statusMessage),
			footerStyle.Render(help),
		)
	}

	return footerStyle.Render(help)
}

// getWeapons returns weapons from the character's inventory.
func (m *MainSheetModel) getWeapons() []models.Item {
	var weapons []models.Item
	equip := &m.character.Inventory.Equipment
	
	// Only show equipped weapons (main hand and off hand)
	if equip.MainHand != nil && equip.MainHand.Type == models.ItemTypeWeapon && equip.MainHand.Damage != "" {
		weapons = append(weapons, *equip.MainHand)
	}
	if equip.OffHand != nil && equip.OffHand.Type == models.ItemTypeWeapon && equip.OffHand.Damage != "" {
		// Don't duplicate if same weapon in both hands
		if equip.MainHand == nil || equip.OffHand.ID != equip.MainHand.ID {
			weapons = append(weapons, *equip.OffHand)
		}
	}
	return weapons
}

// getWeaponAttackBonus calculates the attack bonus for a weapon.
func (m *MainSheetModel) getWeaponAttackBonus(weapon models.Item) int {
	char := m.character

	// Get ability modifier
	abilityMod := m.getWeaponAbilityMod(weapon)

	// Add proficiency bonus only if proficient
	profBonus := 0
	if m.isProficientWithWeapon(weapon) {
		profBonus = char.GetProficiencyBonus()
	}

	// Add magic bonus if any
	magicBonus := weapon.MagicBonus

	return abilityMod + profBonus + magicBonus
}

// getWeaponDamageMod returns the damage modifier for a weapon (ability mod + magic bonus).
func (m *MainSheetModel) getWeaponDamageMod(weapon models.Item) int {
	return m.getWeaponAbilityMod(weapon) + weapon.MagicBonus
}

// getWeaponAbilityMod returns the ability modifier used for a weapon.
func (m *MainSheetModel) getWeaponAbilityMod(weapon models.Item) int {
	char := m.character

	// Check for finesse property - use better of STR/DEX
	isFinesse := false
	isRanged := false
	for _, prop := range weapon.WeaponProps {
		if prop == "finesse" {
			isFinesse = true
		}
		if prop == "ammunition" {
			isRanged = true
		}
	}

	if isFinesse {
		strMod := char.AbilityScores.Strength.Modifier()
		dexMod := char.AbilityScores.Dexterity.Modifier()
		if dexMod > strMod {
			return dexMod
		}
		return strMod
	} else if isRanged {
		return char.AbilityScores.Dexterity.Modifier()
	}
	return char.AbilityScores.Strength.Modifier()
}

// isProficientWithWeapon checks if the character is proficient with the given weapon.
func (m *MainSheetModel) isProficientWithWeapon(weapon models.Item) bool {
	char := m.character
	weaponName := strings.ToLower(weapon.Name)
	subCategory := strings.ToLower(weapon.SubCategory)

	for _, prof := range char.Proficiencies.Weapons {
		profLower := strings.ToLower(prof)
		
		// Check for category proficiency (e.g., "Simple Weapons", "Martial Weapons")
		if profLower == "simple weapons" && strings.Contains(subCategory, "simple") {
			return true
		}
		if profLower == "martial weapons" && strings.Contains(subCategory, "martial") {
			return true
		}
		
		// Check for specific weapon proficiency (e.g., "Longsword", "Hand Crossbows")
		if strings.Contains(weaponName, strings.TrimSuffix(profLower, "s")) {
			return true
		}
		if strings.Contains(profLower, weaponName) {
			return true
		}
	}

	return false
}

// getSpellSaveDC calculates the spell save DC for the character.
func (m *MainSheetModel) getSpellSaveDC() int {
	if m.character == nil || m.character.Spellcasting == nil {
		return 10
	}

	sc := m.character.Spellcasting
	abilityMod := m.character.AbilityScores.GetModifier(sc.Ability)
	profBonus := m.character.Info.ProficiencyBonus()

	return models.CalculateSpellSaveDC(abilityMod, profBonus)
}

// wordWrap wraps text to the specified width.
func (m *MainSheetModel) wordWrap(text string, width int) []string {
	if width <= 0 {
		width = 40
	}

	var lines []string
	words := strings.Fields(text)

	if len(words) == 0 {
		return lines
	}

	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	lines = append(lines, currentLine)
	return lines
}

// calculateUpcastEffect calculates and formats the upcast effect for a given slot level.
func (m *MainSheetModel) calculateUpcastEffect(slotLevel int, spell *data.SpellData) string {
	if spell == nil {
		return ""
	}

	levelsAbove := slotLevel - spell.Level

	upcastLower := strings.ToLower(spell.Upcast)

	// Handle dart/projectile increases (like Magic Missile)
	if strings.Contains(upcastLower, "dart") {
		baseDarts := 3
		totalDarts := baseDarts + levelsAbove
		if spell.Damage != "" {
			return fmt.Sprintf(" (%d darts, %s each)", totalDarts, spell.Damage)
		}
		return fmt.Sprintf(" (%d darts)", totalDarts)
	}

	// Handle target increases (like Scorching Ray)
	if strings.Contains(upcastLower, "ray") || strings.Contains(upcastLower, "beam") {
		var bonusPerLevel int
		fmt.Sscanf(spell.Upcast, "+%d", &bonusPerLevel)
		if bonusPerLevel == 0 {
			bonusPerLevel = 1
		}

		baseRays := 3
		totalRays := baseRays + (bonusPerLevel * levelsAbove)
		if spell.Damage != "" {
			return fmt.Sprintf(" (%d rays, %s each)", totalRays, spell.Damage)
		}
		return fmt.Sprintf(" (%d rays)", totalRays)
	}

	// Handle dice damage increases (like Burning Hands, Fireball)
	var diceStr string
	for _, part := range strings.Fields(spell.Upcast) {
		if strings.Contains(part, "d") && len(part) >= 3 {
			cleaned := strings.TrimPrefix(strings.TrimPrefix(part, "+"), " ")
			dIndex := strings.Index(cleaned, "d")
			if dIndex > 0 && dIndex < len(cleaned)-1 {
				beforeD := cleaned[dIndex-1]
				afterD := cleaned[dIndex+1]
				if beforeD >= '0' && beforeD <= '9' && afterD >= '0' && afterD <= '9' {
					diceStr = cleaned
					break
				}
			}
		}
	}

	if diceStr != "" {
		parts := strings.Split(diceStr, "d")
		if len(parts) == 2 {
			var dicePerLevel int
			fmt.Sscanf(parts[0], "%d", &dicePerLevel)
			if dicePerLevel == 0 {
				dicePerLevel = 1
			}

			diceType := strings.TrimRight(parts[1], ".,;:")

			if spell.Damage != "" {
				baseParts := strings.Split(spell.Damage, "d")
				if len(baseParts) == 2 {
					var baseDice int
					fmt.Sscanf(baseParts[0], "%d", &baseDice)
					totalDice := baseDice + (dicePerLevel * levelsAbove)
					return fmt.Sprintf(" (%dd%s damage)", totalDice, diceType)
				}
				if levelsAbove > 0 {
					bonusDice := dicePerLevel * levelsAbove
					return fmt.Sprintf(" (%s + %dd%s damage)", spell.Damage, bonusDice, diceType)
				}
				return fmt.Sprintf(" (%s damage)", spell.Damage)
			}
			if levelsAbove > 0 {
				bonusDice := dicePerLevel * levelsAbove
				return fmt.Sprintf(" (+%dd%s damage)", bonusDice, diceType)
			}
		}
	}

	// Handle healing increases
	if strings.Contains(upcastLower, "heal") {
		if spell.Damage != "" {
			return fmt.Sprintf(" (%s healing)", spell.Damage)
		}
	}

	// Fallback: show base damage if available
	if spell.Damage != "" {
		return fmt.Sprintf(" (%s damage)", spell.Damage)
	}

	// If upcasting and we have no other info, show generic upcast indicator
	if levelsAbove > 0 {
		return fmt.Sprintf(" (upcast +%d)", levelsAbove)
	}

	return ""
}

// getAvailableCastLevels returns the spell slot levels available for casting a spell.
func (m *MainSheetModel) getAvailableCastLevels(spellLevel int) []int {
	if m.character == nil || m.character.Spellcasting == nil {
		return []int{}
	}

	sc := m.character.Spellcasting
	var available []int

	// Check pact magic first
	if sc.PactMagic != nil && sc.PactMagic.Total > 0 && sc.PactMagic.Remaining > 0 {
		if sc.PactMagic.SlotLevel >= spellLevel {
			available = append(available, sc.PactMagic.SlotLevel)
		}
	}

	// Check regular spell slots from spell level to 9
	for level := spellLevel; level <= 9; level++ {
		slot := sc.SpellSlots.GetSlot(level)
		if slot != nil && slot.Remaining > 0 {
			available = append(available, level)
		}
	}

	return available
}

// formatModifier formats an integer as a modifier string (e.g., +2 or -1).
func formatModifier(mod int) string {
	if mod >= 0 {
		return fmt.Sprintf("+%d", mod)
	}
	return fmt.Sprintf("%d", mod)
}

// BackToSelectionMsg is sent when the user wants to return to character selection.
type BackToSelectionMsg struct{}
