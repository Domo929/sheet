package views

import (
	"fmt"
	"strings"

	"github.com/Domo929/sheet/internal/data"
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

	// Quit confirmation
	confirmingQuit bool

	// Item list navigation
	itemCursor   int
	itemScroll   int
	itemsPerPage int

	// Container viewing (for backpacks, etc.)
	viewingContainer  *models.Item // The container we're viewing into
	containerCursor   int
	containerParentID string // ID of parent item when in container view

	// Equipment slot navigation
	equipCursor int

	// Quantity adjustment mode
	quantityMode   bool
	quantityBuffer string
	quantityItem   *models.Item // Item being adjusted

	// Delete confirmation
	confirmingDelete bool
	deleteItem       *models.Item

	// Currency editing
	currencyMode   bool
	currencyType   int // 0=CP, 1=SP, 2=EP, 3=GP, 4=PP
	currencyBuffer string
	currencyAdding bool

	// Add item mode
	addingItem     bool
	itemSearchTerm string
	searchResults  []models.Item
	searchCursor   int
}

type inventoryKeyMap struct {
	Quit      key.Binding
	ForceQuit key.Binding
	Tab       key.Binding
	ShiftTab  key.Binding
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	Back      key.Binding
	Quantity  key.Binding
	Delete    key.Binding
	Equip     key.Binding
	Add       key.Binding
	Spend     key.Binding
}

