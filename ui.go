package main

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// --- Color Palette (btop-inspired: dim chrome, bright data) ---

var (
	colorBorder   = lipgloss.Color("#555566") // dim border
	colorTitle    = lipgloss.Color("#e94560") // panel titles
	colorLabel    = lipgloss.Color("#888899") // secondary text
	colorValue    = lipgloss.Color("#eaeaea") // primary data
	colorAccent   = lipgloss.Color("#e94560") // highlights, cursor
	colorMuted    = lipgloss.Color("#666677") // de-emphasized
	colorGood     = lipgloss.Color("#53d769") // positive signals
	colorSelected = lipgloss.Color("#ffc107") // selection marker
	colorTrack    = lipgloss.Color("#333344") // unfilled bar track
)

// --- Styles ---

var (
	styleLabel = lipgloss.NewStyle().
			Foreground(colorLabel)

	styleValue = lipgloss.NewStyle().
			Foreground(colorValue)

	styleValueBold = lipgloss.NewStyle().
			Foreground(colorValue).
			Bold(true)

	styleMuted = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleGood = lipgloss.NewStyle().
			Foreground(colorGood).
			Bold(true)

	styleHelp = lipgloss.NewStyle().
			Foreground(colorMuted)
)

// degreeColors for degree breakdown (best-for-caster to worst)
var degreeColors = [4]lipgloss.Color{
	lipgloss.Color("#53d769"), // Best: bright green
	lipgloss.Color("#4ecdc4"), // Good: teal
	lipgloss.Color("#ffc107"), // Bad: amber
	lipgloss.Color("#e94560"), // Worst: red
}

// degreeColorsDim — muted versions for bars (btop style: bars are subtle, values are bright)
var degreeColorsDim = [4]lipgloss.Color{
	lipgloss.Color("#2a6b35"), // dim green
	lipgloss.Color("#286b62"), // dim teal
	lipgloss.Color("#6b5a15"), // dim amber
	lipgloss.Color("#6b2535"), // dim red
}

// chartGradient for braille histogram — bottom (dim) to top (bright)
var chartGradient = []lipgloss.Color{
	lipgloss.Color("#884466"),
	lipgloss.Color("#bb5566"),
	lipgloss.Color("#dd5566"),
	lipgloss.Color("#e94560"),
	lipgloss.Color("#ff6680"),
}

// barGradient goes from green (best) to red (worst relative performance).
var barGradient = []lipgloss.Color{
	lipgloss.Color("#53d769"),
	lipgloss.Color("#8cd769"),
	lipgloss.Color("#c7d769"),
	lipgloss.Color("#d7c069"),
	lipgloss.Color("#d79b69"),
	lipgloss.Color("#d76969"),
	lipgloss.Color("#e94560"),
}

// Save profile colors (from caster's perspective: low save = easy = green)
var profileColors = [3]lipgloss.Color{
	lipgloss.Color("#53d769"), // Low save → easy for caster
	lipgloss.Color("#888899"), // Med → neutral
	lipgloss.Color("#e94560"), // High save → hard for caster
}

// ============================================================================
// Panel Rendering — btop-style title-in-border
// ============================================================================

func (m model) renderPanel(title string, w int, content string) string {
	body := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderTop(false).
		BorderForeground(colorBorder).
		Padding(0, 1).
		Width(w).
		Render(content)

	outerW := lipgloss.Width(body)

	borderFg := lipgloss.NewStyle().Foreground(colorBorder)
	titleFg := lipgloss.NewStyle().Foreground(colorTitle).Bold(true)

	repeatCount := outerW - 2
	if repeatCount < 0 {
		repeatCount = 0
	}

	var top string
	if title == "" {
		top = borderFg.Render("╭" + strings.Repeat("─", repeatCount) + "╮")
	} else {
		titleStr := titleFg.Render(title)
		titleW := lipgloss.Width(titleStr)
		dashes := outerW - titleW - 5
		if dashes < 1 {
			dashes = 1
		}
		top = borderFg.Render("╭─") + " " + titleStr + " " +
			borderFg.Render(strings.Repeat("─", dashes)+"╮")
	}

	return top + "\n" + body
}

// ============================================================================
// Dynamic Layout
// ============================================================================

