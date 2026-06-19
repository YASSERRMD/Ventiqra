package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	for _, k := range []string{"ENV", "BACKEND_PORT", "BACKEND_LOG_LEVEL", "BACKEND_CORS_ORIGINS"} {
		t.Setenv(k, "")
	}
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HTTP.Port != "8080" {
		t.Errorf("default port = %q, want 8080", cfg.HTTP.Port)
	}
	if cfg.Env != "development" {
		t.Errorf("default env = %q, want development", cfg.Env)
	}
	if cfg.IsProduction() {
		t.Error("expected IsProduction() false by default")
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("BACKEND_PORT", "9090")
	t.Setenv("ENV", "production")
	t.Setenv("BACKEND_CORS_ORIGINS", "http://a.test, http://b.test")
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HTTP.Port != "9090" {
		t.Errorf("port = %q, want 9090", cfg.HTTP.Port)
	}
	if !cfg.IsProduction() {
		t.Error("expected IsProduction() true")
	}
	if len(cfg.CORSOrigins) != 2 {
		t.Errorf("cors origins len = %d, want 2", len(cfg.CORSOrigins))
	}
}

func TestLoadDotEnvDoesNotOverrideRealEnv(t *testing.T) {
	t.Setenv("BACKEND_PORT", "7777")
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := []byte("BACKEND_PORT=1111\nENV=production\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HTTP.Port != "7777" {
		t.Errorf("port = %q, want 7777 (real env must win)", cfg.HTTP.Port)
	}
	if cfg.Env != "production" {
		t.Errorf("env = %q, want production", cfg.Env)
	}
}

func TestDSNFallsBackToComponents(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("POSTGRES_USER", "u")
	t.Setenv("POSTGRES_PASSWORD", "p")
	t.Setenv("POSTGRES_HOST", "h")
	t.Setenv("POSTGRES_PORT", "5433")
	t.Setenv("POSTGRES_DB", "d")
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := "postgres://u:p@h:5433/d?sslmode=disable"
	if got := cfg.DSN(); got != want {
		t.Errorf("DSN() = %q, want %q", got, want)
	}
}

func TestDSNPrefersURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x:y@db:5432/zz?sslmode=require")
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DSN() != "postgres://x:y@db:5432/zz?sslmode=require" {
		t.Errorf("DSN() = %q, want explicit DATABASE_URL", cfg.DSN())
	}
}