func defaultInventoryKeyMap() inventoryKeyMap {
	return inventoryKeyMap{
		Quit:      key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "force quit")),
		Tab:       key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next panel")),
		ShiftTab:  key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev panel")),
		Up:        key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:      key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Enter:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Back:      key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Quantity:  key.NewBinding(key.WithKeys("n", "#"), key.WithHelp("n/#", "adjust quantity")),
		Delete:    key.NewBinding(key.WithKeys("x", "delete"), key.WithHelp("x", "delete item")),
		Equip:     key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "equip/unequip")),
		Add:       key.NewBinding(key.WithKeys("a", "+"), key.WithHelp("a/+", "add item/currency")),
		Spend:     key.NewBinding(key.WithKeys("s", "-"), key.WithHelp("s/-", "spend currency")),
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

		// Handle delete confirmation
		if m.confirmingDelete {
			switch msg.String() {
			case "y", "Y":
				m.performDelete()
				return m, nil
			default:
				m.confirmingDelete = false
				m.deleteItem = nil
				m.statusMessage = "Delete cancelled"
				return m, nil
			}
		}

		// Handle quantity mode
		if m.quantityMode {
			return m.handleQuantityInput(msg)
		}

		// Handle add item mode
		if m.addingItem {
			return m.handleAddItemInput(msg)
		}

		// Ctrl+C always quits immediately
		if key.Matches(msg, m.keys.ForceQuit) {
			return m, tea.Quit
		}

		// Handle currency input mode
		if m.currencyMode {
			return m.handleCurrencyInput(msg)
		}

		m.statusMessage = ""

		switch {
		case key.Matches(msg, m.keys.Quit):
			m.confirmingQuit = true
			m.statusMessage = "Quit? (y/n)"
			return m, nil
		case key.Matches(msg, m.keys.Back):
			// If viewing a container, exit to main inventory first
			if m.viewingContainer != nil {
				m.viewingContainer = nil
				m.containerParentID = ""
				m.containerCursor = 0
				m.itemScroll = 0
				m.statusMessage = "Back to inventory"
				return m, nil
			}
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
		case key.Matches(msg, m.keys.Quantity):
			return m.handleQuantity()
		case key.Matches(msg, m.keys.Delete):
			return m.handleDelete()
		case key.Matches(msg, m.keys.Equip):
			return m.handleEquip()
		case key.Matches(msg, m.keys.Add):
			return m.handleAdd()
		case key.Matches(msg, m.keys.Spend):
			return m.handleSpend()
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
		cursor := m.getCurrentCursor()
		if cursor > 0 {
			m.setCurrentCursor(cursor - 1)
			if m.getCurrentCursor() < m.itemScroll {
				m.itemScroll = m.getCurrentCursor()
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
		items := m.getDisplayItems()
		cursor := m.getCurrentCursor()
		if cursor < len(items)-1 {
			m.setCurrentCursor(cursor + 1)
			if m.getCurrentCursor() >= m.itemScroll+m.itemsPerPage {
				m.itemScroll++
			}
		}
	case FocusCurrency:
		if m.currencyType < 5 { // 0-4 are currencies, 5 is "Total"
			m.currencyType++
		}
	}
	return m, nil
}

func (m *InventoryModel) handleEnter() (*InventoryModel, tea.Cmd) {
	switch m.focus {
	case FocusEquipment:
		// Allow viewing pack contents from equipment panel
		equip := &m.character.Inventory.Equipment
		slots := []models.EquipmentSlot{
			models.SlotMainHand,
			models.SlotOffHand,
			models.SlotHead,
			models.SlotBody,
			models.SlotCloak,
			models.SlotGloves,
			models.SlotBoots,
			models.SlotAmulet,
		}
		if m.equipCursor < len(slots) {
			item := equip.GetSlot(slots[m.equipCursor])
			if item != nil && len(item.Contents) > 0 {
				m.viewingContainer = item
				m.containerParentID = item.ID
				m.containerCursor = 0
				m.focus = FocusItems // Switch to items panel to view contents
				m.statusMessage = fmt.Sprintf("Viewing contents of %s", item.Name)
			}
		}
	case FocusItems:
		// If viewing a container, Enter does nothing special
		if m.viewingContainer != nil {
			return m, nil
		}
		
		// Check if hovered item is a container with contents
		items := m.character.Inventory.Items
		if m.itemCursor < len(items) {
			item := &items[m.itemCursor]
			if len(item.Contents) > 0 {
				// Enter the container
				m.viewingContainer = item
				m.containerParentID = item.ID
				m.containerCursor = 0
				m.statusMessage = fmt.Sprintf("Viewing contents of %s", item.Name)
			}
		}
	case FocusCurrency:
		m.currencyMode = true
		m.currencyBuffer = ""
		m.currencyAdding = true
	}
	return m, nil
}

// getDisplayItems returns either main inventory or container contents
func (m *InventoryModel) getDisplayItems() []models.Item {
	if m.viewingContainer != nil {
		return m.viewingContainer.Contents
	}
	return m.character.Inventory.Items
}

// getCurrentCursor returns the appropriate cursor for current view
func (m *InventoryModel) getCurrentCursor() int {
	if m.viewingContainer != nil {
		return m.containerCursor
	}
	return m.itemCursor
}

// setCurrentCursor sets the appropriate cursor for current view
func (m *InventoryModel) setCurrentCursor(val int) {
	if m.viewingContainer != nil {
		m.containerCursor = val
	} else {
		m.itemCursor = val
	}
}

// handleQuantity opens quantity adjustment mode for the hovered item
func (m *InventoryModel) handleQuantity() (*InventoryModel, tea.Cmd) {
	if m.focus != FocusItems {
		return m, nil
	}

	items := m.getDisplayItems()
	cursor := m.getCurrentCursor()
	if cursor >= len(items) {
		return m, nil
	}

	m.quantityItem = &items[cursor]
	m.quantityMode = true
	m.quantityBuffer = fmt.Sprintf("%d", m.quantityItem.Quantity)
	m.statusMessage = fmt.Sprintf("Adjust quantity for %s", m.quantityItem.Name)
	return m, nil
}

// handleQuantityInput handles input while in quantity adjustment mode
func (m *InventoryModel) handleQuantityInput(msg tea.KeyMsg) (*InventoryModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.quantityMode = false
		m.quantityBuffer = ""
		m.quantityItem = nil
		m.statusMessage = ""
		return m, nil

	case tea.KeyEnter:
		if m.quantityBuffer != "" && m.quantityItem != nil {
			var newQty int
			fmt.Sscanf(m.quantityBuffer, "%d", &newQty)
			if newQty < 1 {
				newQty = 1
			}
			m.quantityItem.Quantity = newQty
			m.saveCharacter()
			m.statusMessage = fmt.Sprintf("Set %s quantity to %d", m.quantityItem.Name, newQty)
		}
		m.quantityMode = false
		m.quantityBuffer = ""
		m.quantityItem = nil
		return m, nil

	case tea.KeyUp:
		// Increment
		if m.quantityItem != nil {
			m.quantityItem.Quantity++
			m.quantityBuffer = fmt.Sprintf("%d", m.quantityItem.Quantity)
			m.saveCharacter()
		}
		return m, nil

	case tea.KeyDown:
		// Decrement (min 1)
		if m.quantityItem != nil && m.quantityItem.Quantity > 1 {
			m.quantityItem.Quantity--
			m.quantityBuffer = fmt.Sprintf("%d", m.quantityItem.Quantity)
			m.saveCharacter()
		}
		return m, nil

	case tea.KeyBackspace:
		if len(m.quantityBuffer) > 0 {
			m.quantityBuffer = m.quantityBuffer[:len(m.quantityBuffer)-1]
		}
		return m, nil

	case tea.KeyRunes:
		for _, r := range msg.Runes {
			if r >= '0' && r <= '9' {
				m.quantityBuffer += string(r)
			}
		}
		return m, nil
	}

	return m, nil
}