// leftWidth returns the outer width for the left column (encounter + spells + comparison).
// Scales with terminal width instead of being capped at a fixed size.
func (m model) leftWidth() int {
	w := m.windowWidth * 55 / 100
	if w < 56 {
		w = 56
	}
	// Leave at least 38 for detail pane + margin
	maxLeft := m.windowWidth - 38
	if maxLeft < 56 {
		maxLeft = 56
	}
	if w > maxLeft {
		w = maxLeft
	}
	return w
}

func (m model) detailPaneWidth() int {
	remaining := m.windowWidth - m.leftWidth() - 3
	if remaining < 34 {
		remaining = 34
	}
	return remaining
}

// leftContentWidth returns the usable width inside a left-column panel (after border + padding).
func (m model) leftContentWidth() int {
	return m.leftWidth() - 4
}

// ============================================================================
// Main View — Single Page Layout
// ============================================================================

func (m model) viewList() string {
	help := m.renderListHelp()
	helpH := lipgloss.Height(help)

	bodyH := m.windowHeight - helpH
	if bodyH < 10 {
		bodyH = 10
	}

	encPanel := m.renderEncounterPanel()
	encHeight := lipgloss.Height(encPanel)

	var compSection string
	compHeight := 0
	if len(m.selected) >= 2 {
		compSection = m.renderComparison()
		compHeight = lipgloss.Height(compSection)
	}

	flashSection := ""
	flashHeight := 0
	if m.flashText != "" {
		flashSection = lipgloss.NewStyle().
			Foreground(colorGood).Bold(true).
			Render("  " + m.flashText)
		flashHeight = 1
	}

	var body string

	if len(m.spells) > 0 {
		fixedHeight := encHeight + compHeight + flashHeight
		spellListRows := bodyH - fixedHeight - 2 // -2 for spell panel borders
		if spellListRows < 5 {
			spellListRows = 5
		}

		spellList := m.renderSpellListFixed(spellListRows)

		var leftParts []string
		leftParts = append(leftParts, encPanel)
		leftParts = append(leftParts, spellList)
		if compSection != "" {
			leftParts = append(leftParts, compSection)
		}
		if flashSection != "" {
			leftParts = append(leftParts, flashSection)
		}
		leftCol := lipgloss.JoinVertical(lipgloss.Left, leftParts...)

		paneWidth := m.detailPaneWidth()
		detailContent := m.renderDetailContent()

		leftHeight := lipgloss.Height(leftCol)
		detailPanel := m.renderPanel("Detail", paneWidth, detailContent)
		detailHeight := lipgloss.Height(detailPanel)
		if detailHeight < leftHeight {
			extra := leftHeight - detailHeight
			detailPanel = m.renderPanel("Detail", paneWidth, detailContent+strings.Repeat("\n", extra))
		}

		rightCol := lipgloss.NewStyle().MarginLeft(1).Render(detailPanel)
		body = lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)
	} else {
		spellListRows := bodyH - encHeight - 2 // -2 for spell panel border chrome
		if spellListRows < 5 {
			spellListRows = 5
		}
		spellList := m.renderSpellListFixed(spellListRows)
		body = lipgloss.JoinVertical(lipgloss.Left, encPanel, spellList)
	}

	// Ensure total output fits exactly in the window (ANSI-safe via lipgloss).
	body = lipgloss.NewStyle().MaxWidth(m.windowWidth).Render(body)

	bodyHeight := lipgloss.Height(body)
	if bodyHeight < bodyH {
		body += strings.Repeat("\n", bodyH-bodyHeight)
	} else if bodyHeight > bodyH {
		body = lipgloss.NewStyle().MaxHeight(bodyH).Render(body)
	}

	return lipgloss.JoinVertical(lipgloss.Left, body, help)
}

// ============================================================================
// Encounter Panel
// ============================================================================

