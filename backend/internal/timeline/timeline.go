// Package timeline models the company history view: a unified, chronologically
// ordered stream of milestone and notable events, plus a monthly summary roll-up.
// It is pure data and helpers so the server can assemble and serve it.
package timeline

// Kind classifies a timeline entry.
type Kind string

const (
	KindMilestone Kind = "milestone"
	KindLaunch    Kind = "launch"
	KindFunding   Kind = "funding"
	KindDecision  Kind = "decision"
	KindCrisis    Kind = "crisis"
	KindEvent     Kind = "event"
	KindReputation Kind = "reputation"
)

// Entry is a single chronological timeline row.
type Entry struct {
	ID          string `json:"id"`
	Kind        Kind   `json:"kind"`
	Title       string `json:"title"`
	Description string `json:"description"`
	SimDay      int    `json:"sim_day"`
	CreatedAt   string `json:"created_at"`
}

// MonthlySummary aggregates a company's progress for a completed 30-day month.
type MonthlySummary struct {
	Month         int   `json:"month"`          // 1-indexed month number
	StartDay      int   `json:"start_day"`      // first day of the month (inclusive)
	EndDay        int   `json:"end_day"`        // last day of the month (inclusive)
	CashChange    int64 `json:"cash_change"`    // net cash delta over the month (cents)
	RevenueEnd    int64 `json:"revenue_end"`    // daily revenue at month end (cents)
	BurnEnd       int64 `json:"burn_end"`       // monthly burn at month end (cents)
	EventsCount   int   `json:"events_count"`   // timeline entries in this month
}

// DaysPerMonth is the simulation's month length (mirrors the finance module).
const DaysPerMonth = 30

// MonthForDay returns the 1-indexed month number containing the given sim day.
// Day 0 is month 0 (pre-start); days 1-30 are month 1, etc.
func MonthForDay(day int) int {
	if day <= 0 {
		return 0
	}
	return (day-1)/DaysPerMonth + 1
}

// MonthBounds returns the inclusive [startDay, endDay] for a 1-indexed month.
func MonthBounds(month int) (startDay, endDay int) {
	if month <= 0 {
		return 0, 0
	}
	startDay = (month-1)*DaysPerMonth + 1
	endDay = month * DaysPerMonth
	return startDay, endDay
}

// IsInMonth reports whether a sim day falls within the given 1-indexed month.
func IsInMonth(day, month int) bool {
	startDay, endDay := MonthBounds(month)
	return day >= startDay && day <= endDay
}

// SummarizeMonths builds monthly summaries from the company's run. It takes the
// current sim day, the ending cash/revenue/burn, and per-month cash deltas, and
// returns one summary per completed month plus the current (in-progress) month.
func SummarizeMonths(currentDay int, endingRevenue, endingBurn int64, monthlyCashDeltas []int64) []MonthlySummary {
	months := MonthForDay(currentDay)
	if months < 1 {
		months = 1
	}
	out := make([]MonthlySummary, 0, months)
	for m := 1; m <= months; m++ {
		startDay, endDay := MonthBounds(m)
		var delta int64
		if m-1 < len(monthlyCashDeltas) {
			delta = monthlyCashDeltas[m-1]
		}
		out = append(out, MonthlySummary{
			Month: m, StartDay: startDay, EndDay: endDay,
			CashChange: delta,
			RevenueEnd: endingRevenue, BurnEnd: endingBurn,
		})
	}
	return out
}
