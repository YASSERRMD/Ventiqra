package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrConflict is returned when an insert/update violates a unique constraint.
var ErrConflict = errors.New("repository: conflict")

// User is the application-level user model.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserRepo provides persistence for users.
type UserRepo struct {
	*Repository
}

// NewUserRepo creates a UserRepo over the shared pool.
func NewUserRepo(base *Repository) *UserRepo {
	return &UserRepo{Repository: base}
}

// CreateUser inserts a new user and returns the persisted record.
func (r *UserRepo) CreateUser(ctx context.Context, email, passwordHash, name string) (*User, error) {
	const q = `INSERT INTO users (email, password_hash, name)
	           VALUES ($1, $2, $3)
	           RETURNING id, email, password_hash, name, created_at, updated_at`

	var u User
	row := r.pool.QueryRow(ctx, q, email, passwordHash, name)
	if err := scanUser(row, &u); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &u, nil
}

// GetUserByEmail returns the user with the given email or ErrNotFound.
func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	const q = `SELECT id, email, password_hash, name, created_at, updated_at
	           FROM users WHERE email = $1`

	var u User
	if err := scanUser(r.pool.QueryRow(ctx, q, email), &u); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

// GetUserByID returns the user with the given id or ErrNotFound.
func (r *UserRepo) GetUserByID(ctx context.Context, id string) (*User, error) {
	const q = `SELECT id, email, password_hash, name, created_at, updated_at
	           FROM users WHERE id = $1`

	var u User
	if err := scanUser(r.pool.QueryRow(ctx, q, id), &u); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner, u *User) error {
	return row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt, &u.UpdatedAt)
}
