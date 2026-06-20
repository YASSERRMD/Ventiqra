package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func launchReadyProduct(t *testing.T, srv *Server, token, name string) productResponse {
	t.Helper()
	p := buildProduct(t, srv, token, name)
	setProgress(t, srv, token, p.ID, 100)
	setStage(t, srv, token, p.ID, "building")
	rec := doJSON(t, srv, "POST", "/api/v1/products/"+p.ID+"/launch", nil, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("launch %s: %d body=%s", name, rec.Code, rec.Body.String())
	}
	return p
}

func TestLaunchSeedsCustomers(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	p := launchReadyProduct(t, srv, token, "Seeded")

	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/customers", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list customers: %d body=%s", rec.Code, rec.Body.String())
	}
	var list []customerStateResponse
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(list) != 1 || list[0].ProductID != p.ID {
		t.Fatalf("expected one customer state for %s, got %+v", p.ID, list)
	}
	if list[0].Total <= 0 {
		t.Errorf("seeded total = %d, want > 0", list[0].Total)
	}
}

func TestCustomersEvolveOnTick(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	p := launchReadyProduct(t, srv, token, "Live")

	before := customerTotal(t, srv, token, p.ID)
	for i := 0; i < 10; i++ {
		tickOnce(t, srv, token)
	}
	after := customerTotal(t, srv, token, p.ID)

	// At high satisfaction the customer base should be non-decreasing over a
	// short window, and at any rate must reflect the deterministic evolution.
	if after < 0 {
		t.Errorf("customer total = %d, want >= 0", after)
	}
	if after == before {
		// churn/acquisition should move the number over 10 days.
		t.Errorf("customers did not evolve over 10 ticks: %d -> %d", before, after)
	}
}

func TestCustomersRequireCompany(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/customers", nil, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestCustomersRequireAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/customers", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func customerTotal(t *testing.T, srv *Server, token, productID string) int {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/customers", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list customers: %d", rec.Code)
	}
	var list []customerStateResponse
	_ = json.NewDecoder(rec.Body).Decode(&list)
	for _, c := range list {
		if c.ProductID == productID {
			return c.Total
		}
	}
	t.Fatalf("no customer state for %s", productID)
	return 0
}
