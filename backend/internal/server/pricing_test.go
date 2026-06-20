package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestSetProductPrice(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	p := buildProduct(t, srv, token, "Priced")

	rec := doJSON(t, srv, "PATCH", "/api/v1/products/"+p.ID+"/price", map[string]any{"price_cents": 1999}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("set price status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var updated productResponse
	if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if updated.PriceCents == nil || *updated.PriceCents != 1999 {
		t.Errorf("price = %v, want 1999", updated.PriceCents)
	}
}

func TestSetProductPriceValidation(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	p := buildProduct(t, srv, token, "P")

	cases := []map[string]any{
		{},
		{"price_cents": -5},
	}
	for _, body := range cases {
		rec := doJSON(t, srv, "PATCH", "/api/v1/products/"+p.ID+"/price", body, token)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	}
}

func TestPricingExperimentRecorded(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	p := buildProduct(t, srv, token, "Experiments")

	doJSON(t, srv, "PATCH", "/api/v1/products/"+p.ID+"/price", map[string]any{"price_cents": 1000}, token)
	doJSON(t, srv, "PATCH", "/api/v1/products/"+p.ID+"/price", map[string]any{"price_cents": 1500}, token)

	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/pricing-experiments", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list experiments: %d", rec.Code)
	}
	var list []pricingExperimentResponse
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("experiments len = %d, want 2", len(list))
	}
	// Newest first: second change on top.
	if list[0].NewPriceCents != 1500 {
		t.Errorf("newest new price = %d, want 1500", list[0].NewPriceCents)
	}
	if list[1].OldPriceCents != nil {
		t.Errorf("first change should have nil old price, got %v", list[1].OldPriceCents)
	}
}

func TestPricingExperimentsRequireAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/pricing-experiments", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestPricingAffectsRevenueOnTick(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	p := launchReadyProduct(t, srv, token, "Paid")

	// Set a monthly price; revenue should appear after a tick.
	doJSON(t, srv, "PATCH", "/api/v1/products/"+p.ID+"/price", map[string]any{"price_cents": 1000}, token)
	tickOnce(t, srv, token)

	m := metricsFor(t, srv, token)
	if m.RevenueCents <= 0 {
		t.Errorf("expected positive daily revenue after pricing + tick, got %d", m.RevenueCents)
	}
}
