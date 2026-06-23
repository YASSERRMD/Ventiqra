// Achievement handlers and tick evaluation. The endpoint lists awarded
// achievements; the tick evaluates company state and awards newly-qualified
// milestones.
package server

import (
	"context"
	"net/http"

	"github.com/YASSERRMD/Ventiqra/backend/internal/achievements"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
)

type achievementResponse struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AwardedDay  int    `json:"awarded_day"`
}

func (s *Server) handleListAchievements(w http.ResponseWriter, r *http.Request) {
	if s.achievements == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "achievements service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	list, err := s.achievements.ListByCompany(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load achievements")
		return
	}
	out := make([]achievementResponse, 0, len(list))
	for _, a := range list {
		name, desc := a.Key, ""
		if d, ok := achievements.FindDefinition(achievements.Key(a.Key)); ok {
			name, desc = d.Name, d.Description
		}
		out = append(out, achievementResponse{Key: a.Key, Name: name, Description: desc, AwardedDay: a.AwardedDay})
	}
	writeJSON(w, http.StatusOK, out)
}

// evaluateAchievements checks the company's state against the achievement
// catalog and awards any newly-qualified milestones. Called from the tick.
func (s *Server) evaluateAchievements(ctx context.Context, companyID string, day int) {
	if s.achievements == nil {
		return
	}
	already, err := s.achievements.AwardedSet(ctx, companyID)
	if err != nil {
		return
	}
	aw := achievements.Awarded{}
	for k := range already {
		aw[achievements.Key(k)] = true
	}
	state := s.achievementState(ctx, companyID)
	newly := achievements.Evaluate(state, aw)
	for _, k := range newly {
		if err := s.achievements.Award(ctx, companyID, string(k), day); err != nil {
			continue
		}
		if d, ok := achievements.FindDefinition(k); ok {
			s.recordTimeline(ctx, companyID, "milestone", "Achievement: "+d.Name, d.Description, day)
		}
	}
}

// achievementState assembles the snapshot the engine evaluates.
func (s *Server) achievementState(ctx context.Context, companyID string) achievements.State {
	st := achievements.State{}
	if s.products != nil {
		if list, err := s.products.ListProductsByCompany(ctx, companyID); err == nil {
			for _, p := range list {
				if p.Stage == repository.ProductLaunched {
					st.ProductsLaunched++
				}
			}
		}
	}
	if s.funding != nil {
		if count, err := s.funding.CountByCompany(ctx, companyID); err == nil {
			st.FundingRounds = count
		}
	}
	if s.sim != nil {
		if state, err := s.sim.Get(ctx, companyID); err == nil {
			st.RevenuePerMonth = state.Revenue * 30
			st.MonthlyBurn = state.MonthlyBurn
		}
	}
	if s.employees != nil {
		if list, err := s.employees.ListEmployeesByCompany(ctx, companyID); err == nil {
			st.Employees = len(list)
		}
	}
	st.Customers = s.currentCustomerCount(ctx, companyID)
	if s.snapshots != nil {
		if list, err := s.snapshots.ListByCompany(ctx, companyID, 1); err == nil && len(list) > 0 {
			st.ValuationCents = list[0].ValuationCents
		}
	}
	return st
}

var _ = repository.ErrNotFound
