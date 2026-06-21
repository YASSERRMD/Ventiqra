// Package funding models startup fundraising: valuation, the equity (dilution)
// a round costs, investor interest, and the resulting founder ownership. It also
// generates negotiable investor offers. All formulas are pure and deterministic.
package funding

import (
	mathRand "math/rand/v2"
	"math"
)

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

// Offer is a single investor proposal: capital offered for an equity stake.
type Offer struct {
	Index         int     `json:"index"`
	InvestorName  string  `json:"investor_name"`
	AmountCents   int64   `json:"amount_cents"`
	EquityPercent float64 `json:"equity_percent"`
}

var investorNames = []string{
	"Northwind Capital", "Apex Ventures", "Lighthouse Partners",
	"Meridian Fund", "Catalyst VC", "Summit Equity",
}

// OfferCount is the number of offers generated per solicitation round.
const OfferCount = 3

// offerStreamSalt distinguishes the offer-generation RNG stream.
const offerStreamSalt uint64 = 9517530246846464646

// GenerateOffers returns a deterministic set of investor offers for a round
// identified by (seed, day) and the company's current pre-money valuation.
// Offers vary in amount and asked equity so the player faces a real trade-off.
func GenerateOffers(seed, day int64, preMoney int64) []Offer {
	r := newOfferRand(seed, day)
	offers := make([]Offer, 0, OfferCount)
	for i := 0; i < OfferCount; i++ {
		// Amount varies around 5-15% of valuation.
		frac := 0.05 + r.Float64()*0.10
		amount := int64(float64(preMoney) * frac)
		if amount <= 0 {
			amount = 100_000
		}
		// Investors ask for equity above the fair dilution, with variation; the
		// "cheaper" offer asks closer to fair value.
		fair := EquityPercent(amount, preMoney)
		askMultiplier := 1.15 + r.Float64()*0.45 // 1.15x..1.60x of fair
		ask := fair * askMultiplier
		if ask > 90 {
			ask = 90
		}
		offers = append(offers, Offer{
			Index:         i,
			InvestorName:  investorNames[r.IntN(len(investorNames))],
			AmountCents:   amount,
			EquityPercent: ask,
		})
	}
	return offers
}

// NegotiateOutcome attempts to improve an offer's equity terms. On success the
// investor concedes a reduction; on failure the investor walks and the offer is
// withdrawn. The outcome is deterministic for the (seed, offerIndex) pair.
func NegotiateOutcome(seed, day int64, offerIndex int, currentEquity float64) (newEquity float64, withdrawn bool) {
	r := newOfferRand(seed^881726454626, day^int64(offerIndex+1))
	if r.Float64() < NegotiationSuccessChance {
		// Concede 15% off the asked equity.
		reduced := currentEquity * 0.85
		return reduced, false
	}
	return currentEquity, true
}

// NegotiationSuccessChance is the probability a negotiation succeeds (improves
// terms) rather than causing the investor to withdraw.
const NegotiationSuccessChance = 0.5

func newOfferRand(seed, day int64) *mathRand.Rand {
	return mathRand.New(mathRand.NewPCG(uint64(seed)^offerStreamSalt, uint64(day)^offerStreamSalt))
}
