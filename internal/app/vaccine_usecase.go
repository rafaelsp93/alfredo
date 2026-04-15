package app

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/rafaelsoares/alfredo/internal/gcalendar"
	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
)

// VaccineUseCase wraps VaccineService and orchestrates calendar side effects.
type VaccineUseCase struct {
	vaccine  VaccineServicer
	pets     PetNameGetter
	txRunner PetCareTxRunner
	calendar CalendarPort
	logger   *zap.Logger
	timezone string
}

func NewVaccineUseCase(vaccine VaccineServicer, pets PetNameGetter, txRunner PetCareTxRunner, calendar CalendarPort, timezone string, logger *zap.Logger) *VaccineUseCase {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &VaccineUseCase{vaccine: vaccine, pets: pets, txRunner: txRunner, calendar: calendar, timezone: timezone, logger: logger}
}

func (uc *VaccineUseCase) ListVaccines(ctx context.Context, petID string) ([]domain.Vaccine, error) {
	return uc.vaccine.ListVaccines(ctx, petID)
}

func (uc *VaccineUseCase) RecordVaccine(ctx context.Context, in service.RecordVaccineInput) (*domain.Vaccine, error) {
	pet, err := uc.pets.GetByID(ctx, in.PetID)
	if err != nil {
		return nil, fmt.Errorf("load pet %q: %w", in.PetID, err)
	}
	if pet.GoogleCalendarID == "" {
		return nil, fmt.Errorf("pet %q is missing google calendar id", in.PetID)
	}

	if in.RecurrenceDays != nil && *in.RecurrenceDays > 0 {
		nextDue := in.AdministeredAt.AddDate(0, 0, *in.RecurrenceDays)
		in.NextDueAt = &nextDue
	}

	eventTime := in.AdministeredAt
	reminderMin := 0
	if in.NextDueAt != nil {
		eventTime = *in.NextDueAt
		reminderMin = 7 * 24 * 60
	}
	eventID, err := uc.calendar.CreateEvent(ctx, pet.GoogleCalendarID, gcalendar.Event{
		Title:       in.Name,
		Description: fmt.Sprintf("Pet: %s", pet.Name),
		StartTime:   eventTime,
		EndTime:     eventTime,
		ReminderMin: reminderMin,
		TimeZone:    uc.timezone,
	})
	if err != nil {
		return nil, fmt.Errorf("create vaccine calendar event: %w", err)
	}

	in.GoogleCalendarEventID = eventID
	var vaccine *domain.Vaccine
	err = uc.txRunner.WithinTx(ctx, func(_ *service.PetService, vaccines *service.VaccineService, _ *service.TreatmentService, _ *service.DoseService) error {
		recorded, err := vaccines.RecordVaccine(ctx, in)
		if err != nil {
			return fmt.Errorf("record vaccine: %w", err)
		}
		vaccine = recorded
		return nil
	})
	if err != nil {
		if delErr := uc.calendar.DeleteEvent(ctx, pet.GoogleCalendarID, eventID); delErr != nil {
			uc.logger.Error("calendar compensation failed after vaccine create error",
				zap.String("pet_id", in.PetID),
				zap.String("calendar_id", pet.GoogleCalendarID),
				zap.String("event_id", eventID),
				zap.Error(delErr),
			)
		}
		return nil, err
	}
	return vaccine, nil
}

func (uc *VaccineUseCase) DeleteVaccine(ctx context.Context, petID, vaccineID string) error {
	var (
		pet             *domain.Pet
		vaccine         *domain.Vaccine
		externalDeleted bool
	)

	err := uc.txRunner.WithinTx(ctx, func(pets *service.PetService, vaccines *service.VaccineService, _ *service.TreatmentService, _ *service.DoseService) error {
		loadedPet, err := pets.GetByID(ctx, petID)
		if err != nil {
			return fmt.Errorf("load pet %q: %w", petID, err)
		}
		pet = loadedPet
		loadedVaccine, err := vaccines.GetVaccine(ctx, petID, vaccineID)
		if err != nil {
			return fmt.Errorf("load vaccine %q: %w", vaccineID, err)
		}
		vaccine = loadedVaccine
		if vaccine.GoogleCalendarEventID != "" {
			if err := uc.calendar.DeleteEvent(ctx, pet.GoogleCalendarID, vaccine.GoogleCalendarEventID); err != nil {
				return fmt.Errorf("delete vaccine calendar event %q: %w", vaccine.GoogleCalendarEventID, err)
			}
			externalDeleted = true
		}
		if err := vaccines.DeleteVaccine(ctx, petID, vaccineID); err != nil {
			return fmt.Errorf("delete vaccine %q: %w", vaccineID, err)
		}
		return nil
	})
	if err != nil && externalDeleted && errors.Is(err, ErrTxCommit) {
		uc.logger.Error("vaccine delete committed external change before local commit failed",
			zap.String("pet_id", petID),
			zap.String("calendar_id", pet.GoogleCalendarID),
			zap.String("event_id", vaccine.GoogleCalendarEventID),
			zap.Error(err),
		)
	}
	return err
}
