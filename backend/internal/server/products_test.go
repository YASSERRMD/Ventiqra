package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateAndListProducts(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)

	// A company is required before products can be created.
	if rec := doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Product Co"}, token); rec.Code != http.StatusCreated {
		t.Fatalf("create company: %d body=%s", rec.Code, rec.Body.String())
	}

	// Empty list before any products exist.
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/products", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list empty status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var empty []productResponse
	if err := json.NewDecoder(rec.Body).Decode(&empty); err != nil {
		t.Fatalf("decode empty list: %v", err)
	}
	if empty == nil || len(empty) != 0 {
		t.Errorf("expected JSON [] not nil, got %v", empty)
	}

	// Create first product.
	rec = doJSON(t, srv, "POST", "/api/v1/companies/me/products", map[string]any{"name": "Acme App"}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create product status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var p1 productResponse
	if err := json.NewDecoder(rec.Body).Decode(&p1); err != nil {
		t.Fatalf("decode product: %v", err)
	}
	if p1.Name != "Acme App" || p1.Slug != "acme-app" {
		t.Errorf("product = %+v", p1)
	}
	if p1.Stage != "idea" {
		t.Errorf("stage = %q, want idea", p1.Stage)
	}
	if p1.DevProgress != 0 {
		t.Errorf("dev_progress = %v, want 0", p1.DevProgress)
	}
	if p1.PriceCents != nil {
		t.Errorf("price_cents = %v, want nil", *p1.PriceCents)
	}

	// Create a second product with the same name (suffixed slug).
	rec = doJSON(t, srv, "POST", "/api/v1/companies/me/products", map[string]any{"name": "Acme App"}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create duplicate-name product: %d", rec.Code)
	}
	var p2 productResponse
	_ = json.NewDecoder(rec.Body).Decode(&p2)
	if p2.Slug != "acme-app-1" {
		t.Errorf("second slug = %q, want acme-app-1", p2.Slug)
	}

	// List returns both in creation order.
	rec = doJSON(t, srv, "GET", "/api/v1/companies/me/products", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d", rec.Code)
	}
	var list []productResponse
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("list len = %d, want 2", len(list))
	}
	if list[0].ID != p1.ID || list[1].ID != p2.ID {
		t.Errorf("order mismatch")
	}
}

func TestCreateProductValidation(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	cases := []struct {
		name string
		body map[string]any
	}{
		{"missing name", map[string]any{}},
		{"empty name", map[string]any{"name": ""}},
		{"name too long", map[string]any{"name": string(make([]byte, 121))}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := doJSON(t, srv, "POST", "/api/v1/companies/me/products", c.body, token)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("%s: status = %d, want 400", c.name, rec.Code)
			}
		})
	}
}

func TestCreateProductRequiresCompany(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/products", map[string]any{"name": "X"}, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestCreateProductRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/products", map[string]any{"name": "X"}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestUpdateProductStage(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/products", map[string]any{"name": "Stageful"}, token)
	var p productResponse
	_ = json.NewDecoder(rec.Body).Decode(&p)

	rec = doJSON(t, srv, "PATCH", "/api/v1/products/"+p.ID+"/stage", map[string]any{"stage": "building"}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("update stage status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var updated productResponse
	if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if updated.Stage != "building" {
		t.Errorf("stage = %q, want building", updated.Stage)
	}
}

func TestUpdateProductStageInvalid(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/products", map[string]any{"name": "Stageful"}, token)
	var p productResponse
	_ = json.NewDecoder(rec.Body).Decode(&p)

	rec = doJSON(t, srv, "PATCH", "/api/v1/products/"+p.ID+"/stage", map[string]any{"stage": "nope"}, token)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestUpdateProductProgress(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/products", map[string]any{"name": "Progressful"}, token)
	var p productResponse
	_ = json.NewDecoder(rec.Body).Decode(&p)

	progress := 73.5
	rec = doJSON(t, srv, "PATCH", "/api/v1/products/"+p.ID+"/progress", map[string]any{"progress": progress}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("update progress status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var updated productResponse
	if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if updated.DevProgress < 73.4 || updated.DevProgress > 73.6 {
		t.Errorf("dev_progress = %v, want ~73.5", updated.DevProgress)
	}
}

func TestUpdateProductProgressOutOfRange(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/products", map[string]any{"name": "P"}, token)
	var p productResponse
	_ = json.NewDecoder(rec.Body).Decode(&p)

	cases := []struct {
		name string
		body map[string]any
	}{
		{"over 100", map[string]any{"progress": 150}},
		{"under 0", map[string]any{"progress": -1}},
		{"missing", map[string]any{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := doJSON(t, srv, "PATCH", "/api/v1/products/"+p.ID+"/progress", c.body, token)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("%s: status = %d, want 400", c.name, rec.Code)
			}
		})
	}
}

func TestProductForbiddenForOtherOwner(t *testing.T) {
	srv := newAuthTestServer(t)

	// Owner A creates a company and a product.
	tokenA := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "A Co"}, tokenA)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/products", map[string]any{"name": "A Product"}, tokenA)
	var p productResponse
	_ = json.NewDecoder(rec.Body).Decode(&p)

	// Owner B registers and creates their own company.
	recB := doJSON(t, srv, "POST", "/api/v1/auth/register",
		map[string]string{"email": "other@example.com", "password": "password123", "name": "Other"}, "")
	if recB.Code != http.StatusOK {
		t.Fatalf("register B: %d", recB.Code)
	}
	var auth authResponse
	_ = json.NewDecoder(recB.Body).Decode(&auth)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "B Co"}, auth.Token)

	// B cannot mutate A's product.
	rec = doJSON(t, srv, "PATCH", "/api/v1/products/"+p.ID+"/stage", map[string]any{"stage": "launched"}, auth.Token)
	if rec.Code != http.StatusForbidden {
		t.Errorf("cross-owner stage status = %d, want 403", rec.Code)
	}
	rec = doJSON(t, srv, "PATCH", "/api/v1/products/"+p.ID+"/progress", map[string]any{"progress": 50}, auth.Token)
	if rec.Code != http.StatusForbidden {
		t.Errorf("cross-owner progress status = %d, want 403", rec.Code)
	}
}

func TestUpdateProductMissingIs404(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	missing := "00000000-0000-0000-0000-000000000000"
	rec := doJSON(t, srv, "PATCH", "/api/v1/products/"+missing+"/stage", map[string]any{"stage": "launched"}, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("stage missing status = %d, want 404", rec.Code)
	}
	rec = doJSON(t, srv, "PATCH", "/api/v1/products/"+missing+"/progress", map[string]any{"progress": 50}, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("progress missing status = %d, want 404", rec.Code)
	}
}
