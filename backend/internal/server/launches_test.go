package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func buildProduct(t *testing.T, srv *Server, token, name string) productResponse {
	t.Helper()
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/products", map[string]any{"name": name}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create product: %d body=%s", rec.Code, rec.Body.String())
	}
	var p productResponse
	_ = json.NewDecoder(rec.Body).Decode(&p)
	return p
}

func setStage(t *testing.T, srv *Server, token, id, stage string) {
	t.Helper()
	rec := doJSON(t, srv, "PATCH", "/api/v1/products/"+id+"/stage", map[string]any{"stage": stage}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("set stage %s: %d body=%s", stage, rec.Code, rec.Body.String())
	}
}

func setProgress(t *testing.T, srv *Server, token, id string, progress float64) {
	t.Helper()
	rec := doJSON(t, srv, "PATCH", "/api/v1/products/"+id+"/progress", map[string]any{"progress": progress}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("set progress: %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLaunchRequiresBuildingStage(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	p := buildProduct(t, srv, token, "Idea Only")

	// Idea-stage product cannot be launched.
	rec := doJSON(t, srv, "POST", "/api/v1/products/"+p.ID+"/launch", nil, token)
	if rec.Code != http.StatusConflict {
		t.Errorf("launch idea status = %d, want 409", rec.Code)
	}
}

func TestLaunchNotReady(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	p := buildProduct(t, srv, token, "Early")
	setStage(t, srv, token, p.ID, "building")
	// Zero progress, no team → readiness 0 → not ready.

	rec := doJSON(t, srv, "POST", "/api/v1/products/"+p.ID+"/launch", nil, token)
	if rec.Code != http.StatusConflict {
		t.Fatalf("launch not-ready status = %d, want 409, body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body["readiness"] == nil {
		t.Errorf("expected readiness in not-ready response, got %v", body)
	}
}

func TestLaunchReadyFlow(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	p := buildProduct(t, srv, token, "Ready App")
	setProgress(t, srv, token, p.ID, 100)
	setStage(t, srv, token, p.ID, "building")

	rec := doJSON(t, srv, "POST", "/api/v1/products/"+p.ID+"/launch", nil, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("launch status = %d, want 201, body=%s", rec.Code, rec.Body.String())
	}
	var res launchResultResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode launch: %v", err)
	}
	if res.Readiness < 40 {
		t.Errorf("readiness = %v, want >= 40", res.Readiness)
	}
	if res.InitialCustomers <= 0 {
		t.Errorf("initial customers = %d, want > 0", res.InitialCustomers)
	}
	if res.Product.Stage != "launched" {
		t.Errorf("product stage = %q, want launched", res.Product.Stage)
	}

	// A second launch of the now-launched product is forbidden.
	rec = doJSON(t, srv, "POST", "/api/v1/products/"+p.ID+"/launch", nil, token)
	if rec.Code != http.StatusConflict {
		t.Errorf("re-launch status = %d, want 409", rec.Code)
	}
}

func TestLaunchHistory(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	p := buildProduct(t, srv, token, "Historic")
	setProgress(t, srv, token, p.ID, 100)
	setStage(t, srv, token, p.ID, "building")
	if rec := doJSON(t, srv, "POST", "/api/v1/products/"+p.ID+"/launch", nil, token); rec.Code != http.StatusCreated {
		t.Fatalf("launch: %d body=%s", rec.Code, rec.Body.String())
	}

	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/launches", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list launches: %d", rec.Code)
	}
	var list []launchResponse
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode launches: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("launches len = %d, want 1", len(list))
	}
	if list[0].ProductID != p.ID || list[0].ProductName != "Historic" {
		t.Errorf("launch entry = %+v", list[0])
	}
}

func TestLaunchRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "POST", "/api/v1/products/00000000-0000-0000-0000-000000000000/launch", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestLaunchForbiddenForOtherOwner(t *testing.T) {
	srv := newAuthTestServer(t)
	tokenA := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "A Co"}, tokenA)
	p := buildProduct(t, srv, tokenA, "A Product")
	setProgress(t, srv, tokenA, p.ID, 100)
	setStage(t, srv, tokenA, p.ID, "building")

	recB := doJSON(t, srv, "POST", "/api/v1/auth/register",
		map[string]string{"email": "other@example.com", "password": "password123", "name": "Other"}, "")
	var auth authResponse
	_ = json.NewDecoder(recB.Body).Decode(&auth)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "B Co"}, auth.Token)

	rec := doJSON(t, srv, "POST", "/api/v1/products/"+p.ID+"/launch", nil, auth.Token)
	if rec.Code != http.StatusForbidden {
		t.Errorf("cross-owner launch status = %d, want 403", rec.Code)
	}
}
