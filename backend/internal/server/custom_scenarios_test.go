package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func customScenarioList(t *testing.T, srv *Server, token string) []customScenarioResponse {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/scenarios/custom", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list custom: %d body=%s", rec.Code, rec.Body.String())
	}
	var list []customScenarioResponse
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return list
}

func validCustomBody() map[string]any {
	return map[string]any{
		"name": "My Custom", "difficulty": "hard", "industry": "Fintech",
		"starting_cash_cents": 2_000_000_00, "starting_burn_cents": 150_000_00,
		"market_tam": 100_000, "market_growth_rate": 0.1, "market_trend": 1.1,
	}
}

func TestCustomScenariosStartEmpty(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	list := customScenarioList(t, srv, token)
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d", len(list))
	}
}

func TestCreateCustomScenarioPersists(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	rec := doJSON(t, srv, "POST", "/api/v1/scenarios/custom", validCustomBody(), token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d body=%s", rec.Code, rec.Body.String())
	}
	var c customScenarioResponse
	if err := json.NewDecoder(rec.Body).Decode(&c); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if c.Name != "My Custom" || c.Difficulty != "hard" || c.Industry != "Fintech" {
		t.Errorf("unexpected scenario: %+v", c)
	}
	if c.Market.TAM != 100_000 {
		t.Errorf("TAM = %d, want 100000", c.Market.TAM)
	}
	// It now appears in the list.
	list := customScenarioList(t, srv, token)
	if len(list) != 1 {
		t.Fatalf("expected 1 saved scenario, got %d", len(list))
	}
}

func TestCreateCustomScenarioRejectsInvalid(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	cases := []struct {
		name string
		body map[string]any
	}{
		{"empty name", map[string]any{"name": "", "starting_cash_cents": 100_000_00, "starting_burn_cents": 10_000_00, "market_tam": 5000, "market_growth_rate": 0.05, "market_trend": 1.0}},
		{"cash too low", map[string]any{"name": "X", "starting_cash_cents": 100, "starting_burn_cents": 10_000_00, "market_tam": 5000, "market_growth_rate": 0.05, "market_trend": 1.0}},
		{"tam too low", map[string]any{"name": "X", "starting_cash_cents": 100_000_00, "starting_burn_cents": 10_000_00, "market_tam": 10, "market_growth_rate": 0.05, "market_trend": 1.0}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := doJSON(t, srv, "POST", "/api/v1/scenarios/custom", c.body, token)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("%s: status = %d, want 400", c.name, rec.Code)
			}
		})
	}
}

func TestUpdateCustomScenario(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	rec := doJSON(t, srv, "POST", "/api/v1/scenarios/custom", validCustomBody(), token)
	var c customScenarioResponse
	_ = json.NewDecoder(rec.Body).Decode(&c)

	body := validCustomBody()
	body["name"] = "Renamed"
	body["difficulty"] = "brutal"
	rec = doJSON(t, srv, "PATCH", "/api/v1/scenarios/custom/"+c.ID, body, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("update: %d body=%s", rec.Code, rec.Body.String())
	}
	var updated customScenarioResponse
	_ = json.NewDecoder(rec.Body).Decode(&updated)
	if updated.Name != "Renamed" || updated.Difficulty != "brutal" {
		t.Errorf("update did not apply: %+v", updated)
	}
}

func TestDeleteCustomScenario(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	rec := doJSON(t, srv, "POST", "/api/v1/scenarios/custom", validCustomBody(), token)
	var c customScenarioResponse
	_ = json.NewDecoder(rec.Body).Decode(&c)

	rec = doJSON(t, srv, "DELETE", "/api/v1/scenarios/custom/"+c.ID, nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete: %d", rec.Code)
	}
	list := customScenarioList(t, srv, token)
	if len(list) != 0 {
		t.Errorf("expected empty after delete, got %d", len(list))
	}
}

func TestApplyCustomScenarioSetsCompany(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/scenarios/custom", validCustomBody(), token)
	var c customScenarioResponse
	_ = json.NewDecoder(rec.Body).Decode(&c)

	rec = doJSON(t, srv, "POST", "/api/v1/scenarios/custom/"+c.ID+"/apply", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("apply: %d body=%s", rec.Code, rec.Body.String())
	}
	var res applyScenarioResponse
	_ = json.NewDecoder(rec.Body).Decode(&res)
	if res.Company.Industry != "Fintech" {
		t.Errorf("industry = %q, want Fintech", res.Company.Industry)
	}
	if res.Company.CashCents != 2_000_000_00 {
		t.Errorf("cash = %d, want 200000000", res.Company.CashCents)
	}
}

func TestCustomScenarioOwnerIsolation(t *testing.T) {
	srv := newAuthTestServer(t)
	tokenA := registerAndLogin(t, srv)
	rec := doJSON(t, srv, "POST", "/api/v1/scenarios/custom", validCustomBody(), tokenA)
	var c customScenarioResponse
	_ = json.NewDecoder(rec.Body).Decode(&c)

	// Second user cannot see or touch A's scenario.
	recB := doJSON(t, srv, "POST", "/api/v1/auth/register",
		map[string]string{"email": "other@example.com", "password": "password123", "name": "Other"}, "")
	var auth authResponse
	_ = json.NewDecoder(recB.Body).Decode(&auth)

	list := customScenarioList(t, srv, auth.Token)
	if len(list) != 0 {
		t.Errorf("user B should see no scenarios, got %d", len(list))
	}
	// Apply across owners is forbidden: the handler loads the scenario and
	// rejects it when the owner doesn't match (403), or 404 if not visible.
	rec = doJSON(t, srv, "POST", "/api/v1/scenarios/custom/"+c.ID+"/apply", nil, auth.Token)
	if rec.Code != http.StatusNotFound && rec.Code != http.StatusForbidden {
		t.Errorf("cross-owner apply = %d, want 404 or 403", rec.Code)
	}
}

func TestCustomScenariosRequireAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	if rec := doJSON(t, srv, "GET", "/api/v1/scenarios/custom", nil, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("list no-auth = %d, want 401", rec.Code)
	}
	rec := doJSON(t, srv, "POST", "/api/v1/scenarios/custom", validCustomBody(), "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("create no-auth = %d, want 401", rec.Code)
	}
}
