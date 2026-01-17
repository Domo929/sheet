package views

import (
	"fmt"
	"strings"

	"github.com/Domo929/sheet/internal/models"
	"github.com/Domo929/sheet/internal/storage"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MainSheetModel is the model for the main character sheet view.
type MainSheetModel struct {
	character *models.Character
	storage   *storage.CharacterStorage
	width     int
	height    int
	focusArea FocusArea
	keys      mainSheetKeyMap
}

// FocusArea represents which panel is currently focused.
type FocusArea int

const (
	FocusAbilitiesAndSaves FocusArea = iota
	FocusSkills
	FocusCombat
)

const numFocusAreas = 3

type mainSheetKeyMap struct {
	Quit       key.Binding
	Tab        key.Binding
	ShiftTab   key.Binding
	Inventory  key.Binding
	Spellbook  key.Binding
	Info       key.Binding
	Combat     key.Binding
	Rest       key.Binding
	Navigation key.Binding
}

func defaultMainSheetKeyMap() mainSheetKeyMap {
	return mainSheetKeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next panel"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev panel"),
		),
		Inventory: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "inventory"),
		),
		Spellbook: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "spellbook"),
		),
		Info: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "character info"),
		),
		Combat: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "combat"),
		),
		Rest: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rest"),
		),
		Navigation: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back to selection"),
		),
	}
}

// NewMainSheetModel creates a new main sheet model.
func NewMainSheetModel(character *models.Character, storage *storage.CharacterStorage) *MainSheetModel {
	return &MainSheetModel{
		character: character,
		storage:   storage,
		focusArea: FocusAbilitiesAndSaves,
		keys:      defaultMainSheetKeyMap(),
	}
}

// Init initializes the model.
func (m *MainSheetModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m *MainSheetModel) Update(msg tea.Msg) (*MainSheetModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Tab):
			m.focusArea = (m.focusArea + 1) % numFocusAreas
			return m, nil
		case key.Matches(msg, m.keys.ShiftTab):
			if m.focusArea == 0 {
				m.focusArea = numFocusAreas - 1
			} else {
				m.focusArea--
			}
			return m, nil
		case key.Matches(msg, m.keys.Navigation):
			return m, func() tea.Msg { return BackToSelectionMsg{} }
		}
	}

	return m, nil
}

// View renders the main sheet.
func (m *MainSheetModel) View() string {
	if m.character == nil {
		return "No character loaded"
	}

	// Calculate available width
	width := m.width
	if width == 0 {
		width = 120
	}
	height := m.height
	if height == 0 {
		height = 40
	}

	// Render sections
	header := m.renderHeader(width)
	
	// Main content area - three columns
	leftWidth := 22
	middleWidth := 30
	rightWidth := width - leftWidth - middleWidth - 8

	abilities := m.renderAbilities(leftWidth)
	skills := m.renderSkills(middleWidth)
	combat := m.renderCombatStats(rightWidth)

	// Join columns horizontally
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		abilities,
		"  ",
		skills,
		"  ",
		combat,
	)

	// Footer with navigation help
	footer := m.renderFooter(width)

	// Join all sections vertically
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		mainContent,
		"",
		footer,
	)
}

func (m *MainSheetModel) renderHeader(width int) string {
	char := m.character

	// Title style
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		MarginBottom(1)

	// Info style
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))

	// Build header content
	name := titleStyle.Render(char.Info.Name)

	// Race and class line
	raceClass := fmt.Sprintf("%s %s %d",
		char.Info.Race,
		char.Info.Class,
		char.Info.Level,
	)

	// XP or Milestone
	var progression string
	if char.Info.ProgressionType == models.ProgressionMilestone {
		progression = "Milestone"
	} else {
		progression = fmt.Sprintf("XP: %d / %d", char.Info.ExperiencePoints, models.XPForNextLevel(char.Info.Level))
	}

	// Inspiration
	inspiration := ""
	if char.Info.Inspiration {
		inspiration = " ★ Inspired"
	}

	// Build header
	headerLeft := lipgloss.JoinVertical(
		lipgloss.Left,
		name,
		infoStyle.Render(raceClass),
	)

	headerRight := lipgloss.JoinVertical(
		lipgloss.Right,
		labelStyle.Render("Proficiency: ")+infoStyle.Render(fmt.Sprintf("+%d", char.GetProficiencyBonus())),
		labelStyle.Render(progression)+infoStyle.Render(inspiration),
	)

	// Join header left and right
	headerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 2).
		Width(width - 2)

	headerContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		headerLeft,
		strings.Repeat(" ", max(1, width-lipgloss.Width(headerLeft)-lipgloss.Width(headerRight)-10)),
		headerRight,
	)

	return headerStyle.Render(headerContent)
}

