package views

import (
	"fmt"
	"strings"

	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// OpenCharacterInfoMsg signals to open the character info view.
type OpenCharacterInfoMsg struct{}

// CharacterInfoFocus represents which panel is focused.
type CharacterInfoFocus int

const (
	CharInfoFocusPersonality CharacterInfoFocus = iota
	CharInfoFocusFeatures
)

// FeatureCategory represents the active feature tab.
type FeatureCategory int

const (
	FeatureCategoryRacial FeatureCategory = iota
	FeatureCategoryClass
	FeatureCategorySubclass
	FeatureCategoryFeats
)

// PersonalitySection represents sections in the personality panel.
type PersonalitySection int

const (
	PersonalitySectionTraits PersonalitySection = iota
	PersonalitySectionIdeals
	PersonalitySectionBonds
	PersonalitySectionFlaws
	PersonalitySectionBackstory
)

// CharacterInfoModel is the model for the character info view.
type CharacterInfoModel struct {
	character *models.Character
	storage   *storage.CharacterStorage
	width     int
	height    int
	keys      charInfoKeyMap

	// Focus
	focus CharacterInfoFocus

	// Features panel state
	featureCategory FeatureCategory
	featureCursor   int
	featureScroll   int

	// Personality panel state
	personalitySection PersonalitySection
	personalityCursor  int // index within current section's items
	personalityScroll  int

	// Edit modal state (for Task 7 — define fields but don't implement editing yet)
	editMode    bool
	editBuffer  string
	editAction  string // "edit", "add"
	editSection PersonalitySection
	editIndex   int

	// Backstory expanded
	backstoryExpanded bool

	// Delete confirmation
	confirmingDelete bool

	// Status
	statusMessage string
}

type charInfoKeyMap struct {
	Quit     key.Binding // ctrl+c
	Back     key.Binding // esc
	Tab      key.Binding // tab
	ShiftTab key.Binding // shift+tab
	Up       key.Binding // up, k
	Down     key.Binding // down, j
	Left     key.Binding // left, h
	Right    key.Binding // right, l
	Select   key.Binding // enter
	Edit     key.Binding // e
	Add      key.Binding // a
	Delete   key.Binding // d
	Notes    key.Binding // n
}

func defaultCharInfoKeyMap() charInfoKeyMap {
	return charInfoKeyMap{
		Quit:     key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
		Back:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next panel")),
		ShiftTab: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev panel")),
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Left:     key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "left")),
		Right:    key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "right")),
		Select:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Edit:     key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		Add:      key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
		Delete:   key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		Notes:    key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "notes")),
	}
}

// NewCharacterInfoModel creates a new character info model.
func NewCharacterInfoModel(char *models.Character, storage *storage.CharacterStorage) *CharacterInfoModel {
	return &CharacterInfoModel{
		character: char,
		storage:   storage,
		keys:      defaultCharInfoKeyMap(),
	}
}

// Init initializes the model.
func (m *CharacterInfoModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m *CharacterInfoModel) Update(msg tea.Msg) (*CharacterInfoModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Ctrl+C always quits
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}

		// Handle edit mode input
		if m.editMode {
			return m.handleEditMode(msg)
		}

		// Handle delete confirmation
		if m.confirmingDelete {
			return m.handleConfirmDelete(msg)
		}

		m.statusMessage = ""

		switch {
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return BackToSheetMsg{} }

		case key.Matches(msg, m.keys.Tab):
			if m.focus == CharInfoFocusPersonality {
				m.focus = CharInfoFocusFeatures
			} else {
				m.focus = CharInfoFocusPersonality
			}
			return m, nil

		case key.Matches(msg, m.keys.ShiftTab):
			if m.focus == CharInfoFocusPersonality {
				m.focus = CharInfoFocusFeatures
			} else {
				m.focus = CharInfoFocusPersonality
			}
			return m, nil

		case key.Matches(msg, m.keys.Notes):
			return m, func() tea.Msg { return OpenNotesMsg{ReturnTo: "charinfo"} }

		case key.Matches(msg, m.keys.Up):
			return m.handleUp()

		case key.Matches(msg, m.keys.Down):
			return m.handleDown()

		case key.Matches(msg, m.keys.Left):
			return m.handleLeft()

		case key.Matches(msg, m.keys.Right):
			return m.handleRight()

		case key.Matches(msg, m.keys.Select):
			return m.handleSelect()

		case key.Matches(msg, m.keys.Edit):
			if m.focus == CharInfoFocusPersonality {
				return m.startEdit()
			}

		case key.Matches(msg, m.keys.Add):
			if m.focus == CharInfoFocusPersonality {
				return m.startAdd()
			}

		case key.Matches(msg, m.keys.Delete):
			if m.focus == CharInfoFocusPersonality {
				return m.startDelete()
			}
		}
	}

	return m, nil
}

