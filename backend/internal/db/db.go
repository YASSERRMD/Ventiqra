// Package db manages the PostgreSQL connection pool and related helpers.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig tunes the connection pool used by Connect.
type PoolConfig struct {
	MinConns        int32
	MaxConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	ConnectTimeout  time.Duration
}

// DefaultPoolConfig returns production-reasonable defaults.
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MinConns:        2,
		MaxConns:        10,
		MaxConnLifetime: 30 * time.Minute,
		MaxConnIdleTime: 5 * time.Minute,
		ConnectTimeout:  10 * time.Second,
	}
}

// Connect builds and validates a pgxpool connection pool for the given DSN.
// The pool is pinged before being returned so callers fail fast on bad config.
func Connect(ctx context.Context, dsn string, opts PoolConfig) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	cfg.MinConns = opts.MinConns
	cfg.MaxConns = opts.MaxConns
	cfg.MaxConnLifetime = opts.MaxConnLifetime
	cfg.MaxConnIdleTime = opts.MaxConnIdleTime
	if opts.ConnectTimeout > 0 {
		cfg.ConnConfig.ConnectTimeout = opts.ConnectTimeout
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create db pool: %w", err)
	}

	if err := Ping(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

// Ping verifies the pool can reach the database.
func Ping(ctx context.Context, pool *pgxpool.Pool) error {
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}
	return nil
}

// Close safely closes the pool, tolerating nil.
func Close(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
}
