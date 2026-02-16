package components

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Domo929/roll"
)

// rollState represents the current state of the roll engine.
type rollState int

const (
	rollStateIdle       rollState = iota
	rollStateAdvPrompt            // showing Normal/Advantage/Disadvantage picker
	rollStateCustomRoll           // showing die type picker with quantity selector
	rollStateAnimating            // tumbling dice with color flashing
	rollStateShowing              // final result displayed, waiting for dismissal
)

// Animation constants.
const (
	totalAnimFrames   = 12
	baseFrameDelay    = 80 * time.Millisecond
	frameDelayAccel   = 15 * time.Millisecond
	maxVisibleDice    = 10
	customRollMinQty  = 1
	customRollMaxQty  = 100
)

// animColors is the palette cycled during animation: magenta, cyan, yellow, white.
var animColors = []string{"99", "14", "11", "15"}

// dieTypes is the set of available die types for custom rolls.
var dieTypes = []int{4, 6, 8, 10, 12, 20, 100}

// --- Message Types ---

// RequestRollMsg is sent by views to request a dice roll.
type RequestRollMsg struct {
	Label     string
	DiceExpr  string // e.g., "1d20"
	Modifier  int
	RollType  RollType
	AdvPrompt bool
	FollowUp  *RequestRollMsg
}

// RollCompleteMsg is sent by the engine when the user dismisses the result.
type RollCompleteMsg struct {
	Entry RollHistoryEntry
}

// RollTickMsg is an internal animation tick.
type RollTickMsg struct {
	Time time.Time
}

// OpenCustomRollMsg triggers the custom dice roller overlay.
type OpenCustomRollMsg struct{}

// ToggleRollHistoryMsg triggers history toggle.
type ToggleRollHistoryMsg struct{}

// --- Roll Engine ---

// RollEngine implements the central dice-rolling state machine. Views never call
// the roll library directly — they send RequestRollMsg and receive RollCompleteMsg.
type RollEngine struct {
	state rollState

	// Pending roll request
	pendingRoll *RequestRollMsg

	// Animation state
	currentFrame int
	totalFrames  int
	displayVals  []int
	displayColors []string
	colorIndex   int

	// Result
	finalResult  *roll.Result
	rollEntry    RollHistoryEntry
	advantage    bool
	disadvantage bool

	// Follow-up
	followUp *RequestRollMsg

	// Custom roll state
	selectedDie int // index into dieTypes
	quantity    int // 1-100
}

// NewRollEngine creates a new RollEngine in the Idle state.
func NewRollEngine() *RollEngine {
	return &RollEngine{
		state:       rollStateIdle,
		totalFrames: totalAnimFrames,
		quantity:    1,
	}
}

// IsActive returns true if the engine is in any state other than Idle.
func (e *RollEngine) IsActive() bool {
	return e.state != rollStateIdle
}

// State returns the current state (for testing).
func (e *RollEngine) State() rollState {
	return e.state
}

// OpenCustomRoll enters the CustomRoll state.
func (e *RollEngine) OpenCustomRoll() {
	e.state = rollStateCustomRoll
	e.selectedDie = 0
	e.quantity = 1
}

// Update processes messages and returns commands.
func (e *RollEngine) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case RequestRollMsg:
		return e.handleRequestRoll(msg)
	case RollTickMsg:
		return e.handleTick()
	case OpenCustomRollMsg:
		if e.state == rollStateIdle {
			e.OpenCustomRoll()
		}
		return nil
	case tea.KeyMsg:
		return e.handleKey(msg)
	}
	return nil
}

// handleRequestRoll processes an incoming roll request.
func (e *RollEngine) handleRequestRoll(msg RequestRollMsg) tea.Cmd {
	if e.state != rollStateIdle {
		return nil // locked during animation
	}

	e.pendingRoll = &msg
	e.followUp = msg.FollowUp
	e.advantage = false
	e.disadvantage = false

	if msg.AdvPrompt {
		e.state = rollStateAdvPrompt
		return nil
	}

	return e.executeAndAnimate(msg.DiceExpr, msg.Modifier, false, false)
}

// handleKey processes key input based on current state.
func (e *RollEngine) handleKey(msg tea.KeyMsg) tea.Cmd {
	switch e.state {
	case rollStateAdvPrompt:
		return e.handleAdvPromptKey(msg)
	case rollStateShowing:
		return e.handleShowingKey(msg)
	case rollStateCustomRoll:
		return e.handleCustomRollKey(msg)
	}
	return nil
}

