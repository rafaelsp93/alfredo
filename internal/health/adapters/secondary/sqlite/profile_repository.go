package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rafaelsoares/alfredo/internal/health/domain"
)

type ProfileRepository struct {
	db dbtx
}

func NewProfileRepository(db dbtx) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) Get(ctx context.Context) (domain.HealthProfile, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, height_cm, birth_date, sex, created_at, updated_at
		FROM health_profiles
		WHERE id = 1
	`)
	profile, err := scanProfile(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.HealthProfile{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.HealthProfile{}, fmt.Errorf("select health profile: %w", err)
	}
	return profile, nil
}

func (r *ProfileRepository) Upsert(ctx context.Context, profile domain.HealthProfile) (domain.HealthProfile, error) {
	now := time.Now().UTC()
	profile.ID = 1
	createdAt := now
	var createdAtRaw string
	if err := r.db.QueryRowContext(ctx, `SELECT created_at FROM health_profiles WHERE id = 1`).Scan(&createdAtRaw); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return domain.HealthProfile{}, fmt.Errorf("load existing health profile: %w", err)
		}
	} else {
		existingCreatedAt, err := time.Parse(time.RFC3339Nano, createdAtRaw)
		if err != nil {
			return domain.HealthProfile{}, fmt.Errorf("parse created_at %q: %w", createdAtRaw, err)
		}
		createdAt = existingCreatedAt
	}
	profile.CreatedAt = createdAt
	profile.UpdatedAt = now
	_, err := r.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO health_profiles (id, height_cm, birth_date, sex, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		profile.ID,
		profile.HeightCM,
		profile.BirthDate,
		profile.Sex,
		profile.CreatedAt.Format(time.RFC3339Nano),
		profile.UpdatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return domain.HealthProfile{}, fmt.Errorf("upsert health profile: %w", err)
	}
	return profile, nil
}

func scanProfile(s scanner) (domain.HealthProfile, error) {
	var profile domain.HealthProfile
	var createdAt string
	var updatedAt string
	if err := s.Scan(&profile.ID, &profile.HeightCM, &profile.BirthDate, &profile.Sex, &createdAt, &updatedAt); err != nil {
		return domain.HealthProfile{}, err
	}
	created, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return domain.HealthProfile{}, fmt.Errorf("parse created_at %q: %w", createdAt, err)
	}
	updated, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return domain.HealthProfile{}, fmt.Errorf("parse updated_at %q: %w", updatedAt, err)
	}
	profile.CreatedAt = created
	profile.UpdatedAt = updated
	return profile, nil
}

type scanner interface {
	Scan(dest ...any) error
}
