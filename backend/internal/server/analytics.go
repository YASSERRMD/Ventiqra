// Analytics handler and snapshot capture. The analytics endpoint returns the
// plottable metric series; the tick helper records one daily snapshot so the
// series accumulates as the simulation advances.
package server

import (
	"context"
	"net/http"

	"github.com/YASSERRMD/Ventiqra/backend/internal/analytics"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
)

type analyticsPointResponse struct {
	Day            int   `json:"day"`
	CashCents      int64 `json:"cash_cents"`
	RevenueCents   int64 `json:"revenue_cents"`
	MonthlyBurn    int64 `json:"monthly_burn"`
	Customers      int   `json:"customers"`
	ValuationCents int64 `json:"valuation_cents"`
}

type analyticsResponse struct {
	Day       int                    `json:"day"`
	Cash      []analyticsPointResponse `json:"cash"`
	Revenue   []analyticsPointResponse `json:"revenue"`
	Customers []analyticsPointResponse `json:"customers"`
	Burn      []analyticsPointResponse `json:"burn"`
	Valuation []analyticsPointResponse `json:"valuation"`
}

// handleGetAnalytics returns the company's metric series for the dashboard
// charts. Each slice holds the same points; the frontend plots the relevant
// field per chart.
func (s *Server) handleGetAnalytics(w http.ResponseWriter, r *http.Request) {
	if s.snapshots == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "analytics service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	rows, err := s.snapshots.ListByCompany(r.Context(), companyID, 180)
	if err != nil {
		s.log.Error("list snapshots failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load analytics")
		return
	}
	points := make([]analytics.Point, 0, len(rows))
	for _, m := range rows {
		points = append(points, analytics.Point{
			Day: m.SimDay, CashCents: m.CashCents, RevenueCents: m.RevenueCents,
			MonthlyBurn: m.MonthlyBurn, Customers: m.Customers, ValuationCents: m.ValuationCents,
		})
	}
	day := s.currentSimDay(r.Context(), companyID)
	series := analytics.FromSnapshots(points, day)
	writeJSON(w, http.StatusOK, analyticsResponse{
		Day:       series.Day,
		Cash:      toPointResponses(points),
		Revenue:   toPointResponses(points),
		Customers: toPointResponses(points),
		Burn:      toPointResponses(points),
		Valuation: toPointResponses(points),
	})
}

func toPointResponses(points []analytics.Point) []analyticsPointResponse {
	out := make([]analyticsPointResponse, 0, len(points))
	for _, p := range points {
		out = append(out, analyticsPointResponse{
			Day: p.Day, CashCents: p.CashCents, RevenueCents: p.RevenueCents,
			MonthlyBurn: p.MonthlyBurn, Customers: p.Customers, ValuationCents: p.ValuationCents,
		})
	}
	return out
}

// recordSnapshot captures the day's headline metrics after a tick. Best-effort:
// failures are logged but never break the tick.
func (s *Server) recordSnapshot(ctx context.Context, companyID string, day int, cash, revenue, burn int64, customers int) {
	if s.snapshots == nil {
		return
	}
	// Monthly revenue estimate for valuation.
	monthlyRevenue := revenue * 30
	val := analytics.Valuation(cash, monthlyRevenue)
	if err := s.snapshots.Upsert(ctx, &repository.MetricSnapshot{
		CompanyID: companyID, SimDay: day, CashCents: cash, RevenueCents: revenue,
		MonthlyBurn: burn, Customers: customers, ValuationCents: val,
	}); err != nil {
		s.log.Error("record snapshot failed", "error", err)
	}
}
