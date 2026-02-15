package components

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRollHistory_Add(t *testing.T) {
	h := NewRollHistory()
	assert.Empty(t, h.Entries)
	assert.False(t, h.Visible)

	h.Add(RollHistoryEntry{Label: "Test Roll", Total: 15})
	assert.Len(t, h.Entries, 1)
	assert.True(t, h.Visible, "First add should make history visible")
	assert.Equal(t, "Test Roll", h.Entries[0].Label)
}

func TestRollHistory_AddNewestFirst(t *testing.T) {
	h := NewRollHistory()
	h.Add(RollHistoryEntry{Label: "First"})
	h.Add(RollHistoryEntry{Label: "Second"})
	assert.Equal(t, "Second", h.Entries[0].Label)
	assert.Equal(t, "First", h.Entries[1].Label)
}

func TestRollHistory_CapsAt50(t *testing.T) {
	h := NewRollHistory()
	for i := range 60 {
		h.Add(RollHistoryEntry{Label: fmt.Sprintf("Roll %d", i)})
	}
	assert.Len(t, h.Entries, 50)
	assert.Equal(t, "Roll 59", h.Entries[0].Label)
}

func TestRollHistory_Toggle(t *testing.T) {
	h := NewRollHistory()
	h.Add(RollHistoryEntry{Label: "Test"})
	assert.True(t, h.Visible)
	h.Toggle()
	assert.False(t, h.Visible)
	h.Toggle()
	assert.True(t, h.Visible)
}

func TestRollHistory_Clear(t *testing.T) {
	h := NewRollHistory()
	h.Add(RollHistoryEntry{Label: "Test"})
	h.Clear()
	assert.Empty(t, h.Entries)
	assert.False(t, h.Visible)
}

func TestRollTypeIcon(t *testing.T) {
	assert.Equal(t, "⚔", RollTypeIcon(RollAttack))
	assert.Equal(t, "💥", RollTypeIcon(RollDamage))
	assert.Equal(t, "🎯", RollTypeIcon(RollSkillCheck))
	assert.Equal(t, "🛡", RollTypeIcon(RollSavingThrow))
	assert.Equal(t, "🎲", RollTypeIcon(RollLuck))
}

func TestRollHistory_RenderEmpty(t *testing.T) {
	h := NewRollHistory()
	result := h.Render(25, 40)
	assert.Empty(t, result, "Hidden history should render empty")
}

func TestRollHistory_RenderVisible(t *testing.T) {
	h := NewRollHistory()
	h.Add(RollHistoryEntry{Label: "Longsword Attack", Expression: "1d20+7", Total: 19, RollType: RollAttack})
	result := h.Render(25, 40)
	assert.Contains(t, result, "Roll History")
	assert.Contains(t, result, "Longsword Attack")
}

func TestRollHistory_RenderTooNarrow(t *testing.T) {
	h := NewRollHistory()
	h.Add(RollHistoryEntry{Label: "Test"})
	result := h.Render(15, 40)
	assert.Empty(t, result, "Should not render if width < 20")
}