// handleUp moves the cursor up in the focused panel.
func (m *CharacterInfoModel) handleUp() (*CharacterInfoModel, tea.Cmd) {
	switch m.focus {
	case CharInfoFocusFeatures:
		if m.featureCursor > 0 {
			m.featureCursor--
		}
	case CharInfoFocusPersonality:
		m.movePersonalityCursorUp()
	}
	return m, nil
}

// handleDown moves the cursor down in the focused panel.
func (m *CharacterInfoModel) handleDown() (*CharacterInfoModel, tea.Cmd) {
	switch m.focus {
	case CharInfoFocusFeatures:
		features := m.getFeaturesForCategory()
		if m.featureCursor < len(features)-1 {
			m.featureCursor++
		}
	case CharInfoFocusPersonality:
		m.movePersonalityCursorDown()
	}
	return m, nil
}

// handleLeft switches category tab left when features panel is focused.
func (m *CharacterInfoModel) handleLeft() (*CharacterInfoModel, tea.Cmd) {
	if m.focus == CharInfoFocusFeatures {
		if m.featureCategory > FeatureCategoryRacial {
			m.featureCategory--
			m.featureCursor = 0
			m.featureScroll = 0
		}
	}
	return m, nil
}

// handleRight switches category tab right when features panel is focused.
func (m *CharacterInfoModel) handleRight() (*CharacterInfoModel, tea.Cmd) {
	if m.focus == CharInfoFocusFeatures {
		if m.featureCategory < FeatureCategoryFeats {
			m.featureCategory++
			m.featureCursor = 0
			m.featureScroll = 0
		}
	}
	return m, nil
}

// handleSelect handles enter key.
func (m *CharacterInfoModel) handleSelect() (*CharacterInfoModel, tea.Cmd) {
	if m.focus == CharInfoFocusPersonality && m.personalitySection == PersonalitySectionBackstory {
		m.backstoryExpanded = !m.backstoryExpanded
	}
	return m, nil
}

// handleEditMode handles key input when in edit mode.
func (m *CharacterInfoModel) handleEditMode(msg tea.KeyMsg) (*CharacterInfoModel, tea.Cmd) {
	if m.editSection == PersonalitySectionBackstory {
		// Multiline editing
		switch msg.Type {
		case tea.KeyEsc:
			m.editMode = false
			return m, nil
		case tea.KeyEnter:
			m.editBuffer += "\n"
			return m, nil
		case tea.KeyBackspace:
			if len(m.editBuffer) > 0 {
				m.editBuffer = m.editBuffer[:len(m.editBuffer)-1]
			}
			return m, nil
		case tea.KeyRunes:
			m.editBuffer += string(msg.Runes)
			return m, nil
		default:
			// Handle ctrl+s for save
			if msg.String() == "ctrl+s" {
				m.applyEdit()
				m.editMode = false
			}
			return m, nil
		}
	}

	// Single-line editing
	switch msg.Type {
	case tea.KeyEsc:
		m.editMode = false
		return m, nil
	case tea.KeyEnter:
		m.applyEdit()
		m.editMode = false
		return m, nil
	case tea.KeyBackspace:
		if len(m.editBuffer) > 0 {
			m.editBuffer = m.editBuffer[:len(m.editBuffer)-1]
		}
		return m, nil
	case tea.KeyRunes:
		m.editBuffer += string(msg.Runes)
		return m, nil
	}
	return m, nil
}

