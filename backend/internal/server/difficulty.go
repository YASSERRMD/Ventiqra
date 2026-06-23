// Difficulty handlers: get the company's current difficulty and its multipliers,
// and set the difficulty level.
package server

import (
	"net/http"

	"github.com/YASSERRMD/Ventiqra/backend/internal/difficulty"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
)

type difficultyResponse struct {
	Level       string                 `json:"level"`
	Multipliers difficulty.Multipliers `json:"multipliers"`
}

type setDifficultyRequest struct {
	Level *string `json:"level"`
}

func (s *Server) handleGetDifficulty(w http.ResponseWriter, r *http.Request) {
	if s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "company service not configured")
		return
	}
	company, ok := s.ownerCompany(w, r)
	if !ok {
		return
	}
	level := company.Difficulty
	if level == "" {
		level = string(difficulty.LevelNormal)
	}
	writeJSON(w, http.StatusOK, difficultyResponse{
		Level: level, Multipliers: difficulty.For(difficulty.Level(level)),
	})
}

func (s *Server) handleSetDifficulty(w http.ResponseWriter, r *http.Request) {
	if s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "company service not configured")
		return
	}
	var req setDifficultyRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Level == nil || !difficulty.IsValid(difficulty.Level(*req.Level)) {
		writeError(w, http.StatusBadRequest, "level must be easy, normal, hard, brutal, or custom")
		return
	}
	company, ok := s.ownerCompany(w, r)
	if !ok {
		return
	}
	if err := s.companies.SetDifficulty(r.Context(), company.ID, *req.Level); err != nil {
		writeError(w, http.StatusInternalServerError, "could not set difficulty")
		return
	}
	writeJSON(w, http.StatusOK, difficultyResponse{
		Level: *req.Level, Multipliers: difficulty.For(difficulty.Level(*req.Level)),
	})
}

// ownerCompany is a shorthand that resolves the owner's latest company.
func (s *Server) ownerCompany(w http.ResponseWriter, r *http.Request) (*repository.Company, bool) {
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return nil, false
	}
	c, err := s.companies.GetCompany(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusNotFound, "company not found")
		return nil, false
	}
	return c, true
}
