// Package funding models startup fundraising: valuation, the equity (dilution)
// a round costs, investor interest, and the resulting founder ownership. All
// formulas are pure and deterministic.
package funding

import "math"

// RevenueMultiplier is the annual-revenue multiple used for pre-money valuation.
const RevenueMultiplier int64 = 8

// MonthsPerYear annualizes monthly revenue for the valuation.
const MonthsPerYear int64 = 12

// RoundNames is the ordered sequence of round names a company progresses
// through.
var RoundNames = []string{"pre-seed", "seed", "series-a", "series-b", "series-c", "growth"}

// NextRoundName returns the round name following the given number of prior
// closed rounds, capping at the final stage.
func NextRoundName(priorRounds int) string {
	if priorRounds < 0 {
		priorRounds = 0
	}
	if priorRounds >= len(RoundNames) {
		return RoundNames[len(RoundNames)-1]
	}
	return RoundNames[priorRounds]
}

// PreMoneyValuation estimates the pre-money valuation (in cents) from a monthly
// revenue run-rate and cash on hand. It is the larger of an annualized revenue
// multiple and the cash balance, so revenue-backed and asset-backed companies
// both get a meaningful floor.
func PreMoneyValuation(cash, monthlyRevenue int64) int64 {
	byRevenue := monthlyRevenue * MonthsPerYear * RevenueMultiplier
	if cash > byRevenue {
		return cash
	}
	return byRevenue
}

// EquityPercent returns the equity (dilution) an investor receives for a given
// raise amount at a pre-money valuation: amount / (pre + amount) * 100.
func EquityPercent(amount, preMoney int64) float64 {
	if amount <= 0 || preMoney < 0 {
		return 0
	}
	post := float64(preMoney + amount)
	return float64(amount) / post * 100
}

// FounderEquity compounds a new round's dilution onto the founder's prior
// ownership, returning the resulting founder equity percent.
func FounderEquity(priorFounderEquity, newRoundEquity float64) float64 {
	if priorFounderEquity < 0 {
		priorFounderEquity = 0
	}
	if priorFounderEquity > 100 {
		priorFounderEquity = 100
	}
	result := priorFounderEquity * (1 - newRoundEquity/100)
	return math.Round(result*100) / 100
}

// InvestorInterest returns a 0..1 score representing how attractive the company
// is to investors, rising with revenue traction and falling with cash trouble
// (negative cash). It is a coarse heuristic for the dashboard.
func InvestorInterest(monthlyRevenue, cash int64) float64 {
	score := 0.2
	if monthlyRevenue > 0 {
		// Traction boosts interest, logarithmic so it plateaus.
		score += 0.4 * (1 - math.Exp(-float64(monthlyRevenue)/500_000.0))
	}
	if cash < 0 {
		score -= 0.3
	}
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	return score
}
