// Save/load simulation handlers. Players can capture the current run into a
// named slot and later restore it, overwriting their live company/simulation
// state with the snapshot.
package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/YASSERRMD/Ventiqra/backend/internal/middleware"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
	"github.com/YASSERRMD/Ventiqra/backend/internal/saves"
)

type saveSlotResponse struct {
	ID        string    `json:"id"`
	Slot      string    `json:"slot"`
	Label     string    `json:"label"`
	Day       int       `json:"day"`
	CashCents int64     `json:"cash_cents"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

type saveRequest struct {
	Slot  *string `json:"slot"`
	Label *string `json:"label"`
}

type loadResultResponse struct {
	Slot    string          `json:"slot"`
	Company companyResponse `json:"company"`
	Day     int             `json:"day"`
	Loaded  time.Time       `json:"loaded_at"`
}

func toSaveSlotResponse(s *repository.SaveSlot) saveSlotResponse {
	return saveSlotResponse{
		ID: s.ID, Slot: s.Slot, Label: s.Label, Day: s.Day,
		CashCents: s.CashCents, Status: s.Status, UpdatedAt: s.UpdatedAt,
	}
}

// captureSnapshot builds a saves.Snapshot from the owner's current company and
// simulation state.
func (s *Server) captureSnapshot(r *http.Request, companyID string, company *repository.Company) saves.Snapshot {
	snap := saves.Snapshot{
		CashCents: company.Cash, Status: string(company.Status),
		Name: company.Name, Industry: company.Industry,
	}
	if s.sim != nil {
		if state, err := s.sim.Get(r.Context(), companyID); err == nil {
			snap.Day = state.Day
			snap.Seed = state.Seed
			snap.Revenue = state.Revenue
			snap.MonthlyBurn = state.MonthlyBurn
		}
	}
	return snap
}

// handleSaveSlot creates or replaces a named save slot from the owner's current
// company + simulation state.
func (s *Server) handleSaveSlot(w http.ResponseWriter, r *http.Request) {
	if s.saveSlots == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "save service not configured")
		return
	}
	var req saveRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Slot == nil || !saves.IsValidSlot(*req.Slot) {
		writeError(w, http.StatusBadRequest, "slot must be 1-32 chars: lowercase letters, digits, - or _")
		return
	}
	label := ""
	if req.Label != nil {
		label = saves.ClampLabel(*req.Label)
	}
	ownerID := middleware.UserIDFrom(r.Context())
	company, err := s.companies.GetLatestCompanyForOwner(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no company found; create a company first")
			return
		}
		s.log.Error("save: load company failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load company")
		return
	}

	// Enforce the slot limit: if this is a new slot and the owner is at the
	// limit, reject rather than silently evicting.
	existing, _ := s.saveSlots.ListByOwner(r.Context(), ownerID)
	alreadyExists := false
	for _, sl := range existing {
		if sl.Slot == *req.Slot {
			alreadyExists = true
			break
		}
	}
	if !alreadyExists && len(existing) >= saves.SlotLimit {
		writeError(w, http.StatusConflict, "save slot limit reached; delete an existing slot first")
		return
	}

	snap := s.captureSnapshot(r, company.ID, company)
	snapBytes, _ := json.Marshal(snap)
	saved, err := s.saveSlots.Upsert(r.Context(), &repository.SaveSlot{
		OwnerID: ownerID, Slot: *req.Slot, Label: label, CompanyID: company.ID,
		Day: snap.Day, CashCents: snap.CashCents, Status: snap.Status, Snapshot: snapBytes,
	})
	if err != nil {
		s.log.Error("save slot failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not save slot")
		return
	}
	writeJSON(w, http.StatusCreated, toSaveSlotResponse(saved))
}

// handleListSaveSlots returns the owner's save slots.
func (s *Server) handleListSaveSlots(w http.ResponseWriter, r *http.Request) {
	if s.saveSlots == nil {
		writeError(w, http.StatusServiceUnavailable, "save service not configured")
		return
	}
	ownerID := middleware.UserIDFrom(r.Context())
	list, err := s.saveSlots.ListByOwner(r.Context(), ownerID)
	if err != nil {
		s.log.Error("list save slots failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load save slots")
		return
	}
	out := make([]saveSlotResponse, 0, len(list))
	for _, sl := range list {
		out = append(out, toSaveSlotResponse(sl))
	}
	writeJSON(w, http.StatusOK, out)
}

// handleLoadSlot restores a save slot onto the owner's company + simulation
// state, overwriting the live run.
func (s *Server) handleLoadSlot(w http.ResponseWriter, r *http.Request) {
	if s.saveSlots == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "save service not configured")
		return
	}
	slot := r.PathValue("slot")
	if !saves.IsValidSlot(slot) {
		writeError(w, http.StatusBadRequest, "invalid slot name")
		return
	}
	ownerID := middleware.UserIDFrom(r.Context())
	saved, err := s.saveSlots.Get(r.Context(), ownerID, slot)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "save slot not found")
			return
		}
		s.log.Error("load slot failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load slot")
		return
	}

	var snap saves.Snapshot
	if err := json.Unmarshal(saved.Snapshot, &snap); err != nil {
		s.log.Error("decode snapshot failed", "error", err)
		writeError(w, http.StatusInternalServerError, "save slot is corrupted")
		return
	}

	// Restore company cash + status.
	if err := s.companies.UpdateCash(r.Context(), saved.CompanyID, snap.CashCents); err != nil {
		s.log.Error("restore cash failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not restore company")
		return
	}
	if err := s.companies.UpdateStatus(r.Context(), saved.CompanyID, repository.CompanyStatus(snap.Status)); err != nil {
		s.log.Error("restore status failed", "error", err)
	}
	// Restore simulation state if the sim service is configured. Init is
	// idempotent for existing rows and re-stamps the seed.
	if s.sim != nil {
		if snap.Seed != 0 {
			_, _ = s.sim.Init(r.Context(), saved.CompanyID, snap.Seed, snap.CashCents, snap.MonthlyBurn)
		}
		_ = s.sim.Save(r.Context(), saved.CompanyID, snap.Day, snap.CashCents, snap.Revenue, snap.MonthlyBurn)
	}

	updated, err := s.companies.GetCompany(r.Context(), saved.CompanyID)
	if err != nil {
		updated = &repository.Company{ID: saved.CompanyID, Cash: snap.CashCents, Status: repository.CompanyStatus(snap.Status)}
	}
	writeJSON(w, http.StatusOK, loadResultResponse{
		Slot:    saved.Slot,
		Company: toCompanyResponse(updated),
		Day:     snap.Day,
		Loaded:  time.Now().UTC(),
	})
}

// handleDeleteSlot removes an owner's save slot.
func (s *Server) handleDeleteSlot(w http.ResponseWriter, r *http.Request) {
	if s.saveSlots == nil {
		writeError(w, http.StatusServiceUnavailable, "save service not configured")
		return
	}
	slot := r.PathValue("slot")
	if !saves.IsValidSlot(slot) {
		writeError(w, http.StatusBadRequest, "invalid slot name")
		return
	}
	ownerID := middleware.UserIDFrom(r.Context())
	if err := s.saveSlots.Delete(r.Context(), ownerID, slot); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "save slot not found")
			return
		}
		s.log.Error("delete slot failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not delete slot")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"slot": slot, "status": "deleted"})
}
