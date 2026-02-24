package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/Domo929/sheet/internal/ui/components"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// SpellbookMode represents the current mode of the spellbook.
type SpellbookMode int

const (
	ModeSpellList SpellbookMode = iota // Viewing/casting spells
	ModePreparation                     // Preparing/unpreparing spells
	ModeAddSpell                        // Adding a new spell
	ModeSelectCastLevel                 // Selecting spell slot level for casting
	ModeConfirmCast                     // Confirming spell cast with details
)

// SpellbookModel is the model for the spellbook view.
type SpellbookModel struct {
	character     *models.Character
	storage       *storage.CharacterStorage
	loader        *data.Loader
	width         int
	height        int
	mode          SpellbookMode
	keys          spellbookKeyMap
	statusMessage string

	// Quit confirmation
	confirmingQuit bool

	// Spell list navigation
	spellCursor       int
	spellScroll       int
	spellsPerPage     int
	selectedSpellData *data.SpellData // Full spell data from spells.json

	// Filter state
	filterLevel int // -1 = all, 0 = cantrips, 1-9 = spell levels

	// Add spell mode
	spellSearchTerm string
	searchResults   []data.SpellData
	searchCursor    int

	// Delete confirmation
	confirmingRemove bool
	removeSpellName  string

	// Spell casting level selection
	castingSpell      *models.KnownSpell // Spell being cast
	castLevelCursor   int                // Cursor for slot level selection
	availableCastLevels []int            // Available slot levels for upcasting

	// Roll history layout
	rollHistoryVisible bool
	rollHistoryWidth   int

	// Spell data cache
	spellDatabase *data.SpellDatabase
}

type spellbookKeyMap struct {
	Quit          key.Binding
	ForceQuit     key.Binding
	Up            key.Binding
	Down          key.Binding
	Enter         key.Binding
	Back          key.Binding
	Prepare       key.Binding
	Cast          key.Binding
	Add           key.Binding
	Remove        key.Binding
	Filter        key.Binding
	CustomRoll    key.Binding
	HistoryToggle key.Binding
}

func defaultSpellbookKeyMap() spellbookKeyMap {
	return spellbookKeyMap{
		Quit:      key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "force quit")),
		Up:        key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:      key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Enter:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "cast/select")),
		Back:      key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Prepare:   key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "prepare spells")),
		Cast:      key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "cast spell")),
		Add:       key.NewBinding(key.WithKeys("a", "+"), key.WithHelp("a/+", "add spell")),
		Remove:    key.NewBinding(key.WithKeys("x", "delete"), key.WithHelp("x", "remove spell")),
		Filter:        key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "filter level")),
		CustomRoll:    key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "roll dice")),
		HistoryToggle: key.NewBinding(key.WithKeys("H"), key.WithHelp("H", "history")),
	}
}

// SetRollHistoryState updates the spellbook's knowledge of roll history visibility.
func (m *SpellbookModel) SetRollHistoryState(visible bool, width int) {
	m.rollHistoryVisible = visible
	m.rollHistoryWidth = width
}

// NewSpellbookModel creates a new spellbook model.
func NewSpellbookModel(char *models.Character, storage *storage.CharacterStorage, loader *data.Loader) *SpellbookModel {
	return &SpellbookModel{
		character:     char,
		storage:       storage,
		loader:        loader,
		keys:          defaultSpellbookKeyMap(),
		spellsPerPage: 15,
		filterLevel:   -1, // Show all by default
		mode:          ModeSpellList,
	}
}

// Init initializes the model.
func (m *SpellbookModel) Init() tea.Cmd {
	// Load spell database
	return func() tea.Msg {
		db, err := m.loader.GetSpells()
		if err != nil {
			return spellDatabaseLoadedMsg{err: err}
		}
		return spellDatabaseLoadedMsg{database: db}
	}
}

// spellDatabaseLoadedMsg is sent when spell database is loaded.
type spellDatabaseLoadedMsg struct {
	database *data.SpellDatabase
	err      error
}

