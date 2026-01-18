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

// InventoryFocus represents which panel is focused in the inventory view.
type InventoryFocus int

const (
	FocusEquipment InventoryFocus = iota
	FocusItems
	FocusCurrency
	numInventoryFocusAreas
)

// InventoryModel is the model for the inventory view.
type InventoryModel struct {
	character     *models.Character
	storage       *storage.CharacterStorage
	width         int
	height        int
	focus         InventoryFocus
	keys          inventoryKeyMap
	statusMessage string

	// Item list navigation
	itemCursor    int
	itemScroll    int
	selectedItem  *models.Item
	itemsPerPage  int

	// Equipment slot navigation
	equipCursor int

	// Currency editing
	currencyMode   bool
	currencyType   int // 0=CP, 1=SP, 2=EP, 3=GP, 4=PP
	currencyBuffer string
	currencyAdding bool
}

type inventoryKeyMap struct {
	Quit     key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Back     key.Binding
	Add      key.Binding
	Remove   key.Binding
	Equip    key.Binding
}

func defaultInventoryKeyMap() inventoryKeyMap {
	return inventoryKeyMap{
		Quit:     key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next panel")),
		ShiftTab: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev panel")),
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Back:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Add:      key.NewBinding(key.WithKeys("+", "a"), key.WithHelp("+/a", "add")),
		Remove:   key.NewBinding(key.WithKeys("-", "d"), key.WithHelp("-/d", "remove")),
		Equip:    key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "equip/unequip")),
	}
}

// NewInventoryModel creates a new inventory model.
func NewInventoryModel(char *models.Character, storage *storage.CharacterStorage) *InventoryModel {
	return &InventoryModel{
		character:    char,
		storage:      storage,
		keys:         defaultInventoryKeyMap(),
		itemsPerPage: 15,
	}
}

// Init initializes the model.
func (m *InventoryModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m *InventoryModel) Update(msg tea.Msg) (*InventoryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle currency input mode
		if m.currencyMode {
			return m.handleCurrencyInput(msg)
		}

		m.statusMessage = ""

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return BackToSheetMsg{} }
		case key.Matches(msg, m.keys.Tab):
			m.focus = (m.focus + 1) % numInventoryFocusAreas
			return m, nil
		case key.Matches(msg, m.keys.ShiftTab):
			if m.focus == 0 {
				m.focus = numInventoryFocusAreas - 1
			} else {
				m.focus--
			}
			return m, nil
		case key.Matches(msg, m.keys.Up):
			return m.handleUp()
		case key.Matches(msg, m.keys.Down):
			return m.handleDown()
		case key.Matches(msg, m.keys.Enter):
			return m.handleEnter()
		case key.Matches(msg, m.keys.Add):
			return m.handleAdd()
		case key.Matches(msg, m.keys.Remove):
			return m.handleRemove()
		case key.Matches(msg, m.keys.Equip):
			return m.handleEquip()
		}
	}

	return m, nil
}

func (m *InventoryModel) handleUp() (*InventoryModel, tea.Cmd) {
	switch m.focus {
	case FocusEquipment:
		if m.equipCursor > 0 {
			m.equipCursor--
		}
	case FocusItems:
		if m.itemCursor > 0 {
			m.itemCursor--
			if m.itemCursor < m.itemScroll {
				m.itemScroll = m.itemCursor
			}
		}
	case FocusCurrency:
		if m.currencyType > 0 {
			m.currencyType--
		}
	}
	return m, nil
}

func (m *InventoryModel) handleDown() (*InventoryModel, tea.Cmd) {
	switch m.focus {
	case FocusEquipment:
		maxSlots := 9 // 8 slots + rings section
		if m.equipCursor < maxSlots-1 {
			m.equipCursor++
		}
	case FocusItems:
		itemCount := len(m.character.Inventory.Items)
		if m.itemCursor < itemCount-1 {
			m.itemCursor++
			if m.itemCursor >= m.itemScroll+m.itemsPerPage {
				m.itemScroll++
			}
		}
	case FocusCurrency:
		if m.currencyType < 4 {
			m.currencyType++
		}
	}
	return m, nil
}

func (m *InventoryModel) handleEnter() (*InventoryModel, tea.Cmd) {
	switch m.focus {
	case FocusItems:
		items := m.character.Inventory.Items
		if m.itemCursor < len(items) {
			m.selectedItem = &items[m.itemCursor]
			m.statusMessage = fmt.Sprintf("Selected: %s", m.selectedItem.Name)
		}
	case FocusCurrency:
		m.currencyMode = true
		m.currencyBuffer = ""
		m.currencyAdding = true
	}
	return m, nil
}