// handleAdd adds currency (currency panel) or starts add item mode (items panel)
func (m *InventoryModel) handleAdd() (*InventoryModel, tea.Cmd) {
	switch m.focus {
	case FocusCurrency:
		m.currencyMode = true
		m.currencyBuffer = ""
		m.currencyAdding = true
		m.statusMessage = fmt.Sprintf("Add %s: type amount and press Enter", currencyName(m.currencyType))
	case FocusItems:
		m.addingItem = true
		m.itemSearchTerm = ""
		m.searchResults = nil
		m.searchCursor = 0
		m.statusMessage = "Type to search items, Enter to add, Esc to cancel"
	}
	return m, nil
}

// handleSpend starts spend currency mode
func (m *InventoryModel) handleSpend() (*InventoryModel, tea.Cmd) {
	if m.focus != FocusCurrency {
		return m, nil
	}
	m.currencyMode = true
	m.currencyBuffer = ""
	m.currencyAdding = false
	m.statusMessage = fmt.Sprintf("Spend %s: type amount and press Enter", currencyName(m.currencyType))
	return m, nil
}

// handleAddItemInput handles input in add item mode
func (m *InventoryModel) handleAddItemInput(msg tea.KeyMsg) (*InventoryModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.addingItem = false
		m.itemSearchTerm = ""
		m.searchResults = nil
		m.statusMessage = ""
		return m, nil

	case tea.KeyUp:
		if m.searchCursor > 0 {
			m.searchCursor--
		}
		return m, nil

	case tea.KeyDown:
		if m.searchCursor < len(m.searchResults)-1 {
			m.searchCursor++
		}
		return m, nil

	case tea.KeyEnter:
		if len(m.searchResults) > 0 && m.searchCursor < len(m.searchResults) {
			selectedItem := m.searchResults[m.searchCursor]
			// Create a copy with new ID
			newItem := selectedItem
			newItem.ID = fmt.Sprintf("%s-%d", selectedItem.ID, len(m.character.Inventory.Items)+1)
			newItem.Quantity = 1
			m.character.Inventory.AddItem(newItem)
			m.saveCharacter()
			m.statusMessage = fmt.Sprintf("Added %s to inventory", newItem.Name)
		}
		m.addingItem = false
		m.itemSearchTerm = ""
		m.searchResults = nil
		return m, nil

	case tea.KeyBackspace:
		if len(m.itemSearchTerm) > 0 {
			m.itemSearchTerm = m.itemSearchTerm[:len(m.itemSearchTerm)-1]
			m.updateSearchResults()
		}
		return m, nil

	case tea.KeyRunes:
		m.itemSearchTerm += string(msg.Runes)
		m.updateSearchResults()
		m.searchCursor = 0
		return m, nil
	}

	return m, nil
}

