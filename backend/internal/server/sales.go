// Sales handlers: create deals, list the pipeline, and advance deals one stage
// (closing at negotiation). Won deals pay their value into cash.
package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
	"github.com/YASSERRMD/Ventiqra/backend/internal/sales"
)

type dealResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Stage       string `json:"stage"`
	ValueCents  int64  `json:"value_cents"`
	Probability int    `json:"probability"`
	ClosedWon   bool   `json:"closed_won"`
	CreatedDay  int    `json:"created_day"`
	ClosedDay   *int   `json:"closed_day"`
}

type dealInput struct {
	Name       *string `json:"name"`
	ValueCents *int64  `json:"value_cents"`
}

type advanceResponse struct {
	Deal    dealResponse `json:"deal"`
	PaidOut int64        `json:"paid_out_cents"`
}

func toDealResponse(d *repository.Deal) dealResponse {
	return dealResponse{
		ID: d.ID, Name: d.Name, Stage: d.Stage, ValueCents: d.ValueCents,
		Probability: d.Probability, ClosedWon: d.ClosedWon, CreatedDay: d.CreatedDay, ClosedDay: d.ClosedDay,
	}
}

func (s *Server) handleListDeals(w http.ResponseWriter, r *http.Request) {
	if s.deals == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "sales service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	list, err := s.deals.ListByCompany(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load deals")
		return
	}
	out := make([]dealResponse, 0, len(list))
	for _, d := range list {
		out = append(out, toDealResponse(d))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleCreateDeal(w http.ResponseWriter, r *http.Request) {
	if s.deals == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "sales service not configured")
		return
	}
	var req dealInput
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Name == nil || *req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	value := int64(50_000_00)
	if req.ValueCents != nil {
		value = *req.ValueCents
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	day := s.currentSimDay(r.Context(), companyID)
	created, err := s.deals.Create(r.Context(), &repository.Deal{
		CompanyID: companyID, Name: *req.Name, Stage: string(sales.StageLead),
		ValueCents: value, Probability: sales.StageProbability[sales.StageLead], CreatedDay: day,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create deal")
		return
	}
	writeJSON(w, http.StatusCreated, toDealResponse(created))
}

func (s *Server) handleAdvanceDeal(w http.ResponseWriter, r *http.Request) {
	if s.deals == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "sales service not configured")
		return
	}
	id := r.PathValue("id")
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	d, err := s.deals.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "deal not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not load deal")
		return
	}
	if d.CompanyID != companyID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	if d.Stage == string(sales.StageClosedWon) || d.Stage == string(sales.StageClosedLost) {
		writeError(w, http.StatusConflict, "deal already closed")
		return
	}

	prob := d.Probability + sales.AdvanceChanceBoost(s.salesAgentCount(r.Context(), companyID))
	if prob > 95 {
		prob = 95
	}

	seed := s.companySeed(r.Context(), companyID)
	newStage := sales.Advance(sales.Stage(d.Stage), prob, seed^int64(d.CreatedDay))
	closedWon := newStage == sales.StageClosedWon
	var closedDay *int
	if closedWon || newStage == sales.StageClosedLost {
		day := s.currentSimDay(r.Context(), companyID)
		closedDay = &day
	}
	_ = s.deals.UpdateStage(r.Context(), id, string(newStage), prob, closedWon, closedDay)

	var paidOut int64
	if closedWon {
		paidOut = d.ValueCents
		company, err := s.companies.GetCompany(r.Context(), companyID)
		if err == nil {
			_ = s.companies.UpdateCash(r.Context(), companyID, company.Cash+paidOut)
		}
		if closedDay != nil {
			s.recordTimeline(r.Context(), companyID, "milestone", "Won deal: "+d.Name,
				"Value "+formatInt(int(paidOut/100))+" dollars", *closedDay)
		}
	}

	updated, _ := s.deals.Get(r.Context(), id)
	writeJSON(w, http.StatusOK, advanceResponse{Deal: toDealResponse(updated), PaidOut: paidOut})
}

func (s *Server) salesAgentCount(ctx context.Context, companyID string) int {
	if s.employees == nil {
		return 0
	}
	emps, err := s.employees.ListEmployeesByCompany(ctx, companyID)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range emps {
		if e.Role == repository.RoleSales {
			n++
		}
	}
	return n
}