func (m *InventoryModel) handleAdd() (*InventoryModel, tea.Cmd) {
	switch m.focus {
	case FocusItems:
		if m.selectedItem != nil {
			m.selectedItem.Quantity++
			m.saveCharacter()
			m.statusMessage = fmt.Sprintf("Added 1 %s (now %d)", m.selectedItem.Name, m.selectedItem.Quantity)
		}
	case FocusCurrency:
		m.currencyMode = true
		m.currencyBuffer = ""
		m.currencyAdding = true
	}
	return m, nil
}

func (m *InventoryModel) handleRemove() (*InventoryModel, tea.Cmd) {
	switch m.focus {
	case FocusItems:
		if m.selectedItem != nil && m.selectedItem.Quantity > 0 {
			m.selectedItem.Quantity--
			if m.selectedItem.Quantity == 0 {
				// Remove item from inventory
				m.character.Inventory.RemoveItem(m.selectedItem.ID)
				m.selectedItem = nil
				if m.itemCursor > 0 {
					m.itemCursor--
				}
				m.statusMessage = "Item removed from inventory"
			} else {
				m.statusMessage = fmt.Sprintf("Removed 1 (now %d)", m.selectedItem.Quantity)
			}
			m.saveCharacter()
		}
	case FocusCurrency:
		m.currencyMode = true
		m.currencyBuffer = ""
		m.currencyAdding = false
	}
	return m, nil
}

func (m *InventoryModel) handleEquip() (*InventoryModel, tea.Cmd) {
	if m.focus != FocusItems || m.selectedItem == nil {
		return m, nil
	}

	item := m.selectedItem
	slot := item.EquipmentSlot

	if slot == "" {
		m.statusMessage = "This item cannot be equipped"
		return m, nil
	}

	// Check if already equipped
	equipped := m.character.Inventory.Equipment.GetSlot(slot)
	if equipped != nil && equipped.ID == item.ID {
		// Unequip
		m.character.Inventory.Equipment.SetSlot(slot, nil)
		m.statusMessage = fmt.Sprintf("Unequipped %s", item.Name)
	} else {
		// Equip
		m.character.Inventory.Equipment.SetSlot(slot, item)
		m.statusMessage = fmt.Sprintf("Equipped %s to %s", item.Name, slot)
	}

	m.saveCharacter()
	return m, nil
}

func (m *InventoryModel) handleCurrencyInput(msg tea.KeyMsg) (*InventoryModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.currencyMode = false
		m.currencyBuffer = ""
		return m, nil

	case tea.KeyEnter:
		if m.currencyBuffer != "" {
			var amount int
			fmt.Sscanf(m.currencyBuffer, "%d", &amount)
			if amount > 0 {
				currency := &m.character.Inventory.Currency
				if m.currencyAdding {
					switch m.currencyType {
					case 0:
						currency.AddCopper(amount)
					case 1:
						currency.AddSilver(amount)
					case 2:
						currency.AddElectrum(amount)
					case 3:
						currency.AddGold(amount)
					case 4:
						currency.AddPlatinum(amount)
					}
					m.statusMessage = fmt.Sprintf("Added %d %s", amount, currencyName(m.currencyType))
				} else {
					var err error
					switch m.currencyType {
					case 0:
						err = currency.SpendCopper(amount)
					case 1:
						err = currency.SpendSilver(amount)
					case 2:
						err = currency.SpendElectrum(amount)
					case 3:
						err = currency.SpendGold(amount)
					case 4:
						err = currency.SpendPlatinum(amount)
					}
					if err != nil {
						m.statusMessage = "Insufficient funds"
					} else {
						m.statusMessage = fmt.Sprintf("Spent %d %s", amount, currencyName(m.currencyType))
					}
				}
				m.saveCharacter()
			}
		}
		m.currencyMode = false
		m.currencyBuffer = ""
		return m, nil

	case tea.KeyBackspace:
		if len(m.currencyBuffer) > 0 {
			m.currencyBuffer = m.currencyBuffer[:len(m.currencyBuffer)-1]
		}
		return m, nil

	case tea.KeyRunes:
		for _, r := range msg.Runes {
			if r >= '0' && r <= '9' {
				m.currencyBuffer += string(r)
			}
		}
		return m, nil
	}

	return m, nil
}

func currencyName(idx int) string {
	names := []string{"CP", "SP", "EP", "GP", "PP"}
	if idx >= 0 && idx < len(names) {
		return names[idx]
	}
	return ""
}

func (m *InventoryModel) saveCharacter() {
	if m.storage != nil && m.character != nil {
		_ = m.storage.AutoSave(m.character)
	}
}

