package views

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Domo929/sheet/internal/data"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
)

// OpenMulticlassMsg signals to open the multiclass management view.
type OpenMulticlassMsg struct{}

// multiclassMode represents the active input mode in the multiclass view.
type multiclassMode int

const (
	multiclassModeList multiclassMode = iota
	multiclassModeAdd                 // adding a new class (multi-step form)
)

// Add-class form steps.
const (
	mcAddStepClass = iota
	mcAddStepLevel
	mcAddStepSubclass
	mcAddStepCount
)

// MulticlassModel is the model for the multiclass management view.
type MulticlassModel struct {
	character *models.Character
	storage   *storage.CharacterStorage
	loader    *data.Loader
	width     int
	height    int
	keys      multiclassKeyMap

	classNames []string       // all available class names (for the add picker)
	hitDice    map[string]int // class name -> hit die size

	cursor int // selected class entry index
	mode   multiclassMode

	// Add-class form state.
	addStep     int
	addAvail    []string // class names not yet taken (snapshot at form start)
	addClassIdx int
	addLevelBuf string
	addSubclass string

	confirmingDelete bool
	statusMessage    string
}

type multiclassKeyMap struct {
	Quit      key.Binding
	Back      key.Binding
	Up        key.Binding
	Down      key.Binding
	Add       key.Binding
	LevelUp   key.Binding
	LevelDown key.Binding
	Delete    key.Binding
}

func defaultMulticlassKeyMap() multiclassKeyMap {
	return multiclassKeyMap{
		Quit:      key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
		Back:      key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Up:        key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:      key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Add:       key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add class")),
		LevelUp:   key.NewBinding(key.WithKeys("+", "="), key.WithHelp("+", "level up")),
		LevelDown: key.NewBinding(key.WithKeys("-", "_"), key.WithHelp("-", "level down")),
		Delete:    key.NewBinding(key.WithKeys("x", "delete"), key.WithHelp("x", "remove")),
	}
}

// NewMulticlassModel creates a new multiclass model. It seeds the class list
// from the character's single-class Info the first time it is opened, so an
// existing character's primary class always appears as the first entry.
func NewMulticlassModel(char *models.Character, store *storage.CharacterStorage, loader *data.Loader) *MulticlassModel {
	m := &MulticlassModel{
		character:  char,
		storage:    store,
		loader:     loader,
		keys:       defaultMulticlassKeyMap(),
		hitDice:    map[string]int{},
		classNames: []string{},
	}
	m.loadClasses()
	m.seedFromInfo()
	return m
}

// loadClasses populates the class-name picker list and hit-die lookup.
func (m *MulticlassModel) loadClasses() {
	if m.loader == nil {
		return
	}
	classes, err := m.loader.GetClasses()
	if err != nil || classes == nil {
		return
	}
	for _, c := range classes.Classes {
		m.classNames = append(m.classNames, c.Name)
		m.hitDice[c.Name] = parseHitDie(c.HitDice)
	}
	sort.Strings(m.classNames)
}

// seedFromInfo populates Classes with the primary class if it is empty. This is
// held in memory and only persisted once the user makes an actual change.
func (m *MulticlassModel) seedFromInfo() {
	if m.character == nil || len(m.character.Classes) > 0 {
		return
	}
	if m.character.Info.Class == "" {
		return
	}
	m.character.Classes = []models.ClassLevel{{
		Class:    m.character.Info.Class,
		Subclass: m.character.Info.Subclass,
		Level:    m.character.Info.Level,
		HitDie:   m.dieFor(m.character.Info.Class),
	}}
}

// parseHitDie parses a hit-dice string like "d10" or "1d10" into its die size.
func parseHitDie(s string) int {
	if i := strings.Index(s, "d"); i >= 0 && i+1 < len(s) {
		if v, err := strconv.Atoi(strings.TrimSpace(s[i+1:])); err == nil && v > 0 {
			return v
		}
	}
	return 8
}

func (m *MulticlassModel) dieFor(class string) int {
	if d, ok := m.hitDice[class]; ok && d > 0 {
		return d
	}
	return 8
}

