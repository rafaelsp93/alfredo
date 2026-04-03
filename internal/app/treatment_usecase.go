package app

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
	"github.com/rafaelsoares/alfredo/internal/webhook"
)

const doseWindowDays = 90

// TreatmentUseCase orchestrates treatment creation, dose generation, and webhook emission.
type TreatmentUseCase struct {
	treatments TreatmentServicer
	doses      DoseServicer
	pets       PetNameGetter
	emitter    webhook.EventEmitter
	logger     *zap.Logger
}

func NewTreatmentUseCase(
	treatments TreatmentServicer,
	doses DoseServicer,
	pets PetNameGetter,
	emitter webhook.EventEmitter,
	logger *zap.Logger,
) *TreatmentUseCase {
	return &TreatmentUseCase{treatments: treatments, doses: doses, pets: pets, emitter: emitter, logger: logger}
}

func (uc *TreatmentUseCase) petName(ctx context.Context, petID string) string {
	pet, err := uc.pets.GetByID(ctx, petID)
	if err != nil || pet == nil {
		return petID
	}
	return pet.Name
}

// Create starts a treatment, generates doses, and emits treatment.doses_scheduled.
// Returns the treatment and the generated doses.
func (uc *TreatmentUseCase) Create(ctx context.Context, in service.CreateTreatmentInput) (*domain.Treatment, []domain.Dose, error) {
	tr, err := uc.treatments.Create(ctx, in)
	if err != nil {
		return nil, nil, err
	}
	var upTo time.Time
	if tr.EndedAt != nil {
		upTo = *tr.EndedAt
	} else {
		upTo = time.Now().UTC().AddDate(0, 0, doseWindowDays)
	}
	doses := uc.doses.GenerateDoses(*tr, upTo)
	if err := uc.doses.CreateBatch(ctx, doses); err != nil {
		return nil, nil, err
	}
	if len(doses) > 0 {
		uc.emitter.Emit(ctx, "treatment.doses_scheduled", treatmentDosesScheduledPayload{
			PetID:         tr.PetID,
			PetName:       uc.petName(ctx, tr.PetID),
			TreatmentID:   tr.ID,
			TreatmentName: tr.Name,
			DosageAmount:  tr.DosageAmount,
			DosageUnit:    tr.DosageUnit,
			Route:         tr.Route,
			IntervalHours: tr.IntervalHours,
			Doses:         toDosePayloads(doses),
		})
	}
	return tr, doses, nil
}

// GetByID returns a treatment and its doses.
func (uc *TreatmentUseCase) GetByID(ctx context.Context, petID, treatmentID string) (*domain.Treatment, []domain.Dose, error) {
	tr, err := uc.treatments.GetByID(ctx, petID, treatmentID)
	if err != nil {
		return nil, nil, err
	}
	doses, err := uc.doses.ListByTreatment(ctx, treatmentID)
	if err != nil {
		return nil, nil, err
	}
	return tr, doses, nil
}

// List returns all treatments for a pet with their doses.
func (uc *TreatmentUseCase) List(ctx context.Context, petID string) ([]domain.Treatment, map[string][]domain.Dose, error) {
	ts, err := uc.treatments.List(ctx, petID)
	if err != nil {
		return nil, nil, err
	}
	doseMap := make(map[string][]domain.Dose, len(ts))
	for _, t := range ts {
		doses, err := uc.doses.ListByTreatment(ctx, t.ID)
		if err != nil {
			return nil, nil, err
		}
		doseMap[t.ID] = doses
	}
	return ts, doseMap, nil
}

// Stop marks a treatment as stopped, deletes future doses, and emits treatment.stopped.
func (uc *TreatmentUseCase) Stop(ctx context.Context, petID, treatmentID string) error {
	tr, err := uc.treatments.GetByID(ctx, petID, treatmentID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if err := uc.treatments.Stop(ctx, petID, treatmentID); err != nil {
		return err
	}
	deletedIDs, err := uc.doses.DeleteFutureDoses(ctx, treatmentID, now)
	if err != nil {
		return err
	}
	uc.emitter.Emit(ctx, "treatment.stopped", treatmentStoppedPayload{
		PetID:          tr.PetID,
		PetName:        uc.petName(ctx, tr.PetID),
		TreatmentID:    tr.ID,
		TreatmentName:  tr.Name,
		StoppedAt:      now,
		DeletedDoseIDs: deletedIDs,
	})
	return nil
}

// --- Payload types ---

type dosePayload struct {
	DoseID       string    `json:"dose_id"`
	ScheduledFor time.Time `json:"scheduled_for"`
}

type treatmentDosesScheduledPayload struct {
	PetID         string        `json:"pet_id"`
	PetName       string        `json:"pet_name"`
	TreatmentID   string        `json:"treatment_id"`
	TreatmentName string        `json:"treatment_name"`
	DosageAmount  float64       `json:"dosage_amount"`
	DosageUnit    string        `json:"dosage_unit"`
	Route         string        `json:"route"`
	IntervalHours int           `json:"interval_hours"`
	Doses         []dosePayload `json:"doses"`
}

type treatmentStoppedPayload struct {
	PetID          string    `json:"pet_id"`
	PetName        string    `json:"pet_name"`
	TreatmentID    string    `json:"treatment_id"`
	TreatmentName  string    `json:"treatment_name"`
	StoppedAt      time.Time `json:"stopped_at"`
	DeletedDoseIDs []string  `json:"deleted_dose_ids"`
}

func toDosePayloads(doses []domain.Dose) []dosePayload {
	p := make([]dosePayload, len(doses))
	for i, d := range doses {
		p[i] = dosePayload{DoseID: d.ID, ScheduledFor: d.ScheduledFor}
	}
	return p
}