// View renders the inventory view.
func (m *InventoryModel) View() string {
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

	// Three-column layout: Equipment | Items | Currency
	equipWidth := 30
	currencyWidth := 25
	itemsWidth := width - equipWidth - currencyWidth - 6 // borders and padding

	equipment := m.renderEquipment(equipWidth)
	items := m.renderItems(itemsWidth)
	currency := m.renderCurrency(currencyWidth)

	columns := lipgloss.JoinHorizontal(lipgloss.Top, equipment, items, currency)

	// Footer
	footer := m.renderFooter(width)

	return lipgloss.JoinVertical(lipgloss.Left, header, "", columns, "", footer)
}

func (m *InventoryModel) renderHeader(width int) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99"))

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	// Calculate carried weight
	totalWeight := m.character.Inventory.TotalWeight()
	strScore := m.character.AbilityScores.Strength.Total()
	carryCapacity := strScore * 15 // Standard carrying capacity

	title := titleStyle.Render(fmt.Sprintf("📦 %s's Inventory", m.character.Info.Name))
	weight := infoStyle.Render(fmt.Sprintf("Weight: %.1f / %d lbs", totalWeight, carryCapacity))

	// Total value
	gp, cp := m.character.Inventory.Currency.TotalInGold()
	value := infoStyle.Render(fmt.Sprintf("Total: %d GP %d CP", gp, cp))

	return lipgloss.JoinHorizontal(lipgloss.Center,
		title,
		strings.Repeat(" ", 4),
		weight,
		strings.Repeat(" ", 4),
		value,
	)
}

func (m *InventoryModel) renderEquipment(width int) string {
	focused := m.focus == FocusEquipment

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
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)

	var lines []string
	lines = append(lines, titleStyle.Render("⚔️  Equipment"))
	lines = append(lines, "")

	slots := []struct {
		name string
		slot models.EquipmentSlot
	}{
		{"Main Hand", models.SlotMainHand},
		{"Off Hand", models.SlotOffHand},
		{"Head", models.SlotHead},
		{"Body", models.SlotBody},
		{"Cloak", models.SlotCloak},
		{"Gloves", models.SlotGloves},
		{"Boots", models.SlotBoots},
		{"Amulet", models.SlotAmulet},
	}

	equip := &m.character.Inventory.Equipment

	for i, s := range slots {
		item := equip.GetSlot(s.slot)
		var itemName string
		if item != nil {
			itemName = item.Name
			if item.Magical {
				itemName = "✨ " + itemName
			}
		} else {
			itemName = "-"
		}

		label := labelStyle.Render(fmt.Sprintf("%-10s", s.name))
		value := valueStyle.Render(itemName)

		if focused && m.equipCursor == i {
			value = selectedStyle.Render("▶ " + itemName)
		}

		lines = append(lines, fmt.Sprintf("%s %s", label, value))
	}

	// Rings
	rings := equip.GetRings()
	ringLabel := labelStyle.Render(fmt.Sprintf("%-10s", "Rings"))
	if len(rings) == 0 {
		ringValue := valueStyle.Render("-")
		if focused && m.equipCursor == 8 {
			ringValue = selectedStyle.Render("▶ -")
		}
		lines = append(lines, fmt.Sprintf("%s %s", ringLabel, ringValue))
	} else {
		for i, ring := range rings {
			name := ring.Name
			if ring.Magical {
				name = "✨ " + name
			}
			if i == 0 {
				lines = append(lines, fmt.Sprintf("%s %s", ringLabel, valueStyle.Render(name)))
			} else {
				lines = append(lines, fmt.Sprintf("%s %s", strings.Repeat(" ", 10), valueStyle.Render(name)))
			}
		}
	}

	// Attunement count
	lines = append(lines, "")
	attuned := equip.CountAttunedItems()
	attunementStyle := valueStyle
	if attuned >= 3 {
		attunementStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	}
	lines = append(lines, fmt.Sprintf("%s %s",
		labelStyle.Render("Attuned:"),
		attunementStyle.Render(fmt.Sprintf("%d/3", attuned)),
	))

	return panelStyle.Render(strings.Join(lines, "\n"))
}