// Update handles messages.
func (m *SpellbookModel) Update(msg tea.Msg) (*SpellbookModel, tea.Cmd) {
	switch msg := msg.(type) {
	case spellDatabaseLoadedMsg:
		if msg.err != nil {
			m.statusMessage = fmt.Sprintf("Error loading spells: %v", msg.err)
			return m, nil
		}
		m.spellDatabase = msg.database
		m.updateSelectedSpellData()
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
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

		// Handle remove confirmation
		if m.confirmingRemove {
			switch msg.String() {
			case "y", "Y":
				m.performRemoveSpell()
				return m, nil
			default:
				m.confirmingRemove = false
				m.removeSpellName = ""
				m.statusMessage = "Remove cancelled"
				return m, nil
			}
		}

		// Handle add spell mode
		if m.mode == ModeAddSpell {
			return m.handleAddSpellInput(msg)
		}

		// Handle spell cast level selection
		if m.mode == ModeSelectCastLevel {
			return m.handleCastLevelInput(msg)
		}

		// Handle confirm cast mode
		if m.mode == ModeConfirmCast {
			switch {
			case key.Matches(msg, m.keys.Back): // Esc
				// Cancel casting
				m.mode = ModeSpellList
				m.castingSpell = nil
				m.availableCastLevels = nil
				m.statusMessage = "Casting cancelled"
				return m, nil

			case key.Matches(msg, m.keys.Up):
				// Navigate up in slot selection
				if len(m.availableCastLevels) > 1 && m.castLevelCursor > 0 {
					m.castLevelCursor--
				}
				return m, nil

			case key.Matches(msg, m.keys.Down):
				// Navigate down in slot selection
				if len(m.availableCastLevels) > 1 && m.castLevelCursor < len(m.availableCastLevels)-1 {
					m.castLevelCursor++
				}
				return m, nil

			case key.Matches(msg, m.keys.Enter):
				// Confirm cast
				if m.castingSpell.Level == 0 {
					// Cantrip - no slot needed
					m.statusMessage = fmt.Sprintf("Cast %s (no slot required)", m.castingSpell.Name)
					m.mode = ModeSpellList
					m.castingSpell = nil
					return m, m.saveCharacter()
				} else if len(m.availableCastLevels) > 0 {
					selectedLevel := m.availableCastLevels[m.castLevelCursor]
					if selectedLevel == 0 {
						// Ritual cast (sentinel value 0) - no slot needed, takes 10 extra minutes
						m.statusMessage = fmt.Sprintf("Cast %s as ritual (no slot required, takes 10 extra minutes)", m.castingSpell.Name)
						m.mode = ModeSpellList
						m.castingSpell = nil
						return m, m.saveCharacter()
					}
					// Cast with selected slot level
					m.mode = ModeSpellList
					result, rollCmd := m.castSpellAtLevel(m.castingSpell, selectedLevel)
					saveCmd := m.saveCharacter()
					if rollCmd != nil {
						return result, tea.Batch(saveCmd, rollCmd)
					}
					return result, saveCmd
				} else {
					// No slots available (shouldn't reach here)
					m.statusMessage = "No spell slots available"
					m.mode = ModeSpellList
					m.castingSpell = nil
					return m, nil
				}
			}
			return m, nil
		}

		// Ctrl+C always quits immediately
		if key.Matches(msg, m.keys.ForceQuit) {
			return m, tea.Quit
		}

		m.statusMessage = ""

		switch {
		case key.Matches(msg, m.keys.Quit):
			m.confirmingQuit = true
			m.statusMessage = "Quit? (y/n)"
			return m, nil
		case key.Matches(msg, m.keys.Back):
			if m.mode == ModeSelectCastLevel {
				// Cancel spell casting
				m.mode = ModeSpellList
				m.castingSpell = nil
				m.availableCastLevels = nil
				m.statusMessage = "Casting cancelled"
				return m, nil
			}
			if m.mode == ModePreparation {
				// Return to spell list mode
				m.mode = ModeSpellList
				m.spellCursor = 0
				m.spellScroll = 0
				m.updateSelectedSpellData()
				return m, nil
			}
			return m, func() tea.Msg { return BackToSheetMsg{} }
		case key.Matches(msg, m.keys.Up):
			return m.handleUp(), nil
		case key.Matches(msg, m.keys.Down):
			return m.handleDown(), nil
		case key.Matches(msg, m.keys.Prepare):
			if m.mode == ModeSpellList {
				// Enter preparation mode
				m.mode = ModePreparation
				m.spellCursor = 0
				m.spellScroll = 0
				m.updateSelectedSpellData()
				m.statusMessage = "Preparation Mode - [✓]=prepared [●]=always prepared  p:toggle esc:exit"
				return m, nil
			} else if m.mode == ModePreparation {
				// Toggle preparation of current spell
				return m.handlePrepareToggle(), m.saveCharacter()
			}
			return m, nil
		case key.Matches(msg, m.keys.Cast), key.Matches(msg, m.keys.Enter):
			if m.mode == ModeSpellList {
				return m.handleCastSpell(), m.saveCharacter()
			} else if m.mode == ModePreparation {
				// In preparation mode, Enter also toggles preparation
				return m.handlePrepareToggle(), m.saveCharacter()
			}
			return m, nil
		case key.Matches(msg, m.keys.Add):
			m.mode = ModeAddSpell
			m.spellSearchTerm = ""
			m.searchResults = []data.SpellData{}
			m.searchCursor = 0
			return m, nil
		case key.Matches(msg, m.keys.Remove):
			if m.mode == ModePreparation {
				return m.handleRemoveSpell(), nil
			}
			return m, nil
		case key.Matches(msg, m.keys.Filter):
			return m.handleFilterToggle(), nil
		case key.Matches(msg, m.keys.CustomRoll):
			return m, func() tea.Msg { return components.OpenCustomRollMsg{} }
		case key.Matches(msg, m.keys.HistoryToggle):
			return m, func() tea.Msg { return components.ToggleRollHistoryMsg{} }
		}
	}

	return m, nil
}

// View renders the spellbook.
func (m *SpellbookModel) View() string {
	if m.width == 0 {
		return "Loading spellbook..."
	}

	if m.spellDatabase == nil {
		return "Loading spell data..."
	}

	// Build header
	header := m.renderHeader()

	// Build spell list and details panels
	spellList := m.renderSpellList()
	spellDetails := m.renderSpellDetails()

	// Build spell slots panel
	spellSlots := m.renderSpellSlots()

	const compactBreakpoint = 90

	// Subtract roll history width if visible
	availableWidth := m.width
	if m.rollHistoryVisible {
		availableWidth -= m.rollHistoryWidth
	}

	panelHeight := m.height - lipgloss.Height(header) - 2

	var panels string

	if availableWidth < compactBreakpoint {
		// Compact: two panels (list + details), spell slots hidden
		listWidth := availableWidth * 40 / 100 // 40%
		if listWidth < 25 {
			listWidth = 25
		}
		detailsWidth := availableWidth - listWidth

		spellListStyled := lipgloss.NewStyle().
			Width(listWidth).
			Height(panelHeight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12")).
			Render(spellList)

		spellDetailsStyled := lipgloss.NewStyle().
			Width(detailsWidth).
			Height(panelHeight).
			Border(lipgloss.RoundedBorder()).
			Render(spellDetails)

		panels = lipgloss.JoinHorizontal(lipgloss.Top, spellListStyled, spellDetailsStyled)
	} else {
		// Standard: three panels (list | details | slots)
		listWidth := availableWidth / 3
		detailsWidth := availableWidth / 3
		slotsWidth := availableWidth - listWidth - detailsWidth

		spellListStyled := lipgloss.NewStyle().
			Width(listWidth).
			Height(panelHeight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12")).
			Render(spellList)

		spellDetailsStyled := lipgloss.NewStyle().
			Width(detailsWidth).
			Height(panelHeight).
			Border(lipgloss.RoundedBorder()).
			Render(spellDetails)

		spellSlotsStyled := lipgloss.NewStyle().
			Width(slotsWidth).
			Height(panelHeight).
			Border(lipgloss.RoundedBorder()).
			Render(spellSlots)

		panels = lipgloss.JoinHorizontal(lipgloss.Top, spellListStyled, spellDetailsStyled, spellSlotsStyled)
	}

	// Build footer
	footer := m.renderFooter()

	// Handle add spell overlay
	if m.mode == ModeAddSpell {
		overlay := m.renderAddSpellOverlay()
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay)
	}

	// Handle cast level selection overlay
	if m.mode == ModeSelectCastLevel {
		overlay := m.renderCastLevelOverlay()
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay)
	}

	// Handle confirm cast overlay
	if m.mode == ModeConfirmCast {
		overlay := m.renderCastConfirmationModal()
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, panels, footer)
}

