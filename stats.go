package main

import "math"

// --- Dice Statistics (CLT / Normal Approximation) ---

// DiceMean returns the expected value of NdS+B.
// Mean of NdS = N * (S+1) / 2, plus the bonus.
func DiceMean(d DiceFormula) float64 {
	if !d.Valid() {
		return float64(d.Bonus)
	}
	return float64(d.Count)*float64(d.Sides+1)/2.0 + float64(d.Bonus)
}

// DiceVariance returns the variance of NdS+B.
// Variance of one dS = (S^2 - 1) / 12, and N independent dice sum linearly.
// The bonus is constant and contributes zero variance.
func DiceVariance(d DiceFormula) float64 {
	if !d.Valid() {
		return 0
	}
	s := float64(d.Sides)
	return float64(d.Count) * (s*s - 1) / 12.0
}

// DiceStdDev returns the standard deviation of NdS+B.
func DiceStdDev(d DiceFormula) float64 {
	return math.Sqrt(DiceVariance(d))
}

// --- Normal Distribution Helpers ---

// normalPDF returns the probability density function of a standard normal at x.
func normalPDF(x float64) float64 {
	return math.Exp(-x*x/2) / math.Sqrt(2*math.Pi)
}

// normalCDF returns the cumulative distribution function of a standard normal at x
// using the Abramowitz & Stegun approximation (maximum error ~1.5e-7).
func normalCDF(x float64) float64 {
	if x < -8 {
		return 0
	}
	if x > 8 {
		return 1
	}

	neg := x < 0
	if neg {
		x = -x
	}

	const (
		b1 = 0.319381530
		b2 = -0.356563782
		b3 = 1.781477937
		b4 = -1.821255978
		b5 = 1.330274429
		p  = 0.2316419
	)

	t := 1.0 / (1.0 + p*x)
	y := 1.0 - normalPDF(x)*(b1*t+b2*t*t+b3*t*t*t+b4*t*t*t*t+b5*t*t*t*t*t)

	if neg {
		return 1.0 - y
	}
	return y
}

// scaledNormalPDF returns the PDF of a normal distribution with given mean and stddev,
// evaluated at x.
func scaledNormalPDF(x, mean, stddev float64) float64 {
	if stddev <= 0 {
		return 0
	}
	z := (x - mean) / stddev
	return normalPDF(z) / stddev
}

// --- PF2e Degree of Success Probabilities ---

// CalcSaveDegreeProbabilities calculates the probability of each degree of success
// for a save-based spell. Returns probabilities in unified degree order (Best to Worst
// for the caster).
//
// The enemy rolls d20 + saveMod against spellDC.
// PF2e rules:
//   - Critical Success: total >= DC + 10
//   - Success: total >= DC
//   - Failure: total < DC
//   - Critical Failure: total <= DC - 10
//   - Natural 20 upgrades one step; natural 1 downgrades one step.
//
// Degree order (best for caster first):
//
//	[0] Best  = Crit Fail save  (enemy rolls terribly)
//	[1] Good  = Fail save       (enemy fails)
//	[2] Bad   = Success save    (enemy succeeds)
//	[3] Worst = Crit Success save (enemy crits the save)
func CalcSaveDegreeProbabilities(spellDC, saveMod int) DegreeProbabilities {
	var probs DegreeProbabilities

	for roll := 1; roll <= 20; roll++ {
		total := roll + saveMod

		// Determine base degree from the enemy's perspective (save result)
		// 0=crit fail, 1=fail, 2=success, 3=crit success
		saveDegree := baseSaveDegree(total, spellDC)

		// Natural 20 upgrades the save, natural 1 downgrades
		if roll == 20 {
			saveDegree = clampDegree(saveDegree + 1)
		} else if roll == 1 {
			saveDegree = clampDegree(saveDegree - 1)
		}

		// Save degree maps directly to caster degree:
		// saveDegree 0 (crit fail) = casterDegree 0 (Best for caster)
		// saveDegree 3 (crit success) = casterDegree 3 (Worst for caster)
		casterDegree := saveDegree

		p := 1.0 / 20.0
		probs[casterDegree] += p
	}

	return probs
}

