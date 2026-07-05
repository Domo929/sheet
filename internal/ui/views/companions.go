package views

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
)

// OpenCompanionsMsg signals to open the companions view.
type OpenCompanionsMsg struct{}

// companionMode represents the active input mode in the companions view.
type companionMode int

const (
	companionModeList   companionMode = iota
	companionModeAdd                  // adding a new companion (multi-step form)
	companionModeAttack               // adding an attack to the selected companion
	companionModeHP                   // entering a damage/heal amount
)

// Add-form steps.
const (
	addStepName = iota
	addStepKind
	addStepSizeType
	addStepAC
	addStepHP
	addStepSpeed
	addStepAbilities
	addStepCount
)

// Attack-form steps.
const (
	atkStepName = iota
	atkStepBonus
	atkStepDamage
	atkStepCount
)

// CompanionsModel is the model for the companions / summons view.
type CompanionsModel struct {
	character *models.Character
	storage   *storage.CharacterStorage
	width     int
	height    int
	keys      companionsKeyMap

	cursor int // selected companion index
	scroll int

	mode companionMode

	// Add-companion form state.
	addStep      int
	addName      string
	addKindIndex int
	addSizeType  string
	addACBuf     string
	addHPBuf     string
	addSpeed     string
	addAbilBuf   string

	// Add-attack form state.
	atkStep      int
	atkName      string
	atkBonusBuf  string
	atkDamageBuf string

	// HP adjustment state.
	hpBuf    string
	hpIsHeal bool

	confirmingDelete bool
	statusMessage    string
}

type companionsKeyMap struct {
	Quit   key.Binding
	Back   key.Binding
	Up     key.Binding
	Down   key.Binding
	Add    key.Binding
	Attack key.Binding
	Damage key.Binding
	Heal   key.Binding
	Delete key.Binding
}

func defaultCompanionsKeyMap() companionsKeyMap {
	return companionsKeyMap{
		Quit:   key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
		Back:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Add:    key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add companion")),
		Attack: key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "add attack")),
		Damage: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "damage")),
		Heal:   key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "heal")),
		Delete: key.NewBinding(key.WithKeys("x", "delete"), key.WithHelp("x", "delete")),
	}
}

// NewCompanionsModel creates a new companions model.
func NewCompanionsModel(char *models.Character, store *storage.CharacterStorage) *CompanionsModel {
	return &CompanionsModel{
		character: char,
		storage:   store,
		keys:      defaultCompanionsKeyMap(),
	}
}

// Init initializes the model.
func (m *CompanionsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m *CompanionsModel) Update(msg tea.Msg) (*CompanionsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}

		switch m.mode {
		case companionModeAdd:
			return m.handleAddInput(msg)
		case companionModeAttack:
			return m.handleAttackInput(msg)
		case companionModeHP:
			return m.handleHPInput(msg)
		}

		if m.confirmingDelete {
			switch msg.String() {
			case "y", "Y":
				m.performDelete()
			default:
				m.confirmingDelete = false
				m.statusMessage = "Delete cancelled"
			}
			return m, nil
		}

		return m.handleListInput(msg)
	}
	return m, nil
}

func (m *CompanionsModel) handleListInput(msg tea.KeyPressMsg) (*CompanionsModel, tea.Cmd) {
	m.statusMessage = ""
	switch {
	case key.Matches(msg, m.keys.Back):
		return m, func() tea.Msg { return BackToSheetMsg{} }
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(m.character.Companions)-1 {
			m.cursor++
		}
	case key.Matches(msg, m.keys.Add):
		m.startAdd()
	case key.Matches(msg, m.keys.Attack):
		if m.hasSelection() {
			m.mode = companionModeAttack
			m.atkStep = atkStepName
			m.atkName = ""
			m.atkBonusBuf = ""
			m.atkDamageBuf = ""
			m.statusMessage = "New attack: type a name, Enter to continue, Esc to cancel"
		}
	case key.Matches(msg, m.keys.Damage):
		if m.hasSelection() {
			m.mode = companionModeHP
			m.hpIsHeal = false
			m.hpBuf = ""
			m.statusMessage = "Damage: type amount, Enter to apply, Esc to cancel"
		}
	case key.Matches(msg, m.keys.Heal):
		if m.hasSelection() {
			m.mode = companionModeHP
			m.hpIsHeal = true
			m.hpBuf = ""
			m.statusMessage = "Heal: type amount, Enter to apply, Esc to cancel"
		}
	case key.Matches(msg, m.keys.Delete):
		if m.hasSelection() {
			m.confirmingDelete = true
			m.statusMessage = fmt.Sprintf("Delete %s? (y/n)", m.selected().Name)
		}
	}
	return m, nil
}

