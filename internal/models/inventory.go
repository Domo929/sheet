package models

import (
	"errors"

	"github.com/Domo929/sheet/internal/domain"
)

// Error sentinels for inventory operations
var (
	ErrInsufficientFunds            = errors.New("insufficient funds")
	ErrItemDoesNotRequireAttunement = errors.New("item does not require attunement")
	ErrItemAlreadyAttuned           = errors.New("item is already attuned")
	ErrMaxAttunementReached         = errors.New("already attuned to maximum of 3 items")
	ErrItemNotAttuned               = errors.New("item is not attuned")
)

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

// TotalInGold converts all currency to gold pieces (integer).
// Returns whole gold pieces and remaining copper pieces.
// Conversion: 100 CP = 1 GP, 10 SP = 1 GP, 2 EP = 1 GP, 10 PP = 100 GP
func (c *Currency) TotalInGold() (goldPieces int, remainingCopper int) {
	totalCopper := c.Copper +
		c.Silver*10 +
		c.Electrum*50 +
		c.Gold*100 +
		c.Platinum*1000

	goldPieces = totalCopper / 100
	remainingCopper = totalCopper % 100
	return goldPieces, remainingCopper
}

// TotalCopper returns the total value of the pouch in copper pieces.
func (c *Currency) TotalCopper() int {
	return c.Copper + c.Silver*10 + c.Electrum*50 + c.Gold*100 + c.Platinum*1000
}

// distributeFromCopper resets the pouch to an optimal coin layout (platinum,
// gold, silver, copper; electrum unused) for the given total copper value.
func (c *Currency) distributeFromCopper(totalCopper int) {
	c.Platinum = totalCopper / 1000
	totalCopper %= 1000
	c.Gold = totalCopper / 100
	totalCopper %= 100
	c.Electrum = 0
	c.Silver = totalCopper / 10
	totalCopper %= 10
	c.Copper = totalCopper
}

// CurrencyFromCopper builds an optimally distributed Currency from a copper total.
func CurrencyFromCopper(totalCopper int) Currency {
	var c Currency
	c.distributeFromCopper(totalCopper)
	return c
}

// SpendCoins deducts the given cost, making change from larger coins only when
// a denomination is short. Denominations the player already holds are otherwise
// preserved (e.g. paying 15 gp from 50 gp leaves 35 gp). Returns
// ErrInsufficientFunds when the pouch cannot cover the cost.
func (c *Currency) SpendCoins(cost Currency) error {
	if c.TotalCopper() < cost.TotalCopper() {
		return ErrInsufficientFunds
	}
	remaining := cost.TotalCopper()
	denoms := []struct {
		ptr *int
		val int
	}{
		{&c.Copper, 1}, {&c.Silver, 10}, {&c.Electrum, 50}, {&c.Gold, 100}, {&c.Platinum, 1000},
	}
	for remaining > 0 {
		// Pay with exact coins, smallest first, without overpaying.
		for i := 0; i < len(denoms) && remaining > 0; i++ {
			n := remaining / denoms[i].val
			if n > *denoms[i].ptr {
				n = *denoms[i].ptr
			}
			*denoms[i].ptr -= n
			remaining -= n * denoms[i].val
		}
		if remaining == 0 {
			break
		}
		// Break the smallest coin larger than the remainder into change.
		broke := false
		for i := 1; i < len(denoms); i++ {
			if denoms[i].val > remaining && *denoms[i].ptr > 0 {
				*denoms[i].ptr--
				lower := i - 1
				if denoms[lower].val == 50 { // skip electrum when making change
					lower = 1 // silver
				}
				*denoms[lower].ptr += denoms[i].val / denoms[lower].val
				broke = true
				break
			}
		}
		if !broke {
			return ErrInsufficientFunds
		}
	}
	return nil
}

// AddValue adds the given value to the pouch, preserving denominations.
func (c *Currency) AddValue(v Currency) {
	c.Copper += v.Copper
	c.Silver += v.Silver
	c.Electrum += v.Electrum
	c.Gold += v.Gold
	c.Platinum += v.Platinum
}

// Add adds currency to the pouch.
func (c *Currency) Add(cp, sp, ep, gp, pp int) {
	c.Copper += cp
	c.Silver += sp
	c.Electrum += ep
	c.Gold += gp
	c.Platinum += pp
}

