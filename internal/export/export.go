// Package export renders a character to shareable, read-only formats (Markdown,
// JSON, and a printable HTML sheet) suitable for printing or posting alongside
// a game.
package export

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Domo929/sheet/internal/models"
)

// ToJSON serializes the character to indented JSON bytes.
func ToJSON(c *models.Character) ([]byte, error) {
	return c.ToJSON()
}

// SanitizeFilename converts a character name into a safe base filename.
func SanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	for _, char := range []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"} {
		name = strings.ReplaceAll(name, char, "")
	}
	if name == "" {
		name = "character"
	}
	return name
}

// WriteFiles writes "<name>.md", "<name>.json", and a printable "<name>.html"
// into dir, creating dir if necessary, and returns the Markdown and JSON paths.
// The HTML file is a self-contained, print-ready sheet (open in a browser and
// Save as PDF).
func WriteFiles(c *models.Character, dir string) (mdPath, jsonPath string, err error) {
	if c == nil {
		return "", "", fmt.Errorf("character is nil")
	}
	if err = os.MkdirAll(dir, 0o755); err != nil {
		return "", "", fmt.Errorf("create export dir: %w", err)
	}
	base := SanitizeFilename(c.Info.Name)
	mdPath = filepath.Join(dir, base+".md")
	jsonPath = filepath.Join(dir, base+".json")
	htmlPath := filepath.Join(dir, base+".html")

	if err = os.WriteFile(mdPath, []byte(ToMarkdown(c)), 0o644); err != nil {
		return "", "", fmt.Errorf("write markdown: %w", err)
	}
	jsonBytes, jerrr := ToJSON(c)
	if jerrr != nil {
		return "", "", fmt.Errorf("encode json: %w", jerrr)
	}
	if err = os.WriteFile(jsonPath, jsonBytes, 0o644); err != nil {
		return "", "", fmt.Errorf("write json: %w", err)
	}
	if err = os.WriteFile(htmlPath, []byte(ToHTML(c)), 0o644); err != nil {
		return "", "", fmt.Errorf("write html: %w", err)
	}
	return mdPath, jsonPath, nil
}

// signed formats an integer modifier with an explicit sign (e.g. +3, -1, +0).
func signed(n int) string {
	if n >= 0 {
		return fmt.Sprintf("+%d", n)
	}
	return fmt.Sprintf("%d", n)
}

var abilityOrder = []struct {
	Ability models.Ability
	Label   string
}{
	{models.AbilityStrength, "Strength"},
	{models.AbilityDexterity, "Dexterity"},
	{models.AbilityConstitution, "Constitution"},
	{models.AbilityIntelligence, "Intelligence"},
	{models.AbilityWisdom, "Wisdom"},
	{models.AbilityCharisma, "Charisma"},
}

