package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func analyticsFor(t *testing.T, srv *Server, token string) analyticsResponse {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/analytics", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("analytics: %d body=%s", rec.Code, rec.Body.String())
	}
	var a analyticsResponse
	if err := json.NewDecoder(rec.Body).Decode(&a); err != nil {
		t.Fatalf("decode analytics: %v", err)
	}
	return a
}

func TestAnalyticsEmptyForNewCompany(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	a := analyticsFor(t, srv, token)
	if len(a.Cash) != 0 {
		t.Errorf("expected empty cash series, got %d points", len(a.Cash))
	}
}

func TestAnalyticsAccumulatesAfterTicks(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co", "starting_cash_cents": 50_000_000_00}, token)
	for i := 0; i < 5; i++ {
		if rec := doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token); rec.Code != http.StatusOK {
			t.Fatalf("tick %d: %d", i, rec.Code)
		}
	}
	a := analyticsFor(t, srv, token)
	// One snapshot per tick (5 days → days 1..5).
	if len(a.Cash) != 5 {
		t.Errorf("cash series len = %d, want 5", len(a.Cash))
	}
	if len(a.Burn) != 5 || len(a.Valuation) != 5 {
		t.Errorf("series len burn=%d valuation=%d, want 5", len(a.Burn), len(a.Valuation))
	}
	// Days are ascending.
	if a.Cash[0].Day != 1 || a.Cash[4].Day != 5 {
		t.Errorf("day order wrong: first=%d last=%d", a.Cash[0].Day, a.Cash[4].Day)
	}
}

func TestAnalyticsSnapshotRecordsValuation(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co", "starting_cash_cents": 10_000_000_00}, token)
	doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token)
	a := analyticsFor(t, srv, token)
	if len(a.Valuation) == 0 {
		t.Fatal("no valuation points")
	}
	if a.Valuation[0].ValuationCents <= 0 {
		t.Errorf("valuation = %d, want > 0", a.Valuation[0].ValuationCents)
	}
}

func TestAnalyticsRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/analytics", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("no-auth = %d, want 401", rec.Code)
	}
}
