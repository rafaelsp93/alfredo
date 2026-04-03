package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/rafaelsoares/alfredo/internal/app"
	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
	"go.uber.org/zap"
)

// --- stubs ---

type stubTreatmentService struct {
	treatment *domain.Treatment
}

func (s *stubTreatmentService) Create(_ context.Context, _ service.CreateTreatmentInput) (*domain.Treatment, error) {
	return s.treatment, nil
}
func (s *stubTreatmentService) GetByID(_ context.Context, _, _ string) (*domain.Treatment, error) {
	return s.treatment, nil
}
func (s *stubTreatmentService) List(_ context.Context, _ string) ([]domain.Treatment, error) {
	return []domain.Treatment{*s.treatment}, nil
}
func (s *stubTreatmentService) Stop(_ context.Context, _, _ string) error { return nil }

type stubDoseService struct {
	doses []domain.Dose
}

func (s *stubDoseService) GenerateDoses(_ domain.Treatment, _ time.Time) []domain.Dose {
	return s.doses
}
func (s *stubDoseService) CreateBatch(_ context.Context, _ []domain.Dose) error { return nil }
func (s *stubDoseService) ListByTreatment(_ context.Context, _ string) ([]domain.Dose, error) {
	return s.doses, nil
}
func (s *stubDoseService) DeleteFutureDoses(_ context.Context, _ string, _ time.Time) ([]string, error) {
	return []string{"dose-1", "dose-2"}, nil
}
func (s *stubDoseService) ListOpenEndedActiveTreatments(_ context.Context) ([]domain.Treatment, error) {
	return nil, nil
}
func (s *stubDoseService) ExtendOpenEnded(_ context.Context, _ domain.Treatment, _ time.Time) ([]domain.Dose, error) {
	return s.doses, nil
}

// --- tests ---

func TestTreatmentUseCase_Create_EmitsDosesScheduled(t *testing.T) {
	spy := &spyEmitter{}
	tr := &domain.Treatment{ID: "t1", PetID: "p1", Name: "Amoxicillin", StartedAt: time.Now()}
	doses := []domain.Dose{{ID: "d1", TreatmentID: "t1", ScheduledFor: time.Now()}}
	uc := app.NewTreatmentUseCase(&stubTreatmentService{treatment: tr}, &stubDoseService{doses: doses}, &fakePetGetter{}, spy, zap.NewNop())

	_, _, err := uc.Create(context.Background(), service.CreateTreatmentInput{PetID: "p1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spy.events) != 1 || spy.events[0] != "treatment.doses_scheduled" {
		t.Errorf("events = %v, want [treatment.doses_scheduled]", spy.events)
	}
}

func TestTreatmentUseCase_Stop_EmitsTreatmentStopped(t *testing.T) {
	spy := &spyEmitter{}
	tr := &domain.Treatment{ID: "t1", PetID: "p1", Name: "Amoxicillin", StartedAt: time.Now()}
	uc := app.NewTreatmentUseCase(&stubTreatmentService{treatment: tr}, &stubDoseService{}, &fakePetGetter{}, spy, zap.NewNop())

	if err := uc.Stop(context.Background(), "p1", "t1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spy.events) != 1 || spy.events[0] != "treatment.stopped" {
		t.Errorf("events = %v, want [treatment.stopped]", spy.events)
	}
}
