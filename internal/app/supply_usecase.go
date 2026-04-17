package app

import (
	"context"
	"fmt"

	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
)

// SupplyUseCase wraps SupplyService and validates parent pet ownership.
type SupplyUseCase struct {
	supplies SupplyServicer
	pets     PetNameGetter
}

func NewSupplyUseCase(supplies SupplyServicer, pets PetNameGetter) *SupplyUseCase {
	return &SupplyUseCase{supplies: supplies, pets: pets}
}

func (uc *SupplyUseCase) Create(ctx context.Context, in service.CreateSupplyInput) (*domain.Supply, error) {
	if _, err := uc.pets.GetByID(ctx, in.PetID); err != nil {
		return nil, fmt.Errorf("load pet %q: %w", in.PetID, err)
	}
	return uc.supplies.Create(ctx, in)
}

func (uc *SupplyUseCase) GetByID(ctx context.Context, petID, supplyID string) (*domain.Supply, error) {
	return uc.supplies.GetByID(ctx, petID, supplyID)
}

func (uc *SupplyUseCase) List(ctx context.Context, petID string) ([]domain.Supply, error) {
	return uc.supplies.List(ctx, petID)
}

func (uc *SupplyUseCase) Update(ctx context.Context, petID, supplyID string, in service.UpdateSupplyInput) (*domain.Supply, error) {
	return uc.supplies.Update(ctx, petID, supplyID, in)
}

func (uc *SupplyUseCase) Delete(ctx context.Context, petID, supplyID string) error {
	return uc.supplies.Delete(ctx, petID, supplyID)
}