func (m model) renderEncounterPanel() string {
	contentW := m.leftContentWidth()

	renderField := func(idx int, label string, extraWidth int) string {
		focused := m.encFocus == idx
		paddedLabel := fmt.Sprintf("%-*s", 7+extraWidth, label+":")

		indicator := "  "
		labelStyle := styleLabel
		if focused {
			indicator = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("> ")
			labelStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
		}
		return indicator + labelStyle.Render(paddedLabel) + m.encInputs[idx].View()
	}

	renderSaveField := func(idx int, label string, profile SaveProfile) string {
		focused := m.encFocus == idx
		profStyle := lipgloss.NewStyle().Foreground(profileColors[profile])
		profLabel := profStyle.Render(fmt.Sprintf("%-4s", profile.String()))

		paddedLabel := fmt.Sprintf("%-7s", label+":")
		indicator := "  "
		labelStyle := styleLabel
		if focused {
			indicator = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("> ")
			labelStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
		}
		hint := ""
		if focused {
			hint = styleMuted.Render(" ␣:cycle")
		}
		return indicator + labelStyle.Render(paddedLabel) + profLabel + " " + m.encInputs[idx].View() + hint
	}

	// You column
	var leftLines []string
	leftLines = append(leftLines, styleLabel.Render("  You"))
	leftLines = append(leftLines, renderField(encPCLevel, "Level", 0))
	leftLines = append(leftLines, renderField(encSpellDC, "DC", 0))
	leftLines = append(leftLines, renderField(encAttackMod, "Atk", 0))
	leftCol := strings.Join(leftLines, "\n")

	// Enemy column
	var rightLines []string
	rightLines = append(rightLines, styleLabel.Render("  Enemy"))
	rightLines = append(rightLines, renderField(encEnemyAC, "AC", 0))
	rightLines = append(rightLines, renderSaveField(encRefMod, "Ref", m.encounter.RefProfile))
	rightLines = append(rightLines, renderSaveField(encFortMod, "Fort", m.encounter.FortProfile))
	rightLines = append(rightLines, renderSaveField(encWillMod, "Will", m.encounter.WillProfile))
	rightCol := strings.Join(rightLines, "\n")

	colWidth := contentW / 2
	if colWidth < 28 {
		colWidth = 28
	}
	leftSized := lipgloss.NewStyle().Width(colWidth).Render(leftCol)
	grid := lipgloss.JoinHorizontal(lipgloss.Top, leftSized, rightCol)

	return m.renderPanel("Encounter", m.leftWidth(), grid)
}

// ============================================================================
// Spell List Panel
// ============================================================================