// renderHeader renders the spellbook header with spell stats.
func (m *SpellbookModel) renderHeader() string {
	if m.character == nil || m.character.Spellcasting == nil {
		return lipgloss.NewStyle().Bold(true).Render("Spellbook") + "\n"
	}

	sc := m.character.Spellcasting

	// Calculate spell stats
	abilityMod := m.character.AbilityScores.GetModifier(sc.Ability)
	profBonus := m.character.Info.ProficiencyBonus()
	saveDC := models.CalculateSpellSaveDC(abilityMod, profBonus)
	attackBonus := models.CalculateSpellAttackBonus(abilityMod, profBonus)

	modeTitle := "Spellbook"
	if m.mode == ModePreparation {
		if sc.PreparesSpells {
			modeTitle = "Spellbook - Preparation Mode"
		} else {
			modeTitle = "Spellbook - Manage Spells"
		}
	}
	title := lipgloss.NewStyle().Bold(true).Render(modeTitle)

	// Capitalize the ability name for display
	abilityDisplay := strings.Title(strings.ToLower(string(sc.Ability)))
	stats := fmt.Sprintf("Ability: %s | Save DC: %d | Attack: +%d",
		abilityDisplay, saveDC, attackBonus)

	// Show prepared spell count for preparing classes, known spell count for others
	prepInfo := ""
	if sc.PreparesSpells && sc.MaxPrepared > 0 {
		prepCount := sc.CountPreparedSpells()
		prepInfo = fmt.Sprintf(" | Prepared Spells: %d/%d", prepCount, sc.MaxPrepared)
	} else if len(sc.KnownSpells) > 0 {
		// Non-preparing classes (Warlock, Bard, Sorcerer): show known spells count
		prepInfo = fmt.Sprintf(" | Known Spells: %d", len(sc.KnownSpells))
	}

	// Show cantrip count
	cantripInfo := ""
	if len(sc.CantripsKnown) > 0 {
		cantripInfo = fmt.Sprintf(" | Cantrips: %d", len(sc.CantripsKnown))
	}

	statsStyled := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Render(stats + prepInfo + cantripInfo)

	return lipgloss.JoinVertical(lipgloss.Left, title, statsStyled) + "\n"
}

// renderSpellList renders the spell list panel.
func (m *SpellbookModel) renderSpellList() string {
	if m.character == nil || m.character.Spellcasting == nil {
		return "No spellcasting ability"
	}

	sc := m.character.Spellcasting
	var lines []string

	if m.mode == ModePreparation {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render("=== Preparation Mode ==="))
		lines = append(lines, "Press 'p' or Enter to toggle")
		lines = append(lines, "Press 'x' to remove spell")
		lines = append(lines, "")
	}

	// Get the display list (prepared only in spell list mode, all in preparation mode)
	displaySpells := m.getDisplaySpells()

	if len(displaySpells) == 0 && m.mode == ModeSpellList {
		lines = append(lines, "No spells prepared")
		lines = append(lines, "Press 'p' to prepare spells")
		return strings.Join(lines, "\n")
	}

	// Group spells by level
	spellsByLevel := make(map[int][]models.KnownSpell)
	for _, spell := range displaySpells {
		spellsByLevel[spell.Level] = append(spellsByLevel[spell.Level], spell)
	}

	// Sort levels
	levels := make([]int, 0, len(spellsByLevel))
	for level := range spellsByLevel {
		levels = append(levels, level)
	}
	sort.Ints(levels)

	currentIndex := 0

	for _, level := range levels {
		spells := spellsByLevel[level]

		// Sort spells alphabetically within level
		sort.Slice(spells, func(i, j int) bool {
			return spells[i].Name < spells[j].Name
		})

		// Header for level
		levelHeader := fmt.Sprintf("Level %d", level)
		if level == 0 {
			levelHeader = "Cantrips"
		}
		lines = append(lines, lipgloss.NewStyle().Bold(true).Render(levelHeader))

		for _, spell := range spells {
			cursor := "  "
			if currentIndex == m.spellCursor {
				cursor = "> "
			}

			prepMarker := " "
			if m.mode == ModePreparation {
				// For non-preparing classes (Warlock, Bard, Sorcerer),
				// all known spells are always "prepared" (available)
				if !sc.PreparesSpells {
					prepMarker = "✓"
				} else if spell.AlwaysPrepared {
					// Always prepared from class features (can't be unprepared)
					prepMarker = "●"
				} else if spell.Prepared {
					prepMarker = "✓"
				} else if spell.Ritual && sc.RitualCasterUnprepared {
					// Wizards can cast ritual spells without preparing them
					prepMarker = "●"
				} else {
					prepMarker = " "
				}
			}

			ritualMarker := ""
			if spell.Ritual {
				ritualMarker = " (R)"
			}

			var line string
			if m.mode == ModePreparation {
				line = fmt.Sprintf("%s[%s] %s%s", cursor, prepMarker, spell.Name, ritualMarker)
			} else {
				line = fmt.Sprintf("%s%s%s", cursor, spell.Name, ritualMarker)
			}

			if currentIndex == m.spellCursor {
				line = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(line)
			}

			lines = append(lines, line)
			currentIndex++
		}
		lines = append(lines, "")
	}

	if len(displaySpells) == 0 && len(sc.CantripsKnown) == 0 {
		lines = append(lines, "No spells known")
		lines = append(lines, "Press 'a' to add a spell")
	}

	return strings.Join(lines, "\n")
}

// getSpellSaveDC calculates the spell save DC for the character.
func (m *SpellbookModel) getSpellSaveDC() int {
	if m.character == nil || m.character.Spellcasting == nil {
		return 10
	}

	sc := m.character.Spellcasting
	abilityMod := m.character.AbilityScores.GetModifier(sc.Ability)
	profBonus := m.character.Info.ProficiencyBonus()

	return models.CalculateSpellSaveDC(abilityMod, profBonus)
}