func (m *MainSheetModel) renderAbilities(width int) string {
	char := m.character

	// Styles
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(width)

	if m.focusArea == FocusAbilitiesAndSaves {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("99"))
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99"))

	scoreStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("252"))

	modStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))

	// Build ability scores list
	abilities := []struct {
		name  string
		short string
		score models.AbilityScore
	}{
		{"Strength", "STR", char.AbilityScores.Strength},
		{"Dexterity", "DEX", char.AbilityScores.Dexterity},
		{"Constitution", "CON", char.AbilityScores.Constitution},
		{"Intelligence", "INT", char.AbilityScores.Intelligence},
		{"Wisdom", "WIS", char.AbilityScores.Wisdom},
		{"Charisma", "CHA", char.AbilityScores.Charisma},
	}

	var lines []string
	lines = append(lines, titleStyle.Render("Abilities"))
	lines = append(lines, "")

	for _, a := range abilities {
		mod := a.score.Modifier()
		modStr := formatModifier(mod)

		line := fmt.Sprintf("%-3s %s %s",
			a.short,
			scoreStyle.Render(fmt.Sprintf("%2d", a.score.Total())),
			modStyle.Render(fmt.Sprintf("(%s)", modStr)),
		)
		lines = append(lines, line)
	}

	// Add saving throws section
	lines = append(lines, "")
	lines = append(lines, titleStyle.Render("Saving Throws"))
	lines = append(lines, "")

	saves := []struct {
		name   string
		short  string
		save   *models.SavingThrow
		ability models.Ability
	}{
		{"Strength", "STR", &char.SavingThrows.Strength, models.AbilityStrength},
		{"Dexterity", "DEX", &char.SavingThrows.Dexterity, models.AbilityDexterity},
		{"Constitution", "CON", &char.SavingThrows.Constitution, models.AbilityConstitution},
		{"Intelligence", "INT", &char.SavingThrows.Intelligence, models.AbilityIntelligence},
		{"Wisdom", "WIS", &char.SavingThrows.Wisdom, models.AbilityWisdom},
		{"Charisma", "CHA", &char.SavingThrows.Charisma, models.AbilityCharisma},
	}

	profIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("76")).Render("●")
	noProfIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("○")

	for _, s := range saves {
		mod := char.GetSavingThrowModifier(s.ability)
		modStr := formatModifier(mod)

		icon := noProfIcon
		if s.save.Proficient {
			icon = profIcon
		}

		line := fmt.Sprintf("%s %-3s %s",
			icon,
			s.short,
			modStyle.Render(modStr),
		)
		lines = append(lines, line)
	}

	return panelStyle.Render(strings.Join(lines, "\n"))
}

func (m *MainSheetModel) renderSkills(width int) string {
	char := m.character

	// Styles
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(width)

	if m.focusArea == FocusSkills {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("99"))
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99"))

	modStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))

	profIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("76")).Render("●")
	expertIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("◆")
	noProfIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("○")

	// Skill display names
	skillNames := map[models.SkillName]string{
		models.SkillAcrobatics:     "Acrobatics",
		models.SkillAnimalHandling: "Animal Handling",
		models.SkillArcana:         "Arcana",
		models.SkillAthletics:      "Athletics",
		models.SkillDeception:      "Deception",
		models.SkillHistory:        "History",
		models.SkillInsight:        "Insight",
		models.SkillIntimidation:   "Intimidation",
		models.SkillInvestigation:  "Investigation",
		models.SkillMedicine:       "Medicine",
		models.SkillNature:         "Nature",
		models.SkillPerception:     "Perception",
		models.SkillPerformance:    "Performance",
		models.SkillPersuasion:     "Persuasion",
		models.SkillReligion:       "Religion",
		models.SkillSleightOfHand:  "Sleight of Hand",
		models.SkillStealth:        "Stealth",
		models.SkillSurvival:       "Survival",
	}

	// Skill ability abbreviations
	skillAbilityAbbr := map[models.Ability]string{
		models.AbilityStrength:     "STR",
		models.AbilityDexterity:    "DEX",
		models.AbilityConstitution: "CON",
		models.AbilityIntelligence: "INT",
		models.AbilityWisdom:       "WIS",
		models.AbilityCharisma:     "CHA",
	}

	var lines []string
	lines = append(lines, titleStyle.Render("Skills"))
	lines = append(lines, "")

	for _, skillName := range models.AllSkills() {
		skill := char.Skills.Get(skillName)
		ability := models.GetSkillAbility(skillName)
		mod := char.GetSkillModifier(skillName)
		modStr := formatModifier(mod)

		icon := noProfIcon
		if skill.Proficiency == models.Proficient {
			icon = profIcon
		} else if skill.Proficiency == models.Expertise {
			icon = expertIcon
		}

		displayName := skillNames[skillName]
		abilityAbbr := skillAbilityAbbr[ability]

		line := fmt.Sprintf("%s %s %-15s %s",
			icon,
			modStyle.Render(fmt.Sprintf("(%s)", abilityAbbr)),
			displayName,
			modStyle.Render(modStr),
		)
		lines = append(lines, line)
	}

	return panelStyle.Render(strings.Join(lines, "\n"))
}

