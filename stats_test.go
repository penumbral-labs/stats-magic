package main

import (
	"math"
	"testing"
)

const epsilon = 1e-9

func approxEqual(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}

func TestDiceMean(t *testing.T) {
	tests := []struct {
		dice DiceFormula
		want float64
	}{
		{DiceFormula{1, 6, 0}, 3.5},
		{DiceFormula{2, 6, 0}, 7.0},
		{DiceFormula{8, 6, 0}, 28.0},
		{DiceFormula{4, 10, 4}, 26.0},
		{DiceFormula{1, 20, 0}, 10.5},
		{DiceFormula{0, 6, 5}, 5.0},
	}

	for _, tt := range tests {
		got := DiceMean(tt.dice)
		if !approxEqual(got, tt.want, epsilon) {
			t.Errorf("DiceMean(%v) = %f, want %f", tt.dice, got, tt.want)
		}
	}
}

func TestDiceVariance(t *testing.T) {
	tests := []struct {
		dice DiceFormula
		want float64
	}{
		{DiceFormula{1, 6, 0}, 35.0 / 12.0},
		{DiceFormula{2, 6, 0}, 70.0 / 12.0},
		{DiceFormula{8, 6, 0}, 280.0 / 12.0},
		{DiceFormula{1, 20, 0}, 399.0 / 12.0},
	}

	for _, tt := range tests {
		got := DiceVariance(tt.dice)
		if !approxEqual(got, tt.want, epsilon) {
			t.Errorf("DiceVariance(%v) = %f, want %f", tt.dice, got, tt.want)
		}
	}
}

