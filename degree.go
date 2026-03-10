package main

// Degree represents the four PF2e degrees of success, ordered from best outcome
// for the caster (most damage) to worst outcome for the caster (least damage).
//
//	Index 0: Best for caster  — Crit Fail on save / Crit Hit on attack  → 2x damage
//	Index 1: Good for caster  — Fail on save / Hit on attack            → 1x damage
//	Index 2: Bad for caster   — Success on save / Miss on attack        → 0.5x or 0x
//	Index 3: Worst for caster — Crit Success on save / Crit Miss on attack → 0x
type Degree int

const (
	DegreeBest    Degree = iota // Crit Fail save / Crit Hit attack
	DegreeGood                  // Fail save / Hit attack
	DegreeBad                   // Success save / Miss attack
	DegreeWorst                 // Crit Success save / Crit Miss attack
	DegreeCount                 // sentinel
)

// DegreeMultipliers holds the damage multiplier for each degree, ordered
// from best outcome for caster to worst.
type DegreeMultipliers struct {
	Best  float64 // Crit Fail save / Crit Hit attack (typically 2.0)
	Good  float64 // Fail save / Hit attack (typically 1.0)
	Bad   float64 // Success save / Miss attack (typically 0.5 or 0.0)
	Worst float64 // Crit Success save / Crit Miss attack (typically 0.0)
}

// AsSlice returns the multipliers in degree order (Best to Worst).
func (dm DegreeMultipliers) AsSlice() [4]float64 {
	return [4]float64{dm.Best, dm.Good, dm.Bad, dm.Worst}
}

// DefaultSaveMultipliers returns the standard PF2e multipliers for save-based spells.
// Ordered: Crit Fail (2x), Fail (1x), Success (0.5x), Crit Success (0x).
func DefaultSaveMultipliers() DegreeMultipliers {
	return DegreeMultipliers{
		Best:  2.0, // crit fail save
		Good:  1.0, // fail save
		Bad:   0.5, // success save
		Worst: 0.0, // crit success save
	}
}

// DefaultAttackMultipliers returns the standard PF2e multipliers for attack-based spells.
// Ordered: Crit Hit (2x), Hit (1x), Miss (0x), Crit Miss (0x).
func DefaultAttackMultipliers() DegreeMultipliers {
	return DegreeMultipliers{
		Best:  2.0, // crit hit
		Good:  1.0, // hit
		Bad:   0.0, // miss
		Worst: 0.0, // crit miss
	}
}

// DegreeLabels returns human-readable labels for each degree, ordered Best to Worst.
func DegreeLabels(st SpellType) [4]string {
	if st == SpellTypeSave {
		return [4]string{"Crit Fail", "Failure", "Success", "Crit Success"}
	}
	return [4]string{"Crit Hit", "Hit", "Miss", "Crit Miss"}
}

// DegreeProbabilities holds the probability of each degree, ordered Best to Worst.
type DegreeProbabilities [4]float64
