// internal/petcare/adapters/secondary/sqlite/dose_repository.go
package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
)

type DoseRepository struct{ db dbtx }

func NewDoseRepository(db dbtx) *DoseRepository {
	return &DoseRepository{db: db}
}

func (r *DoseRepository) CreateBatch(ctx context.Context, doses []domain.Dose) error {
	if len(doses) == 0 {
		return nil
	}
	stmt, err := r.db.PrepareContext(ctx, `INSERT INTO doses (id, treatment_id, pet_id, scheduled_for, google_calendar_event_id) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close() //nolint:errcheck
	for _, d := range doses {
		if _, err := stmt.ExecContext(ctx, d.ID, d.TreatmentID, d.PetID, d.ScheduledFor.Format(time.RFC3339), d.GoogleCalendarEventID); err != nil {
			return err
		}
	}
	return nil
}

func (r *DoseRepository) ListByTreatment(ctx context.Context, treatmentID string) ([]domain.Dose, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, treatment_id, pet_id, scheduled_for, google_calendar_event_id FROM doses WHERE treatment_id = ? ORDER BY scheduled_for ASC`,
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

func (r *DoseRepository) ListFutureByTreatment(ctx context.Context, treatmentID string, after time.Time) ([]domain.Dose, error) {
	afterStr := after.Format(time.RFC3339)
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, treatment_id, pet_id, scheduled_for, google_calendar_event_id FROM doses WHERE treatment_id = ? AND scheduled_for > ? ORDER BY scheduled_for ASC`,
		treatmentID, afterStr)
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

func (r *DoseRepository) DeleteFutureByTreatment(ctx context.Context, treatmentID string, after time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM doses WHERE treatment_id = ? AND scheduled_for > ?`,
		treatmentID, after.Format(time.RFC3339))
	return err
}

func scanDose(s scanner) (*domain.Dose, error) {
	var d domain.Dose
	var scheduledFor string
	if err := s.Scan(&d.ID, &d.TreatmentID, &d.PetID, &scheduledFor, &d.GoogleCalendarEventID); err != nil {
		return nil, err
	}
	var err error
	d.ScheduledFor, err = time.Parse(time.RFC3339, scheduledFor)
	if err != nil {
		return nil, fmt.Errorf("parse scheduled_for %q: %w", scheduledFor, err)
	}
	return &d, nil
}
