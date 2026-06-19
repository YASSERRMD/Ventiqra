package server

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/YASSERRMD/Ventiqra/backend/internal/sim"
)

func TestMetricsRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/metrics", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestMetricsNoCompany(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)

	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/metrics", nil, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestMetricsInitializesAndReturns(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)

	// Create a company with the default starting cash.
	rec := doJSON(t, srv, "POST", "/api/v1/companies",
		map[string]any{"name": "Metrics Inc"}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create company: %d body=%s", rec.Code, rec.Body.String())
	}
	var c companyResponse
	if err := json.NewDecoder(rec.Body).Decode(&c); err != nil {
		t.Fatalf("decode company: %v", err)
	}

	// Fetching metrics initializes the simulation state lazily (day 0).
	rec = doJSON(t, srv, "GET", "/api/v1/companies/me/metrics", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var m metricsResponse
	if err := json.NewDecoder(rec.Body).Decode(&m); err != nil {
		t.Fatalf("decode metrics: %v", err)
	}

	if m.Day != 0 {
		t.Errorf("day = %d, want 0 (lazy init)", m.Day)
	}
	if m.CashCents != c.CashCents {
		t.Errorf("cash_cents = %d, want company cash %d", m.CashCents, c.CashCents)
	}
	if m.RevenueCents != 0 {
		t.Errorf("revenue_cents = %d, want 0", m.RevenueCents)
	}
	if m.BurnCentsPerMonth != sim.BaseMonthlyBurnCents {
		t.Errorf("burn_cents_per_month = %d, want %d", m.BurnCentsPerMonth, sim.BaseMonthlyBurnCents)
	}
	// No revenue yet, so valuation floors to cash.
	if m.ValuationCents != c.CashCents {
		t.Errorf("valuation_cents = %d, want cash %d", m.ValuationCents, c.CashCents)
	}
	wantRunway := float64(c.CashCents) / float64(sim.BaseMonthlyBurnCents)
	if m.RunwayMonths < wantRunway-1e-9 || m.RunwayMonths > wantRunway+1e-9 {
		t.Errorf("runway_months = %v, want %v", m.RunwayMonths, wantRunway)
	}

	// Advancing the simulation via tick must be reflected on subsequent reads.
	rec = doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("tick status = %d, body = %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, srv, "GET", "/api/v1/companies/me/metrics", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics after tick status = %d", rec.Code)
	}
	var m2 metricsResponse
	if err := json.NewDecoder(rec.Body).Decode(&m2); err != nil {
		t.Fatalf("decode metrics after tick: %v", err)
	}
	if m2.Day != 1 {
		t.Errorf("day after tick = %d, want 1", m2.Day)
	}
	if m2.CashCents == c.CashCents {
		t.Errorf("cash unchanged after tick; expected a deterministic delta")
	}
}
