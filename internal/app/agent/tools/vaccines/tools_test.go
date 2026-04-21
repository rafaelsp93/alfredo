package vaccines

import (
	"context"
	"testing"
	"time"

	agentcontracts "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
)

func TestVaccineHandlers(t *testing.T) {
	handlers := Handlers(agentcontracts.VaccineToolsDeps{Vaccines: fakeVaccineService{}, Location: time.UTC})
	if Specs()[0].Name != "list_vaccines" || handlers[0].Spec().Name != "list_vaccines" || handlers[1].Spec().Name != "record_vaccine" {
		t.Fatalf("unexpected specs")
	}
	if _, err := handlers[0].Handle(context.Background(), map[string]any{"pet_id": "pet-1"}); err != nil {
		t.Fatalf("list err = %v", err)
	}
	if _, err := handlers[1].Handle(context.Background(), map[string]any{"pet_id": "pet-1", "name": "V10", "date": "2026-04-21T10:00:00"}); err != nil {
		t.Fatalf("record err = %v", err)
	}
	if _, err := handlers[1].Handle(context.Background(), map[string]any{"pet_id": "pet-1"}); err == nil {
		t.Fatal("expected decode error")
	}
	if _, err := handlers[0].Handle(context.Background(), map[string]any{}); err == nil {
		t.Fatal("expected list decode error")
	}
	if _, err := handlers[1].Handle(context.Background(), map[string]any{"pet_id": "pet-1", "name": "V10", "date": "bad"}); err == nil {
		t.Fatal("expected date error")
	}
	if _, err := handlers[1].Handle(context.Background(), map[string]any{"pet_id": "pet-1", "name": "V10", "date": "2026-04-21T10:00:00", "recurrence_days": "bad"}); err == nil {
		t.Fatal("expected recurrence error")
	}
}

type fakeVaccineService struct{}

func (fakeVaccineService) ListVaccines(context.Context, string) ([]domain.Vaccine, error) {
	return nil, nil
}
func (fakeVaccineService) RecordVaccine(context.Context, service.RecordVaccineInput) (*domain.Vaccine, error) {
	return &domain.Vaccine{}, nil
}
