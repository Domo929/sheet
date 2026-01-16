package models

// SpellSlots tracks spell slots by level.
type SpellSlots struct {
	Level1 SlotTracker `json:"level1"`
	Level2 SlotTracker `json:"level2"`
	Level3 SlotTracker `json:"level3"`
	Level4 SlotTracker `json:"level4"`
	Level5 SlotTracker `json:"level5"`
	Level6 SlotTracker `json:"level6"`
	Level7 SlotTracker `json:"level7"`
	Level8 SlotTracker `json:"level8"`
	Level9 SlotTracker `json:"level9"`
}

// SlotTracker tracks total and remaining spell slots.
type SlotTracker struct {
	Total     int `json:"total"`
	Remaining int `json:"remaining"`
}

// NewSlotTracker creates a slot tracker with the given total.
func NewSlotTracker(total int) SlotTracker {
	return SlotTracker{Total: total, Remaining: total}
}

// Use expends a slot if available. Returns true if successful.
func (st *SlotTracker) Use() bool {
	if st.Remaining > 0 {
		st.Remaining--
		return true
	}
	return false
}

// Restore restores all slots.
func (st *SlotTracker) Restore() {
	st.Remaining = st.Total
}

// NewSpellSlots creates empty spell slots.
func NewSpellSlots() SpellSlots {
	return SpellSlots{}
}

// GetSlot returns the slot tracker for a given spell level (1-9).
func (ss *SpellSlots) GetSlot(level int) *SlotTracker {
	switch level {
	case 1:
		return &ss.Level1
	case 2:
		return &ss.Level2
	case 3:
		return &ss.Level3
	case 4:
		return &ss.Level4
	case 5:
		return &ss.Level5
	case 6:
		return &ss.Level6
	case 7:
		return &ss.Level7
	case 8:
		return &ss.Level8
	case 9:
		return &ss.Level9
	default:
		return nil
	}
}

// SetSlots sets the total slots for a given level.
func (ss *SpellSlots) SetSlots(level, total int) {
	slot := ss.GetSlot(level)
	if slot != nil {
		slot.Total = total
		slot.Remaining = total
	}
}

// UseSlot expends a spell slot at the given level. Returns true if successful.
func (ss *SpellSlots) UseSlot(level int) bool {
	slot := ss.GetSlot(level)
	if slot == nil {
		return false
	}
	return slot.Use()
}

// RestoreAll restores all spell slots (long rest).
func (ss *SpellSlots) RestoreAll() {
	for level := 1; level <= 9; level++ {
		slot := ss.GetSlot(level)
		if slot != nil {
			slot.Restore()
		}
	}
}

// KnownSpell represents a spell the character knows or has prepared.
type KnownSpell struct {
	Name     string `json:"name"`
	Level    int    `json:"level"` // 0 for cantrip
	Prepared bool   `json:"prepared,omitempty"`
	Ritual   bool   `json:"ritual,omitempty"`
}

// PactMagic represents Warlock-style pact magic slots.
type PactMagic struct {
	SlotLevel int `json:"slotLevel"` // All pact slots are this level
	Total     int `json:"total"`
	Remaining int `json:"remaining"`
}

// NewPactMagic creates pact magic slots.
func NewPactMagic(slotLevel, total int) PactMagic {
	return PactMagic{
		SlotLevel: slotLevel,
		Total:     total,
		Remaining: total,
	}
}

// Use expends a pact magic slot. Returns true if successful.
func (pm *PactMagic) Use() bool {
	if pm.Remaining > 0 {
		pm.Remaining--
		return true
	}
	return false
}

// Restore restores all pact magic slots.
func (pm *PactMagic) Restore() {
	pm.Remaining = pm.Total
}

// Spellcasting contains all spellcasting information for a character.
type Spellcasting struct {
	Ability        Ability      `json:"ability"`
	SpellSlots     SpellSlots   `json:"spellSlots"`
	KnownSpells    []KnownSpell `json:"knownSpells,omitempty"`
	CantripsKnown  []string     `json:"cantripsKnown,omitempty"`
	PreparesSpells bool         `json:"preparesSpells"` // Whether class prepares spells
	MaxPrepared    int          `json:"maxPrepared,omitempty"`
	PactMagic      *PactMagic   `json:"pactMagic,omitempty"` // For Warlocks
	RitualCaster   bool         `json:"ritualCaster,omitempty"`
}

// NewSpellcasting creates a new spellcasting tracker.
func NewSpellcasting(ability Ability) Spellcasting {
	return Spellcasting{
		Ability:     ability,
		SpellSlots:  NewSpellSlots(),
		KnownSpells: []KnownSpell{},
	}
}

// AddSpell adds a known spell.
func (sc *Spellcasting) AddSpell(name string, level int) {
	sc.KnownSpells = append(sc.KnownSpells, KnownSpell{
		Name:  name,
		Level: level,
	})
}

// AddCantrip adds a known cantrip.
func (sc *Spellcasting) AddCantrip(name string) {
	sc.CantripsKnown = append(sc.CantripsKnown, name)
}

// PrepareSpell sets a spell as prepared.
func (sc *Spellcasting) PrepareSpell(name string, prepared bool) bool {
	for i := range sc.KnownSpells {
		if sc.KnownSpells[i].Name == name {
			sc.KnownSpells[i].Prepared = prepared
			return true
		}
	}
	return false
}

// CountPreparedSpells returns the number of prepared spells.
func (sc *Spellcasting) CountPreparedSpells() int {
	count := 0
	for _, spell := range sc.KnownSpells {
		if spell.Prepared {
			count++
		}
	}
	return count
}

// GetPreparedSpells returns all prepared spells.
func (sc *Spellcasting) GetPreparedSpells() []KnownSpell {
	prepared := []KnownSpell{}
	for _, spell := range sc.KnownSpells {
		if spell.Prepared {
			prepared = append(prepared, spell)
		}
	}
	return prepared
}

// CalculateSpellSaveDC calculates the spell save DC.
func CalculateSpellSaveDC(abilityMod, proficiencyBonus int) int {
	return 8 + abilityMod + proficiencyBonus
}

// CalculateSpellAttackBonus calculates the spell attack bonus.
func CalculateSpellAttackBonus(abilityMod, proficiencyBonus int) int {
	return abilityMod + proficiencyBonus
}
