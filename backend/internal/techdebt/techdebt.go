// Package techdebt models the company's technical-debt state: a 0-100 debt score
// that accumulates as features ship, a quality score (100 - debt), and outage
// risk derived from debt. Refactoring pays debt down.
package techdebt

// MaxDebt is the ceiling on the debt score.
const MaxDebt = 100

// ShipDebtGain is the debt added each time a feature ships (before refactoring).
const ShipDebtGain = 8

// RefactorReduction is the debt removed per refactor action.
const RefactorReduction = 25

// RefactorCostCents is the cash cost of one refactor action.
const RefactorCostCents int64 = 20_000_00 // $20,000

// Quality returns the code-quality score (100 - debt), clamped to [0,100].
func Quality(debt int) int {
	q := MaxDebt - debt
	if q < 0 {
		return 0
	}
	if q > MaxDebt {
		return MaxDebt
	}
	return q
}

// OutageRisk returns the probability [0,1] that debt causes an outage this tick.
// Below 40 debt the risk is ~0; above 80 it approaches certainty.
func OutageRisk(debt int) float64 {
	if debt <= 40 {
		return 0
	}
	if debt >= 90 {
		return 0.9
	}
	return float64(debt-40) / 100.0
}

// Accumulate adds debt for a shipped feature, clamped at MaxDebt.
func Accumulate(debt int) int {
	d := debt + ShipDebtGain
	if d > MaxDebt {
		return MaxDebt
	}
	return d
}

// Refactor reduces debt by RefactorReduction, clamped at 0.
func Refactor(debt int) int {
	d := debt - RefactorReduction
	if d < 0 {
		return 0
	}
	return d
}
