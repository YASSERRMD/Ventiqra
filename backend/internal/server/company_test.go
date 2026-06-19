package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateCompanyAndProfile(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)

	rec := doJSON(t, srv, "POST", "/api/v1/companies",
		map[string]any{"name": "Acme", "industry": "Aerospace", "description": "Rockets"}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var c companyResponse
	if err := json.NewDecoder(rec.Body).Decode(&c); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if c.Name != "Acme" || c.Slug != "acme" {
		t.Errorf("company = %+v", c)
	}
	if c.CashCents != defaultStartingCashCents {
		t.Errorf("cash = %d, want default %d", c.CashCents, defaultStartingCashCents)
	}

	// GET /me returns the created company.
	rec = doJSON(t, srv, "GET", "/api/v1/companies/me", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("me status = %d", rec.Code)
	}
	var me companyResponse
	_ = json.NewDecoder(rec.Body).Decode(&me)
	if me.ID != c.ID {
		t.Errorf("me id mismatch %s != %s", me.ID, c.ID)
	}

	// GET by id.
	rec = doJSON(t, srv, "GET", "/api/v1/companies/"+c.ID, nil, token)
	if rec.Code != http.StatusOK {
		t.Errorf("get by id status = %d", rec.Code)
	}
}

func TestCreateCompanyValidation(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)

	cases := []struct {
		name string
		body map[string]any
	}{
		{"missing name", map[string]any{"industry": "X"}},
		{"empty name", map[string]any{"name": ""}},
		{"negative cash", map[string]any{"name": "Co", "starting_cash_cents": -1}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := doJSON(t, srv, "POST", "/api/v1/companies", c.body, token)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("%s: status = %d, want 400", c.name, rec.Code)
			}
		})
	}
}

func TestCreateCompanyRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "X"}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestMyCompanyNotFound(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me", nil, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestGetCompanyForbiddenForOtherOwner(t *testing.T) {
	srv := newAuthTestServer(t)

	// Owner A creates a company.
	tokenA := registerAndLogin(t, srv)
	rec := doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "A Co"}, tokenA)
	var c companyResponse
	_ = json.NewDecoder(rec.Body).Decode(&c)

	// Owner B registers and tries to view A's company.
	tokenB := doJSON(t, srv, "POST", "/api/v1/auth/register",
		map[string]string{"email": "other@example.com", "password": "password123", "name": "Other"}, "")
	if tokenB.Code != http.StatusOK {
		t.Fatalf("register B: %d", tokenB.Code)
	}
	var auth authResponse
	_ = json.NewDecoder(tokenB.Body).Decode(&auth)

	rec = doJSON(t, srv, "GET", "/api/v1/companies/"+c.ID, nil, auth.Token)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}
