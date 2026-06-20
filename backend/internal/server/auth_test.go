package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"log/slog"

	"github.com/YASSERRMD/Ventiqra/backend/internal/auth"
	"github.com/YASSERRMD/Ventiqra/backend/internal/config"
	"github.com/YASSERRMD/Ventiqra/backend/internal/db"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
	"github.com/YASSERRMD/Ventiqra/backend/internal/testutil"
)

func newAuthTestServer(t *testing.T) *Server {
	t.Helper()
	t.Cleanup(testutil.LockDB())
	dsn := os.Getenv("DATABASE_TEST_URL")
	if dsn == "" {
		dsn = "postgres://ventiqra:changeme@localhost:5432/ventiqra?sslmode=disable"
	}
	pool, err := db.Connect(context.Background(), dsn, db.DefaultPoolConfig())
	if err != nil {
		t.Skipf("database not available, skipping: %v", err)
	}
	t.Cleanup(func() { db.Close(pool) })

	ctx := context.Background()
	if _, err := pool.Exec(ctx, `DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;`); err != nil {
		t.Fatalf("reset schema: %v", err)
	}
	if _, err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	base := repository.New(pool)
	tm, err := auth.NewTokenManager("test-secret", 0)
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	cfg := config.Config{Env: "test"}
	return New(cfg, slog.Default(),
		WithDB(pool),
		WithAuth(repository.NewUserRepo(base), tm),
		WithCompany(repository.NewCompanyRepo(base)),
		WithSim(repository.NewSimStateRepo(base)),
		WithProducts(repository.NewProductRepo(base)),
	)
}

func doJSON(t *testing.T, srv *Server, method, path string, body any, auth string) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Set("Authorization", "Bearer "+auth)
	}
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	return rec
}

func registerAndLogin(t *testing.T, srv *Server) (token string) {
	t.Helper()
	rec := doJSON(t, srv, "POST", "/api/v1/auth/register",
		map[string]string{"email": "carol@example.com", "password": "password123", "name": "Carol"}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("register status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var res authResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode register: %v", err)
	}
	return res.Token
}

func TestRegisterThenMe(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	if token == "" {
		t.Fatal("empty token")
	}

	rec := doJSON(t, srv, "GET", "/api/v1/me", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("me status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var me struct {
		User publicUser `json:"user"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&me); err != nil {
		t.Fatalf("decode me: %v", err)
	}
	if me.User.Email != "carol@example.com" {
		t.Errorf("email = %q, want carol@example.com", me.User.Email)
	}
}

func TestRegisterDuplicateIs409(t *testing.T) {
	srv := newAuthTestServer(t)
	body := map[string]string{"email": "dup@example.com", "password": "password123", "name": "Dup"}
	if rec := doJSON(t, srv, "POST", "/api/v1/auth/register", body, ""); rec.Code != http.StatusOK {
		t.Fatalf("first register = %d", rec.Code)
	}
	rec := doJSON(t, srv, "POST", "/api/v1/auth/register", body, "")
	if rec.Code != http.StatusConflict {
		t.Errorf("duplicate status = %d, want 409", rec.Code)
	}
}

func TestRegisterValidation(t *testing.T) {
	srv := newAuthTestServer(t)
	cases := []struct {
		name string
		body map[string]string
	}{
		{"bad email", map[string]string{"email": "nope", "password": "password123", "name": "X"}},
		{"short password", map[string]string{"email": "x@y.com", "password": "short", "name": "X"}},
		{"missing name", map[string]string{"email": "x@y.com", "password": "password123", "name": ""}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := doJSON(t, srv, "POST", "/api/v1/auth/register", c.body, "")
			if rec.Code != http.StatusBadRequest {
				t.Errorf("%s: status = %d, want 400", c.name, rec.Code)
			}
		})
	}
}

func TestLoginSuccess(t *testing.T) {
	srv := newAuthTestServer(t)
	registerAndLogin(t, srv) // seeds carol via register

	rec := doJSON(t, srv, "POST", "/api/v1/auth/login",
		map[string]string{"email": "carol@example.com", "password": "password123"}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"token"`) {
		t.Errorf("login body missing token: %s", rec.Body.String())
	}
}

func TestLoginWrongPassword(t *testing.T) {
	srv := newAuthTestServer(t)
	registerAndLogin(t, srv)

	rec := doJSON(t, srv, "POST", "/api/v1/auth/login",
		map[string]string{"email": "carol@example.com", "password": "wrongpassword"}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestMeUnauthorized(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "GET", "/api/v1/me", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestMeRejectsBadToken(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "GET", "/api/v1/me", nil, "garbage")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
