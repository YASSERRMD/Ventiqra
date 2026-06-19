package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"
)

var okParser TokenParser = func(token string) (string, error) {
	if token == "valid" {
		return "user-1", nil
	}
	return "", errors.New("bad token")
}

func TestBearerToken(t *testing.T) {
	cases := []struct {
		header string
		want   string
	}{
		{"Bearer valid", "valid"},
		{"bearer valid", "valid"},
		{"Token valid", ""},
		{"", ""},
		{"Bearer", ""},
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if c.header != "" {
			req.Header.Set("Authorization", c.header)
		}
		if got := bearerToken(req); got != c.want {
			t.Errorf("bearerToken(%q) = %q, want %q", c.header, got, c.want)
		}
	}
}

func TestAuthRequiredAllowsValidToken(t *testing.T) {
	called := false
	h := AuthRequired(okParser, slog.Default())(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		called = true
		if UserIDFrom(r.Context()) != "user-1" {
			t.Error("user id not in context")
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid")
	h.ServeHTTP(httptest.NewRecorder(), req)
	if !called {
		t.Error("next handler not called")
	}
}

func TestAuthRequiredRejectsMissing(t *testing.T) {
	h := AuthRequired(okParser, slog.Default())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("next should not be called")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAuthRequiredRejectsInvalid(t *testing.T) {
	h := AuthRequired(okParser, slog.Default())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("next should not be called")
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer nope")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestUserIDFromEmptyContext(t *testing.T) {
	if UserIDFrom(context.Background()) != "" {
		t.Error("expected empty user id from background context")
	}
}
