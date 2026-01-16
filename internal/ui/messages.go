package ui

// NavigateMsg is sent to navigate to a different view.
type NavigateMsg struct {
	View ViewType
}

// LoadCharacterMsg is sent to load a character.
type LoadCharacterMsg struct {
	CharacterName string
}

// ErrorMsg is sent when an error occurs.
type ErrorMsg struct {
	Err error
}

// SaveCharacterMsg is sent to save the current character.
type SaveCharacterMsg struct{}

// TakeDamageMsg is sent when the character takes damage.
type TakeDamageMsg struct {
	Amount int
}

// HealMsg is sent when the character is healed.
type HealMsg struct {
	Amount int
}

// GainXPMsg is sent when the character gains experience points.
type GainXPMsg struct {
	Amount int
}

// LevelUpMsg is sent when the character levels up.
type LevelUpMsg struct{}

// ShortRestMsg is sent when the character takes a short rest.
type ShortRestMsg struct{}

// LongRestMsg is sent when the character takes a long rest.
type LongRestMsg struct{}

// EquipItemMsg is sent when an item is equipped.
type EquipItemMsg struct {
	ItemID string
}

// UnequipItemMsg is sent when an item is unequipped.
type UnequipItemMsg struct {
	ItemID string
}

// UseItemMsg is sent when an item is used.
type UseItemMsg struct {
	ItemID string
}

// CastSpellMsg is sent when a spell is cast.
type CastSpellMsg struct {
	SpellName string
	Level     int
}

// AddConditionMsg is sent when a condition is added to the character.
type AddConditionMsg struct {
	Condition string
}

// RemoveConditionMsg is sent when a condition is removed from the character.
type RemoveConditionMsg struct {
	Condition string
}
