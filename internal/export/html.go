package export

import (
	"fmt"
	"html"
	"sort"
	"strings"

	"github.com/Domo929/sheet/internal/models"
)

// esc HTML-escapes a dynamic string for safe embedding.
func esc(s string) string { return html.EscapeString(s) }

// printCSS is the embedded stylesheet for the printable sheet. It targets
// Letter paper with print-friendly colors and avoids splitting cards/sections
// across page breaks.
const printCSS = `
@page { size: Letter; margin: 0.5in; }
* { box-sizing: border-box; }
body { font-family: 'Segoe UI', Helvetica, Arial, sans-serif; color: #1a1a1a;
  font-size: 12px; line-height: 1.4; max-width: 8in; margin: 0 auto; padding: 16px; }
h1 { font-size: 26px; margin: 0 0 2px; }
.subtitle { font-size: 14px; color: #333; margin-bottom: 2px; }
.meta { color: #666; font-size: 11px; margin-bottom: 12px; }
h2 { font-size: 15px; border-bottom: 2px solid #7c3f00; padding-bottom: 2px;
  margin: 16px 0 8px; color: #7c3f00; }
h3 { font-size: 13px; margin: 10px 0 4px; }
table { border-collapse: collapse; width: 100%; margin-bottom: 8px; }
th, td { border: 1px solid #cbb; padding: 3px 6px; text-align: left; }
th { background: #f3ede4; }
td.c, th.c { text-align: center; }
.cards { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 8px; }
.card { border: 1px solid #cbb; border-radius: 6px; padding: 6px 10px;
  min-width: 92px; text-align: center; }
.card .label { font-size: 10px; text-transform: uppercase; color: #666; }
.card .value { font-size: 18px; font-weight: bold; }
ul { margin: 4px 0 8px; padding-left: 18px; }
li { margin: 2px 0; }
.section { break-inside: avoid; page-break-inside: avoid; }
.muted { color: #666; }
@media print { body { padding: 0; } }
`

