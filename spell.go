package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SpellType distinguishes save-based spells from attack-based spells.
type SpellType int

const (
	SpellTypeSave SpellType = iota
	SpellTypeAttack
)

func (st SpellType) String() string {
	if st == SpellTypeSave {
		return "Save-based"
	}
	return "Attack-based"
}

// DiceFormula represents a parsed NdS+B dice expression.
type DiceFormula struct {
	Count int // N - number of dice
	Sides int // S - sides per die
	Bonus int // B - flat modifier
}

// String returns the canonical dice notation.
func (d DiceFormula) String() string {
	s := fmt.Sprintf("%dd%d", d.Count, d.Sides)
	if d.Bonus > 0 {
		s += fmt.Sprintf("+%d", d.Bonus)
	} else if d.Bonus < 0 {
		s += fmt.Sprintf("%d", d.Bonus)
	}
	return s
}

// Valid returns true if the formula has at least one die with at least two sides.
func (d DiceFormula) Valid() bool {
	return d.Count > 0 && d.Sides >= 2
}

var diceRegex = regexp.MustCompile(`^(\d+)d(\d+)(?:([+-])(\d+))?$`)

// ParseDice parses a dice notation string like "8d6", "4d10+4", or "2d8-1".
// Returns a zero DiceFormula if the input is invalid.
func ParseDice(s string) DiceFormula {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	m := diceRegex.FindStringSubmatch(s)
	if m == nil {
		return DiceFormula{}
	}

	count, _ := strconv.Atoi(m[1])
	sides, _ := strconv.Atoi(m[2])
	bonus := 0
	if m[4] != "" {
		bonus, _ = strconv.Atoi(m[4])
		if m[3] == "-" {
			bonus = -bonus
		}
	}

	return DiceFormula{Count: count, Sides: sides, Bonus: bonus}
}

// Spell holds the intrinsic properties of a spell. Encounter-specific values
// (DC, save mods, AC) live in EncounterState instead.
type Spell struct {
	Name        string
	Type        SpellType
	SaveType    string // "Reflex", "Fortitude", "Will", "" for attacks
	Dice        DiceFormula
	Multipliers DegreeMultipliers
	BaseRank    int // Spell rank (1-10), 0 means unset
	HeightenDie int // Extra dice per heighten step above base (0 = no heightening)
	HeightenStep int // Ranks per heighten step (1 = every rank, 2 = every 2 ranks; 0 defaults to 1)
}

// NewSaveSpell creates a new save-based spell with default multipliers.
func NewSaveSpell(name string) Spell {
	return Spell{
		Name:        name,
		Type:        SpellTypeSave,
		SaveType:    "Reflex",
		Multipliers: DefaultSaveMultipliers(),
	}
}

// NewAttackSpell creates a new attack-based spell with default multipliers.
func NewAttackSpell(name string) Spell {
	return Spell{
		Name:        name,
		Type:        SpellTypeAttack,
		Multipliers: DefaultAttackMultipliers(),
	}
}

// effectiveHeightenStep returns the heighten step, defaulting to 1.
func (s *Spell) effectiveHeightenStep() int {
	if s.HeightenStep > 0 {
		return s.HeightenStep
	}
	return 1
}

// EffectiveDice returns the dice formula adjusted for heightening.
// If BaseRank is set and HeightenDie > 0, adds extra dice per heighten step above base.
func (s *Spell) EffectiveDice(rank int) DiceFormula {
	d := s.Dice
	if s.BaseRank > 0 && s.HeightenDie > 0 && rank > s.BaseRank {
		steps := (rank - s.BaseRank) / s.effectiveHeightenStep()
		d.Count += steps * s.HeightenDie
	}
	return d
}

// DefenseLabel returns a short description of what defense this spell targets.
func (s *Spell) DefenseLabel() string {
	if s.Type == SpellTypeAttack {
		return "Attack"
	}
	return saveAbbrev(s.SaveType) + " save"
}

// ShortDefenseLabel returns a compact defense label for tight layouts.
func (s *Spell) ShortDefenseLabel() string {
	if s.Type == SpellTypeAttack {
		return "Atk"
	}
	return saveAbbrev(s.SaveType)
}

// saveAbbrev returns a 3-letter abbreviation for a save type.
func saveAbbrev(saveType string) string {
	if len(saveType) >= 3 {
		return saveType[:3]
	}
	if saveType == "" {
		return "Ref"
	}
	return saveType
}