// updateSearchResults searches equipment database for matching items
func (m *InventoryModel) updateSearchResults() {
	if len(m.itemSearchTerm) < 2 {
		m.searchResults = nil
		return
	}

	searchLower := strings.ToLower(m.itemSearchTerm)
	var results []models.Item

	// Get equipment from the data package using loader
	loader := data.NewLoader("data")
	equipment, err := loader.GetEquipment()
	if err != nil {
		m.statusMessage = "Error loading equipment"
		return
	}

	// Search weapons
	for _, w := range equipment.Weapons.GetAllWeapons() {
		if strings.Contains(strings.ToLower(w.Name), searchLower) {
			results = append(results, models.Item{
				ID:       w.ID,
				Name:     w.Name,
				Type:     models.ItemTypeWeapon,
				Weight:   w.Weight,
				Value:    costMapToCurrency(w.Cost),
				Quantity: 1,
			})
			if len(results) >= 10 {
				break
			}
		}
	}

	// Search armor
	if len(results) < 10 {
		for _, a := range equipment.Armor.GetAllArmor() {
			if strings.Contains(strings.ToLower(a.Name), searchLower) {
				results = append(results, models.Item{
					ID:       a.ID,
					Name:     a.Name,
					Type:     models.ItemTypeArmor,
					Weight:   a.Weight,
					Value:    costMapToCurrency(a.Cost),
					Quantity: 1,
				})
				if len(results) >= 10 {
					break
				}
			}
		}
	}

	// Search gear
	if len(results) < 10 {
		for _, g := range equipment.Gear {
			if strings.Contains(strings.ToLower(g.Name), searchLower) {
				results = append(results, models.Item{
					ID:       g.ID,
					Name:     g.Name,
					Type:     models.ItemTypeGeneral,
					Weight:   g.Weight,
					Value:    costMapToCurrency(g.Cost),
					Quantity: 1,
				})
				if len(results) >= 10 {
					break
				}
			}
		}
	}

	// Search tools
	if len(results) < 10 {
		for _, t := range equipment.Tools {
			if strings.Contains(strings.ToLower(t.Name), searchLower) {
				results = append(results, models.Item{
					ID:       t.ID,
					Name:     t.Name,
					Type:     models.ItemTypeTool,
					Weight:   t.Weight,
					Value:    costMapToCurrency(t.Cost),
					Quantity: 1,
				})
				if len(results) >= 10 {
					break
				}
			}
		}
	}

	m.searchResults = results
}

// costMapToCurrency converts a cost map to a Currency struct
func costMapToCurrency(cost map[string]int) models.Currency {
	return models.Currency{
		Copper:   cost["cp"],
		Silver:   cost["sp"],
		Electrum: cost["ep"],
		Gold:     cost["gp"],
		Platinum: cost["pp"],
	}
}

// formatCost formats a Currency struct into a display string
func formatCost(c models.Currency) string {
	if c.Platinum > 0 {
		return fmt.Sprintf("%d PP", c.Platinum)
	}
	if c.Gold > 0 {
		return fmt.Sprintf("%d GP", c.Gold)
	}
	if c.Electrum > 0 {
		return fmt.Sprintf("%d EP", c.Electrum)
	}
	if c.Silver > 0 {
		return fmt.Sprintf("%d SP", c.Silver)
	}
	if c.Copper > 0 {
		return fmt.Sprintf("%d CP", c.Copper)
	}
	return "—"
}

// handleDelete prompts to delete the hovered item
func (m *InventoryModel) handleDelete() (*InventoryModel, tea.Cmd) {
	if m.focus != FocusItems {
		return m, nil
	}

	items := m.getDisplayItems()
	cursor := m.getCurrentCursor()
	if cursor >= len(items) {
		return m, nil
	}

	m.deleteItem = &items[cursor]
	m.confirmingDelete = true
	m.statusMessage = fmt.Sprintf("Delete %s? (y/n)", m.deleteItem.Name)
	return m, nil
}

// performDelete actually removes the item after confirmation
func (m *InventoryModel) performDelete() {
	if m.deleteItem == nil {
		return
	}

	cursor := m.getCurrentCursor()

	if m.viewingContainer != nil {
		// Remove from container
		if cursor < len(m.viewingContainer.Contents) {
			m.viewingContainer.Contents = append(m.viewingContainer.Contents[:cursor], m.viewingContainer.Contents[cursor+1:]...)
			if cursor > 0 && cursor >= len(m.viewingContainer.Contents) {
				m.containerCursor--
			}
		}
	} else {
		// Remove from main inventory
		m.character.Inventory.RemoveItem(m.deleteItem.ID)
		if cursor > 0 && cursor >= len(m.character.Inventory.Items) {
			m.itemCursor--
		}
	}

	m.saveCharacter()
	m.statusMessage = fmt.Sprintf("Deleted %s", m.deleteItem.Name)
	m.confirmingDelete = false
	m.deleteItem = nil
}