// handleAdvPromptKey handles keys in the AdvPrompt state.
func (e *RollEngine) handleAdvPromptKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "n", "N":
		return e.executeAndAnimate(e.pendingRoll.DiceExpr, e.pendingRoll.Modifier, false, false)
	case "a", "A":
		return e.executeAndAnimate(e.pendingRoll.DiceExpr, e.pendingRoll.Modifier, true, false)
	case "d", "D":
		return e.executeAndAnimate(e.pendingRoll.DiceExpr, e.pendingRoll.Modifier, false, true)
	case "esc":
		e.resetToIdle()
		return nil
	}
	return nil
}

// handleShowingKey handles keys in the Showing state.
func (e *RollEngine) handleShowingKey(msg tea.KeyMsg) tea.Cmd {
	if e.followUp != nil {
		switch msg.String() {
		case "enter":
			// Start the follow-up roll
			followUp := *e.followUp
			e.resetToIdle()
			return func() tea.Msg { return followUp }
		case "esc":
			entry := e.rollEntry
			e.resetToIdle()
			return func() tea.Msg { return RollCompleteMsg{Entry: entry} }
		}
		return nil
	}

	// No follow-up: any key dismisses
	entry := e.rollEntry
	e.resetToIdle()
	return func() tea.Msg { return RollCompleteMsg{Entry: entry} }
}

// handleCustomRollKey handles keys in the CustomRoll state.
func (e *RollEngine) handleCustomRollKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "left", "h":
		if e.selectedDie > 0 {
			e.selectedDie--
		}
	case "right", "l":
		if e.selectedDie < len(dieTypes)-1 {
			e.selectedDie++
		}
	case "up", "k":
		if e.quantity < customRollMaxQty {
			e.quantity++
		}
	case "down", "j":
		if e.quantity > customRollMinQty {
			e.quantity--
		}
	case "enter":
		sides := dieTypes[e.selectedDie]
		expr := fmt.Sprintf("%dd%d", e.quantity, sides)
		e.pendingRoll = &RequestRollMsg{
			Label:    fmt.Sprintf("Custom Roll: %s", expr),
			DiceExpr: expr,
			RollType: RollCustom,
		}
		e.followUp = nil
		return e.executeAndAnimate(expr, 0, false, false)
	case "esc":
		e.resetToIdle()
	}
	return nil
}

// executeAndAnimate performs the roll and starts the animation.
func (e *RollEngine) executeAndAnimate(expr string, modifier int, adv, disadv bool) tea.Cmd {
	e.advantage = adv
	e.disadvantage = disadv

	result, err := e.executeRoll(expr, modifier, adv, disadv)
	if err != nil {
		// On error, just reset — in a real app we might show an error
		e.resetToIdle()
		return nil
	}

	e.finalResult = result
	e.state = rollStateAnimating
	e.currentFrame = 0
	e.colorIndex = 0

	// Generate random starting display values
	numDice := len(result.Kept) + len(result.Dropped)
	if numDice == 0 {
		numDice = len(result.Rolls)
	}
	sides := result.Sides
	if sides == 0 {
		sides = 20
	}

	e.displayVals = make([]int, numDice)
	for i := range e.displayVals {
		e.displayVals[i] = rand.Intn(sides) + 1
	}

	e.displayColors = make([]string, numDice)
	for i := range e.displayColors {
		e.displayColors[i] = animColors[0]
	}

	// Build history entry now (values are final even if animation hasn't finished)
	e.rollEntry = e.buildHistoryEntry()

	delay := baseFrameDelay + time.Duration(e.currentFrame)*frameDelayAccel
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return RollTickMsg{Time: t}
	})
}

// handleTick processes an animation tick.
func (e *RollEngine) handleTick() tea.Cmd {
	if e.state != rollStateAnimating {
		return nil
	}

	e.currentFrame++
	numDice := len(e.displayVals)
	sides := e.finalResult.Sides
	if sides == 0 {
		sides = 20
	}

	// Combine kept + dropped to get all dice in order for final values
	allFinalVals := e.getAllFinalVals()

	// Cycle color index
	e.colorIndex = (e.colorIndex + 1) % len(animColors)

	for i := 0; i < numDice; i++ {
		if e.isDieLanded(i, numDice) {
			// Landed: show final value with green color
			if i < len(allFinalVals) {
				e.displayVals[i] = allFinalVals[i]
			}
			e.displayColors[i] = "10" // green
		} else {
			// Still tumbling: random value, cycling color
			e.displayVals[i] = rand.Intn(sides) + 1
			e.displayColors[i] = animColors[e.colorIndex]
		}
	}

	if e.currentFrame >= e.totalFrames {
		// Ensure all dice show final values
		for i := 0; i < numDice; i++ {
			if i < len(allFinalVals) {
				e.displayVals[i] = allFinalVals[i]
			}
			e.displayColors[i] = "10"
		}
		e.state = rollStateShowing
		return nil
	}

	delay := baseFrameDelay + time.Duration(e.currentFrame)*frameDelayAccel
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return RollTickMsg{Time: t}
	})
}

