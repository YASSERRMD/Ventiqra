package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func reputationFor(t *testing.T, srv *Server, token string) reputationResponse {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/reputation", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("reputation: %d body=%s", rec.Code, rec.Body.String())
	}
	var r reputationResponse
	if err := json.NewDecoder(rec.Body).Decode(&r); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return r
}

func TestReputationStartsNeutral(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	r := reputationFor(t, srv, token)
	if r.Score != 50 {
		t.Errorf("initial score = %d, want 50", r.Score)
	}
	if r.Growth < 0.99 || r.Growth > 1.01 {
		t.Errorf("neutral growth = %v, want ~1.0", r.Growth)
	}
}

func TestLaunchBoostsReputation(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	launchReadyProduct(t, srv, token, "Big Launch")

	r := reputationFor(t, srv, token)
	if r.Score <= 50 {
		t.Errorf("launch should boost reputation: %d", r.Score)
	}
	// An event for the launch was recorded.
	found := false
	for _, e := range r.Events {
		if e.Delta > 0 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a positive reputation event after launch")
	}
}

func TestFundingBoostsReputation(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/funding/raise", map[string]any{"amount_cents": 1_000_000}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("raise: %d", rec.Code)
	}
	r := reputationFor(t, srv, token)
	if r.Score <= 50 {
		t.Errorf("funding should boost reputation: %d", r.Score)
	}
}

func TestReputationRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/reputation", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
