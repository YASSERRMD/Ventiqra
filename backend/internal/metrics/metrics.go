// Package metrics computes derived company metrics from raw simulation state.
//
// All functions here are pure: they accept primitive inputs and return a
// Metrics value with no side effects and no I/O. This keeps the derivation
// trivially testable and fully deterministic (identical inputs always yield
// identical outputs), which is central to Ventiqra's reproducible simulation.
package metrics

// ValuationMultiplier is the annual-revenue multiple used in the valuation
// estimate. It is deliberately conservative for an early-stage simulation.
const ValuationMultiplier int64 = 4

// MonthsPerYear annualizes a revenue figure for the valuation formula.
const MonthsPerYear int64 = 12

// InfiniteRunway is returned for RunwayMonths when a company has no monthly
// burn: it cannot run out of cash, so the runway is unbounded. A negative
// sentinel keeps the value distinguishable from any real (non-negative)
// runway and serializes cleanly to JSON.
const InfiniteRunway float64 = -1

// Health thresholds (in months of runway) defining warning and critical states.
const (
	WarningRunwayMonths  = 6.0
	CriticalRunwayMonths = 3.0
)

// Health status values reported alongside metrics and on the tick response.
const (
	HealthHealthy  = "healthy"
	HealthWarning  = "warning"
	HealthCritical = "critical"
	HealthBankrupt = "bankrupt"
)

// Health derives a coarse status from cash and runway. Bankruptcy (negative
// cash) always dominates; otherwise the runway bands drive warning/critical.
func Health(cash int64, runwayMonths float64) string {
	if cash < 0 {
		return HealthBankrupt
	}
	if runwayMonths == InfiniteRunway {
		return HealthHealthy
	}
	if runwayMonths <= CriticalRunwayMonths {
		return HealthCritical
	}
	if runwayMonths <= WarningRunwayMonths {
		return HealthWarning
	}
	return HealthHealthy
}

// Metrics is a derived snapshot of a company's financial position. All money
// fields are expressed in cents to match the rest of the codebase.
type Metrics struct {
	CashCents         int64
	RevenueCents      int64
	BurnCentsPerMonth int64
	ValuationCents    int64
	RunwayMonths      float64
}

// Compute derives a Metrics snapshot from raw inputs.
//
// The valuation estimate is max(cash, revenue * MonthsPerYear *
// ValuationMultiplier): a company is worth at least its cash on hand, and an
// annualized revenue multiple once recurring revenue exceeds its balance.
// RunwayMonths is cash / monthlyBurn as a floating-point number of months, or
// InfiniteRunway when monthlyBurn <= 0 (the company cannot go broke).
//
// employeesCount and day are accepted to keep the signature stable as the
// model grows (per-employee burn and day-aware projections arrive in later
// phases) but do not affect the result today.
func Compute(cash, revenue, monthlyBurn int64, employeesCount int, day int) Metrics {
	_ = employeesCount
	_ = day

	annualizedRevenueValue := revenue * MonthsPerYear * ValuationMultiplier
	valuation := cash
	if annualizedRevenueValue > valuation {
		valuation = annualizedRevenueValue
	}

	runway := InfiniteRunway
	if monthlyBurn > 0 {
		runway = float64(cash) / float64(monthlyBurn)
	}

	return Metrics{
		CashCents:         cash,
		RevenueCents:      revenue,
		BurnCentsPerMonth: monthlyBurn,
		ValuationCents:    valuation,
		RunwayMonths:      runway,
	}
}
