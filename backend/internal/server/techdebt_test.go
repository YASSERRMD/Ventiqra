package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func getTechDebt(t *testing.T, srv *Server, token string) techDebtResponse {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/tech-debt", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("get tech debt: %d", rec.Code)
	}
	var td techDebtResponse
	_ = json.NewDecoder(rec.Body).Decode(&td)
	return td
}

func TestTechDebtStartsAtZero(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	td := getTechDebt(t, srv, token)
	if td.Debt != 0 || td.Quality != 100 {
		t.Errorf("initial = %+v, want debt=0 quality=100", td)
	}
}

func TestShippingFeatureAccumulatesDebt(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/features", map[string]any{"name": "F"}, token)
	var f featureResponse
	_ = json.NewDecoder(rec.Body).Decode(&f)
	// Ship it in one develop call.
	doJSON(t, srv, "POST", "/api/v1/companies/me/features/"+f.ID+"/develop", map[string]any{"points": 100}, token)
	td := getTechDebt(t, srv, token)
	if td.Debt == 0 {
		t.Error("expected debt after shipping, got 0")
	}
}

func TestRefactorReducesDebt(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co", "starting_cash_cents": 10_000_000_00}, token)
	// Ship two features to build up debt.
	for i := 0; i < 2; i++ {
		rec := doJSON(t, srv, "POST", "/api/v1/companies/me/features", map[string]any{"name": "F"}, token)
		var f featureResponse
		_ = json.NewDecoder(rec.Body).Decode(&f)
		doJSON(t, srv, "POST", "/api/v1/companies/me/features/"+f.ID+"/develop", map[string]any{"points": 100}, token)
	}
	before := getTechDebt(t, srv, token)
	if before.Debt == 0 {
		t.Fatal("expected debt before refactor")
	}
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/tech-debt/refactor", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("refactor: %d body=%s", rec.Code, rec.Body.String())
	}
	after := getTechDebt(t, srv, token)
	if after.Debt >= before.Debt {
		t.Errorf("debt did not decrease: before=%d after=%d", before.Debt, after.Debt)
	}
	if after.Refactors != before.Refactors+1 {
		t.Errorf("refactor count = %d, want %d", after.Refactors, before.Refactors+1)
	}
}

func TestRefactorAtZeroDebtIsConflict(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/tech-debt/refactor", nil, token)
	if rec.Code != http.StatusConflict {
		t.Errorf("refactor at 0 debt: %d, want 409", rec.Code)
	}
}

func TestTechDebtRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	if rec := doJSON(t, srv, "GET", "/api/v1/companies/me/tech-debt", nil, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("no-auth = %d, want 401", rec.Code)
	}
}