func (m *InventoryModel) handleEquip() (*InventoryModel, tea.Cmd) {
	// Handle Equipment panel - unequip items
	if m.focus == FocusEquipment {
		return m.handleUnequipFromSlot()
	}

	if m.focus != FocusItems {
		return m, nil
	}

	// Use hovered item (cursor position) rather than requiring selection
	items := m.character.Inventory.Items
	if m.itemCursor >= len(items) {
		return m, nil
	}
	item := &items[m.itemCursor]
	equip := &m.character.Inventory.Equipment

	// Weapons go to main hand or off hand
	if item.Type == models.ItemTypeWeapon {
		// Check if already in main hand
		if equip.MainHand != nil && equip.MainHand.ID == item.ID {
			equip.MainHand = nil
			m.statusMessage = fmt.Sprintf("Unequipped %s from Main Hand", item.Name)
		} else if equip.OffHand != nil && equip.OffHand.ID == item.ID {
			// Already in off hand, unequip
			equip.OffHand = nil
			m.statusMessage = fmt.Sprintf("Unequipped %s from Off Hand", item.Name)
		} else if equip.MainHand == nil {
			// Equip to main hand
			equip.MainHand = item
			m.statusMessage = fmt.Sprintf("Equipped %s to Main Hand", item.Name)
		} else if equip.OffHand == nil {
			// Main hand full, equip to off hand
			equip.OffHand = item
			m.statusMessage = fmt.Sprintf("Equipped %s to Off Hand", item.Name)
		} else {
			// Both hands full, replace main hand
			equip.MainHand = item
			m.statusMessage = fmt.Sprintf("Equipped %s to Main Hand (replaced)", item.Name)
		}
		m.saveCharacter()
		return m, nil
	}

	// Shields go to off hand
	if item.Type == models.ItemTypeShield {
		if equip.OffHand != nil && equip.OffHand.ID == item.ID {
			equip.OffHand = nil
			m.statusMessage = fmt.Sprintf("Unequipped %s", item.Name)
		} else {
			equip.OffHand = item
			m.statusMessage = fmt.Sprintf("Equipped %s to Off Hand", item.Name)
		}
		m.saveCharacter()
		return m, nil
	}

	// Other equipment uses its slot
	slot := item.EquipmentSlot
	if slot == "" {
		m.statusMessage = "This item cannot be equipped"
		return m, nil
	}

	// Check if already equipped
	equipped := equip.GetSlot(slot)
	if equipped != nil && equipped.ID == item.ID {
		// Unequip
		equip.SetSlot(slot, nil)
		m.statusMessage = fmt.Sprintf("Unequipped %s", item.Name)
	} else {
		// Equip
		equip.SetSlot(slot, item)
		m.statusMessage = fmt.Sprintf("Equipped %s to %s", item.Name, slot)
	}

	m.saveCharacter()
	return m, nil
}

