// internal/petcare/adapters/secondary/sqlite/dose_repository.go
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
)

type DoseRepository struct{ db *sql.DB }

func NewDoseRepository(db *sql.DB) *DoseRepository {
	return &DoseRepository{db: db}
}

func (r *DoseRepository) CreateBatch(ctx context.Context, doses []domain.Dose) error {
	if len(doses) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO doses (id, treatment_id, pet_id, scheduled_for) VALUES (?, ?, ?, ?)`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close() //nolint:errcheck
	for _, d := range doses {
		if _, err := stmt.ExecContext(ctx, d.ID, d.TreatmentID, d.PetID, d.ScheduledFor.Format(time.RFC3339)); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (r *DoseRepository) ListByTreatment(ctx context.Context, treatmentID string) ([]domain.Dose, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, treatment_id, pet_id, scheduled_for FROM doses WHERE treatment_id = ? ORDER BY scheduled_for ASC`,
		treatmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	var doses []domain.Dose
	for rows.Next() {
		d, err := scanDose(rows)
		if err != nil {
			return nil, err
		}
		doses = append(doses, *d)
	}
	return doses, rows.Err()
}

func (r *DoseRepository) DeleteFutureDoses(ctx context.Context, treatmentID string, after time.Time) ([]string, error) {
	afterStr := after.Format(time.RFC3339)
	rows, err := r.db.QueryContext(ctx,
		`SELECT id FROM doses WHERE treatment_id = ? AND scheduled_for > ?`,
		treatmentID, afterStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	if _, err := r.db.ExecContext(ctx,
		`DELETE FROM doses WHERE treatment_id = ? AND scheduled_for > ?`,
		treatmentID, afterStr); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *DoseRepository) ListOpenEndedActiveTreatments(ctx context.Context) ([]domain.Treatment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, pet_id, name, dosage_amount, dosage_unit, route, interval_hours, started_at, ended_at, stopped_at, vet_name, notes, created_at
		 FROM treatments WHERE ended_at IS NULL AND stopped_at IS NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	var ts []domain.Treatment
	for rows.Next() {
		t, err := scanTreatment(rows)
		if err != nil {
			return nil, err
		}
		ts = append(ts, *t)
	}
	return ts, rows.Err()
}

func (r *DoseRepository) LatestDoseFor(ctx context.Context, treatmentID string) (*domain.Dose, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, treatment_id, pet_id, scheduled_for FROM doses WHERE treatment_id = ? ORDER BY scheduled_for DESC LIMIT 1`,
		treatmentID)
	d, err := scanDose(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return d, nil
}

func scanDose(s scanner) (*domain.Dose, error) {
	var d domain.Dose
	var scheduledFor string
	if err := s.Scan(&d.ID, &d.TreatmentID, &d.PetID, &scheduledFor); err != nil {
		return nil, err
	}
	var err error
	d.ScheduledFor, err = time.Parse(time.RFC3339, scheduledFor)
	if err != nil {
		return nil, fmt.Errorf("parse scheduled_for %q: %w", scheduledFor, err)
	}
	return &d, nil
}
