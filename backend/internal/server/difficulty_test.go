package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func getDifficulty(t *testing.T, srv *Server, token string) difficultyResponse {
	t.Helper()
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/difficulty", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("get difficulty: %d", rec.Code)
	}
	var d difficultyResponse
	_ = json.NewDecoder(rec.Body).Decode(&d)
	return d
}

func TestDifficultyDefaultsNormal(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	d := getDifficulty(t, srv, token)
	if d.Level != "normal" {
		t.Errorf("default = %s, want normal", d.Level)
	}
	if d.Multipliers.BurnMultiplier != 1.0 {
		t.Errorf("normal burn = %v, want 1.0", d.Multipliers.BurnMultiplier)
	}
}

func TestSetDifficulty(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/difficulty", map[string]any{"level": "brutal"}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("set: %d body=%s", rec.Code, rec.Body.String())
	}
	d := getDifficulty(t, srv, token)
	if d.Level != "brutal" {
		t.Errorf("level = %s, want brutal", d.Level)
	}
	if d.Multipliers.BurnMultiplier <= 1.0 {
		t.Error("brutal burn should be > 1.0")
	}
}

func TestSetDifficultyValidates(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	for _, lvl := range []string{"easy", "normal", "hard", "brutal", "custom"} {
		rec := doJSON(t, srv, "POST", "/api/v1/companies/me/difficulty", map[string]any{"level": lvl}, token)
		if rec.Code != http.StatusOK {
			t.Errorf("level %s: %d, want 200", lvl, rec.Code)
		}
	}
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/difficulty", map[string]any{"level": "nightmare"}, token)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("bad level: %d, want 400", rec.Code)
	}
}

func TestDifficultyRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	if rec := doJSON(t, srv, "GET", "/api/v1/companies/me/difficulty", nil, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("no-auth = %d, want 401", rec.Code)
	}
}