// Init initializes the model.
func (m *MulticlassModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m *MulticlassModel) Update(msg tea.Msg) (*MulticlassModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}

		if m.mode == multiclassModeAdd {
			return m.handleAddInput(msg)
		}

		if m.confirmingDelete {
			switch msg.String() {
			case "y", "Y":
				m.performDelete()
			default:
				m.confirmingDelete = false
				m.statusMessage = "Removal cancelled"
			}
			return m, nil
		}

		return m.handleListInput(msg)
	}
	return m, nil
}

func (m *MulticlassModel) handleListInput(msg tea.KeyPressMsg) (*MulticlassModel, tea.Cmd) {
	m.statusMessage = ""
	switch {
	case key.Matches(msg, m.keys.Back):
		return m, func() tea.Msg { return BackToSheetMsg{} }
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(m.character.Classes)-1 {
			m.cursor++
		}
	case key.Matches(msg, m.keys.Add):
		m.startAdd()
	case key.Matches(msg, m.keys.LevelUp):
		m.adjustLevel(1)
	case key.Matches(msg, m.keys.LevelDown):
		m.adjustLevel(-1)
	case key.Matches(msg, m.keys.Delete):
		if len(m.character.Classes) <= 1 {
			m.statusMessage = "A character must have at least one class"
		} else if m.hasSelection() {
			m.confirmingDelete = true
			m.statusMessage = fmt.Sprintf("Remove %s? (y/n)", m.selected().Class)
		}
	}
	return m, nil
}

func (m *MulticlassModel) hasSelection() bool {
	return m.cursor >= 0 && m.cursor < len(m.character.Classes)
}

func (m *MulticlassModel) selected() *models.ClassLevel {
	if !m.hasSelection() {
		return nil
	}
	return &m.character.Classes[m.cursor]
}

// availableClasses returns class names not already taken by the character.
func (m *MulticlassModel) availableClasses() []string {
	taken := map[string]bool{}
	for _, cl := range m.character.Classes {
		taken[strings.ToLower(cl.Class)] = true
	}
	var out []string
	for _, name := range m.classNames {
		if !taken[strings.ToLower(name)] {
			out = append(out, name)
		}
	}
	return out
}

func (m *MulticlassModel) startAdd() {
	if m.character.TotalLevel() >= 20 {
		m.statusMessage = "Total character level is already 20"
		return
	}
	avail := m.availableClasses()
	if len(avail) == 0 {
		m.statusMessage = "No more classes available"
		return
	}
	m.mode = multiclassModeAdd
	m.addStep = mcAddStepClass
	m.addAvail = avail
	m.addClassIdx = 0
	m.addLevelBuf = "1"
	m.addSubclass = ""
	m.statusMessage = "Add class: ↑/↓ to pick, Enter to continue, Esc to cancel"
}

func (m *MulticlassModel) handleAddInput(msg tea.KeyPressMsg) (*MulticlassModel, tea.Cmd) {
	if msg.Code == tea.KeyEscape {
		m.mode = multiclassModeList
		m.statusMessage = "Add cancelled"
		return m, nil
	}

	switch m.addStep {
	case mcAddStepClass:
		switch msg.Code {
		case tea.KeyUp:
			if m.addClassIdx > 0 {
				m.addClassIdx--
			}
		case tea.KeyDown:
			if m.addClassIdx < len(m.addAvail)-1 {
				m.addClassIdx++
			}
		case tea.KeyEnter:
			m.addStep++
		}
	case mcAddStepLevel:
		switch msg.Code {
		case tea.KeyEnter:
			m.addStep++
		case tea.KeyBackspace:
			if len(m.addLevelBuf) > 0 {
				m.addLevelBuf = m.addLevelBuf[:len(m.addLevelBuf)-1]
			}
		default:
			if isDigitStr(msg.Text) {
				m.addLevelBuf += msg.Text
			}
		}
	case mcAddStepSubclass:
		switch msg.Code {
		case tea.KeyEnter:
			m.finalizeAdd()
		case tea.KeyBackspace:
			if len(m.addSubclass) > 0 {
				m.addSubclass = m.addSubclass[:len(m.addSubclass)-1]
			}
		default:
			if msg.Text != "" {
				m.addSubclass += msg.Text
			}
		}
	}
	return m, nil
}

