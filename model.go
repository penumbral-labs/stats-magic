package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// appMode represents the three primary interaction modes.
type appMode int

const (
	modeList appMode = iota // Default: spell list + detail pane
	modeEdit               // Full edit mode (power user)
)

// Encounter input field indices
const (
	encSpellDC = iota
	encAttackMod
	encRefMod
	encFortMod
	encWillMod
	encEnemyAC
	encFieldCount
)

// Edit mode field indices
type editFieldIndex int

const (
	editName editFieldIndex = iota
	editDice
	editSaveType
	editMultBest
	editMultGood
	editMultBad
	editMultWorst
	editBaseRank
	editHeightenDie
	editFieldCount
)

// model is the top-level Bubble Tea model.
type model struct {
	mode      appMode
	encounter EncounterState
	encInputs [encFieldCount]textinput.Model
	encFocus  int // Which encounter field is focused (-1 = none)

	spells   []spellEntry
	cursor   int          // Selected spell in list
	selected map[int]bool // Toggled for comparison

	picker    pickerModel
	flashText string

	windowWidth  int
	windowHeight int
	quitting     bool
}

// spellEntry holds one spell and its computed stats.
type spellEntry struct {
	spell    Spell
	stats    SpellStats
	castRank int // Current rank for heightening (0 = use base rank)

	// Edit mode fields — only populated when entering edit mode.
	editInputs [editFieldCount]textinput.Model
	editFocus  editFieldIndex
}

// effectiveCastRank returns the rank this spell is being cast at.
func (e *spellEntry) effectiveCastRank() int {
	if e.castRank > 0 {
		return e.castRank
	}
	if e.spell.BaseRank > 0 {
		return e.spell.BaseRank
	}
	return 0
}

func (e *spellEntry) recalc(enc EncounterState) {
	e.stats = CalcSpellStatsAtRank(e.spell, enc, e.effectiveCastRank())
}

func parseIntOr(s string, fallback int) int {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "+")
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}

func parseFloatOr(s string, fallback float64) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return fallback
	}
	return v
}

func newInput(placeholder, value string, charLimit int) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = charLimit
	ti.Width = charLimit + 2
	ti.TextStyle = styleValue
	ti.PlaceholderStyle = styleMuted
	ti.PromptStyle = styleMuted
	ti.Prompt = ""
	if value != "" {
		ti.SetValue(value)
	}
	return ti
}

func newEncounterInputs(enc EncounterState) [encFieldCount]textinput.Model {
	var inputs [encFieldCount]textinput.Model
	inputs[encSpellDC] = newInput("DC", strconv.Itoa(enc.SpellDC), 4)
	inputs[encAttackMod] = newInput("Atk", fmt.Sprintf("+%d", enc.AttackMod), 5)
	inputs[encRefMod] = newInput("Ref", fmt.Sprintf("+%d", enc.RefMod), 5)
	inputs[encFortMod] = newInput("Fort", fmt.Sprintf("+%d", enc.FortMod), 5)
	inputs[encWillMod] = newInput("Will", fmt.Sprintf("+%d", enc.WillMod), 5)
	inputs[encEnemyAC] = newInput("AC", strconv.Itoa(enc.EnemyAC), 4)
	return inputs
}

func (m *model) syncEncounterFromInputs() {
	m.encounter.SpellDC = parseIntOr(m.encInputs[encSpellDC].Value(), m.encounter.SpellDC)
	m.encounter.AttackMod = parseIntOr(m.encInputs[encAttackMod].Value(), m.encounter.AttackMod)
	m.encounter.RefMod = parseIntOr(m.encInputs[encRefMod].Value(), m.encounter.RefMod)
	m.encounter.FortMod = parseIntOr(m.encInputs[encFortMod].Value(), m.encounter.FortMod)
	m.encounter.WillMod = parseIntOr(m.encInputs[encWillMod].Value(), m.encounter.WillMod)
	m.encounter.EnemyAC = parseIntOr(m.encInputs[encEnemyAC].Value(), m.encounter.EnemyAC)
}

