package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestLeaderboardStartsEmpty(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	rec := doJSON(t, srv, "GET", "/api/v1/leaderboard", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("leaderboard: %d", rec.Code)
	}
	var list []leaderboardEntryResponse
	_ = json.NewDecoder(rec.Body).Decode(&list)
	if len(list) != 0 {
		t.Errorf("expected empty, got %d", len(list))
	}
}

func TestBankruptcyRecordsLeaderboard(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	// Near-zero cash → bankrupts within a few ticks.
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Flash Co", "starting_cash_cents": 100}, token)
	// Tick until the company is bankrupt (tick returns 409 once bankrupt).
	for i := 0; i < 2000; i++ {
		rec := doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token)
		if rec.Code == http.StatusConflict {
			break
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("tick %d: %d", i, rec.Code)
		}
	}
	// The leaderboard should now have an entry.
	rec := doJSON(t, srv, "GET", "/api/v1/leaderboard", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("leaderboard after bankrupt: %d", rec.Code)
	}
	var list []leaderboardEntryResponse
	_ = json.NewDecoder(rec.Body).Decode(&list)
	found := false
	for _, e := range list {
		if e.CompanyName == "Flash Co" {
			found = true
			if e.Outcome != "bankrupt" {
				t.Errorf("outcome = %s, want bankrupt", e.Outcome)
			}
		}
	}
	if !found {
		t.Error("bankrupt company not on leaderboard")
	}
}

func TestLeaderboardRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	if rec := doJSON(t, srv, "GET", "/api/v1/leaderboard", nil, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("no-auth = %d, want 401", rec.Code)
	}
}
