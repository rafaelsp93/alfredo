package pets

import (
	"context"
	"testing"

	agentcontracts "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
)

func TestPetHandlers(t *testing.T) {
	handlers := Handlers(agentcontracts.PetToolsDeps{Pets: fakePetService{}})
	specs := Specs()
	if len(specs) != 2 || len(handlers) != 2 {
		t.Fatalf("unexpected counts")
	}
	if specs[0].Name != "list_pets" || handlers[0].Spec().Name != "list_pets" || handlers[1].Spec().Name != "get_pet" {
		t.Fatalf("unexpected specs")
	}
	if _, err := handlers[0].Handle(context.Background(), nil); err != nil {
		t.Fatalf("list pets err = %v", err)
	}
	if _, err := handlers[1].Handle(context.Background(), map[string]any{"pet_id": "pet-1"}); err != nil {
		t.Fatalf("get pet err = %v", err)
	}
	if _, err := handlers[1].Handle(context.Background(), map[string]any{}); err == nil {
		t.Fatal("expected missing pet_id error")
	}
}

type fakePetService struct{}

func (fakePetService) List(context.Context) ([]domain.Pet, error) {
	return []domain.Pet{{ID: "pet-1"}}, nil
}
func (fakePetService) GetByID(context.Context, string) (*domain.Pet, error) {
	return &domain.Pet{ID: "pet-1"}, nil
}
