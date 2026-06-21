// Package analytics derives the time-series and headline figures the analytics
// dashboard plots: cash, revenue, burn, customers, and valuation over the run.
// Pure functions over snapshot data so the server and tests share one source of
// truth.
package analytics

// Point is a single day's metric sample.
type Point struct {
	Day             int   `json:"day"`
	CashCents       int64 `json:"cash_cents"`
	RevenueCents    int64 `json:"revenue_cents"`
	MonthlyBurn     int64 `json:"monthly_burn"`
	Customers       int   `json:"customers"`
	ValuationCents  int64 `json:"valuation_cents"`
}

// Series is the full plottable dataset returned by the analytics endpoint.
type Series struct {
	Day        int     `json:"day"`
	Cash       []Point `json:"cash"`
	Revenue    []Point `json:"revenue"`
	Customers  []Point `json:"customers"`
	Burn       []Point `json:"burn"`
	Valuation  []Point `json:"valuation"`
}

// RevenueMultiple is the annual-revenue multiple used to derive valuation.
const RevenueMultiple int64 = 8

// MonthsPerYear annualizes monthly figures.
const MonthsPerYear int64 = 12

// Valuation returns a simple valuation (cents) from cash and monthly revenue:
// the larger of the cash balance and an annualized revenue multiple.
func Valuation(cash, monthlyRevenue int64) int64 {
	byRevenue := monthlyRevenue * MonthsPerYear * RevenueMultiple
	if cash > byRevenue {
		return cash
	}
	return byRevenue
}

// FromSnapshots builds the plottable Series from a chronologically ordered list
// of daily points (oldest first). Each chart slice references the same points;
// the frontend picks the field to plot. If points is empty the Series is empty.
func FromSnapshots(points []Point, currentDay int) Series {
	return Series{
		Day:       currentDay,
		Cash:      points,
		Revenue:   points,
		Customers: points,
		Burn:      points,
		Valuation: points,
	}
}

// Latest returns the most recent point, or a zero Point if none.
func Latest(points []Point) Point {
	if len(points) == 0 {
		return Point{}
	}
	return points[len(points)-1]
}
