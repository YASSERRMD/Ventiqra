// Package config loads Ventiqra backend configuration from environment
// variables and optional .env files.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// Config holds all runtime configuration for the API service.
type Config struct {
	Env         string
	HTTP        HTTPConfig
	Log         LogConfig
	CORSOrigins []string

	Postgres PostgresConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

// HTTPConfig holds the HTTP server settings.
type HTTPConfig struct {
	Host string
	Port string
}

// LogConfig holds logger settings.
type LogConfig struct {
	Level  string
	Format string // "json" or "text"
}

// PostgresConfig holds PostgreSQL connection settings.
type PostgresConfig struct {
	Host        string
	Port        string
	Database    string
	User        string
	Password    string
	DatabaseURL string
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host string
	Port string
	URL  string
}

// JWTConfig holds authentication settings.
type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// Load reads configuration from the environment. If envFile is non-empty and
// the file exists, its values are loaded first (without overriding variables
// already present in the real environment).
func Load(envFile string) (Config, error) {
	if envFile != "" {
		if err := loadDotEnv(envFile); err != nil {
			return Config{}, fmt.Errorf("load env file %q: %w", envFile, err)
		}
	}

	cfg := Config{
		Env: getenv("ENV", "development"),
		HTTP: HTTPConfig{
			Host: getenv("BACKEND_HOST", "0.0.0.0"),
			Port: getenv("BACKEND_PORT", "8080"),
		},
		Log: LogConfig{
			Level:  getenv("BACKEND_LOG_LEVEL", "info"),
			Format: getenv("BACKEND_LOG_FORMAT", "json"),
		},
		CORSOrigins: splitList(getenv("BACKEND_CORS_ORIGINS", "http://localhost:3000")),
		Postgres: PostgresConfig{
			Host:        getenv("POSTGRES_HOST", "localhost"),
			Port:        getenv("POSTGRES_PORT", "5432"),
			Database:    getenv("POSTGRES_DB", "ventiqra"),
			User:        getenv("POSTGRES_USER", "ventiqra"),
			Password:    getenv("POSTGRES_PASSWORD", ""),
			DatabaseURL: getenv("DATABASE_URL", ""),
		},
		Redis: RedisConfig{
			Host: getenv("REDIS_HOST", "localhost"),
			Port: getenv("REDIS_PORT", "6379"),
			URL:  getenv("REDIS_URL", ""),
		},
		JWT: JWTConfig{
			Secret:     getenv("JWT_SECRET", ""),
			AccessTTL:  getduration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL: getduration("JWT_REFRESH_TTL", 168*time.Hour),
		},
	}

	return cfg, nil
}

// IsProduction reports whether the service runs in production mode.
func (c Config) IsProduction() bool { return c.Env == "production" }

// DSN returns the PostgreSQL connection string, preferring DATABASE_URL and
// falling back to the assembled host/port/db/user/password values.
func (c Config) DSN() string {
	if c.Postgres.DatabaseURL != "" {
		return c.Postgres.DatabaseURL
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.Postgres.User, c.Postgres.Password, c.Postgres.Host, c.Postgres.Port, c.Postgres.Database,
	)
}

// loadDotEnv parses a simple KEY=VALUE .env file and sets values for any keys
// that are not already defined in the environment. Existing environment
// variables always take precedence.
func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
	return scanner.Err()
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getduration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func splitList(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
