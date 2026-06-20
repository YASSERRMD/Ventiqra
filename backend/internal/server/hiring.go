package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/YASSERRMD/Ventiqra/backend/internal/hiring"
	"github.com/YASSERRMD/Ventiqra/backend/internal/middleware"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
	"github.com/YASSERRMD/Ventiqra/backend/internal/sim"
)

type candidateResponse struct {
	Index                    int     `json:"index"`
	Role                     string  `json:"role"`
	Name                     string  `json:"name"`
	Quality                  string  `json:"quality"`
	Skill                    int     `json:"skill"`
	SalaryExpectationCents   int64   `json:"salary_expectation_cents"`
	HiringFeeCents           int64   `json:"hiring_fee_cents"`
	AcceptanceChance         float64 `json:"acceptance_chance"`
}

type candidatePoolResponse struct {
	Day         int                 `json:"day"`
	Candidates  []candidateResponse `json:"candidates"`
}

type hireResultResponse struct {
	Accepted bool              `json:"accepted"`
	Message  string            `json:"message"`
	Employee *employeeResponse `json:"employee,omitempty"`
}

// loadCompanyRound resolves the owner's latest company plus its deterministic
// (seed, day) round pair, using the persisted simulation state or a lazy day-0
// default.
func (s *Server) loadCompanyRound(r *http.Request) (*repository.Company, int64, int, error) {
	ownerID := middleware.UserIDFrom(r.Context())
	company, err := s.companies.GetLatestCompanyForOwner(r.Context(), ownerID)
	if err != nil {
		return nil, 0, 0, err
	}
	seed := sim.SeedFromCompanyID(company.ID)
	day := 0
	if s.sim != nil {
		if state, err := s.sim.Get(r.Context(), company.ID); err == nil {
			seed = state.Seed
			day = state.Day
		} else if !errors.Is(err, repository.ErrNotFound) {
			return nil, 0, 0, err
		}
	}
	return company, seed, day, nil
}

// handleListCandidates returns the deterministic candidate pool for the round.
func (s *Server) handleListCandidates(w http.ResponseWriter, r *http.Request) {
	if s.employees == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "hiring service not configured")
		return
	}
	company, seed, day, err := s.loadCompanyRound(r)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no company found")
			return
		}
		s.log.Error("candidates: load company failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load candidates")
		return
	}
	_ = company

	pool := hiring.GeneratePool(seed, int64(day))
	out := make([]candidateResponse, 0, len(pool))
	for _, c := range pool {
		out = append(out, toCandidateResponse(c))
	}
	writeJSON(w, http.StatusOK, candidatePoolResponse{Day: day, Candidates: out})
}

// handleHireCandidate attempts to hire a candidate by round index. The decision
// is deterministic for the round; on acceptance the hiring fee is deducted from
// company cash and the employee is created.
func (s *Server) handleHireCandidate(w http.ResponseWriter, r *http.Request) {
	if s.employees == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "hiring service not configured")
		return
	}
	company, seed, day, err := s.loadCompanyRound(r)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no company found")
			return
		}
		s.log.Error("hire: load company failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not process offer")
		return
	}

	indexStr := r.PathValue("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 || index >= hiring.PoolSize {
		writeError(w, http.StatusBadRequest, "invalid candidate index")
		return
	}

	pool := hiring.GeneratePool(seed, int64(day))
	candidate := pool[index]

	if company.Cash < candidate.HiringFee {
		writeJSON(w, http.StatusOK, hireResultResponse{
			Accepted: false,
			Message:  "Insufficient cash to cover the hiring fee.",
		})
		return
	}

	if !hiring.OfferAccepted(seed, int64(day), index) {
		writeJSON(w, http.StatusOK, hireResultResponse{
			Accepted: false,
			Message:  candidate.Name + " declined your offer.",
		})
		return
	}

	emp, err := s.employees.CreateEmployee(r.Context(), company.ID, candidate.Name,
		repository.EmployeeRole(candidate.Role), candidate.SalaryExpectation, candidate.Skill, 70)
	if err != nil {
		s.log.Error("hire: create employee failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not hire employee")
		return
	}

	newCash := company.Cash - candidate.HiringFee
	if err := s.companies.UpdateCash(r.Context(), company.ID, newCash); err != nil {
		s.log.Error("hire: deduct fee failed", "error", err)
	}

	resp := toEmployeeResponse(emp)
	writeJSON(w, http.StatusCreated, hireResultResponse{
		Accepted: true,
		Message:  candidate.Name + " accepted and joined the team!",
		Employee: &resp,
	})
}

func toCandidateResponse(c hiring.Candidate) candidateResponse {
	return candidateResponse{
		Index: c.Index, Role: c.Role, Name: c.Name, Quality: c.Quality,
		Skill: c.Skill, SalaryExpectationCents: c.SalaryExpectation,
		HiringFeeCents: c.HiringFee, AcceptanceChance: c.AcceptanceChance,
	}
}