// handleConfirmDelete handles key input during delete confirmation.
func (m *CharacterInfoModel) handleConfirmDelete(msg tea.KeyMsg) (*CharacterInfoModel, tea.Cmd) {
	switch {
	case msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'y':
		m.deleteCurrentItem()
		m.confirmingDelete = false
		return m, nil
	case msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'n',
		msg.Type == tea.KeyEsc:
		m.confirmingDelete = false
		return m, nil
	}
	return m, nil
}

// startEdit begins editing the currently focused personality item.
func (m *CharacterInfoModel) startEdit() (*CharacterInfoModel, tea.Cmd) {
	items := m.getAllPersonalityItems()
	if m.personalityCursor < 0 || m.personalityCursor >= len(items) {
		return m, nil
	}

	item := items[m.personalityCursor]
	if item.header {
		return m, nil
	}

	m.editMode = true
	m.editAction = "edit"
	m.editSection = item.section
	m.editIndex = item.index

	if item.text == "(none)" {
		m.editBuffer = ""
	} else {
		m.editBuffer = item.text
	}

	return m, nil
}

// startAdd begins adding a new item to the current personality section.
func (m *CharacterInfoModel) startAdd() (*CharacterInfoModel, tea.Cmd) {
	// Determine which section we're in based on cursor position
	items := m.getAllPersonalityItems()
	if m.personalityCursor < 0 || m.personalityCursor >= len(items) {
		return m, nil
	}

	section := items[m.personalityCursor].section

	m.editMode = true
	m.editAction = "add"
	m.editSection = section
	m.editIndex = -1
	m.editBuffer = ""

	return m, nil
}

// startDelete begins delete confirmation for the currently focused item.
func (m *CharacterInfoModel) startDelete() (*CharacterInfoModel, tea.Cmd) {
	items := m.getAllPersonalityItems()
	if m.personalityCursor < 0 || m.personalityCursor >= len(items) {
		return m, nil
	}

	item := items[m.personalityCursor]
	if item.header {
		return m, nil
	}

	m.confirmingDelete = true
	return m, nil
}

// applyEdit saves the current edit to the character.
func (m *CharacterInfoModel) applyEdit() {
	text := strings.TrimSpace(m.editBuffer)
	if text == "" && m.editSection != PersonalitySectionBackstory {
		return // don't save empty items (except backstory can be cleared)
	}

	switch m.editSection {
	case PersonalitySectionTraits:
		if m.editAction == "add" {
			m.character.Personality.AddTrait(text)
		} else {
			if m.editIndex >= 0 && m.editIndex < len(m.character.Personality.Traits) {
				m.character.Personality.Traits[m.editIndex] = text
			}
		}
	case PersonalitySectionIdeals:
		if m.editAction == "add" {
			m.character.Personality.AddIdeal(text)
		} else {
			if m.editIndex >= 0 && m.editIndex < len(m.character.Personality.Ideals) {
				m.character.Personality.Ideals[m.editIndex] = text
			}
		}
	case PersonalitySectionBonds:
		if m.editAction == "add" {
			m.character.Personality.AddBond(text)
		} else {
			if m.editIndex >= 0 && m.editIndex < len(m.character.Personality.Bonds) {
				m.character.Personality.Bonds[m.editIndex] = text
			}
		}
	case PersonalitySectionFlaws:
		if m.editAction == "add" {
			m.character.Personality.AddFlaw(text)
		} else {
			if m.editIndex >= 0 && m.editIndex < len(m.character.Personality.Flaws) {
				m.character.Personality.Flaws[m.editIndex] = text
			}
		}
	case PersonalitySectionBackstory:
		m.character.Personality.Backstory = text
	}

	m.saveCharacter()
}

