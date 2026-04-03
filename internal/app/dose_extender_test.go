package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/rafaelsoares/alfredo/internal/app"
	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"go.uber.org/zap"
)

// stubDoseServiceForExtender implements DoseServicer for the extender tests.
type stubDoseServiceForExtender struct {
	treatments []domain.Treatment
	extended   []string // treatment IDs passed to ExtendOpenEnded
	newDoses   []domain.Dose
}

func (s *stubDoseServiceForExtender) GenerateDoses(_ domain.Treatment, _ time.Time) []domain.Dose {
	return nil
}
func (s *stubDoseServiceForExtender) CreateBatch(_ context.Context, _ []domain.Dose) error {
	return nil
}
func (s *stubDoseServiceForExtender) ListByTreatment(_ context.Context, _ string) ([]domain.Dose, error) {
	return nil, nil
}
func (s *stubDoseServiceForExtender) DeleteFutureDoses(_ context.Context, _ string, _ time.Time) ([]string, error) {
	return nil, nil
}
func (s *stubDoseServiceForExtender) ListOpenEndedActiveTreatments(_ context.Context) ([]domain.Treatment, error) {
	return s.treatments, nil
}
func (s *stubDoseServiceForExtender) ExtendOpenEnded(_ context.Context, t domain.Treatment, _ time.Time) ([]domain.Dose, error) {
	s.extended = append(s.extended, t.ID)
	return s.newDoses, nil
}

type stubExtenderEmitter struct {
	events []string
}

func (s *stubExtenderEmitter) Emit(_ context.Context, event string, _ any) {
	s.events = append(s.events, event)
}

func TestDoseExtender_ExtendOnce_EmitsForNewDoses(t *testing.T) {
	tr := domain.Treatment{ID: "t1", PetID: "p1", Name: "Drug", IntervalHours: 24, StartedAt: time.Now()}
	doseSvc := &stubDoseServiceForExtender{
		treatments: []domain.Treatment{tr},
		newDoses:   []domain.Dose{{ID: "d1", TreatmentID: "t1", ScheduledFor: time.Now().Add(24 * time.Hour)}},
	}
	spy := &stubExtenderEmitter{}
	ext := app.NewDoseExtender(doseSvc, &fakePetGetter{}, spy, zap.NewNop())

	ext.ExtendOnce(context.Background())

	if len(doseSvc.extended) != 1 || doseSvc.extended[0] != "t1" {
		t.Errorf("extended treatments = %v, want [t1]", doseSvc.extended)
	}
	if len(spy.events) != 1 || spy.events[0] != "treatment.doses_scheduled" {
		t.Errorf("events = %v, want [treatment.doses_scheduled]", spy.events)
	}
}

func TestDoseExtender_ExtendOnce_NoEmitWhenNoDoses(t *testing.T) {
	tr := domain.Treatment{ID: "t1", PetID: "p1", Name: "Drug", IntervalHours: 24, StartedAt: time.Now()}
	doseSvc := &stubDoseServiceForExtender{
		treatments: []domain.Treatment{tr},
		newDoses:   nil, // ExtendOpenEnded returns no new doses
	}
	spy := &stubExtenderEmitter{}
	ext := app.NewDoseExtender(doseSvc, &fakePetGetter{}, spy, zap.NewNop())

	ext.ExtendOnce(context.Background())

	if len(spy.events) != 0 {
		t.Errorf("expected no events when no new doses, got %v", spy.events)
	}
}
