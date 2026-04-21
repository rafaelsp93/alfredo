package pets

import (
	"context"

	agentdomain "github.com/rafaelsoares/alfredo/internal/agent/domain"
	"github.com/rafaelsoares/alfredo/internal/app/agent/args"
	appagent "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	"github.com/rafaelsoares/alfredo/internal/app/agent/registry"
)

func Specs() []agentdomain.Tool {
	return []agentdomain.Tool{
		registry.Tool("list_pets", "List every pet Rafael has registered.", registry.ObjectSchema(nil, nil)),
		registry.Tool("get_pet", "Get one pet by pet_id.", registry.ObjectSchema(registry.Properties("pet_id", "string"), []string{"pet_id"})),
	}
}

func Handlers(deps appagent.PetToolsDeps) []registry.ToolHandler {
	return []registry.ToolHandler{
		listPetsHandler{pets: deps.Pets},
		getPetHandler{pets: deps.Pets},
	}
}

type listPetsHandler struct{ pets appagent.PetCareServicer }

func (h listPetsHandler) Spec() agentdomain.Tool { return Specs()[0] }

func (h listPetsHandler) Handle(ctx context.Context, _ map[string]any) (any, error) {
	return h.pets.List(ctx)
}

type getPetHandler struct{ pets appagent.PetCareServicer }

func (h getPetHandler) Spec() agentdomain.Tool { return Specs()[1] }

func (h getPetHandler) Handle(ctx context.Context, values map[string]any) (any, error) {
	petID, err := args.RequireString(values, "pet_id")
	if err != nil {
		return nil, err
	}
	return h.pets.GetByID(ctx, petID)
}
