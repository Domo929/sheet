package models

// Currency represents D&D 5e currency.
type Currency struct {
	Copper   int `json:"copper"`
	Silver   int `json:"silver"`
	Electrum int `json:"electrum"`
	Gold     int `json:"gold"`
	Platinum int `json:"platinum"`
}

// NewCurrency creates an empty currency pouch.
func NewCurrency() Currency {
	return Currency{}
}

// TotalInGold converts all currency to gold pieces for comparison.
// CP=0.01, SP=0.1, EP=0.5, GP=1, PP=10
func (c *Currency) TotalInGold() float64 {
	return float64(c.Copper)/100 +
		float64(c.Silver)/10 +
		float64(c.Electrum)/2 +
		float64(c.Gold) +
		float64(c.Platinum)*10
}

// Add adds currency to the pouch.
func (c *Currency) Add(cp, sp, ep, gp, pp int) {
	c.Copper += cp
	c.Silver += sp
	c.Electrum += ep
	c.Gold += gp
	c.Platinum += pp
}

// ItemType categorizes items.
type ItemType string

const (
	ItemTypeWeapon     ItemType = "weapon"
	ItemTypeArmor      ItemType = "armor"
	ItemTypeShield     ItemType = "shield"
	ItemTypeConsumable ItemType = "consumable"
	ItemTypeMagicItem  ItemType = "magicItem"
	ItemTypeTool       ItemType = "tool"
	ItemTypeGeneral    ItemType = "general"
)

// EquipmentSlot identifies where an item can be equipped.
type EquipmentSlot string

const (
	SlotMainHand EquipmentSlot = "mainHand"
	SlotOffHand  EquipmentSlot = "offHand"
	SlotHead     EquipmentSlot = "head"
	SlotBody     EquipmentSlot = "body"
	SlotCloak    EquipmentSlot = "cloak"
	SlotGloves   EquipmentSlot = "gloves"
	SlotBoots    EquipmentSlot = "boots"
	SlotAmulet   EquipmentSlot = "amulet"
	SlotRing1    EquipmentSlot = "ring1"
	SlotRing2    EquipmentSlot = "ring2"
)

// Item represents an inventory item.
type Item struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Type          ItemType      `json:"type"`
	Description   string        `json:"description,omitempty"`
	Quantity      int           `json:"quantity"`
	Weight        float64       `json:"weight,omitempty"`
	Value         Currency      `json:"value,omitempty"`
	EquipmentSlot EquipmentSlot `json:"equipmentSlot,omitempty"`

	// Weapon properties
	Damage         string   `json:"damage,omitempty"`         // e.g., "1d8"
	DamageType     string   `json:"damageType,omitempty"`     // e.g., "slashing"
	WeaponProps    []string `json:"weaponProperties,omitempty"` // e.g., ["finesse", "light"]

	// Armor properties
	ArmorClass int  `json:"armorClass,omitempty"`
	StealthDisadvantage bool `json:"stealthDisadvantage,omitempty"`

	// Magic item properties
	Magical       bool   `json:"magical,omitempty"`
	RequiresAttunement bool `json:"requiresAttunement,omitempty"`
	Attuned       bool   `json:"attuned,omitempty"`
	MagicBonus    int    `json:"magicBonus,omitempty"` // +1, +2, +3

	// Consumable properties
	Charges    int `json:"charges,omitempty"`
	MaxCharges int `json:"maxCharges,omitempty"`
}

// NewItem creates a basic item with the given name and type.
func NewItem(id, name string, itemType ItemType) Item {
	return Item{
		ID:       id,
		Name:     name,
		Type:     itemType,
		Quantity: 1,
	}
}

// UseCharge decrements charges if available. Returns true if successful.
func (i *Item) UseCharge() bool {
	if i.Charges > 0 {
		i.Charges--
		return true
	}
	return false
}

// Recharge restores charges up to maximum.
func (i *Item) Recharge(amount int) {
	i.Charges += amount
	if i.Charges > i.MaxCharges {
		i.Charges = i.MaxCharges
	}
}

