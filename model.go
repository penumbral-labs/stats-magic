package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// appMode represents the primary interaction modes.
type appMode int

const (
	modeList     appMode = iota // Default: spell list + detail pane
	modeEdit                    // Full edit mode (power user)
	modeNewSpell                // New spell: enter name or AoN URL
)

// Encounter input field indices
const (
	encPCLevel = iota
	encSpellDC
	encAttackMod
	encEnemyAC
	encRefMod
	encFortMod
	encWillMod
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
	editHeightenStep
	editFieldCount
)

// Save type options for cycling in edit mode.
var saveTypeCycle = []string{"Reflex", "Fortitude", "Will", "Attack"}

// model is the top-level Bubble Tea model.
type model struct {
	mode      appMode
	encounter EncounterState
	encInputs [encFieldCount]textinput.Model
	encFocus  int // Which encounter field is focused (-1 = none)

	spells      []spellEntry
	nextSpellID int          // Monotonic ID counter for unique spell identification
	cursor      int          // Selected spell in list
	selected    map[int]bool // Spell ID → selected for comparison

	picker    pickerModel
	flashText string
	flashID   int // Prevents stale timers from clearing newer messages

	// New spell mode input
	newSpellInput textinput.Model

	windowWidth  int
	windowHeight int
	quitting     bool
}

// spellEntry holds one spell and its computed stats.
type spellEntry struct {
	id       int // Unique identifier (never reused)
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
	inputs[encPCLevel] = newInput("Lvl", strconv.Itoa(enc.PCLevel), 4)
	inputs[encSpellDC] = newInput("DC", strconv.Itoa(enc.SpellDC), 4)
	inputs[encAttackMod] = newInput("Atk", fmt.Sprintf("+%d", enc.AttackMod), 5)
	inputs[encEnemyAC] = newInput("AC", strconv.Itoa(enc.EnemyAC), 4)
	inputs[encRefMod] = newInput("Ref", fmt.Sprintf("+%d", enc.RefMod), 5)
	inputs[encFortMod] = newInput("Fort", fmt.Sprintf("+%d", enc.FortMod), 5)
	inputs[encWillMod] = newInput("Will", fmt.Sprintf("+%d", enc.WillMod), 5)
	return inputs
}

