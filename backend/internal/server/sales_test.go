package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func dealList(t *testing.T, srv *Server, token string) []dealResponse {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/deals", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list deals: %d", rec.Code)
	}
	var list []dealResponse
	_ = json.NewDecoder(rec.Body).Decode(&list)
	return list
}

func TestDealsStartEmpty(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	if list := dealList(t, srv, token); len(list) != 0 {
		t.Errorf("expected empty, got %d", len(list))
	}
}

func TestCreateDeal(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/deals", map[string]any{
		"name": "Acme Corp", "value_cents": 100_000_00,
	}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d body=%s", rec.Code, rec.Body.String())
	}
	var d dealResponse
	_ = json.NewDecoder(rec.Body).Decode(&d)
	if d.Name != "Acme Corp" || d.Stage != "lead" || d.ValueCents != 100_000_00 {
		t.Errorf("unexpected deal: %+v", d)
	}
}

func TestAdvanceDealProgresses(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/deals", map[string]any{"name": "D"}, token)
	var d dealResponse
	_ = json.NewDecoder(rec.Body).Decode(&d)
	// Advance: lead → qualified
	rec = doJSON(t, srv, "POST", "/api/v1/companies/me/deals/"+d.ID+"/advance", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("advance: %d", rec.Code)
	}
	var res advanceResponse
	_ = json.NewDecoder(rec.Body).Decode(&res)
	if res.Deal.Stage != "qualified" {
		t.Errorf("stage = %s, want qualified", res.Deal.Stage)
	}
}

func TestAdvanceClosesAtNegotiation(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/deals", map[string]any{"name": "D"}, token)
	var d dealResponse
	_ = json.NewDecoder(rec.Body).Decode(&d)
	// Advance through lead → qualified → proposal → negotiation → close.
	for i := 0; i < 4; i++ {
		rec = doJSON(t, srv, "POST", "/api/v1/companies/me/deals/"+d.ID+"/advance", nil, token)
		if rec.Code != http.StatusOK {
			t.Fatalf("advance %d: %d", i, rec.Code)
		}
	}
	var res advanceResponse
	_ = json.NewDecoder(rec.Body).Decode(&res)
	if res.Deal.Stage != "closed_won" && res.Deal.Stage != "closed_lost" {
		t.Errorf("final stage = %s, want closed", res.Deal.Stage)
	}
}

func TestAdvanceClosedDealIsConflict(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/deals", map[string]any{"name": "D"}, token)
	var d dealResponse
	_ = json.NewDecoder(rec.Body).Decode(&d)
	for i := 0; i < 4; i++ {
		doJSON(t, srv, "POST", "/api/v1/companies/me/deals/"+d.ID+"/advance", nil, token)
	}
	rec = doJSON(t, srv, "POST", "/api/v1/companies/me/deals/"+d.ID+"/advance", nil, token)
	if rec.Code != http.StatusConflict {
		t.Errorf("advance closed: %d, want 409", rec.Code)
	}
}

func TestDealsRequireAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	if rec := doJSON(t, srv, "GET", "/api/v1/companies/me/deals", nil, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("no-auth = %d, want 401", rec.Code)
	}
}
