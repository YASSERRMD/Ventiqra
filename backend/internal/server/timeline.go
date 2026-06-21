// Timeline handlers and milestone recording helpers. The timeline unifies
// discrete company milestones (founding, launches, funding, decisions, crises)
// into one chronological history the dashboard renders.
package server

import (
	"context"
	"net/http"

	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
	"github.com/YASSERRMD/Ventiqra/backend/internal/timeline"
)

type timelineEntryResponse struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Title       string `json:"title"`
	Description string `json:"description"`
	SimDay      int    `json:"sim_day"`
	CreatedAt   string `json:"created_at"`
}

type timelineResponse struct {
	Day     int                       `json:"day"`
	Entries []timelineEntryResponse   `json:"entries"`
	Monthly []timeline.MonthlySummary `json:"monthly_summary"`
}

// handleGetTimeline returns the company's unified timeline and monthly summary.
func (s *Server) handleGetTimeline(w http.ResponseWriter, r *http.Request) {
	if s.timeline == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "timeline service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	entries, err := s.timeline.ListByCompany(r.Context(), companyID, 50)
	if err != nil {
		s.log.Error("list timeline failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load timeline")
		return
	}
	out := make([]timelineEntryResponse, 0, len(entries))
	for _, e := range entries {
		out = append(out, timelineEntryResponse{
			ID: e.ID, Kind: e.Kind, Title: e.Title, Description: e.Description,
			SimDay: e.SimDay, CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	day := s.currentSimDay(r.Context(), companyID)
	monthly := s.buildMonthlySummary(r.Context(), companyID, day)

	writeJSON(w, http.StatusOK, timelineResponse{
		Day:     day,
		Entries: out,
		Monthly: monthly,
	})
}

// buildMonthlySummary assembles per-month summaries from the current sim state
// and per-month timeline event counts.
func (s *Server) buildMonthlySummary(ctx context.Context, companyID string, day int) []timeline.MonthlySummary {
	var revenue, burn int64
	if s.sim != nil {
		if state, err := s.sim.Get(ctx, companyID); err == nil {
			revenue = state.Revenue
			burn = state.MonthlyBurn
		}
	}
	months := timeline.MonthForDay(day)
	if months < 1 {
		months = 1
	}
	out := make([]timeline.MonthlySummary, 0, months)
	for m := 1; m <= months; m++ {
		startDay, endDay := timeline.MonthBounds(m)
		count := 0
		if s.timeline != nil {
			if n, err := s.timeline.CountInDayRange(ctx, companyID, startDay, endDay); err == nil {
				count = n
			}
		}
		out = append(out, timeline.MonthlySummary{
			Month: m, StartDay: startDay, EndDay: endDay,
			RevenueEnd: revenue, BurnEnd: burn, EventsCount: count,
		})
	}
	return out
}

// recordTimeline is a best-effort helper that appends a milestone entry. Errors
// are logged but never returned, so callers can fire-and-forget from flows that
// must not fail because the timeline write failed.
func (s *Server) recordTimeline(ctx context.Context, companyID, kind, title, description string, day int) {
	if s.timeline == nil {
		return
	}
	_, err := s.timeline.Record(ctx, &repository.TimelineEvent{
		CompanyID: companyID, Kind: kind, Title: title, Description: description, SimDay: day,
	})
	if err != nil {
		s.log.Error("record timeline failed", "kind", kind, "error", err)
	}
}