func (m *model) syncEncounterFromInputs() {
	newLevel := parseIntOr(m.encInputs[encPCLevel].Value(), m.encounter.PCLevel)
	if newLevel < 1 {
		newLevel = 1
	}
	if newLevel > 20 {
		newLevel = 20
	}
	levelChanged := newLevel != m.encounter.PCLevel

	m.encounter.PCLevel = newLevel

	if levelChanged {
		m.encounter.RecalcFromLevel()
		// Update all input values to match
		m.encInputs[encSpellDC].SetValue(strconv.Itoa(m.encounter.SpellDC))
		m.encInputs[encAttackMod].SetValue(fmt.Sprintf("+%d", m.encounter.AttackMod))
		m.encInputs[encEnemyAC].SetValue(strconv.Itoa(m.encounter.EnemyAC))
		m.encInputs[encRefMod].SetValue(fmt.Sprintf("+%d", m.encounter.RefMod))
		m.encInputs[encFortMod].SetValue(fmt.Sprintf("+%d", m.encounter.FortMod))
		m.encInputs[encWillMod].SetValue(fmt.Sprintf("+%d", m.encounter.WillMod))
	} else {
		m.encounter.SpellDC = parseIntOr(m.encInputs[encSpellDC].Value(), m.encounter.SpellDC)
		m.encounter.AttackMod = parseIntOr(m.encInputs[encAttackMod].Value(), m.encounter.AttackMod)
		m.encounter.EnemyAC = parseIntOr(m.encInputs[encEnemyAC].Value(), m.encounter.EnemyAC)
		m.encounter.RefMod = parseIntOr(m.encInputs[encRefMod].Value(), m.encounter.RefMod)
		m.encounter.FortMod = parseIntOr(m.encInputs[encFortMod].Value(), m.encounter.FortMod)
		m.encounter.WillMod = parseIntOr(m.encInputs[encWillMod].Value(), m.encounter.WillMod)
	}
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

// cycleSaveProfile cycles the save profile for the currently focused encounter field.
// Returns true if a profile was cycled.
func (m *model) cycleSaveProfile() bool {
	switch m.encFocus {
	case encRefMod:
		m.encounter.RefProfile = m.encounter.RefProfile.Next()
		m.encounter.RefMod = MonsterSave(m.encounter.PCLevel, m.encounter.RefProfile)
		m.encInputs[encRefMod].SetValue(fmt.Sprintf("+%d", m.encounter.RefMod))
	case encFortMod:
		m.encounter.FortProfile = m.encounter.FortProfile.Next()
		m.encounter.FortMod = MonsterSave(m.encounter.PCLevel, m.encounter.FortProfile)
		m.encInputs[encFortMod].SetValue(fmt.Sprintf("+%d", m.encounter.FortMod))
	case encWillMod:
		m.encounter.WillProfile = m.encounter.WillProfile.Next()
		m.encounter.WillMod = MonsterSave(m.encounter.PCLevel, m.encounter.WillProfile)
		m.encInputs[encWillMod].SetValue(fmt.Sprintf("+%d", m.encounter.WillMod))
	default:
		return false
	}
	return true
}

// addSpellEntry creates a new spell entry with a unique ID and adds it to the model.
func (m *model) addSpellEntry(sp Spell) {
	m.nextSpellID++
	entry := spellEntry{id: m.nextSpellID, spell: sp}
	entry.recalc(m.encounter)
	m.spells = append(m.spells, entry)
	m.cursor = len(m.spells) - 1
}

// flash sets a flash message with a unique ID to prevent stale timer collisions.
func (m *model) flash(text string) tea.Cmd {
	m.flashID++
	m.flashText = text
	return flashAfter(3*time.Second, m.flashID)
}

// initEditInputs populates edit mode text inputs from the spell.
func (e *spellEntry) initEditInputs() {
	sp := e.spell
	e.editInputs[editName] = newInput("Name", sp.Name, 30)
	e.editInputs[editDice] = newInput("Dice", sp.Dice.String(), 12)

	saveTypeVal := sp.SaveType
	if sp.Type == SpellTypeAttack {
		saveTypeVal = "Attack"
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
	e.editInputs[editHeightenDie] = newInput("+dice", hdStr, 4)

	hsStr := ""
	if sp.HeightenStep > 1 {
		hsStr = strconv.Itoa(sp.HeightenStep)
	}
	e.editInputs[editHeightenStep] = newInput("step", hsStr, 4)

	e.editFocus = editName
	e.editInputs[editName].Focus()
}

// cycleSaveType cycles the save type input through Reflex → Fortitude → Will → Attack.
func (e *spellEntry) cycleSaveType() {
	current := e.editInputs[editSaveType].Value()
	idx := 0
	for i, opt := range saveTypeCycle {
		if strings.EqualFold(opt, current) {
			idx = i
			break
		}
	}
	next := (idx + 1) % len(saveTypeCycle)
	e.editInputs[editSaveType].SetValue(saveTypeCycle[next])

	// Update multipliers when switching between save/attack
	newType := saveTypeCycle[next]
	if newType == "Attack" {
		e.editInputs[editMultBest].SetValue("2.0")
		e.editInputs[editMultGood].SetValue("1.0")
		e.editInputs[editMultBad].SetValue("0.0")
		e.editInputs[editMultWorst].SetValue("0.0")
	} else if strings.EqualFold(current, "Attack") {
		// Switching FROM attack TO save — set save defaults
		e.editInputs[editMultBest].SetValue("2.0")
		e.editInputs[editMultGood].SetValue("1.0")
		e.editInputs[editMultBad].SetValue("0.5")
		e.editInputs[editMultWorst].SetValue("0.0")
	}
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
	e.spell.HeightenStep = parseIntOr(e.editInputs[editHeightenStep].Value(), 0)
}

// isAoNURL returns true if the input is a valid Archives of Nethys URL.
func isAoNURL(input string) bool {
	u, err := url.Parse(input)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return u.Scheme == "https" &&
		(host == "2e.aonprd.com" || host == "aonprd.com")
}

// --- AoN Import Messages ---

type aonSpellMsg struct {
	spell Spell
}

type aonErrorMsg struct {
	err error
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
		m.flashID++
	case len(savedSpells) > 0:
		m.encounter = savedEnc
		m.encInputs = newEncounterInputs(savedEnc)
		for _, sp := range savedSpells {
			m.nextSpellID++
			entry := spellEntry{id: m.nextSpellID, spell: sp}
			entry.recalc(m.encounter)
			m.spells = append(m.spells, entry)
		}
		m.flashText = fmt.Sprintf("Loaded %d spell(s)", len(m.spells))
		m.flashID++
	}

	m.picker = newPickerModel(m.windowWidth, m.windowHeight)
	return m
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{textinput.Blink}
	if m.flashText != "" {
		cmds = append(cmds, flashAfter(3*time.Second, m.flashID))
	}
	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle messages that can arrive in any mode
	switch msg := msg.(type) {
	case clearFlashMsg:
		if msg.id == m.flashID {
			m.flashText = ""
		}
		return m, nil

	case pickerMsg:
		sp := msg.preset.ToSpell()
		m.addSpellEntry(sp)
		return m, m.flash(fmt.Sprintf("Added: %s", msg.preset.Name))

	case pickerCancelMsg:
		return m, nil

	case aonSpellMsg:
		m.addSpellEntry(msg.spell)
		m.spells[m.cursor].initEditInputs()
		m.mode = modeEdit
		return m, m.flash(fmt.Sprintf("Imported: %s", msg.spell.Name))

	case aonErrorMsg:
		m.flashID++
		m.flashText = fmt.Sprintf("Import failed: %v", msg.err)
		return m, flashAfter(5*time.Second, m.flashID)
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
	case modeNewSpell:
		return m.updateNewSpell(msg)
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
		case "ctrl+c", "q":
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
				id := m.spells[m.cursor].id
				if m.selected[id] {
					delete(m.selected, id)
				} else {
					m.selected[id] = true
				}
			}
			return m, nil

		case "a":
			m.picker = newPickerModel(m.windowWidth, m.windowHeight)
			m.picker.active = true
			return m, nil

		case "n":
			m.newSpellInput = newInput("AoN URL or spell name", "", 60)
			m.newSpellInput.Focus()
			m.mode = modeNewSpell
			return m, nil

		case "d":
			if len(m.spells) > 0 {
				delete(m.selected, m.spells[m.cursor].id)
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
				return m, m.flash(fmt.Sprintf("Save failed: %v", err))
			}
			return m, m.flash(fmt.Sprintf("Saved %d spell(s)!", len(m.spells)))
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

		case " ":
			// Space cycles save profiles on save fields
			if m.cycleSaveProfile() {
				m.recalcAll()
				return m, nil
			}
		}

		// Forward to active input
		var cmd tea.Cmd
		m.encInputs[m.encFocus], cmd = m.encInputs[m.encFocus].Update(msg)
		m.syncEncounterFromInputs()
		m.recalcAll()
		return m, cmd
	}

	// Forward non-key messages (e.g., blink) to the active input
	var cmd tea.Cmd
	m.encInputs[m.encFocus], cmd = m.encInputs[m.encFocus].Update(msg)
	return m, cmd
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
			m.mode = modeList
			return m, nil

		case "enter":
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

		case " ":
			// Space cycles save type when on that field
			if entry.editFocus == editSaveType {
				entry.cycleSaveType()
				return m, nil
			}
		}

		// Forward to active edit input
		var cmd tea.Cmd
		entry.editInputs[entry.editFocus], cmd = entry.editInputs[entry.editFocus].Update(msg)
		return m, cmd
	}

	// Forward non-key messages (e.g., blink) to the active input
	var cmd tea.Cmd
	entry.editInputs[entry.editFocus], cmd = entry.editInputs[entry.editFocus].Update(msg)
	return m, cmd
}

func (m model) updateNewSpell(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.mode = modeList
			return m, nil

		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			input := strings.TrimSpace(m.newSpellInput.Value())
			m.mode = modeList

			if isAoNURL(input) {
				// AoN URL — fetch asynchronously
				return m, fetchAoNSpell(input)
			}

			// Create new spell with the given name (or default)
			name := input
			if name == "" {
				name = fmt.Sprintf("Spell %d", len(m.spells)+1)
			}
			sp := NewSaveSpell(name)
			m.addSpellEntry(sp)
			m.spells[m.cursor].initEditInputs()
			m.mode = modeEdit
			m.blurEncFields()
			return m, nil
		}

		// Forward to input
		var cmd tea.Cmd
		m.newSpellInput, cmd = m.newSpellInput.Update(msg)
		return m, cmd
	}

	// Forward non-key messages (e.g., blink)
	var cmd tea.Cmd
	m.newSpellInput, cmd = m.newSpellInput.Update(msg)
	return m, cmd
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
	case modeNewSpell:
		return m.viewNewSpell()
	}

	return ""
}