// Spend removes currency from the pouch. Returns error if insufficient funds.
func (c *Currency) Spend(cp, sp, ep, gp, pp int) error {
	// Calculate total value in copper pieces
	currentTotal := c.Copper + c.Silver*10 + c.Electrum*50 + c.Gold*100 + c.Platinum*1000
	costTotal := cp + sp*10 + ep*50 + gp*100 + pp*1000

	if currentTotal < costTotal {
		return ErrInsufficientFunds
	}

	// Deduct from largest denominations first
	remaining := costTotal

	// Try to pay from exact denominations first
	if c.Platinum >= pp {
		c.Platinum -= pp
		remaining -= pp * 1000
	}
	if c.Gold >= gp {
		c.Gold -= gp
		remaining -= gp * 100
	}
	if c.Electrum >= ep {
		c.Electrum -= ep
		remaining -= ep * 50
	}
	if c.Silver >= sp {
		c.Silver -= sp
		remaining -= sp * 10
	}
	if c.Copper >= cp {
		c.Copper -= cp
		remaining -= cp
	}

	// If still remaining, convert from larger denominations
	for remaining > 0 {
		if c.Platinum > 0 {
			c.Platinum--
			c.Gold += 10
		} else if c.Gold > 0 {
			c.Gold--
			c.Silver += 10
		} else if c.Electrum > 0 {
			c.Electrum--
			c.Silver += 5
		} else if c.Silver > 0 {
			c.Silver--
			c.Copper += 10
		}

		// Deduct from copper
		if c.Copper >= remaining {
			c.Copper -= remaining
			remaining = 0
		}
	}

	return nil
}

// AddCopper adds copper pieces.
func (c *Currency) AddCopper(amount int) {
	c.Copper += amount
}

// AddSilver adds silver pieces.
func (c *Currency) AddSilver(amount int) {
	c.Silver += amount
}

// AddElectrum adds electrum pieces.
func (c *Currency) AddElectrum(amount int) {
	c.Electrum += amount
}

// AddGold adds gold pieces.
func (c *Currency) AddGold(amount int) {
	c.Gold += amount
}

// AddPlatinum adds platinum pieces.
func (c *Currency) AddPlatinum(amount int) {
	c.Platinum += amount
}

// SpendCopper removes copper pieces. Returns error if insufficient.
func (c *Currency) SpendCopper(amount int) error {
	return c.Spend(amount, 0, 0, 0, 0)
}

// SpendSilver removes silver pieces. Returns error if insufficient.
func (c *Currency) SpendSilver(amount int) error {
	return c.Spend(0, amount, 0, 0, 0)
}

// SpendElectrum removes electrum pieces. Returns error if insufficient.
func (c *Currency) SpendElectrum(amount int) error {
	return c.Spend(0, 0, amount, 0, 0)
}

// SpendGold removes gold pieces. Returns error if insufficient.
func (c *Currency) SpendGold(amount int) error {
	return c.Spend(0, 0, 0, amount, 0)
}

// SpendPlatinum removes platinum pieces. Returns error if insufficient.
func (c *Currency) SpendPlatinum(amount int) error {
	return c.Spend(0, 0, 0, 0, amount)
}

// SpendFromTotal spends an amount in gold pieces from total wealth.
// Automatically converts currency as needed and redistributes optimally.
// For example, spending 20 GP when you have 3 PP and 7 GP will work.
func (c *Currency) SpendFromTotal(goldAmount int) error {
	// Convert cost to copper
	costInCopper := goldAmount * 100

	// Calculate total wealth in copper
	totalCopper := c.Copper +
		c.Silver*10 +
		c.Electrum*50 +
		c.Gold*100 +
		c.Platinum*1000

	if totalCopper < costInCopper {
		return ErrInsufficientFunds
	}

	// Calculate remaining copper after spending
	remainingCopper := totalCopper - costInCopper

	// Redistribute optimally (largest denominations first)
	c.Platinum = remainingCopper / 1000
	remainingCopper %= 1000

	c.Gold = remainingCopper / 100
	remainingCopper %= 100

	c.Electrum = 0 // Skip electrum for cleaner distribution

	c.Silver = remainingCopper / 10
	remainingCopper %= 10

	c.Copper = remainingCopper

	return nil
}

