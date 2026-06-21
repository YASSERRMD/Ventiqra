package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func marketingFor(t *testing.T, srv *Server, token string) marketingResponse {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/marketing", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("marketing: %d body=%s", rec.Code, rec.Body.String())
	}
	var m marketingResponse
	if err := json.NewDecoder(rec.Body).Decode(&m); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return m
}

func TestGetMarketingZeroBudget(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	m := marketingFor(t, srv, token)
	if m.MonthlyBudgetCents != 0 || m.DailyConversions != 0 {
		t.Errorf("zero budget should yield no conversions: %+v", m)
	}
	if len(m.Channels) != 4 {
		t.Errorf("channels = %d, want 4", len(m.Channels))
	}
}

func TestMarketingSpendProducesConversions(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	// Set a marketing budget.
	doJSON(t, srv, "PATCH", "/api/v1/companies/me/finance", map[string]any{"marketing_budget_cents": 1_000_000}, token)

	m := marketingFor(t, srv, token)
	if m.DailyConversions <= 0 {
		t.Errorf("expected conversions from budget: %+v", m)
	}
	if m.ConversionRate <= 0 {
		t.Errorf("expected positive conversion rate: %v", m.ConversionRate)
	}
	// CAC computed on monthly conversions should be positive.
	monthly := m.DailyConversions * 30
	if m.CACCents != 1_000_000/int64(monthly) {
		t.Errorf("CAC = %d, want %d", m.CACCents, 1_000_000/int64(monthly))
	}
}

func TestMarketingBoostsCustomersOnTick(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	p := launchReadyProduct(t, srv, token, "Marketed")
	doJSON(t, srv, "PATCH", "/api/v1/products/"+p.ID+"/price", map[string]any{"price_cents": 1000}, token)

	// Baseline customers with no marketing.
	for i := 0; i < 3; i++ {
		tickOnce(t, srv, token)
	}
	baseline := customerTotal(t, srv, token, p.ID)

	// Restart-style: set a big budget and tick; customers should grow faster.
	doJSON(t, srv, "PATCH", "/api/v1/companies/me/finance", map[string]any{"marketing_budget_cents": 5_000_000}, token)
	for i := 0; i < 3; i++ {
		tickOnce(t, srv, token)
	}
	withMarketing := customerTotal(t, srv, token, p.ID)
	if withMarketing <= baseline {
		t.Errorf("marketing should boost customers: baseline=%d with=%d", baseline, withMarketing)
	}
}

func TestMarketingRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/marketing", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