func (m *CompanionsModel) hasSelection() bool {
	return m.cursor >= 0 && m.cursor < len(m.character.Companions)
}

func (m *CompanionsModel) selected() *models.Companion {
	if !m.hasSelection() {
		return nil
	}
	return &m.character.Companions[m.cursor]
}

func (m *CompanionsModel) startAdd() {
	m.mode = companionModeAdd
	m.addStep = addStepName
	m.addName = ""
	m.addKindIndex = 0
	m.addSizeType = ""
	m.addACBuf = ""
	m.addHPBuf = ""
	m.addSpeed = ""
	m.addAbilBuf = ""
	m.statusMessage = "New companion: type a name, Enter to continue, Esc to cancel"
}

func (m *CompanionsModel) handleAddInput(msg tea.KeyPressMsg) (*CompanionsModel, tea.Cmd) {
	if msg.Code == tea.KeyEscape {
		m.mode = companionModeList
		m.statusMessage = "Add cancelled"
		return m, nil
	}

	kinds := models.AllCompanionKinds()

	switch m.addStep {
	case addStepName:
		m.editText(&m.addName, msg, func() bool {
			if strings.TrimSpace(m.addName) == "" {
				m.statusMessage = "Name cannot be empty"
				return false
			}
			return true
		})
	case addStepKind:
		switch msg.Code {
		case tea.KeyUp:
			if m.addKindIndex > 0 {
				m.addKindIndex--
			}
		case tea.KeyDown:
			if m.addKindIndex < len(kinds)-1 {
				m.addKindIndex++
			}
		case tea.KeyEnter:
			m.addStep++
		}
	case addStepSizeType:
		m.editText(&m.addSizeType, msg, nil)
	case addStepAC:
		m.editNumeric(&m.addACBuf, msg, nil)
	case addStepHP:
		m.editNumeric(&m.addHPBuf, msg, nil)
	case addStepSpeed:
		m.editText(&m.addSpeed, msg, nil)
	case addStepAbilities:
		m.editText(&m.addAbilBuf, msg, func() bool {
			m.finalizeAdd()
			return false // finalizeAdd handles mode transition
		})
	}
	return m, nil
}

// editText applies a keypress to a text buffer. onEnter (if non-nil) is called
// on Enter; when it returns true the form advances to the next step.
func (m *CompanionsModel) editText(buf *string, msg tea.KeyPressMsg, onEnter func() bool) {
	switch msg.Code {
	case tea.KeyEnter:
		if onEnter == nil {
			m.addStep++
			return
		}
		if onEnter() {
			m.addStep++
		}
	case tea.KeyBackspace:
		if len(*buf) > 0 {
			*buf = (*buf)[:len(*buf)-1]
		}
	default:
		if msg.Text != "" {
			*buf += msg.Text
		}
	}
}

// editNumeric is like editText but only accepts digits and a minus sign.
func (m *CompanionsModel) editNumeric(buf *string, msg tea.KeyPressMsg, onEnter func() bool) {
	switch msg.Code {
	case tea.KeyEnter:
		if onEnter == nil {
			m.addStep++
			return
		}
		if onEnter() {
			m.addStep++
		}
	case tea.KeyBackspace:
		if len(*buf) > 0 {
			*buf = (*buf)[:len(*buf)-1]
		}
	default:
		if msg.Text != "" && (isDigitStr(msg.Text) || (msg.Text == "-" && len(*buf) == 0)) {
			*buf += msg.Text
		}
	}
}

func (m *CompanionsModel) finalizeAdd() {
	kinds := models.AllCompanionKinds()
	kind := kinds[m.addKindIndex]

	ac, _ := strconv.Atoi(m.addACBuf)
	hp, _ := strconv.Atoi(m.addHPBuf)

	size, ctype := splitSizeType(m.addSizeType)

	comp := models.Companion{
		ID:        fmt.Sprintf("comp-%d", time.Now().UnixNano()),
		Name:      strings.TrimSpace(m.addName),
		Kind:      kind,
		Size:      size,
		Type:      ctype,
		AC:        ac,
		MaxHP:     hp,
		CurrentHP: hp,
		Speed:     strings.TrimSpace(m.addSpeed),
		Abilities: parseAbilities(m.addAbilBuf),
	}

	m.character.AddCompanion(comp)
	m.saveCharacter()
	m.cursor = len(m.character.Companions) - 1
	m.mode = companionModeList
	m.statusMessage = fmt.Sprintf("Added %s: %s", strings.ToLower(string(kind)), comp.Name)
}