// SpendFromTotalWithChange spends an amount and returns change breakdown.
// Returns the amount spent from each denomination for display purposes.
func (c *Currency) SpendFromTotalWithChange(goldAmount int) (spent Currency, err error) {
	// Store original values
	original := Currency{
		Copper:   c.Copper,
		Silver:   c.Silver,
		Electrum: c.Electrum,
		Gold:     c.Gold,
		Platinum: c.Platinum,
	}

	err = c.SpendFromTotal(goldAmount)
	if err != nil {
		return Currency{}, err
	}

	// Calculate what was spent from each denomination
	spent = Currency{
		Copper:   original.Copper - c.Copper,
		Silver:   original.Silver - c.Silver,
		Electrum: original.Electrum - c.Electrum,
		Gold:     original.Gold - c.Gold,
		Platinum: original.Platinum - c.Platinum,
	}

	return spent, nil
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
	SlotRing     EquipmentSlot = "ring" // Changed to support multiple rings
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
	SubCategory   string        `json:"subCategory,omitempty"` // e.g., "Simple Melee Weapons", "Martial Ranged Weapons"

	// Weapon properties
	Damage          string                  `json:"damage,omitempty"`           // e.g., "1d8"
	DamageType      domain.DamageType       `json:"damageType,omitempty"`       // e.g., "slashing"
	WeaponProps     []domain.WeaponProperty `json:"weaponProperties,omitempty"` // e.g., ["finesse", "light"]
	Mastery         domain.WeaponMastery    `json:"mastery,omitempty"`          // 2024 weapon mastery property, e.g., "topple"
	VersatileDamage string   `json:"versatileDamage,omitempty"`  // e.g., "1d10" for versatile weapons
	RangeNormal     int      `json:"rangeNormal,omitempty"`      // Normal range in feet
	RangeLong       int      `json:"rangeLong,omitempty"`        // Long range in feet

	// Armor properties
	ArmorClass          int  `json:"armorClass,omitempty"`
	StealthDisadvantage bool `json:"stealthDisadvantage,omitempty"`

	// Magic item properties
	Magical            bool `json:"magical,omitempty"`
	RequiresAttunement bool `json:"requiresAttunement,omitempty"`
	Attuned            bool `json:"attuned,omitempty"`
	MagicBonus         int  `json:"magicBonus,omitempty"` // +1, +2, +3

	// Consumable properties
	Charges    int `json:"charges,omitempty"`
	MaxCharges int `json:"maxCharges,omitempty"`

	// Container properties (for backpacks, pouches, etc.)
	Contents []Item `json:"contents,omitempty"`

	// Custom marks a user-created (homebrew) item.
	Custom bool `json:"custom,omitempty"`
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
	MainHand *Item   `json:"mainHand,omitempty"`
	OffHand  *Item   `json:"offHand,omitempty"`
	Head     *Item   `json:"head,omitempty"`
	Body     *Item   `json:"body,omitempty"`
	Cloak    *Item   `json:"cloak,omitempty"`
	Gloves   *Item   `json:"gloves,omitempty"`
	Boots    *Item   `json:"boots,omitempty"`
	Amulet   *Item   `json:"amulet,omitempty"`
	Rings    []*Item `json:"rings,omitempty"` // Unlimited rings per 2024 rules
}

// NewEquipment creates empty equipment slots.
func NewEquipment() Equipment {
	return Equipment{}
}

// GetSlot returns the item in the given slot.
// Note: For rings, use GetRings() instead as there can be multiple.
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
	default:
		return nil
	}
}

// SetSlot sets the item in the given slot. Returns the previously equipped item.
// Note: For rings, use EquipRing() instead as there can be multiple.
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
	}
	return previous
}

// EquipRing adds a ring to the equipped rings.
func (e *Equipment) EquipRing(ring *Item) {
	e.Rings = append(e.Rings, ring)
}

// UnequipRing removes a ring by index. Returns the removed ring.
func (e *Equipment) UnequipRing(index int) *Item {
	if index < 0 || index >= len(e.Rings) {
		return nil
	}
	ring := e.Rings[index]
	e.Rings = append(e.Rings[:index], e.Rings[index+1:]...)
	return ring
}

// GetRings returns all equipped rings.
func (e *Equipment) GetRings() []*Item {
	return e.Rings
}

// CountAttunedItems returns the number of attuned magic items.
func (e *Equipment) CountAttunedItems() int {
	count := 0
	slots := []*Item{
		e.MainHand, e.OffHand, e.Head, e.Body, e.Cloak,
		e.Gloves, e.Boots, e.Amulet,
	}
	for _, item := range slots {
		if item != nil && item.Attuned {
			count++
		}
	}
	// Also count rings
	for _, ring := range e.Rings {
		if ring != nil && ring.Attuned {
			count++
		}
	}
	return count
}

// AttuneItem attunes a magic item. Returns error if already at max attunement (3 items).
func (e *Equipment) AttuneItem(item *Item) error {
	if !item.RequiresAttunement {
		return ErrItemDoesNotRequireAttunement
	}
	if item.Attuned {
		return ErrItemAlreadyAttuned
	}
	if e.CountAttunedItems() >= 3 {
		return ErrMaxAttunementReached
	}
	item.Attuned = true
	return nil
}

// UnattuneItem removes attunement from an item.
func (e *Equipment) UnattuneItem(item *Item) error {
	if !item.Attuned {
		return ErrItemNotAttuned
	}
	item.Attuned = false
	return nil
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

// CountAttunedItems returns the number of attuned items in the inventory.
func (inv *Inventory) CountAttunedItems() int {
	count := 0
	for i := range inv.Items {
		if inv.Items[i].Attuned {
			count++
		}
	}
	return count
}

// ToggleAttunement flips attunement on the item with the given ID. Attuning
// requires the item to allow attunement and respects the maximum of 3.
func (inv *Inventory) ToggleAttunement(id string) error {
	item := inv.FindItem(id)
	if item == nil {
		return ErrItemNotAttuned
	}
	if item.Attuned {
		item.Attuned = false
		return nil
	}
	if !item.RequiresAttunement {
		return ErrItemDoesNotRequireAttunement
	}
	if inv.CountAttunedItems() >= 3 {
		return ErrMaxAttunementReached
	}
	item.Attuned = true
	return nil
}
