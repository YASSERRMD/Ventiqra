package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetMarketDefaults(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/market", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("market: %d body=%s", rec.Code, rec.Body.String())
	}
	var m marketResponse
	if err := json.NewDecoder(rec.Body).Decode(&m); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if m.TAM != 100_000 {
		t.Errorf("TAM = %d, want 100000", m.TAM)
	}
	if m.TrendMultiplier <= 0 || m.TrendMultiplier > 5 {
		t.Errorf("trend = %v out of range", m.TrendMultiplier)
	}
}

func TestMarketGrowsOnTick(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	before := marketTAM(t, srv, token)
	for i := 0; i < 10; i++ {
		tickOnce(t, srv, token)
	}
	after := marketTAM(t, srv, token)
	if after <= before {
		t.Errorf("TAM should grow over ticks: %d -> %d", before, after)
	}
}

func TestMarketRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/market", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func marketTAM(t *testing.T, srv *Server, token string) int64 {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/market", nil, token)
	var m marketResponse
	_ = json.NewDecoder(rec.Body).Decode(&m)
	return m.TAM
}