// handleUnequipFromSlot unequips an item from the selected equipment slot
func (m *InventoryModel) handleUnequipFromSlot() (*InventoryModel, tea.Cmd) {
	equip := &m.character.Inventory.Equipment
	slots := []models.EquipmentSlot{
		models.SlotMainHand,
		models.SlotOffHand,
		models.SlotHead,
		models.SlotBody,
		models.SlotCloak,
		models.SlotGloves,
		models.SlotBoots,
		models.SlotAmulet,
	}

	if m.equipCursor < len(slots) {
		slot := slots[m.equipCursor]
		item := equip.GetSlot(slot)
		if item != nil {
			equip.SetSlot(slot, nil)
			m.saveCharacter()
			m.statusMessage = fmt.Sprintf("Unequipped %s", item.Name)
		} else {
			m.statusMessage = "Nothing equipped in this slot"
		}
	} else if m.equipCursor == len(slots) {
		// Rings - unequip first ring if any
		rings := equip.GetRings()
		if len(rings) > 0 {
			equip.UnequipRing(0)
			m.saveCharacter()
			m.statusMessage = fmt.Sprintf("Unequipped %s", rings[0].Name)
		} else {
			m.statusMessage = "No rings equipped"
		}
	}
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
					case 5:
						currency.AddGold(amount) // Total adds as gold
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
					case 5:
						// Spend from total - auto-converts currency
						err = currency.SpendFromTotal(amount)
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
	names := []string{"CP", "SP", "EP", "GP", "PP", "GP (from total)"}
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
	// Use proportional widths with minimums
	const compactBreakpoint = 80

	if width < compactBreakpoint {
		// Compact layout: stack panels vertically
		panelWidth := width - 4

		var sections []string
		sections = append(sections, m.renderEquipment(panelWidth))
		sections = append(sections, m.renderItems(panelWidth))
		sections = append(sections, m.renderCurrency(panelWidth))

		columns := lipgloss.JoinVertical(lipgloss.Left, sections...)

		// Add item overlay
		if m.addingItem {
			overlay := m.renderAddItemOverlay(width)
			columns = lipgloss.JoinVertical(lipgloss.Left, columns, "", overlay)
		}

		footer := m.renderFooter(width)
		return lipgloss.JoinVertical(lipgloss.Left, header, "", columns, "", footer)
	}

	// Standard layout: three columns with proportional widths
	equipWidth := width * 25 / 100 // 25%
	if equipWidth < 24 {
		equipWidth = 24
	}
	currencyWidth := width * 20 / 100 // 20%
	if currencyWidth < 20 {
		currencyWidth = 20
	}
	itemsWidth := width - equipWidth - currencyWidth - 6 // borders and padding
	if itemsWidth < 20 {
		itemsWidth = 20
	}

	equipment := m.renderEquipment(equipWidth)
	items := m.renderItems(itemsWidth)
	currency := m.renderCurrency(currencyWidth)

	columns := lipgloss.JoinHorizontal(lipgloss.Top, equipment, items, currency)

	// Add item overlay
	if m.addingItem {
		overlay := m.renderAddItemOverlay(width)
		columns = lipgloss.JoinVertical(lipgloss.Left, columns, "", overlay)
	}

	// Footer
	footer := m.renderFooter(width)

	return lipgloss.JoinVertical(lipgloss.Left, header, "", columns, "", footer)
}

