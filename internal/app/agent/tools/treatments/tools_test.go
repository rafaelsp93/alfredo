package treatments

import (
	"context"
	"testing"
	"time"

	agentcontracts "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
)

func TestTreatmentHandlers(t *testing.T) {
	handlers := Handlers(agentcontracts.TreatmentToolsDeps{Treatments: fakeTreatmentService{}, Location: time.UTC})
	if len(Specs()) != 2 || handlers[0].Spec().Name != "list_treatments" || handlers[1].Spec().Name != "start_treatment" {
		t.Fatalf("unexpected specs")
	}
	if _, err := handlers[0].Handle(context.Background(), map[string]any{"pet_id": "pet-1"}); err != nil {
		t.Fatalf("list err = %v", err)
	}
	if _, err := handlers[1].Handle(context.Background(), map[string]any{
		"pet_id": "pet-1", "name": "med", "dosage_amount": 1.0, "dosage_unit": "ml", "route": "oral", "interval_hours": 12, "started_at": "2026-04-21T10:00:00", "ended_at": "2026-04-22T10:00:00",
	}); err != nil {
		t.Fatalf("create err = %v", err)
	}
	if _, err := handlers[1].Handle(context.Background(), map[string]any{"pet_id": "pet-1"}); err == nil {
		t.Fatal("expected decode error")
	}
	if _, err := handlers[0].Handle(context.Background(), map[string]any{}); err == nil {
		t.Fatal("expected list decode error")
	}
}

type fakeTreatmentService struct{}

func (fakeTreatmentService) Create(context.Context, service.CreateTreatmentInput) (*domain.Treatment, []domain.Dose, error) {
	return &domain.Treatment{}, []domain.Dose{{ID: "dose-1"}}, nil
}
func (fakeTreatmentService) List(context.Context, string) ([]domain.Treatment, map[string][]domain.Dose, error) {
	return nil, map[string][]domain.Dose{}, nil
}
