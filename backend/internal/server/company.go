package server

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/YASSERRMD/Ventiqra/backend/internal/middleware"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
)

// Default starting capital for a new company (in cents): $500,000.
const defaultStartingCashCents int64 = 500_000_00

type companyResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Industry      string    `json:"industry"`
	Description   string    `json:"description"`
	FoundedAt     time.Time `json:"founded_at"`
	CashCents     int64     `json:"cash_cents"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type createCompanyRequest struct {
	Name              string `json:"name"`
	Industry          string `json:"industry"`
	Description       string `json:"description"`
	StartingCashCents *int64 `json:"starting_cash_cents"`
}

func toCompanyResponse(c *repository.Company) companyResponse {
	return companyResponse{
		ID: c.ID, Name: c.Name, Slug: c.Slug, Industry: c.Industry,
		Description: c.Description, FoundedAt: c.FoundedAt, CashCents: c.Cash,
		Status: string(c.Status), CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

func (s *Server) handleCreateCompany(w http.ResponseWriter, r *http.Request) {
	if s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "company service not configured")
		return
	}

	var req createCompanyRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) < 1 || len(req.Name) > 120 {
		writeError(w, http.StatusBadRequest, "name must be 1-120 characters")
		return
	}
	if len(req.Description) > 2000 {
		writeError(w, http.StatusBadRequest, "description must be 2000 characters or fewer")
		return
	}

	cash := defaultStartingCashCents
	if req.StartingCashCents != nil {
		if *req.StartingCashCents < 0 {
			writeError(w, http.StatusBadRequest, "starting_cash_cents must be non-negative")
			return
		}
		cash = *req.StartingCashCents
	}

	ownerID := middleware.UserIDFrom(r.Context())
	company := &repository.Company{
		OwnerID:     ownerID,
		Name:        req.Name,
		Industry:    strings.TrimSpace(req.Industry),
		Description: strings.TrimSpace(req.Description),
		FoundedAt:   time.Now().UTC(),
		Cash:        cash,
	}

	created, err := s.companies.CreateCompany(r.Context(), company)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			writeError(w, http.StatusConflict, "company slug already exists")
			return
		}
		s.log.Error("create company failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not create company")
		return
	}

	// Record the founding milestone on the timeline.
	s.recordTimeline(r.Context(), created.ID, "milestone", "Founded "+created.Name, "The company was founded.", 0)

	writeJSON(w, http.StatusCreated, toCompanyResponse(created))
}

func (s *Server) handleMyCompany(w http.ResponseWriter, r *http.Request) {
	if s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "company service not configured")
		return
	}
	ownerID := middleware.UserIDFrom(r.Context())
	company, err := s.companies.GetLatestCompanyForOwner(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no company found")
			return
		}
		s.log.Error("get my company failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load company")
		return
	}
	writeJSON(w, http.StatusOK, toCompanyResponse(company))
}

func (s *Server) handleGetCompany(w http.ResponseWriter, r *http.Request) {
	if s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "company service not configured")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "company id is required")
		return
	}
	company, err := s.companies.GetCompany(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "company not found")
			return
		}
		s.log.Error("get company failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load company")
		return
	}
	// Only the owner may view their own company for now.
	if company.OwnerID != middleware.UserIDFrom(r.Context()) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	writeJSON(w, http.StatusOK, toCompanyResponse(company))
}

// parseCashParam is reserved for future query-param based endpoints.
func parseCashParam(v string) (int64, bool) {
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}