func (m *model) recalcAll() {
	for i := range m.spells {
		m.spells[i].recalc(m.encounter)
	}
}

func (m *model) focusEncField(idx int) {
	if m.encFocus >= 0 && m.encFocus < encFieldCount {
		m.encInputs[m.encFocus].Blur()
	}
	m.encFocus = idx
	if idx >= 0 && idx < encFieldCount {
		m.encInputs[idx].Focus()
	}
}

func (m *model) blurEncFields() {
	if m.encFocus >= 0 && m.encFocus < encFieldCount {
		m.encInputs[m.encFocus].Blur()
	}
	m.encFocus = -1
}

// initEditInputs populates edit mode text inputs from the spell.
func (e *spellEntry) initEditInputs() {
	sp := e.spell
	e.editInputs[editName] = newInput("Name", sp.Name, 30)
	e.editInputs[editDice] = newInput("Dice", sp.Dice.String(), 12)

	saveTypeVal := sp.SaveType
	if sp.Type == SpellTypeAttack {
		saveTypeVal = "attack"
	}
	e.editInputs[editSaveType] = newInput("Save", saveTypeVal, 12)

	e.editInputs[editMultBest] = newInput("Best", fmt.Sprintf("%.1f", sp.Multipliers.Best), 6)
	e.editInputs[editMultGood] = newInput("Good", fmt.Sprintf("%.1f", sp.Multipliers.Good), 6)
	e.editInputs[editMultBad] = newInput("Bad", fmt.Sprintf("%.1f", sp.Multipliers.Bad), 6)
	e.editInputs[editMultWorst] = newInput("Worst", fmt.Sprintf("%.1f", sp.Multipliers.Worst), 6)

	rankStr := ""
	if sp.BaseRank > 0 {
		rankStr = strconv.Itoa(sp.BaseRank)
	}
	e.editInputs[editBaseRank] = newInput("Rank", rankStr, 4)

	hdStr := ""
	if sp.HeightenDie > 0 {
		hdStr = strconv.Itoa(sp.HeightenDie)
	}
	e.editInputs[editHeightenDie] = newInput("+dice/rank", hdStr, 4)

	e.editFocus = editName
	e.editInputs[editName].Focus()
}

// syncSpellFromEditInputs reads edit inputs back into the spell.
func (e *spellEntry) syncSpellFromEditInputs() {
	e.spell.Name = e.editInputs[editName].Value()
	e.spell.Dice = ParseDice(e.editInputs[editDice].Value())

	saveVal := strings.TrimSpace(e.editInputs[editSaveType].Value())
	switch strings.ToLower(saveVal) {
	case "attack", "atk", "":
		e.spell.Type = SpellTypeAttack
		e.spell.SaveType = ""
		// Set attack multipliers if they look like save defaults
		if e.spell.Multipliers.Bad == 0.5 {
			e.spell.Multipliers = DefaultAttackMultipliers()
		}
	default:
		e.spell.Type = SpellTypeSave
		switch {
		case strings.HasPrefix(strings.ToLower(saveVal), "fort"):
			e.spell.SaveType = "Fortitude"
		case strings.HasPrefix(strings.ToLower(saveVal), "will"):
			e.spell.SaveType = "Will"
		default:
			e.spell.SaveType = "Reflex"
		}
	}

	e.spell.Multipliers.Best = parseFloatOr(e.editInputs[editMultBest].Value(), e.spell.Multipliers.Best)
	e.spell.Multipliers.Good = parseFloatOr(e.editInputs[editMultGood].Value(), e.spell.Multipliers.Good)
	e.spell.Multipliers.Bad = parseFloatOr(e.editInputs[editMultBad].Value(), e.spell.Multipliers.Bad)
	e.spell.Multipliers.Worst = parseFloatOr(e.editInputs[editMultWorst].Value(), e.spell.Multipliers.Worst)

	e.spell.BaseRank = parseIntOr(e.editInputs[editBaseRank].Value(), 0)
	e.spell.HeightenDie = parseIntOr(e.editInputs[editHeightenDie].Value(), 0)
}

