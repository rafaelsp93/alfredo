package summary

import (
	"context"
	"testing"

	agentcontracts "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
)

func TestSummaryHandler(t *testing.T) {
	handlers := Handlers(agentcontracts.SummaryToolsDeps{Summary: fakeSummaryService{}})
	if len(Specs()) != 1 || handlers[0].Spec().Name != "get_pet_summary" {
		t.Fatalf("unexpected specs")
	}
	if _, err := handlers[0].Handle(context.Background(), nil); err != nil {
		t.Fatalf("summary err = %v", err)
	}
	handlers = Handlers(agentcontracts.SummaryToolsDeps{})
	if _, err := handlers[0].Handle(context.Background(), nil); err == nil {
		t.Fatal("expected missing summary error")
	}
}

type fakeSummaryService struct{}

func (fakeSummaryService) AllPets(context.Context) (domain.AllPetsSummary, error) {
	return domain.AllPetsSummary{}, nil
}
