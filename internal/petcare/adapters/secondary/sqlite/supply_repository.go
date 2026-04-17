package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
)

type SupplyRepository struct{ db dbtx }

func NewSupplyRepository(db dbtx) *SupplyRepository {
	return &SupplyRepository{db: db}
}

func (r *SupplyRepository) Create(ctx context.Context, supply domain.Supply) (*domain.Supply, error) {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO supplies (id, pet_id, name, last_purchased_at, estimated_days_supply, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		supply.ID,
		supply.PetID,
		supply.Name,
		formatSupplyDate(supply.LastPurchasedAt),
		supply.EstimatedDaysSupply,
		supply.Notes,
		supply.CreatedAt.Format(time.RFC3339),
		supply.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("insert supply %q for pet %q: %w", supply.ID, supply.PetID, err)
	}
	return &supply, nil
}

func (r *SupplyRepository) GetByID(ctx context.Context, petID, supplyID string) (*domain.Supply, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, pet_id, name, last_purchased_at, estimated_days_supply, notes, created_at, updated_at
		FROM supplies
		WHERE id = ? AND pet_id = ?`, supplyID, petID)
	supply, err := scanSupply(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select supply %q for pet %q: %w", supplyID, petID, err)
	}
	return supply, nil
}

func (r *SupplyRepository) List(ctx context.Context, petID string) ([]domain.Supply, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, pet_id, name, last_purchased_at, estimated_days_supply, notes, created_at, updated_at
		FROM supplies
		WHERE pet_id = ?
		ORDER BY date(last_purchased_at, '+' || estimated_days_supply || ' days') ASC, name ASC`, petID)
	if err != nil {
		return nil, fmt.Errorf("query supplies for pet %q: %w", petID, err)
	}
	defer rows.Close() //nolint:errcheck

	supplies := make([]domain.Supply, 0)
	for rows.Next() {
		supply, err := scanSupply(rows)
		if err != nil {
			return nil, fmt.Errorf("scan supply for pet %q: %w", petID, err)
		}
		supplies = append(supplies, *supply)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate supplies for pet %q: %w", petID, err)
	}
	return supplies, nil
}

func (r *SupplyRepository) Update(ctx context.Context, supply domain.Supply) (*domain.Supply, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE supplies
		SET name = ?, last_purchased_at = ?, estimated_days_supply = ?, notes = ?, updated_at = ?
		WHERE id = ? AND pet_id = ?`,
		supply.Name,
		formatSupplyDate(supply.LastPurchasedAt),
		supply.EstimatedDaysSupply,
		supply.Notes,
		supply.UpdatedAt.Format(time.RFC3339),
		supply.ID,
		supply.PetID,
	)
	if err != nil {
		return nil, fmt.Errorf("update supply %q for pet %q: %w", supply.ID, supply.PetID, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, domain.ErrNotFound
	}
	return &supply, nil
}

func (r *SupplyRepository) Delete(ctx context.Context, petID, supplyID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM supplies WHERE id = ? AND pet_id = ?`, supplyID, petID)
	if err != nil {
		return fmt.Errorf("delete supply %q for pet %q: %w", supplyID, petID, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func scanSupply(s scanner) (*domain.Supply, error) {
	var supply domain.Supply
	var lastPurchasedAt string
	var createdAt string
	var updatedAt string
	var notes sql.NullString

	err := s.Scan(
		&supply.ID,
		&supply.PetID,
		&supply.Name,
		&lastPurchasedAt,
		&supply.EstimatedDaysSupply,
		&notes,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}
	supply.LastPurchasedAt, err = time.Parse("2006-01-02", lastPurchasedAt)
	if err != nil {
		return nil, fmt.Errorf("parse last_purchased_at %q: %w", lastPurchasedAt, err)
	}
	supply.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at %q: %w", createdAt, err)
	}
	supply.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at %q: %w", updatedAt, err)
	}
	if notes.Valid {
		supply.Notes = &notes.String
	}
	return &supply, nil
}

func formatSupplyDate(t time.Time) string {
	return t.Format("2006-01-02")
}