func (m model) renderSpellListFixed(maxRows int) string {
	w := m.leftWidth()
	contentW := m.leftContentWidth()

	if len(m.spells) == 0 {
		empty := styleMuted.Render("  No spells yet. Press  a  to add from presets,  n  for custom.")
		content := empty
		padLines := maxRows - 1
		if padLines > 0 {
			content += strings.Repeat("\n", padLines)
		}
		return m.renderPanel("Spells", w, content)
	}

	maxDmg := 0.0
	for _, e := range m.spells {
		if e.spell.Dice.Valid() && e.stats.ExpectedDamage > maxDmg {
			maxDmg = e.stats.ExpectedDamage
		}
	}

	// Dynamic sparkline width based on available space
	sparkWidth := 16
	// Name gets remaining space after spark + dice + damage + prefixes
	nameWidth := contentW - sparkWidth - 16
	if nameWidth < 10 {
		nameWidth = 10
	}
	if nameWidth > 24 {
		nameWidth = 24
	}

	var allLines []string
	for i, e := range m.spells {
		allLines = append(allLines, m.renderSpellRow(i, &e, nameWidth, sparkWidth, maxDmg))
	}

	visibleRows := maxRows - 1
	if visibleRows < 3 {
		visibleRows = 3
	}

	startIdx := 0
	if len(allLines) > visibleRows {
		startIdx = m.cursor - visibleRows/2
		if startIdx < 0 {
			startIdx = 0
		}
		if startIdx+visibleRows > len(allLines) {
			startIdx = len(allLines) - visibleRows
		}
	}

	endIdx := startIdx + visibleRows
	if endIdx > len(allLines) {
		endIdx = len(allLines)
	}

	var lines []string

	if startIdx > 0 {
		lines = append(lines, styleMuted.Render(fmt.Sprintf("  ↑ %d more", startIdx)))
	}

	lines = append(lines, allLines[startIdx:endIdx]...)

	remaining := len(allLines) - endIdx
	if remaining > 0 {
		lines = append(lines, styleMuted.Render(fmt.Sprintf("  ↓ %d more", remaining)))
	}

	for len(lines) < maxRows {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return m.renderPanel("Spells", w, content)
}

func (m model) renderSpellRow(i int, e *spellEntry, nameWidth, sparkWidth int, maxDmg float64) string {
	isCursor := i == m.cursor
	isSel := m.selected[e.id]

	prefix := "  "
	if isCursor && isSel {
		prefix = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render(">*")
	} else if isCursor {
		prefix = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("> ")
	} else if isSel {
		prefix = lipgloss.NewStyle().Foreground(colorSelected).Render(" *")
	}

	name := e.spell.Name
	if name == "" {
		name = fmt.Sprintf("Spell %d", i+1)
	}
	nameRunes := []rune(name)
	if len(nameRunes) > nameWidth {
		name = string(nameRunes[:nameWidth-3]) + "..."
	}

	nameStyle := styleValue
	if isCursor {
		nameStyle = lipgloss.NewStyle().Foreground(colorValue).Bold(true)
	}

	paddedName := fmt.Sprintf("%-*s", nameWidth, name)
	paddedDice := fmt.Sprintf("%6s", e.spell.Dice.String())

	var spark string
	if e.spell.Dice.Valid() && len(e.stats.MixturePDF) > 0 {
		sparkRaw := RenderBrailleSparkline(e.stats.MixturePDF, sparkWidth)
		ratio := 0.0
		if maxDmg > 0 {
			ratio = e.stats.ExpectedDamage / maxDmg
		}
		colorIdx := int((1.0 - ratio) * float64(len(barGradient)-1))
		if colorIdx < 0 {
			colorIdx = 0
		}
		if colorIdx >= len(barGradient) {
			colorIdx = len(barGradient) - 1
		}
		spark = lipgloss.NewStyle().Foreground(barGradient[colorIdx]).Render(sparkRaw)
	} else {
		spark = strings.Repeat(" ", sparkWidth)
	}

	var dmgStr string
	if e.spell.Dice.Valid() {
		dmgStr = fmt.Sprintf("~%.0f", e.stats.ExpectedDamage)
	}

	return prefix + " " +
		nameStyle.Render(paddedName) + " " +
		styleMuted.Render(paddedDice) + " " +
		spark + " " +
		styleValue.Render(fmt.Sprintf("%4s", dmgStr))
}

// ============================================================================
// Detail Pane
// ============================================================================

func (m model) renderDetailContent() string {
	entry := &m.spells[m.cursor]
	sp := &entry.spell
	st := &entry.stats
	labels := DegreeLabels(sp.Type)
	innerWidth := m.detailPaneWidth() - 4

	var lines []string

	castRank := entry.effectiveCastRank()

	// Spell name + metadata
	namePart := styleValueBold.Render(sp.Name)
	var metaParts []string
	if sp.BaseRank > 0 {
		if castRank > sp.BaseRank {
			metaParts = append(metaParts, styleGood.Render(fmt.Sprintf("R%d", castRank))+
				styleMuted.Render(fmt.Sprintf("/%d", sp.BaseRank)))
		} else {
			metaParts = append(metaParts, styleMuted.Render(fmt.Sprintf("R%d", sp.BaseRank)))
		}
	}
	metaParts = append(metaParts, styleMuted.Render(sp.ShortDefenseLabel()))
	lines = append(lines, namePart+"  "+strings.Join(metaParts, styleMuted.Render(" · ")))

	effectiveDice := sp.Dice
	if castRank > 0 {
		effectiveDice = sp.EffectiveDice(castRank)
	}
	diceInfo := styleValueBold.Render(effectiveDice.String())
	if sp.HeightenDie > 0 {
		step := sp.effectiveHeightenStep()
		if step > 1 {
			diceInfo += styleMuted.Render(fmt.Sprintf(" +%dd%d/%d ranks", sp.HeightenDie, sp.Dice.Sides, step))
		} else {
			diceInfo += styleMuted.Render(fmt.Sprintf(" +%dd%d/rank", sp.HeightenDie, sp.Dice.Sides))
		}
	}
	if sp.Type == SpellTypeSave {
		saveMod := m.encounter.SaveModFor(sp.SaveType)
		diceInfo += "  " + styleMuted.Render(fmt.Sprintf("vs %s +%d (DC %d)", sp.SaveType, saveMod, m.encounter.SpellDC))
	} else {
		diceInfo += "  " + styleMuted.Render(fmt.Sprintf("+%d vs AC %d", m.encounter.AttackMod, m.encounter.EnemyAC))
	}
	lines = append(lines, diceInfo)

	if !effectiveDice.Valid() {
		lines = append(lines, styleMuted.Render("Invalid dice formula"))
		return strings.Join(lines, "\n")
	}

	// Degree breakdown — btop-style thin bars with track
	probs := st.DegreeProb
	mults := sp.Multipliers.AsSlice()
	maxProb := 0.0
	for _, p := range probs {
		if p > maxProb {
			maxProb = p
		}
	}

	probBarWidth := innerWidth - 28
	if probBarWidth < 6 {
		probBarWidth = 6
	}
	if probBarWidth > 16 {
		probBarWidth = 16
	}

	trackStyle := lipgloss.NewStyle().Foreground(colorTrack)

	for i := 0; i < 4; i++ {
		pct := probs[i] * 100
		barStyle := lipgloss.NewStyle().Foreground(degreeColorsDim[i])

		barLen := 0
		if maxProb > 0 {
			barLen = int(math.Round(probs[i] / maxProb * float64(probBarWidth)))
		}
		bar := barStyle.Render(strings.Repeat("━", barLen)) +
			trackStyle.Render(strings.Repeat("━", probBarWidth-barLen))

		multStr := fmt.Sprintf("%.0fx", mults[i])
		if mults[i] == 0.5 {
			multStr = ".5x"
		}

		dmgStr := "  —"
		if mults[i] > 0 {
			dmgStr = fmt.Sprintf("%3s", fmt.Sprintf("~%.0f", st.DegreeMean[i]))
		}

		paddedLabel := fmt.Sprintf("%-13s", labels[i])
		degStyle := lipgloss.NewStyle().Foreground(degreeColors[i])

		line := paddedLabel +
			styleMuted.Render(fmt.Sprintf("%3s", multStr)) +
			fmt.Sprintf(" %3.0f%% ", pct) +
			bar + " " +
			degStyle.Render(fmt.Sprintf("%4s", dmgStr))
		lines = append(lines, line)
	}

	// Summary
	lines = append(lines, "")
	lines = append(lines, styleGood.Render(fmt.Sprintf("~%.0f damage on average", st.ExpectedDamage)))
	lines = append(lines, styleMuted.Render(fmt.Sprintf("  typically %.0f–%.0f",
		math.Max(0, st.ExpectedDamage-st.OverallStdDev),
		st.ExpectedDamage+st.OverallStdDev)))
	lines = append(lines, styleMuted.Render(fmt.Sprintf("  %.0f%% chance to deal damage", st.AnyDamageProb*100)))
	lines = append(lines, "")

	// Braille histogram
	if len(st.MixturePDF) > 0 {
		chartW := innerWidth - 6
		if chartW < 16 {
			chartW = 16
		}
		chartH := 8
		rows := RenderBrailleChart(st.MixturePDF, chartW, chartH, chartGradient)
		for _, row := range rows {
			lines = append(lines, "   "+row)
		}
		loLabel := fmt.Sprintf("%.0f", st.MixtureLo)
		hiLabel := fmt.Sprintf("%.0f", st.MixtureHi)
		axisGap := chartW - len(loLabel) - len(hiLabel)
		if axisGap < 1 {
			axisGap = 1
		}
		lines = append(lines, "   "+styleMuted.Render(loLabel+strings.Repeat(" ", axisGap)+hiLabel))
	}

	// Heightening
	if len(st.HeightenTable) > 1 {
		lines = append(lines, "")
		lines = append(lines, styleLabel.Render("Heightening"))
		for _, row := range st.HeightenTable {
			marker := "  "
			rowStyle := styleMuted
			if row.Rank == castRank {
				marker = lipgloss.NewStyle().Foreground(colorAccent).Render("> ")
				rowStyle = styleGood
			} else if row.Rank == sp.BaseRank {
				rowStyle = styleLabel
			}
			line := fmt.Sprintf("%s%-2d  %-8s  %6.0f",
				marker, row.Rank, row.Dice, row.Expected)
			lines = append(lines, rowStyle.Render(line))
		}
	}

	return strings.Join(lines, "\n")
}

// ============================================================================
// Comparison Table
// ============================================================================

func (m model) renderComparison() string {
	contentW := m.leftContentWidth()

	type ranked struct {
		name string
		sp   *Spell
		st   *SpellStats
	}
	var items []ranked
	for i := range m.spells {
		e := &m.spells[i]
		if !m.selected[e.id] {
			continue
		}
		name := e.spell.Name
		if name == "" {
			name = fmt.Sprintf("Spell %d", i+1)
		}
		items = append(items, ranked{name: name, sp: &e.spell, st: &e.stats})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].st.ExpectedDamage > items[j].st.ExpectedDamage
	})

	if len(items) == 0 {
		return ""
	}

	bestDmg := items[0].st.ExpectedDamage

	// Dynamic name width: fill remaining space after fixed columns
	fixedCols := 8 + 12 + 8 + 4 // E[dmg] + ±1σ + Δ + prefix/gaps
	nameWidth := contentW - fixedCols
	if nameWidth < 12 {
		nameWidth = 12
	}

	var lines []string
	header := "  " +
		styleLabel.Render(fmt.Sprintf("%-*s", nameWidth, "Spell")) +
		styleLabel.Render(fmt.Sprintf("%8s", "E[dmg]")) +
		styleLabel.Render(fmt.Sprintf("%12s", "±1σ")) +
		styleLabel.Render(fmt.Sprintf("%8s", "Δ"))
	lines = append(lines, header)
	lines = append(lines, styleMuted.Render("  "+strings.Repeat("─", contentW-4)))

	for _, item := range items {
		name := item.name
		nameRunes := []rune(name)
		if len(nameRunes) > nameWidth {
			name = string(nameRunes[:nameWidth-3]) + "..."
		}
		paddedName := fmt.Sprintf("%-*s", nameWidth, name)

		if item.sp.Dice.Valid() {
			avg := item.st.ExpectedDamage
			lo := math.Max(0, avg-item.st.OverallStdDev)
			hi := avg + item.st.OverallStdDev
			rangeStr := fmt.Sprintf("%.0f–%.0f", lo, hi)

			delta := avg - bestDmg
			var deltaCol string
			if delta == 0 {
				deltaCol = styleGood.Render(fmt.Sprintf("%8s", "best"))
			} else {
				deltaCol = styleValue.Render(fmt.Sprintf("%+8.0f", delta))
			}

			line := "  " +
				styleValue.Render(paddedName) +
				styleValue.Render(fmt.Sprintf("%8.1f", avg)) +
				styleMuted.Render(fmt.Sprintf("%12s", rangeStr)) +
				deltaCol
			lines = append(lines, line)
		} else {
			line := "  " +
				styleMuted.Render(paddedName) +
				styleMuted.Render(fmt.Sprintf("%8s", "—")) +
				styleMuted.Render(fmt.Sprintf("%12s", "—")) +
				styleMuted.Render(fmt.Sprintf("%8s", "—"))
			lines = append(lines, line)
		}
	}

	content := strings.Join(lines, "\n")
	return m.renderPanel("Comparison", m.leftWidth(), content)
}

