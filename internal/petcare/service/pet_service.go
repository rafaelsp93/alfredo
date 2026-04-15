package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/port"
)

type CreatePetInput struct {
	Name             string
	Species          string
	Breed            *string
	BirthDate        *time.Time
	WeightKg         *float64
	DailyFoodGrams   *float64
	PhotoPath        *string
	GoogleCalendarID string
}

type UpdatePetInput struct {
	Name           string
	Species        string
	Breed          *string
	BirthDate      *time.Time
	WeightKg       *float64
	DailyFoodGrams *float64
	PhotoPath      *string
}

type PetService struct {
	repo port.PetRepository
}

func NewPetService(repo port.PetRepository) *PetService {
	return &PetService{repo: repo}
}

func (s *PetService) List(ctx context.Context) ([]domain.Pet, error) {
	return s.repo.List(ctx)
}

func (s *PetService) Create(ctx context.Context, in CreatePetInput) (*domain.Pet, error) {
	if in.Name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrValidation)
	}
	if in.Species == "" {
		return nil, fmt.Errorf("%w: species is required", domain.ErrValidation)
	}
	pet := domain.Pet{
		ID:               uuid.New().String(),
		Name:             in.Name,
		Species:          in.Species,
		Breed:            in.Breed,
		BirthDate:        in.BirthDate,
		WeightKg:         in.WeightKg,
		DailyFoodGrams:   in.DailyFoodGrams,
		PhotoPath:        in.PhotoPath,
		GoogleCalendarID: in.GoogleCalendarID,
		CreatedAt:        time.Now().UTC(),
	}
	created, err := s.repo.Create(ctx, pet)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *PetService) GetByID(ctx context.Context, id string) (*domain.Pet, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PetService) Update(ctx context.Context, id string, in UpdatePetInput) (*domain.Pet, error) {
	if in.Name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrValidation)
	}
	if in.Species == "" {
		return nil, fmt.Errorf("%w: species is required", domain.ErrValidation)
	}
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	existing.Name = in.Name
	existing.Species = in.Species
	existing.Breed = in.Breed
	existing.BirthDate = in.BirthDate
	existing.WeightKg = in.WeightKg
	existing.DailyFoodGrams = in.DailyFoodGrams
	existing.PhotoPath = in.PhotoPath
	updated, err := s.repo.Update(ctx, *existing)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *PetService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