// Equipment tracks what items are currently equipped.
type Equipment struct {
	MainHand *Item `json:"mainHand,omitempty"`
	OffHand  *Item `json:"offHand,omitempty"`
	Head     *Item `json:"head,omitempty"`
	Body     *Item `json:"body,omitempty"`
	Cloak    *Item `json:"cloak,omitempty"`
	Gloves   *Item `json:"gloves,omitempty"`
	Boots    *Item `json:"boots,omitempty"`
	Amulet   *Item `json:"amulet,omitempty"`
	Ring1    *Item `json:"ring1,omitempty"`
	Ring2    *Item `json:"ring2,omitempty"`
}

// NewEquipment creates empty equipment slots.
func NewEquipment() Equipment {
	return Equipment{}
}

// GetSlot returns the item in the given slot.
func (e *Equipment) GetSlot(slot EquipmentSlot) *Item {
	switch slot {
	case SlotMainHand:
		return e.MainHand
	case SlotOffHand:
		return e.OffHand
	case SlotHead:
		return e.Head
	case SlotBody:
		return e.Body
	case SlotCloak:
		return e.Cloak
	case SlotGloves:
		return e.Gloves
	case SlotBoots:
		return e.Boots
	case SlotAmulet:
		return e.Amulet
	case SlotRing1:
		return e.Ring1
	case SlotRing2:
		return e.Ring2
	default:
		return nil
	}
}

// SetSlot sets the item in the given slot. Returns the previously equipped item.
func (e *Equipment) SetSlot(slot EquipmentSlot, item *Item) *Item {
	var previous *Item
	switch slot {
	case SlotMainHand:
		previous = e.MainHand
		e.MainHand = item
	case SlotOffHand:
		previous = e.OffHand
		e.OffHand = item
	case SlotHead:
		previous = e.Head
		e.Head = item
	case SlotBody:
		previous = e.Body
		e.Body = item
	case SlotCloak:
		previous = e.Cloak
		e.Cloak = item
	case SlotGloves:
		previous = e.Gloves
		e.Gloves = item
	case SlotBoots:
		previous = e.Boots
		e.Boots = item
	case SlotAmulet:
		previous = e.Amulet
		e.Amulet = item
	case SlotRing1:
		previous = e.Ring1
		e.Ring1 = item
	case SlotRing2:
		previous = e.Ring2
		e.Ring2 = item
	}
	return previous
}

// CountAttunedItems returns the number of attuned magic items.
func (e *Equipment) CountAttunedItems() int {
	count := 0
	slots := []*Item{
		e.MainHand, e.OffHand, e.Head, e.Body, e.Cloak,
		e.Gloves, e.Boots, e.Amulet, e.Ring1, e.Ring2,
	}
	for _, item := range slots {
		if item != nil && item.Attuned {
			count++
		}
	}
	return count
}

// Inventory contains all character items and equipment.
type Inventory struct {
	Items     []Item    `json:"items"`
	Equipment Equipment `json:"equipment"`
	Currency  Currency  `json:"currency"`
}

// NewInventory creates an empty inventory.
func NewInventory() Inventory {
	return Inventory{
		Items:     []Item{},
		Equipment: NewEquipment(),
		Currency:  NewCurrency(),
	}
}

// AddItem adds an item to the inventory.
func (inv *Inventory) AddItem(item Item) {
	inv.Items = append(inv.Items, item)
}

// RemoveItem removes an item by ID. Returns the removed item or nil.
func (inv *Inventory) RemoveItem(id string) *Item {
	for i, item := range inv.Items {
		if item.ID == id {
			removed := inv.Items[i]
			inv.Items = append(inv.Items[:i], inv.Items[i+1:]...)
			return &removed
		}
	}
	return nil
}

// FindItem finds an item by ID.
func (inv *Inventory) FindItem(id string) *Item {
	for i := range inv.Items {
		if inv.Items[i].ID == id {
			return &inv.Items[i]
		}
	}
	return nil
}

// TotalWeight calculates the total weight of all items.
func (inv *Inventory) TotalWeight() float64 {
	total := 0.0
	for _, item := range inv.Items {
		total += item.Weight * float64(item.Quantity)
	}
	return total
}
