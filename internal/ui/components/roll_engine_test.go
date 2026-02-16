package components

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// advanceToShowing sends enough RollTickMsgs to bring the engine to the Showing state.
func advanceToShowing(t *testing.T, e *RollEngine) {
	t.Helper()
	for i := 0; i < totalAnimFrames+5; i++ {
		cmd := e.Update(RollTickMsg{Time: time.Now()})
		if e.state == rollStateShowing {
			return
		}
		if cmd == nil {
			break
		}
	}
	assert.Equal(t, rollStateShowing, e.state, "Expected engine to reach Showing state")
}

func TestRollEngine_RequestRollWithAdvPrompt(t *testing.T) {
	e := NewRollEngine()
	e.Update(RequestRollMsg{DiceExpr: "1d20", AdvPrompt: true})

	assert.Equal(t, rollStateAdvPrompt, e.State())
	assert.True(t, e.IsActive())
}

func TestRollEngine_RequestRollWithoutAdvPrompt(t *testing.T) {
	e := NewRollEngine()
	cmd := e.Update(RequestRollMsg{DiceExpr: "1d6", Modifier: 0, AdvPrompt: false})

	assert.Equal(t, rollStateAnimating, e.State())
	assert.NotNil(t, cmd, "Should return a tick cmd")
}

func TestRollEngine_AdvPromptNormal(t *testing.T) {
	e := NewRollEngine()
	e.Update(RequestRollMsg{DiceExpr: "1d20", AdvPrompt: true})

	e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	assert.Equal(t, rollStateAnimating, e.state)
}

func TestRollEngine_AdvPromptAdvantage(t *testing.T) {
	e := NewRollEngine()
	e.Update(RequestRollMsg{DiceExpr: "1d20", AdvPrompt: true})

	e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	assert.Equal(t, rollStateAnimating, e.state)
	assert.True(t, e.advantage)
}

func TestRollEngine_AdvPromptDisadvantage(t *testing.T) {
	e := NewRollEngine()
	e.Update(RequestRollMsg{DiceExpr: "1d20", AdvPrompt: true})

	e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	assert.Equal(t, rollStateAnimating, e.state)
	assert.True(t, e.disadvantage)
}

func TestRollEngine_AdvPromptEsc(t *testing.T) {
	e := NewRollEngine()
	e.Update(RequestRollMsg{DiceExpr: "1d20", AdvPrompt: true})

	e.Update(tea.KeyMsg{Type: tea.KeyEsc})

	assert.Equal(t, rollStateIdle, e.state)
	assert.False(t, e.IsActive())
}

func TestRollEngine_AnimationCompletes(t *testing.T) {
	e := NewRollEngine()
	e.Update(RequestRollMsg{DiceExpr: "1d6", AdvPrompt: false})

	advanceToShowing(t, e)

	assert.Equal(t, rollStateShowing, e.state)
}

func TestRollEngine_ShowingDismiss(t *testing.T) {
	e := NewRollEngine()
	e.Update(RequestRollMsg{DiceExpr: "1d6", AdvPrompt: false})
	advanceToShowing(t, e)

	// No follow-up set, any key dismisses
	cmd := e.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.Equal(t, rollStateIdle, e.state)
	assert.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(RollCompleteMsg)
	assert.True(t, ok, "Expected RollCompleteMsg")
}

func TestRollEngine_ShowingWithFollowUp(t *testing.T) {
	followUp := &RequestRollMsg{Label: "Damage", DiceExpr: "1d8"}
	e := NewRollEngine()
	e.Update(RequestRollMsg{DiceExpr: "1d20", AdvPrompt: false, FollowUp: followUp})
	advanceToShowing(t, e)

	// Press Enter to trigger follow-up
	cmd := e.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, cmd)

	msg := cmd()
	reqMsg, ok := msg.(RequestRollMsg)
	assert.True(t, ok, "Expected RequestRollMsg (follow-up)")
	assert.Equal(t, "Damage", reqMsg.Label)
	assert.Equal(t, "1d8", reqMsg.DiceExpr)

	// Process the follow-up request — engine was reset to idle, so this starts a new roll
	cmd2 := e.Update(reqMsg)
	assert.Equal(t, rollStateAnimating, e.state, "Follow-up should start animating")
	assert.NotNil(t, cmd2)
}

func TestRollEngine_ShowingSkipFollowUp(t *testing.T) {
	followUp := &RequestRollMsg{Label: "Damage", DiceExpr: "1d8"}
	e := NewRollEngine()
	e.Update(RequestRollMsg{DiceExpr: "1d20", AdvPrompt: false, FollowUp: followUp})
	advanceToShowing(t, e)

	// Press Esc to skip follow-up
	cmd := e.Update(tea.KeyMsg{Type: tea.KeyEsc})

	assert.Equal(t, rollStateIdle, e.state)
	assert.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(RollCompleteMsg)
	assert.True(t, ok, "Expected RollCompleteMsg when skipping follow-up")
}

func TestRollEngine_CustomRollNavigation(t *testing.T) {
	e := NewRollEngine()
	e.OpenCustomRoll()

	assert.Equal(t, rollStateCustomRoll, e.state)
	assert.Equal(t, 0, e.selectedDie)
	assert.Equal(t, 1, e.quantity)

	// Right → selectedDie == 1
	e.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, 1, e.selectedDie)

	// Left → selectedDie == 0
	e.Update(tea.KeyMsg{Type: tea.KeyLeft})
	assert.Equal(t, 0, e.selectedDie)

	// Up → quantity == 2
	e.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 2, e.quantity)

	// Down → quantity == 1
	e.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, e.quantity)
}

func TestRollEngine_CustomRollEnter(t *testing.T) {
	e := NewRollEngine()
	e.OpenCustomRoll()

	// Select d8 (index 2): Right, Right
	e.Update(tea.KeyMsg{Type: tea.KeyRight})
	e.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, 2, e.selectedDie)

	// Set quantity to 3: Up, Up
	e.Update(tea.KeyMsg{Type: tea.KeyUp})
	e.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 3, e.quantity)

	// Press Enter to roll
	cmd := e.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, rollStateAnimating, e.state, "Custom roll should start animating")
	assert.NotNil(t, cmd)
}

func TestRollEngine_CustomRollEsc(t *testing.T) {
	e := NewRollEngine()
	e.OpenCustomRoll()

	e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Equal(t, rollStateIdle, e.state)
}

func TestRollEngine_CustomRollQuantityBounds(t *testing.T) {
	e := NewRollEngine()
	e.OpenCustomRoll()

	// Down from 1 should stay at 1 (minimum)
	e.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, e.quantity, "Quantity should not go below 1")

	// Set quantity to max by pressing up many times
	for i := 0; i < 110; i++ {
		e.Update(tea.KeyMsg{Type: tea.KeyUp})
	}
	assert.Equal(t, customRollMaxQty, e.quantity, "Quantity should not exceed max (100)")
}

func TestRollEngine_IsActive(t *testing.T) {
	e := NewRollEngine()
	assert.False(t, e.IsActive(), "New engine should be idle")

	e.OpenCustomRoll()
	assert.True(t, e.IsActive(), "Engine should be active in custom roll state")

	// Reset to idle via Esc
	e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, e.IsActive(), "Engine should be idle after Esc")
}

func TestRollEngine_ViewRendersModal(t *testing.T) {
	e := NewRollEngine()
	e.Update(RequestRollMsg{DiceExpr: "1d6", AdvPrompt: false, Label: "Test Roll"})

	view := e.View(80, 24)
	assert.NotEmpty(t, view, "View should not be empty during animation")
	assert.Contains(t, view, "Test Roll")
}
