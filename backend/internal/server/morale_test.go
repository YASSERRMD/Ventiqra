package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func moraleSummary(t *testing.T, srv *Server, token string) moraleSummaryResponse {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/morale", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("morale: %d body=%s", rec.Code, rec.Body.String())
	}
	var m moraleSummaryResponse
	if err := json.NewDecoder(rec.Body).Decode(&m); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return m
}

func TestMoraleSummaryEmpty(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	m := moraleSummary(t, srv, token)
	if m.Headcount != 0 {
		t.Errorf("empty headcount = %d, want 0", m.Headcount)
	}
}

func TestMoraleDecaysOnTick(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	// Hire an engineer at morale 80.
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/employees",
		map[string]any{"name": "Dev", "role": "engineer", "skill": 100, "morale": 80}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("hire: %d", rec.Code)
	}

	// Tick a few times; morale should drift down toward the decay floor.
	for i := 0; i < 5; i++ {
		tickOnce(t, srv, token)
	}
	m := moraleSummary(t, srv, token)
	if m.AverageMorale >= 80 {
		t.Errorf("morale should decay from 80: got %d", m.AverageMorale)
	}
}

func TestMoraleBoostOnFunding(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	// Hire an engineer near the decay floor.
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/employees",
		map[string]any{"name": "Dev", "role": "engineer", "skill": 100, "morale": 50}, token)
	_ = rec
	before := moraleSummary(t, srv, token).AverageMorale

	// Raise a round → team morale boost.
	doJSON(t, srv, "POST", "/api/v1/companies/me/funding/raise", map[string]any{"amount_cents": 500_000}, token)
	after := moraleSummary(t, srv, token).AverageMorale
	if after <= before {
		t.Errorf("funding should boost morale: %d -> %d", before, after)
	}
}

func TestBurntOutEmployeesCanResign(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	// Hire several engineers with zero morale (burnt out).
	for i := 0; i < 6; i++ {
		doJSON(t, srv, "POST", "/api/v1/companies/me/employees",
			map[string]any{"name": "Burnt" + string(rune('A'+i)), "role": "engineer", "skill": 50, "morale": 0}, token)
	}

	// Tick enough times for at least one resignation.
	startCount := moraleSummary(t, srv, token).Headcount
	resigned := false
	for i := 0; i < 30; i++ {
		tickOnce(t, srv, token)
		if moraleSummary(t, srv, token).Headcount < startCount {
			resigned = true
			break
		}
	}
	if !resigned {
		t.Errorf("expected at least one resignation among burnt-out employees")
	}
}

func TestMoraleRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/morale", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
