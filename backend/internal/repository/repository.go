// Package repository implements the data-access (repository) pattern used
// across Ventiqra. A base Repository exposes the connection pool; concrete
// repositories embed it and add entity-specific queries.
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when a query expects exactly one row but finds none.
var ErrNotFound = errors.New("repository: not found")

// Repository is the base repository holding the shared connection pool.
// Concrete repositories embed it to gain access to the pool.
type Repository struct {
	pool *pgxpool.Pool
}

// New creates a base Repository bound to the given pool.
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Pool returns the underlying connection pool.
func (r *Repository) Pool() *pgxpool.Pool { return r.pool }

// TxRunner lets a repository method participate in a caller-controlled
// transaction via the db.WithTx helper.
type TxRunner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Tx returns a transaction-capable handle (the pool itself satisfies this).
func (r *Repository) Tx() TxRunner { return r.pool }

// IsNotFound reports whether err is ErrNotFound.
func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }
