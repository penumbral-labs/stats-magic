package main

// SaveProfile represents the difficulty tier of an enemy's saving throw.
type SaveProfile int

const (
	SaveLow SaveProfile = iota
	SaveMod
	SaveHigh
)

func (p SaveProfile) String() string {
	switch p {
	case SaveLow:
		return "Low"
	case SaveHigh:
		return "High"
	default:
		return "Mod"
	}
}

// Next cycles to the next profile: Low → Mod → High → Low.
func (p SaveProfile) Next() SaveProfile {
	return (p + 1) % 3
}

// ParseSaveProfile converts a string to a SaveProfile.
func ParseSaveProfile(s string) SaveProfile {
	switch s {
	case "Low":
		return SaveLow
	case "High":
		return SaveHigh
	default:
		return SaveMod
	}
}

// EncounterState holds combat parameters derived from PC level and enemy profiles.
type EncounterState struct {
	PCLevel   int // PC level (1-20)
	SpellDC   int // Your spell DC
	AttackMod int // Your attack modifier
	EnemyAC   int // Enemy AC
	RefMod    int // Enemy Reflex save modifier
	FortMod   int // Enemy Fortitude save modifier
	WillMod   int // Enemy Will save modifier

	// Save difficulty profiles for the enemy (Low/Mod/High per save).
	RefProfile  SaveProfile
	FortProfile SaveProfile
	WillProfile SaveProfile
}

// DefaultEncounter returns encounter defaults for a level 5 PC vs moderate enemies.
func DefaultEncounter() EncounterState {
	return EncounterForLevel(5)
}

// EncounterForLevel creates an encounter with all moderate enemy profiles at the given PC level.
func EncounterForLevel(level int) EncounterState {
	enc := EncounterState{
		PCLevel:     level,
		RefProfile:  SaveMod,
		FortProfile: SaveMod,
		WillProfile: SaveMod,
	}
	enc.RecalcFromLevel()
	return enc
}

// RecalcFromLevel recomputes all derived values from PCLevel and save profiles.
func (e *EncounterState) RecalcFromLevel() {
	e.SpellDC = PCSpellDC(e.PCLevel)
	e.AttackMod = PCAttackMod(e.PCLevel)
	e.EnemyAC = MonsterAC(e.PCLevel)
	e.RefMod = MonsterSave(e.PCLevel, e.RefProfile)
	e.FortMod = MonsterSave(e.PCLevel, e.FortProfile)
	e.WillMod = MonsterSave(e.PCLevel, e.WillProfile)
}

// SaveModFor returns the appropriate enemy save modifier for the given save type.
func (e EncounterState) SaveModFor(saveType string) int {
	switch saveType {
	case "Fortitude":
		return e.FortMod
	case "Will":
		return e.WillMod
	default:
		return e.RefMod
	}
}

// --- PF2e Lookup Tables ---
//
// PC stats: based on Remaster primary caster progression.
// Trained (1-6), Expert (7-14), Master (15-18), Legendary (19+).
// Key ability: 18 at 1, +2 at 5/10/15/20 → 20/22/24/26 → mods +5/+6/+7/+8.
//
// Monster stats: based on GM Core Table 2-5 (AC) and Table 2-6 (Saving Throws).

// pcSpellDCTable: spell DC by PC level for a primary caster.
// DC = 10 + proficiency + ability_mod + level.
var pcSpellDCTable = [21]int{
	0,              // level 0 (unused)
	17, 18, 19, 20, // levels 1-4:  trained +2, ability +4
	22, 23, // levels 5-6:  trained +2, ability +5
	26, 27, 28, // levels 7-9:  expert +4, ability +5
	30, 31, 32, 33, 34, // levels 10-14: expert +4, ability +6
	38, 39, 40, 41, // levels 15-18: master +6, ability +7
	44, 46, // levels 19-20: legendary +8, ability +7/+8
}

// PCSpellDC returns the expected spell DC for a primary caster at the given level.
func PCSpellDC(level int) int {
	if level < 1 {
		return pcSpellDCTable[1]
	}
	if level > 20 {
		return pcSpellDCTable[20]
	}
	return pcSpellDCTable[level]
}

// PCAttackMod returns the expected spell attack modifier for a primary caster.
func PCAttackMod(level int) int {
	return PCSpellDC(level) - 10
}

// monsterACTable: moderate AC by creature level (GM Core Table 2-5).
var monsterACTable = [21]int{
	0,
	16, 18, 19, 21, 22, 24, 25, 27, 28, 30,
	31, 33, 34, 36, 37, 39, 40, 42, 43, 45,
}

// MonsterAC returns the moderate AC for a creature at the given level.
func MonsterAC(level int) int {
	if level < 1 {
		return monsterACTable[1]
	}
	if level > 20 {
		return monsterACTable[20]
	}
	return monsterACTable[level]
}

// monsterSaveTable: save modifiers by creature level, indexed [level][profile].
// Values from GM Core Table 2-6 (Saving Throws) for Low, Moderate, High.
var monsterSaveTable = [21][3]int{
	{0, 0, 0},    // level 0 (unused)
	{4, 7, 10},   // level 1
	{5, 8, 11},   // level 2
	{6, 9, 12},   // level 3
	{8, 11, 14},  // level 4
	{9, 12, 15},  // level 5
	{11, 14, 17}, // level 6
	{12, 15, 18}, // level 7
	{13, 16, 19}, // level 8
	{15, 18, 21}, // level 9
	{16, 19, 22}, // level 10
	{18, 21, 24}, // level 11
	{19, 22, 25}, // level 12
	{20, 23, 26}, // level 13
	{22, 25, 28}, // level 14
	{23, 26, 29}, // level 15
	{25, 28, 30}, // level 16
	{26, 29, 32}, // level 17
	{27, 30, 33}, // level 18
	{29, 32, 35}, // level 19
	{30, 33, 36}, // level 20
}

// MonsterSave returns the save modifier for a creature at the given level and profile.
func MonsterSave(level int, profile SaveProfile) int {
	if level < 1 {
		return monsterSaveTable[1][profile]
	}
	if level > 20 {
		return monsterSaveTable[20][profile]
	}
	return monsterSaveTable[level][profile]
}
