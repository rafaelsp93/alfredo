package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/port"
)

type CreateSupplyInput struct {
	PetID               string
	Name                string
	LastPurchasedAt     time.Time
	EstimatedDaysSupply int
	Notes               *string
}

type UpdateSupplyInput struct {
	Name                *string
	LastPurchasedAt     *time.Time
	EstimatedDaysSupply *int
	Notes               *string
}

type SupplyService struct {
	repo port.SupplyRepository
}

func NewSupplyService(repo port.SupplyRepository) *SupplyService {
	return &SupplyService{repo: repo}
}

func (s *SupplyService) Create(ctx context.Context, in CreateSupplyInput) (*domain.Supply, error) {
	if in.PetID == "" {
		return nil, fmt.Errorf("%w: pet_id is required", domain.ErrValidation)
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrValidation)
	}
	if in.LastPurchasedAt.IsZero() {
		return nil, fmt.Errorf("%w: last_purchased_at is required", domain.ErrValidation)
	}
	if in.EstimatedDaysSupply <= 0 {
		return nil, fmt.Errorf("%w: estimated_days_supply must be greater than zero", domain.ErrValidation)
	}
	now := time.Now().UTC()
	supply := domain.Supply{
		ID:                  uuid.New().String(),
		PetID:               in.PetID,
		Name:                strings.TrimSpace(in.Name),
		LastPurchasedAt:     dateOnly(in.LastPurchasedAt),
		EstimatedDaysSupply: in.EstimatedDaysSupply,
		Notes:               in.Notes,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	created, err := s.repo.Create(ctx, supply)
	if err != nil {
		return nil, fmt.Errorf("create supply: %w", err)
	}
	return created, nil
}

func (s *SupplyService) GetByID(ctx context.Context, petID, supplyID string) (*domain.Supply, error) {
	supply, err := s.repo.GetByID(ctx, petID, supplyID)
	if err != nil {
		return nil, fmt.Errorf("get supply %q for pet %q: %w", supplyID, petID, err)
	}
	return supply, nil
}

func (s *SupplyService) List(ctx context.Context, petID string) ([]domain.Supply, error) {
	supplies, err := s.repo.List(ctx, petID)
	if err != nil {
		return nil, fmt.Errorf("list supplies for pet %q: %w", petID, err)
	}
	return supplies, nil
}

func (s *SupplyService) Update(ctx context.Context, petID, supplyID string, in UpdateSupplyInput) (*domain.Supply, error) {
	existing, err := s.GetByID(ctx, petID, supplyID)
	if err != nil {
		return nil, fmt.Errorf("update supply: %w", err)
	}
	if in.Name != nil {
		if strings.TrimSpace(*in.Name) == "" {
			return nil, fmt.Errorf("%w: name is required", domain.ErrValidation)
		}
		existing.Name = strings.TrimSpace(*in.Name)
	}
	if in.LastPurchasedAt != nil {
		if in.LastPurchasedAt.IsZero() {
			return nil, fmt.Errorf("%w: last_purchased_at is required", domain.ErrValidation)
		}
		existing.LastPurchasedAt = dateOnly(*in.LastPurchasedAt)
	}
	if in.EstimatedDaysSupply != nil {
		if *in.EstimatedDaysSupply <= 0 {
			return nil, fmt.Errorf("%w: estimated_days_supply must be greater than zero", domain.ErrValidation)
		}
		existing.EstimatedDaysSupply = *in.EstimatedDaysSupply
	}
	if in.Notes != nil {
		existing.Notes = in.Notes
	}
	existing.UpdatedAt = time.Now().UTC()
	updated, err := s.repo.Update(ctx, *existing)
	if err != nil {
		return nil, fmt.Errorf("update supply: %w", err)
	}
	return updated, nil
}

func (s *SupplyService) Delete(ctx context.Context, petID, supplyID string) error {
	if err := s.repo.Delete(ctx, petID, supplyID); err != nil {
		return fmt.Errorf("delete supply %q for pet %q: %w", supplyID, petID, err)
	}
	return nil
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