func (m *InventoryModel) renderItems(width int) string {
	focused := m.focus == FocusItems

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
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141"))

	var lines []string
	lines = append(lines, titleStyle.Render("🎒 Items"))
	lines = append(lines, "")

	items := m.character.Inventory.Items
	if len(items) == 0 {
		lines = append(lines, dimStyle.Render("  No items in inventory"))
	} else {
		// Show items with scrolling
		endIdx := m.itemScroll + m.itemsPerPage
		if endIdx > len(items) {
			endIdx = len(items)
		}

		for i := m.itemScroll; i < endIdx; i++ {
			item := items[i]
			prefix := "  "
			style := itemStyle

			if focused && i == m.itemCursor {
				prefix = "▶ "
				style = selectedStyle
			}

			// Format: name (qty) - type
			name := item.Name
			if item.Magical {
				name = "✨ " + name
			}

			qty := ""
			if item.Quantity > 1 {
				qty = fmt.Sprintf(" ×%d", item.Quantity)
			}

			typeTag := typeStyle.Render(fmt.Sprintf("[%s]", item.Type))

			line := fmt.Sprintf("%s%s%s %s", prefix, style.Render(name), dimStyle.Render(qty), typeTag)
			lines = append(lines, line)
		}

		// Scroll indicator
		if len(items) > m.itemsPerPage {
			indicator := fmt.Sprintf("  [%d-%d of %d]", m.itemScroll+1, endIdx, len(items))
			lines = append(lines, dimStyle.Render(indicator))
		}
	}

	// Selected item details
	if m.selectedItem != nil && focused {
		lines = append(lines, "")
		lines = append(lines, dimStyle.Render("─── Details ───"))
		lines = append(lines, itemStyle.Render(m.selectedItem.Name))
		if m.selectedItem.Description != "" {
			desc := m.selectedItem.Description
			if len(desc) > width-6 {
				desc = desc[:width-9] + "..."
			}
			lines = append(lines, dimStyle.Render(desc))
		}
		if m.selectedItem.Weight > 0 {
			lines = append(lines, dimStyle.Render(fmt.Sprintf("Weight: %.1f lbs", m.selectedItem.Weight)))
		}
		if m.selectedItem.Damage != "" {
			lines = append(lines, dimStyle.Render(fmt.Sprintf("Damage: %s %s", m.selectedItem.Damage, m.selectedItem.DamageType)))
		}
		if m.selectedItem.Charges > 0 || m.selectedItem.MaxCharges > 0 {
			lines = append(lines, dimStyle.Render(fmt.Sprintf("Charges: %d/%d", m.selectedItem.Charges, m.selectedItem.MaxCharges)))
		}
	}

	return panelStyle.Render(strings.Join(lines, "\n"))
}

func (m *InventoryModel) renderCurrency(width int) string {
	focused := m.focus == FocusCurrency

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
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220")) // Gold color
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)

	var lines []string
	lines = append(lines, titleStyle.Render("💰 Currency"))
	lines = append(lines, "")

	currency := m.character.Inventory.Currency
	currencies := []struct {
		name   string
		symbol string
		amount int
		color  string
	}{
		{"Copper", "CP", currency.Copper, "166"},
		{"Silver", "SP", currency.Silver, "252"},
		{"Electrum", "EP", currency.Electrum, "81"},
		{"Gold", "GP", currency.Gold, "220"},
		{"Platinum", "PP", currency.Platinum, "255"},
	}

	for i, c := range currencies {
		label := labelStyle.Render(fmt.Sprintf("%-8s", c.name))
		coinStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(c.color))
		value := coinStyle.Render(fmt.Sprintf("%d %s", c.amount, c.symbol))

		if focused && m.currencyType == i {
			value = selectedStyle.Render(fmt.Sprintf("▶ %d %s", c.amount, c.symbol))
		}

		lines = append(lines, fmt.Sprintf("%s %s", label, value))
	}

	// Total value
	lines = append(lines, "")
	gp, cp := currency.TotalInGold()
	lines = append(lines, labelStyle.Render("─── Total ───"))
	lines = append(lines, valueStyle.Render(fmt.Sprintf("%d GP %d CP", gp, cp)))

	// Currency input mode
	if m.currencyMode {
		lines = append(lines, "")
		action := "Add"
		if !m.currencyAdding {
			action = "Spend"
		}
		lines = append(lines, titleStyle.Render(fmt.Sprintf("%s %s:", action, currencyName(m.currencyType))))
		lines = append(lines, valueStyle.Render(m.currencyBuffer+"_"))
	}

	return panelStyle.Render(strings.Join(lines, "\n"))
}

func (m *InventoryModel) renderFooter(width int) string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Width(width)

	var help string
	switch m.focus {
	case FocusEquipment:
		help = "↑/↓: navigate • tab: next panel • esc: back to sheet"
	case FocusItems:
		help = "↑/↓: navigate • enter: select • +/a: add qty • -/d: remove • e: equip • tab: next panel • esc: back"
	case FocusCurrency:
		help = "↑/↓: select currency • +/a: add • -/d: spend • enter: edit • tab: next panel • esc: back"
	}

	if m.statusMessage != "" {
		help = m.statusMessage + " | " + help
	}

	return footerStyle.Render(help)
}

// BackToSheetMsg signals to return to the main sheet.
type BackToSheetMsg struct{}

// OpenInventoryMsg signals to open the inventory view.
type OpenInventoryMsg struct{}
