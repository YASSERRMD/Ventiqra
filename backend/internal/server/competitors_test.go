package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestListCompetitorsSeeds(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/competitors", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list competitors: %d body=%s", rec.Code, rec.Body.String())
	}
	var list []competitorResponse
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 seeded competitors, got %d", len(list))
	}
	for _, c := range list {
		if c.Name == "" || c.Strength < 0 || c.Strength > 100 {
			t.Errorf("malformed competitor: %+v", c)
		}
	}
}

func TestCompetitorsDeterministic(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/competitors", nil, token)
	var first []competitorResponse
	_ = json.NewDecoder(rec.Body).Decode(&first)
	rec = doJSON(t, srv, "GET", "/api/v1/companies/me/competitors", nil, token)
	var second []competitorResponse
	_ = json.NewDecoder(rec.Body).Decode(&second)
	if len(first) != len(second) {
		t.Fatalf("differing counts")
	}
	for i := range first {
		if first[i].Name != second[i].Name {
			t.Errorf("competitor %d name not deterministic: %q vs %q", i, first[i].Name, second[i].Name)
		}
	}
}

func TestCompetitorsGrowOnTick(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	// Seed via list, then advance several ticks.
	doJSON(t, srv, "GET", "/api/v1/companies/me/competitors", nil, token)
	for i := 0; i < 5; i++ {
		tickOnce(t, srv, token)
	}

	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/competitors", nil, token)
	var list []competitorResponse
	_ = json.NewDecoder(rec.Body).Decode(&list)
	// After 5 ticks at +0..2/day each, at least one rival should have grown.
	grew := false
	for _, c := range list {
		if c.Strength > 15 { // initial floor is 15
			grew = true
		}
	}
	if !grew {
		t.Errorf("expected at least one competitor to grow after ticks")
	}
}

func TestCompetitorsRequireAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/competitors", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
