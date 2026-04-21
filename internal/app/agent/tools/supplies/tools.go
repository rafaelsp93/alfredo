package supplies

import (
	"context"
	"fmt"
	"strings"

	agentdomain "github.com/rafaelsoares/alfredo/internal/agent/domain"
	"github.com/rafaelsoares/alfredo/internal/app/agent/args"
	appagent "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	"github.com/rafaelsoares/alfredo/internal/app/agent/registry"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
)

func Specs() []agentdomain.Tool {
	return []agentdomain.Tool{
		registry.Tool("list_supplies", "List supply records for one pet.", registry.ObjectSchema(registry.Properties("pet_id", "string"), []string{"pet_id"})),
		registry.Tool("get_supply", "Get one supply record by pet_id and supply_id.", registry.ObjectSchema(registry.Properties("pet_id", "string", "supply_id", "string"), []string{"pet_id", "supply_id"})),
		registry.Tool("create_supply", "Create a supply record for one pet.", registry.ObjectSchema(registry.Properties("pet_id", "string", "name", "string", "last_purchased_at", "string", "estimated_days_supply", "integer", "notes", "string"), []string{"pet_id", "name", "last_purchased_at", "estimated_days_supply"})),
		registry.Tool("update_supply", "Update a supply record for one pet.", registry.ObjectSchema(registry.Properties("pet_id", "string", "supply_id", "string", "name", "string", "last_purchased_at", "string", "estimated_days_supply", "integer", "notes", "string"), []string{"pet_id", "supply_id"})),
	}
}

func Handlers(deps appagent.SupplyToolsDeps) []registry.ToolHandler {
	return []registry.ToolHandler{
		listSuppliesHandler{supplies: deps.Supplies},
		getSupplyHandler{supplies: deps.Supplies},
		createSupplyHandler{supplies: deps.Supplies},
		updateSupplyHandler{supplies: deps.Supplies},
	}
}

type listSuppliesHandler struct{ supplies appagent.SupplyServicer }

func (h listSuppliesHandler) Spec() agentdomain.Tool { return Specs()[0] }
func (h listSuppliesHandler) Handle(ctx context.Context, values map[string]any) (any, error) {
	petID, err := args.RequireString(values, "pet_id")
	if err != nil {
		return nil, err
	}
	return h.supplies.List(ctx, petID)
}

type getSupplyHandler struct{ supplies appagent.SupplyServicer }

func (h getSupplyHandler) Spec() agentdomain.Tool { return Specs()[1] }
func (h getSupplyHandler) Handle(ctx context.Context, values map[string]any) (any, error) {
	petID, supplyID, err := args.RequireTwoStrings(values, "pet_id", "supply_id")
	if err != nil {
		return nil, err
	}
	return h.supplies.GetByID(ctx, petID, supplyID)
}

type createSupplyHandler struct{ supplies appagent.SupplyServicer }

func (h createSupplyHandler) Spec() agentdomain.Tool { return Specs()[2] }
func (h createSupplyHandler) Handle(ctx context.Context, values map[string]any) (any, error) {
	in, err := decodeCreateSupply(values)
	if err != nil {
		return nil, err
	}
	return h.supplies.Create(ctx, in)
}

type updateSupplyHandler struct{ supplies appagent.SupplyServicer }

func (h updateSupplyHandler) Spec() agentdomain.Tool { return Specs()[3] }
func (h updateSupplyHandler) Handle(ctx context.Context, values map[string]any) (any, error) {
	petID, supplyID, err := args.RequireTwoStrings(values, "pet_id", "supply_id")
	if err != nil {
		return nil, err
	}
	in, err := decodeUpdateSupply(values)
	if err != nil {
		return nil, err
	}
	return h.supplies.Update(ctx, petID, supplyID, in)
}

func decodeCreateSupply(values map[string]any) (service.CreateSupplyInput, error) {
	petID, err := args.RequireString(values, "pet_id")
	if err != nil {
		return service.CreateSupplyInput{}, err
	}
	name, err := args.RequireString(values, "name")
	if err != nil {
		return service.CreateSupplyInput{}, err
	}
	lastPurchasedAt, err := args.RequireDate(values, "last_purchased_at")
	if err != nil {
		return service.CreateSupplyInput{}, err
	}
	estimated, err := args.RequireInt(values, "estimated_days_supply")
	if err != nil {
		return service.CreateSupplyInput{}, err
	}
	return service.CreateSupplyInput{
		PetID:               petID,
		Name:                name,
		LastPurchasedAt:     lastPurchasedAt,
		EstimatedDaysSupply: estimated,
		Notes:               args.OptionalString(values, "notes"),
	}, nil
}

func decodeUpdateSupply(values map[string]any) (service.UpdateSupplyInput, error) {
	var in service.UpdateSupplyInput
	in.Name = args.OptionalString(values, "name")
	if value, ok := values["last_purchased_at"]; ok && value != nil && strings.TrimSpace(fmt.Sprint(value)) != "" {
		out, err := args.ParseDate(fmt.Sprint(value), "last_purchased_at")
		if err != nil {
			return service.UpdateSupplyInput{}, err
		}
		in.LastPurchasedAt = &out
	}
	if value, ok := values["estimated_days_supply"]; ok && value != nil {
		out, err := args.NumberToInt(value, "estimated_days_supply")
		if err != nil {
			return service.UpdateSupplyInput{}, err
		}
		in.EstimatedDaysSupply = &out
	}
	in.Notes = args.OptionalString(values, "notes")
	return in, nil
}