// --- Bubble Tea Interface ---

func initialModel() model {
	enc := DefaultEncounter()
	m := model{
		mode:         modeList,
		encounter:    enc,
		encInputs:    newEncounterInputs(enc),
		encFocus:     -1,
		selected:     make(map[int]bool),
		windowWidth:  100,
		windowHeight: 40,
	}

	// Try loading saved data
	savedEnc, savedSpells, err := loadData()
	switch {
	case err != nil && !os.IsNotExist(err):
		m.flashText = fmt.Sprintf("Load failed: %v", err)
	case len(savedSpells) > 0:
		m.encounter = savedEnc
		m.encInputs = newEncounterInputs(savedEnc)
		for _, sp := range savedSpells {
			entry := spellEntry{spell: sp}
			entry.recalc(m.encounter)
			m.spells = append(m.spells, entry)
		}
		m.flashText = fmt.Sprintf("Loaded %d spell(s)", len(m.spells))
	}

	m.picker = newPickerModel(m.windowWidth, m.windowHeight)
	return m
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{textinput.Blink}
	if m.flashText != "" {
		cmds = append(cmds, flashAfter(3*time.Second))
	}
	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle messages that can arrive after picker closes
	switch msg := msg.(type) {
	case clearFlashMsg:
		m.flashText = ""
		return m, nil

	case pickerMsg:
		sp := msg.preset.ToSpell()
		entry := spellEntry{spell: sp}
		entry.recalc(m.encounter)
		m.spells = append(m.spells, entry)
		m.cursor = len(m.spells) - 1
		m.flashText = fmt.Sprintf("Added: %s", msg.preset.Name)
		return m, flashAfter(3 * time.Second)

	case pickerCancelMsg:
		return m, nil
	}

	// Handle picker overlay
	if m.picker.active {
		return m.updatePicker(msg)
	}

	switch m.mode {
	case modeList:
		return m.updateList(msg)
	case modeEdit:
		return m.updateEdit(msg)
	}

	return m, nil
}

func (m model) updatePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.picker, cmd = m.picker.Update(msg)
	return m, cmd
}

