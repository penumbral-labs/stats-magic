package main

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// --- Color Palette ---

var (
	colorBorder    = lipgloss.Color("#e94560")
	colorBorderDim = lipgloss.Color("#533483")
	colorTitle     = lipgloss.Color("#e94560")
	colorLabel     = lipgloss.Color("#a8a8b3")
	colorValue     = lipgloss.Color("#eaeaea")
	colorHighlight = lipgloss.Color("#e94560")
	colorMuted     = lipgloss.Color("#888899")
	colorGood      = lipgloss.Color("#53d769")
	colorBar       = lipgloss.Color("#e94560")
	colorCursor    = lipgloss.Color("#e94560")
	colorSelected  = lipgloss.Color("#ffc107")
)

// --- Styles ---

var (
	stylePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 2)

	stylePanelDim = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorderDim).
			Padding(0, 2)

	styleTitle = lipgloss.NewStyle().
			Foreground(colorTitle).
			Bold(true)

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
			Foreground(colorMuted).
			Italic(true)

	styleDistribution = lipgloss.NewStyle().
				Foreground(colorBar)
)

// degreeColors for degree breakdown (best-for-caster to worst)
var degreeColors = [4]lipgloss.Color{
	lipgloss.Color("#53d769"), // Best: bright green
	lipgloss.Color("#4ecdc4"), // Good: teal-green
	lipgloss.Color("#ffc107"), // Bad: yellow
	lipgloss.Color("#e94560"), // Worst: red
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

// ============================================================================
// Main View — Single Page Layout
// ============================================================================

func (m model) viewList() string {
	header := m.renderHeader()
	help := m.renderListHelp()
	headerH := lipgloss.Height(header)
	helpH := lipgloss.Height(help)

	// Available height for the main body (between header and help)
	bodyH := m.windowHeight - headerH - helpH
	if bodyH < 10 {
		bodyH = 10
	}

	// Render fixed-height components first to measure
	encPanel := m.renderEncounterPanel()
	encHeight := lipgloss.Height(encPanel)

	var compSection string
	compHeight := 0
	if len(m.selected) >= 2 {
		compSection = m.renderComparison()
		compHeight = lipgloss.Height(compSection) + 1 // +1 for gap line
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
		// Calculate spell list rows from available body height
		// body = enc(+1 gap) + spellList + comp(+1 gap) + flash
		fixedHeight := encHeight + 1 + compHeight + flashHeight
		spellListRows := bodyH - fixedHeight
		if spellListRows < 5 {
			spellListRows = 5
		}

		spellList := m.renderSpellListFixed(spellListRows)

		// Assemble left column
		var leftParts []string
		leftParts = append(leftParts, encPanel)
		leftParts = append(leftParts, "")
		leftParts = append(leftParts, spellList)
		if compSection != "" {
			leftParts = append(leftParts, "")
			leftParts = append(leftParts, compSection)
		}
		if flashSection != "" {
			leftParts = append(leftParts, flashSection)
		}
		leftCol := lipgloss.JoinVertical(lipgloss.Left, leftParts...)

		// Detail pane fills remaining width
		paneWidth := m.detailPaneWidth()
		detailContent := m.renderDetailContent()

		// Match detail pane height to left column
		leftHeight := lipgloss.Height(leftCol)
		detailPanel := stylePanel.Width(paneWidth).Render(detailContent)
		detailHeight := lipgloss.Height(detailPanel)
		if detailHeight < leftHeight {
			extra := leftHeight - detailHeight
			detailPanel = stylePanel.Width(paneWidth).Render(detailContent + strings.Repeat("\n", extra))
		}

		rightCol := lipgloss.NewStyle().MarginLeft(1).Render(detailPanel)
		body = lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)
	} else {
		// No spells: just encounter + empty spell list
		spellListRows := bodyH - encHeight - 1
		if spellListRows < 5 {
			spellListRows = 5
		}
		spellList := m.renderSpellListFixed(spellListRows)
		body = lipgloss.JoinVertical(lipgloss.Left, encPanel, "", spellList)
	}

	// Pad body to fill available height so help bar stays at the bottom
	bodyHeight := lipgloss.Height(body)
	if bodyHeight < bodyH {
		body += strings.Repeat("\n", bodyH-bodyHeight)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, body, help)
}

