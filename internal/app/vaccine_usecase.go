package app

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
	"github.com/rafaelsoares/alfredo/internal/webhook"
)

// VaccineUseCase wraps VaccineService and emits domain events for calendar/notification workflows.
type VaccineUseCase struct {
	vaccine VaccineServicer
	pets    PetNameGetter
	emitter webhook.EventEmitter
}

func NewVaccineUseCase(vaccine VaccineServicer, pets PetNameGetter, emitter webhook.EventEmitter, _ *zap.Logger) *VaccineUseCase {
	return &VaccineUseCase{vaccine: vaccine, pets: pets, emitter: emitter}
}

func (uc *VaccineUseCase) petName(ctx context.Context, petID string) string {
	pet, err := uc.pets.GetByID(ctx, petID)
	if err != nil || pet == nil {
		return petID
	}
	return pet.Name
}

func (uc *VaccineUseCase) ListVaccines(ctx context.Context, petID string) ([]domain.Vaccine, error) {
	return uc.vaccine.ListVaccines(ctx, petID)
}

func (uc *VaccineUseCase) RecordVaccine(ctx context.Context, in service.RecordVaccineInput) (*domain.Vaccine, error) {
	v, err := uc.vaccine.RecordVaccine(ctx, in)
	if err != nil {
		return nil, err
	}
	uc.emitter.Emit(ctx, "vaccine.taken", vaccineTakenPayload{
		PetID:       v.PetID,
		PetName:     uc.petName(ctx, v.PetID),
		VaccineID:   v.ID,
		VaccineName: v.Name,
		Date:        v.AdministeredAt,
	})
	if v.NextDueAt != nil {
		uc.emitter.Emit(ctx, "vaccine.expire", vaccineExpirePayload{
			PetID:       v.PetID,
			PetName:     uc.petName(ctx, v.PetID),
			VaccineID:   v.ID,
			VaccineName: v.Name,
			ExpireAt:    *v.NextDueAt,
		})
	}
	return v, nil
}

func (uc *VaccineUseCase) DeleteVaccine(ctx context.Context, petID, vaccineID string) error {
	if err := uc.vaccine.DeleteVaccine(ctx, petID, vaccineID); err != nil {
		return err
	}
	uc.emitter.Emit(ctx, "vaccine.deleted", vaccineDeletedPayload{
		PetID:     petID,
		PetName:   uc.petName(ctx, petID),
		VaccineID: vaccineID,
	})
	return nil
}

// --- Payload types ---

type vaccineTakenPayload struct {
	PetID       string    `json:"pet_id"`
	PetName     string    `json:"pet_name"`
	VaccineID   string    `json:"vaccine_id"`
	VaccineName string    `json:"vaccine_name"`
	Date        time.Time `json:"date"`
}

type vaccineExpirePayload struct {
	PetID       string    `json:"pet_id"`
	PetName     string    `json:"pet_name"`
	VaccineID   string    `json:"vaccine_id"`
	VaccineName string    `json:"vaccine_name"`
	ExpireAt    time.Time `json:"expire_at"`
}

type vaccineDeletedPayload struct {
	PetID     string `json:"pet_id"`
	PetName   string `json:"pet_name"`
	VaccineID string `json:"vaccine_id"`
}
