package summary

import (
	"context"
	"fmt"

	agentdomain "github.com/rafaelsoares/alfredo/internal/agent/domain"
	appagent "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	"github.com/rafaelsoares/alfredo/internal/app/agent/registry"
)

func Specs() []agentdomain.Tool {
	return []agentdomain.Tool{
		registry.Tool("get_pet_summary", "Get the all-pets daily digest data with vaccines due soon, active treatments, upcoming appointments, recent observations, and supplies needing reorder. Use this for resumo diário, digest, pendências, or priorities across all pets.", registry.ObjectSchema(nil, nil)),
	}
}

func Handlers(deps appagent.SummaryToolsDeps) []registry.ToolHandler {
	return []registry.ToolHandler{getPetSummaryHandler{summary: deps.Summary}}
}

type getPetSummaryHandler struct{ summary appagent.SummaryUseCaser }

func (h getPetSummaryHandler) Spec() agentdomain.Tool { return Specs()[0] }

func (h getPetSummaryHandler) Handle(ctx context.Context, _ map[string]any) (any, error) {
	if h.summary == nil {
		return nil, fmt.Errorf("summary use case is not configured")
	}
	return h.summary.AllPets(ctx)
}
