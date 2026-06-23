// Package balance codifies the simulation's tuned economic constants and the
// invariants that guard them. Centralizing these here means a single source of
// truth for revenue, churn, funding, and hiring formulas, and a test suite that
// fails when a tuning change breaks a gameplay invariant.
package balance

// --- Revenue formula ---
// Daily revenue = customers × ARPU × priceMultiplier.
// ARPU is the average revenue per user per day in cents.

const (
	// ARPUCents is the base average revenue per customer per day (cents).
	ARPUCents int64 = 50 // $0.50/day → ~$15/mo per customer at neutral

	// PriceElasticity is how strongly price changes move demand (higher = more
	// elastic). Used by the pricing module to scale acquisition.
	PriceElasticity = 1.5

	// MinPriceMultiplier caps the discount effect.
	MinPriceMultiplier = 0.5

	// MaxPriceMultiplier caps the premium effect.
	MaxPriceMultiplier = 2.0
)

// RevenuePerDay returns the daily revenue (cents) for a customer base at a given
// price multiplier.
func RevenuePerDay(customers int, priceMultiplier float64) int64 {
	if customers <= 0 {
		return 0
	}
	if priceMultiplier < MinPriceMultiplier {
		priceMultiplier = MinPriceMultiplier
	}
	if priceMultiplier > MaxPriceMultiplier {
		priceMultiplier = MaxPriceMultiplier
	}
	return int64(float64(customers) * float64(ARPUCents) * priceMultiplier)
}

// --- Churn formula ---
// Daily churn = customers × baseChurnRate × satisfactionFactor × difficultyMultiplier.

const (
	// BaseChurnRate is the fraction of customers lost per day at neutral
	// satisfaction and difficulty.
	BaseChurnRate = 0.01 // 1%/day → ~26%/mo
)

// ChurnPerDay returns the daily customer churn count.
func ChurnPerDay(customers, satisfaction int, churnMultiplier float64) int {
	if customers <= 0 {
		return 0
	}
	if churnMultiplier <= 0 {
		churnMultiplier = 1
	}
	// Satisfaction factor: 100 sat → 0.5× churn; 50 sat → 1× churn; 0 sat → 2× churn.
	satFactor := 1.0 + (50.0-float64(satisfaction))/50.0
	if satFactor < 0.1 {
		satFactor = 0.1
	}
	rate := BaseChurnRate * satFactor * churnMultiplier
	lost := int(float64(customers) * rate)
	if lost > customers {
		lost = customers
	}
	return lost
}

// --- Funding probability ---
// The chance a funding round succeeds rises with traction (customers) and falls
// with the ask size.

const (
	// FundingBaseChance is the base success probability for a reasonably-sized
	// round with decent traction.
	FundingBaseChance = 0.5

	// FundingTractionCustomers is the customer count that yields the base chance.
	FundingTractionCustomers = 500

	// FundingAskScaleCents is the round size (cents) at which the base chance
	// applies; larger asks reduce the chance.
	FundingAskScaleCents int64 = 2_000_000_00
)

// FundingChance returns the probability [0,1] that a round succeeds.
func FundingChance(customers int, askCents int64, chanceMultiplier float64) float64 {
	if chanceMultiplier <= 0 {
		chanceMultiplier = 1
	}
	// Traction bonus: more customers than the threshold raises the chance.
	traction := float64(customers) / float64(FundingTractionCustomers)
	if traction > 2 {
		traction = 2
	}
	// Ask penalty: larger asks reduce the chance.
	askFactor := float64(FundingAskScaleCents) / float64(askCents)
	if askCents <= 0 {
		askFactor = 1
	}
	if askFactor > 2 {
		askFactor = 2
	}
	chance := FundingBaseChance * (0.5 + 0.5*traction) * askFactor * chanceMultiplier
	if chance > 0.95 {
		chance = 0.95
	}
	if chance < 0.05 {
		chance = 0.05
	}
	return chance
}

// --- Hiring market ---
// Candidate quality and cost scale with the target role's seniority.

const (
	// HireCostBaseCents is the base one-time hiring cost.
	HireCostBaseCents int64 = 5_000_00 // $5,000

	// OfferAcceptBase is the base probability a candidate accepts an offer.
	OfferAcceptBase = 0.7
)

// HireCost returns the one-time cost to hire for a role.
func HireCost(role string) int64 {
	switch role {
	case "engineer", "developer":
		return 8_000_00
	case "sales":
		return 6_000_00
	case "support":
		return 4_000_00
	default:
		return HireCostBaseCents
	}
}