func (m *InventoryModel) renderAddItemOverlay(width int) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(0, 1).
		Width(width - 4)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	var lines []string
	lines = append(lines, titleStyle.Render("Add Item to Inventory"))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Search: %s_", m.itemSearchTerm))
	lines = append(lines, "")

	if len(m.searchResults) == 0 {
		if len(m.itemSearchTerm) < 2 {
			lines = append(lines, normalStyle.Render("Type at least 2 characters to search..."))
		} else {
			lines = append(lines, normalStyle.Render("No items found"))
		}
	} else {
		for i, item := range m.searchResults {
			costStr := formatCost(item.Value)
			var line string
			if i == m.searchCursor {
				line = selectedStyle.Render(fmt.Sprintf("▶ %s (%s)", item.Name, costStr))
			} else {
				line = normalStyle.Render(fmt.Sprintf("  %s (%s)", item.Name, costStr))
			}
			lines = append(lines, line)
		}
	}

	return boxStyle.Render(strings.Join(lines, "\n"))
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
	
	// Title changes based on whether we're in a container
	if m.viewingContainer != nil {
		lines = append(lines, titleStyle.Render(fmt.Sprintf("📦 %s Contents", m.viewingContainer.Name)))
		lines = append(lines, dimStyle.Render("  (esc to go back)"))
	} else {
		lines = append(lines, titleStyle.Render("🎒 Items"))
	}
	lines = append(lines, "")

	items := m.getDisplayItems()
	cursor := m.getCurrentCursor()
	
	if len(items) == 0 {
		if m.viewingContainer != nil {
			lines = append(lines, dimStyle.Render("  Container is empty"))
		} else {
			lines = append(lines, dimStyle.Render("  No items in inventory"))
		}
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

			if focused && i == cursor {
				prefix = "▶ "
				style = selectedStyle
			}

			// Format: name (qty) - type
			name := item.Name
			if item.Magical {
				name = "✨ " + name
			}
			
			// Show container indicator if item has contents
			if len(item.Contents) > 0 {
				name = "📦 " + name
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

	// Show details for hovered item (no selection required)
	if len(items) > 0 && cursor < len(items) {
		hoveredItem := items[cursor]
		lines = append(lines, "")
		lines = append(lines, dimStyle.Render("─── Details ───"))
		lines = append(lines, itemStyle.Render(hoveredItem.Name))
		if hoveredItem.Description != "" {
			desc := hoveredItem.Description
			if len(desc) > width-6 {
				desc = desc[:width-9] + "..."
			}
			lines = append(lines, dimStyle.Render(desc))
		}
		if hoveredItem.Weight > 0 {
			lines = append(lines, dimStyle.Render(fmt.Sprintf("Weight: %.1f lbs", hoveredItem.Weight)))
		}
		if hoveredItem.Damage != "" {
			lines = append(lines, dimStyle.Render(fmt.Sprintf("Damage: %s %s", hoveredItem.Damage, hoveredItem.DamageType)))
		}
		if hoveredItem.Charges > 0 || hoveredItem.MaxCharges > 0 {
			lines = append(lines, dimStyle.Render(fmt.Sprintf("Charges: %d/%d", hoveredItem.Charges, hoveredItem.MaxCharges)))
		}
		if len(hoveredItem.Contents) > 0 {
			lines = append(lines, dimStyle.Render(fmt.Sprintf("Contains: %d items (Enter to view)", len(hoveredItem.Contents))))
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

	// Total value (selectable for smart spending)
	lines = append(lines, "")
	gp, cp := currency.TotalInGold()
	gpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220")) // Gold color
	cpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("166")) // Copper color

	if focused && m.currencyType == 5 {
		// Total is selected
		totalLabel := selectedStyle.Render("▶ Total")
		totalValue := gpStyle.Render(fmt.Sprintf("%d GP", gp)) + " " + cpStyle.Render(fmt.Sprintf("%d CP", cp))
		lines = append(lines, fmt.Sprintf("%s  %s", totalLabel, totalValue))
		lines = append(lines, labelStyle.Render("  (auto-converts)"))
	} else {
		lines = append(lines, labelStyle.Render("─── Total ───"))
		totalLine := gpStyle.Render(fmt.Sprintf("%d GP", gp)) + " " + cpStyle.Render(fmt.Sprintf("%d CP", cp))
		lines = append(lines, totalLine)
	}

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
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Width(width)

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("42")).
		Bold(true).
		Width(width)

	var help string

	// Show special prompts
	if m.quantityMode {
		help = "↑/↓: adjust • type number • enter: confirm • esc: cancel"
	} else if m.confirmingDelete {
		help = "y: confirm delete • any other key: cancel"
	} else if m.confirmingQuit {
		help = "y: quit • any other key: cancel"
	} else if m.addingItem {
		help = "type to search • ↑/↓: select • enter: add item • esc: cancel"
	} else {
		switch m.focus {
		case FocusEquipment:
			help = "↑/↓: navigate • e: unequip • enter: view pack • tab: next panel • esc: back"
		case FocusItems:
			if m.viewingContainer != nil {
				help = "↑/↓: navigate • a: add item • n: quantity • x: delete • esc: back to inventory"
			} else {
				help = "↑/↓: navigate • a: add item • n: quantity • x: delete • e: equip • tab: next • esc: back"
			}
		case FocusCurrency:
			help = "↑/↓: select • a: add • s: spend • tab: next panel • esc: back"
		}
	}

	var lines []string
	if m.statusMessage != "" {
		lines = append(lines, statusStyle.Render(m.statusMessage))
	}
	lines = append(lines, helpStyle.Render(help))

	return strings.Join(lines, "\n")
}

// BackToSheetMsg signals to return to the main sheet.
type BackToSheetMsg struct{}

// OpenInventoryMsg signals to open the inventory view.
type OpenInventoryMsg struct{}

// OpenSpellbookMsg signals to open the spellbook view.
type OpenSpellbookMsg struct{}