// ToHTML renders a complete, print-ready character sheet as a self-contained
// HTML document. Open it in a browser and use Print (or Save as PDF) to produce
// a printable PDF; no external assets or fonts are required.
func ToHTML(c *models.Character) string {
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	b.WriteString("<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	fmt.Fprintf(&b, "<title>%s — Character Sheet</title>\n", esc(c.Info.Name))
	fmt.Fprintf(&b, "<style>%s</style>\n", printCSS)
	b.WriteString("</head>\n<body>\n")

	htmlHeader(&b, c)
	htmlCombat(&b, c)
	htmlAbilities(&b, c)
	htmlSkills(&b, c)
	htmlProficiencies(&b, c)
	htmlFeatures(&b, c)
	htmlSpellcasting(&b, c)
	htmlResources(&b, c)
	htmlInventory(&b, c)
	htmlCompanions(&b, c)
	htmlPersonality(&b, c)

	b.WriteString("</body>\n</html>\n")
	return b.String()
}

func htmlHeader(b *strings.Builder, c *models.Character) {
	info := c.Info
	fmt.Fprintf(b, "<h1>%s</h1>\n", esc(info.Name))

	race := info.Race
	if info.Subrace != "" {
		race = fmt.Sprintf("%s (%s)", info.Race, info.Subrace)
	}
	class := info.Class
	if info.Subclass != "" {
		class = fmt.Sprintf("%s (%s)", info.Class, info.Subclass)
	}
	if c.IsMulticlass() {
		fmt.Fprintf(b, "<div class=\"subtitle\">Level %d %s — %s</div>\n",
			c.TotalLevel(), esc(race), esc(c.ClassSummary()))
	} else {
		fmt.Fprintf(b, "<div class=\"subtitle\">Level %d %s %s</div>\n",
			info.Level, esc(race), esc(class))
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
		fmt.Fprintf(b, "<div class=\"meta\">%s</div>\n", esc(strings.Join(meta, " • ")))
	}
}

func htmlCard(b *strings.Builder, label, value string) {
	fmt.Fprintf(b, "<div class=\"card\"><div class=\"label\">%s</div><div class=\"value\">%s</div></div>\n",
		esc(label), esc(value))
}

func htmlCombat(b *strings.Builder, c *models.Character) {
	hp := c.CombatStats.HitPoints
	hpStr := fmt.Sprintf("%d / %d", hp.Current, hp.Maximum)
	if hp.Temporary > 0 {
		hpStr += fmt.Sprintf(" +%d", hp.Temporary)
	}

	b.WriteString("<div class=\"section\"><h2>Combat</h2>\n<div class=\"cards\">\n")
	htmlCard(b, "Armor Class", fmt.Sprintf("%d", c.CombatStats.ArmorClass))
	htmlCard(b, "Hit Points", hpStr)
	htmlCard(b, "Initiative", signed(c.GetInitiative()))
	htmlCard(b, "Speed", speedString(c))
	htmlCard(b, "Prof. Bonus", signed(c.GetProficiencyBonus()))
	htmlCard(b, "Passive Perc.", fmt.Sprintf("%d", 10+c.GetSkillModifier(models.SkillPerception)))
	htmlCard(b, "Hit Dice", fmt.Sprintf("%d/%d d%d",
		c.CombatStats.HitDice.Remaining, c.CombatStats.HitDice.Total, c.CombatStats.HitDice.DieType))
	b.WriteString("</div>\n")

	var extra []string
	if c.CombatStats.ExhaustionLevel > 0 {
		extra = append(extra, fmt.Sprintf("Exhaustion Level %d", c.CombatStats.ExhaustionLevel))
	}
	if senses := c.CombatStats.Senses.List(); len(senses) > 0 {
		extra = append(extra, "Senses: "+strings.Join(senses, ", "))
	}
	if len(c.CombatStats.Conditions) > 0 {
		conds := make([]string, len(c.CombatStats.Conditions))
		for i, cond := range c.CombatStats.Conditions {
			conds[i] = string(cond)
		}
		extra = append(extra, "Conditions: "+strings.Join(conds, ", "))
	}
	if len(extra) > 0 {
		fmt.Fprintf(b, "<div class=\"meta\">%s</div>\n", esc(strings.Join(extra, " • ")))
	}
	b.WriteString("</div>\n")
}

func htmlAbilities(b *strings.Builder, c *models.Character) {
	b.WriteString("<div class=\"section\"><h2>Ability Scores</h2>\n<table>\n")
	b.WriteString("<tr><th>Ability</th><th class=\"c\">Score</th><th class=\"c\">Mod</th><th class=\"c\">Save</th></tr>\n")
	for _, a := range abilityOrder {
		score := c.AbilityScores.Get(a.Ability)
		save := signed(c.GetSavingThrowModifier(a.Ability))
		if c.SavingThrows.IsProficient(a.Ability) {
			save += " ●"
		}
		fmt.Fprintf(b, "<tr><td>%s</td><td class=\"c\">%d</td><td class=\"c\">%s</td><td class=\"c\">%s</td></tr>\n",
			esc(a.Label), score.Total(), esc(signed(score.Modifier())), esc(save))
	}
	b.WriteString("</table></div>\n")
}

func htmlSkills(b *strings.Builder, c *models.Character) {
	b.WriteString("<div class=\"section\"><h2>Skills</h2>\n<table>\n")
	b.WriteString("<tr><th>Skill</th><th class=\"c\">Ability</th><th class=\"c\">Bonus</th><th class=\"c\">Prof</th></tr>\n")
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
			prof = "●● exp"
		}
		fmt.Fprintf(b, "<tr><td>%s</td><td class=\"c\">%s</td><td class=\"c\">%s</td><td class=\"c\">%s</td></tr>\n",
			esc(display), esc(abbr), esc(signed(c.GetSkillModifier(name))), esc(prof))
	}
	b.WriteString("</table></div>\n")
}

func htmlProficiencies(b *strings.Builder, c *models.Character) {
	p := c.Proficiencies
	if len(p.Armor) == 0 && len(p.Weapons) == 0 && len(p.Tools) == 0 && len(p.Languages) == 0 {
		return
	}
	b.WriteString("<div class=\"section\"><h2>Proficiencies &amp; Languages</h2>\n<ul>\n")
	item := func(label string, items []string) {
		if len(items) > 0 {
			fmt.Fprintf(b, "<li><strong>%s:</strong> %s</li>\n", esc(label), esc(strings.Join(items, ", ")))
		}
	}
	item("Armor", p.Armor)
	item("Weapons", p.Weapons)
	item("Tools", p.Tools)
	item("Languages", p.Languages)
	b.WriteString("</ul></div>\n")
}

func htmlFeatures(b *strings.Builder, c *models.Character) {
	f := c.Features
	if len(f.RacialTraits) == 0 && len(f.ClassFeatures) == 0 && len(f.Feats) == 0 {
		return
	}
	b.WriteString("<div class=\"section\"><h2>Features &amp; Traits</h2>\n")
	section := func(title string, feats []models.Feature) {
		if len(feats) == 0 {
			return
		}
		fmt.Fprintf(b, "<h3>%s</h3>\n<ul>\n", esc(title))
		for _, ft := range feats {
			line := "<strong>" + esc(ft.Name) + "</strong>"
			if ft.Source != "" {
				line += " <span class=\"muted\">(" + esc(ft.Source) + ")</span>"
			}
			if ft.Description != "" {
				line += " — " + esc(ft.Description)
			}
			fmt.Fprintf(b, "<li>%s</li>\n", line)
		}
		b.WriteString("</ul>\n")
	}
	section("Racial Traits", f.RacialTraits)
	section("Class Features", f.ClassFeatures)
	section("Feats", f.Feats)
	b.WriteString("</div>\n")
}

