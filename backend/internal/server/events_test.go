package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func eventsList(t *testing.T, srv *Server, token string) []gameEventResponse {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/events", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("events: %d body=%s", rec.Code, rec.Body.String())
	}
	var list []gameEventResponse
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return list
}

func TestEventsEmptyInitially(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	if list := eventsList(t, srv, token); len(list) != 0 {
		t.Errorf("expected 0 events initially, got %d", len(list))
	}
}

func TestEventsFireOverManyTicks(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co", "starting_cash_cents": 50_000_000}, token)

	for i := 0; i < 200; i++ {
		tickOnce(t, srv, token)
	}
	list := eventsList(t, srv, token)
	if len(list) == 0 {
		t.Fatalf("expected events to fire over 200 ticks")
	}

	// Every recorded event has a valid kind.
	kinds := map[string]bool{"positive": true, "negative": true, "neutral": true, "crisis": true}
	for _, e := range list {
		if !kinds[e.Kind] {
			t.Errorf("event kind %q invalid", e.Kind)
		}
		if e.Title == "" {
			t.Errorf("event missing title: %+v", e)
		}
	}
}

func TestEventsRequireAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/events", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
