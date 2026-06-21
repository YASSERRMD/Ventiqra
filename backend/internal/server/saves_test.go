package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func saveSlotsList(t *testing.T, srv *Server, token string) []saveSlotResponse {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/saves", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list saves: %d body=%s", rec.Code, rec.Body.String())
	}
	var list []saveSlotResponse
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode saves: %v", err)
	}
	return list
}

func TestSaveSlotsStartEmpty(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	if list := saveSlotsList(t, srv, token); len(list) != 0 {
		t.Errorf("expected empty, got %d", len(list))
	}
}

func TestSaveRequiresCompany(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	rec := doJSON(t, srv, "POST", "/api/v1/saves", map[string]any{"slot": "a"}, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("save without company = %d, want 404", rec.Code)
	}
}

func TestSaveRejectsBadSlotName(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	cases := []string{"", "UPPER", "slot 1", "slot.1"}
	for _, c := range cases {
		rec := doJSON(t, srv, "POST", "/api/v1/saves", map[string]any{"slot": c}, token)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("slot %q: status = %d, want 400", c, rec.Code)
		}
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co", "starting_cash_cents": 5_000_000_00}, token)

	// Advance a few days so the snapshot has non-trivial state.
	for i := 0; i < 3; i++ {
		if rec := doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token); rec.Code != http.StatusOK {
			t.Fatalf("tick %d: %d", i, rec.Code)
		}
	}

	// Save into slot "run1".
	rec := doJSON(t, srv, "POST", "/api/v1/saves", map[string]any{"slot": "run1", "label": "checkpoint"}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save: %d body=%s", rec.Code, rec.Body.String())
	}
	var saved saveSlotResponse
	_ = json.NewDecoder(rec.Body).Decode(&saved)
	if saved.Slot != "run1" || saved.Label != "checkpoint" {
		t.Errorf("saved slot = %+v", saved)
	}
	if saved.Day < 1 {
		t.Errorf("saved day = %d, want >= 1", saved.Day)
	}

	// Advance more days to diverge the live state.
	for i := 0; i < 5; i++ {
		doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token)
	}

	// Load the slot back; the day should revert to the saved day.
	rec = doJSON(t, srv, "POST", "/api/v1/saves/run1/load", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("load: %d body=%s", rec.Code, rec.Body.String())
	}
	var loaded loadResultResponse
	_ = json.NewDecoder(rec.Body).Decode(&loaded)
	if loaded.Day != saved.Day {
		t.Errorf("after load day = %d, want %d", loaded.Day, saved.Day)
	}
	if loaded.Company.CashCents != saved.CashCents {
		t.Errorf("after load cash = %d, want %d", loaded.Company.CashCents, saved.CashCents)
	}
}

func TestLoadUnknownSlotIs404(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/saves/nope/load", nil, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("load unknown = %d, want 404", rec.Code)
	}
}

func TestDeleteSlot(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	doJSON(t, srv, "POST", "/api/v1/saves", map[string]any{"slot": "del"}, token)

	rec := doJSON(t, srv, "DELETE", "/api/v1/saves/del", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete: %d", rec.Code)
	}
	if list := saveSlotsList(t, srv, token); len(list) != 0 {
		t.Errorf("expected empty after delete, got %d", len(list))
	}
}

func TestSaveSlotLimitEnforced(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	// Fill up to the limit (5 distinct slots).
	for i := 0; i < 5; i++ {
		rec := doJSON(t, srv, "POST", "/api/v1/saves", map[string]any{"slot": "slot" + string(rune('1'+i))}, token)
		if rec.Code != http.StatusCreated {
			t.Fatalf("fill slot %d: %d body=%s", i, rec.Code, rec.Body.String())
		}
	}
	// A 6th distinct slot should be rejected.
	rec := doJSON(t, srv, "POST", "/api/v1/saves", map[string]any{"slot": "slot6"}, token)
	if rec.Code != http.StatusConflict {
		t.Errorf("6th slot = %d, want 409", rec.Code)
	}
	// Re-saving an existing slot is allowed (not a new slot).
	rec = doJSON(t, srv, "POST", "/api/v1/saves", map[string]any{"slot": "slot1"}, token)
	if rec.Code != http.StatusCreated {
		t.Errorf("overwrite slot1 = %d, want 201", rec.Code)
	}
}

func TestSavesRequireAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	if rec := doJSON(t, srv, "GET", "/api/v1/saves", nil, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("list no-auth = %d, want 401", rec.Code)
	}
	rec := doJSON(t, srv, "POST", "/api/v1/saves", map[string]any{"slot": "a"}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("save no-auth = %d, want 401", rec.Code)
	}
}