func (m model) renderHeader() string {
	title := lipgloss.NewStyle().
		Foreground(colorHighlight).Bold(true).
		Render("Stats Magic")
	subtitle := styleMuted.Render(" — PF2e Spell Damage Calculator")
	return "\n " + title + subtitle
}

// leftWidth returns the outer width for both the encounter panel and spell list.
func (m model) leftWidth() int {
	// Target: enough for the spell list content. The encounter grid fits within this.
	w := m.windowWidth/2 + 4
	if w < 56 {
		w = 56
	}
	if w > 72 {
		w = 72
	}
	return w
}

func (m model) detailPaneWidth() int {
	// Fill remaining width: window - left column - gap - border overhead
	remaining := m.windowWidth - m.leftWidth() - 5
	if remaining < 34 {
		remaining = 34
	}
	return remaining
}

func (m model) renderEncounterPanel() string {
	encLabels := [encFieldCount]string{"Spell DC", "Attack Mod", "Reflex", "Fortitude", "Will", "AC"}

	renderField := func(idx int) string {
		focused := m.encFocus == idx
		label := fmt.Sprintf("%-12s", encLabels[idx]+":")
		indicator := "  "
		labelStyle := styleLabel
		if focused {
			indicator = lipgloss.NewStyle().Foreground(colorHighlight).Bold(true).Render("> ")
			labelStyle = lipgloss.NewStyle().Foreground(colorHighlight).Bold(true)
		}
		return indicator + labelStyle.Render(label) + m.encInputs[idx].View()
	}

	// Build each column as a single string block, let lipgloss handle alignment
	youFields := []int{encSpellDC, encAttackMod}
	enemyFields := []int{encRefMod, encFortMod, encWillMod, encEnemyAC}

	var leftLines []string
	leftLines = append(leftLines, styleLabel.Render("  You"))
	leftLines = append(leftLines, "")
	for _, idx := range youFields {
		leftLines = append(leftLines, renderField(idx))
	}
	leftCol := strings.Join(leftLines, "\n")

	var rightLines []string
	rightLines = append(rightLines, styleLabel.Render("  Enemy"))
	rightLines = append(rightLines, "")
	for _, idx := range enemyFields {
		rightLines = append(rightLines, renderField(idx))
	}
	rightCol := strings.Join(rightLines, "\n")

	// Use lipgloss to join — it handles ANSI width correctly
	colWidth := 28
	leftSized := lipgloss.NewStyle().Width(colWidth).Render(leftCol)
	grid := lipgloss.JoinHorizontal(lipgloss.Top, leftSized, rightCol)

	return stylePanelDim.Width(m.leftWidth()).Render(grid)
}

