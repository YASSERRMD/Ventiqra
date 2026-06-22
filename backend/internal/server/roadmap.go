// Roadmap handlers: feature backlog CRUD and a develop action that advances
// the active feature. Shipped features contribute to product value.
package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
	"github.com/YASSERRMD/Ventiqra/backend/internal/roadmap"
)

type featureResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Priority    int     `json:"priority"`
	Status      string  `json:"status"`
	Progress    int     `json:"progress"`
	ValuePoints int     `json:"value_points"`
	StartedDay  *int    `json:"started_day"`
	ShippedDay  *int    `json:"shipped_day"`
	ProductID   *string `json:"product_id"`
}

type featureInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Priority    *int    `json:"priority"`
	ValuePoints *int    `json:"value_points"`
	ProductID   *string `json:"product_id"`
}

type developRequest struct {
	Points *int `json:"points"`
}

func toFeatureResponse(f *repository.Feature) featureResponse {
	return featureResponse{
		ID: f.ID, Name: f.Name, Description: f.Description, Priority: f.Priority,
		Status: f.Status, Progress: f.Progress, ValuePoints: f.ValuePoints,
		StartedDay: f.StartedDay, ShippedDay: f.ShippedDay, ProductID: f.ProductID,
	}
}

func (s *Server) handleListFeatures(w http.ResponseWriter, r *http.Request) {
	if s.features == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "roadmap service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	list, err := s.features.ListByCompany(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load features")
		return
	}
	out := make([]featureResponse, 0, len(list))
	for _, f := range list {
		out = append(out, toFeatureResponse(f))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleCreateFeature(w http.ResponseWriter, r *http.Request) {
	if s.features == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "roadmap service not configured")
		return
	}
	var req featureInput
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Name == nil || *req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	priority := 0
	if req.Priority != nil {
		priority = *req.Priority
	}
	value := 10
	if req.ValuePoints != nil {
		value = *req.ValuePoints
	}
	desc := ""
	if req.Description != nil {
		desc = *req.Description
	}
	created, err := s.features.Create(r.Context(), &repository.Feature{
		CompanyID: companyID, ProductID: req.ProductID, Name: *req.Name,
		Description: desc, Priority: priority, Status: string(roadmap.StatusBacklog),
		Progress: 0, ValuePoints: value,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create feature")
		return
	}
	writeJSON(w, http.StatusCreated, toFeatureResponse(created))
}

func (s *Server) handleDeleteFeature(w http.ResponseWriter, r *http.Request) {
	if s.features == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "roadmap service not configured")
		return
	}
	id := r.PathValue("id")
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	if err := s.features.Delete(r.Context(), id, companyID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "feature not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not delete feature")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "status": "deleted"})
}

func (s *Server) handleDevelopFeature(w http.ResponseWriter, r *http.Request) {
	if s.features == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "roadmap service not configured")
		return
	}
	id := r.PathValue("id")
	var req developRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	points := 10
	if req.Points != nil {
		points = *req.Points
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	f, err := s.features.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "feature not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not load feature")
		return
	}
	if f.CompanyID != companyID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	if f.Status == string(roadmap.StatusShipped) {
		writeError(w, http.StatusConflict, "feature already shipped")
		return
	}
	rf := roadmap.Feature{Progress: f.Progress, Status: roadmap.Status(f.Status)}
	newProgress, shipped := roadmap.DevelopProgress(&rf, points)
	status := string(roadmap.StatusDeveloping)
	var shippedDay *int
	if shipped {
		status = string(roadmap.StatusShipped)
		day := s.currentSimDay(r.Context(), companyID)
		shippedDay = &day
		if err := s.features.UpdateProgress(r.Context(), id, newProgress, status, shippedDay); err == nil {
				s.recordTimeline(r.Context(), companyID, "milestone", "Shipped feature: "+f.Name,
					"Value points: "+formatInt(rf.ValuePoints), day)
		}
	} else {
		_ = s.features.UpdateProgress(r.Context(), id, newProgress, status, nil)
	}
	updated, _ := s.features.Get(r.Context(), id)
	writeJSON(w, http.StatusOK, map[string]any{
		"feature":  toFeatureResponse(updated),
		"shipped":  shipped,
	})
}

// formatInt is a local int→string helper (avoids importing strconv for one call
// and avoids colliding with test-only helpers named itoa).
func formatInt(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

var _ = time.Now
