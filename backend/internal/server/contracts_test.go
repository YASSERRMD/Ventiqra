package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func contractList(t *testing.T, srv *Server, token string) []contractResponse {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/contracts", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list contracts: %d", rec.Code)
	}
	var list []contractResponse
	_ = json.NewDecoder(rec.Body).Decode(&list)
	return list
}

func TestContractsStartEmpty(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	if list := contractList(t, srv, token); len(list) != 0 {
		t.Errorf("expected empty, got %d", len(list))
	}
}

func TestSignContract(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/contracts", map[string]any{
		"customer_name": "Globex", "annual_value": 1_200_000_00, "discount_pct": 10, "term_years": 2,
	}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("sign: %d body=%s", rec.Code, rec.Body.String())
	}
	var c contractResponse
	_ = json.NewDecoder(rec.Body).Decode(&c)
	if c.CustomerName != "Globex" || c.Status != "active" {
		t.Errorf("unexpected contract: %+v", c)
	}
	// 10% discount: 1.2M → 1.08M
	if c.AnnualValue != 1_080_000_00 {
		t.Errorf("annual value = %d, want 1080000 (10%% off 1.2M)", c.AnnualValue)
	}
	if c.TermDays != 720 {
		t.Errorf("term = %d, want 720 (2 years)", c.TermDays)
	}
}

func TestSignContractValidates(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	cases := []map[string]any{
		{"customer_name": "", "annual_value": 100_000_00},
		{"customer_name": "X", "annual_value": 0},
		{"customer_name": "X", "annual_value": 100_000_00, "discount_pct": 80},
	}
	for _, body := range cases {
		rec := doJSON(t, srv, "POST", "/api/v1/companies/me/contracts", body, token)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("validation: %d, want 400 for %+v", rec.Code, body)
		}
	}
}

func TestContractsRequireAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	if rec := doJSON(t, srv, "GET", "/api/v1/companies/me/contracts", nil, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("no-auth = %d, want 401", rec.Code)
	}
}
