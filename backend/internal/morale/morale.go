// Package morale models team morale dynamics: daily decay, boosts, burnout, and
// resignation risk. Formulas are pure and deterministic.
package morale

import "math/rand/v2"

// DecayFloor is the equilibrium morale trends toward: routine work wears it
// down to here, and events/boosts move it above or below.
const DecayFloor = 50

// BurnoutThreshold is the morale value at/below which an employee is considered
// burned out (near-zero effective output and high resignation risk).
const BurnoutThreshold = 30

// moraleStreamSalt separates the morale RNG stream.
const moraleStreamSalt uint64 = 7777555533331111

// DailyDecay returns the daily morale change from routine wear. Morale above the
// decay floor drifts down by 1; at or below the floor it holds steady.
func DailyDecay(morale int) int {
	if morale > DecayFloor {
		return -1
	}
	return 0
}

// Boost applies a positive morale bump, clamped to 100.
func Boost(current, amount int) int {
	current += amount
	if current > 100 {
		return 100
	}
	if current < 0 {
		return 0
	}
	return current
}

// BurnoutFactor scales effective output by morale, dropping sharply at/below the
// burnout threshold to near zero.
func BurnoutFactor(morale int) float64 {
	if morale <= 0 {
		return 0
	}
	if morale <= BurnoutThreshold {
		return float64(morale) / float64(BurnoutThreshold) * 0.3
	}
	return 1.0
}

// Resigns reports deterministically whether an employee resigns on a given day.
// Resignation risk rises sharply as morale falls below the burnout threshold.
func Resigns(morale int, seed, day int64, employeeIndex int) bool {
	if morale > BurnoutThreshold {
		return false
	}
	r := rand.New(rand.NewPCG(uint64(seed)^moraleStreamSalt, uint64(day)^uint64(employeeIndex+1)))
	// At morale 0 → ~60% daily chance; at threshold → ~10%.
	chance := 0.6 * float64(BurnoutThreshold-morale+1) / float64(BurnoutThreshold+1)
	return r.Float64() < chance
}