// deleteCurrentItem deletes the currently focused personality item.
func (m *CharacterInfoModel) deleteCurrentItem() {
	items := m.getAllPersonalityItems()
	if m.personalityCursor < 0 || m.personalityCursor >= len(items) {
		return
	}

	item := items[m.personalityCursor]
	if item.header {
		return
	}

	switch item.section {
	case PersonalitySectionTraits:
		if item.index >= 0 && item.index < len(m.character.Personality.Traits) {
			m.character.Personality.Traits = append(
				m.character.Personality.Traits[:item.index],
				m.character.Personality.Traits[item.index+1:]...,
			)
		}
	case PersonalitySectionIdeals:
		if item.index >= 0 && item.index < len(m.character.Personality.Ideals) {
			m.character.Personality.Ideals = append(
				m.character.Personality.Ideals[:item.index],
				m.character.Personality.Ideals[item.index+1:]...,
			)
		}
	case PersonalitySectionBonds:
		if item.index >= 0 && item.index < len(m.character.Personality.Bonds) {
			m.character.Personality.Bonds = append(
				m.character.Personality.Bonds[:item.index],
				m.character.Personality.Bonds[item.index+1:]...,
			)
		}
	case PersonalitySectionFlaws:
		if item.index >= 0 && item.index < len(m.character.Personality.Flaws) {
			m.character.Personality.Flaws = append(
				m.character.Personality.Flaws[:item.index],
				m.character.Personality.Flaws[item.index+1:]...,
			)
		}
	case PersonalitySectionBackstory:
		m.character.Personality.Backstory = ""
	}

	m.saveCharacter()

	// Adjust cursor - rebuild items after deletion and clamp
	newItems := m.getAllPersonalityItems()
	if m.personalityCursor >= len(newItems) && m.personalityCursor > 0 {
		m.personalityCursor--
	}
	m.updatePersonalitySection()
}

// getCurrentSectionItems returns items for the current personality section.
func (m *CharacterInfoModel) getCurrentSectionItems() []string {
	if m.character == nil {
		return nil
	}
	switch m.personalitySection {
	case PersonalitySectionTraits:
		return m.character.Personality.Traits
	case PersonalitySectionIdeals:
		return m.character.Personality.Ideals
	case PersonalitySectionBonds:
		return m.character.Personality.Bonds
	case PersonalitySectionFlaws:
		return m.character.Personality.Flaws
	case PersonalitySectionBackstory:
		return []string{m.character.Personality.Backstory}
	}
	return nil
}

// saveCharacter saves the character to storage.
func (m *CharacterInfoModel) saveCharacter() {
	if m.storage != nil {
		_ = m.storage.AutoSave(m.character)
	}
}

// sectionName returns a human-readable name for a personality section.
func sectionName(section PersonalitySection) string {
	switch section {
	case PersonalitySectionTraits:
		return "Trait"
	case PersonalitySectionIdeals:
		return "Ideal"
	case PersonalitySectionBonds:
		return "Bond"
	case PersonalitySectionFlaws:
		return "Flaw"
	case PersonalitySectionBackstory:
		return "Backstory"
	}
	return "Item"
}

// movePersonalityCursorUp moves the personality cursor up across sections.
func (m *CharacterInfoModel) movePersonalityCursorUp() {
	items := m.getAllPersonalityItems()
	if len(items) == 0 {
		return
	}
	if m.personalityCursor > 0 {
		m.personalityCursor--
	}
	m.updatePersonalitySection()
}

// movePersonalityCursorDown moves the personality cursor down across sections.
func (m *CharacterInfoModel) movePersonalityCursorDown() {
	items := m.getAllPersonalityItems()
	if len(items) == 0 {
		return
	}
	if m.personalityCursor < len(items)-1 {
		m.personalityCursor++
	}
	m.updatePersonalitySection()
}

// personalityItem represents an item in the personality panel for unified cursor navigation.
type personalityItem struct {
	section PersonalitySection
	index   int
	text    string
	header  bool // true if this is a section header
}

