package supplies

import (
	"context"
	"testing"

	agentcontracts "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
)

func TestSupplyHandlers(t *testing.T) {
	handlers := Handlers(agentcontracts.SupplyToolsDeps{Supplies: fakeSupplyService{}})
	if len(Specs()) != 4 || handlers[0].Spec().Name != "list_supplies" || handlers[1].Spec().Name != "get_supply" || handlers[2].Spec().Name != "create_supply" || handlers[3].Spec().Name != "update_supply" {
		t.Fatalf("unexpected specs")
	}
	if _, err := handlers[0].Handle(context.Background(), map[string]any{"pet_id": "pet-1"}); err != nil {
		t.Fatalf("list err = %v", err)
	}
	if _, err := handlers[1].Handle(context.Background(), map[string]any{"pet_id": "pet-1", "supply_id": "supply-1"}); err != nil {
		t.Fatalf("get err = %v", err)
	}
	if _, err := handlers[2].Handle(context.Background(), map[string]any{"pet_id": "pet-1", "name": "food", "last_purchased_at": "2026-04-21", "estimated_days_supply": 10}); err != nil {
		t.Fatalf("create err = %v", err)
	}
	if _, err := handlers[3].Handle(context.Background(), map[string]any{"pet_id": "pet-1", "supply_id": "supply-1", "name": "food"}); err != nil {
		t.Fatalf("update err = %v", err)
	}
	if _, err := handlers[2].Handle(context.Background(), map[string]any{"pet_id": "pet-1"}); err == nil {
		t.Fatal("expected create decode error")
	}
	if _, err := handlers[3].Handle(context.Background(), map[string]any{"pet_id": "pet-1", "supply_id": "supply-1", "estimated_days_supply": "bad"}); err == nil {
		t.Fatal("expected update decode error")
	}
	if _, err := handlers[0].Handle(context.Background(), map[string]any{}); err == nil {
		t.Fatal("expected list decode error")
	}
	if _, err := handlers[1].Handle(context.Background(), map[string]any{"pet_id": "pet-1"}); err == nil {
		t.Fatal("expected get decode error")
	}
}

type fakeSupplyService struct{}

func (fakeSupplyService) Create(context.Context, service.CreateSupplyInput) (*domain.Supply, error) {
	return &domain.Supply{}, nil
}
func (fakeSupplyService) GetByID(context.Context, string, string) (*domain.Supply, error) {
	return &domain.Supply{}, nil
}
func (fakeSupplyService) List(context.Context, string) ([]domain.Supply, error) { return nil, nil }
func (fakeSupplyService) Update(context.Context, string, string, service.UpdateSupplyInput) (*domain.Supply, error) {
	return &domain.Supply{}, nil
}
