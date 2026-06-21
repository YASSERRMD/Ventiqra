package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func fundingSummary(t *testing.T, srv *Server, token string) fundingSummaryResponse {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/funding", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("funding summary: %d body=%s", rec.Code, rec.Body.String())
	}
	var s fundingSummaryResponse
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return s
}

func TestFundingSummaryDefaults(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	s := fundingSummary(t, srv, token)
	if s.RoundsRaised != 0 {
		t.Errorf("rounds = %d, want 0", s.RoundsRaised)
	}
	if s.FounderEquity != 100 {
		t.Errorf("founder equity = %v, want 100", s.FounderEquity)
	}
	if s.PreMoneyCents <= 0 {
		t.Errorf("pre-money should be positive")
	}
}

func TestRaiseFundingAddsCashAndDilutes(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	before := companyCash(t, srv, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/funding/raise", map[string]any{"amount_cents": 1_000_000}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("raise: %d body=%s", rec.Code, rec.Body.String())
	}

	after := companyCash(t, srv, token)
	if after != before+1_000_000 {
		t.Errorf("cash after raise = %d, want %d", after, before+1_000_000)
	}

	s := fundingSummary(t, srv, token)
	if s.RoundsRaised != 1 {
		t.Errorf("rounds = %d, want 1", s.RoundsRaised)
	}
	if s.FounderEquity >= 100 {
		t.Errorf("founder equity = %v, should be diluted below 100", s.FounderEquity)
	}
	if len(s.Rounds) != 1 || s.Rounds[0].RoundName != "pre-seed" {
		t.Errorf("round = %+v", s.Rounds)
	}
}

func TestRaiseFundingValidation(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	cases := []map[string]any{
		{},
		{"amount_cents": 0},
		{"amount_cents": -5},
	}
	for _, body := range cases {
		rec := doJSON(t, srv, "POST", "/api/v1/companies/me/funding/raise", body, token)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	}
}

func TestRaiseFundingProgressesRoundNames(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	for i := range []string{"pre-seed", "seed", "series-a"} {
		rec := doJSON(t, srv, "POST", "/api/v1/companies/me/funding/raise", map[string]any{"amount_cents": 500_000}, token)
		if rec.Code != http.StatusCreated {
			t.Fatalf("raise %d: %d", i, rec.Code)
		}
	}
	s := fundingSummary(t, srv, token)
	// Rounds newest-first: series-a, seed, pre-seed.
	if len(s.Rounds) != 3 || s.Rounds[0].RoundName != "series-a" || s.Rounds[2].RoundName != "pre-seed" {
		t.Errorf("round order = %+v", s.Rounds)
	}
}

func TestRaiseFundingRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/funding/raise", map[string]any{"amount_cents": 100}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func companyCash(t *testing.T, srv *Server, token string) int64 {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me", nil, token)
	var c companyResponse
	_ = json.NewDecoder(rec.Body).Decode(&c)
	return c.CashCents
}
