package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

// tickOnce performs a single simulation tick for the owner's company.
func tickOnce(t *testing.T, srv *Server, token string) {
	t.Helper()
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("sim tick status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestPayrollRaisesBurn(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	// Baseline burn after a tick with no employees reflects base overhead + infra
	// floor (no payroll, no marketing).
	tickOnce(t, srv, token)
	before := metricsFor(t, srv, token)
	if before.BurnCentsPerMonth != 550_000 {
		t.Fatalf("baseline burn = %d, want 550000", before.BurnCentsPerMonth)
	}

	// Hire an engineer with a known monthly salary.
	hire := doJSON(t, srv, "POST", "/api/v1/companies/me/employees",
		map[string]any{"name": "Builder", "role": "engineer", "salary_cents": 1200000, "skill": 100, "morale": 70}, token)
	if hire.Code != http.StatusCreated {
		t.Fatalf("hire: %d body=%s", hire.Code, hire.Body.String())
	}

	tickOnce(t, srv, token)
	after := metricsFor(t, srv, token)
	if after.BurnCentsPerMonth != 550_000+1_200_000 {
		t.Errorf("burn with payroll = %d, want %d", after.BurnCentsPerMonth, 550_000+1_200_000)
	}
}

func TestBuildingProductsAdvanceWithDevelopers(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	// Create a product and move it into the building stage.
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/products", map[string]any{"name": "Dev App"}, token)
	var p productResponse
	_ = json.NewDecoder(rec.Body).Decode(&p)
	if code := doJSON(t, srv, "PATCH", "/api/v1/products/"+p.ID+"/stage", map[string]any{"stage": "building"}, token).Code; code != http.StatusOK {
		t.Fatalf("stage to building: %d", code)
	}

	// Without developers, a tick should not advance progress.
	tickOnce(t, srv, token)
	if got := productProgress(t, srv, token, p.ID); got != 0 {
		t.Errorf("progress without developers = %v, want 0", got)
	}

	// Hire a strong, motivated engineer (skill 100, morale 70 → 0.5%/day).
	hire := doJSON(t, srv, "POST", "/api/v1/companies/me/employees",
		map[string]any{"name": "Dev", "role": "engineer", "skill": 100, "morale": 70}, token)
	if hire.Code != http.StatusCreated {
		t.Fatalf("hire: %d body=%s", hire.Code, hire.Body.String())
	}

	tickOnce(t, srv, token)
	progress := productProgress(t, srv, token, p.ID)
	if progress < 0.4 || progress > 0.6 {
		t.Errorf("progress after one dev-day = %v, want ~0.5", progress)
	}

	// A second tick accumulates further.
	tickOnce(t, srv, token)
	if got := productProgress(t, srv, token, p.ID); got < progress {
		t.Errorf("progress should accumulate: %v then %v", progress, got)
	}
}

func TestNonBuildingProductsDoNotAdvance(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	// Idea-stage product stays at 0 even with developers.
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/products", map[string]any{"name": "Idea App"}, token)
	var p productResponse
	_ = json.NewDecoder(rec.Body).Decode(&p)
	doJSON(t, srv, "POST", "/api/v1/companies/me/employees",
		map[string]any{"name": "Dev", "role": "engineer", "skill": 100, "morale": 70}, token)
	tickOnce(t, srv, token)
	if got := productProgress(t, srv, token, p.ID); got != 0 {
		t.Errorf("idea product advanced: %v, want 0", got)
	}
}

func metricsFor(t *testing.T, srv *Server, token string) metricsResponse {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/metrics", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics: %d body=%s", rec.Code, rec.Body.String())
	}
	var m metricsResponse
	if err := json.NewDecoder(rec.Body).Decode(&m); err != nil {
		t.Fatalf("decode metrics: %v", err)
	}
	return m
}

func productProgress(t *testing.T, srv *Server, token, id string) float64 {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/products", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list products: %d", rec.Code)
	}
	var list []productResponse
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode products: %v", err)
	}
	for _, p := range list {
		if p.ID == id {
			return p.DevProgress
		}
	}
	t.Fatalf("product %s not found", id)
	return 0
}