func htmlSpellcasting(b *strings.Builder, c *models.Character) {
	sc := c.Spellcasting
	if sc == nil {
		return
	}
	b.WriteString("<div class=\"section\"><h2>Spellcasting</h2>\n<div class=\"cards\">\n")
	htmlCard(b, "Ability", abilityAbbrev[sc.Ability])
	htmlCard(b, "Save DC", fmt.Sprintf("%d", c.GetSpellSaveDC()))
	htmlCard(b, "Attack", signed(c.GetSpellAttackBonus()))
	b.WriteString("</div>\n")

	var slotParts []string
	for lvl := 1; lvl <= 9; lvl++ {
		if slot := sc.SpellSlots.GetSlot(lvl); slot != nil && slot.Total > 0 {
			slotParts = append(slotParts, fmt.Sprintf("L%d: %d/%d", lvl, slot.Remaining, slot.Total))
		}
	}
	if sc.PactMagic != nil && sc.PactMagic.Total > 0 {
		slotParts = append(slotParts, fmt.Sprintf("Pact L%d: %d/%d",
			sc.PactMagic.SlotLevel, sc.PactMagic.Remaining, sc.PactMagic.Total))
	}
	if len(slotParts) > 0 {
		fmt.Fprintf(b, "<p><strong>Spell Slots:</strong> %s</p>\n", esc(strings.Join(slotParts, ", ")))
	}

	if len(sc.CantripsKnown) > 0 {
		cantrips := append([]string(nil), sc.CantripsKnown...)
		sort.Strings(cantrips)
		fmt.Fprintf(b, "<p><strong>Cantrips:</strong> %s</p>\n", esc(strings.Join(cantrips, ", ")))
	}

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
		fmt.Fprintf(b, "<h3>Level %d</h3>\n<ul>\n", lvl)
		for _, sp := range spells {
			marker := ""
			if sp.Prepared || sp.AlwaysPrepared {
				marker += " <span class=\"muted\">(prepared)</span>"
			}
			if sp.Ritual {
				marker += " <span class=\"muted\">(ritual)</span>"
			}
			fmt.Fprintf(b, "<li>%s%s</li>\n", esc(sp.Name), marker)
		}
		b.WriteString("</ul>\n")
	}
	b.WriteString("</div>\n")
}

func htmlResources(b *strings.Builder, c *models.Character) {
	if len(c.Resources) == 0 {
		return
	}
	b.WriteString("<div class=\"section\"><h2>Class Resources</h2>\n<table>\n")
	b.WriteString("<tr><th>Resource</th><th class=\"c\">Uses</th><th class=\"c\">Recharge</th></tr>\n")
	for _, p := range c.Resources {
		recharge := "Long Rest"
		if p.Recharge == "short" {
			recharge = "Short Rest"
		} else if p.ShortRestRecovery > 0 {
			recharge = fmt.Sprintf("Long Rest (+%d/Short)", p.ShortRestRecovery)
		}
		fmt.Fprintf(b, "<tr><td>%s</td><td class=\"c\">%d / %d</td><td class=\"c\">%s</td></tr>\n",
			esc(p.Name), p.Current, p.Max, esc(recharge))
	}
	b.WriteString("</table></div>\n")
}

func htmlInventory(b *strings.Builder, c *models.Character) {
	inv := c.Inventory
	b.WriteString("<div class=\"section\"><h2>Inventory</h2>\n")

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
		fmt.Fprintf(b, "<p><strong>Equipped:</strong> %s</p>\n", esc(strings.Join(equipped, " • ")))
	}

	if len(inv.Items) > 0 {
		b.WriteString("<table>\n<tr><th>Item</th><th class=\"c\">Qty</th><th class=\"c\">Weight</th></tr>\n")
		for _, it := range inv.Items {
			name := it.Name
			if it.RequiresAttunement && it.Attuned {
				name += " ✦"
			}
			fmt.Fprintf(b, "<tr><td>%s</td><td class=\"c\">%d</td><td class=\"c\">%s</td></tr>\n",
				esc(name), it.Quantity, esc(formatWeight(it.Weight*float64(it.Quantity))))
		}
		b.WriteString("</table>\n")
	}

	fmt.Fprintf(b, "<p><strong>Carried Weight:</strong> %s / %s lb (%s)</p>\n",
		esc(formatWeight(c.CarriedWeight())), esc(formatWeight(c.CarryingCapacity())), esc(c.Encumbrance().String()))

	cur := inv.Currency
	var coins []string
	for _, d := range []struct {
		n int
		s string
	}{{cur.Platinum, "pp"}, {cur.Gold, "gp"}, {cur.Electrum, "ep"}, {cur.Silver, "sp"}, {cur.Copper, "cp"}} {
		if d.n > 0 {
			coins = append(coins, fmt.Sprintf("%d %s", d.n, d.s))
		}
	}
	if len(coins) > 0 {
		fmt.Fprintf(b, "<p><strong>Currency:</strong> %s</p>\n", esc(strings.Join(coins, ", ")))
	}
	b.WriteString("</div>\n")
}

