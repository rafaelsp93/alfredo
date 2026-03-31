package app

import (
	"context"
	"time"

	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
	"github.com/rafaelsoares/alfredo/internal/webhook"
)

type PetUseCase struct {
	svc     PetCareServicer
	emitter webhook.EventEmitter
}

func NewPetUseCase(svc PetCareServicer, emitter webhook.EventEmitter) *PetUseCase {
	return &PetUseCase{svc: svc, emitter: emitter}
}

func (uc *PetUseCase) List(ctx context.Context) ([]domain.Pet, error) {
	return uc.svc.List(ctx)
}

func (uc *PetUseCase) Create(ctx context.Context, in service.CreatePetInput) (*domain.Pet, error) {
	pet, err := uc.svc.Create(ctx, in)
	if err != nil {
		return nil, err
	}
	uc.emitter.Emit(ctx, "pet.created", petCreatedPayload{
		ID:        pet.ID,
		Name:      pet.Name,
		Species:   pet.Species,
		Breed:     pet.Breed,
		BirthDate: pet.BirthDate,
	})
	return pet, nil
}

type petCreatedPayload struct {
	ID        string     `json:"id"`
	Name      string     `json:"pet_name"`
	Species   string     `json:"species"`
	Breed     *string    `json:"breed"`
	BirthDate *time.Time `json:"birth_date"`
}

func (uc *PetUseCase) GetByID(ctx context.Context, id string) (*domain.Pet, error) {
	return uc.svc.GetByID(ctx, id)
}

func (uc *PetUseCase) Update(ctx context.Context, id string, in service.UpdatePetInput) (*domain.Pet, error) {
	return uc.svc.Update(ctx, id, in)
}

func (uc *PetUseCase) Delete(ctx context.Context, id string) error {
	pet, err := uc.svc.GetByID(ctx, id)
	if err != nil {
		return err
	}

	err = uc.svc.Delete(ctx, id)
	if err != nil {
		return err
	}

	uc.emitter.Emit(ctx, "pet.deleted", petDeletedPayload{
		PetID: pet.ID,
		Name:  pet.Name,
	})
	return nil
}

type petDeletedPayload struct {
	PetID string `json:"pet_id"`
	Name  string `json:"pet_name"`
}
