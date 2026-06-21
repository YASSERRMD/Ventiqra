package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// pendingDecisionFor fetches the owner's pending decision card, returning the
// recorder so callers can assert on status codes (e.g. 404 when none pending).
func pendingDecisionFor(t *testing.T, srv *Server, token string) (pendingDecisionResponse, *httptest.ResponseRecorder) {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/decisions/pending", nil, token)
	if rec.Code != http.StatusOK {
		return pendingDecisionResponse{}, rec
	}
	var d pendingDecisionResponse
	if err := json.NewDecoder(rec.Body).Decode(&d); err != nil {
		t.Fatalf("decode pending decision: %v", err)
	}
	return d, rec
}

// tickN advances the owner's simulation by n days, failing on any non-200.
func tickN(t *testing.T, srv *Server, token string, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		if rec := doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token); rec.Code != http.StatusOK {
			t.Fatalf("tick %d: %d body=%s", i, rec.Code, rec.Body.String())
		}
	}
}

func TestNoPendingDecisionInitially(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	_, rec := pendingDecisionFor(t, srv, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("initial pending status = %d, want 404", rec.Code)
	}
}

func TestDecisionOfferedOnCadence(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	// Enough cash that the company survives 10 days of burn.
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co", "starting_cash_cents": 50_000_000_00}, token)

	// Advance to day 10; a decision card should be offered (cadence = 10).
	tickN(t, srv, token, 10)

	card, rec := pendingDecisionFor(t, srv, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("pending after 10 ticks: %d body=%s", rec.Code, rec.Body.String())
	}
	if card.Title == "" {
		t.Errorf("expected a card with a title, got %+v", card)
	}
	if len(card.Choices) != 2 {
		t.Errorf("expected 2 choices, got %d", len(card.Choices))
	}
	if card.SimDayOffered != 10 {
		t.Errorf("offered day = %d, want 10", card.SimDayOffered)
	}
}

func TestResolveDecisionAppliesEffects(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co", "starting_cash_cents": 50_000_000_00}, token)
	tickN(t, srv, token, 10)

	card, rec := pendingDecisionFor(t, srv, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("pending: %d", rec.Code)
	}
	// Resolve with the first choice.
	choiceID := card.Choices[0].ID
	rec = doJSON(t, srv, "POST", "/api/v1/companies/me/decisions/"+card.ID+"/resolve",
		map[string]any{"choice_id": choiceID}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("resolve status = %d body=%s", rec.Code, rec.Body.String())
	}
	var res resolveDecisionResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode resolve: %v", err)
	}
	if res.Outcome != "success" && res.Outcome != "failure" {
		t.Errorf("outcome = %q, want success/failure", res.Outcome)
	}

	// The card is no longer pending.
	_, rec = pendingDecisionFor(t, srv, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("after resolve, pending status = %d, want 404", rec.Code)
	}
}

func TestResolveTwiceIsConflict(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co", "starting_cash_cents": 50_000_000_00}, token)
	tickN(t, srv, token, 10)

	card, rec := pendingDecisionFor(t, srv, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("pending: %d", rec.Code)
	}
	choiceID := card.Choices[0].ID
	// First resolve succeeds.
	if rec := doJSON(t, srv, "POST", "/api/v1/companies/me/decisions/"+card.ID+"/resolve",
		map[string]any{"choice_id": choiceID}, token); rec.Code != http.StatusOK {
		t.Fatalf("first resolve = %d body=%s", rec.Code, rec.Body.String())
	}
	// Second resolve is a conflict (no longer pending).
	rec = doJSON(t, srv, "POST", "/api/v1/companies/me/decisions/"+card.ID+"/resolve",
		map[string]any{"choice_id": choiceID}, token)
	if rec.Code != http.StatusConflict {
		t.Errorf("second resolve status = %d, want 409", rec.Code)
	}
}

func TestResolveRequiresChoiceID(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co", "starting_cash_cents": 50_000_000_00}, token)
	tickN(t, srv, token, 10)

	card, _ := pendingDecisionFor(t, srv, token)
	// Missing choice_id.
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/decisions/"+card.ID+"/resolve",
		map[string]any{}, token)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("missing choice_id status = %d, want 400", rec.Code)
	}
	// Bogus choice_id.
	rec = doJSON(t, srv, "POST", "/api/v1/companies/me/decisions/"+card.ID+"/resolve",
		map[string]any{"choice_id": "does-not-exist"}, token)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("bogus choice_id status = %d, want 400", rec.Code)
	}
}

func TestDecisionRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	if rec := doJSON(t, srv, "GET", "/api/v1/companies/me/decisions/pending", nil, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("pending no-auth status = %d, want 401", rec.Code)
	}
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/decisions/00000000-0000-0000-0000-000000000000/resolve",
		map[string]any{"choice_id": "x"}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("resolve no-auth status = %d, want 401", rec.Code)
	}
}