// renderSpellListFixed renders the spell list panel with a fixed number of visible rows.
// If there are more spells than rows, it scrolls to keep the cursor visible.
func (m model) renderSpellListFixed(maxRows int) string {
	w := m.leftWidth()

	if len(m.spells) == 0 {
		empty := styleMuted.Render("  No spells yet. Press  a  to add from presets.")
		content := styleTitle.Render("Spells") + "\n\n" + empty
		// Pad to target height (maxRows lines + title + gap = maxRows+2, minus the 3 we already have)
		padLines := maxRows - 1
		if padLines > 0 {
			content += strings.Repeat("\n", padLines)
		}
		return stylePanelDim.Width(w).Render(content)
	}

	maxDmg := 0.0
	for _, e := range m.spells {
		if e.spell.Dice.Valid() && e.stats.ExpectedDamage > maxDmg {
			maxDmg = e.stats.ExpectedDamage
		}
	}

	sparkWidth := 16

	// Build all spell lines
	var allLines []string
	for i, e := range m.spells {
		allLines = append(allLines, m.renderSpellRow(i, &e, sparkWidth, maxDmg))
	}

	// Determine visible window around cursor
	visibleRows := maxRows - 2 // subtract title + gap
	if visibleRows < 3 {
		visibleRows = 3
	}

	startIdx := 0
	if len(allLines) > visibleRows {
		// Keep cursor centered in the window
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
	lines = append(lines, styleTitle.Render("Spells"))
	lines = append(lines, "")

	// Scroll indicator top
	if startIdx > 0 {
		lines = append(lines, styleMuted.Render(fmt.Sprintf("  ↑ %d more", startIdx)))
	}

	lines = append(lines, allLines[startIdx:endIdx]...)

	// Scroll indicator bottom
	remaining := len(allLines) - endIdx
	if remaining > 0 {
		lines = append(lines, styleMuted.Render(fmt.Sprintf("  ↓ %d more", remaining)))
	}

	// Pad to fill target height
	currentLines := len(lines)
	targetLines := maxRows
	if currentLines < targetLines {
		for i := 0; i < targetLines-currentLines; i++ {
			lines = append(lines, "")
		}
	}

	content := strings.Join(lines, "\n")
	return stylePanel.Width(w).Render(content)
}

// renderSpellRow renders a single spell line for the list.
func (m model) renderSpellRow(i int, e *spellEntry, sparkWidth int, maxDmg float64) string {
	isCursor := i == m.cursor
	isSel := m.selected[i]

	prefix := "  "
	if isCursor && isSel {
		prefix = lipgloss.NewStyle().Foreground(colorCursor).Bold(true).Render(">*")
	} else if isCursor {
		prefix = lipgloss.NewStyle().Foreground(colorCursor).Bold(true).Render("> ")
	} else if isSel {
		prefix = lipgloss.NewStyle().Foreground(colorSelected).Render(" *")
	}

	name := e.spell.Name
	if name == "" {
		name = fmt.Sprintf("Spell %d", i+1)
	}
	if len(name) > 16 {
		name = name[:13] + "..."
	}

	nameStyle := styleValue
	if isCursor {
		nameStyle = lipgloss.NewStyle().Foreground(colorValue).Bold(true)
	}

	paddedName := fmt.Sprintf("%-16s", name)
	paddedDice := fmt.Sprintf("%6s", e.spell.Dice.String())

	// Per-spell sparkline
	var spark string
	if e.spell.Dice.Valid() && len(e.stats.MixturePDF) > 0 {
		sparkRaw := RenderCompactHistogram(e.stats.MixturePDF, e.stats.MixtureLo, e.stats.MixtureHi, sparkWidth)
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

// renderDetailContent builds the inner content for the detail pane.
func (m model) renderDetailContent() string {
	entry := &m.spells[m.cursor]
	sp := &entry.spell
	st := &entry.stats
	labels := DegreeLabels(sp.Type)
	innerWidth := m.detailPaneWidth() - 6 // border + padding

	var lines []string

	castRank := entry.effectiveCastRank()
	lines = append(lines, styleValueBold.Render(sp.Name))

	var meta []string
	if sp.BaseRank > 0 {
		if castRank > sp.BaseRank {
			meta = append(meta, styleGood.Render(fmt.Sprintf("Rank %d", castRank))+
				styleMuted.Render(fmt.Sprintf(" (base %d)", sp.BaseRank)))
		} else {
			meta = append(meta, styleLabel.Render(fmt.Sprintf("Rank %d", sp.BaseRank)))
		}
	}
	meta = append(meta, styleLabel.Render(sp.DefenseLabel()))
	lines = append(lines, strings.Join(meta, styleMuted.Render(" · ")))

	effectiveDice := sp.Dice
	if castRank > 0 {
		effectiveDice = sp.EffectiveDice(castRank)
	}
	diceInfo := styleValueBold.Render(effectiveDice.String())
	if sp.HeightenDie > 0 {
		diceInfo += styleMuted.Render(fmt.Sprintf(" (+%dd/rank)", sp.HeightenDie))
	}
	lines = append(lines, diceInfo)

	if !effectiveDice.Valid() {
		lines = append(lines, "")
		lines = append(lines, styleMuted.Render("Invalid dice formula"))
		return strings.Join(lines, "\n")
	}

	if sp.Type == SpellTypeSave {
		saveMod := m.encounter.SaveModFor(sp.SaveType)
		lines = append(lines, styleMuted.Render(
			fmt.Sprintf("vs %s +%d (DC %d)", sp.SaveType, saveMod, m.encounter.SpellDC)))
	} else {
		lines = append(lines, styleMuted.Render(
			fmt.Sprintf("+%d vs AC %d", m.encounter.AttackMod, m.encounter.EnemyAC)))
	}
	lines = append(lines, "")

	// Degree breakdown
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

	for i := 0; i < 4; i++ {
		pct := probs[i] * 100
		degStyle := lipgloss.NewStyle().Foreground(degreeColors[i])

		barLen := 0
		if maxProb > 0 {
			barLen = int(math.Round(probs[i] / maxProb * float64(probBarWidth)))
		}
		bar := degStyle.Render(strings.Repeat("█", barLen)) +
			strings.Repeat(" ", probBarWidth-barLen)

		multStr := fmt.Sprintf("%.0fx", mults[i])
		if mults[i] == 0.5 {
			multStr = ".5x"
		}

		dmgStr := "  —"
		if mults[i] > 0 {
			dmgStr = fmt.Sprintf("%3s", fmt.Sprintf("~%.0f", st.DegreeMean[i]))
		}

		paddedLabel := fmt.Sprintf("%-13s", labels[i])

		line := paddedLabel +
			styleMuted.Render(fmt.Sprintf("%3s", multStr)) +
			fmt.Sprintf(" %3.0f%% ", pct) +
			bar + " " +
			degStyle.Render(fmt.Sprintf("%4s", dmgStr))
		lines = append(lines, line)
	}
	lines = append(lines, "")

	// Summary
	lines = append(lines, styleGood.Render(fmt.Sprintf("~%.0f damage on average", st.ExpectedDamage)))
	lines = append(lines, styleMuted.Render(fmt.Sprintf("  typically %.0f–%.0f",
		math.Max(0, st.ExpectedDamage-st.OverallStdDev),
		st.ExpectedDamage+st.OverallStdDev)))
	lines = append(lines, styleMuted.Render(fmt.Sprintf("  %.0f%% chance to deal damage", st.AnyDamageProb*100)))

	// Tall histogram
	if len(st.MixturePDF) > 0 {
		lines = append(lines, "")
		chartW := innerWidth - 8
		if chartW < 16 {
			chartW = 16
		}
		chartH := 6
		rows := RenderTallHistogram(st.MixturePDF, st.MixtureLo, st.MixtureHi, chartW, chartH, colorBar)
		for _, row := range rows {
			lines = append(lines, "    "+row)
		}
		// Axis labels — show "dmg" suffix for clarity
		loLabel := fmt.Sprintf("%.0f", st.MixtureLo)
		hiLabel := fmt.Sprintf("%.0f", st.MixtureHi)
		axisGap := chartW - len(loLabel) - len(hiLabel)
		if axisGap < 1 {
			axisGap = 1
		}
		lines = append(lines, "    "+styleMuted.Render(loLabel+strings.Repeat(" ", axisGap)+hiLabel))
	}

	// Heightening
	if len(st.HeightenTable) > 1 {
		lines = append(lines, "")
		lines = append(lines, styleLabel.Render("Heightening:"))
		for _, row := range st.HeightenTable {
			marker := "  "
			rowStyle := styleMuted
			if row.Rank == castRank {
				marker = "> "
				rowStyle = styleGood
			} else if row.Rank == sp.BaseRank {
				rowStyle = styleLabel
			}
			line := fmt.Sprintf("%s%d  %-8s  ~%.0f",
				marker, row.Rank, row.Dice, row.Expected)
			lines = append(lines, rowStyle.Render(line))
		}
	}

	return strings.Join(lines, "\n")
}

func (m model) renderComparison() string {
	type ranked struct {
		name string
		sp   *Spell
		st   *SpellStats
	}
	var items []ranked
	for idx := range m.selected {
		if idx < len(m.spells) {
			e := &m.spells[idx]
			name := e.spell.Name
			if name == "" {
				name = fmt.Sprintf("Spell %d", idx+1)
			}
			items = append(items, ranked{name: name, sp: &e.spell, st: &e.stats})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].st.ExpectedDamage > items[j].st.ExpectedDamage
	})

	if len(items) == 0 {
		return ""
	}

	bestDmg := items[0].st.ExpectedDamage

	var lines []string
	// Clear column headers: Spell name, expected damage, ± range, difference from best
	header := "  " +
		styleLabel.Render(fmt.Sprintf("%-18s", "Spell")) +
		styleLabel.Render(fmt.Sprintf("%8s", "E[dmg]")) +
		styleLabel.Render(fmt.Sprintf("%12s", "±1σ")) +
		styleLabel.Render(fmt.Sprintf("%8s", "Δ"))
	lines = append(lines, header)
	lines = append(lines, styleMuted.Render("  "+strings.Repeat("─", 46)))

	for _, item := range items {
		name := item.name
		if len(name) > 18 {
			name = name[:15] + "..."
		}
		paddedName := fmt.Sprintf("%-18s", name)

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
	return "\n" + styleTitle.Render(" Comparison") + "\n" + content
}

func (m model) renderListHelp() string {
	var parts []string
	parts = append(parts, "j/k: navigate")
	parts = append(parts, "Space: compare")
	parts = append(parts, "a: add")
	parts = append(parts, "d: remove")
	parts = append(parts, "e: edit")
	parts = append(parts, "Tab: encounter")

	if len(m.spells) > 0 {
		entry := &m.spells[m.cursor]
		if entry.spell.BaseRank > 0 && entry.spell.HeightenDie > 0 {
			parts = append(parts, "+/-: cast rank")
		}
	}

	parts = append(parts, "Ctrl+S: save")
	parts = append(parts, "q: quit")
	help := " " + strings.Join(parts, "  ")
	return "\n" + styleHelp.Render(help)
}

// ============================================================================
// Edit Mode View
// ============================================================================

func (m model) viewEdit() string {
	if len(m.spells) == 0 {
		return m.renderHeader() + "\n" + styleMuted.Render("  No spells.")
	}

	entry := &m.spells[m.cursor]

	var sections []string

	breadcrumb := lipgloss.NewStyle().Foreground(colorHighlight).Bold(true).
		Render("Stats Magic") +
		styleMuted.Render(" > ") +
		styleValueBold.Render("Edit: "+entry.spell.Name)
	sections = append(sections, "\n "+breadcrumb+"\n")

	editLabels := [editFieldCount]string{
		"Name", "Dice", "Save Type",
		"Mult Best", "Mult Good", "Mult Bad", "Mult Worst",
		"Base Rank", "+Dice/Rank",
	}

	var lines []string
	for i := editFieldIndex(0); i < editFieldCount; i++ {
		focused := entry.editFocus == i
		label := fmt.Sprintf("%-14s", editLabels[i]+":")

		indicator := "  "
		labelStyle := styleLabel
		if focused {
			indicator = lipgloss.NewStyle().Foreground(colorHighlight).Bold(true).Render("> ")
			labelStyle = lipgloss.NewStyle().Foreground(colorHighlight).Bold(true)
		}

		lines = append(lines, indicator+labelStyle.Render(label)+" "+entry.editInputs[i].View())
	}

	lines = append(lines, "")
	lines = append(lines, styleMuted.Render("  Save Type: Reflex, Fortitude, Will, or Attack"))

	content := strings.Join(lines, "\n")
	maxWidth := m.windowWidth - 4
	if maxWidth < 50 {
		maxWidth = 50
	}
	if maxWidth > 70 {
		maxWidth = 70
	}

	panel := stylePanel.Width(maxWidth).Render(content)
	sections = append(sections, panel)

	help := " Tab/Shift+Tab: fields  Enter: confirm  Esc: cancel"
	sections = append(sections, "\n"+styleHelp.Render(help))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
