package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func getControl(t *testing.T, srv *Server, token string) simControlResponse {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/sim/control", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("get control: %d body=%s", rec.Code, rec.Body.String())
	}
	var c simControlResponse
	_ = json.NewDecoder(rec.Body).Decode(&c)
	return c
}

func TestSimControlDefaultsPaused1x(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	c := getControl(t, srv, token)
	if c.Mode != "paused" || c.Speed != 1 {
		t.Errorf("default = %+v, want paused/1", c)
	}
}

func TestPauseAndResume(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	if rec := doJSON(t, srv, "POST", "/api/v1/companies/me/sim/resume", nil, token); rec.Code != http.StatusOK {
		t.Fatalf("resume: %d", rec.Code)
	}
	if c := getControl(t, srv, token); c.Mode != "auto" {
		t.Errorf("after resume mode = %q, want auto", c.Mode)
	}
	if rec := doJSON(t, srv, "POST", "/api/v1/companies/me/sim/pause", nil, token); rec.Code != http.StatusOK {
		t.Fatalf("pause: %d", rec.Code)
	}
	if c := getControl(t, srv, token); c.Mode != "paused" {
		t.Errorf("after pause mode = %q, want paused", c.Mode)
	}
}

func TestSetSpeedValidatesRange(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	for _, sp := range []int{1, 5, 30} {
		rec := doJSON(t, srv, "POST", "/api/v1/companies/me/sim/speed", map[string]any{"speed": sp}, token)
		if rec.Code != http.StatusOK {
			t.Errorf("speed %d: %d, want 200", sp, rec.Code)
		}
	}
	for _, sp := range []int{0, 2, 7, 50} {
		rec := doJSON(t, srv, "POST", "/api/v1/companies/me/sim/speed", map[string]any{"speed": sp}, token)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("bad speed %d: %d, want 400", sp, rec.Code)
		}
	}
}

func TestSetSpeedPersists(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	doJSON(t, srv, "POST", "/api/v1/companies/me/sim/speed", map[string]any{"speed": 30}, token)
	if c := getControl(t, srv, token); c.Speed != 30 {
		t.Errorf("speed = %d, want 30", c.Speed)
	}
}

func TestSimControlRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	if rec := doJSON(t, srv, "GET", "/api/v1/companies/me/sim/control", nil, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("no-auth = %d, want 401", rec.Code)
	}
}