func TestParseDice(t *testing.T) {
	tests := []struct {
		input string
		want  DiceFormula
	}{
		{"8d6", DiceFormula{8, 6, 0}},
		{"4d10+4", DiceFormula{4, 10, 4}},
		{"2d8-1", DiceFormula{2, 8, -1}},
		{"1d20", DiceFormula{1, 20, 0}},
		{"invalid", DiceFormula{}},
		{"", DiceFormula{}},
		{"d6", DiceFormula{}},
	}

	for _, tt := range tests {
		got := ParseDice(tt.input)
		if got != tt.want {
			t.Errorf("ParseDice(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestSaveDegreeProbabilities_SumToOne(t *testing.T) {
	for dc := 10; dc <= 40; dc += 5 {
		for saveMod := -5; saveMod <= 25; saveMod += 3 {
			probs := CalcSaveDegreeProbabilities(dc, saveMod)
			sum := probs[0] + probs[1] + probs[2] + probs[3]
			if !approxEqual(sum, 1.0, 1e-10) {
				t.Errorf("DC=%d saveMod=%d: probabilities sum to %f, want 1.0", dc, saveMod, sum)
			}
		}
	}
}

func TestAttackDegreeProbabilities_SumToOne(t *testing.T) {
	for ac := 10; ac <= 40; ac += 5 {
		for attackMod := -5; attackMod <= 25; attackMod += 3 {
			probs := CalcAttackDegreeProbabilities(attackMod, ac)
			sum := probs[0] + probs[1] + probs[2] + probs[3]
			if !approxEqual(sum, 1.0, 1e-10) {
				t.Errorf("AC=%d attackMod=%d: probabilities sum to %f, want 1.0", ac, attackMod, sum)
			}
		}
	}
}

func TestSaveDegreeProbabilities_KnownCase(t *testing.T) {
	probs := CalcSaveDegreeProbabilities(20, 5)

	if !approxEqual(probs[0], 5.0/20.0, epsilon) {
		t.Errorf("Best (Crit Fail save) = %f, want %f", probs[0], 5.0/20.0)
	}
	if !approxEqual(probs[1], 9.0/20.0, epsilon) {
		t.Errorf("Good (Fail save) = %f, want %f", probs[1], 9.0/20.0)
	}
	if !approxEqual(probs[2], 5.0/20.0, epsilon) {
		t.Errorf("Bad (Success save) = %f, want %f", probs[2], 5.0/20.0)
	}
	if !approxEqual(probs[3], 1.0/20.0, epsilon) {
		t.Errorf("Worst (Crit Success save) = %f, want %f", probs[3], 1.0/20.0)
	}
}

func TestSaveDegreeProbabilities_CritFailThreshold(t *testing.T) {
	probs := CalcSaveDegreeProbabilities(20, 0)

	if !approxEqual(probs[0], 10.0/20.0, epsilon) {
		t.Errorf("Best (Crit Fail save) = %f, want %f (crit fail threshold test)", probs[0], 10.0/20.0)
	}
	if !approxEqual(probs[1], 9.0/20.0, epsilon) {
		t.Errorf("Good (Fail save) = %f, want %f", probs[1], 9.0/20.0)
	}
	if !approxEqual(probs[2], 0.0, epsilon) {
		t.Errorf("Bad (Success save) = %f, want 0.0", probs[2])
	}
	if !approxEqual(probs[3], 1.0/20.0, epsilon) {
		t.Errorf("Worst (Crit Success save) = %f, want %f", probs[3], 1.0/20.0)
	}
}

func TestAttackDegreeProbabilities_KnownCase(t *testing.T) {
	probs := CalcAttackDegreeProbabilities(10, 20)

	if !approxEqual(probs[0], 1.0/20.0, epsilon) {
		t.Errorf("Best (Crit Hit) = %f, want %f", probs[0], 1.0/20.0)
	}
	if !approxEqual(probs[1], 10.0/20.0, epsilon) {
		t.Errorf("Good (Hit) = %f, want %f", probs[1], 10.0/20.0)
	}
	if !approxEqual(probs[2], 8.0/20.0, epsilon) {
		t.Errorf("Bad (Miss) = %f, want %f", probs[2], 8.0/20.0)
	}
	if !approxEqual(probs[3], 1.0/20.0, epsilon) {
		t.Errorf("Worst (Crit Miss) = %f, want %f", probs[3], 1.0/20.0)
	}
}

func TestCalcSpellStats_Fireball(t *testing.T) {
	spell := Spell{
		Name:        "Fireball",
		Type:        SpellTypeSave,
		SaveType:    "Reflex",
		Dice:        ParseDice("6d6"),
		Multipliers: DefaultSaveMultipliers(),
		BaseRank:    3,
		HeightenDie: 2,
	}
	enc := EncounterState{
		SpellDC: 20,
		RefMod:  5,
	}

	stats := CalcSpellStats(spell, enc)

	if !approxEqual(stats.BaseMean, 21.0, epsilon) {
		t.Errorf("BaseMean = %f, want 21.0", stats.BaseMean)
	}

	if stats.ExpectedDamage <= 0 || stats.ExpectedDamage > 2*stats.BaseMean {
		t.Errorf("ExpectedDamage = %f, out of reasonable range", stats.ExpectedDamage)
	}

	if stats.OverallStdDev <= 0 {
		t.Errorf("OverallStdDev = %f, should be positive", stats.OverallStdDev)
	}

	if stats.AnyDamageProb <= 0 || stats.AnyDamageProb > 1 {
		t.Errorf("AnyDamageProb = %f, should be between 0 and 1", stats.AnyDamageProb)
	}

	// Should have heightening table from rank 3 to 10
	if len(stats.HeightenTable) != 8 {
		t.Errorf("HeightenTable length = %d, want 8", len(stats.HeightenTable))
	}

	if len(stats.HeightenTable) > 0 {
		if stats.HeightenTable[0].Rank != 3 {
			t.Errorf("HeightenTable[0].Rank = %d, want 3", stats.HeightenTable[0].Rank)
		}
	}

	if len(stats.MixturePDF) == 0 {
		t.Error("MixturePDF should be populated")
	}
}

func TestCalcSpellStats_AttackSpell(t *testing.T) {
	spell := Spell{
		Name:        "Shocking Grasp",
		Type:        SpellTypeAttack,
		Dice:        ParseDice("2d12"),
		Multipliers: DefaultAttackMultipliers(),
	}
	enc := EncounterState{
		AttackMod: 10,
		EnemyAC:   20,
	}

	stats := CalcSpellStats(spell, enc)

	if !approxEqual(stats.BaseMean, 13.0, epsilon) {
		t.Errorf("BaseMean = %f, want 13.0", stats.BaseMean)
	}

	if stats.ExpectedDamage <= 0 {
		t.Errorf("ExpectedDamage = %f, should be positive", stats.ExpectedDamage)
	}

	// For attack, only hit and crit hit deal damage
	expectedAny := stats.DegreeProb[0] + stats.DegreeProb[1]
	if !approxEqual(stats.AnyDamageProb, expectedAny, epsilon) {
		t.Errorf("AnyDamageProb = %f, want %f", stats.AnyDamageProb, expectedAny)
	}
}

func TestCalcSpellStats_SaveTypeRouting(t *testing.T) {
	// Test that save type correctly routes to the right encounter modifier
	spell := Spell{
		Name:        "Sound Burst",
		Type:        SpellTypeSave,
		SaveType:    "Fortitude",
		Dice:        ParseDice("2d10"),
		Multipliers: DefaultSaveMultipliers(),
	}
	enc := EncounterState{
		SpellDC: 20,
		RefMod:  5,
		FortMod: 15, // Much higher fort save
		WillMod: 3,
	}

	stats := CalcSpellStats(spell, enc)

	// With fort mod +15 vs DC 20, enemy saves more often
	// Compare to a reflex-based spell with the same dice
	spellRef := spell
	spellRef.SaveType = "Reflex"
	statsRef := CalcSpellStats(spellRef, enc)

	// Fort save is higher, so Fortitude spell should deal less expected damage
	if stats.ExpectedDamage >= statsRef.ExpectedDamage {
		t.Errorf("Fortitude spell (E[Dmg]=%.1f) should deal less than Reflex spell (E[Dmg]=%.1f) when FortMod > RefMod",
			stats.ExpectedDamage, statsRef.ExpectedDamage)
	}
}

func TestEffectiveDice(t *testing.T) {
	spell := Spell{
		Dice:        ParseDice("6d6"),
		BaseRank:    3,
		HeightenDie: 2,
	}

	d := spell.EffectiveDice(3)
	if d.Count != 6 {
		t.Errorf("EffectiveDice(3).Count = %d, want 6", d.Count)
	}

	d = spell.EffectiveDice(5)
	if d.Count != 10 {
		t.Errorf("EffectiveDice(5).Count = %d, want 10", d.Count)
	}

	d = spell.EffectiveDice(10)
	if d.Count != 20 {
		t.Errorf("EffectiveDice(10).Count = %d, want 20", d.Count)
	}
}

func TestNormalCDF_KnownValues(t *testing.T) {
	tests := []struct {
		x    float64
		want float64
		tol  float64
	}{
		{0, 0.5, 1e-6},
		{-8, 0, 1e-6},
		{8, 1, 1e-6},
		{1.96, 0.975, 1e-3},
		{-1.96, 0.025, 1e-3},
	}

	for _, tt := range tests {
		got := normalCDF(tt.x)
		if !approxEqual(got, tt.want, tt.tol) {
			t.Errorf("normalCDF(%f) = %f, want ~%f", tt.x, got, tt.want)
		}
	}
}

func TestPresetToSpell(t *testing.T) {
	presets := AllPresets()
	if len(presets) < 15 {
		t.Errorf("Expected at least 15 presets, got %d", len(presets))
	}

	enc := EncounterState{
		SpellDC:   20,
		AttackMod: 10,
		RefMod:    5,
		FortMod:   5,
		WillMod:   5,
		EnemyAC:   20,
	}

	for _, p := range presets {
		sp := p.ToSpell()
		if sp.Name == "" {
			t.Error("Preset produced spell with empty name")
		}
		if !sp.Dice.Valid() {
			t.Errorf("Preset %q produced invalid dice formula", p.Name)
		}
		if sp.Type != p.Type {
			t.Errorf("Preset %q: spell type %v != preset type %v", p.Name, sp.Type, p.Type)
		}
		if sp.Type == SpellTypeSave && sp.SaveType == "" {
			t.Errorf("Preset %q: save spell has empty SaveType", p.Name)
		}

		// CalcSpellStats should not panic
		stats := CalcSpellStats(sp, enc)
		if stats.ExpectedDamage < 0 {
			t.Errorf("Preset %q: negative expected damage %f", p.Name, stats.ExpectedDamage)
		}
	}
}

func TestDegreeMultiplierOrder(t *testing.T) {
	sm := DefaultSaveMultipliers()
	s := sm.AsSlice()
	if s[0] < s[1] || s[1] < s[2] || s[2] < s[3] {
		t.Errorf("Save multipliers not in descending order: %v", s)
	}

	am := DefaultAttackMultipliers()
	a := am.AsSlice()
	if a[0] < a[1] || a[1] < a[2] {
		t.Errorf("Attack multipliers not in descending order: %v", a)
	}
}

func TestEncounterState_SaveModFor(t *testing.T) {
	enc := EncounterState{
		RefMod:  10,
		FortMod: 15,
		WillMod: 8,
	}

	if got := enc.SaveModFor("Reflex"); got != 10 {
		t.Errorf("SaveModFor(Reflex) = %d, want 10", got)
	}
	if got := enc.SaveModFor("Fortitude"); got != 15 {
		t.Errorf("SaveModFor(Fortitude) = %d, want 15", got)
	}
	if got := enc.SaveModFor("Will"); got != 8 {
		t.Errorf("SaveModFor(Will) = %d, want 8", got)
	}
	// Default to Reflex
	if got := enc.SaveModFor(""); got != 10 {
		t.Errorf("SaveModFor('') = %d, want 10 (default Reflex)", got)
	}
}

func TestSpellDefenseLabel(t *testing.T) {
	save := Spell{Type: SpellTypeSave, SaveType: "Reflex"}
	if got := save.DefenseLabel(); got != "Ref save" {
		t.Errorf("DefenseLabel() = %q, want 'Ref save'", got)
	}

	fort := Spell{Type: SpellTypeSave, SaveType: "Fortitude"}
	if got := fort.DefenseLabel(); got != "For save" {
		t.Errorf("DefenseLabel() = %q, want 'For save'", got)
	}

	atk := Spell{Type: SpellTypeAttack}
	if got := atk.DefenseLabel(); got != "Attack" {
		t.Errorf("DefenseLabel() = %q, want 'Attack'", got)
	}
}
