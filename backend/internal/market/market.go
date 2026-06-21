// Package market models the addressable market a company sells into: its size,
// growth, demand, and a trend multiplier. Formulas are pure and deterministic.
package market

import "math/rand/v2"

// Model holds the market state advanced by the model.
type Model struct {
	TAM             int64   // total addressable market (customers)
	GrowthRate      float64 // fraction per simulated month
	TrendMultiplier float64 // demand cycle factor around 1.0
}

// DefaultModel is the market every new company starts with.
var DefaultModel = Model{TAM: 100_000, GrowthRate: 0.01, TrendMultiplier: 1.0}

// marketStreamSalt separates the market RNG stream.
const marketStreamSalt uint64 = 845678901234567890

// DailyGrowth returns the daily increase in the addressable market derived from
// a monthly growth rate (over a 30-day month).
func DailyGrowth(m Model) int64 {
	monthly := float64(m.TAM) * m.GrowthRate
	return int64(monthly / 30.0)
}

// Advance applies one simulated day: the market grows and the trend multiplier
// drifts in a bounded cycle.
func Advance(m Model, seed, day int64) Model {
	r := rand.New(rand.NewPCG(uint64(seed)^marketStreamSalt, uint64(day)))
	tam := m.TAM + DailyGrowth(m)
	// Trend drifts by up to ±2%, bounded to [0.5, 1.5].
	trend := m.TrendMultiplier + (r.Float64()*0.04 - 0.02)
	if trend < 0.5 {
		trend = 0.5
	}
	if trend > 1.5 {
		trend = 1.5
	}
	return Model{TAM: tam, GrowthRate: m.GrowthRate, TrendMultiplier: trend}
}

// DemandPenetration returns how many potential customers remain un-served, as a
// fraction of the TAM (0..1). It bounds how large a customer base can grow.
func DemandPenetration(customers int, tam int64) float64 {
	if tam <= 0 {
		return 0
	}
	pen := 1.0 - float64(customers)/float64(tam)
	if pen < 0 {
		return 0
	}
	if pen > 1 {
		return 1
	}
	return pen
}
