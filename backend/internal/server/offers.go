package server

import (
	"errors"
	"hash/fnv"
	"net/http"
	"time"

	"github.com/YASSERRMD/Ventiqra/backend/internal/funding"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
	"github.com/YASSERRMD/Ventiqra/backend/internal/sim"
)

type offerResponse struct {
	ID            string    `json:"id"`
	InvestorName  string    `json:"investor_name"`
	AmountCents   int64     `json:"amount_cents"`
	EquityPercent float64   `json:"equity_percent"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type negotiateResultResponse struct {
	Withdrawn bool           `json:"withdrawn"`
	Message   string         `json:"message"`
	Offer     *offerResponse `json:"offer,omitempty"`
}

func toOfferResponse(o *repository.InvestorOffer) offerResponse {
	return offerResponse{
		ID: o.ID, InvestorName: o.InvestorName, AmountCents: o.AmountCents,
		EquityPercent: o.EquityPercent, Status: string(o.Status), CreatedAt: o.CreatedAt,
	}
}

// companyRoundKey returns the deterministic (seed, day) pair for offer generation.
func (s *Server) companyRoundKey(r *http.Request, companyID string) (int64, int) {
	seed := sim.SeedFromCompanyID(companyID)
	day := 0
	if s.sim != nil {
		if state, err := s.sim.Get(r.Context(), companyID); err == nil {
			seed = state.Seed
			day = state.Day
		}
	}
	return seed, day
}

// handleSolicitOffers generates a fresh deterministic batch of investor offers,
// replacing any pending offers.
func (s *Server) handleSolicitOffers(w http.ResponseWriter, r *http.Request) {
	if s.offers == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "offer service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	company, err := s.companies.GetCompany(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusNotFound, "company not found")
		return
	}
	if company.Status == repository.CompanyBankrupt {
		writeError(w, http.StatusConflict, "bankrupt companies cannot solicit funding")
		return
	}

	seed, day := s.companyRoundKey(r, companyID)
	preMoney := funding.PreMoneyValuation(company.Cash, s.companyMonthlyRevenue(r.Context(), companyID))

	if err := s.offers.ClearPending(r.Context(), companyID); err != nil {
		s.log.Error("clear pending offers failed", "error", err)
	}
	offers := funding.GenerateOffers(seed, int64(day), preMoney)
	out := make([]offerResponse, 0, len(offers))
	for _, of := range offers {
		inserted, err := s.offers.Insert(r.Context(), &repository.InvestorOffer{
			CompanyID:     companyID,
			InvestorName:  of.InvestorName,
			AmountCents:   of.AmountCents,
			EquityPercent: of.EquityPercent,
			RoundSeed:     seed,
		})
		if err != nil {
			s.log.Error("insert offer failed", "error", err)
			continue
		}
		out = append(out, toOfferResponse(inserted))
	}
	writeJSON(w, http.StatusCreated, out)
}

// handleListOffers returns the company's pending offers.
func (s *Server) handleListOffers(w http.ResponseWriter, r *http.Request) {
	if s.offers == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "offer service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	list, err := s.offers.ListPendingByCompany(r.Context(), companyID)
	if err != nil {
		s.log.Error("list offers failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load offers")
		return
	}
	out := make([]offerResponse, 0, len(list))
	for _, o := range list {
		out = append(out, toOfferResponse(o))
	}
	writeJSON(w, http.StatusOK, out)
}

// loadOwnedOffer loads an offer and confirms it belongs to the owner's company.
func (s *Server) loadOwnedOffer(w http.ResponseWriter, r *http.Request) (*repository.InvestorOffer, bool) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "offer id is required")
		return nil, false
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return nil, false
	}
	offer, err := s.offers.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "offer not found")
			return nil, false
		}
		s.log.Error("load offer failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load offer")
		return nil, false
	}
	if offer.CompanyID != companyID {
		writeError(w, http.StatusForbidden, "forbidden")
		return nil, false
	}
	return offer, true
}

// handleAcceptOffer accepts an offer: adds cash, records the funding round, and
// rejects all other pending offers.
func (s *Server) handleAcceptOffer(w http.ResponseWriter, r *http.Request) {
	if s.offers == nil || s.companies == nil || s.funding == nil {
		writeError(w, http.StatusServiceUnavailable, "offer service not configured")
		return
	}
	offer, ok := s.loadOwnedOffer(w, r)
	if !ok {
		return
	}
	if offer.Status != repository.OfferPending {
		writeError(w, http.StatusConflict, "offer is no longer pending")
		return
	}

	company, err := s.companies.GetCompany(r.Context(), offer.CompanyID)
	if err != nil {
		writeError(w, http.StatusNotFound, "company not found")
		return
	}
	preMoney := funding.PreMoneyValuation(company.Cash, s.companyMonthlyRevenue(r.Context(), company.ID))
	prior, _ := s.funding.CountByCompany(r.Context(), company.ID)
	day := s.currentSimDay(r.Context(), company.ID)

	if _, err := s.funding.Record(r.Context(), &repository.FundingRound{
		CompanyID:     company.ID,
		RoundName:     funding.NextRoundName(prior),
		AmountCents:   offer.AmountCents,
		PreMoneyCents: preMoney,
		EquityPercent: offer.EquityPercent,
		SimDay:        day,
	}); err != nil {
		s.log.Error("record accepted round failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not close round")
		return
	}

	newCash := company.Cash + offer.AmountCents
	if err := s.companies.UpdateCash(r.Context(), company.ID, newCash); err != nil {
		s.log.Error("apply offer cash failed", "error", err)
	}
	_ = s.offers.UpdateStatus(r.Context(), offer.ID, repository.OfferAccepted)

	// Reject remaining pending offers (only one round closes at a time).
	if pending, err := s.offers.ListPendingByCompany(r.Context(), company.ID); err == nil {
		for _, p := range pending {
			_ = s.offers.UpdateStatus(r.Context(), p.ID, repository.OfferRejected)
		}
	}

	writeJSON(w, http.StatusOK, toOfferResponse(&repository.InvestorOffer{
		ID: offer.ID, CompanyID: offer.CompanyID, InvestorName: offer.InvestorName,
		AmountCents: offer.AmountCents, EquityPercent: offer.EquityPercent,
		Status: repository.OfferAccepted, RoundSeed: offer.RoundSeed,
		CreatedAt: offer.CreatedAt, UpdatedAt: time.Now(),
	}))
}

// handleRejectOffer rejects a pending offer.
func (s *Server) handleRejectOffer(w http.ResponseWriter, r *http.Request) {
	if s.offers == nil {
		writeError(w, http.StatusServiceUnavailable, "offer service not configured")
		return
	}
	offer, ok := s.loadOwnedOffer(w, r)
	if !ok {
		return
	}
	if offer.Status != repository.OfferPending {
		writeError(w, http.StatusConflict, "offer is no longer pending")
		return
	}
	if err := s.offers.UpdateStatus(r.Context(), offer.ID, repository.OfferRejected); err != nil {
		s.log.Error("reject offer failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not reject offer")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": offer.ID, "status": "rejected"})
}

// handleNegotiateOffer attempts to improve an offer's terms; on failure the
// investor withdraws.
func (s *Server) handleNegotiateOffer(w http.ResponseWriter, r *http.Request) {
	if s.offers == nil {
		writeError(w, http.StatusServiceUnavailable, "offer service not configured")
		return
	}
	offer, ok := s.loadOwnedOffer(w, r)
	if !ok {
		return
	}
	if offer.Status != repository.OfferPending {
		writeError(w, http.StatusConflict, "offer is no longer pending")
		return
	}

	_, day := s.companyRoundKey(r, offer.CompanyID)
	newEquity, withdrawn := funding.NegotiateOutcome(offer.RoundSeed, int64(day), offerIndexFromID(offer), offer.EquityPercent)
	if withdrawn {
		_ = s.offers.UpdateStatus(r.Context(), offer.ID, repository.OfferWithdrawn)
		writeJSON(w, http.StatusOK, negotiateResultResponse{
			Withdrawn: true,
			Message:   offer.InvestorName + " withdrew their offer during negotiation.",
		})
		return
	}
	_ = s.offers.UpdateEquity(r.Context(), offer.ID, newEquity)
	updated, _ := s.offers.Get(r.Context(), offer.ID)
	resp := toOfferResponse(updated)
	writeJSON(w, http.StatusOK, negotiateResultResponse{
		Withdrawn: false,
		Message:   offer.InvestorName + " conceded to better terms!",
		Offer:     &resp,
	})
}

// offerIndexFromID derives a stable index in [0, OfferCount) from the offer id,
// so each offer's negotiation outcome is independently yet reproducibly decided.
func offerIndexFromID(o *repository.InvestorOffer) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(o.ID))
	return int(h.Sum32() % uint32(funding.OfferCount))
}
