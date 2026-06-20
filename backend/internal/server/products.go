package server

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/YASSERRMD/Ventiqra/backend/internal/middleware"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
)

type productResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Stage       string    `json:"stage"`
	DevProgress float64   `json:"dev_progress"`
	PriceCents  *int64    `json:"price_cents"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type createProductRequest struct {
	Name string `json:"name"`
}

type updateProductStageRequest struct {
	Stage string `json:"stage"`
}

type updateProductProgressRequest struct {
	Progress *float64 `json:"progress"`
}

func toProductResponse(p *repository.Product) productResponse {
	return productResponse{
		ID: p.ID, Name: p.Name, Slug: p.Slug, Stage: string(p.Stage),
		DevProgress: p.DevProgress, PriceCents: p.PriceCents,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

// handleCreateProduct creates a new product for the owner's latest company.
func (s *Server) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	if s.products == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "product service not configured")
		return
	}

	var req createProductRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) < 1 || len(req.Name) > 120 {
		writeError(w, http.StatusBadRequest, "name must be 1-120 characters")
		return
	}

	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}

	product, err := s.products.CreateProduct(r.Context(), companyID, req.Name)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			writeError(w, http.StatusConflict, "product slug already exists for this company")
			return
		}
		s.log.Error("create product failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not create product")
		return
	}
	writeJSON(w, http.StatusCreated, toProductResponse(product))
}

// handleListProducts returns the products for the owner's latest company.
func (s *Server) handleListProducts(w http.ResponseWriter, r *http.Request) {
	if s.products == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "product service not configured")
		return
	}

	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}

	list, err := s.products.ListProductsByCompany(r.Context(), companyID)
	if err != nil {
		s.log.Error("list products failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load products")
		return
	}
	if list == nil {
		list = []*repository.Product{}
	}
	resp := make([]productResponse, 0, len(list))
	for _, p := range list {
		resp = append(resp, toProductResponse(p))
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleUpdateProductStage updates a product's lifecycle stage.
func (s *Server) handleUpdateProductStage(w http.ResponseWriter, r *http.Request) {
	if s.products == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "product service not configured")
		return
	}

	var req updateProductStageRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Stage = strings.TrimSpace(req.Stage)
	if !repository.ValidProductStages[repository.ProductStage(req.Stage)] {
		writeError(w, http.StatusBadRequest, "stage must be one of idea, building, launched, retired")
		return
	}

	_, ok := s.loadOwnedProduct(w, r)
	if !ok {
		return
	}

	id := r.PathValue("id")
	if err := s.products.UpdateStage(r.Context(), id, repository.ProductStage(req.Stage)); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "product not found")
			return
		}
		s.log.Error("update product stage failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update product")
		return
	}
	updated, err := s.products.GetProduct(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"id": id, "stage": req.Stage})
		return
	}
	writeJSON(w, http.StatusOK, toProductResponse(updated))
}

// handleUpdateProductProgress updates a product's development progress.
func (s *Server) handleUpdateProductProgress(w http.ResponseWriter, r *http.Request) {
	if s.products == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "product service not configured")
		return
	}

	var req updateProductProgressRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Progress == nil {
		writeError(w, http.StatusBadRequest, "progress is required")
		return
	}
	if *req.Progress < 0 || *req.Progress > 100 {
		writeError(w, http.StatusBadRequest, "progress must be between 0 and 100")
		return
	}

	_, ok := s.loadOwnedProduct(w, r)
	if !ok {
		return
	}

	id := r.PathValue("id")
	if err := s.products.UpdateProgress(r.Context(), id, *req.Progress); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "product not found")
			return
		}
		s.log.Error("update product progress failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update product")
		return
	}
	updated, err := s.products.GetProduct(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"id": id, "dev_progress": *req.Progress})
		return
	}
	writeJSON(w, http.StatusOK, toProductResponse(updated))
}

// ownerCompanyID resolves the id of the owner's latest company, writing the
// appropriate error response when it cannot.
func (s *Server) ownerCompanyID(w http.ResponseWriter, r *http.Request) (string, bool) {
	ownerID := middleware.UserIDFrom(r.Context())
	company, err := s.companies.GetLatestCompanyForOwner(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no company found")
			return "", false
		}
		s.log.Error("load company failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load company")
		return "", false
	}
	return company.ID, true
}

// loadOwnedProduct loads a product by path id and confirms it belongs to the
// authenticated owner's latest company.
func (s *Server) loadOwnedProduct(w http.ResponseWriter, r *http.Request) (*repository.Product, bool) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "product id is required")
		return nil, false
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return nil, false
	}
	product, err := s.products.GetProduct(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "product not found")
			return nil, false
		}
		s.log.Error("load product failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load product")
		return nil, false
	}
	if product.CompanyID != companyID {
		writeError(w, http.StatusForbidden, "forbidden")
		return nil, false
	}
	return product, true
}
