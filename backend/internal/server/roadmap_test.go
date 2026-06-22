package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func featureList(t *testing.T, srv *Server, token string) []featureResponse {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/features", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list features: %d", rec.Code)
	}
	var list []featureResponse
	_ = json.NewDecoder(rec.Body).Decode(&list)
	return list
}

func TestFeaturesStartEmpty(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	if list := featureList(t, srv, token); len(list) != 0 {
		t.Errorf("expected empty, got %d", len(list))
	}
}

func TestCreateFeaturePersists(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/features", map[string]any{
		"name": "Dark mode", "priority": 5, "value_points": 15,
	}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d body=%s", rec.Code, rec.Body.String())
	}
	var f featureResponse
	_ = json.NewDecoder(rec.Body).Decode(&f)
	if f.Name != "Dark mode" || f.Priority != 5 || f.Status != "backlog" {
		t.Errorf("unexpected feature: %+v", f)
	}
	if len(featureList(t, srv, token)) != 1 {
		t.Error("feature not in list")
	}
}

func TestCreateFeatureRequiresName(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/features", map[string]any{"priority": 1}, token)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("no name: %d, want 400", rec.Code)
	}
}

func TestDevelopFeatureProgressesAndShips(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/features", map[string]any{
		"name": "F", "value_points": 10,
	}, token)
	var f featureResponse
	_ = json.NewDecoder(rec.Body).Decode(&f)

	// Develop by 60 → progress 60, not shipped.
	rec = doJSON(t, srv, "POST", "/api/v1/companies/me/features/"+f.ID+"/develop", map[string]any{"points": 60}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("develop1: %d", rec.Code)
	}
	var res1 struct {
		Feature featureResponse `json:"feature"`
		Shipped bool            `json:"shipped"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&res1)
	if res1.Feature.Progress != 60 || res1.Shipped {
		t.Errorf("develop1: progress=%d shipped=%v", res1.Feature.Progress, res1.Shipped)
	}

	// Develop by 50 → clamps to 100, ships.
	rec = doJSON(t, srv, "POST", "/api/v1/companies/me/features/"+f.ID+"/develop", map[string]any{"points": 50}, token)
	var res2 struct {
		Feature featureResponse `json:"feature"`
		Shipped bool            `json:"shipped"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&res2)
	if res2.Feature.Progress != 100 || !res2.Shipped {
		t.Errorf("develop2: progress=%d shipped=%v", res2.Feature.Progress, res2.Shipped)
	}
	if res2.Feature.Status != "shipped" {
		t.Errorf("status = %q, want shipped", res2.Feature.Status)
	}

	// Further develop is a conflict.
	rec = doJSON(t, srv, "POST", "/api/v1/companies/me/features/"+f.ID+"/develop", map[string]any{"points": 10}, token)
	if rec.Code != http.StatusConflict {
		t.Errorf("develop shipped: %d, want 409", rec.Code)
	}
}

func TestDeleteFeature(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/features", map[string]any{"name": "X"}, token)
	var f featureResponse
	_ = json.NewDecoder(rec.Body).Decode(&f)
	rec = doJSON(t, srv, "DELETE", "/api/v1/companies/me/features/"+f.ID, nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete: %d", rec.Code)
	}
	if len(featureList(t, srv, token)) != 0 {
		t.Error("not deleted")
	}
}

func TestFeaturesRequireAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	if rec := doJSON(t, srv, "GET", "/api/v1/companies/me/features", nil, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("no-auth = %d, want 401", rec.Code)
	}
}