func htmlCompanions(b *strings.Builder, c *models.Character) {
	if len(c.Companions) == 0 {
		return
	}
	b.WriteString("<div class=\"section\"><h2>Companions &amp; Summons</h2>\n")
	for i := range c.Companions {
		comp := &c.Companions[i]
		fmt.Fprintf(b, "<h3>%s</h3>\n", esc(comp.Name))
		descr := string(comp.Kind)
		if st := strings.TrimSpace(comp.Size + " " + comp.Type); st != "" {
			descr += " · " + st
		}
		fmt.Fprintf(b, "<div class=\"meta\">%s</div>\n<ul>\n", esc(descr))
		fmt.Fprintf(b, "<li><strong>AC:</strong> %d</li>\n", comp.AC)
		hp := fmt.Sprintf("%d/%d", comp.CurrentHP, comp.MaxHP)
		if comp.TempHP > 0 {
			hp += fmt.Sprintf(" (+%d temp)", comp.TempHP)
		}
		fmt.Fprintf(b, "<li><strong>HP:</strong> %s</li>\n", esc(hp))
		if comp.Speed != "" {
			fmt.Fprintf(b, "<li><strong>Speed:</strong> %s</li>\n", esc(comp.Speed))
		}
		var abils []string
		for j := 0; j < 6; j++ {
			abils = append(abils, fmt.Sprintf("%s %d (%s)",
				models.CompanionAbilityLabel(j), comp.Abilities[j], models.FormatModifier(comp.Modifier(j))))
		}
		fmt.Fprintf(b, "<li><strong>Abilities:</strong> %s</li>\n", esc(strings.Join(abils, ", ")))
		b.WriteString("</ul>\n")

		if len(comp.Attacks) > 0 {
			b.WriteString("<p><strong>Attacks:</strong></p>\n<ul>\n")
			for _, a := range comp.Attacks {
				line := fmt.Sprintf("%s: %s to hit", a.Name, models.FormatModifier(a.Bonus))
				if a.Damage != "" {
					line += ", " + a.Damage
				}
				fmt.Fprintf(b, "<li>%s</li>\n", esc(line))
			}
			b.WriteString("</ul>\n")
		}
		if len(comp.Traits) > 0 {
			b.WriteString("<p><strong>Traits:</strong></p>\n<ul>\n")
			for _, t := range comp.Traits {
				if t.Text != "" {
					fmt.Fprintf(b, "<li><strong>%s.</strong> %s</li>\n", esc(t.Name), esc(t.Text))
				} else {
					fmt.Fprintf(b, "<li>%s</li>\n", esc(t.Name))
				}
			}
			b.WriteString("</ul>\n")
		}
		if comp.Notes != "" {
			fmt.Fprintf(b, "<p>%s</p>\n", esc(comp.Notes))
		}
	}
	b.WriteString("</div>\n")
}

func htmlPersonality(b *strings.Builder, c *models.Character) {
	p := c.Personality
	if len(p.Traits) == 0 && len(p.Ideals) == 0 && len(p.Bonds) == 0 && len(p.Flaws) == 0 && p.Backstory == "" {
		return
	}
	b.WriteString("<div class=\"section\"><h2>Personality</h2>\n<ul>\n")
	item := func(label string, items []string) {
		if len(items) > 0 {
			fmt.Fprintf(b, "<li><strong>%s:</strong> %s</li>\n", esc(label), esc(strings.Join(items, "; ")))
		}
	}
	item("Traits", p.Traits)
	item("Ideals", p.Ideals)
	item("Bonds", p.Bonds)
	item("Flaws", p.Flaws)
	b.WriteString("</ul>\n")
	if p.Backstory != "" {
		fmt.Fprintf(b, "<h3>Backstory</h3>\n<p>%s</p>\n", esc(p.Backstory))
	}
	b.WriteString("</div>\n")
}