// getAllPersonalityItems returns all personality items in display order for cursor navigation.
func (m *CharacterInfoModel) getAllPersonalityItems() []personalityItem {
	if m.character == nil {
		return nil
	}

	var items []personalityItem
	p := &m.character.Personality

	// Traits
	items = append(items, personalityItem{section: PersonalitySectionTraits, index: -1, text: "Traits:", header: true})
	if len(p.Traits) == 0 {
		items = append(items, personalityItem{section: PersonalitySectionTraits, index: 0, text: "(none)"})
	} else {
		for i, t := range p.Traits {
			items = append(items, personalityItem{section: PersonalitySectionTraits, index: i, text: t})
		}
	}

	// Ideals
	items = append(items, personalityItem{section: PersonalitySectionIdeals, index: -1, text: "Ideals:", header: true})
	if len(p.Ideals) == 0 {
		items = append(items, personalityItem{section: PersonalitySectionIdeals, index: 0, text: "(none)"})
	} else {
		for i, t := range p.Ideals {
			items = append(items, personalityItem{section: PersonalitySectionIdeals, index: i, text: t})
		}
	}

	// Bonds
	items = append(items, personalityItem{section: PersonalitySectionBonds, index: -1, text: "Bonds:", header: true})
	if len(p.Bonds) == 0 {
		items = append(items, personalityItem{section: PersonalitySectionBonds, index: 0, text: "(none)"})
	} else {
		for i, t := range p.Bonds {
			items = append(items, personalityItem{section: PersonalitySectionBonds, index: i, text: t})
		}
	}

	// Flaws
	items = append(items, personalityItem{section: PersonalitySectionFlaws, index: -1, text: "Flaws:", header: true})
	if len(p.Flaws) == 0 {
		items = append(items, personalityItem{section: PersonalitySectionFlaws, index: 0, text: "(none)"})
	} else {
		for i, t := range p.Flaws {
			items = append(items, personalityItem{section: PersonalitySectionFlaws, index: i, text: t})
		}
	}

	// Backstory
	items = append(items, personalityItem{section: PersonalitySectionBackstory, index: -1, text: "Backstory:", header: true})
	if p.Backstory == "" {
		items = append(items, personalityItem{section: PersonalitySectionBackstory, index: 0, text: "(none)"})
	} else {
		items = append(items, personalityItem{section: PersonalitySectionBackstory, index: 0, text: p.Backstory})
	}

	return items
}

// updatePersonalitySection updates the current section based on cursor position.
func (m *CharacterInfoModel) updatePersonalitySection() {
	items := m.getAllPersonalityItems()
	if m.personalityCursor >= 0 && m.personalityCursor < len(items) {
		m.personalitySection = items[m.personalityCursor].section
	}
}

// getFeaturesForCategory returns features for the current category.
func (m *CharacterInfoModel) getFeaturesForCategory() []models.Feature {
	if m.character == nil {
		return nil
	}

	switch m.featureCategory {
	case FeatureCategoryRacial:
		return m.character.Features.RacialTraits
	case FeatureCategoryClass:
		var features []models.Feature
		subclass := m.character.Info.Subclass
		for _, f := range m.character.Features.ClassFeatures {
			if subclass != "" && strings.Contains(f.Source, subclass) {
				continue // exclude subclass features
			}
			features = append(features, f)
		}
		return features
	case FeatureCategorySubclass:
		var features []models.Feature
		subclass := m.character.Info.Subclass
		if subclass == "" {
			return features
		}
		for _, f := range m.character.Features.ClassFeatures {
			if strings.Contains(f.Source, subclass) {
				features = append(features, f)
			}
		}
		return features
	case FeatureCategoryFeats:
		return m.character.Features.Feats
	}
	return nil
}

// View renders the character info view.
func (m *CharacterInfoModel) View() string {
	if m.character == nil {
		return "No character loaded"
	}

	width := m.width
	if width == 0 {
		width = 120
	}
	height := m.height
	if height == 0 {
		height = 40
	}

	// Header
	header := m.renderHeader(width)

	// Two-panel layout: Personality (45%) | Features (55%)
	gap := 2
	leftWidth := (width - gap) * 45 / 100
	rightWidth := width - leftWidth - gap

	personality := m.renderPersonalityPanel(leftWidth)
	features := m.renderFeaturesPanel(rightWidth)

	gapStr := strings.Repeat(" ", gap)
	columns := lipgloss.JoinHorizontal(lipgloss.Top, personality, gapStr, features)

	// Footer
	footer := m.renderFooter(width)

	base := lipgloss.JoinVertical(lipgloss.Left, header, "", columns, "", footer)

	// Overlay edit modal if in edit mode
	if m.editMode {
		modal := m.renderEditModal(width, height)
		base = lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modal,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("236")),
		)
	}

	// Show delete confirmation overlay
	if m.confirmingDelete {
		items := m.getAllPersonalityItems()
		sectionLabel := "item"
		if m.personalityCursor >= 0 && m.personalityCursor < len(items) {
			sectionLabel = strings.ToLower(sectionName(items[m.personalityCursor].section))
		}
		confirmModal := m.renderConfirmDeleteModal(width, height, sectionLabel)
		base = lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, confirmModal,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("236")),
		)
	}

	return base
}

