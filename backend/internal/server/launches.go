package server

import (
	"context"
	"net/http"
	"time"

	"github.com/YASSERRMD/Ventiqra/backend/internal/launch"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
	"github.com/YASSERRMD/Ventiqra/backend/internal/sim"
)

type launchResponse struct {
	ID               string    `json:"id"`
	ProductID        string    `json:"product_id"`
	ProductName      string    `json:"product_name"`
	Readiness        float64   `json:"readiness"`
	InitialCustomers int       `json:"initial_customers"`
	LaunchedAt       time.Time `json:"launched_at"`
}

type launchResultResponse struct {
	Readiness        float64         `json:"readiness"`
	InitialCustomers int             `json:"initial_customers"`
	Product          productResponse `json:"product"`
}

// handleLaunchProduct launches a building-stage product once its readiness
// clears the threshold, recording the event and granting initial customers.
func (s *Server) handleLaunchProduct(w http.ResponseWriter, r *http.Request) {
	if s.launches == nil || s.products == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "launch service not configured")
		return
	}

	product, ok := s.loadOwnedProduct(w, r)
	if !ok {
		return
	}
	if product.Stage != repository.ProductBuilding {
		writeError(w, http.StatusConflict, "only products in the building stage can be launched")
		return
	}

	companyID := product.CompanyID
	avgSkill, teamSize := s.teamReadinessInputs(r.Context(), companyID)
	readiness := launch.Readiness(launch.Inputs{
		DevProgress: product.DevProgress,
		AvgSkill:    avgSkill,
		TeamSize:    teamSize,
	})
	if !launch.CanLaunch(readiness) {
		writeJSON(w, http.StatusConflict, map[string]any{
			"error":     "product is not ready to launch",
			"readiness": readiness,
			"required":  launch.MinReadiness,
		})
		return
	}

	seed := sim.SeedFromCompanyID(companyID)
	customers := launch.InitialCustomers(readiness, seed)

	recorded, err := s.launches.Record(r.Context(), product.ID, companyID, readiness, customers)
	if err != nil {
		s.log.Error("record launch failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not record launch")
		return
	}

	if err := s.products.UpdateStage(r.Context(), product.ID, repository.ProductLaunched); err != nil {
		s.log.Error("set launched stage failed", "error", err)
	}

	// Launching is a reputation milestone.
	s.recordReputationEvent(r.Context(), companyID, "launched "+product.Name, 3, s.currentSimDay(r.Context(), companyID))

	// Seed the product's customer state from the launch (initial customers +
	// readiness-derived satisfaction). Idempotent if a state already exists.
	if s.customers != nil {
		if err := s.customers.InitForLaunch(r.Context(), product.ID, companyID,
			customers, int(readiness)); err != nil {
			s.log.Error("init customer state failed", "error", err)
		}
	}

	updated, err := s.products.GetProduct(r.Context(), product.ID)
	if err != nil {
		updated = product
	}
	writeJSON(w, http.StatusCreated, launchResultResponse{
		Readiness:        recorded.Readiness,
		InitialCustomers: recorded.InitialCustomers,
		Product:          toProductResponse(updated),
	})
}

// handleListLaunches returns the owner's launch event history, newest first.
func (s *Server) handleListLaunches(w http.ResponseWriter, r *http.Request) {
	if s.launches == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "launch service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	list, err := s.launches.ListByCompany(r.Context(), companyID)
	if err != nil {
		s.log.Error("list launches failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load launches")
		return
	}

	// Resolve product names for display.
	names := map[string]string{}
	if s.products != nil {
		for _, l := range list {
			if _, ok := names[l.ProductID]; ok {
				continue
			}
			if p, err := s.products.GetProduct(r.Context(), l.ProductID); err == nil {
				names[l.ProductID] = p.Name
			}
		}
	}

	out := make([]launchResponse, 0, len(list))
	for _, l := range list {
		name := names[l.ProductID]
		if name == "" {
			name = l.ProductID
		}
		out = append(out, launchResponse{
			ID: l.ID, ProductID: l.ProductID, ProductName: name,
			Readiness: l.Readiness, InitialCustomers: l.InitialCustomers, LaunchedAt: l.LaunchedAt,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// teamReadinessInputs returns the average skill and headcount of the company's
// team for readiness scoring, or (0, 0) when there is no team.
func (s *Server) teamReadinessInputs(ctx context.Context, companyID string) (float64, int) {
	if s.employees == nil {
		return 0, 0
	}
	team, err := s.employees.ListEmployeesByCompany(ctx, companyID)
	if err != nil || len(team) == 0 {
		return 0, 0
	}
	var skillSum int
	for _, e := range team {
		skillSum += e.Skill
	}
	return float64(skillSum) / float64(len(team)), len(team)
}
