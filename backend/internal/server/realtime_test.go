package server

import (
	"net/http"
	"testing"
)

func TestWebSocketRequiresToken(t *testing.T) {
	srv := newAuthTestServer(t)
	// The WS handler responds to a plain GET (no upgrade) with the same auth
	// gating. Without a token query param it must 401 before upgrading.
	rec := doJSON(t, srv, "GET", "/api/v1/realtime", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("no token: status = %d, want 401", rec.Code)
	}
}

func TestWebSocketRejectsBadToken(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "GET", "/api/v1/realtime?token=not-a-real-jwt", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("bad token: status = %d, want 401", rec.Code)
	}
}

func TestWebSocketRequiresCompany(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	// Valid token, but no company yet → 404 before upgrade.
	rec := doJSON(t, srv, "GET", "/api/v1/realtime?token="+token, nil, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("no company: status = %d, want 404", rec.Code)
	}
}

func TestBroadcastHelpersAreNilSafe(t *testing.T) {
	// With a hub configured, broadcasting must not panic on a fresh company.
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "RT Co"}, token)
	// A tick triggers broadcastTick internally; it must not error.
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/sim/tick", nil, token)
	if rec.Code != http.StatusOK {
		t.Errorf("tick after broadcast = %d, want 200", rec.Code)
	}
}