func (m *CompanionsModel) handleAttackInput(msg tea.KeyPressMsg) (*CompanionsModel, tea.Cmd) {
	if msg.Code == tea.KeyEscape {
		m.mode = companionModeList
		m.statusMessage = "Attack cancelled"
		return m, nil
	}
	comp := m.selected()
	if comp == nil {
		m.mode = companionModeList
		return m, nil
	}

	switch m.atkStep {
	case atkStepName:
		switch msg.Code {
		case tea.KeyEnter:
			if strings.TrimSpace(m.atkName) == "" {
				m.statusMessage = "Attack name cannot be empty"
				return m, nil
			}
			m.atkStep++
		case tea.KeyBackspace:
			if len(m.atkName) > 0 {
				m.atkName = m.atkName[:len(m.atkName)-1]
			}
		default:
			if msg.Text != "" {
				m.atkName += msg.Text
			}
		}
	case atkStepBonus:
		switch msg.Code {
		case tea.KeyEnter:
			m.atkStep++
		case tea.KeyBackspace:
			if len(m.atkBonusBuf) > 0 {
				m.atkBonusBuf = m.atkBonusBuf[:len(m.atkBonusBuf)-1]
			}
		default:
			if msg.Text != "" && (isDigitStr(msg.Text) || (msg.Text == "-" && len(m.atkBonusBuf) == 0)) {
				m.atkBonusBuf += msg.Text
			}
		}
	case atkStepDamage:
		switch msg.Code {
		case tea.KeyEnter:
			bonus, _ := strconv.Atoi(m.atkBonusBuf)
			comp.Attacks = append(comp.Attacks, models.CompanionAttack{
				Name:   strings.TrimSpace(m.atkName),
				Bonus:  bonus,
				Damage: strings.TrimSpace(m.atkDamageBuf),
			})
			m.saveCharacter()
			m.mode = companionModeList
			m.statusMessage = fmt.Sprintf("Added attack: %s", strings.TrimSpace(m.atkName))
		case tea.KeyBackspace:
			if len(m.atkDamageBuf) > 0 {
				m.atkDamageBuf = m.atkDamageBuf[:len(m.atkDamageBuf)-1]
			}
		default:
			if msg.Text != "" {
				m.atkDamageBuf += msg.Text
			}
		}
	}
	return m, nil
}

func (m *CompanionsModel) handleHPInput(msg tea.KeyPressMsg) (*CompanionsModel, tea.Cmd) {
	if msg.Code == tea.KeyEscape {
		m.mode = companionModeList
		m.statusMessage = ""
		return m, nil
	}
	comp := m.selected()
	if comp == nil {
		m.mode = companionModeList
		return m, nil
	}

	switch msg.Code {
	case tea.KeyEnter:
		amount, _ := strconv.Atoi(m.hpBuf)
		if m.hpIsHeal {
			comp.Heal(amount)
			m.statusMessage = fmt.Sprintf("%s healed %d (now %d/%d)", comp.Name, amount, comp.CurrentHP, comp.MaxHP)
		} else {
			comp.Damage(amount)
			m.statusMessage = fmt.Sprintf("%s took %d damage (now %d/%d)", comp.Name, amount, comp.CurrentHP, comp.MaxHP)
		}
		m.saveCharacter()
		m.mode = companionModeList
	case tea.KeyBackspace:
		if len(m.hpBuf) > 0 {
			m.hpBuf = m.hpBuf[:len(m.hpBuf)-1]
		}
	default:
		if isDigitStr(msg.Text) {
			m.hpBuf += msg.Text
		}
	}
	return m, nil
}

func (m *CompanionsModel) performDelete() {
	if comp := m.selected(); comp != nil {
		name := comp.Name
		m.character.RemoveCompanion(comp.ID)
		m.saveCharacter()
		if m.cursor >= len(m.character.Companions) && m.cursor > 0 {
			m.cursor--
		}
		m.statusMessage = fmt.Sprintf("Deleted %s", name)
	}
	m.confirmingDelete = false
}

func (m *CompanionsModel) saveCharacter() {
	if m.storage != nil {
		_ = m.storage.AutoSave(m.character)
	}
}

// splitSizeType splits a "Size Type" string (e.g., "Medium Beast") into its
// size and type components. A single token is treated as the type.
func splitSizeType(s string) (size, ctype string) {
	fields := strings.Fields(s)
	switch len(fields) {
	case 0:
		return "", ""
	case 1:
		return "", fields[0]
	default:
		return fields[0], strings.Join(fields[1:], " ")
	}
}

