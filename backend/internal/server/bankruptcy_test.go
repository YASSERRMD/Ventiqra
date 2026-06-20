package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func tickUntil(t *testing.T, srv *Server, token string, maxTicks int, cond func(simTickResponse) bool) (simTickResponse, bool) {
	t.Helper()
	var last simTickResponse
	for i := 0; i < maxTicks; i++ {
		rec := doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token)
		if rec.Code != http.StatusOK {
			return last, false
		}
		_ = json.NewDecoder(rec.Body).Decode(&last)
		if cond(last) {
			return last, true
		}
	}
	return last, false
}

func TestTickReportsHealth(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("tick: %d", rec.Code)
	}
	var res simTickResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if res.Health == "" || res.Status == "" {
		t.Errorf("expected health/status in tick response, got %+v", res)
	}
}

func TestMetricsReportsHealth(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	m := metricsFor(t, srv, token)
	if m.Health == "" {
		t.Errorf("expected health in metrics")
	}
}

func TestBankruptcyAndGameOver(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	// Start with tiny cash so bankruptcy is reachable within a bounded number
	// of ticks.
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co", "starting_cash_cents": 100}, token)

	last, ok := tickUntil(t, srv, token, 2000, func(r simTickResponse) bool {
		return r.Status == "bankrupt"
	})
	if !ok {
		t.Fatalf("did not reach bankruptcy within ticks; last=%+v", last)
	}

	// Game over: further ticks are rejected.
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token)
	if rec.Code != http.StatusConflict {
		t.Errorf("tick after bankruptcy status = %d, want 409", rec.Code)
	}

	// Metrics reflect bankruptcy.
	m := metricsFor(t, srv, token)
	if m.Health != "bankrupt" {
		t.Errorf("metrics health = %q, want bankrupt", m.Health)
	}
}

func TestRestartClearsBankruptcy(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co", "starting_cash_cents": 100}, token)

	if _, ok := tickUntil(t, srv, token, 2000, func(r simTickResponse) bool { return r.Status == "bankrupt" }); !ok {
		t.Fatalf("did not reach bankruptcy")
	}

	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/restart", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("restart: %d body=%s", rec.Code, rec.Body.String())
	}
	var res restartResponse
	_ = json.NewDecoder(rec.Body).Decode(&res)
	if res.Status != "active" || res.Day != 0 {
		t.Errorf("restart result = %+v", res)
	}

	// Ticking works again after restart.
	if rec := doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token); rec.Code != http.StatusOK {
		t.Errorf("tick after restart = %d, want 200", rec.Code)
	}
}

func TestRestartRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/restart", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
