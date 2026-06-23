package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

// TestSimulationEngineFullCycle exercises a complete simulation lifecycle:
// create → tick repeatedly → verify day advances, cash changes, and a snapshot
// is recorded for analytics. This is the integration-level engine test.
func TestSimulationEngineFullCycle(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{
		"name": "Engine Co", "starting_cash_cents": 50_000_000_00,
	}, token)

	for i := 0; i < 10; i++ {
		rec := doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token)
		if rec.Code != http.StatusOK {
			t.Fatalf("tick %d: %d body=%s", i, rec.Code, rec.Body.String())
		}
	}

	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/metrics", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics: %d", rec.Code)
	}
	var metrics map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&metrics); err != nil {
		t.Fatalf("decode metrics: %v", err)
	}
	if day, ok := metrics["day"].(float64); !ok || day < 10 {
		t.Errorf("day = %v, want >= 10", metrics["day"])
	}

	rec = doJSON(t, srv, "GET", "/api/v1/companies/me/analytics", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("analytics: %d", rec.Code)
	}
	var analytics analyticsResponse
	if err := json.NewDecoder(rec.Body).Decode(&analytics); err != nil {
		t.Fatalf("decode analytics: %v", err)
	}
	if len(analytics.Cash) != 10 {
		t.Errorf("analytics cash points = %d, want 10", len(analytics.Cash))
	}
}

// TestSimulationEngineBankruptcyPath verifies that a company with near-zero
// cash goes bankrupt and the tick then returns 409.
func TestSimulationEngineBankruptcyPath(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{
		"name": "Doomed Co", "starting_cash_cents": 100,
	}, token)

	bankrupt := false
	for i := 0; i < 2000; i++ {
		rec := doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token)
		if rec.Code == http.StatusConflict {
			bankrupt = true
			break
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("tick %d: %d", i, rec.Code)
		}
	}
	if !bankrupt {
		t.Fatal("company never went bankrupt")
	}

	rec := doJSON(t, srv, "GET", "/api/v1/companies/me", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("company: %d", rec.Code)
	}
	var company companyResponse
	_ = json.NewDecoder(rec.Body).Decode(&company)
	if company.Status != "bankrupt" {
		t.Errorf("status = %s, want bankrupt", company.Status)
	}

	rec = doJSON(t, srv, "GET", "/api/v1/leaderboard", nil, token)
	var lb []leaderboardEntryResponse
	_ = json.NewDecoder(rec.Body).Decode(&lb)
	found := false
	for _, e := range lb {
		if e.CompanyName == "Doomed Co" {
			found = true
		}
	}
	if !found {
		t.Error("bankrupt company not on leaderboard")
	}
}

// TestFundingRaiseFlow verifies the full funding flow: raise → company cash
// increases → timeline records it.
func TestFundingRaiseFlow(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Fund Co"}, token)

	rec := doJSON(t, srv, "GET", "/api/v1/companies/me", nil, token)
	var before companyResponse
	_ = json.NewDecoder(rec.Body).Decode(&before)

	rec = doJSON(t, srv, "POST", "/api/v1/companies/me/funding/raise", map[string]any{
		"amount_cents": 1_000_000_00,
	}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("raise: %d body=%s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, srv, "GET", "/api/v1/companies/me", nil, token)
	var after companyResponse
	_ = json.NewDecoder(rec.Body).Decode(&after)
	if after.CashCents <= before.CashCents {
		t.Errorf("cash did not increase: before=%d after=%d", before.CashCents, after.CashCents)
	}

	rec = doJSON(t, srv, "GET", "/api/v1/companies/me/timeline", nil, token)
	var tl timelineResponse
	_ = json.NewDecoder(rec.Body).Decode(&tl)
	foundFunding := false
	for _, e := range tl.Entries {
		if e.Kind == "funding" {
			foundFunding = true
		}
	}
	if !foundFunding {
		t.Error("funding not recorded in timeline")
	}
}

// TestSimulationEngineDayProgression verifies that N ticks advance the day by N.
func TestSimulationEngineDayProgression(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{
		"name": "Day Co", "starting_cash_cents": 50_000_000_00,
	}, token)

	for i := 0; i < 5; i++ {
		doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token)
	}
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/metrics", nil, token)
	var m map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&m)
	if day, _ := m["day"].(float64); day != 5 {
		t.Errorf("day after 5 ticks = %v, want 5", day)
	}
}