var skillDisplay = map[models.SkillName]string{
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

var abilityAbbrev = map[models.Ability]string{
	models.AbilityStrength:     "STR",
	models.AbilityDexterity:    "DEX",
	models.AbilityConstitution: "CON",
	models.AbilityIntelligence: "INT",
	models.AbilityWisdom:       "WIS",
	models.AbilityCharisma:     "CHA",
}

// ToMarkdown renders a complete, read-only character sheet as Markdown.
func ToMarkdown(c *models.Character) string {
	var b strings.Builder
	writeHeader(&b, c)
	writeCoreStats(&b, c)
	writeAbilities(&b, c)
	writeSkills(&b, c)
	writeProficiencies(&b, c)
	writeFeatures(&b, c)
	writeSpellcasting(&b, c)
	writeInventory(&b, c)
	writeCompanions(&b, c)
	writePersonality(&b, c)
	return b.String()
}

func writeHeader(b *strings.Builder, c *models.Character) {
	info := c.Info
	fmt.Fprintf(b, "# %s\n\n", info.Name)

	race := info.Race
	if info.Subrace != "" {
		race = fmt.Sprintf("%s (%s)", info.Race, info.Subrace)
	}
	class := info.Class
	if info.Subclass != "" {
		class = fmt.Sprintf("%s (%s)", info.Class, info.Subclass)
	}
	if c.IsMulticlass() {
		fmt.Fprintf(b, "**Level %d %s — %s**\n\n", c.TotalLevel(), race, c.ClassSummary())
	} else {
		fmt.Fprintf(b, "**Level %d %s %s**\n\n", info.Level, race, class)
	}

	var meta []string
	if info.Background != "" {
		meta = append(meta, "Background: "+info.Background)
	}
	if info.Alignment != "" {
		meta = append(meta, "Alignment: "+string(info.Alignment))
	}
	if info.PlayerName != "" {
		meta = append(meta, "Player: "+info.PlayerName)
	}
	if len(meta) > 0 {
		fmt.Fprintf(b, "%s\n\n", strings.Join(meta, " • "))
	}
}

func writeCoreStats(b *strings.Builder, c *models.Character) {
	hp := c.CombatStats.HitPoints
	hpStr := fmt.Sprintf("%d / %d", hp.Current, hp.Maximum)
	if hp.Temporary > 0 {
		hpStr += fmt.Sprintf(" (+%d temp)", hp.Temporary)
	}

	b.WriteString("## Combat\n\n")
	b.WriteString("| Stat | Value |\n|------|-------|\n")
	fmt.Fprintf(b, "| Armor Class | %d |\n", c.CombatStats.ArmorClass)
	fmt.Fprintf(b, "| Hit Points | %s |\n", hpStr)
	fmt.Fprintf(b, "| Hit Dice | %d/%d d%d |\n", c.CombatStats.HitDice.Remaining, c.CombatStats.HitDice.Total, c.CombatStats.HitDice.DieType)
	fmt.Fprintf(b, "| Initiative | %s |\n", signed(c.GetInitiative()))
	fmt.Fprintf(b, "| Speed | %s |\n", speedString(c))
	fmt.Fprintf(b, "| Proficiency Bonus | %s |\n", signed(c.GetProficiencyBonus()))
	fmt.Fprintf(b, "| Passive Perception | %d |\n", 10+c.GetSkillModifier(models.SkillPerception))
	if c.CombatStats.ExhaustionLevel > 0 {
		fmt.Fprintf(b, "| Exhaustion | Level %d |\n", c.CombatStats.ExhaustionLevel)
	}
	if senses := c.CombatStats.Senses.List(); len(senses) > 0 {
		fmt.Fprintf(b, "| Senses | %s |\n", strings.Join(senses, ", "))
	}
	if len(c.CombatStats.Conditions) > 0 {
		conds := make([]string, len(c.CombatStats.Conditions))
		for i, cond := range c.CombatStats.Conditions {
			conds[i] = string(cond)
		}
		fmt.Fprintf(b, "| Conditions | %s |\n", strings.Join(conds, ", "))
	}
	b.WriteString("\n")
}

func speedString(c *models.Character) string {
	parts := []string{fmt.Sprintf("%d ft", c.GetEffectiveSpeed())}
	if c.CombatStats.FlySpeed > 0 {
		parts = append(parts, fmt.Sprintf("Fly %d", c.CombatStats.FlySpeed))
	}
	if c.CombatStats.SwimSpeed > 0 {
		parts = append(parts, fmt.Sprintf("Swim %d", c.CombatStats.SwimSpeed))
	}
	if c.CombatStats.ClimbSpeed > 0 {
		parts = append(parts, fmt.Sprintf("Climb %d", c.CombatStats.ClimbSpeed))
	}
	if c.CombatStats.BurrowSpeed > 0 {
		parts = append(parts, fmt.Sprintf("Burrow %d", c.CombatStats.BurrowSpeed))
	}
	return strings.Join(parts, ", ")
}

func writeAbilities(b *strings.Builder, c *models.Character) {
	b.WriteString("## Ability Scores\n\n")
	b.WriteString("| Ability | Score | Mod | Save |\n|---------|:-----:|:---:|:----:|\n")
	for _, a := range abilityOrder {
		score := c.AbilityScores.Get(a.Ability)
		save := signed(c.GetSavingThrowModifier(a.Ability))
		if c.SavingThrows.IsProficient(a.Ability) {
			save += " *(prof)*"
		}
		fmt.Fprintf(b, "| %s | %d | %s | %s |\n", a.Label, score.Total(), signed(score.Modifier()), save)
	}
	b.WriteString("\n")
}

func writeSkills(b *strings.Builder, c *models.Character) {
	b.WriteString("## Skills\n\n")
	b.WriteString("| Skill | Ability | Bonus | Proficiency |\n|-------|:-------:|:-----:|:-----------:|\n")
	for _, name := range models.AllSkills() {
		display := skillDisplay[name]
		if display == "" {
			display = string(name)
		}
		abbr := abilityAbbrev[models.GetSkillAbility(name)]
		prof := ""
		switch c.Skills.Get(name).Proficiency {
		case models.Proficient:
			prof = "●"
		case models.Expertise:
			prof = "●● (expertise)"
		}
		fmt.Fprintf(b, "| %s | %s | %s | %s |\n", display, abbr, signed(c.GetSkillModifier(name)), prof)
	}
	b.WriteString("\n")
}

func writeProficiencies(b *strings.Builder, c *models.Character) {
	p := c.Proficiencies
	if len(p.Armor) == 0 && len(p.Weapons) == 0 && len(p.Tools) == 0 && len(p.Languages) == 0 {
		return
	}
	b.WriteString("## Proficiencies & Languages\n\n")
	writeList := func(label string, items []string) {
		if len(items) > 0 {
			fmt.Fprintf(b, "- **%s:** %s\n", label, strings.Join(items, ", "))
		}
	}
	writeList("Armor", p.Armor)
	writeList("Weapons", p.Weapons)
	writeList("Tools", p.Tools)
	writeList("Languages", p.Languages)
	b.WriteString("\n")
}

func writeFeatures(b *strings.Builder, c *models.Character) {
	f := c.Features
	if len(f.RacialTraits) == 0 && len(f.ClassFeatures) == 0 && len(f.Feats) == 0 {
		return
	}
	b.WriteString("## Features & Traits\n\n")
	section := func(title string, feats []models.Feature) {
		if len(feats) == 0 {
			return
		}
		fmt.Fprintf(b, "### %s\n\n", title)
		for _, ft := range feats {
			if ft.Source != "" {
				fmt.Fprintf(b, "- **%s** *(%s)*", ft.Name, ft.Source)
			} else {
				fmt.Fprintf(b, "- **%s**", ft.Name)
			}
			if ft.Description != "" {
				fmt.Fprintf(b, " — %s", ft.Description)
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	section("Racial Traits", f.RacialTraits)
	section("Class Features", f.ClassFeatures)
	section("Feats", f.Feats)
}

func writeSpellcasting(b *strings.Builder, c *models.Character) {
	sc := c.Spellcasting
	if sc == nil {
		return
	}
	b.WriteString("## Spellcasting\n\n")
	ability := abilityAbbrev[sc.Ability]
	fmt.Fprintf(b, "- **Spellcasting Ability:** %s\n", ability)
	fmt.Fprintf(b, "- **Spell Save DC:** %d\n", c.GetSpellSaveDC())
	fmt.Fprintf(b, "- **Spell Attack Bonus:** %s\n", signed(c.GetSpellAttackBonus()))

	// Spell slots.
	var slotParts []string
	for lvl := 1; lvl <= 9; lvl++ {
		if slot := sc.SpellSlots.GetSlot(lvl); slot != nil && slot.Total > 0 {
			slotParts = append(slotParts, fmt.Sprintf("L%d: %d/%d", lvl, slot.Remaining, slot.Total))
		}
	}
	if sc.PactMagic != nil && sc.PactMagic.Total > 0 {
		slotParts = append(slotParts, fmt.Sprintf("Pact (L%d): %d/%d", sc.PactMagic.SlotLevel, sc.PactMagic.Remaining, sc.PactMagic.Total))
	}
	if len(slotParts) > 0 {
		fmt.Fprintf(b, "- **Spell Slots:** %s\n", strings.Join(slotParts, ", "))
	}
	b.WriteString("\n")

	if len(sc.CantripsKnown) > 0 {
		cantrips := append([]string(nil), sc.CantripsKnown...)
		sort.Strings(cantrips)
		fmt.Fprintf(b, "**Cantrips:** %s\n\n", strings.Join(cantrips, ", "))
	}

	// Group leveled spells.
	byLevel := map[int][]models.KnownSpell{}
	maxLvl := 0
	for _, sp := range sc.KnownSpells {
		if sp.Level == 0 {
			continue
		}
		byLevel[sp.Level] = append(byLevel[sp.Level], sp)
		if sp.Level > maxLvl {
			maxLvl = sp.Level
		}
	}
	for lvl := 1; lvl <= maxLvl; lvl++ {
		spells := byLevel[lvl]
		if len(spells) == 0 {
			continue
		}
		sort.Slice(spells, func(i, j int) bool { return spells[i].Name < spells[j].Name })
		fmt.Fprintf(b, "**Level %d:**\n", lvl)
		for _, sp := range spells {
			marker := ""
			if sp.Prepared || sp.AlwaysPrepared {
				marker = " *(prepared)*"
			}
			if sp.Ritual {
				marker += " *(ritual)*"
			}
			fmt.Fprintf(b, "- %s%s\n", sp.Name, marker)
		}
		b.WriteString("\n")
	}
}

func writeInventory(b *strings.Builder, c *models.Character) {
	inv := c.Inventory
	b.WriteString("## Inventory\n\n")

	// Equipped items.
	eq := inv.Equipment
	var equipped []string
	if eq.MainHand != nil {
		equipped = append(equipped, "Main Hand: "+eq.MainHand.Name)
	}
	if eq.OffHand != nil {
		equipped = append(equipped, "Off Hand: "+eq.OffHand.Name)
	}
	if eq.Body != nil {
		equipped = append(equipped, "Armor: "+eq.Body.Name)
	}
	if len(equipped) > 0 {
		fmt.Fprintf(b, "**Equipped:** %s\n\n", strings.Join(equipped, " • "))
	}

	if len(inv.Items) > 0 {
		b.WriteString("| Item | Qty | Weight |\n|------|:---:|:------:|\n")
		for _, it := range inv.Items {
			fmt.Fprintf(b, "| %s | %d | %s |\n", it.Name, it.Quantity, formatWeight(it.Weight*float64(it.Quantity)))
		}
		b.WriteString("\n")
	}

	// Weight & encumbrance.
	load := c.Encumbrance().String()
	fmt.Fprintf(b, "**Carried Weight:** %s / %s lb (%s)\n\n",
		formatWeight(c.CarriedWeight()), formatWeight(c.CarryingCapacity()), load)

	// Currency.
	cur := inv.Currency
	var coins []string
	if cur.Platinum > 0 {
		coins = append(coins, fmt.Sprintf("%d pp", cur.Platinum))
	}
	if cur.Gold > 0 {
		coins = append(coins, fmt.Sprintf("%d gp", cur.Gold))
	}
	if cur.Electrum > 0 {
		coins = append(coins, fmt.Sprintf("%d ep", cur.Electrum))
	}
	if cur.Silver > 0 {
		coins = append(coins, fmt.Sprintf("%d sp", cur.Silver))
	}
	if cur.Copper > 0 {
		coins = append(coins, fmt.Sprintf("%d cp", cur.Copper))
	}
	if len(coins) > 0 {
		fmt.Fprintf(b, "**Currency:** %s\n\n", strings.Join(coins, ", "))
	}
}

func formatWeight(w float64) string {
	if w == float64(int(w)) {
		return fmt.Sprintf("%d", int(w))
	}
	return fmt.Sprintf("%.1f", w)
}

func writePersonality(b *strings.Builder, c *models.Character) {
	p := c.Personality
	if len(p.Traits) == 0 && len(p.Ideals) == 0 && len(p.Bonds) == 0 && len(p.Flaws) == 0 && p.Backstory == "" {
		return
	}
	b.WriteString("## Personality\n\n")
	writeList := func(label string, items []string) {
		if len(items) > 0 {
			fmt.Fprintf(b, "- **%s:** %s\n", label, strings.Join(items, "; "))
		}
	}
	writeList("Traits", p.Traits)
	writeList("Ideals", p.Ideals)
	writeList("Bonds", p.Bonds)
	writeList("Flaws", p.Flaws)
	if p.Backstory != "" {
		fmt.Fprintf(b, "\n### Backstory\n\n%s\n", p.Backstory)
	}
	b.WriteString("\n")
}

func writeCompanions(b *strings.Builder, c *models.Character) {
	if len(c.Companions) == 0 {
		return
	}
	b.WriteString("## Companions & Summons\n\n")
	for i := range c.Companions {
		comp := &c.Companions[i]
		fmt.Fprintf(b, "### %s\n\n", comp.Name)

		descr := string(comp.Kind)
		if st := strings.TrimSpace(comp.Size + " " + comp.Type); st != "" {
			descr += " · " + st
		}
		fmt.Fprintf(b, "*%s*\n\n", descr)

		fmt.Fprintf(b, "- **AC:** %d\n", comp.AC)
		hp := fmt.Sprintf("%d/%d", comp.CurrentHP, comp.MaxHP)
		if comp.TempHP > 0 {
			hp += fmt.Sprintf(" (+%d temp)", comp.TempHP)
		}
		fmt.Fprintf(b, "- **HP:** %s\n", hp)
		if comp.Speed != "" {
			fmt.Fprintf(b, "- **Speed:** %s\n", comp.Speed)
		}

		var abils []string
		for j := 0; j < 6; j++ {
			abils = append(abils, fmt.Sprintf("%s %d (%s)",
				models.CompanionAbilityLabel(j), comp.Abilities[j],
				models.FormatModifier(comp.Modifier(j))))
		}
		fmt.Fprintf(b, "- **Abilities:** %s\n", strings.Join(abils, ", "))

		if len(comp.Attacks) > 0 {
			b.WriteString("\n**Attacks:**\n\n")
			for _, a := range comp.Attacks {
				line := fmt.Sprintf("- %s: %s to hit", a.Name, models.FormatModifier(a.Bonus))
				if a.Damage != "" {
					line += fmt.Sprintf(", %s", a.Damage)
				}
				b.WriteString(line + "\n")
			}
		}

		if len(comp.Traits) > 0 {
			b.WriteString("\n**Traits:**\n\n")
			for _, t := range comp.Traits {
				if t.Text != "" {
					fmt.Fprintf(b, "- **%s.** %s\n", t.Name, t.Text)
				} else {
					fmt.Fprintf(b, "- %s\n", t.Name)
				}
			}
		}

		if comp.Notes != "" {
			fmt.Fprintf(b, "\n%s\n", comp.Notes)
		}
		b.WriteString("\n")
	}
}
