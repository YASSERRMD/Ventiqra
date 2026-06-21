package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// SaveSlot is a named, restorable snapshot of an owner's simulation run.
type SaveSlot struct {
	ID         string
	OwnerID    string
	Slot       string
	Label      string
	CompanyID  string
	Day        int
	CashCents  int64
	Status     string
	Snapshot   []byte // raw JSONB
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// SaveSlotRepo provides persistence for simulation save slots.
type SaveSlotRepo struct {
	*Repository
}

// NewSaveSlotRepo creates a SaveSlotRepo over the shared pool.
func NewSaveSlotRepo(base *Repository) *SaveSlotRepo {
	return &SaveSlotRepo{Repository: base}
}

// Upsert creates or replaces a slot for the owner, returning it.
func (r *SaveSlotRepo) Upsert(ctx context.Context, s *SaveSlot) (*SaveSlot, error) {
	const q = `INSERT INTO save_slots (owner_id, slot, label, company_id, day, cash_cents, status, snapshot)
	           VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	           ON CONFLICT (owner_id, slot) DO UPDATE
	             SET label = EXCLUDED.label, company_id = EXCLUDED.company_id,
	                 day = EXCLUDED.day, cash_cents = EXCLUDED.cash_cents,
	                 status = EXCLUDED.status, snapshot = EXCLUDED.snapshot, updated_at = NOW()
	           RETURNING id, owner_id, slot, label, company_id, day, cash_cents, status, snapshot, created_at, updated_at`
	return scanSaveSlot(r.pool.QueryRow(ctx, q,
		s.OwnerID, s.Slot, s.Label, s.CompanyID, s.Day, s.CashCents, s.Status, s.Snapshot))
}

// Get returns a slot by owner+slot name, or ErrNotFound.
func (r *SaveSlotRepo) Get(ctx context.Context, ownerID, slot string) (*SaveSlot, error) {
	const q = `SELECT id, owner_id, slot, label, company_id, day, cash_cents, status, snapshot, created_at, updated_at
	           FROM save_slots WHERE owner_id = $1 AND slot = $2`
	s, err := scanSaveSlot(r.pool.QueryRow(ctx, q, ownerID, slot))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get save slot: %w", err)
	}
	return s, nil
}

// ListByOwner returns an owner's save slots, newest first.
func (r *SaveSlotRepo) ListByOwner(ctx context.Context, ownerID string) ([]*SaveSlot, error) {
	const q = `SELECT id, owner_id, slot, label, company_id, day, cash_cents, status, snapshot, created_at, updated_at
	           FROM save_slots WHERE owner_id = $1 ORDER BY updated_at DESC`
	rows, err := r.pool.Query(ctx, q, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list save slots: %w", err)
	}
	defer rows.Close()
	var out []*SaveSlot
	for rows.Next() {
		s, err := scanSaveSlot(rows)
		if err != nil {
			return nil, fmt.Errorf("scan save slot: %w", err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("save slots rows: %w", err)
	}
	return out, nil
}

// Delete removes an owner's slot.
func (r *SaveSlotRepo) Delete(ctx context.Context, ownerID, slot string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM save_slots WHERE owner_id = $1 AND slot = $2`, ownerID, slot)
	if err != nil {
		return fmt.Errorf("delete save slot: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// SnapshotJSON is a convenience accessor that returns the parsed snapshot.
func (s *SaveSlot) SnapshotJSON() map[string]any {
	if len(s.Snapshot) == 0 {
		return nil
	}
	var m map[string]any
	_ = json.Unmarshal(s.Snapshot, &m)
	return m
}

type saveSlotScanner interface {
	Scan(dest ...any) error
}

func scanSaveSlot(row saveSlotScanner) (*SaveSlot, error) {
	var s SaveSlot
	if err := row.Scan(
		&s.ID, &s.OwnerID, &s.Slot, &s.Label, &s.CompanyID, &s.Day, &s.CashCents,
		&s.Status, &s.Snapshot, &s.CreatedAt, &s.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &s, nil
}