func (m *MulticlassModel) finalizeAdd() {
	if len(m.addAvail) == 0 || m.addClassIdx >= len(m.addAvail) {
		m.mode = multiclassModeList
		return
	}
	class := m.addAvail[m.addClassIdx]

	level, _ := strconv.Atoi(m.addLevelBuf)
	if level < 1 {
		level = 1
	}
	// Cap total character level at 20.
	if room := 20 - m.character.TotalLevel(); level > room {
		level = room
	}
	if level < 1 {
		m.mode = multiclassModeList
		m.statusMessage = "Total character level is already 20"
		return
	}

	m.character.Classes = append(m.character.Classes, models.ClassLevel{
		Class:    class,
		Subclass: strings.TrimSpace(m.addSubclass),
		Level:    level,
		HitDie:   m.dieFor(class),
	})
	m.recomputeDerived()
	m.cursor = len(m.character.Classes) - 1
	m.mode = multiclassModeList
	m.statusMessage = fmt.Sprintf("Added %s %d", class, level)
}

// adjustLevel changes the selected class's level by delta, keeping each class in
// 1..20 and the total character level at or below 20.
func (m *MulticlassModel) adjustLevel(delta int) {
	entry := m.selected()
	if entry == nil {
		return
	}
	next := entry.Level + delta
	if next < 1 {
		m.statusMessage = "A class cannot go below level 1 (remove it with x)"
		return
	}
	if next > 20 {
		return
	}
	if delta > 0 && m.character.TotalLevel()+delta > 20 {
		m.statusMessage = "Total character level is already 20"
		return
	}
	entry.Level = next
	m.recomputeDerived()
	m.statusMessage = fmt.Sprintf("%s is now level %d", entry.Class, entry.Level)
}

func (m *MulticlassModel) performDelete() {
	if entry := m.selected(); entry != nil && len(m.character.Classes) > 1 {
		name := entry.Class
		m.character.Classes = append(m.character.Classes[:m.cursor], m.character.Classes[m.cursor+1:]...)
		if m.cursor >= len(m.character.Classes) && m.cursor > 0 {
			m.cursor--
		}
		m.recomputeDerived()
		m.statusMessage = fmt.Sprintf("Removed %s", name)
	}
	m.confirmingDelete = false
}

// recomputeDerived re-syncs the primary class, hit dice, and spell slots after
// any change to the class breakdown, then persists the character.
func (m *MulticlassModel) recomputeDerived() {
	m.character.SyncPrimaryClass()

	// Update total hit dice, preserving how many have already been spent.
	total := m.character.TotalLevel()
	hd := &m.character.CombatStats.HitDice
	used := hd.Total - hd.Remaining
	if used < 0 {
		used = 0
	}
	hd.Total = total
	hd.Remaining = total - used
	if hd.Remaining < 0 {
		hd.Remaining = 0
	}
	if hd.Remaining > total {
		hd.Remaining = total
	}
	if len(m.character.Classes) > 0 && m.character.Classes[0].HitDie > 0 {
		hd.DieType = m.character.Classes[0].HitDie
	}

	// Recompute spell slots.
	if m.character.IsMulticlass() {
		m.character.ApplyMulticlassSpellSlots()
	} else if len(m.character.Classes) == 1 {
		m.applySingleClassSlots(m.character.Classes[0])
	}

	m.saveCharacter()
}