// CalcAttackDegreeProbabilities calculates the probability of each degree of success
// for an attack-based spell. Returns probabilities in unified degree order (Best to Worst
// for the caster).
//
// The caster rolls d20 + attackMod against enemyAC.
// PF2e rules:
//   - Critical Hit: total >= AC + 10
//   - Hit: total >= AC
//   - Miss: total < AC
//   - Critical Miss: total <= AC - 10
//   - Natural 20 upgrades one step; natural 1 downgrades one step.
//
// Degree order (best for caster first):
//
//	[0] Best  = Crit Hit   (caster crits)
//	[1] Good  = Hit        (caster hits)
//	[2] Bad   = Miss       (caster misses)
//	[3] Worst = Crit Miss  (caster fumbles)
func CalcAttackDegreeProbabilities(attackMod, enemyAC int) DegreeProbabilities {
	var probs DegreeProbabilities

	for roll := 1; roll <= 20; roll++ {
		total := roll + attackMod

		// Base attack degree: 0=crit miss, 1=miss, 2=hit, 3=crit hit
		attackDegree := baseAttackDegree(total, enemyAC)

		// Natural 20 upgrades, natural 1 downgrades
		if roll == 20 {
			attackDegree = clampDegree(attackDegree + 1)
		} else if roll == 1 {
			attackDegree = clampDegree(attackDegree - 1)
		}

		// Map to caster degree: crit hit (3) = Best (0), crit miss (0) = Worst (3)
		casterDegree := 3 - attackDegree

		p := 1.0 / 20.0
		probs[casterDegree] += p
	}

	return probs
}

// baseSaveDegree converts a save total vs DC into a save degree (from the enemy's perspective).
// 0 = crit fail, 1 = fail, 2 = success, 3 = crit success.
func baseSaveDegree(total, dc int) int {
	switch {
	case total >= dc+10:
		return 3 // crit success
	case total >= dc:
		return 2 // success
	case total <= dc-10:
		return 0 // crit failure
	default:
		return 1 // failure
	}
}

// baseAttackDegree converts an attack total vs AC into an attack degree.
// 0 = crit miss, 1 = miss, 2 = hit, 3 = crit hit.
func baseAttackDegree(total, ac int) int {
	switch {
	case total >= ac+10:
		return 3 // crit hit
	case total >= ac:
		return 2 // hit
	case total <= ac-10:
		return 0 // crit miss
	default:
		return 1 // miss
	}
}

func clampDegree(d int) int {
	if d < 0 {
		return 0
	}
	if d > 3 {
		return 3
	}
	return d
}

// --- Spell Damage Statistics ---

// SpellStats holds the computed damage statistics for a spell.
type SpellStats struct {
	BaseMean   float64 // Mean of the raw dice roll
	BaseStdDev float64 // StdDev of the raw dice roll

	DegreeProb [4]float64 // Probability of each degree (Best to Worst)
	DegreeMean [4]float64 // Expected damage per degree

	ExpectedDamage float64 // Overall weighted expected damage
	OverallStdDev  float64 // Overall standard deviation across all degrees
	AnyDamageProb  float64 // Probability of dealing any damage at all

	// Weighted mixture distribution samples for visualization.
	MixturePDF []float64 // Sampled PDF values across the damage range
	MixtureLo  float64   // Low end of the sample range
	MixtureHi  float64   // High end of the sample range

	// Heightening table: expected damage at each rank from BaseRank to 10.
	HeightenTable []HeightenRow
}

// HeightenRow holds the expected damage for a spell at a given rank.
type HeightenRow struct {
	Rank     int
	Dice     string
	Mean     float64
	Expected float64
}

// CalcSpellStats computes full damage statistics for a spell at its base rank.
func CalcSpellStats(spell Spell, enc EncounterState) SpellStats {
	return CalcSpellStatsAtRank(spell, enc, 0)
}

