package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})
}

func TestChainOrder(t *testing.T) {
	var calls []string
	mk := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls = append(calls, "pre-"+name)
				next.ServeHTTP(w, r)
				calls = append(calls, "post-"+name)
			})
		}
	}

	h := Chain(okHandler(), mk("a"), mk("b"), mk("c"))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	want := []string{"pre-a", "pre-b", "pre-c", "post-c", "post-b", "post-a"}
	for i, w := range want {
		if i >= len(calls) || calls[i] != w {
			t.Fatalf("call order = %v, want %v", calls, want)
		}
	}
}

func TestRequestIDGeneratedAndEchoed(t *testing.T) {
	rec := httptest.NewRecorder()
	RequestID(okHandler()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if id := rec.Header().Get("X-Request-Id"); id == "" {
		t.Error("expected generated X-Request-Id header")
	}
}

func TestRequestIDPreservedFromClient(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Id", "client-id")
	RequestID(okHandler()).ServeHTTP(rec, req)
	if id := rec.Header().Get("X-Request-Id"); id != "client-id" {
		t.Errorf("X-Request-Id = %q, want client-id", id)
	}
}

func TestRecoverCatchesPanic(t *testing.T) {
	panicHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	})
	rec := httptest.NewRecorder()
	Recover(slog.Default())(panicHandler).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestCORSAllowedOrigin(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://allowed.test")
	CORS([]string{"http://allowed.test"})(okHandler()).ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://allowed.test" {
		t.Errorf("allow-origin = %q, want http://allowed.test", got)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary = %q, want Origin", got)
	}
}

func TestCORSDisallowedOrigin(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://evil.test")
	CORS([]string{"http://allowed.test"})(okHandler()).ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("allow-origin = %q, want empty", got)
	}
}

func TestCORSWildcard(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://anything.test")
	CORS([]string{"*"})(okHandler()).ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://anything.test" {
		t.Errorf("allow-origin = %q, want echoed origin", got)
	}
}

func TestCORSPreflightReturnsNoContent(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://allowed.test")
	CORS([]string{"http://allowed.test"})(okHandler()).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("preflight status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestLoggerDoesNotBreakRequest(t *testing.T) {
	rec := httptest.NewRecorder()
	Logger(slog.Default())(okHandler()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
