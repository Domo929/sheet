package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SpellbookFocus represents which panel is focused in the spellbook view.
type SpellbookFocus int

const (
	FocusSpellList SpellbookFocus = iota
	FocusSpellSlots
	numSpellbookFocusAreas
)

// SpellbookModel is the model for the spellbook view.
type SpellbookModel struct {
	character     *models.Character
	storage       *storage.CharacterStorage
	loader        *data.Loader
	width         int
	height        int
	focus         SpellbookFocus
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
	addingSpell    bool
	spellSearchTerm string
	searchResults  []data.SpellData
	searchCursor   int

	// Delete confirmation
	confirmingRemove bool
	removeSpellName  string

	// Spell data cache
	spellDatabase *data.SpellDatabase
}

type spellbookKeyMap struct {
	Quit      key.Binding
	ForceQuit key.Binding
	Tab       key.Binding
	ShiftTab  key.Binding
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	Back      key.Binding
	Prepare   key.Binding
	Cast      key.Binding
	Add       key.Binding
	Remove    key.Binding
	Filter    key.Binding
}

func defaultSpellbookKeyMap() spellbookKeyMap {
	return spellbookKeyMap{
		Quit:      key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "force quit")),
		Tab:       key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next panel")),
		ShiftTab:  key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev panel")),
		Up:        key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:      key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Enter:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "cast spell")),
		Back:      key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Prepare:   key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "prepare/unprepare")),
		Cast:      key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "cast spell")),
		Add:       key.NewBinding(key.WithKeys("a", "+"), key.WithHelp("a/+", "add spell")),
		Remove:    key.NewBinding(key.WithKeys("x", "delete"), key.WithHelp("x", "remove spell")),
		Filter:    key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "filter level")),
	}
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
		if m.addingSpell {
			return m.handleAddSpellInput(msg)
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
			return m, func() tea.Msg { return BackToSheetMsg{} }
		case key.Matches(msg, m.keys.Tab):
			m.focus = (m.focus + 1) % numSpellbookFocusAreas
			return m, nil
		case key.Matches(msg, m.keys.ShiftTab):
			if m.focus == 0 {
				m.focus = numSpellbookFocusAreas - 1
			} else {
				m.focus--
			}
			return m, nil
		case key.Matches(msg, m.keys.Up):
			return m.handleUp(), nil
		case key.Matches(msg, m.keys.Down):
			return m.handleDown(), nil
		case key.Matches(msg, m.keys.Prepare):
			return m.handlePrepareToggle(), m.saveCharacter()
		case key.Matches(msg, m.keys.Cast), key.Matches(msg, m.keys.Enter):
			return m.handleCastSpell(), m.saveCharacter()
		case key.Matches(msg, m.keys.Add):
			m.addingSpell = true
			m.spellSearchTerm = ""
			m.searchResults = []data.SpellData{}
			m.searchCursor = 0
			return m, nil
		case key.Matches(msg, m.keys.Remove):
			return m.handleRemoveSpell(), nil
		case key.Matches(msg, m.keys.Filter):
			return m.handleFilterToggle(), nil
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

	// Layout: Header at top, then three panels (list | details | slots)
	listWidth := m.width / 3
	detailsWidth := m.width / 3
	slotsWidth := m.width - listWidth - detailsWidth

	spellListStyled := lipgloss.NewStyle().
		Width(listWidth).
		Height(m.height - lipgloss.Height(header) - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.borderColor(FocusSpellList)).
		Render(spellList)

	spellDetailsStyled := lipgloss.NewStyle().
		Width(detailsWidth).
		Height(m.height - lipgloss.Height(header) - 2).
		Border(lipgloss.RoundedBorder()).
		Render(spellDetails)

	spellSlotsStyled := lipgloss.NewStyle().
		Width(slotsWidth).
		Height(m.height - lipgloss.Height(header) - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.borderColor(FocusSpellSlots)).
		Render(spellSlots)

	panels := lipgloss.JoinHorizontal(lipgloss.Top, spellListStyled, spellDetailsStyled, spellSlotsStyled)

	// Build footer
	footer := m.renderFooter()

	// Handle add spell overlay
	if m.addingSpell {
		overlay := m.renderAddSpellOverlay()
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

	title := lipgloss.NewStyle().Bold(true).Render("Spellbook")

	stats := fmt.Sprintf("Ability: %s | Save DC: %d | Attack: +%d",
		sc.Ability, saveDC, attackBonus)

	// Show prepared spell count if applicable
	prepInfo := ""
	if sc.PreparesSpells && sc.MaxPrepared > 0 {
		prepCount := sc.CountPreparedSpells()
		prepInfo = fmt.Sprintf(" | Prepared: %d/%d", prepCount, sc.MaxPrepared)
	}

	statsStyled := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(stats + prepInfo)

	return lipgloss.JoinVertical(lipgloss.Left, title, statsStyled) + "\n"
}

// renderSpellList renders the spell list panel.
func (m *SpellbookModel) renderSpellList() string {
	if m.character == nil || m.character.Spellcasting == nil {
		return "No spellcasting ability"
	}

	sc := m.character.Spellcasting
	var lines []string

	// Group spells by level
	spellsByLevel := make(map[int][]models.KnownSpell)
	for _, spell := range sc.KnownSpells {
		if m.filterLevel == -1 || spell.Level == m.filterLevel {
			spellsByLevel[spell.Level] = append(spellsByLevel[spell.Level], spell)
		}
	}

	// Add cantrips
	if m.filterLevel == -1 || m.filterLevel == 0 {
		if len(sc.CantripsKnown) > 0 {
			lines = append(lines, lipgloss.NewStyle().Bold(true).Render("Cantrips"))
			for _, cantrip := range sc.CantripsKnown {
				lines = append(lines, fmt.Sprintf("  %s", cantrip))
			}
			lines = append(lines, "")
		}
	}

	// Add leveled spells
	levels := make([]int, 0, len(spellsByLevel))
	for level := range spellsByLevel {
		levels = append(levels, level)
	}
	sort.Ints(levels)

	flatSpells := m.getFlatSpellList()
	currentIndex := 0

	for _, level := range levels {
		spells := spellsByLevel[level]
		lines = append(lines, lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Level %d", level)))

		for _, spell := range spells {
			cursor := "  "
			if currentIndex == m.spellCursor {
				cursor = "> "
			}

			prepMarker := " "
			if spell.Prepared {
				prepMarker = "✓"
			} else if sc.PreparesSpells {
				prepMarker = " "
			}

			ritualMarker := ""
			if spell.Ritual {
				ritualMarker = " (R)"
			}

			line := fmt.Sprintf("%s[%s] %s%s", cursor, prepMarker, spell.Name, ritualMarker)

			if currentIndex == m.spellCursor {
				line = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(line)
			}

			lines = append(lines, line)
			currentIndex++
		}
		lines = append(lines, "")
	}

	if len(flatSpells) == 0 && len(sc.CantripsKnown) == 0 {
		lines = append(lines, "No spells known")
		lines = append(lines, "Press 'a' to add a spell")
	}

	// Apply scrolling
	maxHeight := m.height - 10
	if len(lines) > maxHeight {
		start := m.spellScroll
		end := start + maxHeight
		if end > len(lines) {
			end = len(lines)
		}
		lines = lines[start:end]
	}

	return strings.Join(lines, "\n")
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
	lines = append(lines, fmt.Sprintf("Components: %s", strings.Join(spell.Components, ", ")))
	lines = append(lines, fmt.Sprintf("Duration: %s", spell.Duration))
	lines = append(lines, "")

	// Word wrap description
	descLines := m.wordWrap(spell.Description, m.width/3-4)
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

	if m.character.Spellcasting != nil && m.character.Spellcasting.PreparesSpells {
		helps = append(helps, "p: prepare")
	}

	helps = append(helps, "c: cast")
	helps = append(helps, "a: add spell")
	helps = append(helps, "x: remove")
	helps = append(helps, "f: filter")
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

func (m *SpellbookModel) borderColor(focus SpellbookFocus) lipgloss.Color {
	if m.focus == focus {
		return lipgloss.Color("12")
	}
	return lipgloss.Color("240")
}

func (m *SpellbookModel) handleUp() *SpellbookModel {
	if m.focus == FocusSpellList {
		if m.spellCursor > 0 {
			m.spellCursor--
			m.updateSelectedSpellData()

			// Adjust scroll if needed
			if m.spellCursor < m.spellScroll {
				m.spellScroll = m.spellCursor
			}
		}
	}
	return m
}

func (m *SpellbookModel) handleDown() *SpellbookModel {
	if m.focus == FocusSpellList {
		flatSpells := m.getFlatSpellList()
		if m.spellCursor < len(flatSpells)-1 {
			m.spellCursor++
			m.updateSelectedSpellData()

			// Adjust scroll if needed
			maxHeight := m.height - 10
			if m.spellCursor >= m.spellScroll+maxHeight {
				m.spellScroll = m.spellCursor - maxHeight + 1
			}
		}
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

	flatSpells := m.getFlatSpellList()
	if m.spellCursor >= len(flatSpells) {
		return m
	}

	spell := flatSpells[m.spellCursor]

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
	flatSpells := m.getFlatSpellList()
	if m.spellCursor >= len(flatSpells) {
		// Check if a cantrip is selected
		if m.spellCursor-len(flatSpells) < len(sc.CantripsKnown) {
			cantripIdx := m.spellCursor - len(flatSpells)
			m.statusMessage = fmt.Sprintf("Cast %s (no slot required)", sc.CantripsKnown[cantripIdx])
			return m
		}
		return m
	}

	spell := flatSpells[m.spellCursor]

	// Check if spell is prepared (if applicable)
	if sc.PreparesSpells && !spell.Prepared && !spell.Ritual {
		m.statusMessage = fmt.Sprintf("%s is not prepared", spell.Name)
		return m
	}

	// Cantrips don't use slots
	if spell.Level == 0 {
		m.statusMessage = fmt.Sprintf("Cast %s (no slot required)", spell.Name)
		return m
	}

	// Try to use a spell slot
	if sc.SpellSlots.UseSlot(spell.Level) {
		m.statusMessage = fmt.Sprintf("Cast %s (used level %d slot)", spell.Name, spell.Level)
		return m
	}

	// Try pact magic if applicable
	if sc.PactMagic != nil && sc.PactMagic.SlotLevel >= spell.Level {
		if sc.PactMagic.Use() {
			m.statusMessage = fmt.Sprintf("Cast %s (used pact magic slot)", spell.Name)
			return m
		}
	}

	m.statusMessage = fmt.Sprintf("No spell slots available for %s", spell.Name)
	return m
}

func (m *SpellbookModel) handleRemoveSpell() *SpellbookModel {
	flatSpells := m.getFlatSpellList()
	if m.spellCursor >= len(flatSpells) {
		return m
	}

	spell := flatSpells[m.spellCursor]
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
			flatSpells := m.getFlatSpellList()
			if m.spellCursor >= len(flatSpells) && m.spellCursor > 0 {
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

func (m *SpellbookModel) handleAddSpellInput(msg tea.KeyMsg) (*SpellbookModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.addingSpell = false
		m.spellSearchTerm = ""
		m.searchResults = []data.SpellData{}
		m.statusMessage = "Add spell cancelled"
		return m, nil

	case "enter":
		if len(m.searchResults) > 0 && m.searchCursor < len(m.searchResults) {
			spell := m.searchResults[m.searchCursor]
			m.addSpellToCharacter(spell)
			m.addingSpell = false
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

func (m *SpellbookModel) getFlatSpellList() []models.KnownSpell {
	if m.character == nil || m.character.Spellcasting == nil {
		return []models.KnownSpell{}
	}

	sc := m.character.Spellcasting
	var flat []models.KnownSpell

	// Apply filter
	for _, spell := range sc.KnownSpells {
		if m.filterLevel == -1 || spell.Level == m.filterLevel {
			flat = append(flat, spell)
		}
	}

	// Sort by level, then by name
	sort.Slice(flat, func(i, j int) bool {
		if flat[i].Level != flat[j].Level {
			return flat[i].Level < flat[j].Level
		}
		return flat[i].Name < flat[j].Name
	})

	return flat
}

func (m *SpellbookModel) updateSelectedSpellData() {
	flatSpells := m.getFlatSpellList()

	if m.spellCursor >= len(flatSpells) || m.spellDatabase == nil {
		m.selectedSpellData = nil
		return
	}

	spell := flatSpells[m.spellCursor]

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
