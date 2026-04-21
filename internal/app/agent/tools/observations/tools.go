package observations

import (
	"context"
	"time"

	agentdomain "github.com/rafaelsoares/alfredo/internal/agent/domain"
	"github.com/rafaelsoares/alfredo/internal/app/agent/args"
	appagent "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	"github.com/rafaelsoares/alfredo/internal/app/agent/registry"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
)

func Specs() []agentdomain.Tool {
	return []agentdomain.Tool{
		registry.Tool("list_observations", "List observation history for one pet.", registry.ObjectSchema(registry.Properties("pet_id", "string"), []string{"pet_id"})),
		registry.Tool("log_observation", "Create a new observation entry for one pet.", registry.ObjectSchema(registry.Properties("pet_id", "string", "observed_at", "string", "description", "string"), []string{"pet_id", "observed_at", "description"})),
	}
}

func Handlers(deps appagent.ObservationToolsDeps) []registry.ToolHandler {
	return []registry.ToolHandler{
		listObservationsHandler{observations: deps.Observations},
		logObservationHandler{observations: deps.Observations, location: deps.Location},
	}
}

type listObservationsHandler struct{ observations appagent.ObservationServicer }

func (h listObservationsHandler) Spec() agentdomain.Tool { return Specs()[0] }

func (h listObservationsHandler) Handle(ctx context.Context, values map[string]any) (any, error) {
	petID, err := args.RequireString(values, "pet_id")
	if err != nil {
		return nil, err
	}
	return h.observations.ListByPet(ctx, petID)
}

type logObservationHandler struct {
	observations appagent.ObservationServicer
	location     *time.Location
}

func (h logObservationHandler) Spec() agentdomain.Tool { return Specs()[1] }

func (h logObservationHandler) Handle(ctx context.Context, values map[string]any) (any, error) {
	in, err := decodeCreateObservation(values, h.location)
	if err != nil {
		return nil, err
	}
	return h.observations.Create(ctx, in)
}

func decodeCreateObservation(values map[string]any, location *time.Location) (service.CreateObservationInput, error) {
	petID, err := args.RequireString(values, "pet_id")
	if err != nil {
		return service.CreateObservationInput{}, err
	}
	observedAt, err := args.RequireUserTime(values, "observed_at", location)
	if err != nil {
		return service.CreateObservationInput{}, err
	}
	description, err := args.RequireString(values, "description")
	if err != nil {
		return service.CreateObservationInput{}, err
	}
	return service.CreateObservationInput{PetID: petID, ObservedAt: observedAt, Description: description}, nil
}