// renderSpellDetails renders the selected spell's details.
func (m *SpellbookModel) renderSpellDetails() string {
	if m.selectedSpellData == nil {
		return "Select a spell to see details"
	}

	spell := m.selectedSpellData
	var lines []string

	lines = append(lines, lipgloss.NewStyle().Bold(true).Render(spell.Name))

	levelSchool := fmt.Sprintf("Level %d %s", spell.Level, spell.School)
	if spell.Level == 0 {
		levelSchool = fmt.Sprintf("%s cantrip", spell.School)
	}
	if spell.Ritual {
		levelSchool += " (ritual)"
	}
	lines = append(lines, levelSchool)
	lines = append(lines, "")

	lines = append(lines, fmt.Sprintf("Casting Time: %s", spell.CastingTime))
	lines = append(lines, fmt.Sprintf("Range: %s", spell.Range))
	lines = append(lines, fmt.Sprintf("Components: %s", strings.Join(data.ComponentsToStrings(spell.Components), ", ")))
	lines = append(lines, fmt.Sprintf("Duration: %s", spell.Duration))

	// Add damage if present
	if spell.Damage != "" {
		damageInfo := spell.Damage
		if spell.DamageType != "" {
			damageInfo = fmt.Sprintf("%s %s", damageInfo, spell.DamageType)
		}
		lines = append(lines, fmt.Sprintf("Damage: %s", damageInfo))
	}

	// Add saving throw if present
	if spell.SavingThrow != "" {
		saveDC := m.getSpellSaveDC()
		lines = append(lines, fmt.Sprintf("Saving Throw: %s DC %d", spell.SavingThrow, saveDC))
	}

	lines = append(lines, "")

	// Word wrap description based on available width
	detailsWrapWidth := m.detailsPanelWidth() - 4
	descLines := m.wordWrap(spell.Description, detailsWrapWidth)
	lines = append(lines, descLines...)

	return strings.Join(lines, "\n")
}

// renderSpellSlots renders the spell slots panel.
func (m *SpellbookModel) renderSpellSlots() string {
	if m.character == nil || m.character.Spellcasting == nil {
		return "No spell slots"
	}

	sc := m.character.Spellcasting
	var lines []string

	lines = append(lines, lipgloss.NewStyle().Bold(true).Render("Spell Slots"))
	lines = append(lines, "")

	// Show pact magic if applicable
	if sc.PactMagic != nil && sc.PactMagic.Total > 0 {
		pm := sc.PactMagic
		slotsDisplay := m.renderSlotBar(pm.Remaining, pm.Total)
		lines = append(lines, fmt.Sprintf("Pact Magic (Level %d)", pm.SlotLevel))
		lines = append(lines, slotsDisplay)
		lines = append(lines, "")
	}

	// Show regular spell slots
	for level := 1; level <= 9; level++ {
		slot := sc.SpellSlots.GetSlot(level)
		if slot != nil && slot.Total > 0 {
			slotsDisplay := m.renderSlotBar(slot.Remaining, slot.Total)
			lines = append(lines, fmt.Sprintf("Level %d:", level))
			lines = append(lines, slotsDisplay)
			lines = append(lines, "")
		}
	}

	if len(lines) == 2 { // Only header
		lines = append(lines, "No spell slots")
	}

	return strings.Join(lines, "\n")
}

// renderSlotBar renders a visual bar for spell slots.
func (m *SpellbookModel) renderSlotBar(remaining, total int) string {
	if total == 0 {
		return ""
	}

	bar := ""
	for i := 0; i < total; i++ {
		if i < remaining {
			bar += "●"
		} else {
			bar += "○"
		}
	}

	return fmt.Sprintf("%s %d/%d", bar, remaining, total)
}

// renderFooter renders the help footer.
func (m *SpellbookModel) renderFooter() string {
	if m.statusMessage != "" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(m.statusMessage)
	}

	var helps []string
	helps = append(helps, "esc: back")
	helps = append(helps, "↑↓: navigate")

	if m.mode == ModeSpellList {
		if m.character.Spellcasting != nil {
			if m.character.Spellcasting.PreparesSpells {
				helps = append(helps, "p: prepare spells")
			} else {
				helps = append(helps, "p: manage spells")
			}
		}
		helps = append(helps, "c/enter: cast")
	} else if m.mode == ModePreparation {
		helps = append(helps, "p/enter: toggle")
		helps = append(helps, "x: remove")
	}

	helps = append(helps, "a: add spell")
	helps = append(helps, "f: filter")
	helps = append(helps, "/: roll dice")
	helps = append(helps, "H: history")
	helps = append(helps, "q: quit")

	return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Join(helps, " | "))
}

// renderAddSpellOverlay renders the add spell modal.
func (m *SpellbookModel) renderAddSpellOverlay() string {
	var lines []string

	lines = append(lines, lipgloss.NewStyle().Bold(true).Render("Add Spell"))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Search: %s_", m.spellSearchTerm))
	lines = append(lines, "")

	if len(m.searchResults) == 0 && m.spellSearchTerm != "" {
		lines = append(lines, "No spells found")
	} else {
		for i, spell := range m.searchResults {
			cursor := "  "
			if i == m.searchCursor {
				cursor = "> "
			}

			levelStr := fmt.Sprintf("L%d", spell.Level)
			if spell.Level == 0 {
				levelStr = "Can"
			}

			line := fmt.Sprintf("%s[%s] %s (%s)", cursor, levelStr, spell.Name, spell.School)

			if i == m.searchCursor {
				line = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(line)
			}

			lines = append(lines, line)
		}
	}

	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Type to search | Enter: add | Esc: cancel"))

	content := strings.Join(lines, "\n")

	width := 60
	height := len(lines) + 2

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Render(content)
}

// Helper functions

func (m *SpellbookModel) handleUp() *SpellbookModel {
	if m.spellCursor > 0 {
		m.spellCursor--
		m.updateSelectedSpellData()
	}
	return m
}

func (m *SpellbookModel) handleDown() *SpellbookModel {
	displaySpells := m.getDisplaySpells()
	if m.spellCursor < len(displaySpells)-1 {
		m.spellCursor++
		m.updateSelectedSpellData()
	}
	return m
}