func (m *MainSheetModel) renderCombatStats(width int) string {
	char := m.character

	// Styles
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(width)

	if m.focusArea == FocusCombat {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("99"))
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))

	valueStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("252"))

	var lines []string
	lines = append(lines, titleStyle.Render("Combat"))
	lines = append(lines, "")

	// HP
	hp := char.CombatStats.HitPoints
	hpPercent := float64(hp.Current) / float64(hp.Maximum)
	hpColor := "76" // green
	if hpPercent < 0.25 {
		hpColor = "196" // red
	} else if hpPercent < 0.5 {
		hpColor = "214" // yellow
	}

	hpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(hpColor))
	hpLine := fmt.Sprintf("%s %s / %s",
		labelStyle.Render("HP:"),
		hpStyle.Render(fmt.Sprintf("%d", hp.Current)),
		valueStyle.Render(fmt.Sprintf("%d", hp.Maximum)),
	)
	if hp.Temporary > 0 {
		hpLine += labelStyle.Render(fmt.Sprintf(" (+%d temp)", hp.Temporary))
	}
	lines = append(lines, hpLine)

	// AC
	lines = append(lines, fmt.Sprintf("%s %s",
		labelStyle.Render("AC:"),
		valueStyle.Render(fmt.Sprintf("%d", char.CombatStats.ArmorClass)),
	))

	// Initiative
	init := char.GetInitiative()
	lines = append(lines, fmt.Sprintf("%s %s",
		labelStyle.Render("Initiative:"),
		valueStyle.Render(formatModifier(init)),
	))

	// Speed
	lines = append(lines, fmt.Sprintf("%s %s ft",
		labelStyle.Render("Speed:"),
		valueStyle.Render(fmt.Sprintf("%d", char.CombatStats.Speed)),
	))

	// Hit Dice
	hd := char.CombatStats.HitDice
	lines = append(lines, fmt.Sprintf("%s %s/%s d%d",
		labelStyle.Render("Hit Dice:"),
		valueStyle.Render(fmt.Sprintf("%d", hd.Remaining)),
		valueStyle.Render(fmt.Sprintf("%d", hd.Total)),
		hd.DieType,
	))

	// Death Saves (only show if relevant)
	if hp.Current == 0 {
		lines = append(lines, "")
		lines = append(lines, titleStyle.Render("Death Saves"))
		ds := char.CombatStats.DeathSaves
		successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("76"))
		failStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

		successes := strings.Repeat(successStyle.Render("●"), ds.Successes) +
			strings.Repeat(labelStyle.Render("○"), 3-ds.Successes)
		failures := strings.Repeat(failStyle.Render("●"), ds.Failures) +
			strings.Repeat(labelStyle.Render("○"), 3-ds.Failures)

		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Successes:"), successes))
		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Failures: "), failures))
	}

	// Spellcasting info if applicable
	if char.IsSpellcaster() {
		lines = append(lines, "")
		lines = append(lines, titleStyle.Render("Spellcasting"))

		lines = append(lines, fmt.Sprintf("%s %s",
			labelStyle.Render("Spell Save DC:"),
			valueStyle.Render(fmt.Sprintf("%d", char.GetSpellSaveDC())),
		))
		lines = append(lines, fmt.Sprintf("%s %s",
			labelStyle.Render("Spell Attack:"),
			valueStyle.Render(formatModifier(char.GetSpellAttackBonus())),
		))
	}

	// Active conditions
	if len(char.CombatStats.Conditions) > 0 {
		lines = append(lines, "")
		lines = append(lines, titleStyle.Render("Conditions"))
		condStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
		for _, cond := range char.CombatStats.Conditions {
			lines = append(lines, condStyle.Render("• "+string(cond)))
		}
	}

	return panelStyle.Render(strings.Join(lines, "\n"))
}

func (m *MainSheetModel) renderFooter(width int) string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Width(width)

	help := "tab/shift+tab: navigate panels • i: inventory • s: spellbook • c: character info • r: rest • q: quit"

	return footerStyle.Render(help)
}

// formatModifier formats an integer as a modifier string (e.g., +2 or -1).
func formatModifier(mod int) string {
	if mod >= 0 {
		return fmt.Sprintf("+%d", mod)
	}
	return fmt.Sprintf("%d", mod)
}

// BackToSelectionMsg is sent when the user wants to return to character selection.
type BackToSelectionMsg struct{}