func (m *CharacterInfoModel) renderHeader(width int) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99"))

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	title := titleStyle.Render(fmt.Sprintf("%s — Character Info", m.character.Info.Name))

	details := []string{
		m.character.Info.Race,
		fmt.Sprintf("%s %d", m.character.Info.Class, m.character.Info.Level),
	}
	if m.character.Info.Subclass != "" {
		details = append(details, m.character.Info.Subclass)
	}
	if m.character.Info.Background != "" {
		details = append(details, m.character.Info.Background)
	}

	info := infoStyle.Render(strings.Join(details, " • "))

	return lipgloss.JoinHorizontal(lipgloss.Center,
		title,
		strings.Repeat(" ", 4),
		info,
	)
}

func (m *CharacterInfoModel) renderPersonalityPanel(width int) string {
	focused := m.focus == CharInfoFocusPersonality

	borderColor := lipgloss.Color("240")
	if focused {
		borderColor = lipgloss.Color("99")
	}

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(width)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)

	var lines []string
	lines = append(lines, titleStyle.Render("Personality"))
	lines = append(lines, "")

	p := &m.character.Personality
	items := m.getAllPersonalityItems()

	// Content width inside panel (accounting for border and padding)
	contentWidth := width - 4
	if contentWidth < 20 {
		contentWidth = 20
	}

	for idx, item := range items {
		if item.header {
			lines = append(lines, headerStyle.Render(item.text))
			continue
		}

		prefix := "  "
		style := itemStyle

		if focused && idx == m.personalityCursor {
			prefix = "> "
			style = selectedStyle
		}

		if item.text == "(none)" {
			lines = append(lines, prefix+dimStyle.Render(item.text))
			continue
		}

		if item.section == PersonalitySectionBackstory {
			// Backstory: word-wrapped, truncated if long and not expanded
			backstory := p.Backstory
			if backstory == "" {
				lines = append(lines, prefix+dimStyle.Render("(none)"))
			} else {
				wrapped := wordWrapText(backstory, contentWidth-4)
				if !m.backstoryExpanded && len(wrapped) > 3 {
					wrapped = wrapped[:3]
					wrapped = append(wrapped, "...")
				}
				for i, wl := range wrapped {
					if i == 0 {
						lines = append(lines, prefix+style.Render("• "+wl))
					} else {
						lines = append(lines, "    "+style.Render(wl))
					}
				}
			}
		} else {
			lines = append(lines, prefix+style.Render("• "+item.text))
		}
	}

	return panelStyle.Render(strings.Join(lines, "\n"))
}

