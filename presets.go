package main

import (
	"fmt"
	"strings"
)

// SpellPreset defines a preconfigured PF2e spell with accurate Remaster data.
// All data sourced from Archives of Nethys (2e.aonprd.com), Player Core / Player Core 2.
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

// AllPresets returns the full library of PF2e Remaster spell presets.
func AllPresets() []SpellPreset {
	return []SpellPreset{
		// =====================================================================
		// Cantrips (base rank 1, scale with character level)
		// =====================================================================
		{
			Name:        "Caustic Blast",
			Rank:        1,
			Dice:        "1d8",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 0, // +1d8 per 2 ranks (not per rank)
			Description: "5-ft burst, 30 ft, basic Reflex, acid",
		},
		{
			Name:        "Divine Lance",
			Rank:        1,
			Dice:        "2d4",
			Type:        SpellTypeAttack,
			Multipliers: DefaultAttackMultipliers(),
			HeightenDie: 1,
			Description: "Spell attack, 60 ft, spirit damage, sanctified",
		},
		{
			Name:        "Electric Arc",
			Rank:        1,
			Dice:        "2d4",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "1-2 targets, 30 ft, basic Reflex, electricity",
		},
		{
			Name:        "Frostbite",
			Rank:        1,
			Dice:        "2d4",
			Type:        SpellTypeSave,
			SaveType:    "Fortitude",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "60 ft, basic Fortitude, cold",
		},
		{
			Name:        "Gouging Claw",
			Rank:        1,
			Dice:        "2d6",
			Type:        SpellTypeAttack,
			Multipliers: DefaultAttackMultipliers(),
			HeightenDie: 1,
			Description: "Melee spell attack, slashing/piercing + bleed",
		},
		{
			Name:        "Ignition",
			Rank:        1,
			Dice:        "2d4",
			Type:        SpellTypeAttack,
			Multipliers: DefaultAttackMultipliers(),
			HeightenDie: 1,
			Description: "Spell attack, fire; persistent 1d4 on crit",
		},
		{
			Name:        "Scatter Scree",
			Rank:        1,
			Dice:        "2d4",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "Two 5-ft cubes, 30 ft, basic Reflex, bludgeoning",
		},
		{
			Name:        "Vitality Lash",
			Rank:        1,
			Dice:        "2d6",
			Type:        SpellTypeSave,
			SaveType:    "Fortitude",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "30 ft, basic Fort, vitality; undead/void only",
		},

		// =====================================================================
		// Rank 1
		// =====================================================================
		{
			Name:        "Breathe Fire",
			Rank:        1,
			Dice:        "2d6",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 2,
			Description: "15-ft cone, basic Reflex, fire",
		},
		{
			Name:        "Chilling Spray",
			Rank:        1,
			Dice:        "2d4",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 2,
			Description: "15-ft cone, Reflex, cold; fail: slowed 1",
		},
		{
			Name:        "Force Barrage",
			Rank:        1,
			Dice:        "1d4+1",
			Type:        SpellTypeAttack,
			Multipliers: DegreeMultipliers{Best: 1, Good: 1, Bad: 1, Worst: 1}, // auto-hit, no degrees
			HeightenDie: 0, // +1 missile per 2 ranks
			Description: "120 ft, auto-hit, force; 1-3 missiles by actions",
		},
		{
			Name:        "Harm",
			Rank:        1,
			Dice:        "1d8",
			Type:        SpellTypeSave,
			SaveType:    "Fortitude",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "Touch/30 ft/30-ft emanation, basic Fort, void",
		},
		{
			Name:        "Noxious Vapors",
			Rank:        1,
			Dice:        "1d6",
			Type:        SpellTypeSave,
			SaveType:    "Fortitude",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "10-ft emanation, basic Fort, poison; sickened on crit fail",
		},
		{
			Name:        "Phantom Pain",
			Rank:        1,
			Dice:        "2d4",
			Type:        SpellTypeSave,
			SaveType:    "Will",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 2,
			Description: "30 ft, Will, mental; +1d4 persistent, sickened on fail",
		},
		{
			Name:        "Pummeling Rubble",
			Rank:        1,
			Dice:        "2d4",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 2,
			Description: "15-ft cone, Reflex, bludgeoning; knockback on fail",
		},
		{
			Name:        "Thunderstrike",
			Rank:        1,
			Dice:        "1d12",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "120 ft, basic Reflex, elec+sonic; -1 save vs metal",
		},

		// =====================================================================
		// Rank 2
		// =====================================================================
		{
			Name:        "Acid Grip",
			Rank:        2,
			Dice:        "2d8",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "Basic Reflex, acid + persistent 1d6; -10 ft speed",
		},
		{
			Name:        "Blazing Bolt",
			Rank:        2,
			Dice:        "2d6",
			Type:        SpellTypeAttack,
			Multipliers: DefaultAttackMultipliers(),
			HeightenDie: 1,
			Description: "Spell attack, fire; 1 ray per action (up to 3)",
		},
		{
			Name:        "Ice Storm",
			Rank:        2,
			Dice:        "2d8",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 0, // +1d8 per 2 ranks, not per rank
			Description: "20-ft burst, basic Reflex, bludgeoning+cold; sustained",
		},
		{
			Name:        "Noise Blast",
			Rank:        2,
			Dice:        "2d10",
			Type:        SpellTypeSave,
			SaveType:    "Fortitude",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "10-ft burst, basic Fort, sonic; deafened on fail",
		},

		// =====================================================================
		// Rank 3
		// =====================================================================
		{
			Name:        "Chilling Darkness",
			Rank:        3,
			Dice:        "5d6",
			Type:        SpellTypeAttack,
			Multipliers: DefaultAttackMultipliers(),
			HeightenDie: 2,
			Description: "Spell attack, 120 ft, cold (+spirit vs holy)",
		},
		{
			Name:        "Fireball",
			Rank:        3,
			Dice:        "6d6",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 2,
			Description: "20-ft burst, basic Reflex, fire",
		},
		{
			Name:        "Holy Light",
			Rank:        3,
			Dice:        "5d6",
			Type:        SpellTypeAttack,
			Multipliers: DefaultAttackMultipliers(),
			HeightenDie: 2,
			Description: "Spell attack, 120 ft, fire (+spirit vs unholy)",
		},
		{
			Name:        "Lightning Bolt",
			Rank:        3,
			Dice:        "4d12",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "120-ft line, basic Reflex, electricity",
		},
		{
			Name:        "Vampiric Feast",
			Rank:        3,
			Dice:        "6d6",
			Type:        SpellTypeSave,
			SaveType:    "Fortitude",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 2,
			Description: "Touch, basic Fort, void; heals half damage dealt",
		},

		// =====================================================================
		// Rank 4
		// =====================================================================
		{
			Name:        "Divine Wrath",
			Rank:        4,
			Dice:        "4d10",
			Type:        SpellTypeSave,
			SaveType:    "Fortitude",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "20-ft burst, 120 ft, Fort, spirit; sickened on fail",
		},
		{
			Name:        "Hydraulic Torrent",
			Rank:        4,
			Dice:        "8d6",
			Type:        SpellTypeSave,
			SaveType:    "Fortitude",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 2,
			Description: "60-ft line, basic Fort, bludgeoning; knockback",
		},
		{
			Name:        "Vision of Death",
			Rank:        4,
			Dice:        "8d6",
			Type:        SpellTypeSave,
			SaveType:    "Will",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 2,
			Description: "Basic Will, mental; frightened on fail, death at 0 HP",
		},

		// =====================================================================
		// Rank 5
		// =====================================================================
		{
			Name:        "Howling Blizzard",
			Rank:        5,
			Dice:        "10d6",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 2,
			Description: "60-ft cone, basic Reflex, cold; difficult terrain",
		},
		{
			Name:        "Impaling Spike",
			Rank:        5,
			Dice:        "8d6",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 2,
			Description: "30 ft, Reflex, piercing (cold iron); immobilized on fail",
		},
		{
			Name:        "Shadow Blast",
			Rank:        5,
			Dice:        "6d8",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "Cone/burst/line, basic save, choose damage type",
		},

		// =====================================================================
		// Rank 6
		// =====================================================================
		{
			Name:        "Chain Lightning",
			Rank:        6,
			Dice:        "8d12",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "Chains to nearby targets, basic Reflex, electricity",
		},
		{
			Name:        "Disintegrate",
			Rank:        6,
			Dice:        "12d10",
			Type:        SpellTypeAttack,
			Multipliers: DefaultAttackMultipliers(),
			HeightenDie: 2,
			Description: "Spell attack → basic Fort, untyped; destroys at 0 HP",
		},
		{
			Name:        "Visions of Danger",
			Rank:        6,
			Dice:        "8d8",
			Type:        SpellTypeSave,
			SaveType:    "Will",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "30-ft burst, 500 ft, basic Will, mental; illusion",
		},

		// =====================================================================
		// Rank 7
		// =====================================================================
		{
			Name:        "Eclipse Burst",
			Rank:        7,
			Dice:        "8d10",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "60-ft burst, basic Reflex, cold+void; blinds on crit fail",
		},
		{
			Name:        "Sunburst",
			Rank:        7,
			Dice:        "8d10",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "60-ft burst, basic Reflex, fire+vitality; blinds on crit fail",
		},

		// =====================================================================
		// Rank 8
		// =====================================================================
		{
			Name:        "Arctic Rift",
			Rank:        8,
			Dice:        "12d8",
			Type:        SpellTypeSave,
			SaveType:    "Fortitude",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "120-ft line, basic Fort, cold; immobilizes on crit fail",
		},
		{
			Name:        "Desiccate",
			Rank:        8,
			Dice:        "10d10",
			Type:        SpellTypeSave,
			SaveType:    "Fortitude",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "500 ft, basic Fort, void; plants/water creatures worse",
		},

		// =====================================================================
		// Rank 9
		// =====================================================================
		{
			Name:        "Falling Stars (fire)",
			Rank:        9,
			Dice:        "14d6",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 0,
			Description: "4 meteors, 40-ft bursts, basic Reflex, fire portion",
		},
		{
			Name:        "Falling Stars (bludg.)",
			Rank:        9,
			Dice:        "6d10",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 0,
			Description: "4 meteors, 40-ft bursts, basic Reflex, bludg. portion",
		},
		{
			Name:        "Massacre",
			Rank:        9,
			Dice:        "9d6",
			Type:        SpellTypeSave,
			SaveType:    "Fortitude",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 1,
			Description: "60-ft line, Fort, void; 100 dmg on fail, death on crit fail",
		},

		// =====================================================================
		// Rank 10
		// =====================================================================
		{
			Name:        "Cataclysm",
			Rank:        10,
			Dice:        "15d10",
			Type:        SpellTypeSave,
			SaveType:    "Reflex",
			Multipliers: DefaultSaveMultipliers(),
			HeightenDie: 0,
			Description: "60-ft burst, basic Reflex, 5 damage types; resistances -10",
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
