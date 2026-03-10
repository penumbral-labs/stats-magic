package main

import (
	"fmt"
	"strings"
)

// SpellPreset defines a preconfigured PF2e spell with accurate game data.
type SpellPreset struct {
	Name        string
	Rank        int
	Dice        string
	Type        SpellType
	SaveType    string // "Reflex", "Fortitude", "Will", or "" for attacks
	Multipliers DegreeMultipliers
	HeightenDie int // Extra dice per rank above base (0 = none)
	Description string
}

// AllPresets returns the full library of PF2e spell presets.
func AllPresets() []SpellPreset {
	return []SpellPreset{
		// --- Rank 1 ---
		{
			Name:        "Breathe Fire",
			Rank:        1,
			Dice:        "2d6",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "15-ft cone of fire, basic Reflex save",
		},
		{
			Name:        "Horizon Thunder Sphere (ranged)",
			Rank:        1,
			Dice:        "1d6+1",
			Type:        SpellTypeAttack,
			Multipliers: DefaultAttackMultipliers(),
			HeightenDie: 1,
			Description: "Ranged spell attack, electricity damage",
		},
		{
			Name:        "Horizon Thunder Sphere (melee)",
			Rank:        1,
			Dice:        "2d6+1",
			Type:        SpellTypeAttack,
			Multipliers: DefaultAttackMultipliers(),
			HeightenDie: 1,
			Description: "Melee spell attack, electricity damage",
		},
		{
			Name:        "Shocking Grasp",
			Rank:        1,
			Dice:        "2d12",
			Type:        SpellTypeAttack,
			Multipliers: DefaultAttackMultipliers(),
			HeightenDie: 1,
			Description: "Melee spell attack, electricity, +1 vs metal armor",
		},

		// --- Rank 2 ---
		{
			Name:        "Scorching Ray (1 ray)",
			Rank:        2,
			Dice:        "2d6",
			Type:        SpellTypeAttack,
			Multipliers: DefaultAttackMultipliers(),
			HeightenDie: 0,
			Description: "Ranged spell attack, fire; 2 rays at rank 2",
		},
		{
			Name:        "Acid Arrow",
			Rank:        2,
			Dice:        "3d8",
			Type:        SpellTypeAttack,
			Multipliers: DefaultAttackMultipliers(),
			HeightenDie: 1,
			Description: "Ranged spell attack, acid + persistent 1d6",
		},
		{
			Name:        "Sound Burst",
			Rank:        2,
			Dice:        "2d10",
			Type:        SpellTypeSave,
			SaveType:    "Fortitude",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "10-ft burst, basic Fortitude save, sonic",
		},

		// --- Rank 3 ---
		{
			Name:        "Fireball",
			Rank:        3,
			Dice:        "6d6",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 2,
			Description: "20-ft burst, basic Reflex save, fire",
		},
		{
			Name:        "Lightning Bolt",
			Rank:        3,
			Dice:        "4d12",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "120-ft line, basic Reflex save, electricity",
		},

		// --- Rank 4 ---
		{
			Name:        "Fire Shield",
			Rank:        4,
			Dice:        "2d6",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "Reactive fire damage when struck, basic Reflex",
		},
		{
			Name:        "Hydraulic Torrent",
			Rank:        4,
			Dice:        "8d6",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 2,
			Description: "60-ft line, basic Reflex save, bludgeoning",
		},

		// --- Rank 5 ---
		{
			Name:        "Cone of Cold",
			Rank:        5,
			Dice:        "12d6",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 2,
			Description: "60-ft cone, basic Reflex save, cold",
		},
		{
			Name:        "Blazing Fissure",
			Rank:        5,
			Dice:        "4d6",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "Line, basic Reflex save, fire (4d6 per 10 ft)",
		},

		// --- Rank 6 ---
		{
			Name:        "Chain Lightning",
			Rank:        6,
			Dice:        "8d12",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "Primary target + chain, basic Reflex, electricity",
		},

		// --- Rank 7 ---
		{
			Name:        "Sunburst",
			Rank:        7,
			Dice:        "8d10",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "60-ft burst, basic Reflex save, fire + good",
		},
		{
			Name:        "Eclipse Burst",
			Rank:        7,
			Dice:        "8d10",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "60-ft burst, basic Reflex save, negative",
		},

		// --- Rank 8 ---
		{
			Name:        "Polar Ray",
			Rank:        8,
			Dice:        "10d8",
			Type:        SpellTypeAttack,
			Multipliers: DefaultAttackMultipliers(),
			HeightenDie: 1,
			Description: "Ranged spell attack, cold + drained",
		},
		{
			Name:        "Horrid Wilting",
			Rank:        8,
			Dice:        "10d10",
			Type:        SpellTypeSave,
			SaveType:    "Fortitude",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "30-ft burst, basic Fortitude save, negative",
		},

		// --- Rank 9 ---
		{
			Name:        "Meteor Swarm (fire)",
			Rank:        9,
			Dice:        "14d6",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 0,
			Description: "4 meteors, 40-ft burst each, fire portion",
		},
		{
			Name:        "Meteor Swarm (bludgeoning)",
			Rank:        9,
			Dice:        "6d10",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 0,
			Description: "4 meteors, 40-ft burst each, bludgeoning portion",
		},

		// --- Rank 10 ---
		{
			Name:        "Cataclysm (fire)",
			Rank:        10,
			Dice:        "3d10",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 0,
			Description: "60-ft burst, fire portion of six damage types",
		},
	}
}

// PresetFilterString returns a searchable string for fuzzy matching.
func (p SpellPreset) PresetFilterString() string {
	return p.Name + " " + p.Description
}

// ToSpell converts a preset into a Spell with its intrinsic properties.
// Encounter-specific values (DC, save mods, AC) live in EncounterState.
func (p SpellPreset) ToSpell() Spell {
	return Spell{
		Name:        p.Name,
		Type:        p.Type,
		SaveType:    p.SaveType,
		Dice:        ParseDice(p.Dice),
		Multipliers: p.Multipliers,
		BaseRank:    p.Rank,
		HeightenDie: p.HeightenDie,
	}
}

// PickerLabel returns a formatted string for the picker display.
func (p SpellPreset) PickerLabel() string {
	parts := []string{p.Dice}
	if p.Type == SpellTypeSave {
		parts = append(parts, p.SaveType+" save")
	} else {
		parts = append(parts, "attack roll")
	}
	if p.HeightenDie > 0 {
		parts = append(parts, fmt.Sprintf("+%dd/rank", p.HeightenDie))
	}
	return strings.Join(parts, " | ")
}