func (m *CharacterInfoModel) renderFeaturesPanel(width int) string {
	focused := m.focus == CharInfoFocusFeatures

	borderColor := lipgloss.Color("240")
	if focused {
		borderColor = lipgloss.Color("99")
	}

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(width)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	boldStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))

	var lines []string
	lines = append(lines, titleStyle.Render("Features"))
	lines = append(lines, "")

	// Category tabs
	lines = append(lines, m.renderCategoryTabs())
	lines = append(lines, "")

	// Content width inside panel
	contentWidth := width - 4
	if contentWidth < 20 {
		contentWidth = 20
	}

	// Feature list
	features := m.getFeaturesForCategory()
	if len(features) == 0 {
		lines = append(lines, dimStyle.Render("  No features in this category"))
	} else {
		for i, f := range features {
			prefix := "  "
			style := itemStyle

			if focused && i == m.featureCursor {
				prefix = "> "
				style = selectedStyle
			}

			lines = append(lines, prefix+style.Render(f.Name))
		}
	}

	// Detail pane
	if len(features) > 0 && m.featureCursor < len(features) {
		selected := features[m.featureCursor]

		lines = append(lines, "")
		lines = append(lines, dimStyle.Render(strings.Repeat("─", contentWidth)))
		lines = append(lines, boldStyle.Render(selected.Name))
		lines = append(lines, dimStyle.Render("Source: "+selected.Source))

		if selected.Level > 0 {
			lines = append(lines, dimStyle.Render(fmt.Sprintf("Level: %d", selected.Level)))
		}
		if selected.Activation != "" {
			lines = append(lines, dimStyle.Render(fmt.Sprintf("Activation: %s", string(selected.Activation))))
		}

		if selected.Description != "" {
			lines = append(lines, "")
			wrapped := wordWrapText(selected.Description, contentWidth)
			for _, wl := range wrapped {
				lines = append(lines, itemStyle.Render(wl))
			}
		}
	}

	return panelStyle.Render(strings.Join(lines, "\n"))
}

func (m *CharacterInfoModel) renderCategoryTabs() string {
	categories := []struct {
		label    string
		category FeatureCategory
	}{
		{"Racial", FeatureCategoryRacial},
		{"Class", FeatureCategoryClass},
		{"Subclass", FeatureCategorySubclass},
		{"Feats", FeatureCategoryFeats},
	}

	activeStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	var tabs []string
	for _, cat := range categories {
		style := inactiveStyle
		if cat.category == m.featureCategory {
			style = activeStyle
		}
		tabs = append(tabs, style.Render("["+cat.label+"]"))
	}

	return strings.Join(tabs, " ")
}

// renderEditModal renders the centered edit modal overlay.
func (m *CharacterInfoModel) renderEditModal(width, height int) string {
	modalWidth := width * 50 / 100
	if modalWidth < 40 {
		modalWidth = 40
	}
	if modalWidth > 80 {
		modalWidth = 80
	}

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1, 2).
		Width(modalWidth)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	// Title
	action := "Edit"
	if m.editAction == "add" {
		action = "Add"
	}
	title := titleStyle.Render(fmt.Sprintf("%s %s", action, sectionName(m.editSection)))

	// Input area
	displayText := m.editBuffer + "█"
	input := inputStyle.Render(displayText)

	// Help text
	var helpText string
	if m.editSection == PersonalitySectionBackstory {
		helpText = dimStyle.Render("Ctrl+S: save • Esc: cancel • Enter: newline")
	} else {
		helpText = dimStyle.Render("Enter: save • Esc: cancel")
	}

	content := lipgloss.JoinVertical(lipgloss.Left, title, "", input, "", helpText)

	return borderStyle.Render(content)
}

// renderConfirmDeleteModal renders a delete confirmation modal.
func (m *CharacterInfoModel) renderConfirmDeleteModal(width, height int, sectionLabel string) string {
	modalWidth := 50
	if modalWidth > width-4 {
		modalWidth = width - 4
	}

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 2).
		Width(modalWidth)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	title := titleStyle.Render(fmt.Sprintf("Delete this %s?", sectionLabel))
	help := dimStyle.Render("y: confirm • n/Esc: cancel")

	content := lipgloss.JoinVertical(lipgloss.Left, title, "", help)

	return borderStyle.Render(content)
}

func (m *CharacterInfoModel) renderFooter(width int) string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Width(width)

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("42")).
		Bold(true).
		Width(width)

	var help string
	switch m.focus {
	case CharInfoFocusPersonality:
		help = "↑/↓: navigate • e: edit • a: add • d: delete • tab: features panel • n: notes • esc: back"
	case CharInfoFocusFeatures:
		help = "←/→: category • ↑/↓: navigate • tab: personality panel • n: notes • esc: back"
	}

	var lines []string
	if m.statusMessage != "" {
		lines = append(lines, statusStyle.Render(m.statusMessage))
	}
	lines = append(lines, helpStyle.Render(help))

	return strings.Join(lines, "\n")
}

// wordWrapText wraps text to fit within the given width.
func wordWrapText(text string, width int) []string {
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
