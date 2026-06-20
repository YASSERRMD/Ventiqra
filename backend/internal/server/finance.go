package server

import (
	"context"
	"net/http"

	"github.com/YASSERRMD/Ventiqra/backend/internal/develop"
	"github.com/YASSERRMD/Ventiqra/backend/internal/finance"
)

type financeBreakdownResponse struct {
	Base      int64 `json:"base_cents"`
	Salaries  int64 `json:"salary_cents"`
	Infra     int64 `json:"infra_cents"`
	Marketing int64 `json:"marketing_cents"`
	Total     int64 `json:"total_burn_cents"`
}

type financeResponse struct {
	MarketingBudgetCents int64                  `json:"marketing_budget_cents"`
	MonthlyRevenueCents  int64                  `json:"monthly_revenue_cents"`
	Burn                 financeBreakdownResponse `json:"burn"`
	ProfitLossCents      int64                  `json:"profit_loss_cents"`
	Day                  int                    `json:"day"`
}

type updateFinanceRequest struct {
	MarketingBudgetCents *int64 `json:"marketing_budget_cents"`
}

// handleGetFinance returns the company's finance breakdown and monthly P&L.
func (s *Server) handleGetFinance(w http.ResponseWriter, r *http.Request) {
	if s.finance == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "finance service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}

	fin, err := s.finance.GetOrCreate(r.Context(), companyID)
	if err != nil {
		s.log.Error("finance load failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load finance")
		return
	}

	payroll, totalCustomers, dailyRevenue, day := s.financeInputs(r.Context(), companyID)
	burn := finance.MonthlyBreakdown(payroll, fin.MarketingBudgetCents, totalCustomers)
	monthlyRevenue := finance.MonthlyRevenueFromDaily(dailyRevenue)

	writeJSON(w, http.StatusOK, financeResponse{
		MarketingBudgetCents: fin.MarketingBudgetCents,
		MonthlyRevenueCents:  monthlyRevenue,
		Burn: financeBreakdownResponse{
			Base: burn.Base, Salaries: burn.Salaries, Infra: burn.Infra,
			Marketing: burn.Marketing, Total: burn.Total(),
		},
		ProfitLossCents: finance.ProfitLoss(monthlyRevenue, burn.Total()),
		Day:             day,
	})
}

// handleUpdateFinance updates the company's finance settings (marketing budget).
func (s *Server) handleUpdateFinance(w http.ResponseWriter, r *http.Request) {
	if s.finance == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "finance service not configured")
		return
	}
	var req updateFinanceRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.MarketingBudgetCents == nil || *req.MarketingBudgetCents < 0 {
		writeError(w, http.StatusBadRequest, "marketing_budget_cents must be a non-negative integer")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	fin, err := s.finance.SetMarketingBudget(r.Context(), companyID, *req.MarketingBudgetCents)
	if err != nil {
		s.log.Error("update finance failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update finance")
		return
	}

	payroll, totalCustomers, dailyRevenue, day := s.financeInputs(r.Context(), companyID)
	burn := finance.MonthlyBreakdown(payroll, fin.MarketingBudgetCents, totalCustomers)
	monthlyRevenue := finance.MonthlyRevenueFromDaily(dailyRevenue)
	writeJSON(w, http.StatusOK, financeResponse{
		MarketingBudgetCents: fin.MarketingBudgetCents,
		MonthlyRevenueCents:  monthlyRevenue,
		Burn: financeBreakdownResponse{
			Base: burn.Base, Salaries: burn.Salaries, Infra: burn.Infra,
			Marketing: burn.Marketing, Total: burn.Total(),
		},
		ProfitLossCents: finance.ProfitLoss(monthlyRevenue, burn.Total()),
		Day:             day,
	})
}

// financeInputs gathers the live payroll, total customers, daily revenue, and
// simulated day used to render the finance breakdown.
func (s *Server) financeInputs(ctx context.Context, companyID string) (payroll int64, totalCustomers int, dailyRevenue int64, day int) {
	day = 0
	if s.sim != nil {
		if state, err := s.sim.Get(ctx, companyID); err == nil {
			day = state.Day
			dailyRevenue = state.Revenue
		}
	}
	if s.employees != nil {
		if team, err := s.employees.ListEmployeesByCompany(ctx, companyID); err == nil {
			sals := make([]int64, 0, len(team))
			for _, e := range team {
				sals = append(sals, e.SalaryCents)
			}
			payroll = develop.MonthlyPayroll(sals)
		}
	}
	if s.customers != nil {
		if list, err := s.customers.ListByCompany(ctx, companyID); err == nil {
			for _, c := range list {
				totalCustomers += c.Total
			}
		}
	}
	return payroll, totalCustomers, dailyRevenue, day
}
