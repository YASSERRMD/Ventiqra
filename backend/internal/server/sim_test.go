package server

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/YASSERRMD/Ventiqra/backend/internal/sim"
)

func TestSimTickRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestSimTickNoCompany(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)

	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestSimTickInitializesAndAdvances(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)

	// Create a company first.
	rec := doJSON(t, srv, "POST", "/api/v1/companies",
		map[string]any{"name": "Sim Inc"}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create company: %d body=%s", rec.Code, rec.Body.String())
	}
	var c companyResponse
	if err := json.NewDecoder(rec.Body).Decode(&c); err != nil {
		t.Fatalf("decode company: %v", err)
	}
	startCash := c.CashCents

	// First tick initializes state at day 0 then advances to day 1.
	rec = doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("tick status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var tick1 simTickResponse
	if err := json.NewDecoder(rec.Body).Decode(&tick1); err != nil {
		t.Fatalf("decode tick: %v", err)
	}
	if tick1.Day != 1 {
		t.Errorf("day = %d, want 1", tick1.Day)
	}
	if tick1.CompanyID != c.ID {
		t.Errorf("company_id = %s, want %s", tick1.CompanyID, c.ID)
	}
	if tick1.Seed == 0 {
		t.Errorf("seed should be derived and non-zero")
	}
	delta := tick1.CashCents - startCash
	if delta < -100_000 || delta > 50_000 {
		t.Errorf("first-tick delta %d out of range", delta)
	}

	// The company's cash must mirror the simulated cash.
	rec = doJSON(t, srv, "GET", "/api/v1/companies/me", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("me status = %d", rec.Code)
	}
	var me companyResponse
	_ = json.NewDecoder(rec.Body).Decode(&me)
	if me.CashCents != tick1.CashCents {
		t.Errorf("company cash %d != sim cash %d after tick", me.CashCents, tick1.CashCents)
	}

	// Second tick advances to day 2 deterministically.
	rec = doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("tick2 status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var tick2 simTickResponse
	if err := json.NewDecoder(rec.Body).Decode(&tick2); err != nil {
		t.Fatalf("decode tick2: %v", err)
	}
	if tick2.Day != 2 {
		t.Errorf("day = %d, want 2", tick2.Day)
	}
	if tick2.Seed != tick1.Seed {
		t.Errorf("seed changed between ticks: %d != %d", tick2.Seed, tick1.Seed)
	}
}

func TestSimTickSeedIsDeterministicallyDerived(t *testing.T) {
	// The persisted seed must equal the deterministic derivation from the
	// company id, so the same company always yields the same stream regardless
	// of how many times its state is reloaded.
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	rec := doJSON(t, srv, "POST", "/api/v1/companies",
		map[string]any{"name": "Seed Co"}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d", rec.Code)
	}
	var c companyResponse
	if err := json.NewDecoder(rec.Body).Decode(&c); err != nil {
		t.Fatalf("decode: %v", err)
	}

	wantSeed := sim.SeedFromCompanyID(c.ID)

	rec = doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("tick: %d", rec.Code)
	}
	var tick simTickResponse
	if err := json.NewDecoder(rec.Body).Decode(&tick); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if tick.Seed != wantSeed {
		t.Errorf("seed = %d, want derived %d", tick.Seed, wantSeed)
	}
}