// applySingleClassSlots restores regular spell slots for a single-class caster
// from that class's own progression table, preserving expended slots. It only
// acts when the class has a slot row for the level (i.e., full/half casters);
// subclass casters and non-casters are left untouched so nothing is wiped.
func (m *MulticlassModel) applySingleClassSlots(entry models.ClassLevel) {
	if m.loader == nil {
		return
	}
	class, err := m.loader.FindClassByName(entry.Class)
	if err != nil || class == nil {
		return
	}
	var row *data.SpellSlot
	for i := range class.SpellSlots {
		if class.SpellSlots[i].Level == entry.Level {
			row = &class.SpellSlots[i]
			break
		}
	}
	if row == nil {
		return
	}
	if m.character.Spellcasting == nil {
		sc := models.NewSpellcasting(models.Ability(strings.ToLower(class.SpellcastingAbility)))
		m.character.Spellcasting = &sc
	}
	for lvl := 1; lvl <= 9; lvl++ {
		count := getSpellSlotCount(*row, lvl)
		tracker := m.character.Spellcasting.SpellSlots.GetSlot(lvl)
		if tracker == nil {
			if count > 0 {
				m.character.Spellcasting.SpellSlots.SetSlots(lvl, count)
			}
			continue
		}
		used := tracker.Total - tracker.Remaining
		if used < 0 {
			used = 0
		}
		tracker.Total = count
		tracker.Remaining = count - used
		if tracker.Remaining < 0 {
			tracker.Remaining = 0
		}
		if tracker.Remaining > count {
			tracker.Remaining = count
		}
	}
}

func (m *MulticlassModel) saveCharacter() {
	if m.storage != nil {
		_ = m.storage.AutoSave(m.character)
	}
}

// View renders the multiclass view.
func (m *MulticlassModel) View() string {
	if m.character == nil {
		return "No character loaded"
	}

	width := m.width
	if width == 0 {
		width = 120
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	header := titleStyle.Render("⚔ Classes & Multiclassing") + "  " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(m.character.Info.Name)

	var body string
	if m.mode == multiclassModeAdd {
		body = m.renderAddForm(width)
	} else {
		body = m.renderList(width)
	}

	footer := m.renderFooter(width)
	return lipgloss.JoinVertical(lipgloss.Left, header, "", body, "", footer)
}

func (m *MulticlassModel) renderList(width int) string {
	if len(m.character.Classes) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(
			"No class data.\n\nPress 'a' to add a class.")
	}

	listWidth := width * 34 / 100
	if listWidth < 26 {
		listWidth = 26
	}
	if listWidth > 44 {
		listWidth = 44
	}

	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	var listLines []string
	for i := range m.character.Classes {
		cl := &m.character.Classes[i]
		prefix := "  "
		style := normalStyle
		if i == m.cursor {
			prefix = "▶ "
			style = selectedStyle
		}
		label := fmt.Sprintf("%s %d", cl.Class, cl.Level)
		if i == 0 {
			label += "  (primary)"
		}
		listLines = append(listLines, fmt.Sprintf("%s%s", prefix, style.Render(label)))
		sub := cl.Subclass
		if sub == "" {
			sub = "—"
		}
		listLines = append(listLines, dimStyle.Render(fmt.Sprintf("   d%d · %s", cl.HitDie, sub)))
	}

	listBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(0, 1).
		Width(listWidth).
		Render(strings.Join(listLines, "\n"))

	detail := m.renderSummary(width - listWidth - 6)

	return lipgloss.JoinHorizontal(lipgloss.Top, listBox, detail)
}

