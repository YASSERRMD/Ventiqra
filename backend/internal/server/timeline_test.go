package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func timelineFor(t *testing.T, srv *Server, token string) timelineResponse {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/timeline", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("timeline: %d body=%s", rec.Code, rec.Body.String())
	}
	var tl timelineResponse
	if err := json.NewDecoder(rec.Body).Decode(&tl); err != nil {
		t.Fatalf("decode timeline: %v", err)
	}
	return tl
}

func TestTimelineRecordsFounding(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Timeline Co"}, token)

	tl := timelineFor(t, srv, token)
	if len(tl.Entries) == 0 {
		t.Fatal("expected at least the founding entry")
	}
	found := false
	for _, e := range tl.Entries {
		if e.Kind == "milestone" && e.Title == "Founded Timeline Co" {
			found = true
		}
	}
	if !found {
		t.Error("founding milestone not recorded")
	}
}

func TestTimelineMonthlySummaryAfterTicks(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co", "starting_cash_cents": 50_000_000_00}, token)
	// Advance into month 2 (day 31+).
	for i := 0; i < 35; i++ {
		if rec := doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token); rec.Code != http.StatusOK {
			t.Fatalf("tick %d: %d", i, rec.Code)
		}
	}
	tl := timelineFor(t, srv, token)
	if len(tl.Monthly) < 2 {
		t.Fatalf("expected at least 2 monthly summaries, got %d", len(tl.Monthly))
	}
	if tl.Monthly[0].Month != 1 || tl.Monthly[1].Month != 2 {
		t.Errorf("months = %d, %d", tl.Monthly[0].Month, tl.Monthly[1].Month)
	}
}

func TestTimelineRecordsFunding(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/funding/raise", map[string]any{"amount_cents": 1_000_000}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("raise: %d", rec.Code)
	}
	tl := timelineFor(t, srv, token)
	found := false
	for _, e := range tl.Entries {
		if e.Kind == "funding" {
			found = true
		}
	}
	if !found {
		t.Error("funding milestone not recorded")
	}
}

func TestTimelineRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/timeline", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("no-auth = %d, want 401", rec.Code)
	}
}

func TestTimelineRequiresCompany(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/timeline", nil, token)
	// ownerCompanyID returns 404 when no company exists.
	if rec.Code != http.StatusNotFound {
		t.Errorf("no company = %d, want 404", rec.Code)
	}
}