func (m *SpellbookModel) handlePrepareToggle() *SpellbookModel {
	if m.character == nil || m.character.Spellcasting == nil {
		m.statusMessage = "No spellcasting ability"
		return m
	}

	sc := m.character.Spellcasting

	if !sc.PreparesSpells {
		m.statusMessage = "This class doesn't prepare spells"
		return m
	}

	displaySpells := m.getDisplaySpells()
	if m.spellCursor >= len(displaySpells) {
		return m
	}

	spell := displaySpells[m.spellCursor]

	// Can't toggle always-prepared spells
	if spell.AlwaysPrepared {
		m.statusMessage = fmt.Sprintf("%s is always prepared (from class feature)", spell.Name)
		return m
	}

	// Can't toggle ritual spells for Wizards (they can cast without preparing)
	if spell.Ritual && sc.RitualCasterUnprepared {
		m.statusMessage = fmt.Sprintf("%s is a ritual spell (can cast from spellbook)", spell.Name)
		return m
	}

	// Check if we can prepare more spells
	if !spell.Prepared && sc.MaxPrepared > 0 {
		if sc.CountPreparedSpells() >= sc.MaxPrepared {
			m.statusMessage = fmt.Sprintf("Cannot prepare more than %d spells", sc.MaxPrepared)
			return m
		}
	}

	// Toggle prepared state
	newState := !spell.Prepared
	sc.PrepareSpell(spell.Name, newState)

	if newState {
		m.statusMessage = fmt.Sprintf("Prepared %s", spell.Name)
	} else {
		m.statusMessage = fmt.Sprintf("Unprepared %s", spell.Name)
	}

	return m
}

func (m *SpellbookModel) handleCastSpell() *SpellbookModel {
	if m.character == nil || m.character.Spellcasting == nil {
		m.statusMessage = "No spellcasting ability"
		return m
	}

	sc := m.character.Spellcasting

	// Check if a spell is selected
	displaySpells := m.getDisplaySpells()
	if m.spellCursor >= len(displaySpells) {
		return m
	}

	spell := displaySpells[m.spellCursor]

	// Check if spell is prepared or can be cast (if applicable)
	if sc.PreparesSpells {
		canCast := spell.Prepared || spell.AlwaysPrepared

		// Wizards can cast ritual spells from spellbook without preparing
		if spell.Ritual && sc.RitualCasterUnprepared {
			canCast = true
		}

		if !canCast {
			m.statusMessage = fmt.Sprintf("%s is not prepared", spell.Name)
			return m
		}
	}

	// Check if spell data is available in database (only warn, don't block)
	// This allows tests to work without full spell database loaded
	if m.selectedSpellData == nil && m.spellDatabase != nil {
		m.statusMessage = fmt.Sprintf("Warning: %s data not found in spell database", spell.Name)
		// Continue anyway - the modal will handle nil gracefully
	}

	// Store casting spell and get available levels
	m.castingSpell = &spell

	// Cantrips don't use slots
	if spell.Level == 0 {
		m.availableCastLevels = []int{}
	} else if spell.Ritual {
		// Ritual spells can be cast as a ritual (no slot, +10 min) OR with a spell slot
		// Use 0 as sentinel value for "cast as ritual" option
		m.availableCastLevels = append([]int{0}, m.getAvailableCastLevels(spell.Level)...)
	} else {
		// Find available spell slot levels
		m.availableCastLevels = m.getAvailableCastLevels(spell.Level)

		if len(m.availableCastLevels) == 0 {
			m.statusMessage = fmt.Sprintf("No spell slots available for %s", spell.Name)
			m.castingSpell = nil
			return m
		}
	}

	// Initialize cursor
	m.castLevelCursor = 0

	// Always enter confirmation modal mode
	m.mode = ModeConfirmCast
	return m
}

func (m *SpellbookModel) handleRemoveSpell() *SpellbookModel {
	displaySpells := m.getDisplaySpells()
	if m.spellCursor >= len(displaySpells) {
		return m
	}

	spell := displaySpells[m.spellCursor]
	m.confirmingRemove = true
	m.removeSpellName = spell.Name
	m.statusMessage = fmt.Sprintf("Remove %s? (y/n)", spell.Name)
	return m
}

func (m *SpellbookModel) performRemoveSpell() {
	if m.character == nil || m.character.Spellcasting == nil {
		return
	}

	sc := m.character.Spellcasting

	// Remove from known spells
	for i, spell := range sc.KnownSpells {
		if spell.Name == m.removeSpellName {
			sc.KnownSpells = append(sc.KnownSpells[:i], sc.KnownSpells[i+1:]...)
			m.statusMessage = fmt.Sprintf("Removed %s", m.removeSpellName)

			// Adjust cursor if needed
			displaySpells := m.getDisplaySpells()
			if m.spellCursor >= len(displaySpells) && m.spellCursor > 0 {
				m.spellCursor--
			}

			m.updateSelectedSpellData()
			m.saveCharacter()
			break
		}
	}

	m.confirmingRemove = false
	m.removeSpellName = ""
}

func (m *SpellbookModel) handleFilterToggle() *SpellbookModel {
	// Cycle through filter levels: all -> cantrips -> 1 -> 2 -> ... -> 9 -> all
	m.filterLevel++
	if m.filterLevel > 9 {
		m.filterLevel = -1
	}

	if m.filterLevel == -1 {
		m.statusMessage = "Showing all spells"
	} else if m.filterLevel == 0 {
		m.statusMessage = "Showing cantrips"
	} else {
		m.statusMessage = fmt.Sprintf("Showing level %d spells", m.filterLevel)
	}

	// Reset cursor
	m.spellCursor = 0
	m.spellScroll = 0
	m.updateSelectedSpellData()

	return m
}