// CalcSpellStatsAtRank computes damage statistics for a spell at a specific rank.
// If rank is 0, uses the spell's base dice.
func CalcSpellStatsAtRank(spell Spell, enc EncounterState, rank int) SpellStats {
	var st SpellStats

	dice := spell.Dice
	if rank > 0 && spell.BaseRank > 0 && spell.HeightenDie > 0 {
		dice = spell.EffectiveDice(rank)
	}

	if !dice.Valid() {
		return st
	}

	st.BaseMean = DiceMean(dice)
	st.BaseStdDev = DiceStdDev(dice)

	// Compute degree probabilities using encounter state
	if spell.Type == SpellTypeSave {
		saveMod := enc.SaveModFor(spell.SaveType)
		st.DegreeProb = CalcSaveDegreeProbabilities(enc.SpellDC, saveMod)
	} else {
		st.DegreeProb = CalcAttackDegreeProbabilities(enc.AttackMod, enc.EnemyAC)
	}

	multipliers := spell.Multipliers.AsSlice()

	// Expected damage per degree = baseMean * multiplier
	for i := 0; i < 4; i++ {
		st.DegreeMean[i] = st.BaseMean * multipliers[i]
	}

	// Overall expected damage = sum(prob_i * mean_i)
	for i := 0; i < 4; i++ {
		st.ExpectedDamage += st.DegreeProb[i] * st.DegreeMean[i]
	}

	// Probability of any damage = sum of probs where multiplier > 0
	for i := 0; i < 4; i++ {
		if multipliers[i] > 0 {
			st.AnyDamageProb += st.DegreeProb[i]
		}
	}

	// Overall variance uses law of total variance:
	// Var(D) = E[Var(D|degree)] + Var(E[D|degree])
	baseVar := DiceVariance(dice)

	eVarGiven := 0.0
	for i := 0; i < 4; i++ {
		eVarGiven += st.DegreeProb[i] * multipliers[i] * multipliers[i] * baseVar
	}

	eMeanSq := 0.0
	for i := 0; i < 4; i++ {
		eMeanSq += st.DegreeProb[i] * st.DegreeMean[i] * st.DegreeMean[i]
	}
	varOfMeans := eMeanSq - st.ExpectedDamage*st.ExpectedDamage

	totalVar := eVarGiven + varOfMeans
	if totalVar < 0 {
		totalVar = 0
	}
	st.OverallStdDev = math.Sqrt(totalVar)

	// Compute weighted mixture distribution
	st.computeMixturePDF(spell, baseVar)

	// Compute heightening table
	st.computeHeightenTable(spell)

	return st
}

// computeMixturePDF samples the weighted mixture of normal distributions across
// damage-dealing degrees of success. Zero-damage degrees (multiplier <= 0) are
// excluded so the histogram shows the damage shape, not a spike at zero.
func (st *SpellStats) computeMixturePDF(spell Spell, baseVar float64) {
	const sampleCount = 60
	multipliers := spell.Multipliers.AsSlice()
	baseSD := math.Sqrt(baseVar)

	// Find the range from the lowest and highest contributing distributions.
	loEdge := math.Inf(1)
	hiEdge := 0.0
	hasDamage := false
	for i := 0; i < 4; i++ {
		if multipliers[i] <= 0 || st.DegreeProb[i] < 0.001 {
			continue
		}
		hasDamage = true
		m := st.BaseMean * multipliers[i]
		s := baseSD * multipliers[i]
		lo := m - 3*s
		hi := m + 3*s
		if lo < loEdge {
			loEdge = lo
		}
		if hi > hiEdge {
			hiEdge = hi
		}
	}

	if !hasDamage {
		return
	}

	// Clamp low end to 0 (damage can't be negative) but don't start at 0
	// if the distribution is far from it — start just below the lowest mean.
	st.MixtureLo = math.Max(0, loEdge)
	st.MixtureHi = hiEdge
	if st.MixtureHi <= st.MixtureLo {
		st.MixtureHi = st.MixtureLo + 1
	}

	step := (st.MixtureHi - st.MixtureLo) / float64(sampleCount)
	st.MixturePDF = make([]float64, sampleCount)

	for i := 0; i < sampleCount; i++ {
		x := st.MixtureLo + (float64(i)+0.5)*step
		val := 0.0
		for d := 0; d < 4; d++ {
			if st.DegreeProb[d] < 0.0001 || multipliers[d] <= 0 {
				continue
			}
			m := st.BaseMean * multipliers[d]
			s := baseSD * multipliers[d]
			if s > 0 {
				val += st.DegreeProb[d] * scaledNormalPDF(x, m, s)
			}
		}
		st.MixturePDF[i] = val
	}
}

// computeHeightenTable fills the heightening table if the spell has BaseRank and HeightenDie set.
func (st *SpellStats) computeHeightenTable(spell Spell) {
	if spell.BaseRank <= 0 || spell.HeightenDie <= 0 {
		return
	}

	for rank := spell.BaseRank; rank <= 10; rank++ {
		d := spell.EffectiveDice(rank)
		mean := DiceMean(d)

		// Expected damage = sum(prob_i * mean * mult_i)
		multipliers := spell.Multipliers.AsSlice()
		expected := 0.0
		for i := 0; i < 4; i++ {
			expected += st.DegreeProb[i] * mean * multipliers[i]
		}

		st.HeightenTable = append(st.HeightenTable, HeightenRow{
			Rank:     rank,
			Dice:     d.String(),
			Mean:     mean,
			Expected: expected,
		})
	}
}
