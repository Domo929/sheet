package components

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

// BuildSpellRollCmd creates a tea.Cmd for a spell's dice roll based on its properties.
// For spell attacks (no saving throw): attack roll with damage follow-up.
// For save-based spells: damage roll only with DC shown in label.
// Returns nil for spells with no damage.
func BuildSpellRollCmd(spellName, damage, damageType, savingThrow string, attackBonus, saveDC int) tea.Cmd {
	if damage == "" {
		return nil
	}

	if savingThrow == "" {
		// Spell attack roll + damage follow-up
		return func() tea.Msg {
			return RequestRollMsg{
				Label:     spellName + " Attack",
				DiceExpr:  "1d20",
				Modifier:  attackBonus,
				RollType:  RollAttack,
				AdvPrompt: true,
				FollowUp: &RequestRollMsg{
					Label:    spellName + " Damage (" + damageType + ")",
					DiceExpr: damage,
					Modifier: 0,
					RollType: RollDamage,
				},
			}
		}
	}

	// Save-based spell — just roll damage, show DC in label
	return func() tea.Msg {
		return RequestRollMsg{
			Label:    fmt.Sprintf("%s Damage (DC %d %s)", spellName, saveDC, savingThrow),
			DiceExpr: damage,
			Modifier: 0,
			RollType: RollDamage,
		}
	}
}