func (m *MulticlassModel) renderSummary(width int) string {
	if width < 28 {
		width = 28
	}
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("111"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	var lines []string
	lines = append(lines, titleStyle.Render(m.character.ClassSummary()))
	lines = append(lines, fmt.Sprintf("%s %s",
		labelStyle.Render("Total level"), valueStyle.Render(strconv.Itoa(m.character.TotalLevel()))))

	// Hit dice, grouped by die size (e.g., 5d10 + 3d6).
	lines = append(lines, fmt.Sprintf("%s %s",
		labelStyle.Render("Hit dice"), valueStyle.Render(m.hitDiceSummary())))
	lines = append(lines, "")

	// Spellcasting summary.
	casterLevel := models.MulticlassCasterLevel(m.character.Classes)
	if m.character.IsMulticlass() {
		lines = append(lines, labelStyle.Render("Multiclass spellcaster"))
		if casterLevel > 0 {
			lines = append(lines, dimStyle.Render(fmt.Sprintf("Combined caster level %d", casterLevel)))
			slots := models.MulticlassSpellSlots(casterLevel)
			lines = append(lines, valueStyle.Render(formatSlotRow(slots)))
		} else {
			lines = append(lines, dimStyle.Render("No combined spell slots"))
		}
		if m.hasWarlock() {
			lines = append(lines, dimStyle.Render("Warlock Pact Magic is tracked separately"))
		}
	} else {
		lines = append(lines, dimStyle.Render("Single class — press 'a' to multiclass"))
	}

	lines = append(lines, "")
	lines = append(lines, dimStyle.Render("Adjust levels here to keep multiclass slots in sync."))
	lines = append(lines, dimStyle.Render("Use Level Up (main sheet) for features & HP."))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(width).
		Render(strings.Join(lines, "\n"))
}

// hitDiceSummary groups class hit dice by die size, e.g. "5d10 + 3d6".
func (m *MulticlassModel) hitDiceSummary() string {
	counts := map[int]int{}
	var order []int
	for _, cl := range m.character.Classes {
		die := cl.HitDie
		if die == 0 {
			die = 8
		}
		if _, seen := counts[die]; !seen {
			order = append(order, die)
		}
		counts[die] += cl.Level
	}
	sort.Sort(sort.Reverse(sort.IntSlice(order)))
	var parts []string
	for _, die := range order {
		parts = append(parts, fmt.Sprintf("%dd%d", counts[die], die))
	}
	return strings.Join(parts, " + ")
}

func (m *MulticlassModel) hasWarlock() bool {
	for _, cl := range m.character.Classes {
		if strings.EqualFold(cl.Class, "Warlock") {
			return true
		}
	}
	return false
}

// formatSlotRow renders a slots-per-level array as "L1:4 L2:3 L3:3 ...".
func formatSlotRow(slots [9]int) string {
	var parts []string
	for i, n := range slots {
		if n > 0 {
			parts = append(parts, fmt.Sprintf("L%d:%d", i+1, n))
		}
	}
	if len(parts) == 0 {
		return "—"
	}
	return strings.Join(parts, "  ")
}

func (m *MulticlassModel) renderAddForm(width int) string {
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	doneStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))

	field := func(step int, label, value string) string {
		style := dimStyle
		cursor := ""
		switch {
		case step == m.addStep:
			style = activeStyle
			cursor = "_"
		case step < m.addStep:
			style = doneStyle
		}
		return style.Render(fmt.Sprintf("%-10s %s%s", label+":", value, cursor))
	}

	className := ""
	if m.addClassIdx < len(m.addAvail) {
		className = m.addAvail[m.addClassIdx]
	}
	if m.addStep == mcAddStepClass {
		className = "◀ " + className + " ▶"
	}

	var lines []string
	lines = append(lines, titleStyle.Render("Add Class"))
	lines = append(lines, "")
	lines = append(lines, field(mcAddStepClass, "Class", className))
	lines = append(lines, field(mcAddStepLevel, "Level", m.addLevelBuf))
	lines = append(lines, field(mcAddStepSubclass, "Subclass", m.addSubclass))
	lines = append(lines, "")
	lines = append(lines, dimStyle.Render("Subclass matters for casters (e.g. Eldritch Knight, Arcane Trickster)."))
	lines = append(lines, dimStyle.Render("↑/↓: pick class • Enter: next/create • Esc: cancel"))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("214")).
		Padding(0, 1).
		Width(width - 4).
		Render(strings.Join(lines, "\n"))
}

func (m *MulticlassModel) renderFooter(width int) string {
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Width(width)
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).Width(width)

	var help string
	switch {
	case m.mode == multiclassModeAdd:
		help = "↑/↓: pick class • type level • enter: next/create • esc: cancel"
	case m.confirmingDelete:
		help = "y: confirm remove • any other key: cancel"
	default:
		help = "↑/↓: select • a: add class • +/-: level • x: remove • esc: back"
	}

	var lines []string
	if m.statusMessage != "" {
		lines = append(lines, statusStyle.Render(m.statusMessage))
	}
	lines = append(lines, helpStyle.Render(help))
	return strings.Join(lines, "\n")
}

// CursorInfo returns cursor settings when a text-input step is active.
func (m *MulticlassModel) CursorInfo() *tea.Cursor {
	if m.mode == multiclassModeAdd && (m.addStep == mcAddStepLevel || m.addStep == mcAddStepSubclass) {
		return &tea.Cursor{Shape: tea.CursorBar, Blink: true}
	}
	return nil
}
