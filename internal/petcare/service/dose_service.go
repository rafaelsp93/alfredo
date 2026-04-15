package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/port"
)

type DoseService struct {
	repo port.DoseRepository
}

func NewDoseService(repo port.DoseRepository) *DoseService {
	return &DoseService{repo: repo}
}

// GenerateDoses produces doses from treatment.StartedAt stepping by IntervalHours up to (but not including) upTo.
// For finite treatments, call with upTo = *treatment.EndedAt.
// For open-ended treatments, call with upTo = now + 90 days.
func (s *DoseService) GenerateDoses(t domain.Treatment, upTo time.Time) []domain.Dose {
	var doses []domain.Dose
	for cur := t.StartedAt; cur.Before(upTo); cur = cur.Add(time.Duration(t.IntervalHours) * time.Hour) {
		doses = append(doses, domain.Dose{
			ID:           uuid.New().String(),
			TreatmentID:  t.ID,
			PetID:        t.PetID,
			ScheduledFor: cur,
		})
	}
	return doses
}

// CreateBatch persists a batch of doses.
func (s *DoseService) CreateBatch(ctx context.Context, doses []domain.Dose) error {
	return s.repo.CreateBatch(ctx, doses)
}

// ListByTreatment returns all doses for a treatment ordered by scheduled_for ASC.
func (s *DoseService) ListByTreatment(ctx context.Context, treatmentID string) ([]domain.Dose, error) {
	return s.repo.ListByTreatment(ctx, treatmentID)
}

// DeleteFutureDoses deletes doses scheduled after `after` and returns their IDs.
func (s *DoseService) ListFutureByTreatment(ctx context.Context, treatmentID string, after time.Time) ([]domain.Dose, error) {
	return s.repo.ListFutureByTreatment(ctx, treatmentID, after)
}

func (s *DoseService) DeleteFutureByTreatment(ctx context.Context, treatmentID string, after time.Time) error {
	return s.repo.DeleteFutureByTreatment(ctx, treatmentID, after)
}