func (m *SpellbookModel) handleAddSpellInput(msg tea.KeyPressMsg) (*SpellbookModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = m.getPreviousMode()
		m.spellSearchTerm = ""
		m.searchResults = []data.SpellData{}
		m.statusMessage = "Add spell cancelled"
		return m, nil

	case "enter":
		if len(m.searchResults) > 0 && m.searchCursor < len(m.searchResults) {
			spell := m.searchResults[m.searchCursor]
			m.addSpellToCharacter(spell)
			m.mode = m.getPreviousMode()
			m.spellSearchTerm = ""
			m.searchResults = []data.SpellData{}
			return m, m.saveCharacter()
		}
		return m, nil

	case "up":
		if m.searchCursor > 0 {
			m.searchCursor--
		}
		return m, nil

	case "down":
		if m.searchCursor < len(m.searchResults)-1 {
			m.searchCursor++
		}
		return m, nil

	case "backspace":
		if len(m.spellSearchTerm) > 0 {
			m.spellSearchTerm = m.spellSearchTerm[:len(m.spellSearchTerm)-1]
			m.updateSearchResults()
		}
		return m, nil

	default:
		// Add character to search term
		if len(msg.String()) == 1 {
			m.spellSearchTerm += msg.String()
			m.updateSearchResults()
		}
		return m, nil
	}
}

func (m *SpellbookModel) getPreviousMode() SpellbookMode {
	// When exiting add spell mode, return to preparation mode if we were there
	// Otherwise return to spell list mode
	if m.mode == ModeAddSpell {
		// Check if we should return to preparation mode or spell list mode
		// For now, always return to preparation mode since that's where adding makes most sense
		return ModePreparation
	}
	return ModeSpellList
}

func (m *SpellbookModel) updateSearchResults() {
	if m.spellDatabase == nil || m.character == nil {
		return
	}

	m.searchResults = []data.SpellData{}
	m.searchCursor = 0

	if m.spellSearchTerm == "" {
		return
	}

	// Get character's class for filtering
	className := m.character.Info.Class

	searchLower := strings.ToLower(m.spellSearchTerm)

	for _, spell := range m.spellDatabase.Spells {
		// Check if spell matches search term
		if !strings.Contains(strings.ToLower(spell.Name), searchLower) {
			continue
		}

		// Check if spell is available to character's class
		classMatch := false
		for _, class := range spell.Classes {
			if strings.EqualFold(class, className) {
				classMatch = true
				break
			}
		}

		if !classMatch {
			continue
		}

		m.searchResults = append(m.searchResults, spell)

		// Limit results
		if len(m.searchResults) >= 10 {
			break
		}
	}
}

func (m *SpellbookModel) addSpellToCharacter(spell data.SpellData) {
	if m.character == nil || m.character.Spellcasting == nil {
		return
	}

	sc := m.character.Spellcasting

	// Check if spell is already known
	for _, known := range sc.KnownSpells {
		if known.Name == spell.Name {
			m.statusMessage = fmt.Sprintf("%s already known", spell.Name)
			return
		}
	}

	// Add as cantrip or spell
	if spell.Level == 0 {
		sc.AddCantrip(spell.Name)
		m.statusMessage = fmt.Sprintf("Added cantrip: %s", spell.Name)
	} else {
		sc.AddSpell(spell.Name, spell.Level)
		// Mark as ritual if applicable
		for i := range sc.KnownSpells {
			if sc.KnownSpells[i].Name == spell.Name {
				sc.KnownSpells[i].Ritual = spell.Ritual
				break
			}
		}
		m.statusMessage = fmt.Sprintf("Added spell: %s", spell.Name)
	}

	m.updateSelectedSpellData()
}

// getDisplaySpells returns the spells to display based on mode and filter.
// In spell list mode, includes cantrips as level 0 "spells" for casting.
func (m *SpellbookModel) getDisplaySpells() []models.KnownSpell {
	if m.character == nil || m.character.Spellcasting == nil {
		return []models.KnownSpell{}
	}

	sc := m.character.Spellcasting
	var display []models.KnownSpell

	// Include cantrips in both spell list mode and preparation mode
	if m.mode == ModeSpellList || m.mode == ModePreparation {
		// Add cantrips as level 0 spells
		if m.filterLevel == -1 || m.filterLevel == 0 {
			for _, cantripName := range sc.CantripsKnown {
				display = append(display, models.KnownSpell{
					Name:     cantripName,
					Level:    0,
					Prepared: true, // Cantrips are always prepared
					Ritual:   false,
				})
			}
		}
	}

	for _, spell := range sc.KnownSpells {
		// Apply filter
		if m.filterLevel != -1 && spell.Level != m.filterLevel {
			continue
		}

		// In spell list mode:
		// - For preparing classes (Wizard, Cleric): only show prepared spells
		// - For non-preparing classes (Warlock, Bard, Sorcerer): show all known spells
		// In preparation mode, show all spells for both
		if m.mode == ModeSpellList {
			if !sc.PreparesSpells {
				// Non-preparing classes: all known spells are available
				display = append(display, spell)
			} else if spell.Prepared || spell.AlwaysPrepared {
				// Preparing classes: show prepared spells and always-prepared spells
				display = append(display, spell)
			} else if spell.Ritual && sc.RitualCasterUnprepared {
				// Wizards can cast ritual spells from their spellbook without preparing
				display = append(display, spell)
			}
		} else {
			display = append(display, spell)
		}
	}

	// Sort by level, then by name
	sort.Slice(display, func(i, j int) bool {
		if display[i].Level != display[j].Level {
			return display[i].Level < display[j].Level
		}
		return display[i].Name < display[j].Name
	})

	return display
}

func (m *SpellbookModel) updateSelectedSpellData() {
	displaySpells := m.getDisplaySpells()

	if m.spellCursor >= len(displaySpells) || m.spellDatabase == nil {
		m.selectedSpellData = nil
		return
	}

	spell := displaySpells[m.spellCursor]

	// Find full spell data
	for i := range m.spellDatabase.Spells {
		if m.spellDatabase.Spells[i].Name == spell.Name {
			m.selectedSpellData = &m.spellDatabase.Spells[i]
			return
		}
	}

	m.selectedSpellData = nil
}

func (m *SpellbookModel) saveCharacter() tea.Cmd {
	return func() tea.Msg {
		if err := m.storage.AutoSave(m.character); err != nil {
			return nil // Silently fail for auto-save
		}
		return nil
	}
}

// detailsPanelWidth returns the width of the details panel based on current layout mode.
func (m *SpellbookModel) detailsPanelWidth() int {
	const compactBreakpoint = 90

	availableWidth := m.width
	if m.rollHistoryVisible {
		availableWidth -= m.rollHistoryWidth
	}

	if availableWidth < compactBreakpoint {
		// Compact mode: details gets 60% of available width
		listWidth := availableWidth * 40 / 100
		if listWidth < 25 {
			listWidth = 25
		}
		return availableWidth - listWidth
	}
	// Standard mode: details gets 1/3 of available width
	return availableWidth / 3
}

