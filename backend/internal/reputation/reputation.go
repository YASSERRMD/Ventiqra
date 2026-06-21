// Package reputation models brand reputation: its score, how it modulates growth,
// and how satisfaction and health drift it. Formulas are pure and deterministic.
package reputation

// NeutralScore is the starting reputation for a new company.
const NeutralScore = 50

// GrowthMultiplier returns the customer-acquisition multiplier for a reputation
// score. Neutral (50) maps to 1.0; top reputation grants up to ~1.3×, poor
// reputation throttles acquisition down toward ~0.7×.
func GrowthMultiplier(score int) float64 {
	s := clamp(score)
	// Linear map: 0 → 0.7, 50 → 1.0, 100 → 1.3.
	return 0.7 + float64(s)/50.0*0.3
}

// SatisfactionDrift returns the daily score delta derived from average customer
// satisfaction (0..100). Happy customers lift reputation; unhappy ones erode it.
func SatisfactionDrift(avgSatisfaction int) int {
	s := clamp(avgSatisfaction)
	switch {
	case s >= 75:
		return 1
	case s < 40:
		return -1
	default:
		return 0
	}
}

// HealthDelta returns the reputation delta for a company health state.
func HealthDelta(health string) int {
	switch health {
	case "bankrupt":
		return -5
	case "critical":
		return -1
	case "warning":
		return 0
	default:
		return 0
	}
}

// Clamp clamps a score to the 0..100 range.
func Clamp(score int) int { return clamp(score) }

func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}