// isDieLanded returns true if die at index i has landed.
// Dice land from left to right: die i lands when currentFrame > totalFrames - numDice + i
func (e *RollEngine) isDieLanded(i, numDice int) bool {
	return e.currentFrame > e.totalFrames-numDice+i
}

// getAllFinalVals returns all final die values (kept followed by dropped).
func (e *RollEngine) getAllFinalVals() []int {
	if e.finalResult == nil {
		return nil
	}
	vals := make([]int, 0, len(e.finalResult.Kept)+len(e.finalResult.Dropped))
	vals = append(vals, e.finalResult.Kept...)
	vals = append(vals, e.finalResult.Dropped...)
	return vals
}

// executeRoll calls the roll library to perform the actual dice roll.
func (e *RollEngine) executeRoll(expr string, modifier int, adv, disadv bool) (*roll.Result, error) {
	if adv {
		return roll.RollAdvantage(modifier)
	}
	if disadv {
		return roll.RollDisadvantage(modifier)
	}

	result, err := roll.RollString(expr)
	if err != nil {
		return nil, err
	}

	// If there's an additional modifier not baked into the expression, add it
	if modifier != 0 && result.Modifier == 0 {
		result.Modifier = modifier
		result.Total += modifier
	}

	return result, nil
}

// buildHistoryEntry constructs a RollHistoryEntry from the current result and metadata.
func (e *RollEngine) buildHistoryEntry() RollHistoryEntry {
	if e.finalResult == nil || e.pendingRoll == nil {
		return RollHistoryEntry{}
	}

	entry := RollHistoryEntry{
		Label:        e.pendingRoll.Label,
		RollType:     e.pendingRoll.RollType,
		Expression:   e.finalResult.Expression,
		Rolls:        e.finalResult.Rolls,
		Kept:         e.finalResult.Kept,
		Dropped:      e.finalResult.Dropped,
		Modifier:     e.finalResult.Modifier,
		Total:        e.finalResult.Total,
		Advantage:    e.advantage,
		Disadvantage: e.disadvantage,
		Timestamp:    time.Now(),
	}

	// Nat crit/fail detection: only on d20 rolls
	if e.isD20Roll() {
		for _, v := range e.finalResult.Kept {
			if v == 20 {
				entry.NatCrit = true
			}
			if v == 1 {
				entry.NatFail = true
			}
		}
	}

	return entry
}

// isD20Roll returns true if this is a d20 roll.
func (e *RollEngine) isD20Roll() bool {
	if e.finalResult != nil && e.finalResult.Sides == 20 {
		return true
	}
	if e.pendingRoll != nil && strings.Contains(strings.ToLower(e.pendingRoll.DiceExpr), "d20") {
		return true
	}
	return false
}

// resetToIdle clears all state and returns to idle.
func (e *RollEngine) resetToIdle() {
	e.state = rollStateIdle
	e.pendingRoll = nil
	e.finalResult = nil
	e.followUp = nil
	e.currentFrame = 0
	e.displayVals = nil
	e.displayColors = nil
	e.colorIndex = 0
	e.advantage = false
	e.disadvantage = false
	e.rollEntry = RollHistoryEntry{}
}

// --- Rendering ---

// View renders the roll engine modal overlay. Returns empty string when idle.
func (e *RollEngine) View(underlayWidth, underlayHeight int) string {
	switch e.state {
	case rollStateAdvPrompt:
		return e.renderAdvPrompt(underlayWidth, underlayHeight)
	case rollStateAnimating:
		return e.renderAnimating(underlayWidth, underlayHeight)
	case rollStateShowing:
		return e.renderShowing(underlayWidth, underlayHeight)
	case rollStateCustomRoll:
		return e.renderCustomRoll(underlayWidth, underlayHeight)
	default:
		return ""
	}
}