func (m *SpellbookModel) wordWrap(text string, width int) []string {
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
func (m *SpellbookModel) calculateUpcastEffect(slotLevel int) string {
	if m.selectedSpellData == nil || m.castingSpell == nil {
		return ""
	}

	spell := m.selectedSpellData
	levelsAbove := slotLevel - m.castingSpell.Level

	upcastLower := strings.ToLower(spell.Upcast)

	// Handle dart/projectile increases (like Magic Missile)
	if strings.Contains(upcastLower, "dart") {
		// Magic Missile: 3 darts base, +1 per level
		baseDarts := 3
		totalDarts := baseDarts + levelsAbove
		if spell.Damage != "" {
			return fmt.Sprintf(" (%d darts, %s each)", totalDarts, spell.Damage)
		}
		return fmt.Sprintf(" (%d darts)", totalDarts)
	}

	// Handle target increases (like Scorching Ray)
	if strings.Contains(upcastLower, "ray") || strings.Contains(upcastLower, "beam") {
		// Extract number from upcast (e.g., "+1 ray" -> 1)
		var bonusPerLevel int
		fmt.Sscanf(spell.Upcast, "+%d", &bonusPerLevel)
		if bonusPerLevel == 0 {
			bonusPerLevel = 1
		}

		// Scorching Ray: 3 rays base, +1 per level
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
		// Look for dice notation (e.g., "+1d6", "1d8")
		// Must contain 'd' followed by a digit, and have a digit before 'd'
		if strings.Contains(part, "d") && len(part) >= 3 {
			// Check if this is actually dice notation (not "2nd", "3rd", etc.)
			// Valid dice notation has: [number]d[number]
			cleaned := strings.TrimPrefix(strings.TrimPrefix(part, "+"), " ")
			dIndex := strings.Index(cleaned, "d")
			if dIndex > 0 && dIndex < len(cleaned)-1 {
				// Check if character before 'd' is a digit
				beforeD := cleaned[dIndex-1]
				// Check if character after 'd' is a digit
				afterD := cleaned[dIndex+1]
				if beforeD >= '0' && beforeD <= '9' && afterD >= '0' && afterD <= '9' {
					diceStr = cleaned
					break
				}
			}
		}
	}

	if diceStr != "" {
		// Parse dice notation
		parts := strings.Split(diceStr, "d")
		if len(parts) == 2 {
			var dicePerLevel int
			fmt.Sscanf(parts[0], "%d", &dicePerLevel)
			if dicePerLevel == 0 {
				dicePerLevel = 1
			}

			diceType := strings.TrimRight(parts[1], ".,;:")

			// Calculate total damage
			if spell.Damage != "" {
				// Parse base damage (e.g., "3d6" -> 3 dice)
				baseParts := strings.Split(spell.Damage, "d")
				if len(baseParts) == 2 {
					var baseDice int
					fmt.Sscanf(baseParts[0], "%d", &baseDice)
					totalDice := baseDice + (dicePerLevel * levelsAbove)
					return fmt.Sprintf(" (%dd%s damage)", totalDice, diceType)
				}
				// If we can't parse base damage, show it separately with bonus
				if levelsAbove > 0 {
					bonusDice := dicePerLevel * levelsAbove
					return fmt.Sprintf(" (%s + %dd%s damage)", spell.Damage, bonusDice, diceType)
				}
				// At base level, just show base damage
				return fmt.Sprintf(" (%s damage)", spell.Damage)
			}
			// No base damage, just show bonus (only if upcasting)
			if levelsAbove > 0 {
				bonusDice := dicePerLevel * levelsAbove
				return fmt.Sprintf(" (+%dd%s damage)", bonusDice, diceType)
			}
		}
	}

	// Handle healing increases
	if strings.Contains(upcastLower, "heal") {
		if spell.Damage != "" {
			// Damage field is used for healing in healing spells
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
// Returns slots at the spell's level and higher that have remaining uses.
func (m *SpellbookModel) getAvailableCastLevels(spellLevel int) []int {
	if m.character == nil || m.character.Spellcasting == nil {
		return []int{}
	}

	sc := m.character.Spellcasting
	var available []int

	// Check pact magic first
	if sc.PactMagic != nil && sc.PactMagic.Total > 0 && sc.PactMagic.Remaining > 0 {
		if sc.PactMagic.SlotLevel >= spellLevel {
			// Add pact magic as a "pseudo-level" (we'll handle it specially)
			// For display, we'll show it as "Pact (Level X)"
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

// castSpellAtLevel casts the spell using the specified slot level.
// Returns the updated model and an optional tea.Cmd for dice rolls.
func (m *SpellbookModel) castSpellAtLevel(spell *models.KnownSpell, slotLevel int) (*SpellbookModel, tea.Cmd) {
	if m.character == nil || m.character.Spellcasting == nil {
		return m, nil
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
			m.mode = ModeSpellList
			m.castingSpell = nil
			m.availableCastLevels = nil
			return m, m.spellRollCmd(spell)
		}
	}

	// Try regular spell slot
	if sc.SpellSlots.UseSlot(slotLevel) {
		upcastMsg := ""
		if slotLevel > spell.Level {
			upcastMsg = fmt.Sprintf(" (upcast to level %d)", slotLevel)
		}
		m.statusMessage = fmt.Sprintf("Cast %s%s using level %d slot", spell.Name, upcastMsg, slotLevel)
		m.mode = ModeSpellList
		m.castingSpell = nil
		m.availableCastLevels = nil
		return m, m.spellRollCmd(spell)
	}

	m.statusMessage = fmt.Sprintf("Failed to use spell slot")
	return m, nil
}

// spellRollCmd returns a tea.Cmd for dice rolls triggered by casting a damage spell.
// It looks up full spell data from the spell database to get damage info.
func (m *SpellbookModel) spellRollCmd(spell *models.KnownSpell) tea.Cmd {
	// Look up full spell data for damage info
	var fullSpell *data.SpellData
	if m.spellDatabase != nil {
		for i := range m.spellDatabase.Spells {
			if m.spellDatabase.Spells[i].Name == spell.Name {
				fullSpell = &m.spellDatabase.Spells[i]
				break
			}
		}
	}
	if fullSpell == nil {
		return nil
	}
	return components.BuildSpellRollCmd(
		fullSpell.Name, fullSpell.Damage, string(fullSpell.DamageType), fullSpell.SavingThrow,
		m.character.GetSpellAttackBonus(), m.getSpellSaveDC(),
	)
}

// handleCastLevelInput handles keyboard input in cast level selection mode.
func (m *SpellbookModel) handleCastLevelInput(msg tea.KeyPressMsg) (*SpellbookModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeSpellList
		m.castingSpell = nil
		m.availableCastLevels = nil
		m.statusMessage = "Casting cancelled"
		return m, nil

	case "enter":
		if m.castingSpell != nil && len(m.availableCastLevels) > 0 && m.castLevelCursor < len(m.availableCastLevels) {
			selectedLevel := m.availableCastLevels[m.castLevelCursor]
			result, rollCmd := m.castSpellAtLevel(m.castingSpell, selectedLevel)
			saveCmd := m.saveCharacter()
			if rollCmd != nil {
				return result, tea.Batch(saveCmd, rollCmd)
			}
			return result, saveCmd
		}
		return m, nil

	case "up", "k":
		if m.castLevelCursor > 0 {
			m.castLevelCursor--
		}
		return m, nil

	case "down", "j":
		if m.castLevelCursor < len(m.availableCastLevels)-1 {
			m.castLevelCursor++
		}
		return m, nil
	}

	return m, nil
}

// renderCastLevelOverlay renders the spell slot level selection modal.
func (m *SpellbookModel) renderCastLevelOverlay() string {
	if m.castingSpell == nil || m.character == nil || m.character.Spellcasting == nil {
		return ""
	}

	sc := m.character.Spellcasting
	var lines []string

	lines = append(lines, lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Cast %s", m.castingSpell.Name)))
	lines = append(lines, "")
	lines = append(lines, "Select spell slot level:")
	lines = append(lines, "")

	for i, level := range m.availableCastLevels {
		cursor := "  "
		if i == m.castLevelCursor {
			cursor = "> "
		}

		var line string
		if level == 0 {
			// Ritual casting option
			line = fmt.Sprintf("%sCast as Ritual (no slot, +10 min)", cursor)
		} else {
			// Check if this is pact magic
			isPactMagic := sc.PactMagic != nil && sc.PactMagic.SlotLevel == level

			if isPactMagic {
				remaining := sc.PactMagic.Remaining
				total := sc.PactMagic.Total
				upcastInfo := m.calculateUpcastEffect(level)
				line = fmt.Sprintf("%sPact Magic - Level %d%s [%d/%d remaining]", cursor, level, upcastInfo, remaining, total)
			} else {
				slot := sc.SpellSlots.GetSlot(level)
				if slot != nil {
					upcastInfo := m.calculateUpcastEffect(level)
					line = fmt.Sprintf("%sLevel %d Slot%s [%d/%d remaining]", cursor, level, upcastInfo, slot.Remaining, slot.Total)
				}
			}
		}

		if i == m.castLevelCursor {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(line)
		}

		lines = append(lines, line)
	}

	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("↑↓: select | Enter: confirm | Esc: cancel"))

	content := strings.Join(lines, "\n")

	width := 60
	height := len(lines) + 2

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Render(content)
}

// renderCastConfirmationModal renders the spell casting confirmation modal.
func (m *SpellbookModel) renderCastConfirmationModal() string {
	if m.castingSpell == nil || m.selectedSpellData == nil {
		return ""
	}

	spell := m.selectedSpellData
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
	// If casting as ritual (cursor on ritual option), show ritual casting time
	castingTime := spell.CastingTime
	isRitualSelected := m.castingSpell.Ritual && len(m.availableCastLevels) > 0 && m.availableCastLevels[m.castLevelCursor] == 0
	if isRitualSelected {
		castingTime = "10 minutes (ritual)"
	}
	lines = append(lines, fmt.Sprintf("Casting Time: %s", castingTime))
	lines = append(lines, fmt.Sprintf("Range: %s", spell.Range))
	lines = append(lines, fmt.Sprintf("Components: %s", strings.Join(data.ComponentsToStrings(spell.Components), ", ")))
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
	if m.castingSpell.Level > 0 && len(m.availableCastLevels) > 0 {
		lines = append(lines, "")
		lines = append(lines, strings.Repeat("─", 60))
		lines = append(lines, "")
	}

	// Slot selection (if not a cantrip and not being cast as ritual)
	if m.castingSpell.Level > 0 && len(m.availableCastLevels) > 0 {
		if len(m.availableCastLevels) == 1 {
			// Single option
			level := m.availableCastLevels[0]
			if level == 0 {
				// Only ritual option available
				lines = append(lines, "Using: Cast as Ritual (no slot, +10 min)")
			} else {
				upcastInfo := m.calculateUpcastEffect(level)
				sc := m.character.Spellcasting

				isPactMagic := sc.PactMagic != nil && sc.PactMagic.SlotLevel == level
				if isPactMagic {
					lines = append(lines, fmt.Sprintf("Using: Pact Magic - Level %d%s", level, upcastInfo))
				} else {
					lines = append(lines, fmt.Sprintf("Using: Level %d Slot%s", level, upcastInfo))
				}
			}
		} else {
			// Multiple options (may include ritual + slot levels)
			lines = append(lines, "Select Casting Method:")
			lines = append(lines, "")

			sc := m.character.Spellcasting
			for i, level := range m.availableCastLevels {
				cursor := "  "
				if i == m.castLevelCursor {
					cursor = "> "
				}

				var line string
				if level == 0 {
					// Ritual casting option
					line = fmt.Sprintf("%sCast as Ritual (no slot, +10 min)", cursor)
				} else {
					isPactMagic := sc.PactMagic != nil && sc.PactMagic.SlotLevel == level
					upcastInfo := m.calculateUpcastEffect(level)

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
	if m.castingSpell.Level == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Enter: cast cantrip | Esc: cancel"))
	} else if len(m.availableCastLevels) <= 1 {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Enter: cast | Esc: cancel"))
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("↑↓: select method | Enter: cast | Esc: cancel"))
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
