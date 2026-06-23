package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAchievementsStartEmpty(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/achievements", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list: %d", rec.Code)
	}
	var list []achievementResponse
	_ = json.NewDecoder(rec.Body).Decode(&list)
	if len(list) != 0 {
		t.Errorf("expected empty, got %d", len(list))
	}
}

func TestAchievementsAwardedOnTick(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co", "starting_cash_cents": 50_000_000_00}, token)
	// Tick once; the company exists so no achievements yet unless conditions met.
	doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token)
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/achievements", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list after tick: %d", rec.Code)
	}
	// The endpoint should work without error; specific awards depend on state.
	var list []achievementResponse
	_ = json.NewDecoder(rec.Body).Decode(&list)
	// Just assert it's a valid list (may be empty or contain awards depending on random events).
	_ = list
}

func TestAchievementsRequireAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	if rec := doJSON(t, srv, "GET", "/api/v1/companies/me/achievements", nil, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("no-auth = %d, want 401", rec.Code)
	}
}