// parseAbilities parses up to six space-separated ability scores. Missing or
// invalid entries default to 10.
func parseAbilities(s string) [6]int {
	abils := [6]int{10, 10, 10, 10, 10, 10}
	for i, f := range strings.Fields(s) {
		if i >= 6 {
			break
		}
		if v, err := strconv.Atoi(f); err == nil {
			abils[i] = v
		}
	}
	return abils
}

// View renders the companions view.
func (m *CompanionsModel) View() string {
	if m.character == nil {
		return "No character loaded"
	}

	width := m.width
	if width == 0 {
		width = 120
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	header := titleStyle.Render("🐾 Companions & Summons") + "  " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(m.character.Info.Name)

	var body string
	switch m.mode {
	case companionModeAdd:
		body = m.renderAddForm(width)
	case companionModeAttack:
		body = m.renderAttackForm(width)
	default:
		body = m.renderList(width)
	}

	footer := m.renderFooter(width)
	return lipgloss.JoinVertical(lipgloss.Left, header, "", body, "", footer)
}

func (m *CompanionsModel) renderList(width int) string {
	if len(m.character.Companions) == 0 {
		empty := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(
			"No companions yet.\n\nPress 'a' to add a pet, familiar, summon, mount, or Wild Shape form.")
		return empty
	}

	listWidth := width * 32 / 100
	if listWidth < 24 {
		listWidth = 24
	}
	if listWidth > 40 {
		listWidth = 40
	}

	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	var listLines []string
	for i := range m.character.Companions {
		c := &m.character.Companions[i]
		prefix := "  "
		style := normalStyle
		if i == m.cursor {
			prefix = "▶ "
			style = selectedStyle
		}
		hp := fmt.Sprintf("%d/%d", c.CurrentHP, c.MaxHP)
		line := fmt.Sprintf("%s%s", prefix, style.Render(c.Name))
		listLines = append(listLines, line)
		listLines = append(listLines, dimStyle.Render(fmt.Sprintf("   %s · HP %s", c.Kind, hp)))
	}

	listBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(0, 1).
		Width(listWidth).
		Render(strings.Join(listLines, "\n"))

	statBox := m.renderStatBlock(width - listWidth - 6)

	return lipgloss.JoinHorizontal(lipgloss.Top, listBox, statBox)
}

func (m *CompanionsModel) renderStatBlock(width int) string {
	if width < 24 {
		width = 24
	}
	comp := m.selected()
	if comp == nil {
		return ""
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("111"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	var lines []string
	lines = append(lines, titleStyle.Render(comp.Name))

	subtitle := string(comp.Kind)
	if comp.Size != "" || comp.Type != "" {
		subtitle += " · " + strings.TrimSpace(comp.Size+" "+comp.Type)
	}
	lines = append(lines, dimStyle.Render(subtitle))
	lines = append(lines, "")

	hpColor := lipgloss.Color("42")
	if comp.MaxHP > 0 && comp.CurrentHP*2 <= comp.MaxHP {
		hpColor = lipgloss.Color("214")
	}
	if comp.CurrentHP == 0 {
		hpColor = lipgloss.Color("196")
	}
	hpStr := fmt.Sprintf("%d/%d", comp.CurrentHP, comp.MaxHP)
	if comp.TempHP > 0 {
		hpStr += fmt.Sprintf(" (+%d temp)", comp.TempHP)
	}
	lines = append(lines, fmt.Sprintf("%s %s   %s %s",
		labelStyle.Render("AC"), valueStyle.Render(strconv.Itoa(comp.AC)),
		labelStyle.Render("HP"), lipgloss.NewStyle().Foreground(hpColor).Bold(true).Render(hpStr)))
	if comp.Speed != "" {
		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Speed"), valueStyle.Render(comp.Speed)))
	}
	lines = append(lines, "")

	// Ability row.
	var abilCells []string
	for i := 0; i < 6; i++ {
		mod := comp.Modifier(i)
		cell := fmt.Sprintf("%s %d (%s)", models.CompanionAbilityLabel(i), comp.Abilities[i], models.FormatModifier(mod))
		abilCells = append(abilCells, cell)
	}
	lines = append(lines, dimStyle.Render(strings.Join(abilCells[:3], "  ")))
	lines = append(lines, dimStyle.Render(strings.Join(abilCells[3:], "  ")))
	lines = append(lines, "")

	// Attacks.
	lines = append(lines, labelStyle.Render("Attacks"))
	if len(comp.Attacks) == 0 {
		lines = append(lines, dimStyle.Render("  (none — press 'w' to add)"))
	} else {
		for _, a := range comp.Attacks {
			atk := fmt.Sprintf("  %s: %s to hit", a.Name, models.FormatModifier(a.Bonus))
			if a.Damage != "" {
				atk += fmt.Sprintf(", %s", a.Damage)
			}
			lines = append(lines, valueStyle.Render(atk))
		}
	}

	// Traits.
	if len(comp.Traits) > 0 {
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Traits"))
		for _, t := range comp.Traits {
			lines = append(lines, valueStyle.Render("  "+t.Name))
			if t.Text != "" {
				lines = append(lines, dimStyle.Render("    "+t.Text))
			}
		}
	}

	if comp.Notes != "" {
		lines = append(lines, "")
		lines = append(lines, dimStyle.Render(comp.Notes))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(width).
		Render(strings.Join(lines, "\n"))
}

func (m *CompanionsModel) renderAddForm(width int) string {
	kinds := models.AllCompanionKinds()
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

	kindLabel := string(kinds[m.addKindIndex])
	if m.addStep == addStepKind {
		kindLabel = "◀ " + kindLabel + " ▶"
	}

	var lines []string
	lines = append(lines, titleStyle.Render("Add Companion"))
	lines = append(lines, "")
	lines = append(lines, field(addStepName, "Name", m.addName))
	lines = append(lines, field(addStepKind, "Kind", kindLabel))
	lines = append(lines, field(addStepSizeType, "Size/Type", m.addSizeType))
	lines = append(lines, field(addStepAC, "AC", m.addACBuf))
	lines = append(lines, field(addStepHP, "Max HP", m.addHPBuf))
	lines = append(lines, field(addStepSpeed, "Speed", m.addSpeed))
	lines = append(lines, field(addStepAbilities, "Abilities", m.addAbilBuf))
	lines = append(lines, "")
	lines = append(lines, dimStyle.Render("Abilities: 6 space-separated scores (STR DEX CON INT WIS CHA), blank = 10s"))
	lines = append(lines, dimStyle.Render("Size/Type example: \"Medium Beast\" · Enter: next/create · Esc: cancel"))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("214")).
		Padding(0, 1).
		Width(width - 4).
		Render(strings.Join(lines, "\n"))
}

func (m *CompanionsModel) renderAttackForm(width int) string {
	comp := m.selected()
	name := ""
	if comp != nil {
		name = comp.Name
	}
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	doneStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))

	field := func(step int, label, value string) string {
		style := dimStyle
		cursor := ""
		switch {
		case step == m.atkStep:
			style = activeStyle
			cursor = "_"
		case step < m.atkStep:
			style = doneStyle
		}
		return style.Render(fmt.Sprintf("%-10s %s%s", label+":", value, cursor))
	}

	var lines []string
	lines = append(lines, titleStyle.Render("Add Attack to "+name))
	lines = append(lines, "")
	lines = append(lines, field(atkStepName, "Name", m.atkName))
	lines = append(lines, field(atkStepBonus, "To hit", m.atkBonusBuf))
	lines = append(lines, field(atkStepDamage, "Damage", m.atkDamageBuf))
	lines = append(lines, "")
	lines = append(lines, dimStyle.Render("To hit: a number like 5 or -1 · Damage: e.g. \"2d6 + 3 slashing\""))
	lines = append(lines, dimStyle.Render("Enter: next/create · Esc: cancel"))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("214")).
		Padding(0, 1).
		Width(width - 4).
		Render(strings.Join(lines, "\n"))
}

func (m *CompanionsModel) renderFooter(width int) string {
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Width(width)
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).Width(width)

	var help string
	switch {
	case m.mode == companionModeAdd || m.mode == companionModeAttack:
		help = "type value • ↑/↓: change kind • enter: next/create • esc: cancel"
	case m.mode == companionModeHP:
		help = "type amount • enter: apply • esc: cancel"
	case m.confirmingDelete:
		help = "y: confirm delete • any other key: cancel"
	default:
		help = "↑/↓: select • a: add • w: attack • d: damage • h: heal • x: delete • esc: back"
	}

	var lines []string
	if m.statusMessage != "" {
		lines = append(lines, statusStyle.Render(m.statusMessage))
	}
	lines = append(lines, helpStyle.Render(help))
	return strings.Join(lines, "\n")
}

// CursorInfo returns cursor settings when a text-input mode is active.
func (m *CompanionsModel) CursorInfo() *tea.Cursor {
	if m.mode == companionModeAdd || m.mode == companionModeAttack || m.mode == companionModeHP {
		return &tea.Cursor{Shape: tea.CursorBar, Blink: true}
	}
	return nil
}