func (m model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		m.picker = newPickerModel(m.windowWidth, m.windowHeight)
		return m, nil

	case tea.KeyMsg:
		// If an encounter field is focused, handle input
		if m.encFocus >= 0 {
			return m.updateEncounterInput(msg)
		}

		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "q":
			m.quitting = true
			return m, tea.Quit

		case "j", "down":
			if len(m.spells) > 0 {
				m.cursor = (m.cursor + 1) % len(m.spells)
			}
			return m, nil

		case "k", "up":
			if len(m.spells) > 0 {
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.spells) - 1
				}
			}
			return m, nil

		case "enter", "e":
			if len(m.spells) > 0 {
				m.spells[m.cursor].initEditInputs()
				m.mode = modeEdit
				m.blurEncFields()
			}
			return m, nil

		case "+", "=":
			if len(m.spells) > 0 {
				entry := &m.spells[m.cursor]
				if entry.spell.BaseRank > 0 && entry.spell.HeightenDie > 0 {
					rank := entry.effectiveCastRank()
					if rank < 10 {
						entry.castRank = rank + 1
						entry.recalc(m.encounter)
					}
				}
			}
			return m, nil

		case "-", "_":
			if len(m.spells) > 0 {
				entry := &m.spells[m.cursor]
				if entry.spell.BaseRank > 0 && entry.spell.HeightenDie > 0 {
					rank := entry.effectiveCastRank()
					if rank > entry.spell.BaseRank {
						entry.castRank = rank - 1
						entry.recalc(m.encounter)
					}
				}
			}
			return m, nil

		case " ":
			if len(m.spells) > 0 {
				if m.selected[m.cursor] {
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = true
				}
			}
			return m, nil

		case "a":
			m.picker = newPickerModel(m.windowWidth, m.windowHeight)
			m.picker.active = true
			return m, nil

		case "n":
			sp := NewSaveSpell(fmt.Sprintf("Spell %d", len(m.spells)+1))
			entry := spellEntry{spell: sp}
			entry.recalc(m.encounter)
			entry.initEditInputs()
			m.spells = append(m.spells, entry)
			m.cursor = len(m.spells) - 1
			m.mode = modeEdit
			m.blurEncFields()
			return m, nil

		case "d":
			if len(m.spells) > 0 {
				// Clear selection references for deleted spell
				delete(m.selected, m.cursor)
				// Rebuild selected map with shifted indices
				newSel := make(map[int]bool)
				for k := range m.selected {
					if k > m.cursor {
						newSel[k-1] = true
					} else {
						newSel[k] = true
					}
				}
				m.selected = newSel

				m.spells = append(m.spells[:m.cursor], m.spells[m.cursor+1:]...)
				if m.cursor >= len(m.spells) && m.cursor > 0 {
					m.cursor = len(m.spells) - 1
				}
			}
			return m, nil

		case "tab":
			m.focusEncField(0)
			return m, nil

		case "ctrl+s":
			if err := saveData(m.encounter, m.spells); err != nil {
				m.flashText = fmt.Sprintf("Save failed: %v", err)
			} else {
				m.flashText = fmt.Sprintf("Saved %d spell(s)!", len(m.spells))
			}
			return m, flashAfter(3 * time.Second)
		}
	}

	return m, nil
}

func (m model) updateEncounterInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			next := (m.encFocus + 1) % encFieldCount
			m.focusEncField(next)
			m.syncEncounterFromInputs()
			m.recalcAll()
			return m, nil

		case "shift+tab":
			prev := m.encFocus - 1
			if prev < 0 {
				prev = encFieldCount - 1
			}
			m.focusEncField(prev)
			m.syncEncounterFromInputs()
			m.recalcAll()
			return m, nil

		case "esc", "enter":
			m.syncEncounterFromInputs()
			m.recalcAll()
			m.blurEncFields()
			return m, nil

		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}

		// Forward to active input
		var cmd tea.Cmd
		m.encInputs[m.encFocus], cmd = m.encInputs[m.encFocus].Update(msg)
		m.syncEncounterFromInputs()
		m.recalcAll()
		return m, cmd
	}

	return m, nil
}

func (m model) updateEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(m.spells) == 0 {
		m.mode = modeList
		return m, nil
	}

	entry := &m.spells[m.cursor]

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Cancel, revert to detail
			m.mode = modeList
			return m, nil

		case "enter":
			// Confirm changes
			entry.syncSpellFromEditInputs()
			entry.recalc(m.encounter)
			m.mode = modeList
			return m, nil

		case "tab":
			entry.editInputs[entry.editFocus].Blur()
			entry.editFocus = (entry.editFocus + 1) % editFieldCount
			entry.editInputs[entry.editFocus].Focus()
			return m, nil

		case "shift+tab":
			entry.editInputs[entry.editFocus].Blur()
			entry.editFocus--
			if entry.editFocus < 0 {
				entry.editFocus = editFieldCount - 1
			}
			entry.editInputs[entry.editFocus].Focus()
			return m, nil

		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}

		// Forward to active edit input
		var cmd tea.Cmd
		entry.editInputs[entry.editFocus], cmd = entry.editInputs[entry.editFocus].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	// Picker overlay
	if m.picker.active {
		return m.picker.View()
	}

	switch m.mode {
	case modeList:
		return m.viewList()
	case modeEdit:
		return m.viewEdit()
	}

	return ""
}