// renderAdvPrompt renders the advantage/disadvantage picker.
func (e *RollEngine) renderAdvPrompt(width, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	optionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	highlightStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))

	var content strings.Builder
	content.WriteString(titleStyle.Render("Roll Mode"))
	content.WriteString("\n\n")
	content.WriteString("  " + highlightStyle.Render("[N]") + optionStyle.Render("ormal"))
	content.WriteString("\n")
	content.WriteString("  " + highlightStyle.Render("[A]") + optionStyle.Render("dvantage"))
	content.WriteString("\n")
	content.WriteString("  " + highlightStyle.Render("[D]") + optionStyle.Render("isadvantage"))

	return e.renderModal(content.String(), width, height)
}

// renderAnimating renders the dice animation.
func (e *RollEngine) renderAnimating(width, height int) string {
	label := ""
	if e.pendingRoll != nil {
		icon := RollTypeIcon(e.pendingRoll.RollType)
		label = fmt.Sprintf("%s %s", icon, e.pendingRoll.Label)
	}

	var content strings.Builder
	if label != "" {
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
		content.WriteString(titleStyle.Render(label))
		content.WriteString("\n")
	}
	content.WriteString("\n")
	content.WriteString(e.renderDiceRow(maxVisibleDice))
	content.WriteString("\n")

	return e.renderModal(content.String(), width, height)
}

// renderShowing renders the final result with breakdown.
func (e *RollEngine) renderShowing(width, height int) string {
	label := ""
	if e.pendingRoll != nil {
		icon := RollTypeIcon(e.pendingRoll.RollType)
		label = fmt.Sprintf("%s %s", icon, e.pendingRoll.Label)
	}

	natCrit := e.rollEntry.NatCrit
	natFail := e.rollEntry.NatFail
	isLuck := e.pendingRoll != nil && e.pendingRoll.RollType == RollLuck

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	totalStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)

	if natCrit {
		totalStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	} else if natFail {
		totalStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	} else if isLuck {
		titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
		totalStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	}

	var content strings.Builder
	if label != "" {
		content.WriteString(titleStyle.Render(label))
		content.WriteString("\n")
	}
	content.WriteString("\n")
	content.WriteString(e.renderDiceRow(maxVisibleDice))
	content.WriteString("\n\n")

	// Breakdown line
	if e.finalResult != nil {
		breakdown := e.buildBreakdown()
		content.WriteString("  " + totalStyle.Render(breakdown))
		content.WriteString("\n")

		// Advantage/disadvantage details
		if e.advantage && len(e.finalResult.Dropped) > 0 {
			advDetail := fmt.Sprintf("  Advantage: kept %s, dropped %s",
				formatIntSlice(e.finalResult.Kept),
				formatIntSlice(e.finalResult.Dropped))
			content.WriteString(dimStyle.Render(advDetail))
			content.WriteString("\n")
		} else if e.disadvantage && len(e.finalResult.Dropped) > 0 {
			disDetail := fmt.Sprintf("  Disadvantage: kept %s, dropped %s",
				formatIntSlice(e.finalResult.Kept),
				formatIntSlice(e.finalResult.Dropped))
			content.WriteString(dimStyle.Render(disDetail))
			content.WriteString("\n")
		}

		// Nat crit/fail indicator
		if natCrit {
			critStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
			content.WriteString("\n  " + critStyle.Render("NATURAL 20!"))
			content.WriteString("\n")
		} else if natFail {
			failStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
			content.WriteString("\n  " + failStyle.Render("NATURAL 1!"))
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")

	// Prompt
	if e.followUp != nil {
		content.WriteString(promptStyle.Render("  Enter: roll damage \u2022 Esc: skip"))
	} else {
		content.WriteString(promptStyle.Render("  Press any key to dismiss"))
	}

	return e.renderModal(content.String(), width, height)
}

// renderCustomRoll renders the custom dice roller.
func (e *RollEngine) renderCustomRoll(width, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")).
		Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("11")).
		Padding(0, 1)
	unselectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).
		Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)
	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	quantityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	var content strings.Builder
	content.WriteString(titleStyle.Render("Custom Roll"))
	content.WriteString("\n\n")

	// Die type selector
	dieButtons := make([]string, len(dieTypes))
	for i, dt := range dieTypes {
		label := fmt.Sprintf("d%d", dt)
		if i == e.selectedDie {
			dieButtons[i] = selectedStyle.Render(label)
		} else {
			dieButtons[i] = unselectedStyle.Render(label)
		}
	}
	content.WriteString("  " + lipgloss.JoinHorizontal(lipgloss.Center, dieButtons...))
	content.WriteString("\n\n")

	// Quantity
	content.WriteString(quantityStyle.Render(fmt.Sprintf("  Quantity: %d", e.quantity)))
	content.WriteString("        ")
	content.WriteString(promptStyle.Render("\u2190 / \u2192 to change die"))
	content.WriteString("\n")
	content.WriteString("                    ")
	content.WriteString(promptStyle.Render("\u2191 / \u2193 to change quantity"))
	content.WriteString("\n\n")

	content.WriteString(promptStyle.Render("  Enter: roll \u2022 Esc: cancel"))

	return e.renderModal(content.String(), width, height)
}

