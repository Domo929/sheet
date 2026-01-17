package data

// Equipment represents the full equipment database.
type Equipment struct {
	Weapons Weapons `json:"weapons"`
	Armor   Armor   `json:"armor"`
	Packs   []Pack  `json:"packs"`
	Gear    []Item  `json:"gear"`
}

// Weapons contains all weapon categories.
type Weapons struct {
	SimpleMelee    []Weapon `json:"simple_melee"`
	SimpleRanged   []Weapon `json:"simple_ranged"`
	MartialMelee   []Weapon `json:"martial_melee"`
	MartialRanged  []Weapon `json:"martial_ranged"`
}

// Armor contains all armor categories.
type Armor struct {
	Light  []ArmorItem `json:"light"`
	Medium []ArmorItem `json:"medium"`
	Heavy  []ArmorItem `json:"heavy"`
	Shield []ArmorItem `json:"shield"`
}

// Weapon represents a weapon item.
type Weapon struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Cost             map[string]int    `json:"cost"`
	Damage           string            `json:"damage"`
	DamageType       string            `json:"damageType"`
	Weight           float64           `json:"weight"`
	Properties       []string          `json:"properties"`
	Range            *WeaponRange      `json:"range,omitempty"`
	VersatileDamage  string            `json:"versatileDamage,omitempty"`
}

// WeaponRange represents weapon range in feet.
type WeaponRange struct {
	Normal int `json:"normal"`
	Long   int `json:"long"`
}

// ArmorItem represents an armor item.
type ArmorItem struct {
	ID                  string         `json:"id"`
	Name                string         `json:"name"`
	Cost                map[string]int `json:"cost"`
	ArmorClass          string         `json:"armorClass,omitempty"`
	Weight              float64        `json:"weight"`
	StealthDisadvantage bool           `json:"stealthDisadvantage,omitempty"`
	StrengthRequired    int            `json:"strengthRequired,omitempty"`
}

// Pack represents an equipment pack.
type Pack struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Cost     map[string]int `json:"cost"`
	Weight   float64        `json:"weight"`
	Contents []string       `json:"contents"`
}

// Item represents a general gear item.
type Item struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Cost   map[string]int `json:"cost"`
	Weight float64        `json:"weight"`
}

// GetAllWeapons returns all weapons in a flat list.
func (w *Weapons) GetAllWeapons() []Weapon {
	var all []Weapon
	all = append(all, w.SimpleMelee...)
	all = append(all, w.SimpleRanged...)
	all = append(all, w.MartialMelee...)
	all = append(all, w.MartialRanged...)
	return all
}

// GetSimpleWeapons returns all simple weapons.
func (w *Weapons) GetSimpleWeapons() []Weapon {
	var simple []Weapon
	simple = append(simple, w.SimpleMelee...)
	simple = append(simple, w.SimpleRanged...)
	return simple
}

// GetMartialWeapons returns all martial weapons.
func (w *Weapons) GetMartialWeapons() []Weapon {
	var martial []Weapon
	martial = append(martial, w.MartialMelee...)
	martial = append(martial, w.MartialRanged...)
	return martial
}

// GetAllArmor returns all armor items in a flat list.
func (a *Armor) GetAllArmor() []ArmorItem {
	var all []ArmorItem
	all = append(all, a.Light...)
	all = append(all, a.Medium...)
	all = append(all, a.Heavy...)
	all = append(all, a.Shield...)
	return all
}
