package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
	"go.uber.org/zap"
)

type PetUseCase struct {
	svc      PetCareServicer
	txRunner PetCareTxRunner
	calendar CalendarPort
	logger   *zap.Logger
}

func NewPetUseCase(svc PetCareServicer, txRunner PetCareTxRunner, calendar CalendarPort, logger *zap.Logger) *PetUseCase {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &PetUseCase{svc: svc, txRunner: txRunner, calendar: calendar, logger: logger}
}

func (uc *PetUseCase) List(ctx context.Context) ([]domain.Pet, error) {
	return uc.svc.List(ctx)
}

func (uc *PetUseCase) Create(ctx context.Context, in service.CreatePetInput) (*domain.Pet, error) {
	calendarID, err := uc.calendar.CreateCalendar(ctx, in.Name)
	if err != nil {
		return nil, fmt.Errorf("create calendar for pet %q: %w", in.Name, err)
	}
	in.GoogleCalendarID = calendarID

	var pet *domain.Pet
	err = uc.txRunner.WithinTx(ctx, func(pets *service.PetService, _ *service.VaccineService, _ *service.TreatmentService, _ *service.DoseService) error {
		created, err := pets.Create(ctx, in)
		if err != nil {
			return fmt.Errorf("create pet: %w", err)
		}
		pet = created
		return nil
	})
	if err != nil {
		if delErr := uc.calendar.DeleteCalendar(ctx, calendarID); delErr != nil {
			uc.logger.Error("calendar compensation failed after pet create error",
				zap.String("calendar_id", calendarID),
				zap.Error(delErr),
			)
		}
		return nil, err
	}
	return pet, nil
}

func (uc *PetUseCase) GetByID(ctx context.Context, id string) (*domain.Pet, error) {
	return uc.svc.GetByID(ctx, id)
}

func (uc *PetUseCase) Update(ctx context.Context, id string, in service.UpdatePetInput) (*domain.Pet, error) {
	return uc.svc.Update(ctx, id, in)
}

func (uc *PetUseCase) Delete(ctx context.Context, id string) error {
	var pet *domain.Pet
	externalDeleted := false

	err := uc.txRunner.WithinTx(ctx, func(pets *service.PetService, _ *service.VaccineService, _ *service.TreatmentService, _ *service.DoseService) error {
		loaded, err := pets.GetByID(ctx, id)
		if err != nil {
			return fmt.Errorf("load pet %q: %w", id, err)
		}
		pet = loaded
		if pet.GoogleCalendarID != "" {
			if err := uc.calendar.DeleteCalendar(ctx, pet.GoogleCalendarID); err != nil {
				return fmt.Errorf("delete calendar %q: %w", pet.GoogleCalendarID, err)
			}
			externalDeleted = true
		}
		if err := pets.Delete(ctx, id); err != nil {
			return fmt.Errorf("delete pet %q: %w", id, err)
		}
		return nil
	})
	if err != nil && externalDeleted && errors.Is(err, ErrTxCommit) {
		uc.logger.Error("pet delete committed external change before local commit failed",
			zap.String("pet_id", id),
			zap.String("calendar_id", pet.GoogleCalendarID),
			zap.Error(err),
		)
	}
	return err
}
