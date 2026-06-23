// Leaderboard handlers and bankruptcy finalization. The endpoint returns the
// local top scores; the tick records a finalized entry when a company goes
// bankrupt.
package server

import (
	"context"
	"net/http"

	"github.com/YASSERRMD/Ventiqra/backend/internal/leaderboard"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
)

type leaderboardEntryResponse struct {
	ID            string `json:"id"`
	CompanyName   string `json:"company_name"`
	Score         int64  `json:"score"`
	DaysSurvived  int    `json:"days_survived"`
	PeakValuation int64  `json:"peak_valuation"`
	Outcome       string `json:"outcome"`
	CreatedAt     string `json:"created_at"`
}

func (s *Server) handleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	if s.leaderboard == nil {
		writeError(w, http.StatusServiceUnavailable, "leaderboard service not configured")
		return
	}
	entries, err := s.leaderboard.Top(r.Context(), 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load leaderboard")
		return
	}
	out := make([]leaderboardEntryResponse, 0, len(entries))
	for _, e := range entries {
		out = append(out, leaderboardEntryResponse{
			ID: e.ID, CompanyName: e.CompanyName, Score: e.Score,
			DaysSurvived: e.DaysSurvived, PeakValuation: e.PeakValuation,
			Outcome: e.Outcome, CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// finalizeRun records a leaderboard entry for a company when it ends
// (bankruptcy). Best-effort: de-duplicates by (owner, company name).
func (s *Server) finalizeRun(ctx context.Context, company *repository.Company, outcome leaderboard.Outcome) {
	if s.leaderboard == nil {
		return
	}
	has, _ := s.leaderboard.HasEntryForCompany(ctx, company.OwnerID, company.Name)
	if has {
		return
	}
	day := s.currentSimDay(ctx, company.ID)
	var peakValuation int64
	if s.snapshots != nil {
		if list, err := s.snapshots.ListByCompany(ctx, company.ID, 180); err == nil {
			for _, m := range list {
				if m.ValuationCents > peakValuation {
					peakValuation = m.ValuationCents
				}
			}
		}
	}
	customers := s.currentCustomerCount(ctx, company.ID)
	achCount := 0
	if s.achievements != nil {
		if list, err := s.achievements.ListByCompany(ctx, company.ID); err == nil {
			achCount = len(list)
		}
	}
	score := leaderboard.Score(leaderboard.Input{
		DaysSurvived: day, PeakValuation: peakValuation,
		Customers: customers, Achievements: achCount, Outcome: outcome,
	})
	_, _ = s.leaderboard.Record(ctx, &repository.LeaderboardEntry{
		OwnerID: company.OwnerID, CompanyName: company.Name, Score: score,
		DaysSurvived: day, PeakValuation: peakValuation, Outcome: string(outcome),
	})
}
