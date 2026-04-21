package observations

import (
	"context"
	"testing"
	"time"

	agentcontracts "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
)

func TestObservationHandlers(t *testing.T) {
	handlers := Handlers(agentcontracts.ObservationToolsDeps{Observations: fakeObservationService{}, Location: time.UTC})
	if len(Specs()) != 2 || handlers[0].Spec().Name != "list_observations" || handlers[1].Spec().Name != "log_observation" {
		t.Fatalf("unexpected specs")
	}
	if _, err := handlers[0].Handle(context.Background(), map[string]any{"pet_id": "pet-1"}); err != nil {
		t.Fatalf("list err = %v", err)
	}
	if _, err := handlers[1].Handle(context.Background(), map[string]any{"pet_id": "pet-1", "observed_at": "2026-04-21T10:00:00", "description": "ok"}); err != nil {
		t.Fatalf("create err = %v", err)
	}
	if _, err := handlers[1].Handle(context.Background(), map[string]any{"pet_id": "pet-1"}); err == nil {
		t.Fatal("expected decode error")
	}
	if _, err := handlers[0].Handle(context.Background(), map[string]any{}); err == nil {
		t.Fatal("expected list decode error")
	}
}

type fakeObservationService struct{}

func (fakeObservationService) Create(context.Context, service.CreateObservationInput) (*domain.Observation, error) {
	return &domain.Observation{}, nil
}
func (fakeObservationService) ListByPet(context.Context, string) ([]domain.Observation, error) {
	return nil, nil
}
