package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestSolicitOffers(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/funding/offers/solicit", nil, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("solicit: %d body=%s", rec.Code, rec.Body.String())
	}
	var offers []offerResponse
	if err := json.NewDecoder(rec.Body).Decode(&offers); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(offers) != 3 {
		t.Fatalf("offers = %d, want 3", len(offers))
	}
	for _, o := range offers {
		if o.InvestorName == "" || o.AmountCents <= 0 || o.EquityPercent <= 0 {
			t.Errorf("malformed offer: %+v", o)
		}
	}
}

func TestListOffers(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	// No offers until solicited.
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/funding/offers", nil, token)
	var empty []offerResponse
	_ = json.NewDecoder(rec.Body).Decode(&empty)
	if len(empty) != 0 {
		t.Fatalf("expected 0 offers initially, got %d", len(empty))
	}

	doJSON(t, srv, "POST", "/api/v1/companies/me/funding/offers/solicit", nil, token)
	rec = doJSON(t, srv, "GET", "/api/v1/companies/me/funding/offers", nil, token)
	var list []offerResponse
	_ = json.NewDecoder(rec.Body).Decode(&list)
	if len(list) != 3 {
		t.Fatalf("expected 3 offers after solicit, got %d", len(list))
	}
}

func TestAcceptOfferClosesRoundAndAddsCash(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	cashBefore := companyCash(t, srv, token)

	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/funding/offers/solicit", nil, token)
	var offers []offerResponse
	_ = json.NewDecoder(rec.Body).Decode(&offers)
	chosen := offers[0]

	rec = doJSON(t, srv, "POST", "/api/v1/companies/me/funding/offers/"+chosen.ID+"/accept", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("accept: %d body=%s", rec.Code, rec.Body.String())
	}

	if cashAfter := companyCash(t, srv, token); cashAfter != cashBefore+chosen.AmountCents {
		t.Errorf("cash = %d, want %d", cashAfter, cashBefore+chosen.AmountCents)
	}

	// Round was recorded.
	s := fundingSummary(t, srv, token)
	if s.RoundsRaised != 1 {
		t.Errorf("rounds = %d, want 1", s.RoundsRaised)
	}

	// Remaining offers were rejected.
	rec = doJSON(t, srv, "GET", "/api/v1/companies/me/funding/offers", nil, token)
	var remaining []offerResponse
	_ = json.NewDecoder(rec.Body).Decode(&remaining)
	if len(remaining) != 0 {
		t.Errorf("pending offers after accept = %d, want 0", len(remaining))
	}
}

func TestRejectOffer(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/funding/offers/solicit", nil, token)
	var offers []offerResponse
	_ = json.NewDecoder(rec.Body).Decode(&offers)

	rec = doJSON(t, srv, "POST", "/api/v1/companies/me/funding/offers/"+offers[0].ID+"/reject", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("reject: %d", rec.Code)
	}
	// One fewer pending offer remains.
	rec = doJSON(t, srv, "GET", "/api/v1/companies/me/funding/offers", nil, token)
	var remaining []offerResponse
	_ = json.NewDecoder(rec.Body).Decode(&remaining)
	if len(remaining) != 2 {
		t.Errorf("pending after reject = %d, want 2", len(remaining))
	}
}

func TestNegotiateOfferEitherImprovesOrWithdraws(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/funding/offers/solicit", nil, token)
	var offers []offerResponse
	_ = json.NewDecoder(rec.Body).Decode(&offers)

	// Negotiate repeatedly across fresh solicit batches until we observe both
	// outcomes. Each offer's outcome is deterministic, so we loop over offers and
	// a few solicit rounds.
	observedWithdraw := false
	observedImprove := false
	for attempt := 0; attempt < 5 && !(observedWithdraw && observedImprove); attempt++ {
		rec = doJSON(t, srv, "POST", "/api/v1/companies/me/funding/offers/solicit", nil, token)
		_ = json.NewDecoder(rec.Body).Decode(&offers)
		for _, o := range offers {
			r := doJSON(t, srv, "POST", "/api/v1/companies/me/funding/offers/"+o.ID+"/negotiate", nil, token)
			if r.Code != http.StatusOK {
				t.Fatalf("negotiate: %d body=%s", r.Code, r.Body.String())
			}
			var res negotiateResultResponse
			_ = json.NewDecoder(r.Body).Decode(&res)
			if res.Withdrawn {
				observedWithdraw = true
			} else if res.Offer != nil && res.Offer.EquityPercent < o.EquityPercent {
				observedImprove = true
			}
		}
	}
	if !observedWithdraw {
		t.Errorf("expected to observe at least one withdrawal")
	}
	if !observedImprove {
		t.Errorf("expected to observe at least one improved offer")
	}
}

func TestOffersRequireAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/funding/offers/solicit", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