// ============================================================================
// Help Bar
// ============================================================================

func (m model) renderListHelp() string {
	type binding struct {
		key, action string
	}

	var bindings []binding

	if m.encFocus >= 0 {
		// Encounter input mode
		bindings = []binding{
			{"Tab", "next field"},
			{"Esc", "done"},
		}
		if m.encFocus == encRefMod || m.encFocus == encFortMod || m.encFocus == encWillMod {
			bindings = append(bindings, binding{"Space", "cycle Low/Med/High"})
		}
		bindings = append(bindings, binding{"^S", "save"}, binding{"^C", "quit"})
	} else {
		bindings = []binding{
			{"j/k", "navigate"},
			{"Space", "compare"},
			{"a", "add"},
			{"n", "new"},
			{"d", "remove"},
			{"e", "edit"},
			{"Tab", "encounter"},
		}

		if len(m.spells) > 0 && m.cursor < len(m.spells) {
			entry := &m.spells[m.cursor]
			if entry.spell.BaseRank > 0 && entry.spell.HeightenDie > 0 {
				bindings = append(bindings, binding{"+/-", "rank"})
			}
		}

		bindings = append(bindings, binding{"^S", "save"}, binding{"q", "quit"})
	}

	keyStyle := lipgloss.NewStyle().Foreground(colorValue)
	actionStyle := lipgloss.NewStyle().Foreground(colorMuted)
	sepStyle := lipgloss.NewStyle().Foreground(colorBorder)

	var parts []string
	for _, b := range bindings {
		parts = append(parts, keyStyle.Render(b.key)+actionStyle.Render(" "+b.action))
	}

	return " " + strings.Join(parts, sepStyle.Render("  │  "))
}

