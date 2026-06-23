// Package contracts models enterprise contracts: multi-year recurring-revenue
// agreements with a negotiated discount and term. Each tick accrues revenue
// and counts down remaining days; on expiry a renewal roll decides renewed vs
// churned.
package contracts

import "math/rand/v2"

// DaysPerYear is the simulation's year length (mirrors timeline.DaysPerMonth*12).
const DaysPerYear = 360

// DefaultTermDays is the standard enterprise contract length.
const DefaultTermDays = 360 // 1 year

// RenewChance is the base probability a contract renews at expiry (before
// satisfaction adjustments).
const RenewChance = 0.7

// TermForYears returns the sim-day term for a given number of years.
func TermForYears(years int) int {
	if years < 1 {
		years = 1
	}
	return years * DaysPerYear
}

// AnnualValueFor returns the effective annual value after applying a discount.
func AnnualValueFor(baseAnnual int64, discountPct int) int64 {
	if discountPct <= 0 {
		return baseAnnual
	}
	if discountPct > 50 {
		discountPct = 50
	}
	return baseAnnual - (baseAnnual*int64(discountPct))/100
}

// DailyRevenue returns the daily revenue (cents) from an active contract.
func DailyRevenue(annualValue int64) int64 {
	if annualValue <= 0 {
		return 0
	}
	return annualValue / DaysPerYear
}

// RenewalRoll decides whether a contract renews at expiry. Higher customer
// satisfaction raises the base chance; seed makes it deterministic.
func RenewalRoll(seed int64, satisfaction int) bool {
	// Each point of satisfaction above 50 adds 0.5% renewal chance (cap +20%).
	boost := 0.0
	if satisfaction > 50 {
		boost = float64(satisfaction-50) * 0.005
	}
	if boost > 0.20 {
		boost = 0.20
	}
	chance := RenewChance + boost
	if chance > 0.95 {
		chance = 0.95
	}
	r := rand.New(rand.NewPCG(uint64(seed)^0xC0DE7AC7, uint64(seed)|7))
	return r.Float64() < chance
}
