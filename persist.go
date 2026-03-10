package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const configDirName = "stats-magic"
const configFileName = "spells.json"

// savedSpell is the JSON-serializable form of a spell configuration.
type savedSpell struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"` // "save" or "attack"
	SaveType    string  `json:"save_type,omitempty"`
	Dice        string  `json:"dice"`
	MultBest    float64 `json:"mult_best"`
	MultGood    float64 `json:"mult_good"`
	MultBad     float64 `json:"mult_bad"`
	MultWorst   float64 `json:"mult_worst"`
	BaseRank     int     `json:"base_rank,omitempty"`
	HeightenDie  int     `json:"heighten_die,omitempty"`
	HeightenStep int     `json:"heighten_step,omitempty"`
}

// savedEncounter is the JSON-serializable form of the encounter state.
type savedEncounter struct {
	PCLevel     int    `json:"pc_level,omitempty"`
	SpellDC     int    `json:"spell_dc"`
	AttackMod   int    `json:"attack_mod"`
	RefMod      int    `json:"reflex_mod"`
	FortMod     int    `json:"fortitude_mod"`
	WillMod     int    `json:"will_mod"`
	EnemyAC     int    `json:"enemy_ac"`
	RefProfile  string `json:"ref_profile,omitempty"`
	FortProfile string `json:"fort_profile,omitempty"`
	WillProfile string `json:"will_profile,omitempty"`
}

// savedData is the top-level JSON structure.
type savedData struct {
	Encounter savedEncounter `json:"encounter"`
	Spells    []savedSpell   `json:"spells"`
}

// clearFlashMsg signals that the flash message should be cleared.
// The id field prevents a stale timer from clearing a newer message.
type clearFlashMsg struct {
	id int
}

func flashAfter(d time.Duration, id int) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return clearFlashMsg{id: id}
	})
}

func configPath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cfgDir, configDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

func saveData(enc EncounterState, spells []spellEntry) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	data := savedData{
		Encounter: savedEncounter{
			PCLevel:     enc.PCLevel,
			SpellDC:     enc.SpellDC,
			AttackMod:   enc.AttackMod,
			RefMod:      enc.RefMod,
			FortMod:     enc.FortMod,
			WillMod:     enc.WillMod,
			EnemyAC:     enc.EnemyAC,
			RefProfile:  enc.RefProfile.String(),
			FortProfile: enc.FortProfile.String(),
			WillProfile: enc.WillProfile.String(),
		},
	}

	for _, e := range spells {
		sp := e.spell
		typeStr := "save"
		if sp.Type == SpellTypeAttack {
			typeStr = "attack"
		}
		data.Spells = append(data.Spells, savedSpell{
			Name:         sp.Name,
			Type:         typeStr,
			SaveType:     sp.SaveType,
			Dice:         sp.Dice.String(),
			MultBest:     sp.Multipliers.Best,
			MultGood:     sp.Multipliers.Good,
			MultBad:      sp.Multipliers.Bad,
			MultWorst:    sp.Multipliers.Worst,
			BaseRank:     sp.BaseRank,
			HeightenDie:  sp.HeightenDie,
			HeightenStep: sp.HeightenStep,
		})
	}

	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, raw, 0o644)
}

func loadData() (EncounterState, []Spell, error) {
	path, err := configPath()
	if err != nil {
		return EncounterState{}, nil, err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return EncounterState{}, nil, err
	}

	var data savedData
	if err := json.Unmarshal(raw, &data); err != nil {
		return EncounterState{}, nil, err
	}

	enc := EncounterState{
		PCLevel:     data.Encounter.PCLevel,
		SpellDC:     data.Encounter.SpellDC,
		AttackMod:   data.Encounter.AttackMod,
		RefMod:      data.Encounter.RefMod,
		FortMod:     data.Encounter.FortMod,
		WillMod:     data.Encounter.WillMod,
		EnemyAC:     data.Encounter.EnemyAC,
		RefProfile:  ParseSaveProfile(data.Encounter.RefProfile),
		FortProfile: ParseSaveProfile(data.Encounter.FortProfile),
		WillProfile: ParseSaveProfile(data.Encounter.WillProfile),
	}

	// Backwards compatibility: old saves without PCLevel
	if enc.PCLevel == 0 {
		if enc.SpellDC == 0 && enc.AttackMod == 0 && enc.EnemyAC == 0 {
			enc = DefaultEncounter()
		} else {
			// Infer level from spell DC
			enc.PCLevel = inferLevelFromDC(enc.SpellDC)
		}
	}

	var spells []Spell
	for _, s := range data.Spells {
		st := SpellTypeSave
		if s.Type == "attack" {
			st = SpellTypeAttack
		}

		sp := Spell{
			Name:     s.Name,
			Type:     st,
			SaveType: s.SaveType,
			Dice:     ParseDice(s.Dice),
			Multipliers: DegreeMultipliers{
				Best:  s.MultBest,
				Good:  s.MultGood,
				Bad:   s.MultBad,
				Worst: s.MultWorst,
			},
			BaseRank:     s.BaseRank,
			HeightenDie:  s.HeightenDie,
			HeightenStep: s.HeightenStep,
		}

		if sp.Type == SpellTypeSave && sp.SaveType == "" {
			sp.SaveType = "Reflex"
		}

		spells = append(spells, sp)
	}

	return enc, spells, nil
}

// inferLevelFromDC finds the closest PC level for a given spell DC.
func inferLevelFromDC(dc int) int {
	best := 1
	bestDist := abs(pcSpellDCTable[1] - dc)
	for lvl := 2; lvl <= 20; lvl++ {
		d := abs(pcSpellDCTable[lvl] - dc)
		if d < bestDist {
			bestDist = d
			best = lvl
		}
	}
	return best
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