// ============================================================================
// Edit Mode View
// ============================================================================

func (m model) viewEdit() string {
	if len(m.spells) == 0 {
		return styleMuted.Render("  No spells.")
	}

	entry := &m.spells[m.cursor]

	var sections []string

	breadcrumb := lipgloss.NewStyle().Foreground(colorAccent).Bold(true).
		Render("Edit") +
		styleMuted.Render(": ") +
		styleValueBold.Render(entry.spell.Name)
	sections = append(sections, " "+breadcrumb)

	editLabels := [editFieldCount]string{
		"Name", "Dice", "Defense",
		"Mult Best", "Mult Good", "Mult Bad", "Mult Worst",
		"Base Rank", "Heighten", "Per Ranks",
	}

	var lines []string
	for i := editFieldIndex(0); i < editFieldCount; i++ {
		focused := entry.editFocus == i
		label := fmt.Sprintf("%-14s", editLabels[i]+":")

		indicator := "  "
		labelStyle := styleLabel
		if focused {
			indicator = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("> ")
			labelStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
		}

		fieldView := entry.editInputs[i].View()

		// Add contextual hints for special fields
		hint := ""
		switch i {
		case editSaveType:
			if focused {
				hint = styleMuted.Render("  (Space to cycle)")
			}
		case editHeightenDie:
			val := parseIntOr(entry.editInputs[editHeightenDie].Value(), 0)
			if val > 0 {
				dice := ParseDice(entry.editInputs[editDice].Value())
				step := parseIntOr(entry.editInputs[editHeightenStep].Value(), 1)
				if step < 1 {
					step = 1
				}
				if dice.Valid() {
					if step > 1 {
						hint = styleMuted.Render(fmt.Sprintf("  +%dd%d per %d ranks", val, dice.Sides, step))
					} else {
						hint = styleMuted.Render(fmt.Sprintf("  +%dd%d per rank", val, dice.Sides))
					}
				}
			}
		case editHeightenStep:
			hint = styleMuted.Render("  (1=every rank, 2=every 2 ranks)")
		}

		lines = append(lines, indicator+labelStyle.Render(label)+" "+fieldView+hint)
	}

	content := strings.Join(lines, "\n")
	maxWidth := m.windowWidth - 4
	if maxWidth < 50 {
		maxWidth = 50
	}
	if maxWidth > 70 {
		maxWidth = 70
	}

	panel := m.renderPanel("Edit", maxWidth, content)
	sections = append(sections, panel)

	help := " Tab/Shift+Tab: fields  Space: cycle defense  Enter: confirm  Esc: cancel"
	sections = append(sections, styleHelp.Render(help))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// ============================================================================
// New Spell View
// ============================================================================

func (m model) viewNewSpell() string {
	var sections []string

	sections = append(sections, "")

	var lines []string
	lines = append(lines, styleLabel.Render("  Enter an Archives of Nethys URL to import, or a spell name for manual entry."))
	lines = append(lines, "")
	lines = append(lines, "  "+m.newSpellInput.View())
	lines = append(lines, "")

	content := strings.Join(lines, "\n")

	maxWidth := m.windowWidth - 8
	if maxWidth < 50 {
		maxWidth = 50
	}
	if maxWidth > 80 {
		maxWidth = 80
	}

	panel := m.renderPanel("New Spell", maxWidth, content)
	sections = append(sections, panel)

	help := " Enter: create  Esc: cancel"
	sections = append(sections, styleHelp.Render(help))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
