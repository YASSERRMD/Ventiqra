// Package finance computes the components of a company's monthly burn and its
// profit/loss. Formulas are pure so the finance breakdown is reproducible.
package finance

// BaseMonthlyOperatingCents is the fixed monthly overhead beyond payroll,
// infrastructure, and marketing (tools, legal, admin).
const BaseMonthlyOperatingCents int64 = 500_000

// InfraBaseCents is the floor monthly infrastructure (hosting) cost.
const InfraBaseCents int64 = 50_000

// InfraPerCustomerCents is the incremental monthly hosting cost per customer.
const InfraPerCustomerCents int64 = 5

// Breakdown itemizes the monthly burn so the dashboard and P&L can show where
// money goes.
type Breakdown struct {
	Base       int64 // fixed operating overhead
	Salaries   int64 // monthly payroll
	Infra      int64 // hosting scaled by customers
	Marketing  int64 // marketing budget
}

// Total returns the sum of all monthly burn components.
func (b Breakdown) Total() int64 {
	return b.Base + b.Salaries + b.Infra + b.Marketing
}

// InfraCost returns the monthly infrastructure cost for a customer base.
func InfraCost(totalCustomers int) int64 {
	if totalCustomers < 0 {
		totalCustomers = 0
	}
	return InfraBaseCents + int64(totalCustomers)*InfraPerCustomerCents
}

// MonthlyBreakdown assembles the burn breakdown from its inputs.
func MonthlyBreakdown(payroll, marketing int64, totalCustomers int) Breakdown {
	if payroll < 0 {
		payroll = 0
	}
	if marketing < 0 {
		marketing = 0
	}
	return Breakdown{
		Base:      BaseMonthlyOperatingCents,
		Salaries:  payroll,
		Infra:     InfraCost(totalCustomers),
		Marketing: marketing,
	}
}

// ProfitLoss returns the monthly profit (positive) or loss (negative) given a
// monthly revenue and total monthly burn.
func ProfitLoss(monthlyRevenue, monthlyBurn int64) int64 {
	return monthlyRevenue - monthlyBurn
}

// MonthlyRevenueFromDaily scales a daily revenue figure to a 30-day month.
func MonthlyRevenueFromDaily(dailyRevenue int64) int64 {
	return dailyRevenue * 30
}
