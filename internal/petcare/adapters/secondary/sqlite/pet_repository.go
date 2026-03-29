package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
)

type PetRepository struct{ db *sql.DB }

func NewPetRepository(db *sql.DB) *PetRepository { return &PetRepository{db: db} }

func (r *PetRepository) Create(ctx context.Context, p domain.Pet) (*domain.Pet, error) {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO pets (id, name, species, breed, birth_date, weight_kg, daily_food_grams, photo_path, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.Species,
		p.Breed, formatDate(p.BirthDate), p.WeightKg, p.DailyFoodGrams, p.PhotoPath,
		p.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PetRepository) GetByID(ctx context.Context, id string) (*domain.Pet, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, species, breed, birth_date, weight_kg, daily_food_grams, photo_path, created_at FROM pets WHERE id = ?`, id)
	p, err := scanPet(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return p, err
}

func (r *PetRepository) List(ctx context.Context) ([]domain.Pet, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, species, breed, birth_date, weight_kg, daily_food_grams, photo_path, created_at FROM pets ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	pets := make([]domain.Pet, 0)
	for rows.Next() {
		p, err := scanPet(rows)
		if err != nil {
			return nil, err
		}
		pets = append(pets, *p)
	}
	return pets, rows.Err()
}

func (r *PetRepository) Update(ctx context.Context, p domain.Pet) (*domain.Pet, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE pets SET name=?, species=?, breed=?, birth_date=?, weight_kg=?, daily_food_grams=?, photo_path=?
		WHERE id=?`,
		p.Name, p.Species, p.Breed, formatDate(p.BirthDate), p.WeightKg, p.DailyFoodGrams, p.PhotoPath, p.ID,
	)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, domain.ErrNotFound
	}
	return &p, nil
}

func (r *PetRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM pets WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// --- helpers ---

type scanner interface {
	Scan(dest ...any) error
}

func scanPet(s scanner) (*domain.Pet, error) {
	var p domain.Pet
	var birthDate sql.NullString
	var createdAt string
	err := s.Scan(&p.ID, &p.Name, &p.Species, &p.Breed, &birthDate, &p.WeightKg, &p.DailyFoodGrams, &p.PhotoPath, &createdAt)
	if err != nil {
		return nil, err
	}
	p.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at %q: %w", createdAt, err)
	}
	if birthDate.Valid && birthDate.String != "" {
		t, err := time.Parse("2006-01-02", birthDate.String)
		if err != nil {
			return nil, fmt.Errorf("parse birth_date %q: %w", birthDate.String, err)
		}
		p.BirthDate = &t
	}
	return &p, nil
}

func formatDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}
