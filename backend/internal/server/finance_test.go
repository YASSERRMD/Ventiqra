package server

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/YASSERRMD/Ventiqra/backend/internal/finance"
)

func TestGetFinanceDefaults(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/finance", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("finance status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var fin financeResponse
	if err := json.NewDecoder(rec.Body).Decode(&fin); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Default burn with no team/customers = base overhead + infra floor.
	if fin.Burn.Infra != finance.InfraCost(0) {
		t.Errorf("default infra = %d, want floor %d", fin.Burn.Infra, finance.InfraCost(0))
	}
	if fin.Burn.Total != fin.Burn.Base+fin.Burn.Infra {
		t.Errorf("default burn total = %d, want base+infra %d", fin.Burn.Total, fin.Burn.Base+fin.Burn.Infra)
	}
	if fin.MarketingBudgetCents != 0 {
		t.Errorf("default marketing = %d, want 0", fin.MarketingBudgetCents)
	}
}

func TestUpdateFinanceMarketingBudget(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	rec := doJSON(t, srv, "PATCH", "/api/v1/companies/me/finance",
		map[string]any{"marketing_budget_cents": 250000}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("update finance: %d body=%s", rec.Code, rec.Body.String())
	}
	var fin financeResponse
	_ = json.NewDecoder(rec.Body).Decode(&fin)
	if fin.MarketingBudgetCents != 250000 {
		t.Errorf("marketing = %d, want 250000", fin.MarketingBudgetCents)
	}
	if fin.Burn.Marketing != 250000 {
		t.Errorf("burn marketing = %d, want 250000", fin.Burn.Marketing)
	}
	if fin.Burn.Total != fin.Burn.Base+fin.Burn.Infra+fin.Burn.Marketing {
		t.Errorf("burn total = %d, want base+infra+marketing", fin.Burn.Total)
	}
}

func TestUpdateFinanceValidation(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	cases := []map[string]any{
		{},
		{"marketing_budget_cents": -1},
	}
	for _, body := range cases {
		rec := doJSON(t, srv, "PATCH", "/api/v1/companies/me/finance", body, token)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	}
}

func TestFinanceReflectsPayrollAndInfra(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	// Hire an engineer with a known salary → payroll appears in the breakdown.
	doJSON(t, srv, "POST", "/api/v1/companies/me/employees",
		map[string]any{"name": "Dev", "role": "engineer", "salary_cents": 1200000, "skill": 100, "morale": 70}, token)

	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/finance", nil, token)
	var fin financeResponse
	_ = json.NewDecoder(rec.Body).Decode(&rec)
	rec = doJSON(t, srv, "GET", "/api/v1/companies/me/finance", nil, token)
	if err := json.NewDecoder(rec.Body).Decode(&fin); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if fin.Burn.Salaries != 1200000 {
		t.Errorf("salaries = %d, want 1200000", fin.Burn.Salaries)
	}
	if fin.Burn.Total != fin.Burn.Base+fin.Burn.Salaries+fin.Burn.Infra+fin.Burn.Marketing {
		t.Errorf("burn total does not match components")
	}
}

func TestFinanceRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/finance", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
