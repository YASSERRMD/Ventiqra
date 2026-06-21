package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func scenariosList(t *testing.T, srv *Server, token string) []scenarioResponse {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/scenarios", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list scenarios: %d body=%s", rec.Code, rec.Body.String())
	}
	var list []scenarioResponse
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode scenarios: %v", err)
	}
	return list
}

func TestListScenariosReturnsCatalog(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)

	list := scenariosList(t, srv, token)
	if len(list) < 4 {
		t.Fatalf("expected at least 4 scenarios, got %d", len(list))
	}
	ids := map[string]bool{}
	for _, sc := range list {
		ids[sc.ID] = true
		if sc.Name == "" || sc.Industry == "" || sc.Difficulty == "" {
			t.Errorf("scenario %+v has empty field", sc)
		}
		if sc.StartingCashCents <= 0 {
			t.Errorf("scenario %q non-positive cash", sc.ID)
		}
		if sc.Market.TAM <= 0 {
			t.Errorf("scenario %q non-positive TAM", sc.ID)
		}
	}
	for _, want := range []string{"bootstrap_saas", "vc_funded_startup", "hardware_startup", "marketplace"} {
		if !ids[want] {
			t.Errorf("missing scenario %q", want)
		}
	}
}

func TestApplyScenarioSetsCashAndIndustry(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	// Create a company with default cash first.
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	rec := doJSON(t, srv, "POST", "/api/v1/scenarios/bootstrap_saas/apply", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("apply: %d body=%s", rec.Code, rec.Body.String())
	}
	var res applyScenarioResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode apply: %v", err)
	}
	if res.Scenario.ID != "bootstrap_saas" {
		t.Errorf("scenario id = %q, want bootstrap_saas", res.Scenario.ID)
	}
	if res.Company.Industry != res.Scenario.Industry {
		t.Errorf("company industry %q != scenario industry %q", res.Company.Industry, res.Scenario.Industry)
	}
	if res.Company.CashCents != res.Scenario.StartingCashCents {
		t.Errorf("company cash %d != scenario cash %d", res.Company.CashCents, res.Scenario.StartingCashCents)
	}
}

func TestApplyUnknownScenarioIs404(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	rec := doJSON(t, srv, "POST", "/api/v1/scenarios/does-not-exist/apply", nil, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("unknown scenario status = %d, want 404", rec.Code)
	}
}

func TestApplyScenarioRequiresCompany(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	// No company created.
	rec := doJSON(t, srv, "POST", "/api/v1/scenarios/bootstrap_saas/apply", nil, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("apply without company status = %d, want 404", rec.Code)
	}
}

func TestApplyScenarioSeedsMarket(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	if rec := doJSON(t, srv, "POST", "/api/v1/scenarios/vc_funded_startup/apply", nil, token); rec.Code != http.StatusOK {
		t.Fatalf("apply: %d body=%s", rec.Code, rec.Body.String())
	}

	// The market endpoint should now reflect the scenario's seeded values.
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/market", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("market: %d body=%s", rec.Code, rec.Body.String())
	}
	var m marketResponse
	if err := json.NewDecoder(rec.Body).Decode(&m); err != nil {
		t.Fatalf("decode market: %v", err)
	}
	if m.TAM == 0 {
		t.Errorf("expected seeded TAM, got 0")
	}
}

func TestScenariosRequireAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	if rec := doJSON(t, srv, "GET", "/api/v1/scenarios", nil, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("list no-auth status = %d, want 401", rec.Code)
	}
	rec := doJSON(t, srv, "POST", "/api/v1/scenarios/bootstrap_saas/apply", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("apply no-auth status = %d, want 401", rec.Code)
	}
}
