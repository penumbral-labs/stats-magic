package main

// EncounterState holds the shared combat parameters: your character's spell DC
// and attack modifier, plus the enemy's save modifiers and AC.
type EncounterState struct {
	SpellDC   int // Your spell DC
	AttackMod int // Your attack modifier
	RefMod    int // Enemy Reflex save modifier
	FortMod   int // Enemy Fortitude save modifier
	WillMod   int // Enemy Will save modifier
	EnemyAC   int // Enemy AC
}

// DefaultEncounter returns sensible defaults for a mid-level encounter.
func DefaultEncounter() EncounterState {
	return EncounterState{
		SpellDC:   30,
		AttackMod: 20,
		RefMod:    14,
		FortMod:   16,
		WillMod:   12,
		EnemyAC:   28,
	}
}

// SaveModFor returns the appropriate enemy save modifier for the given save type.
func (e EncounterState) SaveModFor(saveType string) int {
	switch saveType {
	case "Fortitude":
		return e.FortMod
	case "Will":
		return e.WillMod
	default:
		return e.RefMod // Reflex is the most common default
	}
}
