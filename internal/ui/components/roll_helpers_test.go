package components

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSpellRollCmd(t *testing.T) {
	t.Run("no damage returns nil", func(t *testing.T) {
		cmd := BuildSpellRollCmd("Mage Hand", "", "", "", 5, 0)
		assert.Nil(t, cmd)
	})

	t.Run("spell attack with damage", func(t *testing.T) {
		cmd := BuildSpellRollCmd("Fire Bolt", "1d10", "fire", "", 7, 0)
		require.NotNil(t, cmd)

		msg := cmd()
		rollMsg, ok := msg.(RequestRollMsg)
		require.True(t, ok)
		assert.Equal(t, "Fire Bolt Attack", rollMsg.Label)
		assert.Equal(t, "1d20", rollMsg.DiceExpr)
		assert.Equal(t, 7, rollMsg.Modifier)
		assert.Equal(t, RollAttack, rollMsg.RollType)
		assert.True(t, rollMsg.AdvPrompt)

		// Check follow-up damage roll
		require.NotNil(t, rollMsg.FollowUp)
		assert.Equal(t, "Fire Bolt Damage (fire)", rollMsg.FollowUp.Label)
		assert.Equal(t, "1d10", rollMsg.FollowUp.DiceExpr)
		assert.Equal(t, RollDamage, rollMsg.FollowUp.RollType)
	})

	t.Run("save-based spell with damage", func(t *testing.T) {
		cmd := BuildSpellRollCmd("Fireball", "8d6", "fire", "DEX", 0, 15)
		require.NotNil(t, cmd)

		msg := cmd()
		rollMsg, ok := msg.(RequestRollMsg)
		require.True(t, ok)
		assert.Contains(t, rollMsg.Label, "Fireball Damage")
		assert.Contains(t, rollMsg.Label, "DC 15")
		assert.Contains(t, rollMsg.Label, "DEX")
		assert.Equal(t, "8d6", rollMsg.DiceExpr)
		assert.Equal(t, 0, rollMsg.Modifier)
		assert.Equal(t, RollDamage, rollMsg.RollType)
	})
}

// Verify BuildSpellRollCmd returns proper tea.Cmd type
func TestBuildSpellRollCmdReturnType(t *testing.T) {
	cmd := BuildSpellRollCmd("Test", "1d6", "fire", "", 5, 0)
	require.NotNil(t, cmd)
	var _ tea.Cmd = cmd // compile-time check
}