// renderModal wraps content in a centered bordered modal overlay.
func (e *RollEngine) renderModal(content string, underlayWidth, underlayHeight int) string {
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1, 2)

	rendered := modalStyle.Render(content)

	// Center the modal in the available space
	renderedWidth := lipgloss.Width(rendered)
	renderedHeight := lipgloss.Height(rendered)

	if underlayWidth <= 0 || underlayHeight <= 0 {
		return rendered
	}

	hPad := 0
	if underlayWidth > renderedWidth {
		hPad = (underlayWidth - renderedWidth) / 2
	}
	vPad := 0
	if underlayHeight > renderedHeight {
		vPad = (underlayHeight - renderedHeight) / 2
	}

	centered := lipgloss.NewStyle().
		MarginLeft(hPad).
		MarginTop(vPad).
		Render(rendered)

	return centered
}

// renderDie renders a single die box.
func (e *RollEngine) renderDie(value int, color string, landed bool) string {
	dieColor := lipgloss.Color(color)

	// Check for nat crit/fail on landed d20 dice
	if landed && e.isD20Roll() {
		if value == 20 {
			dieColor = lipgloss.Color("10") // green
		} else if value == 1 {
			dieColor = lipgloss.Color("9") // red
		}
	}

	style := lipgloss.NewStyle().Foreground(dieColor)
	if landed {
		style = style.Bold(true)
	}

	valStr := fmt.Sprintf("%3d", value)
	top := style.Render("\u256d\u2500\u2500\u2500\u256e")
	mid := style.Render("\u2502") + style.Render(valStr) + style.Render("\u2502")
	bot := style.Render("\u2570\u2500\u2500\u2500\u256f")

	return top + "\n" + mid + "\n" + bot
}

// renderDiceRow renders all dice side by side.
func (e *RollEngine) renderDiceRow(maxVisible int) string {
	numDice := len(e.displayVals)
	if numDice == 0 {
		return ""
	}

	visible := numDice
	if visible > maxVisible {
		visible = maxVisible
	}

	diceStrings := make([]string, visible)
	for i := 0; i < visible; i++ {
		color := "15" // default white
		if i < len(e.displayColors) {
			color = e.displayColors[i]
		}
		landed := e.state == rollStateShowing || e.isDieLanded(i, numDice)
		diceStrings[i] = e.renderDie(e.displayVals[i], color, landed)
	}

	row := "  " + lipgloss.JoinHorizontal(lipgloss.Top, intersperse(diceStrings, "  ")...)

	if numDice > maxVisible {
		extra := numDice - maxVisible
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		row += "\n" + dimStyle.Render(fmt.Sprintf("  ... and %d more", extra))
	}

	return row
}

// buildBreakdown builds the result breakdown string, e.g., "(14) + 7 = 21".
func (e *RollEngine) buildBreakdown() string {
	if e.finalResult == nil {
		return ""
	}

	result := e.finalResult

	// Format kept dice
	parts := make([]string, len(result.Kept))
	for i, v := range result.Kept {
		parts[i] = fmt.Sprintf("%d", v)
	}

	var breakdown string
	if len(result.Kept) == 1 {
		breakdown = fmt.Sprintf("(%s)", parts[0])
	} else {
		breakdown = fmt.Sprintf("(%s)", strings.Join(parts, " + "))
	}

	if result.Modifier != 0 {
		breakdown += fmt.Sprintf(" %+d", result.Modifier)
	}

	breakdown += fmt.Sprintf(" = %d", result.Total)

	return breakdown
}

// --- Helpers ---

// formatIntSlice formats a slice of ints as a comma-separated string.
func formatIntSlice(vals []int) string {
	parts := make([]string, len(vals))
	for i, v := range vals {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(parts, ", ")
}

// intersperse inserts a separator element between each element in a slice.
func intersperse(items []string, sep string) []string {
	if len(items) <= 1 {
		return items
	}
	result := make([]string, 0, len(items)*2-1)
	for i, item := range items {
		if i > 0 {
			result = append(result, sep)
		}
		result = append(result, item)
	}
	return result
}
