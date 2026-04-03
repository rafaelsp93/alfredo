package app

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/rafaelsoares/alfredo/internal/webhook"
)

// DoseExtender is a background job that tops up open-ended treatments with a rolling 90-day dose window.
type DoseExtender struct {
	doses   DoseServicer
	pets    PetNameGetter
	emitter webhook.EventEmitter
	logger  *zap.Logger
}

func NewDoseExtender(doses DoseServicer, pets PetNameGetter, emitter webhook.EventEmitter, logger *zap.Logger) *DoseExtender {
	return &DoseExtender{doses: doses, pets: pets, emitter: emitter, logger: logger}
}

// ExtendOnce runs one pass of the extension job. Call this from a goroutine on a ticker.
func (e *DoseExtender) ExtendOnce(ctx context.Context) {
	treatments, err := e.doses.ListOpenEndedActiveTreatments(ctx)
	if err != nil {
		e.logger.Error("dose extender: list open-ended treatments failed", zap.Error(err))
		return
	}
	windowEnd := time.Now().UTC().AddDate(0, 0, doseWindowDays)
	for _, t := range treatments {
		doses, err := e.doses.ExtendOpenEnded(ctx, t, windowEnd)
		if err != nil {
			e.logger.Error("dose extender: extend failed", zap.String("treatment_id", t.ID), zap.Error(err))
			continue
		}
		if len(doses) == 0 {
			continue
		}
		pet, _ := e.pets.GetByID(ctx, t.PetID)
		petName := t.PetID
		if pet != nil {
			petName = pet.Name
		}
		e.emitter.Emit(ctx, "treatment.doses_scheduled", treatmentDosesScheduledPayload{
			PetID:         t.PetID,
			PetName:       petName,
			TreatmentID:   t.ID,
			TreatmentName: t.Name,
			DosageAmount:  t.DosageAmount,
			DosageUnit:    t.DosageUnit,
			Route:         t.Route,
			IntervalHours: t.IntervalHours,
			Doses:         toDosePayloads(doses),
		})
	}
}

// Run starts the extension job on a daily ticker until ctx is cancelled.
func (e *DoseExtender) Run(ctx context.Context) {
	e.ExtendOnce(ctx) // run immediately at startup, then on daily ticker
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			e.ExtendOnce(ctx)
		case <-ctx.Done():
			return
		}
	}
}
