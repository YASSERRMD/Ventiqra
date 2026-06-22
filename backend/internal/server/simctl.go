// Simulation speed-control handlers. The frontend polls the existing tick
// endpoint at the configured speed when in auto mode; these endpoints persist
// the run mode and speed so the state is durable and queryable.
package server

import (
	"context"
	"net/http"

	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
	"github.com/YASSERRMD/Ventiqra/backend/internal/simctl"
)

type simControlResponse struct {
	Mode  string `json:"mode"`   // paused | auto
	Speed int    `json:"speed"`  // 1 | 5 | 30
}

type setSpeedRequest struct {
	Speed *int `json:"speed"`
}

func toSimControlResponse(c *repository.SimControl) simControlResponse {
	return simControlResponse{Mode: c.Mode, Speed: c.Speed}
}

// handleGetSimControl returns the company's current speed/mode.
func (s *Server) handleGetSimControl(w http.ResponseWriter, r *http.Request) {
	if s.simControl == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "sim control not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	c, err := s.simControl.GetOrCreate(r.Context(), companyID)
	if err != nil {
		s.log.Error("get sim control failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load sim control")
		return
	}
	writeJSON(w, http.StatusOK, toSimControlResponse(c))
}

// handlePauseSim sets the mode to paused.
func (s *Server) handlePauseSim(w http.ResponseWriter, r *http.Request) {
	if s.simControl == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "sim control not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	if _, err := s.simControl.GetOrCreate(r.Context(), companyID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not pause")
		return
	}
	if err := s.simControl.SetMode(r.Context(), companyID, string(simctl.ModePaused)); err != nil {
		writeError(w, http.StatusInternalServerError, "could not pause")
		return
	}
	writeJSON(w, http.StatusOK, simControlResponse{Mode: string(simctl.ModePaused), Speed: s.currentSpeed(r.Context(), companyID)})
}

// handleResumeSim sets the mode to auto at the current speed.
func (s *Server) handleResumeSim(w http.ResponseWriter, r *http.Request) {
	if s.simControl == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "sim control not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	c, err := s.simControl.GetOrCreate(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not resume")
		return
	}
	if err := s.simControl.SetMode(r.Context(), companyID, string(simctl.ModeAuto)); err != nil {
		writeError(w, http.StatusInternalServerError, "could not resume")
		return
	}
	writeJSON(w, http.StatusOK, simControlResponse{Mode: string(simctl.ModeAuto), Speed: c.Speed})
}

// handleSetSimSpeed updates the tick speed.
func (s *Server) handleSetSimSpeed(w http.ResponseWriter, r *http.Request) {
	if s.simControl == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "sim control not configured")
		return
	}
	var req setSpeedRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Speed == nil || !simctl.IsValidSpeed(simctl.Speed(*req.Speed)) {
		writeError(w, http.StatusBadRequest, "speed must be 1, 5, or 30")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	if _, err := s.simControl.GetOrCreate(r.Context(), companyID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not set speed")
		return
	}
	if err := s.simControl.SetSpeed(r.Context(), companyID, *req.Speed); err != nil {
		writeError(w, http.StatusInternalServerError, "could not set speed")
		return
	}
	writeJSON(w, http.StatusOK, simControlResponse{Mode: s.currentMode(r.Context(), companyID), Speed: *req.Speed})
}

// currentMode returns the company's mode, defaulting to paused on error.
func (s *Server) currentMode(ctx context.Context, companyID string) string {
	if s.simControl == nil {
		return string(simctl.ModePaused)
	}
	if c, err := s.simControl.Get(ctx, companyID); err == nil {
		return c.Mode
	}
	return string(simctl.ModePaused)
}

// currentSpeed returns the company's speed, defaulting to 1 on error.
func (s *Server) currentSpeed(ctx context.Context, companyID string) int {
	if s.simControl == nil {
		return int(simctl.Speed1x)
	}
	if c, err := s.simControl.Get(ctx, companyID); err == nil {
		return c.Speed
	}
	return int(simctl.Speed1x)
}
