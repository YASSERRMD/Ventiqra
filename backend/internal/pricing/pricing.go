// Package pricing models how a product's price affects demand and revenue.
// Formulas are pure and deterministic so pricing experiments are reproducible.
package pricing

import "math"

// BaselineMonthlyCents is the reference monthly price (~$10) at which the
// demand multiplier is exactly 1.0.
const BaselineMonthlyCents int64 = 1000

// DefaultElasticity governs how aggressively price changes move demand. Higher
// values make customers more price-sensitive.
const DefaultElasticity = 1.5

// DemandMultiplier returns the acquisition multiplier for a monthly price
// relative to the baseline. Prices above baseline reduce demand; below baseline
// boost it. A zero/nil price (free) maps to the maximum boost. The result is
// clamped to [MinDemandMul, MaxDemandMul].
func DemandMultiplier(priceCents, baselineCents int64, elasticity float64) float64 {
	if elasticity <= 0 {
		elasticity = DefaultElasticity
	}
	if baselineCents <= 0 {
		baselineCents = BaselineMonthlyCents
	}
	if priceCents <= 0 {
		// Free product: strong adoption boost, but see DailyRevenueCents for the
		// offsetting zero-revenue effect.
		return MaxDemandMul
	}
	ratio := float64(priceCents) / float64(baselineCents)
	mul := math.Pow(ratio, -elasticity)
	return clamp(mul, MinDemandMul, MaxDemandMul)
}

// DailyRevenueCents returns the deterministic daily revenue (in cents) produced
// by a customer base paying a monthly price. Revenue is monthly price scaled to
// a 30-day month.
func DailyRevenueCents(totalCustomers int, monthlyPriceCents int64) int64 {
	if totalCustomers <= 0 || monthlyPriceCents <= 0 {
		return 0
	}
	return int64(totalCustomers) * monthlyPriceCents / 30
}

// Clamp bounds for the demand multiplier.
const (
	MinDemandMul = 0.1
	MaxDemandMul = 3.0
)

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
